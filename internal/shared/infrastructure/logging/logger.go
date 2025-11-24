// Package logging provides structured logging utilities.
package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents logging level
type LogLevel int

const (
	// DEBUG log level for debug messages.
	DEBUG LogLevel = iota
	// INFO log level for informational messages.
	INFO
	// WARN log level for warning messages.
	WARN
	// ERROR log level for error messages.
	ERROR
)

// Logger provides structured logging
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(level string) *Logger {
	logLevel := parseLogLevel(level)
	return &Logger{
		level:  logLevel,
		logger: log.New(os.Stdout, "", 0),
	}
}

// Debug logs debug message
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	if l.level <= DEBUG {
		l.log("DEBUG", message, fields)
	}
}

// Info logs info message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	if l.level <= INFO {
		l.log("INFO", message, fields)
	}
}

// Warn logs warning message
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	if l.level <= WARN {
		l.log("WARN", message, fields)
	}
}

// Error logs error message
func (l *Logger) Error(message string, fields map[string]interface{}) {
	if l.level <= ERROR {
		l.log("ERROR", message, fields)
	}
}

func (l *Logger) log(level, message string, fields map[string]interface{}) {
	entry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level,
		"message":   message,
	}

	for k, v := range fields {
		entry[k] = v
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}

	l.logger.Println(string(jsonBytes))
}

func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}
