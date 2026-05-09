package usecases

import "context"

// auditSectionResource is the constant resource string every Section
// audit event carries. Distinct from auditResource ("curriculum") so
// operators can grep one stream and see only section events without
// conflating them with curriculum lifecycle events.
const auditSectionResource = "curriculum_section"

// sectionDenialFields composes the canonical {actor_user_id, section_id,
// curriculum_id, reason} field shape every section *_denied event
// carries. Mirror к denialFields(...) for curricula but with the
// section-specific axis (section_id) instead of code.
//
// curriculumID may be 0 in two operational scenarios:
//  1. CreateSection denial when the curriculum lookup itself failed
//     (the user supplied a non-existent curriculum id) — the operator
//     reads in.CurriculumID from the request log to see which.
//  2. Sentinel-not-set defenses inside disambiguation paths.
//
// sectionID may be 0 for Create denials before the row was assigned
// an id by the database.
func sectionDenialFields(actorID, sectionID, curriculumID int64, reason string) map[string]any {
	return map[string]any{
		"actor_user_id": actorID,
		"section_id":    sectionID,
		"curriculum_id": curriculumID,
		"reason":        reason,
	}
}

// emitSectionAudit dispatches an audit event for the section resource.
// A nil sink is treated as a successful no-op (mirror к emitAudit
// for curricula). The resource string is fixed to auditSectionResource
// so a typo cannot drift the event into the curriculum stream.
func emitSectionAudit(sink AuditSink, ctx context.Context, action string, fields map[string]any) {
	if sink == nil {
		return
	}
	sink.LogAuditEvent(ctx, action, auditSectionResource, fields)
}
