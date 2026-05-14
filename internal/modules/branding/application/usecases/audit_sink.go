package usecases

import "context"

// AuditSink is the narrow port the branding use cases call when
// emitting forensic events. Mirror к feedback_audit_emitter_narrow_
// interface — define the interface in the consumer package matching
// the methods used; concrete *logging.AuditLogger satisfies it
// structurally; tests substitute a spy fake.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}
