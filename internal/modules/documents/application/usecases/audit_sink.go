// Package usecases contains the application-layer use cases for the
// documents module.
package usecases

import "context"

// AuditSink is the narrow port the documents workflow use cases (Submit/
// Approve/Reject — v0.148.0 #227) use to emit audit events. The platform
// AuditLogger (*logging.AuditLogger) satisfies this interface structurally,
// keeping use-case tests free of the concrete logger.
//
// Defined в этом package per Clean Architecture: audit-style ports live
// в consumer (use-case) package, не в domain. Mirror к curriculum +
// assignments AuditSink shape exactly so a single concrete logger can
// satisfy all three modules' ports.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}

// auditResource is the constant resource string every documents-workflow
// audit event carries.
const auditResource = "document"

// auditFieldDocumentID centralizes the audit payload key so a single
// rename propagates across all workflow use cases.
const auditFieldDocumentID = "document_id"

// denialFields composes the canonical {actor_user_id, document_id,
// reason} field shape that every *_denied audit event carries. Mirror
// к curriculum.denialFields.
func denialFields(actorID, documentID int64, reason string) map[string]any {
	return map[string]any{
		"actor_user_id":      actorID,
		auditFieldDocumentID: documentID,
		"reason":             reason,
	}
}

// emitAudit dispatches an audit event. Nil sink — silent no-op.
func emitAudit(sink AuditSink, ctx context.Context, action string, fields map[string]any) {
	if sink == nil {
		return
	}
	sink.LogAuditEvent(ctx, action, auditResource, fields)
}
