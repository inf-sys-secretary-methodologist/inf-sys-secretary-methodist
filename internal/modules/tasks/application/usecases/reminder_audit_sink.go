package usecases

import "context"

// AuditSink is the narrow port the task reminder use cases call
// when emitting forensic events. Mirror к feedback_audit_emitter_
// narrow_interface — define the interface in the consumer package
// matching only the methods used; concrete *logging.AuditLogger
// satisfies it structurally so wiring stays one-line in main.go;
// tests substitute a spy fake.
//
// Separate from the existing project/task usecase audit dependency
// в this package (which uses *logging.AuditLogger directly) so the
// reminder use cases can be tested without dragging the heavier
// logging package into their unit tests.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}
