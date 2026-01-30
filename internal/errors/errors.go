package errors

import (
	"fmt"
)

// ErrorType represents different categories of errors
type ErrorType int

const (
	ErrTypeUnknown ErrorType = iota
	ErrTypeGitHub           // GitHub API errors
	ErrTypeCopilot          // Copilot CLI errors
	ErrTypeFilesystem       // File operations errors
	ErrTypeValidation       // Data validation errors
	ErrTypeNetwork          // Network connectivity errors
	ErrTypeAuth             // Authentication errors
)

// SentinelError is a custom error with additional context
type SentinelError struct {
	Type    ErrorType
	Op      string // Operation being performed
	Path    string // File path if applicable
	Err     error  // Underlying error
	Message string // User-friendly message
}

func (e *SentinelError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s (%s): %v", e.Op, e.Message, e.Path, e.Err)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

func (e *SentinelError) Unwrap() error {
	return e.Err
}

// New creates a new SentinelError
func New(errType ErrorType, op, message string, err error) *SentinelError {
	return &SentinelError{
		Type:    errType,
		Op:      op,
		Message: message,
		Err:     err,
	}
}

// WithPath adds a file path context to the error
func (e *SentinelError) WithPath(path string) *SentinelError {
	e.Path = path
	return e
}

// Predefined error constructors for common scenarios
func GitHubAPIError(op string, err error) *SentinelError {
	return New(ErrTypeGitHub, op, "GitHub API request failed", err)
}

func CopilotError(op string, err error) *SentinelError {
	return New(ErrTypeCopilot, op, "GitHub Copilot CLI failed", err)
}

func FilesystemError(op, path string, err error) *SentinelError {
	return New(ErrTypeFilesystem, op, "filesystem operation failed", err).WithPath(path)
}

func ValidationError(op, message string) *SentinelError {
	return New(ErrTypeValidation, op, message, nil)
}

func NetworkError(op string, err error) *SentinelError {
	return New(ErrTypeNetwork, op, "network request failed", err)
}

func AuthError(op string, err error) *SentinelError {
	return New(ErrTypeAuth, op, "authentication failed", err)
}
