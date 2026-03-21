package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		level    string
		expected LogLevel
	}{
		{"debug", DEBUG},
		{"info", INFO},
		{"warn", WARN},
		{"error", ERROR},
		{"unknown", INFO},
		{"", INFO},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			l := NewLogger(tt.level)
			assert.NotNil(t, l)
			assert.Equal(t, tt.expected, l.level)
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	l := NewLogger("debug")
	// Should not panic
	l.Debug("test message", map[string]interface{}{"key": "value"})
}

func TestLogger_Debug_Filtered(t *testing.T) {
	l := NewLogger("info")
	// Debug messages should be filtered out at info level
	l.Debug("should not appear", nil)
}

func TestLogger_Info(t *testing.T) {
	l := NewLogger("info")
	l.Info("test info", map[string]interface{}{"count": 42})
}

func TestLogger_Info_Filtered(t *testing.T) {
	l := NewLogger("warn")
	l.Info("should not appear", nil)
}

func TestLogger_Warn(t *testing.T) {
	l := NewLogger("warn")
	l.Warn("test warning", map[string]interface{}{"warning": true})
}

func TestLogger_Warn_Filtered(t *testing.T) {
	l := NewLogger("error")
	l.Warn("should not appear", nil)
}

func TestLogger_Error(t *testing.T) {
	l := NewLogger("error")
	l.Error("test error", map[string]interface{}{"error": "something bad"})
}

func TestLogger_NilFields(t *testing.T) {
	l := NewLogger("debug")
	l.Debug("message with nil fields", nil)
	l.Info("message with nil fields", nil)
	l.Warn("message with nil fields", nil)
	l.Error("message with nil fields", nil)
}

func TestLogger_LogLevelConstants(t *testing.T) {
	assert.Equal(t, LogLevel(0), DEBUG)
	assert.Equal(t, LogLevel(1), INFO)
	assert.Equal(t, LogLevel(2), WARN)
	assert.Equal(t, LogLevel(3), ERROR)
}

func TestParseLogLevel(t *testing.T) {
	assert.Equal(t, DEBUG, parseLogLevel("debug"))
	assert.Equal(t, INFO, parseLogLevel("info"))
	assert.Equal(t, WARN, parseLogLevel("warn"))
	assert.Equal(t, ERROR, parseLogLevel("error"))
	assert.Equal(t, INFO, parseLogLevel("invalid"))
}
