package usecases

import (
	"context"
	"maps"
)

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

// emitAudit is the package-private helper every assignments use case calls
// to emit an audit event. It enriches the caller-supplied fields with the
// actor user id (forensic trail invariant: every event carries who
// triggered it) and dispatches via AuditSink with resource="assignment".
//
// Centralised here because three call sites (SaveGrade, ReturnSubmission,
// ResubmitSubmission) need the same shape; the v0.111.0 review flagged
// the duplicated method-pair as "extract on N=3", and this is N=3.
//
// Nil sink is treated as a successful no-op so tests can omit the
// dependency without sprinkling nil checks at every call site.
func emitAudit(sink AuditSink, ctx context.Context, actorID int64, action string, fields map[string]any) {
	if sink == nil {
		return
	}
	enriched := map[string]any{"actor_user_id": actorID}
	maps.Copy(enriched, fields)
	sink.LogAuditEvent(ctx, action, "assignment", enriched)
}
