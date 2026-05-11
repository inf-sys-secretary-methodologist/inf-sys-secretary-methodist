package usecases

import (
	"context"
	"maps"
)

// AuditSink is the narrow port both SyncUseCase and ConflictUseCase
// use to emit forensic audit events on 1C synchronization outcomes
// and conflict-resolution decisions. *logging.AuditLogger satisfies
// this structurally; tests substitute a recording fake. Defined next
// to the use cases per the Clean Architecture gate.
//
// Mirror к the assignments + messaging precedent — same shape, same
// helper-on-package level. Two separate consumers in this package
// (sync + conflict) is enough to extract on N=2 because the helper
// also adds the unified actor_user_id semantics required by the
// audit_logs JSONB schema.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}

// emitIntegrationAudit dispatches one audit event with the
// caller-supplied fields. Nil sink is a no-op so existing test
// setups remain backward-compatible.
//
// actorID == 0 (the absent-actor sentinel) skips the actor_user_id
// enrichment — used by background sync events that have no
// authenticated initiator on the application side. The platform
// AuditLogger still copies actor_user_id from the request context
// into the row's typed column when middleware promoted it; the
// in-fields enrichment here is only for human-readable inspection
// of the JSONB payload.
func emitIntegrationAudit(sink AuditSink, ctx context.Context, actorID int64, action, resource string, fields map[string]any) {
	if sink == nil {
		return
	}
	enriched := map[string]any{}
	if actorID != 0 {
		enriched["actor_user_id"] = actorID
	}
	maps.Copy(enriched, fields)
	sink.LogAuditEvent(ctx, action, resource, enriched)
}
