package logging

import (
	"context"
	"time"
)

// SecurityEvent represents a security-related event
type SecurityEvent string

const (
	EventLoginSuccess         SecurityEvent = "login_success"
	EventLoginFailed          SecurityEvent = "login_failed"
	EventRegistrationSuccess  SecurityEvent = "registration_success"
	EventRegistrationFailed   SecurityEvent = "registration_failed"
	EventTokenRefreshSuccess  SecurityEvent = "token_refresh_success"
	EventTokenRefreshFailed   SecurityEvent = "token_refresh_failed"
	EventTokenValidationFailed SecurityEvent = "token_validation_failed"
	EventUnauthorizedAccess   SecurityEvent = "unauthorized_access"
	EventRateLimitExceeded    SecurityEvent = "rate_limit_exceeded"
	EventAccountLocked        SecurityEvent = "account_locked"
	EventPasswordChanged      SecurityEvent = "password_changed"
	EventPermissionDenied     SecurityEvent = "permission_denied"
)

// SecurityLogger provides security event logging with audit trail
type SecurityLogger struct {
	logger *Logger
}

// NewSecurityLogger creates a new security logger
func NewSecurityLogger(logger *Logger) *SecurityLogger {
	return &SecurityLogger{logger: logger}
}

// LogSecurityEvent logs a security event with full context
func (sl *SecurityLogger) LogSecurityEvent(ctx context.Context, event SecurityEvent, fields map[string]interface{}) {
	enrichedFields := map[string]interface{}{
		"event_type":  string(event),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"category":    "security",
	}

	// Extract correlation ID from context if present
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		enrichedFields["correlation_id"] = correlationID
	}

	// Extract user ID from context if present
	if userID := ctx.Value("user_id"); userID != nil {
		enrichedFields["user_id"] = userID
	}

	// Extract IP address from context if present
	if ipAddr := ctx.Value("ip_address"); ipAddr != nil {
		enrichedFields["ip_address"] = ipAddr
	}

	// Extract user agent from context if present
	if userAgent := ctx.Value("user_agent"); userAgent != nil {
		enrichedFields["user_agent"] = userAgent
	}

	// Merge provided fields
	for k, v := range fields {
		enrichedFields[k] = v
	}

	// Determine log level based on event type
	switch event {
	case EventLoginFailed, EventTokenValidationFailed, EventUnauthorizedAccess,
		 EventRateLimitExceeded, EventPermissionDenied:
		sl.logger.Warn("Security event detected", enrichedFields)
	case EventAccountLocked:
		sl.logger.Error("Security event detected", enrichedFields)
	default:
		sl.logger.Info("Security event", enrichedFields)
	}
}

// LogLoginAttempt logs a login attempt
func (sl *SecurityLogger) LogLoginAttempt(ctx context.Context, email string, success bool, reason string) {
	event := EventLoginSuccess
	if !success {
		event = EventLoginFailed
	}

	sl.LogSecurityEvent(ctx, event, map[string]interface{}{
		"email":   email,
		"success": success,
		"reason":  reason,
	})
}

// LogTokenOperation logs token-related operations
func (sl *SecurityLogger) LogTokenOperation(ctx context.Context, operation string, success bool, userID int64) {
	var event SecurityEvent
	switch operation {
	case "refresh":
		if success {
			event = EventTokenRefreshSuccess
		} else {
			event = EventTokenRefreshFailed
		}
	case "validate":
		if !success {
			event = EventTokenValidationFailed
		} else {
			return // Don't log successful validations to reduce noise
		}
	}

	sl.LogSecurityEvent(ctx, event, map[string]interface{}{
		"operation": operation,
		"success":   success,
		"user_id":   userID,
	})
}

// LogRegistration logs user registration
func (sl *SecurityLogger) LogRegistration(ctx context.Context, email string, role string, success bool, reason string) {
	event := EventRegistrationSuccess
	if !success {
		event = EventRegistrationFailed
	}

	sl.LogSecurityEvent(ctx, event, map[string]interface{}{
		"email":   email,
		"role":    role,
		"success": success,
		"reason":  reason,
	})
}

// LogRateLimitExceeded logs rate limit violations
func (sl *SecurityLogger) LogRateLimitExceeded(ctx context.Context, endpoint string) {
	sl.LogSecurityEvent(ctx, EventRateLimitExceeded, map[string]interface{}{
		"endpoint": endpoint,
	})
}

// LogPermissionDenied logs authorization failures
func (sl *SecurityLogger) LogPermissionDenied(ctx context.Context, resource string, requiredRole string, userRole string) {
	sl.LogSecurityEvent(ctx, EventPermissionDenied, map[string]interface{}{
		"resource":      resource,
		"required_role": requiredRole,
		"user_role":     userRole,
	})
}

// AuditLogger provides detailed audit trail for compliance
type AuditLogger struct {
	logger *Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *Logger) *AuditLogger {
	return &AuditLogger{logger: logger}
}

// LogAuditEvent logs an audit event
func (al *AuditLogger) LogAuditEvent(ctx context.Context, action string, resource string, fields map[string]interface{}) {
	enrichedFields := map[string]interface{}{
		"action":    action,
		"resource":  resource,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"category":  "audit",
	}

	// Extract context values
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		enrichedFields["correlation_id"] = correlationID
	}
	if userID := ctx.Value("user_id"); userID != nil {
		enrichedFields["actor_user_id"] = userID
	}
	if ipAddr := ctx.Value("ip_address"); ipAddr != nil {
		enrichedFields["actor_ip"] = ipAddr
	}

	// Merge provided fields
	for k, v := range fields {
		enrichedFields[k] = v
	}

	al.logger.Info("Audit event", enrichedFields)
}

// PerformanceLogger logs performance metrics
type PerformanceLogger struct {
	logger *Logger
}

// NewPerformanceLogger creates a new performance logger
func NewPerformanceLogger(logger *Logger) *PerformanceLogger {
	return &PerformanceLogger{logger: logger}
}

// LogDatabaseQuery logs database query performance
func (pl *PerformanceLogger) LogDatabaseQuery(ctx context.Context, query string, duration time.Duration, rowsAffected int64) {
	fields := map[string]interface{}{
		"query_type":    "database",
		"duration_ms":   duration.Milliseconds(),
		"rows_affected": rowsAffected,
		"category":      "performance",
	}

	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		fields["correlation_id"] = correlationID
	}

	// Warn on slow queries (> 100ms)
	if duration.Milliseconds() > 100 {
		fields["slow_query"] = true
		pl.logger.Warn("Slow database query detected", fields)
	} else {
		pl.logger.Debug("Database query executed", fields)
	}
}

// LogCacheOperation logs cache hit/miss
func (pl *PerformanceLogger) LogCacheOperation(ctx context.Context, operation string, key string, hit bool) {
	fields := map[string]interface{}{
		"operation": operation,
		"cache_key": key,
		"cache_hit": hit,
		"category":  "performance",
	}

	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		fields["correlation_id"] = correlationID
	}

	pl.logger.Debug("Cache operation", fields)
}

// LogHTTPRequest logs HTTP request performance
func (pl *PerformanceLogger) LogHTTPRequest(ctx context.Context, method string, path string, statusCode int, duration time.Duration) {
	fields := map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
		"category":    "performance",
	}

	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		fields["correlation_id"] = correlationID
	}

	// Warn on slow requests (> 500ms)
	if duration.Milliseconds() > 500 {
		fields["slow_request"] = true
		pl.logger.Warn("Slow HTTP request detected", fields)
	} else {
		pl.logger.Info("HTTP request completed", fields)
	}
}
