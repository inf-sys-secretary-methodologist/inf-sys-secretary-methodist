package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// fakeSubmitRepo combines GetByID + Update for the Submit use case.
type fakeSubmitRepo struct {
	loaded      *entities.Curriculum
	loadErr     error
	updateErr   error
	updateCalls int
}

func (f *fakeSubmitRepo) GetByID(_ context.Context, _ int64) (*entities.Curriculum, error) {
	return f.loaded, f.loadErr
}

func (f *fakeSubmitRepo) Update(_ context.Context, _ *entities.Curriculum) error {
	f.updateCalls++
	return f.updateErr
}

func TestNewSubmitForApprovalUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewSubmitForApprovalUseCase(nil, ...) did not panic")
		}
	}()
	NewSubmitForApprovalUseCase(nil, &recordingAuditSink{}, time.Now)
}

func TestSubmitForApprovalUseCase_AuthorSubmitsOwnDraft(t *testing.T) {
	const author = int64(42)
	c := reconstituted(t, 7, author, entities.StatusDraft)
	repo := &fakeSubmitRepo{loaded: c}
	audit := &recordingAuditSink{}
	frozenNow := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	uc := NewSubmitForApprovalUseCase(repo, audit, func() time.Time { return frozenNow })

	got, err := uc.Execute(context.Background(), author, false, SubmitForApprovalInput{ID: 7})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, entities.StatusPendingApproval, got.Status())
	assert.Equal(t, 1, repo.updateCalls)

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "curriculum.submitted", ev.Action)
	assert.Equal(t, author, ev.Fields["actor_user_id"])
	assert.Equal(t, int64(7), ev.Fields["curriculum_id"])
}

func TestSubmitForApprovalUseCase_AdminSubmitsForeignDraft(t *testing.T) {
	const author = int64(42)
	const admin = int64(99)
	c := reconstituted(t, 7, author, entities.StatusDraft)
	repo := &fakeSubmitRepo{loaded: c}
	audit := &recordingAuditSink{}
	uc := NewSubmitForApprovalUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), admin, true, SubmitForApprovalInput{ID: 7})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "curriculum.submitted", audit.events[0].Action)
	assert.Equal(t, admin, audit.events[0].Fields["actor_user_id"])
}

func TestSubmitForApprovalUseCase_StrangerMethodistRejected(t *testing.T) {
	const author = int64(42)
	const stranger = int64(7)
	c := reconstituted(t, 1, author, entities.StatusDraft)
	repo := &fakeSubmitRepo{loaded: c}
	audit := &recordingAuditSink{}
	uc := NewSubmitForApprovalUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), stranger, false, SubmitForApprovalInput{ID: 1})
	assert.True(t, errors.Is(err, entities.ErrCurriculumScopeForbidden))
	assert.Zero(t, repo.updateCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "curriculum.submit_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

func TestSubmitForApprovalUseCase_NonDraftStatusRejected(t *testing.T) {
	const author = int64(42)
	cases := []struct {
		name   string
		status entities.CurriculumStatus
	}{
		{"already pending", entities.StatusPendingApproval},
		{"approved", entities.StatusApproved},
		{"archived", entities.StatusArchived},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := reconstituted(t, 1, author, tc.status)
			repo := &fakeSubmitRepo{loaded: c}
			audit := &recordingAuditSink{}
			uc := NewSubmitForApprovalUseCase(repo, audit, time.Now)

			_, err := uc.Execute(context.Background(), author, false, SubmitForApprovalInput{ID: 1})
			assert.True(t, errors.Is(err, entities.ErrCannotSubmit),
				"expected ErrCannotSubmit, got %v", err)
			assert.Zero(t, repo.updateCalls)

			require.Len(t, audit.events, 1)
			assert.Equal(t, "curriculum.submit_denied", audit.events[0].Action)
			assert.Equal(t, "not_draft", audit.events[0].Fields["reason"])
		})
	}
}

func TestSubmitForApprovalUseCase_NotFoundDeniedAndAudited(t *testing.T) {
	repo := &fakeSubmitRepo{loadErr: repositories.ErrCurriculumNotFound}
	audit := &recordingAuditSink{}
	uc := NewSubmitForApprovalUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), 7, false, SubmitForApprovalInput{ID: 999})
	assert.True(t, errors.Is(err, repositories.ErrCurriculumNotFound))

	require.Len(t, audit.events, 1)
	assert.Equal(t, "curriculum.submit_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestSubmitForApprovalUseCase_TransportErrorsPropagateWithoutAudit(t *testing.T) {
	t.Run("load transport error", func(t *testing.T) {
		transport := errors.New("conn refused on get")
		repo := &fakeSubmitRepo{loadErr: transport}
		audit := &recordingAuditSink{}
		uc := NewSubmitForApprovalUseCase(repo, audit, time.Now)

		_, err := uc.Execute(context.Background(), 7, false, SubmitForApprovalInput{ID: 1})
		assert.ErrorIs(t, err, transport)
		assert.Empty(t, audit.events)
	})

	t.Run("update transport error", func(t *testing.T) {
		const author = int64(42)
		transport := errors.New("conn refused on update")
		c := reconstituted(t, 1, author, entities.StatusDraft)
		repo := &fakeSubmitRepo{loaded: c, updateErr: transport}
		audit := &recordingAuditSink{}
		uc := NewSubmitForApprovalUseCase(repo, audit, time.Now)

		_, err := uc.Execute(context.Background(), author, false, SubmitForApprovalInput{ID: 1})
		assert.ErrorIs(t, err, transport)
		assert.Equal(t, 1, repo.updateCalls)
		assert.Empty(t, audit.events)
	})
}

func TestSubmitForApprovalUseCase_NilSinkTolerated(t *testing.T) {
	const author = int64(42)
	c := reconstituted(t, 1, author, entities.StatusDraft)
	repo := &fakeSubmitRepo{loaded: c}
	uc := NewSubmitForApprovalUseCase(repo, nil, time.Now)

	_, err := uc.Execute(context.Background(), author, false, SubmitForApprovalInput{ID: 1})
	require.NoError(t, err)
}
