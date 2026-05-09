package usecases

import "context"

// auditDisciplineItemResource is the constant resource string every
// DisciplineItem audit event carries. Distinct from auditResource
// ("curriculum") и auditSectionResource ("curriculum_section") so
// operators can grep one stream and see only discipline-item events
// without conflating с parent-aggregate lifecycle events.
const auditDisciplineItemResource = "curriculum_section_item"

// disciplineItemDenialFields composes the canonical {actor_user_id,
// item_id, section_id, curriculum_id, reason} field shape. Mirror к
// sectionDenialFields but adds item_id axis (operators trace failed
// item edit independently от section edit).
//
// Zero values acceptable for fields that aren't yet known at denial time:
//   - itemID may be 0 для Create denials before row assigned
//   - sectionID may be 0 если section lookup itself failed
//   - curriculumID may be 0 если section/curriculum chain broken
func disciplineItemDenialFields(actorID, itemID, sectionID, curriculumID int64, reason string) map[string]any {
	return map[string]any{
		"actor_user_id": actorID,
		"item_id":       itemID,
		"section_id":    sectionID,
		"curriculum_id": curriculumID,
		"reason":        reason,
	}
}

// emitDisciplineItemAudit dispatches an audit event for the
// discipline-item resource. Nil sink is no-op (mirror к emitAudit
// pattern). Resource fixed к auditDisciplineItemResource so a typo
// cannot drift the event into curriculum or section streams.
func emitDisciplineItemAudit(sink AuditSink, ctx context.Context, action string, fields map[string]any) {
	if sink == nil {
		return
	}
	sink.LogAuditEvent(ctx, action, auditDisciplineItemResource, fields)
}
