package orchestrator

import (
	"fmt"
	"strings"

	"gh-sentinel/internal/config"
	"gh-sentinel/internal/logger"
	"gh-sentinel/internal/ui"
	"gh-sentinel/pkg/analyzer"
	"gh-sentinel/pkg/copilot"
	"gh-sentinel/pkg/github"
	"gh-sentinel/pkg/patcher"
)

// Orchestrator coordinates all sentinel operations
type Orchestrator struct {
	config   *config.Config
	logger   *logger.Logger
	github   *github.Client
	copilot  *copilot.Client
	analyzer *analyzer.Analyzer
	patcher  *patcher.Patcher
}

// New creates a new orchestrator instance
func New() (*Orchestrator, error) {
	// Initialize configuration
	cfg := config.Default()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	if err := cfg.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Initialize logger
	log := logger.Default()

	// Initialize GitHub client
	ghClient, err := github.NewClient(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GitHub client: %w", err)
	}

	// Initialize Copilot client
	copilotClient, err := copilot.NewClient(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Copilot client: %w", err)
	}

	// Initialize analyzer and patcher
	analyzer := analyzer.NewAnalyzer(log)
	patcher := patcher.NewPatcher(cfg, log)

	return &Orchestrator{
		config:   cfg,
		logger:   log,
		github:   ghClient,
		copilot:  copilotClient,
		analyzer: analyzer,
		patcher:  patcher,
	}, nil
}

// Run executes the main sentinel workflow
func (o *Orchestrator) Run() error {
	// Display banner
	ui.PrintBanner()

	repo := o.github.GetRepository()
	fmt.Println(ui.FormatInfo(fmt.Sprintf("Repository: %s", ui.FormatHighlight(repo.FullName))))
	fmt.Println(ui.FormatDim("Scanning for failed workflows...\n"))

	// Step 1: Get workflow files list
	workflowFiles, err := o.github.ListWorkflowFiles()
	if err != nil {
		return fmt.Errorf("failed to list workflow files: %w", err)
	}
	o.logger.Debug("Found workflow files: %v", workflowFiles)

	// Step 2: Get failed workflow runs
	runs, err := o.github.GetFailedWorkflowRuns(10)
	if err != nil {
		return fmt.Errorf("failed to get workflow runs: %w", err)
	}

	if len(runs) == 0 {
		fmt.Println(ui.FormatSuccess("System Clean. No failures detected! ✨"))
		return nil
	}

	fmt.Println(ui.FormatWarning(fmt.Sprintf("Found %d failed workflow runs", len(runs))))

	// Step 3: User selects a workflow to analyze
	items := o.convertToUIItems(runs)
	selected, err := ui.ShowWorkflowSelector(items)
	if err != nil {
		return fmt.Errorf("failed to show selector: %w", err)
	}

	if selected == nil {
		fmt.Println(ui.FormatDim("Operation cancelled"))
		return nil
	}

	// Step 4: Analyze the selected run
	return o.analyzeAndFix(selected, workflowFiles)
}

// convertToUIItems converts workflow runs to UI items
func (o *Orchestrator) convertToUIItems(runs []*github.WorkflowRun) []ui.WorkflowItem {
	var items []ui.WorkflowItem
	for _, run := range runs {
		icon := o.getStatusIcon(run.Conclusion)
		items = append(items, ui.WorkflowItem{
			ID:          run.ID,
			TitleText:   run.DisplayTitle,
			DescText:    fmt.Sprintf("Run #%d • %s • %s", run.RunNumber, run.Event, run.UpdatedAt.Format("Jan 02, 15:04")),
			Status:      run.Status,
			Conclusion:  run.Conclusion,
			Path:        run.WorkflowPath,
			Icon:        icon,
		})
	}
	return items
}

// getStatusIcon returns an icon based on workflow conclusion
func (o *Orchestrator) getStatusIcon(conclusion string) string {
	switch conclusion {
	case "success":
		return "✅"
	case "failure":
		return "❌"
	case "cancelled":
		return "🚫"
	case "skipped":
		return "⏭️"
	default:
		return "⏳"
	}
}

// analyzeAndFix performs the full analysis and fix workflow
func (o *Orchestrator) analyzeAndFix(selected *ui.WorkflowItem, workflowFiles []string) error {
	fmt.Println("\n" + ui.FormatHeader("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(ui.FormatHeader(fmt.Sprintf("🔍 Analyzing Run #%d", selected.ID)))
	fmt.Println(ui.FormatHeader("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))

	// Step 1: Fetch logs (if available)
	fmt.Println(ui.FormatInfo("Fetching job logs..."))
	logs, err := o.github.GetWorkflowJobLogs(selected.ID)
	
	// If no job logs, this might be a configuration error
	// Continue anyway and let Copilot analyze the workflow file
	if err != nil {
		o.logger.Warn("Could not retrieve job logs: %v", err)
		fmt.Println(ui.FormatWarning("⚠ No job logs available (possible workflow configuration error)"))
		logs = "[No job execution logs available - workflow may have configuration error]"
		fmt.Println(ui.FormatInfo("Proceeding with workflow file analysis...\n"))
	} else {
		o.logger.Debug("Retrieved %d chars of logs", len(logs))
	}

	// Step 2: Quick pattern analysis (skip if no real logs)
	var analysis *analyzer.Analysis
	if logs != "" && !strings.Contains(logs, "[No job execution logs") {
		fmt.Println(ui.FormatInfo("Running pattern analysis..."))
		analysis = o.analyzer.AnalyzeLogs(logs)

		if len(analysis.Errors) > 0 {
			fmt.Println(ui.FormatWarning(fmt.Sprintf("\nDetected %d potential issues:", len(analysis.Errors))))
			for i, err := range analysis.Errors {
				if i >= 3 {
					break // Show top 3
				}
				fmt.Printf("  %d. %s: %s\n", i+1, ui.FormatHighlight(err.Pattern), err.Message[:min(80, len(err.Message))])
			}
		}

		suggestions := o.analyzer.GetTopSuggestions(analysis, 3)
		if len(suggestions) > 0 {
			fmt.Println(ui.FormatInfo("\n💡 Quick Suggestions:"))
			for i, suggestion := range suggestions {
				fmt.Printf("  %d. %s\n", i+1, suggestion)
			}
		}
		fmt.Println()
	}

	// Step 3: Get file content
	fileContent, err := o.github.GetWorkflowFileContent(selected.Path)
	if err != nil {
		o.logger.Warn("Failed to fetch remote file content: %v", err)
		fileContent = "[Remote file not accessible]"
	}

	// Step 4: AI Diagnosis
	fmt.Println(ui.FormatInfo("Consulting AI for diagnosis..."))
	diagnosisReq := &copilot.DiagnosisRequest{
		ErrorLogs:      logs,
		CurrentFile:    selected.Path,
		FileContent:    fileContent,
		AvailableFiles: workflowFiles,
		WorkflowPath:   selected.Path,
	}

	diagnosis, err := o.copilot.DiagnoseAndFix(diagnosisReq)
	if err != nil {
		return fmt.Errorf("AI diagnosis failed: %w", err)
	}

	// Display results
	o.displayDiagnosisResults(diagnosis, selected.Path)

	// Step 5: Apply fix if available
	if diagnosis.FixedContent != "" && diagnosis.Confidence != "HEALTHY" {
		return o.applyFix(diagnosis)
	}

	fmt.Println(ui.FormatInfo("No actionable fix required"))
	return nil
}

// displayDiagnosisResults shows the diagnosis results
func (o *Orchestrator) displayDiagnosisResults(diagnosis *copilot.DiagnosisResult, originalPath string) {
	fmt.Println("\n" + ui.FormatHeader("━━━━━━━━━━━━━━ DIAGNOSIS REPORT ━━━━━━━━━━━━━━\n"))

	// Target redirection?
	if diagnosis.TargetFile != originalPath {
		fmt.Println(ui.FormatWarning("🎯 Target Redirection Detected"))
		fmt.Printf("   User Selected: %s\n", ui.FormatDim(originalPath))
		fmt.Printf("   AI Identified: %s\n", ui.FormatHighlight(diagnosis.TargetFile))
		fmt.Println()
	}

	// Confidence
	confidenceStyle := ui.FormatSuccess
	if diagnosis.Confidence == "MEDIUM" {
		confidenceStyle = ui.FormatWarning
	} else if diagnosis.Confidence == "LOW" {
		confidenceStyle = ui.FormatError
	}
	fmt.Printf("Confidence: %s\n\n", confidenceStyle(diagnosis.Confidence))

	// Explanation
	fmt.Println(ui.FormatHeader("Root Cause:"))
	fmt.Println(wrapText(diagnosis.Explanation, 80))
	fmt.Println()
}

// applyFix applies the suggested fix
func (o *Orchestrator) applyFix(diagnosis *copilot.DiagnosisResult) error {
	fmt.Println(ui.FormatHeader("━━━━━━━━━━━━━━ PROPOSED FIX ━━━━━━━━━━━━━━\n"))

	// Show diff preview
	diff, err := o.patcher.PreviewDiff(diagnosis.TargetFile, diagnosis.FixedContent)
	if err != nil {
		o.logger.Warn("Could not generate diff preview: %v", err)
	} else {
		// Show first 15 lines of diff
		lines := strings.Split(diff, "\n")
		previewLines := lines
		if len(lines) > 15 {
			previewLines = lines[:15]
		}
		for _, line := range previewLines {
			if strings.HasPrefix(line, "+") {
				fmt.Println(ui.FormatSuccess(line))
			} else if strings.HasPrefix(line, "-") {
				fmt.Println(ui.FormatError(line))
			} else {
				fmt.Println(ui.FormatDim(line))
			}
		}
		if len(lines) > 15 {
			fmt.Println(ui.FormatDim(fmt.Sprintf("... (%d more lines)", len(lines)-15)))
		}
		fmt.Println()
	}

	// Confirm with user
	confirmed, err := ui.ShowConfirmation(
		fmt.Sprintf("Apply patch to %s?", diagnosis.TargetFile),
		"A backup will be created automatically",
	)
	if err != nil {
		return fmt.Errorf("confirmation dialog failed: %w", err)
	}

	if !confirmed {
		fmt.Println(ui.FormatDim("Patch cancelled by user"))
		return nil
	}

	// Apply patch
	fmt.Println(ui.FormatInfo("Applying patch..."))
	patchReq := &patcher.PatchRequest{
		FilePath:    diagnosis.TargetFile,
		NewContent:  diagnosis.FixedContent,
		ValidateYAML: true,
	}

	result, err := o.patcher.Apply(patchReq)
	if err != nil {
		return fmt.Errorf("failed to apply patch: %w", err)
	}

	// Success!
	fmt.Println()
	fmt.Println(ui.FormatSuccess(fmt.Sprintf("✓ %s patched successfully!", diagnosis.TargetFile)))
	if result.BackupPath != "" {
		fmt.Println(ui.FormatDim(fmt.Sprintf("  Backup: %s", result.BackupPath)))
	}
	fmt.Println(ui.FormatDim(fmt.Sprintf("  Changes: +%d -%d lines", result.LinesAdded, result.LinesRemoved)))
	fmt.Println()
	fmt.Println(ui.FormatInfo("💡 Next steps:"))
	fmt.Println("  1. Review the changes")
	fmt.Println("  2. Commit and push to trigger a new workflow run")
	fmt.Println("  3. Monitor the results")

	return nil
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func wrapText(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		if len(currentLine)+len(word)+1 > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		} else {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}
