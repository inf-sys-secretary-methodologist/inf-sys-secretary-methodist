package usecases

import (
	"context"
	"maps"
)

// recordingAuditSink captures audit calls without touching real logging.
// Shared across all use-case tests in this package.
type recordingAuditSink struct {
	events []auditCall
}

type auditCall struct {
	Action   string
	Resource string
	Fields   map[string]any
}

func (r *recordingAuditSink) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	cp := make(map[string]any, len(fields))
	maps.Copy(cp, fields)
	r.events = append(r.events, auditCall{Action: action, Resource: resource, Fields: cp})
}
