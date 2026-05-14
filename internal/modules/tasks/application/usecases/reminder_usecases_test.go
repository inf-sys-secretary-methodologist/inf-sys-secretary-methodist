package usecases

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
)

// fakeReminderRepo is a deterministic in-memory replacement for
// TaskReminderRepository. Sufficient for the unit tests below.
type fakeReminderRepo struct {
	mu      sync.Mutex
	rows    map[int64]*entities.TaskReminder
	nextID  int64
	created int
	deleted int
	// error injection
	createErr  error
	getByIDErr error
	deleteErr  error
	listErr    error
}

func newFakeReminderRepo() *fakeReminderRepo {
	return &fakeReminderRepo{rows: map[int64]*entities.TaskReminder{}, nextID: 100}
}

func (r *fakeReminderRepo) Create(_ context.Context, rem *entities.TaskReminder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.createErr != nil {
		return r.createErr
	}
	r.nextID++
	persisted := entities.HydrateFromPersistence(
		r.nextID,
		rem.TaskID(),
		rem.UserID(),
		rem.ReminderType(),
		rem.MinutesBefore(),
		rem.IsSent(),
		rem.SentAt(),
		rem.CreatedAt(),
	)
	*rem = *persisted
	r.rows[r.nextID] = persisted
	r.created++
	return nil
}

func (r *fakeReminderRepo) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.deleteErr != nil {
		return r.deleteErr
	}
	if _, ok := r.rows[id]; !ok {
		return errors.New("not-found")
	}
	delete(r.rows, id)
	r.deleted++
	return nil
}

func (r *fakeReminderRepo) GetByID(_ context.Context, id int64) (*entities.TaskReminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}
	rem, ok := r.rows[id]
	if !ok {
		return nil, errors.New("not-found")
	}
	return rem, nil
}

func (r *fakeReminderRepo) ListByTaskAndUser(_ context.Context, taskID, userID int64) ([]*entities.TaskReminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.listErr != nil {
		return nil, r.listErr
	}
	out := []*entities.TaskReminder{}
	for _, rem := range r.rows {
		if rem.TaskID() == taskID && rem.UserID() == userID {
			out = append(out, rem)
		}
	}
	return out, nil
}

func (r *fakeReminderRepo) GetPendingReminders(_ context.Context, _ time.Time) ([]*entities.TaskReminder, error) {
	return nil, nil
}

func (r *fakeReminderRepo) MarkSentBatch(_ context.Context, _ []int64, _ time.Time) error {
	return nil
}

// fakeClock returns a fixed time.
type fakeClock struct{ now time.Time }

func (c fakeClock) Now() time.Time { return c.now }

// auditCall records one invocation.
type auditCall struct {
	Action   string
	Resource string
	Fields   map[string]any
}

// fakeAudit mirrors AuditSink.
type fakeAudit struct {
	mu    sync.Mutex
	calls []auditCall
}

func (a *fakeAudit) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.calls = append(a.calls, auditCall{Action: action, Resource: resource, Fields: fields})
}

// TestSetReminderUseCase_HappyPath verifies the full path —
// domain validation passes, repo Create succeeds, audit emitted.
func TestSetReminderUseCase_HappyPath(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	repo := newFakeReminderRepo()
	audit := &fakeAudit{}
	uc := NewSetReminderUseCase(repo, fakeClock{now: now}, audit)

	rem, err := uc.Execute(context.Background(), SetReminderInput{
		TaskID:        42,
		ActorUserID:   7,
		ReminderType:  entities.ReminderTypeTelegram,
		MinutesBefore: 15,
	})
	require.NoError(t, err)
	require.NotNil(t, rem)
	assert.Equal(t, int64(101), rem.ID(), "first created reminder gets id=101")
	assert.Equal(t, int64(42), rem.TaskID())
	assert.Equal(t, int64(7), rem.UserID())
	assert.Equal(t, entities.ReminderTypeTelegram, rem.ReminderType())
	assert.Equal(t, 15, rem.MinutesBefore())
	assert.False(t, rem.IsSent())
	assert.Equal(t, now, rem.CreatedAt(), "created_at = injected clock")

	assert.Equal(t, 1, repo.created)
	require.Len(t, audit.calls, 1)
	assert.Equal(t, "task_reminder.set", audit.calls[0].Action)
	assert.Equal(t, "task_reminder", audit.calls[0].Resource)
	assert.Equal(t, int64(7), audit.calls[0].Fields["user_id"])
	assert.Equal(t, int64(42), audit.calls[0].Fields["task_id"])
}

// TestSetReminderUseCase_InvalidInput maps all 4 domain sentinels
// to use case error propagation. Table-driven (CLAUDE.md ≥3 gate).
func TestSetReminderUseCase_InvalidInput(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		input   SetReminderInput
		wantErr error
	}{
		{"zero_task_id", SetReminderInput{0, 7, entities.ReminderTypeEmail, 15}, entities.ErrInvalidTaskID},
		{"zero_user_id", SetReminderInput{42, 0, entities.ReminderTypeEmail, 15}, entities.ErrInvalidUserID},
		{"unknown_type", SetReminderInput{42, 7, entities.ReminderType("slack"), 15}, entities.ErrInvalidReminderType},
		{"zero_minutes", SetReminderInput{42, 7, entities.ReminderTypeEmail, 0}, entities.ErrInvalidMinutesBefore},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeReminderRepo()
			audit := &fakeAudit{}
			uc := NewSetReminderUseCase(repo, fakeClock{now: now}, audit)
			_, err := uc.Execute(context.Background(), tc.input)
			require.Error(t, err)
			assert.True(t, errors.Is(err, tc.wantErr), "want %v got %v", tc.wantErr, err)
			assert.Equal(t, 0, repo.created, "rejected input does not hit repo")
			assert.Len(t, audit.calls, 0, "rejected input emits no audit")
		})
	}
}

// TestSetReminderUseCase_RepoError surfaces the repo failure
// up to the caller without firing the audit event.
func TestSetReminderUseCase_RepoError(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	repo := newFakeReminderRepo()
	repo.createErr = errors.New("disk full")
	audit := &fakeAudit{}
	uc := NewSetReminderUseCase(repo, fakeClock{now: now}, audit)

	_, err := uc.Execute(context.Background(), SetReminderInput{
		TaskID: 42, ActorUserID: 7, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 30,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disk full")
	assert.Len(t, audit.calls, 0, "repo error → no audit emission")
}

// TestSetReminderUseCase_NilAuditSink confirms the use case is
// resilient к a nil audit dependency (test-friendly default).
func TestSetReminderUseCase_NilAuditSink(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	repo := newFakeReminderRepo()
	uc := NewSetReminderUseCase(repo, fakeClock{now: now}, nil)

	_, err := uc.Execute(context.Background(), SetReminderInput{
		TaskID: 42, ActorUserID: 7, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 30,
	})
	require.NoError(t, err)
}

// TestNewSetReminderUseCase_NilRepo_Panics surfaces misconfigured
// DI early — boot fails fast.
func TestNewSetReminderUseCase_NilRepo_Panics(t *testing.T) {
	assert.Panics(t, func() {
		_ = NewSetReminderUseCase(nil, nil, nil)
	})
}

// TestListTaskRemindersUseCase_FiltersByUser verifies the per-user
// privacy boundary — каждый user sees own reminders only.
func TestListTaskRemindersUseCase_FiltersByUser(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	repo := newFakeReminderRepo()

	// Seed: user 7 has 2 reminders, user 8 has 1, all on task 42.
	for _, in := range []SetReminderInput{
		{TaskID: 42, ActorUserID: 7, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 15},
		{TaskID: 42, ActorUserID: 7, ReminderType: entities.ReminderTypeTelegram, MinutesBefore: 30},
		{TaskID: 42, ActorUserID: 8, ReminderType: entities.ReminderTypeInApp, MinutesBefore: 60},
	} {
		setUC := NewSetReminderUseCase(repo, fakeClock{now: now}, nil)
		_, err := setUC.Execute(context.Background(), in)
		require.NoError(t, err)
	}

	listUC := NewListTaskRemindersUseCase(repo)
	out, err := listUC.Execute(context.Background(), ListTaskRemindersInput{TaskID: 42, ActorUserID: 7})
	require.NoError(t, err)
	assert.Len(t, out, 2, "user 7 sees own 2 reminders")
	for _, rem := range out {
		assert.Equal(t, int64(7), rem.UserID(), "no leakage of other users' reminders")
	}
}

// TestListTaskRemindersUseCase_Empty returns empty non-nil slice.
func TestListTaskRemindersUseCase_Empty(t *testing.T) {
	repo := newFakeReminderRepo()
	uc := NewListTaskRemindersUseCase(repo)
	out, err := uc.Execute(context.Background(), ListTaskRemindersInput{TaskID: 42, ActorUserID: 7})
	require.NoError(t, err)
	require.NotNil(t, out, "empty slice not nil")
	assert.Len(t, out, 0)
}

// TestDeleteReminderUseCase_HappyPath — owner deletes own reminder
// → row gone + audit emitted.
func TestDeleteReminderUseCase_HappyPath(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	repo := newFakeReminderRepo()
	audit := &fakeAudit{}

	setUC := NewSetReminderUseCase(repo, fakeClock{now: now}, nil)
	rem, err := setUC.Execute(context.Background(), SetReminderInput{
		TaskID: 42, ActorUserID: 7, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 15,
	})
	require.NoError(t, err)

	delUC := NewDeleteReminderUseCase(repo, audit)
	require.NoError(t, delUC.Execute(context.Background(), DeleteReminderInput{
		ReminderID: rem.ID(), TaskID: 42, ActorUserID: 7,
	}))
	assert.Equal(t, 1, repo.deleted)
	require.Len(t, audit.calls, 1)
	assert.Equal(t, "task_reminder.deleted", audit.calls[0].Action)
}

// TestDeleteReminderUseCase_NotFound — repo returns error → use
// case propagates without audit emission.
func TestDeleteReminderUseCase_NotFound(t *testing.T) {
	repo := newFakeReminderRepo()
	audit := &fakeAudit{}
	uc := NewDeleteReminderUseCase(repo, audit)
	err := uc.Execute(context.Background(), DeleteReminderInput{
		ReminderID: 999, TaskID: 42, ActorUserID: 7,
	})
	require.Error(t, err)
	assert.Len(t, audit.calls, 0)
}

// TestDeleteReminderUseCase_WrongTask — reminder exists but
// addresses a different task. Returns ErrReminderNotFoundForTask
// (handler maps к 404 без leaking row's actual task_id).
func TestDeleteReminderUseCase_WrongTask(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	repo := newFakeReminderRepo()
	setUC := NewSetReminderUseCase(repo, fakeClock{now: now}, nil)
	rem, err := setUC.Execute(context.Background(), SetReminderInput{
		TaskID: 42, ActorUserID: 7, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 15,
	})
	require.NoError(t, err)

	audit := &fakeAudit{}
	delUC := NewDeleteReminderUseCase(repo, audit)
	err = delUC.Execute(context.Background(), DeleteReminderInput{
		ReminderID: rem.ID(), TaskID: 99, ActorUserID: 7,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrReminderNotFoundForTask), "want sentinel got %v", err)
	assert.Equal(t, 0, repo.deleted, "wrong-task delete must not hit repo")
	assert.Len(t, audit.calls, 0)
}

// TestDeleteReminderUseCase_WrongOwner — reminder belongs к
// another user. Returns ErrReminderOwnerOnly → 403 in handler.
func TestDeleteReminderUseCase_WrongOwner(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	repo := newFakeReminderRepo()
	setUC := NewSetReminderUseCase(repo, fakeClock{now: now}, nil)
	rem, err := setUC.Execute(context.Background(), SetReminderInput{
		TaskID: 42, ActorUserID: 7, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 15,
	})
	require.NoError(t, err)

	audit := &fakeAudit{}
	delUC := NewDeleteReminderUseCase(repo, audit)
	err = delUC.Execute(context.Background(), DeleteReminderInput{
		ReminderID: rem.ID(), TaskID: 42, ActorUserID: 99,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrReminderOwnerOnly), "want sentinel got %v", err)
	assert.Equal(t, 0, repo.deleted, "non-owner delete must not hit repo")
	assert.Len(t, audit.calls, 0)
}
