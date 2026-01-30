package main

import (
	"fmt"
	"os"

	"gh-sentinel/internal/orchestrator"
	"gh-sentinel/internal/ui"
)

const (
	version = "1.0.0"
)

func main() {
	// Create and run orchestrator
	orch, err := orchestrator.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.FormatError(fmt.Sprintf("Initialization failed: %v", err)))
		printHelp()
		os.Exit(1)
	}

	if err := orch.Run(); err != nil {
		fmt.Fprintln(os.Stderr, ui.FormatError(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}

func printHelp() {
	help := `
🛡️  Sentinel CI - AI-Powered CI/CD Pipeline Repair

PREREQUISITES:
  • gh CLI must be installed and authenticated
  • gh copilot extension must be installed
  • Must be run from a git repository

USAGE:
  gh sentinel

SETUP:
  1. Install gh CLI: https://cli.github.com
  2. Authenticate: gh auth login
  3. Install Copilot: gh extension install github/gh-copilot
  4. Install Sentinel: gh extension install .

FEATURES:
  ✓ Automatic detection of failed workflows
  ✓ AI-powered root cause analysis
  ✓ Surgical YAML fixes with backup
  ✓ Interactive TUI for workflow selection
  ✓ Pattern-based error detection
  ✓ Diff preview before applying changes

VERSION: %s

LEARN MORE: https://github.com/YOUR_USERNAME/gh-sentinel
`
	fmt.Printf(help, version)
}