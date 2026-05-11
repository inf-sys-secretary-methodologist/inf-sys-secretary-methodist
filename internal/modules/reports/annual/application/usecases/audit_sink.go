// Package usecases hosts the annual methodist report use cases. Cross-
// module orchestration лежит здесь: this package consumes aggregate
// methods from curriculum / assignments / documents modules to assemble
// the annual report payload, then delegates rendering to the docxgen
// infrastructure.
package usecases

import "context"

// AuditSink is the narrow port the use cases in this package use to emit
// audit events. *logging.AuditLogger satisfies this interface structurally.
// Nil sink is treated as a successful no-op so tests can omit the
// dependency without sprinkling nil checks at each call site.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}
