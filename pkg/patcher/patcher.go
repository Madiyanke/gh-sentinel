package patcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gh-sentinel/internal/config"
	"gh-sentinel/internal/errors"
	"gh-sentinel/internal/logger"
)

// Patcher handles safe file patching with backup and rollback
type Patcher struct {
	config *config.Config
	logger *logger.Logger
}

// NewPatcher creates a new patcher instance
func NewPatcher(cfg *config.Config, log *logger.Logger) *Patcher {
	return &Patcher{
		config: cfg,
		logger: log,
	}
}

// PatchRequest contains information for a patch operation
type PatchRequest struct {
	FilePath    string
	NewContent  string
	ValidateYAML bool
}

// PatchResult contains the result of a patch operation
type PatchResult struct {
	Success     bool
	BackupPath  string
	Message     string
	LinesAdded  int
	LinesRemoved int
}

// Apply applies a patch to a file with automatic backup
func (p *Patcher) Apply(req *PatchRequest) (*PatchResult, error) {
	p.logger.Info("Applying patch to %s", req.FilePath)

	// Validate input
	if req.NewContent == "" {
		return nil, errors.ValidationError("apply_patch", "empty patch content")
	}

	// Basic YAML validation
	if req.ValidateYAML {
		if err := p.validateYAML(req.NewContent); err != nil {
			return nil, err
		}
	}

	// Read original file
	originalContent, err := os.ReadFile(req.FilePath)
	var backupPath string
	
	if err == nil {
		// File exists - create backup
		if p.config.BackupEnabled {
			backupPath, err = p.createBackup(req.FilePath, originalContent)
			if err != nil {
				return nil, err
			}
			p.logger.Info("Created backup at %s", backupPath)
		}
	} else if !os.IsNotExist(err) {
		// Error reading file (not just "doesn't exist")
		return nil, errors.FilesystemError("apply_patch", req.FilePath, err)
	}

	// Calculate diff stats
	result := &PatchResult{
		BackupPath: backupPath,
	}
	
	if len(originalContent) > 0 {
		result.LinesAdded, result.LinesRemoved = p.calculateDiff(string(originalContent), req.NewContent)
	}

	// Write new content
	if err := os.WriteFile(req.FilePath, []byte(req.NewContent), 0644); err != nil {
		return nil, errors.FilesystemError("apply_patch", req.FilePath, err)
	}

	result.Success = true
	result.Message = fmt.Sprintf("Successfully patched %s", filepath.Base(req.FilePath))
	
	p.logger.Info("Patch applied: +%d -%d lines", result.LinesAdded, result.LinesRemoved)
	return result, nil
}

// createBackup creates a timestamped backup of a file
func (p *Patcher) createBackup(filePath string, content []byte) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.%s%s", filePath, timestamp, p.config.BackupSuffix)

	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return "", errors.FilesystemError("create_backup", backupPath, err)
	}

	return backupPath, nil
}

// Rollback reverts a file to its backup
func (p *Patcher) Rollback(filePath, backupPath string) error {
	p.logger.Info("Rolling back %s from %s", filePath, backupPath)

	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		return errors.FilesystemError("rollback", backupPath, err)
	}

	if err := os.WriteFile(filePath, backupContent, 0644); err != nil {
		return errors.FilesystemError("rollback", filePath, err)
	}

	p.logger.Info("Rollback successful")
	return nil
}

// validateYAML performs basic YAML structure validation
func (p *Patcher) validateYAML(content string) error {
	// Basic checks for YAML structure
	if !strings.Contains(content, "name:") && !strings.Contains(content, "jobs:") && !strings.Contains(content, "on:") {
		return errors.ValidationError("validate_yaml", "content doesn't appear to be a valid GitHub Actions workflow")
	}

	// Check for common YAML issues
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Check for tabs (YAML must use spaces)
		if strings.Contains(line, "\t") {
			return errors.ValidationError("validate_yaml", fmt.Sprintf("line %d contains tabs (YAML requires spaces)", i+1))
		}
	}

	return nil
}

// calculateDiff calculates rough diff statistics
func (p *Patcher) calculateDiff(original, new string) (added, removed int) {
	originalLines := strings.Split(original, "\n")
	newLines := strings.Split(new, "\n")

	// Simple line-based diff
	originalSet := make(map[string]bool)
	for _, line := range originalLines {
		originalSet[line] = true
	}

	newSet := make(map[string]bool)
	for _, line := range newLines {
		newSet[line] = true
	}

	// Count additions
	for _, line := range newLines {
		if !originalSet[line] && strings.TrimSpace(line) != "" {
			added++
		}
	}

	// Count removals
	for _, line := range originalLines {
		if !newSet[line] && strings.TrimSpace(line) != "" {
			removed++
		}
	}

	return added, removed
}

// PreviewDiff generates a human-readable diff preview
func (p *Patcher) PreviewDiff(filePath, newContent string) (string, error) {
	originalContent, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "[NEW FILE]\n" + newContent, nil
		}
		return "", errors.FilesystemError("preview_diff", filePath, err)
	}

	var preview strings.Builder
	preview.WriteString(fmt.Sprintf("=== Changes to %s ===\n\n", filepath.Base(filePath)))

	originalLines := strings.Split(string(originalContent), "\n")
	newLines := strings.Split(newContent, "\n")

	// Simple side-by-side preview (first 20 lines)
	maxLines := 20
	if len(newLines) < maxLines {
		maxLines = len(newLines)
	}

	for i := 0; i < maxLines; i++ {
		var orig, new string
		if i < len(originalLines) {
			orig = originalLines[i]
		}
		if i < len(newLines) {
			new = newLines[i]
		}

		if orig != new {
			if orig != "" {
				preview.WriteString(fmt.Sprintf("- %s\n", orig))
			}
			if new != "" {
				preview.WriteString(fmt.Sprintf("+ %s\n", new))
			}
		}
	}

	if len(newLines) > maxLines {
		preview.WriteString(fmt.Sprintf("\n... (%d more lines)\n", len(newLines)-maxLines))
	}

	return preview.String(), nil
}

// ListBackups finds all backup files for a given path
func (p *Patcher) ListBackups(filePath string) ([]string, error) {
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.FilesystemError("list_backups", dir, err)
	}

	var backups []string
	pattern := base + "."
	
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), pattern) && strings.HasSuffix(entry.Name(), p.config.BackupSuffix) {
			backups = append(backups, filepath.Join(dir, entry.Name()))
		}
	}

	return backups, nil
}
