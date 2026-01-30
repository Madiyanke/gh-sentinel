package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds all application configuration
type Config struct {
	Version       string
	UserAgent     string
	MaxLogSize    int
	RequestTimeout time.Duration
	BackupEnabled bool
	BackupSuffix  string
	TempDir       string
	CacheDir      string
}

// Default returns a production-ready configuration
func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".gh-sentinel", "cache")
	tempDir := filepath.Join(homeDir, ".gh-sentinel", "tmp")
	
	return &Config{
		Version:       "1.0.0",
		UserAgent:     "gh-sentinel/1.0.0",
		MaxLogSize:    6000, // Characters (Windows cmd buffer safety)
		RequestTimeout: 30 * time.Second,
		BackupEnabled: true,
		BackupSuffix:  ".sentinel.bak",
		TempDir:       tempDir,
		CacheDir:      cacheDir,
	}
}

// EnsureDirectories creates required directories if they don't exist
func (c *Config) EnsureDirectories() error {
	dirs := []string{c.TempDir, c.CacheDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	if c.MaxLogSize <= 0 {
		return fmt.Errorf("MaxLogSize must be positive")
	}
	if c.RequestTimeout <= 0 {
		return fmt.Errorf("RequestTimeout must be positive")
	}
	return nil
}
