package usecases

import (
	"context"
	"maps"
)

// AuditSink is the narrow port the messaging use case uses to emit
// audit events. *logging.AuditLogger satisfies this structurally;
// tests substitute a recording fake. Defined next to the use case
// per Clean Architecture: repository-style interfaces live in the
// consumer package, not in domain.
//
// Mirror к the assignments package precedent
// (internal/modules/assignments/application/usecases/audit_sink.go):
// pattern is established, copy the shape.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}

// emitMessagingAudit is the package-private helper every messaging
// mutating method calls to emit an audit event. It enriches the
// caller-supplied fields with the actor user id (forensic invariant:
// every event carries who triggered it) and dispatches via AuditSink
// with the caller-supplied resource — distinct values per scope
// (`conversation` vs `message`) so the read API can filter the two
// streams separately.
//
// Nil sink is treated as a successful no-op so existing test setups
// that omit the dependency (and the v0.131.1 backwards-compat
// constructor path) do not need nil checks at every call site.
//
// actorID == 0 (the absent-actor sentinel) skips the actor_user_id
// enrichment so a misconfigured call site never injects a false-zero
// actor into the JSONB payload — same shape as the integration
// helper for defense-in-depth (reviewer Tier 2 #1, v0.131.1).
func emitMessagingAudit(sink AuditSink, ctx context.Context, actorID int64, action, resource string, fields map[string]any) {
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
