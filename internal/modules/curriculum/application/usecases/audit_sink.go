// Package usecases contains the application-layer use cases for the
// curriculum bounded context.
package usecases

import "context"

// AuditSink is the narrow port the curriculum use cases use to emit
// audit events. The platform AuditLogger (*logging.AuditLogger)
// satisfies this interface structurally, keeping use-case tests free
// of the concrete logger and its side effects.
//
// Defined in this package per the Clean Architecture gate:
// repository-style and audit-style ports live in the consumer
// (use-case) package, not in domain.
//
// Mirrors the assignments-module AuditSink shape exactly so the
// adapter wiring in main.go can satisfy both ports with a single
// concrete logger.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}

// auditResource is the constant resource string every curriculum
// audit event carries. Centralising it here prevents typos from
// drifting between use cases (e.g. "curriculum" vs "curricula").
const auditResource = "curriculum"

// emitAudit dispatches an audit event for the curriculum bounded
// context. A nil sink is treated as a successful no-op so use
// cases never need to sprinkle nil checks at every call site
// (v0.116.0 Create / Update used per-method guards that this
// helper now centralises — N=5 trigger reached with v0.117.0
// Submit / Approve / Reject).
//
// The caller supplies the full action string ("curriculum.created"
// / "curriculum.update_denied" / etc.) and the field map. The
// resource argument is fixed to auditResource so a typo can't
// drift the event into the wrong stream.
func emitAudit(sink AuditSink, ctx context.Context, action string, fields map[string]any) {
	if sink == nil {
		return
	}
	sink.LogAuditEvent(ctx, action, auditResource, fields)
}
