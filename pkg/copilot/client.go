package copilot

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"gh-sentinel/internal/config"
	"gh-sentinel/internal/errors"
	"gh-sentinel/internal/logger"
)

// Client handles interaction with GitHub Copilot CLI
type Client struct {
	config *config.Config
	logger *logger.Logger
}

// NewClient creates a new Copilot client
func NewClient(cfg *config.Config, log *logger.Logger) (*Client, error) {
	// Verify gh copilot is available
	cmd := exec.Command("gh", "copilot", "--version")
	if err := cmd.Run(); err != nil {
		return nil, errors.CopilotError("new_client", fmt.Errorf("gh copilot not available - install with: gh extension install github/gh-copilot"))
	}

	return &Client{
		config: cfg,
		logger: log,
	}, nil
}

// DiagnosisRequest contains all information needed for diagnosis
type DiagnosisRequest struct {
	ErrorLogs      string
	CurrentFile    string
	FileContent    string
	AvailableFiles []string
	WorkflowPath   string
}

// DiagnosisResult contains the AI diagnosis and fix suggestion
type DiagnosisResult struct {
	Explanation  string
	FixedContent string
	TargetFile   string
	Confidence   string
}

// DiagnoseAndFix uses GitHub Copilot to analyze errors and suggest fixes
func (c *Client) DiagnoseAndFix(req *DiagnosisRequest) (*DiagnosisResult, error) {
	c.logger.Info("Requesting AI diagnosis for %s", req.CurrentFile)

	// Truncate logs if necessary
	logs := req.ErrorLogs
	if len(logs) > c.config.MaxLogSize {
		logs = "... [Truncated for buffer safety] ...\n" + logs[len(logs)-c.config.MaxLogSize:]
		c.logger.Debug("Truncated logs from %d to %d chars", len(req.ErrorLogs), len(logs))
	}

	// Build context-rich prompt
	prompt := c.buildDiagnosisPrompt(req, logs)

	// Execute gh copilot
	cmd := exec.Command("gh", "copilot", "-p", prompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.CopilotError("diagnose_and_fix", fmt.Errorf("copilot execution failed: %v\nOutput: %s", err, string(output)))
	}

	rawResult := string(output)
	c.logger.Debug("Received %d bytes from Copilot", len(rawResult))

	// Parse the result
	result, err := c.parseResponse(rawResult, req.CurrentFile)
	if err != nil {
		return nil, err
	}

	c.logger.Info("Diagnosis complete - Target: %s, Confidence: %s", result.TargetFile, result.Confidence)
	return result, nil
}

// buildDiagnosisPrompt creates a comprehensive prompt for Copilot
func (c *Client) buildDiagnosisPrompt(req *DiagnosisRequest, logs string) string {
	filesContext := strings.Join(req.AvailableFiles, ", ")
	
	// Escape quotes in logs
	safeErrorLogs := strings.ReplaceAll(logs, `"`, `'`)
	
	prompt := fmt.Sprintf(`### ROLE: Senior DevOps Security Architect

### MISSION: Critical CI/CD Pipeline Recovery

You are analyzing a failed GitHub Actions workflow. Your goal is to:
1. Identify the ROOT CAUSE of the failure
2. Determine the EXACT file that needs modification
3. Provide a COMPLETE, WORKING fix

### CONTEXT

**Repository Workflow Files:**
%s

**Suspected File:** %s

**Current File Content:**
`+"```yaml\n%s\n```"+`

**Failure Logs:**
%s

### ANALYSIS REQUIREMENTS

1. **Root Cause Analysis:** Examine the logs to find the exact error (exit codes, syntax errors, missing dependencies, etc.)
2. **Target Identification:** The suspected file may not be the actual culprit. Check logs for references to other workflow files.
3. **Surgical Fix:** Provide the COMPLETE file content with the fix applied. NO placeholders, NO comments like "# rest of file unchanged"

### OUTPUT FORMAT (STRICT)

FIX_TARGET: [exact-filename.yml]
CONFIDENCE: [HIGH|MEDIUM|LOW]

EXPLANATION:
[Concise 2-3 sentence root cause explanation]

FIXED_CONTENT:
`+"```yaml\n[COMPLETE YAML FILE WITH FIX APPLIED]\n```"+`

### EXAMPLES OF GOOD OUTPUT

FIX_TARGET: ci.yml
CONFIDENCE: HIGH

EXPLANATION: 
The workflow failed because the Node.js setup action version is deprecated. The logs show "Node.js 12 actions are deprecated" on line 23. The fix updates setup-node from v1 to v4.

FIXED_CONTENT:
`+"```yaml\nname: CI\n...[COMPLETE FILE]...\n```"+`

### CRITICAL RULES

- Output ONLY the format above
- Include the ENTIRE file in FIXED_CONTENT
- Match the original indentation exactly
- If the workflow is actually healthy, use: CONFIDENCE: HEALTHY`,
		filesContext,
		req.CurrentFile,
		req.FileContent,
		safeErrorLogs,
	)

	return prompt
}

// parseResponse extracts structured information from Copilot's response
func (c *Client) parseResponse(rawResponse string, defaultTarget string) (*DiagnosisResult, error) {
	result := &DiagnosisResult{
		TargetFile: defaultTarget,
		Confidence: "MEDIUM",
	}

	// Extract target file
	targetRe := regexp.MustCompile(`(?i)FIX_TARGET:\s*([^\s\n\r]+)`)
	if match := targetRe.FindStringSubmatch(rawResponse); len(match) > 1 {
		extracted := strings.Trim(match[1], "[]`* \"'")
		result.TargetFile = c.normalizeWorkflowPath(extracted)
		c.logger.Debug("Extracted target: %s (normalized to %s)", match[1], result.TargetFile)
	}

	// Extract confidence
	confidenceRe := regexp.MustCompile(`(?i)CONFIDENCE:\s*([A-Z]+)`)
	if match := confidenceRe.FindStringSubmatch(rawResponse); len(match) > 1 {
		result.Confidence = strings.ToUpper(match[1])
	}

	// Check if healthy
	if result.Confidence == "HEALTHY" || strings.Contains(strings.ToUpper(rawResponse), "STATUS: HEALTHY") {
		result.Explanation = "Workflow appears healthy according to AI analysis"
		return result, nil
	}

	// Extract explanation
	explanationRe := regexp.MustCompile(`(?is)EXPLANATION:\s*(.+?)(?:FIXED_CONTENT:|$)`)
	if match := explanationRe.FindStringSubmatch(rawResponse); len(match) > 1 {
		result.Explanation = strings.TrimSpace(match[1])
	} else {
		// Fallback: use entire response as explanation
		result.Explanation = rawResponse
	}

	// Extract YAML fix
	yamlRe := regexp.MustCompile("(?s)```(?:yaml|yml)?\\n(.*?)\\n```")
	if match := yamlRe.FindStringSubmatch(rawResponse); len(match) > 1 {
		result.FixedContent = strings.TrimSpace(match[1])
		c.logger.Debug("Extracted YAML fix: %d lines", strings.Count(result.FixedContent, "\n")+1)
	} else {
		c.logger.Warn("No YAML code block found in Copilot response")
	}

	// Validate we got meaningful output
	if result.FixedContent == "" && result.Confidence != "HEALTHY" {
		return nil, errors.ValidationError("parse_copilot_response", "no actionable fix found in AI response")
	}

	return result, nil
}

// normalizeWorkflowPath ensures the path is in the correct format
func (c *Client) normalizeWorkflowPath(path string) string {
	// Remove quotes and extra characters
	path = strings.Trim(path, "[]`* \"'")
	
	// Normalize backslashes to forward slashes
	path = strings.ReplaceAll(path, "\\", "/")
	
	// Ensure it starts with .github/workflows/
	if strings.HasPrefix(path, ".github/workflows/") {
		return path
	}
	
	if strings.HasPrefix(path, "github/workflows/") {
		return "." + path
	}
	
	if strings.HasPrefix(path, "workflows/") {
		return ".github/" + path
	}
	
	// Just a filename - prepend full path
	cleanName := strings.TrimPrefix(path, "/")
	return ".github/workflows/" + cleanName
}

// QuickDiagnose provides a quick diagnosis without full file context
func (c *Client) QuickDiagnose(errorLogs string) (string, error) {
	prompt := fmt.Sprintf(`Analyze this CI/CD failure log and explain the root cause in 2-3 sentences:

%s`, errorLogs)

	cmd := exec.Command("gh", "copilot", "-p", prompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.CopilotError("quick_diagnose", err)
	}

	return string(output), nil
}
