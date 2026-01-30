package ai

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func SuggestFix(errorLogs string, currentSelection string, availableFiles []string) (string, string, string, error) {
	// 1. Lecture du contenu du fichier suspect pour donner le contexte à l'IA
	fileContentBytes, err := os.ReadFile(currentSelection)
	var fileContent string
	if err != nil {
		// Si on ne peut pas lire le fichier (ex: erreur de chemin), on laisse vide, l'IA devinera
		fileContent = "[Local file not accessible, rely on standard structure]"
	} else {
		fileContent = string(fileContentBytes)
	}

	// Troncature des logs (Buffer Safety)
	const maxBuffer = 6000 // On réduit un peu pour laisser de la place au contenu du fichier
	if len(errorLogs) > maxBuffer {
		errorLogs = "... [Truncated LOGS] ...\n" + errorLogs[len(errorLogs)-maxBuffer:]
	}

	filesContext := strings.Join(availableFiles, ", ")

instruction := fmt.Sprintf(
		"### ROLE: Senior DevOps Architect\n"+
		"### GOAL: Analyze the logs, diagnose the crash, and repair the workflow.\n\n"+
		"### CONTEXT\n"+
		"File: %s\n"+
		"Siblings: [%s]\n"+
		"Logs:\n%s\n"+
		"Content:\n```yaml\n%s\n```\n\n"+
		"### INSTRUCTIONS\n"+
		"1. **Analyze**: Read the logs to pinpoint exactly why the job failed (e.g., exit codes, error messages).\n"+
		"2. **Diagnose**: Correlate the log error with the specific line in the 'Content'.\n"+
		"3. **Fix**: Rewrite the YAML to resolve the error while keeping the logic intended by the user (unless the logic itself is the bug).\n"+
		"4. **Output Full File**: Return the ENTIRE, VALID YAML file. DO NOT use placeholders.\n"+
		"5. **Preserve Format**: Keep indentation and structure exactly as is.\n\n"+
		"### OUTPUT FORMAT\n"+
		"FIX_TARGET: [filename]\n"+
		"EXPLANATION: [Concise root cause analysis]\n"+
		"```yaml\n[FULL YAML CONTENT]\n```",
		currentSelection, filesContext, strings.ReplaceAll(errorLogs, "\"", "'"), fileContent,
	)

	cmd := exec.Command("gh", "copilot", "-p", instruction)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", "", fmt.Errorf("Copilot Error: %v", err)
	}

	rawResult := string(out)

	// --- Logique de Sanitization du Chemin (Gardée de l'étape précédente) ---
	targetPath := currentSelection
	targetRe := regexp.MustCompile(`(?i)FIX_TARGET:\s*([^\s\n\r]+)`)
	if targetMatch := targetRe.FindStringSubmatch(rawResult); len(targetMatch) > 1 {
		extracted := strings.Trim(targetMatch[1], "[]`* \"'")
		extracted = strings.ReplaceAll(extracted, "\\", "/")
		if strings.HasPrefix(extracted, ".github/workflows/") {
			targetPath = extracted
		} else if strings.HasPrefix(extracted, "github/workflows/") {
			targetPath = "." + extracted
		} else {
			cleanName := strings.TrimPrefix(extracted, "/")
			targetPath = ".github/workflows/" + cleanName
		}
	}

	re := regexp.MustCompile("(?s)```(?:yaml|yml)?\n(.*?)\n```")
	match := re.FindStringSubmatch(rawResult)
	var codePart string
	if len(match) > 1 {
		codePart = strings.TrimSpace(match[1])
	}

	if strings.Contains(strings.ToUpper(rawResult), "STATUS: HEALTHY") {
		return rawResult, "", targetPath, nil
	}

	return rawResult, codePart, targetPath, nil
}