package usecases

import "context"

// EventNotifier is the narrow notification port. Concrete adapter
// wired в main.go bridges to NotificationUseCase per plan ADR-7 —
// backend slice (v0.162.0) defines the port; production wiring lands
// в v0.163.0 alongside frontend.
//
// Methods correspond to the lifecycle events that surface к target
// audience: published / canceled / completed / updated. Implementations
// resolve the audience cohort и filter per user prefs.
type EventNotifier interface {
	NotifyEventPublished(ctx context.Context, eventID int64, title, audience string)
	NotifyEventCancelled(ctx context.Context, eventID int64, title, audience string)
	NotifyEventUpdated(ctx context.Context, eventID int64, title, audience string)
}

// noopNotifier is the zero-value fallback when DI does not wire a
// concrete notifier (e.g. в unit tests). Methods are no-op so use
// cases never crash on nil dispatch.
type noopNotifier struct{}

func (noopNotifier) NotifyEventPublished(_ context.Context, _ int64, _, _ string) {}
func (noopNotifier) NotifyEventCancelled(_ context.Context, _ int64, _, _ string) {}
func (noopNotifier) NotifyEventUpdated(_ context.Context, _ int64, _, _ string)   {}
