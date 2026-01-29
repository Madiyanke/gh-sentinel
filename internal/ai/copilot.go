package ai

import (
	"fmt"
	"os/exec"
	"strings"
)

func SuggestFix(errorLogs string) (string, error) {
	// Trimming intelligent : on garde le début (contexte) et la fin (erreur)
	lines := strings.Split(errorLogs, "\n")
	content := errorLogs
	if len(lines) > 60 {
		content = strings.Join(lines[:20], "\n") + "\n...[TRUNCATED]...\n" + strings.Join(lines[len(lines)-40:], "\n")
	}

	// Prompt optimisé pour le "Direct Fix"
	prompt := fmt.Sprintf("Act as a Senior DevOps. Analyze this GitHub Action failure and provide the exact file change needed. Logs: %s", content)

	// Note : On utilise 'explain' car 'suggest' est souvent trop interactif pour exec.Command
	cmd := exec.Command("gh", "copilot", "explain", prompt)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Copilot inaccessible : %w", err)
	}

	return string(out), nil
}