package usecases

import (
	"context"
	"time"
)

// DebtNotifier is the narrow port used to notify a student about events
// on their debt (a resit has been scheduled). Best-effort and
// fire-and-forget: no error is returned — a failed notification must not
// fail the academic operation that triggered it. main.go wires a concrete
// adapter over the notifications module (no cross-module import here); an
// unset notifier is a no-op via notifyResitScheduled.
//
// Implementations MUST treat their work as non-blocking or hand off to a
// background worker — use cases call this on the request path.
type DebtNotifier interface {
	NotifyResitScheduled(ctx context.Context, studentUserID, debtID int64, disciplineName string, scheduledDate time.Time)
}

// notifyResitScheduled dispatches the notification only when a notifier is
// wired AND the debt is linked to a local student account (best-effort
// link). A nil notifier or unresolved student is a silent no-op.
func notifyResitScheduled(n DebtNotifier, ctx context.Context, studentUserID *int64, debtID int64, disciplineName string, scheduledDate time.Time) {
	if n == nil || studentUserID == nil {
		return
	}
	n.NotifyResitScheduled(ctx, *studentUserID, debtID, disciplineName, scheduledDate)
}
