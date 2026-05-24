package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// Compile-time assertion that *logging.AuditLogger satisfies the narrow
// AuditSink port. Catches signature drift in either type at build time
// rather than at the main.go DI seam.
var _ AuditSink = (*logging.AuditLogger)(nil)

// AuditSink is the narrow port the users use cases (user / department /
// position) use to emit audit events. *logging.AuditLogger satisfies
// this structurally; tests substitute a recording fake. Defined next
// to the use case per Clean Architecture: repository-style interfaces
// live in the consumer package, not in domain.
//
// Pattern mirror к the messaging / announcements / curriculum
// precedents — closes the v0.160.0 audit T2 finding about users
// usecases holding the concrete *logging.AuditLogger while sibling
// modules had already migrated к the narrow port.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}
