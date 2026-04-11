// Package logging provides structured logging utilities.
package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
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

// ContextLogger wraps Logger and auto-injects trace_id and span_id from context.
type ContextLogger struct {
	logger  *Logger
	traceID string
	spanID  string
}

// WithContext returns a ContextLogger that includes trace_id/span_id from the span in ctx.
// If no active span exists, fields are omitted.
func (l *Logger) WithContext(ctx context.Context) *ContextLogger {
	cl := &ContextLogger{logger: l}
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		cl.traceID = span.SpanContext().TraceID().String()
		cl.spanID = span.SpanContext().SpanID().String()
	}
	return cl
}

func (cl *ContextLogger) enrichFields(fields map[string]interface{}) map[string]interface{} {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if cl.traceID != "" {
		fields["trace_id"] = cl.traceID
		fields["span_id"] = cl.spanID
	}
	return fields
}

// Debug logs debug message with trace context.
func (cl *ContextLogger) Debug(message string, fields map[string]interface{}) {
	cl.logger.Debug(message, cl.enrichFields(fields))
}

// Info logs info message with trace context.
func (cl *ContextLogger) Info(message string, fields map[string]interface{}) {
	cl.logger.Info(message, cl.enrichFields(fields))
}

// Warn logs warning message with trace context.
func (cl *ContextLogger) Warn(message string, fields map[string]interface{}) {
	cl.logger.Warn(message, cl.enrichFields(fields))
}

// Error logs error message with trace context.
func (cl *ContextLogger) Error(message string, fields map[string]interface{}) {
	cl.logger.Error(message, cl.enrichFields(fields))
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
