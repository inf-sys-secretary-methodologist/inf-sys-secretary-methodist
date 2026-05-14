package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

// Clock is the narrow port for current-time injection. Lets tests
// substitute deterministic clocks without rebuilding the use case
// tree.
type Clock interface {
	Now() time.Time
}

// SystemClock returns time.Now() — production wiring in main.go.
type SystemClock struct{}

// Now returns current wall-clock time.
func (SystemClock) Now() time.Time { return time.Now() }

// SetReminderInput is the public DTO for the SetReminder use case.
// ActorUserID comes from JWT context (the caller is the owner —
// per-user privacy boundary; no other user's reminders can be set
// through this endpoint).
type SetReminderInput struct {
	TaskID        int64
	ActorUserID   int64
	ReminderType  entities.ReminderType
	MinutesBefore int
}

// SetReminderUseCase creates a new reminder for the caller against
// the supplied task.
type SetReminderUseCase struct {
	repo  repositories.TaskReminderRepository
	clock Clock
	audit AuditSink
}

// NewSetReminderUseCase constructs the use case. Panics on nil repo
// so misconfigured DI fails at boot. clock defaults to SystemClock
// if nil. audit may be nil — emission is skipped в that case.
func NewSetReminderUseCase(
	repo repositories.TaskReminderRepository,
	clock Clock,
	audit AuditSink,
) *SetReminderUseCase {
	if repo == nil {
		panic("tasks: NewSetReminderUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = SystemClock{}
	}
	return &SetReminderUseCase{repo: repo, clock: clock, audit: audit}
}

// Execute validates input via entities.NewTaskReminder, persists,
// and emits a task_reminder.set audit event on success.
//
// Stub for RED — GREEN replaces the body with the real composition.
func (uc *SetReminderUseCase) Execute(ctx context.Context, in SetReminderInput) (*entities.TaskReminder, error) {
	_ = ctx
	_ = in
	return nil, errors.New("set_reminder: not implemented yet")
}
