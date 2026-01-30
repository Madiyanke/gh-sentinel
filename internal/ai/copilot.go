package ai

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func SuggestFix(errorLogs string, workflowPath string) (string, string, error) {
	instruction := fmt.Sprintf(
		"As a Senior DevOps, analyze workflow '%s' with logs: %s. "+
		"If SUCCESS, say 'Healthy'. If FAILURE, provide:\n"+
		"1. A 2-sentence explanation.\n"+
		"2. The full corrected YAML in a ```yaml block.",
		workflowPath, errorLogs,
	)

	cmd := exec.Command("gh", "copilot", "-p", instruction)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("erreur Copilot CLI: %v", err)
	}

	rawResult := string(out)

	// Regex pour extraire le bloc de code YAML pur
	re := regexp.MustCompile("(?s)```(?:yaml|yml)?\n(.*?)\n```")
	match := re.FindStringSubmatch(rawResult)
	
	var codePart string
	if len(match) > 1 {
		codePart = strings.TrimSpace(match[1])
	}

	explanationPart := rawResult
	if idx := strings.Index(rawResult, "```"); idx != -1 {
		explanationPart = strings.TrimSpace(rawResult[:idx])
	}

	return explanationPart, codePart, nil
}