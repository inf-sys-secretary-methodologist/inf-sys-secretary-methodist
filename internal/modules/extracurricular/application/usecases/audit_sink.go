package usecases

import "context"

// AuditSink is the narrow port the extracurricular use cases use to
// emit audit events. *logging.AuditLogger satisfies it structurally —
// no concrete dependency leaks into the use-case package per CLAUDE.md
// Clean Architecture gate.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}

// auditResource is the canonical resource string for every event in
// the extracurricular bounded context.
const auditResource = "extracurricular_event"

// Pair 5 RED anchor — keeps helpers referenced until GREEN switches
// stubs к real call sites. Removed in GREEN.
//
//nolint:gochecknoglobals,unused // anchor only
var _ = []any{auditResource, emitAudit, actionFields, denialFields}

// emitAudit dispatches an audit event. Nil sink → no-op so each
// use-case site stays free of `if uc.audit != nil` clutter.
func emitAudit(sink AuditSink, ctx context.Context, action string, fields map[string]any) {
	if sink == nil {
		return
	}
	sink.LogAuditEvent(ctx, action, auditResource, fields)
}

// actionFields returns the canonical happy-path field shape for
// successful actions. event_id may be 0 on initial Create — the
// fresh-row id is populated post-Save.
func actionFields(actorID, eventID int64) map[string]any {
	return map[string]any{
		"actor_user_id": actorID,
		"event_id":      eventID,
	}
}

// denialFields composes the canonical denial-record shape. Code is a
// machine-readable category (e.g. "not_found", "forbidden",
// "version_conflict", "invalid") matched in the *_denied test
// assertions so renames surface immediately.
func denialFields(actorID, eventID int64, reason, code string) map[string]any {
	return map[string]any{
		"actor_user_id": actorID,
		"event_id":      eventID,
		"reason":        reason,
		"code":          code,
	}
}
