package log

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
)

// Logger provides structured logging for agent actions.
type Logger struct {
	slog    *slog.Logger
	verbose bool
}

// New creates a new Logger.
func New(verbose bool) *Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return &Logger{
		slog:    slog.New(handler),
		verbose: verbose,
	}
}

// AgentAction logs an agent performing an action.
func (l *Logger) AgentAction(a *agent.Agent, action, detail string) {
	l.slog.Info("agent action",
		"agent", a.ShortName(),
		"role", a.Role,
		"action", action,
		"detail", detail,
	)
}

// ScenarioStart logs the start of a scenario.
func (l *Logger) ScenarioStart(name string) {
	l.slog.Info("scenario started",
		"scenario", name,
		"time", time.Now().Format("15:04:05"),
	)
}

// ScenarioEnd logs the end of a scenario.
func (l *Logger) ScenarioEnd(name string, duration time.Duration) {
	l.slog.Info("scenario completed",
		"scenario", name,
		"duration", duration.Round(time.Second).String(),
	)
}

// StepStart logs the start of a scenario step.
func (l *Logger) StepStart(stepName, agentName string) {
	l.slog.Info("step",
		"step", stepName,
		"agent", agentName,
	)
}

// StepError logs a step failure.
func (l *Logger) StepError(stepName string, err error) {
	l.slog.Error("step failed",
		"step", stepName,
		"error", err.Error(),
	)
}

// Debug logs a debug message (only if verbose).
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

// Error logs an error.
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

// Fatal logs an error and exits.
func (l *Logger) Fatal(msg string, args ...any) {
	l.slog.Error(msg, args...)
	fmt.Fprintf(os.Stderr, "FATAL: %s\n", msg)
	os.Exit(1)
}
