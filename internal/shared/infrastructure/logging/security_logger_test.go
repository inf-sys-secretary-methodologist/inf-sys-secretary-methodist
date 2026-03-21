package logging

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSecurityLogger(t *testing.T) {
	logger := NewLogger("debug")
	sl := NewSecurityLogger(logger)
	assert.NotNil(t, sl)
}

func TestSecurityLogger_LogSecurityEvent(t *testing.T) {
	logger := NewLogger("debug")
	sl := NewSecurityLogger(logger)

	tests := []struct {
		name  string
		event SecurityEvent
	}{
		{"login success", EventLoginSuccess},
		{"login failed", EventLoginFailed},
		{"registration success", EventRegistrationSuccess},
		{"registration failed", EventRegistrationFailed},
		{"token refresh success", EventTokenRefreshSuccess},
		{"token refresh failed", EventTokenRefreshFailed},
		{"token validation failed", EventTokenValidationFailed},
		{"unauthorized access", EventUnauthorizedAccess},
		{"rate limit exceeded", EventRateLimitExceeded},
		{"account locked", EventAccountLocked},
		{"password changed", EventPasswordChanged},
		{"permission denied", EventPermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl.LogSecurityEvent(context.Background(), tt.event, map[string]interface{}{
				"test": true,
			})
		})
	}
}

func TestSecurityLogger_LogSecurityEvent_WithContext(t *testing.T) {
	logger := NewLogger("debug")
	sl := NewSecurityLogger(logger)

	ctx := context.WithValue(context.Background(), ContextKeyCorrelationID, "corr-123")
	ctx = context.WithValue(ctx, ContextKeyUserID, int64(42))
	ctx = context.WithValue(ctx, ContextKeyIPAddress, "192.168.1.1")
	ctx = context.WithValue(ctx, ContextKeyUserAgent, "Test Agent")

	sl.LogSecurityEvent(ctx, EventLoginSuccess, map[string]interface{}{
		"extra": "data",
	})
}

func TestSecurityLogger_LogLoginAttempt(t *testing.T) {
	logger := NewLogger("debug")
	sl := NewSecurityLogger(logger)

	sl.LogLoginAttempt(context.Background(), "user@test.com", true, "")
	sl.LogLoginAttempt(context.Background(), "user@test.com", false, "invalid password")
}

func TestSecurityLogger_LogTokenOperation(t *testing.T) {
	logger := NewLogger("debug")
	sl := NewSecurityLogger(logger)

	sl.LogTokenOperation(context.Background(), "refresh", true, 1)
	sl.LogTokenOperation(context.Background(), "refresh", false, 1)
	sl.LogTokenOperation(context.Background(), "validate", false, 1)
	sl.LogTokenOperation(context.Background(), "validate", true, 1) // Should return early
}

func TestSecurityLogger_LogRegistration(t *testing.T) {
	logger := NewLogger("debug")
	sl := NewSecurityLogger(logger)

	sl.LogRegistration(context.Background(), "test@test.com", "admin", true, "")
	sl.LogRegistration(context.Background(), "test@test.com", "admin", false, "email exists")
}

func TestSecurityLogger_LogRateLimitExceeded(t *testing.T) {
	logger := NewLogger("debug")
	sl := NewSecurityLogger(logger)

	sl.LogRateLimitExceeded(context.Background(), "/api/login")
}

func TestSecurityLogger_LogPermissionDenied(t *testing.T) {
	logger := NewLogger("debug")
	sl := NewSecurityLogger(logger)

	sl.LogPermissionDenied(context.Background(), "/api/users", "admin", "user")
}

func TestNewAuditLogger(t *testing.T) {
	logger := NewLogger("debug")
	al := NewAuditLogger(logger)
	assert.NotNil(t, al)
}

func TestAuditLogger_LogAuditEvent(t *testing.T) {
	logger := NewLogger("debug")
	al := NewAuditLogger(logger)

	al.LogAuditEvent(context.Background(), "create", "user", map[string]interface{}{
		"user_id": 1,
	})
}

func TestAuditLogger_LogAuditEvent_WithContext(t *testing.T) {
	logger := NewLogger("debug")
	al := NewAuditLogger(logger)

	ctx := context.WithValue(context.Background(), ContextKeyCorrelationID, "corr-456")
	ctx = context.WithValue(ctx, ContextKeyUserID, int64(10))
	ctx = context.WithValue(ctx, ContextKeyIPAddress, "10.0.0.1")

	al.LogAuditEvent(ctx, "update", "document", map[string]interface{}{
		"doc_id": 42,
	})
}

func TestNewPerformanceLogger(t *testing.T) {
	logger := NewLogger("debug")
	pl := NewPerformanceLogger(logger)
	assert.NotNil(t, pl)
}

func TestPerformanceLogger_LogDatabaseQuery(t *testing.T) {
	logger := NewLogger("debug")
	pl := NewPerformanceLogger(logger)

	// Fast query
	pl.LogDatabaseQuery(context.Background(), "SELECT * FROM users", 10*time.Millisecond, 5) // 10ms

	// Slow query
	pl.LogDatabaseQuery(context.Background(), "SELECT * FROM users", 200*time.Millisecond, 100) // 200ms
}

func TestPerformanceLogger_LogDatabaseQuery_WithContext(t *testing.T) {
	logger := NewLogger("debug")
	pl := NewPerformanceLogger(logger)

	ctx := context.WithValue(context.Background(), ContextKeyCorrelationID, "corr-789")
	pl.LogDatabaseQuery(ctx, "INSERT INTO users", 5*time.Millisecond, 1)
}

func TestPerformanceLogger_LogCacheOperation(t *testing.T) {
	logger := NewLogger("debug")
	pl := NewPerformanceLogger(logger)

	pl.LogCacheOperation(context.Background(), "get", "user:1", true)
	pl.LogCacheOperation(context.Background(), "get", "user:2", false)
}

func TestPerformanceLogger_LogCacheOperation_WithContext(t *testing.T) {
	logger := NewLogger("debug")
	pl := NewPerformanceLogger(logger)

	ctx := context.WithValue(context.Background(), ContextKeyCorrelationID, "corr-abc")
	pl.LogCacheOperation(ctx, "set", "user:1", true)
}

func TestPerformanceLogger_LogHTTPRequest(t *testing.T) {
	logger := NewLogger("debug")
	pl := NewPerformanceLogger(logger)

	// Fast request
	pl.LogHTTPRequest(context.Background(), "GET", "/api/users", 200, 50*time.Millisecond) // 50ms

	// Slow request
	pl.LogHTTPRequest(context.Background(), "POST", "/api/heavy", 200, 1000*time.Millisecond) // 1000ms
}

func TestPerformanceLogger_LogHTTPRequest_WithContext(t *testing.T) {
	logger := NewLogger("debug")
	pl := NewPerformanceLogger(logger)

	ctx := context.WithValue(context.Background(), ContextKeyCorrelationID, "corr-def")
	pl.LogHTTPRequest(ctx, "GET", "/api/test", 404, 20*time.Millisecond)
}

func TestSecurityEventConstants(t *testing.T) {
	assert.Equal(t, SecurityEvent("login_success"), EventLoginSuccess)
	assert.Equal(t, SecurityEvent("login_failed"), EventLoginFailed)
	assert.Equal(t, SecurityEvent("registration_success"), EventRegistrationSuccess)
	assert.Equal(t, SecurityEvent("registration_failed"), EventRegistrationFailed)
	assert.Equal(t, SecurityEvent("token_refresh_success"), EventTokenRefreshSuccess)
	assert.Equal(t, SecurityEvent("token_refresh_failed"), EventTokenRefreshFailed)
	assert.Equal(t, SecurityEvent("token_validation_failed"), EventTokenValidationFailed)
	assert.Equal(t, SecurityEvent("unauthorized_access"), EventUnauthorizedAccess)
	assert.Equal(t, SecurityEvent("rate_limit_exceeded"), EventRateLimitExceeded)
	assert.Equal(t, SecurityEvent("account_locked"), EventAccountLocked)
	assert.Equal(t, SecurityEvent("password_changed"), EventPasswordChanged)
	assert.Equal(t, SecurityEvent("permission_denied"), EventPermissionDenied)
}
