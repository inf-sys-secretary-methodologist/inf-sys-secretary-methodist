// Package usecases contains the application-layer use cases for the
// work_program bounded context. Use cases orchestrate the WorkProgram
// aggregate (domain), the WorkProgramRepository port (this package),
// and the AuditSink port (this file).
package usecases

import "context"

// AuditSink is the narrow port the work_program use cases use to emit
// forensic audit events. The platform AuditLogger (*logging.AuditLogger)
// satisfies this interface structurally, keeping use-case tests free of
// the concrete logger and its side effects.
//
// Defined in this package per the Clean Architecture gate:
// audit-style ports live in the consumer (use-case) package, not in
// domain. Mirrors the curriculum/assignments AuditSink shape exactly
// so the adapter wiring in main.go can satisfy multiple module ports
// with a single concrete logger.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}

// auditResource is the constant resource string every work_program
// audit event carries. Centralizing it here prevents typos from
// drifting between use cases (e.g. "work_program" vs "workprogram").
const auditResource = "work_program"

// denialFields composes the canonical
// {actor_user_id, work_program_id, reason, specialty_code} field shape
// that every *_denied audit event carries. Centralizing the shape
// keeps the forensic record consistent across use cases so operators
// can grep one column name and see every denial of a given kind.
//
// specialty_code is the most identifying piece of business metadata
// for a WorkProgram (cohort-scoped — see ADR-018 ADR-3), so it plays
// the role that `code` plays in the curriculum module.
func denialFields(actorID, workProgramID int64, reason, specialtyCode string) map[string]any {
	return map[string]any{
		"actor_user_id":   actorID,
		"work_program_id": workProgramID,
		"reason":          reason,
		"specialty_code":  specialtyCode,
	}
}

// emitAudit dispatches an audit event for the work_program bounded
// context. A nil sink is treated as a successful no-op so use cases
// never need to sprinkle nil checks at every call site.
//
// The caller supplies the full action string ("work_program.created"
// / "work_program.submit_denied" / etc.) and the field map. The
// resource argument is fixed to auditResource so a typo can't drift
// the event into the wrong stream.
func emitAudit(sink AuditSink, ctx context.Context, action string, fields map[string]any) {
	if sink == nil {
		return
	}
	sink.LogAuditEvent(ctx, action, auditResource, fields)
}
