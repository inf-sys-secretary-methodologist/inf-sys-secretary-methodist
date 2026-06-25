// Package usecases contains the application-layer use cases for the
// student_debts bounded context. Use cases orchestrate the StudentDebt
// aggregate (domain), the StudentDebtRepository port (this package),
// and the cross-cutting AuditSink / Notifier ports (this package).
package usecases

import "context"

// AuditSink is the narrow port the student_debts use cases use to emit
// forensic audit events. The platform AuditLogger satisfies it
// structurally, keeping use-case tests free of the concrete logger.
//
// Defined in this package per the Clean Architecture gate: audit-style
// ports live in the consumer (use-case) package, not in domain. The
// shape mirrors work_program / curriculum exactly so one concrete logger
// in main.go can satisfy every module's AuditSink.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}

// auditResource is the constant resource string every student_debts
// audit event carries. Centralizing it prevents "student_debts" vs
// "studentdebts" drift between use cases.
const auditResource = "student_debts"

// denialFields composes the canonical
// {actor_user_id, student_debt_id, reason} field shape every *_denied
// audit event carries, so operators can grep one column across denials.
func denialFields(actorID, debtID int64, reason string) map[string]any {
	return map[string]any{
		"actor_user_id":   actorID,
		"student_debt_id": debtID,
		"reason":          reason,
	}
}

// emitAudit dispatches an audit event for the student_debts context. A
// nil sink is a successful no-op so use cases never sprinkle nil checks.
// The resource is fixed to auditResource so a typo can't drift the event
// into the wrong stream.
func emitAudit(sink AuditSink, ctx context.Context, action string, fields map[string]any) {
	if sink == nil {
		return
	}
	sink.LogAuditEvent(ctx, action, auditResource, fields)
}
