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

// SystemNotifier is the narrow port the users use case uses to push
// system-level notifications (role change / status change) к the
// affected user. Concrete adapter lives в cmd/server/main.go DI seam;
// users module no longer imports
// internal/modules/notifications/application/... directly.
//
// v0.160.1 polish Item 3 — closes cross-module-impl class per the
// CLAUDE.md gate ("Cross-module импорты — запрещены. Только через
// адаптеры в main.go / DI-точке"). Mirror к v0.162.1 messaging Item 3
// + v0.163.1 announcement SystemNotifier pattern.
//
// Nil notifier is treated as a successful no-op so existing test
// setups что omit notifications wiring keep working unchanged.
type SystemNotifier interface {
	SendSystemNotification(ctx context.Context, userID int64, title, message string) error
}
