package usecases

import "context"

// AuditSink is the narrow port the use cases in this package use to emit
// audit events. The platform AuditLogger (*logging.AuditLogger) satisfies
// this interface structurally, keeping use-case tests free of the concrete
// logger and its side effects. Defined next to the other narrow ports
// (SaveGradeNotifier, ReturnSubmissionNotifier) per the Clean Architecture
// gate: repository-style interfaces live in the consumer package, not in
// domain.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}
