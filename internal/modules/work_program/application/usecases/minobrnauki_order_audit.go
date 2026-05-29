package usecases

import "context"

// auditResourceOrder is the resource string every MinobrnaukiOrder audit
// event carries — distinct from auditResource ("work_program") because
// orders are a separate artifact (приказ Минобрнауки per ADR-11).
const auditResourceOrder = "minobrnauki_order"

// emitOrderAudit dispatches a MinobrnaukiOrder audit event. A nil sink
// is a successful no-op so use cases never sprinkle nil checks. The
// resource is fixed to auditResourceOrder so a typo can't drift the
// event into the wrong stream.
func emitOrderAudit(sink AuditSink, ctx context.Context, action string, fields map[string]any) {
	if sink == nil {
		return
	}
	sink.LogAuditEvent(ctx, action, auditResourceOrder, fields)
}

// orderDenialFields composes the canonical
// {actor_user_id, minobrnauki_order_id, reason, order_number} shape that
// every MinobrnaukiOrder *_denied audit event carries.
func orderDenialFields(actorID, orderID int64, reason, orderNumber string) map[string]any {
	return map[string]any{
		"actor_user_id":        actorID,
		"minobrnauki_order_id": orderID,
		"reason":               reason,
		"order_number":         orderNumber,
	}
}

// orderSuccessFields composes the canonical
// {actor_user_id, minobrnauki_order_id, order_number, change_scope}
// shape that every successful MinobrnaukiOrder audit event carries.
func orderSuccessFields(actorID, orderID int64, orderNumber, changeScope string) map[string]any {
	return map[string]any{
		"actor_user_id":        actorID,
		"minobrnauki_order_id": orderID,
		"order_number":         orderNumber,
		"change_scope":         changeScope,
	}
}
