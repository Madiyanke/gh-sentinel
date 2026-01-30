package engine

import (
	"fmt"
	"os"
	"strings"
)

func ApplyPatch(filePath string, newContent string) error {
	if !strings.Contains(newContent, "jobs:") && !strings.Contains(newContent, "name:") {
		return fmt.Errorf("le correctif YAML semble invalide")
	}

	old, err := os.ReadFile(filePath)
	if err == nil {
		_ = os.WriteFile(filePath+".bak", old, 0644)
	}

	return os.WriteFile(filePath, []byte(newContent), 0644)
}