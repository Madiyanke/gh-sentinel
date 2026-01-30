package analyzer

import (
	"fmt"
	"regexp"
	"strings"

	"gh-sentinel/internal/logger"
)

// Analyzer performs intelligent log analysis
type Analyzer struct {
	logger *logger.Logger
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(log *logger.Logger) *Analyzer {
	return &Analyzer{logger: log}
}

// ErrorPattern represents a known error pattern
type ErrorPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Severity    string
	Suggestion  string
	Category    string
}

// Analysis contains the results of log analysis
type Analysis struct {
	Errors      []DetectedError
	Warnings    []string
	Summary     string
	Confidence  float64
	Category    string
}

// DetectedError represents an error found in logs
type DetectedError struct {
	Pattern     string
	Message     string
	Line        int
	Severity    string
	Suggestion  string
	Category    string
}

// Common error patterns
var errorPatterns = []ErrorPattern{
	{
		Name:        "Node.js Version Deprecated",
		Pattern:     regexp.MustCompile(`(?i)Node\.js \d+ actions? (?:are|is) deprecated`),
		Severity:    "HIGH",
		Suggestion:  "Update to a newer Node.js version in your workflow",
		Category:    "deprecation",
	},
	{
		Name:        "Command Not Found",
		Pattern:     regexp.MustCompile(`(?i)(?:command not found|command '[\w-]+' not found|bash: [\w-]+: command not found)`),
		Severity:    "HIGH",
		Suggestion:  "Install the missing command or check PATH configuration",
		Category:    "dependency",
	},
	{
		Name:        "Python Import Error",
		Pattern:     regexp.MustCompile(`(?i)ModuleNotFoundError:|ImportError:|No module named`),
		Severity:    "HIGH",
		Suggestion:  "Install missing Python dependencies or check requirements.txt",
		Category:    "dependency",
	},
	{
		Name:        "NPM Install Failed",
		Pattern:     regexp.MustCompile(`(?i)npm ERR!|npm install failed|ENOENT.*package\.json`),
		Severity:    "HIGH",
		Suggestion:  "Check package.json or run npm install locally first",
		Category:    "dependency",
	},
	{
		Name:        "YAML Syntax Error",
		Pattern:     regexp.MustCompile(`(?i)yaml.*syntax error|invalid yaml|mapping values are not allowed`),
		Severity:    "CRITICAL",
		Suggestion:  "Fix YAML indentation or syntax errors",
		Category:    "syntax",
	},
	{
		Name:        "Permission Denied",
		Pattern:     regexp.MustCompile(`(?i)permission denied|EACCES`),
		Severity:    "MEDIUM",
		Suggestion:  "Add execute permissions or check file ownership",
		Category:    "permissions",
	},
	{
		Name:        "Docker Build Failed",
		Pattern:     regexp.MustCompile(`(?i)docker build.*failed|ERROR \[.*\]|failed to solve`),
		Severity:    "HIGH",
		Suggestion:  "Check Dockerfile syntax and build context",
		Category:    "docker",
	},
	{
		Name:        "Test Failure",
		Pattern:     regexp.MustCompile(`(?i)test.*failed|FAIL:|❌.*test|\d+ failed,`),
		Severity:    "MEDIUM",
		Suggestion:  "Review test results and fix failing tests",
		Category:    "testing",
	},
	{
		Name:        "Exit Code Non-Zero",
		Pattern:     regexp.MustCompile(`(?i)exit(?:ed)? (?:with )?code \d+|Process completed with exit code \d+`),
		Severity:    "HIGH",
		Suggestion:  "Check the command output above for the actual error",
		Category:    "exit_code",
	},
	{
		Name:        "GitHub Actions Syntax",
		Pattern:     regexp.MustCompile(`(?i)unexpected value|unexpected symbol|Required property is missing`),
		Severity:    "CRITICAL",
		Suggestion:  "Fix workflow YAML syntax according to GitHub Actions schema",
		Category:    "syntax",
	},
}

// AnalyzeLogs performs comprehensive log analysis
func (a *Analyzer) AnalyzeLogs(logs string) *Analysis {
	a.logger.Debug("Analyzing logs (%d chars)", len(logs))

	analysis := &Analysis{
		Errors:   []DetectedError{},
		Warnings: []string{},
	}

	lines := strings.Split(logs, "\n")

	// Pattern matching
	for i, line := range lines {
		for _, pattern := range errorPatterns {
			if pattern.Pattern.MatchString(line) {
				analysis.Errors = append(analysis.Errors, DetectedError{
					Pattern:    pattern.Name,
					Message:    strings.TrimSpace(line),
					Line:       i + 1,
					Severity:   pattern.Severity,
					Suggestion: pattern.Suggestion,
					Category:   pattern.Category,
				})
				a.logger.Debug("Detected error pattern: %s at line %d", pattern.Name, i+1)
			}
		}
	}

	// Categorize and summarize
	if len(analysis.Errors) > 0 {
		analysis.Category = analysis.Errors[0].Category
		analysis.Summary = a.generateSummary(analysis.Errors)
		analysis.Confidence = a.calculateConfidence(analysis.Errors)
	} else {
		analysis.Summary = "No specific error patterns detected"
		analysis.Confidence = 0.3
	}

	a.logger.Info("Analysis complete: %d errors, confidence %.2f", len(analysis.Errors), analysis.Confidence)
	return analysis
}

// generateSummary creates a human-readable summary
func (a *Analyzer) generateSummary(errors []DetectedError) string {
	if len(errors) == 0 {
		return "No errors detected"
	}

	// Group by category
	categories := make(map[string]int)
	for _, err := range errors {
		categories[err.Category]++
	}

	var parts []string
	for cat, count := range categories {
		if count == 1 {
			parts = append(parts, cat)
		} else {
			parts = append(parts, fmt.Sprintf("%d %s errors", count, cat))
		}
	}

	return fmt.Sprintf("Found: %s", strings.Join(parts, ", "))
}

// calculateConfidence estimates confidence in the analysis
func (a *Analyzer) calculateConfidence(errors []DetectedError) float64 {
	if len(errors) == 0 {
		return 0.3
	}

	// More critical errors = higher confidence
	criticalCount := 0
	for _, err := range errors {
		if err.Severity == "CRITICAL" {
			criticalCount++
		}
	}

	baseConfidence := 0.6
	if criticalCount > 0 {
		baseConfidence = 0.9
	} else if len(errors) > 3 {
		baseConfidence = 0.8
	}

	return baseConfidence
}

// ExtractExitCode attempts to extract the exit code from logs
func (a *Analyzer) ExtractExitCode(logs string) int {
	re := regexp.MustCompile(`(?i)exit(?:ed)? (?:with )?code (\d+)`)
	matches := re.FindStringSubmatch(logs)
	if len(matches) > 1 {
		var code int
		fmt.Sscanf(matches[1], "%d", &code)
		return code
	}
	return -1
}

// GetTopSuggestions returns the most relevant suggestions
func (a *Analyzer) GetTopSuggestions(analysis *Analysis, limit int) []string {
	if len(analysis.Errors) == 0 {
		return []string{}
	}

	// Deduplicate suggestions
	seen := make(map[string]bool)
	var suggestions []string

	for _, err := range analysis.Errors {
		if !seen[err.Suggestion] && err.Suggestion != "" {
			suggestions = append(suggestions, err.Suggestion)
			seen[err.Suggestion] = true
			
			if len(suggestions) >= limit {
				break
			}
		}
	}

	return suggestions
}
