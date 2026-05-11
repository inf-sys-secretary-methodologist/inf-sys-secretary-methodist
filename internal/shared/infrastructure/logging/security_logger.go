package logging

import (
	"context"
	"time"
)

// ContextKey is a custom type for context keys to avoid SA1029.
type ContextKey string

const (
	// ContextKeyCorrelationID is the context key for correlation ID.
	ContextKeyCorrelationID ContextKey = "correlation_id"
	// ContextKeyUserID is the context key for user ID.
	ContextKeyUserID ContextKey = "user_id"
	// ContextKeyIPAddress is the context key for IP address.
	ContextKeyIPAddress ContextKey = "ip_address"
	// ContextKeyUserAgent is the context key for user agent.
	ContextKeyUserAgent ContextKey = "user_agent"
)

// SecurityEvent represents a security-related event
type SecurityEvent string

const (
	// EventLoginSuccess indicates a successful login attempt.
	EventLoginSuccess SecurityEvent = "login_success"
	// EventLoginFailed indicates a failed login attempt.
	EventLoginFailed SecurityEvent = "login_failed"
	// EventRegistrationSuccess indicates a successful registration.
	EventRegistrationSuccess SecurityEvent = "registration_success"
	// EventRegistrationFailed indicates a failed registration.
	EventRegistrationFailed SecurityEvent = "registration_failed"
	// EventTokenRefreshSuccess indicates a successful token refresh.
	EventTokenRefreshSuccess SecurityEvent = "token_refresh_success"
	// EventTokenRefreshFailed indicates a failed token refresh.
	EventTokenRefreshFailed SecurityEvent = "token_refresh_failed"
	// EventTokenValidationFailed indicates a failed token validation.
	EventTokenValidationFailed SecurityEvent = "token_validation_failed"
	// EventUnauthorizedAccess indicates an unauthorized access attempt.
	EventUnauthorizedAccess SecurityEvent = "unauthorized_access"
	// EventRateLimitExceeded indicates that rate limit was exceeded.
	EventRateLimitExceeded SecurityEvent = "rate_limit_exceeded"
	// EventAccountLocked indicates that an account was locked.
	EventAccountLocked SecurityEvent = "account_locked"
	// EventPasswordChanged indicates that a password was changed.
	EventPasswordChanged SecurityEvent = "password_changed"
	// EventPermissionDenied indicates that permission was denied.
	EventPermissionDenied SecurityEvent = "permission_denied"
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
		"event_type": string(event),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"category":   "security",
	}

	// Extract correlation ID from context if present
	if correlationID := ctx.Value(ContextKeyCorrelationID); correlationID != nil {
		enrichedFields["correlation_id"] = correlationID
	}

	// Extract user ID from context if present
	if userID := ctx.Value(ContextKeyUserID); userID != nil {
		enrichedFields["user_id"] = userID
	}

	// Extract IP address from context if present
	if ipAddr := ctx.Value(ContextKeyIPAddress); ipAddr != nil {
		enrichedFields["ip_address"] = ipAddr
	}

	// Extract user agent from context if present
	if userAgent := ctx.Value(ContextKeyUserAgent); userAgent != nil {
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

// AuditLogger provides detailed audit trail for compliance.
//
// The structured stdout emit is the original v0.99 surface. v0.130.0
// adds an optional AuditLogWriter (set via WithRepository) for
// persisting every event to audit_logs (ADR-2 sync write on an
// independent connection; ADR-3 failure is logged and not propagated).
// When writer is nil the logger behaves exactly as before — backwards
// compatible with every existing call site (ADR-7).
type AuditLogger struct {
	logger *Logger
	writer AuditLogWriter // optional; nil → log-only behavior
}

// NewAuditLogger creates a new audit logger emitting to structured log
// only. Use WithRepository to add audit_logs persistence.
func NewAuditLogger(logger *Logger) *AuditLogger {
	return &AuditLogger{logger: logger}
}

// WithRepository attaches an AuditLogWriter so every subsequent
// LogAuditEvent persists to audit_logs in addition to the structured
// stdout emit. Returns the receiver so callers can chain
// NewAuditLogger(l).WithRepository(repo) at DI time.
func (al *AuditLogger) WithRepository(writer AuditLogWriter) *AuditLogger {
	al.writer = writer
	return al
}

// LogAuditEvent emits an audit event to the structured log and, if a
// repository was attached via WithRepository, also persists it to the
// audit_logs table. Per ADR-3 a writer failure does NOT propagate to
// the caller — the original structured log line is still emitted, plus
// a separate error-level log line records the persist failure for
// post-mortem.
func (al *AuditLogger) LogAuditEvent(ctx context.Context, action string, resource string, fields map[string]interface{}) {
	enrichedFields := map[string]interface{}{
		"action":    action,
		"resource":  resource,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"category":  "audit",
	}

	// Extract context values
	if correlationID := ctx.Value(ContextKeyCorrelationID); correlationID != nil {
		enrichedFields["correlation_id"] = correlationID
	}
	if userID := ctx.Value(ContextKeyUserID); userID != nil {
		enrichedFields["actor_user_id"] = userID
	}
	if ipAddr := ctx.Value(ContextKeyIPAddress); ipAddr != nil {
		enrichedFields["actor_ip"] = ipAddr
	}

	// Merge provided fields
	for k, v := range fields {
		enrichedFields[k] = v
	}

	al.logger.Info("Audit event", enrichedFields)

	if al.writer != nil {
		al.persist(ctx, action, resource, fields)
	}
}

// persist builds the AuditLog row from ctx + caller args and writes it.
// Failure is captured and logged at error level — never propagated to
// the caller (ADR-3 fire-and-forget). Keeping this off the hot path of
// LogAuditEvent body keeps the structured-log emission, which is the
// most important forensic record, ahead of any DB latency.
func (al *AuditLogger) persist(ctx context.Context, action, resource string, fields map[string]interface{}) {
	rowFields := fields
	if rowFields == nil {
		rowFields = map[string]interface{}{}
	}
	row := &AuditLog{
		Action:        action,
		Resource:      resource,
		ActorUserID:   extractActorUserID(ctx),
		ActorIP:       extractActorIP(ctx),
		CorrelationID: extractCorrelationID(ctx),
		Fields:        rowFields,
	}
	if err := al.writer.Write(ctx, row); err != nil {
		al.logger.Error("Audit event persistence failed", map[string]interface{}{
			"action":   action,
			"resource": resource,
			"cause":    err.Error(),
		})
	}
}

// extractActorUserID pulls user_id from context. Supports int64 directly
// or any integer type assignable to int64 via a type switch; otherwise
// nil so the row writes SQL NULL.
func extractActorUserID(ctx context.Context) *int64 {
	v := ctx.Value(ContextKeyUserID)
	if v == nil {
		return nil
	}
	if id, ok := v.(int64); ok {
		return &id
	}
	return nil
}

// extractActorIP pulls the source IP from context as a string.
func extractActorIP(ctx context.Context) *string {
	v := ctx.Value(ContextKeyIPAddress)
	if v == nil {
		return nil
	}
	if ip, ok := v.(string); ok && ip != "" {
		return &ip
	}
	return nil
}

// extractCorrelationID pulls the request correlation id from context.
func extractCorrelationID(ctx context.Context) *string {
	v := ctx.Value(ContextKeyCorrelationID)
	if v == nil {
		return nil
	}
	if id, ok := v.(string); ok && id != "" {
		return &id
	}
	return nil
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
func (pl *PerformanceLogger) LogDatabaseQuery(ctx context.Context, _ string, duration time.Duration, rowsAffected int64) {
	fields := map[string]interface{}{
		"query_type":    "database",
		"duration_ms":   duration.Milliseconds(),
		"rows_affected": rowsAffected,
		"category":      "performance",
	}

	if correlationID := ctx.Value(ContextKeyCorrelationID); correlationID != nil {
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

	if correlationID := ctx.Value(ContextKeyCorrelationID); correlationID != nil {
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

	if correlationID := ctx.Value(ContextKeyCorrelationID); correlationID != nil {
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
