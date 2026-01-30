package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents logging levels
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger provides structured logging
type Logger struct {
	level  Level
	output io.Writer
	prefix string
}

// New creates a new logger
func New(level Level, output io.Writer) *Logger {
	if output == nil {
		output = os.Stderr
	}
	return &Logger{
		level:  level,
		output: output,
		prefix: "sentinel",
	}
}

// Default creates a logger with default settings
func Default() *Logger {
	return New(LevelInfo, os.Stderr)
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	var levelStr string
	switch level {
	case LevelDebug:
		levelStr = "DEBUG"
	case LevelInfo:
		levelStr = "INFO"
	case LevelWarn:
		levelStr = "WARN"
	case LevelError:
		levelStr = "ERROR"
	}

	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.output, "[%s] %s: %s\n", timestamp, levelStr, message)
}

// Debug logs debug messages
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs informational messages
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs warning messages
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs error messages
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}
