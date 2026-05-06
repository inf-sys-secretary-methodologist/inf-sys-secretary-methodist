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

// fakeApproveRepo is a separate fake from fakeSubmitRepo so the two
// use-case tests stay independent.
type fakeApproveRepo struct {
	loaded      *entities.Curriculum
	loadErr     error
	updateErr   error
	updateCalls int
}

func (f *fakeApproveRepo) GetByID(_ context.Context, _ int64) (*entities.Curriculum, error) {
	return f.loaded, f.loadErr
}

func (f *fakeApproveRepo) Update(_ context.Context, _ *entities.Curriculum) error {
	f.updateCalls++
	return f.updateErr
}

func TestNewApproveCurriculumUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewApproveCurriculumUseCase(nil, ...) did not panic")
		}
	}()
	NewApproveCurriculumUseCase(nil, &recordingAuditSink{}, time.Now)
}

func TestApproveCurriculumUseCase_HappyPath(t *testing.T) {
	const author = int64(42)
	const admin = int64(99)
	c := reconstituted(t, 7, author, entities.StatusPendingApproval)
	repo := &fakeApproveRepo{loaded: c}
	audit := &recordingAuditSink{}
	frozenNow := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	uc := NewApproveCurriculumUseCase(repo, audit, func() time.Time { return frozenNow })

	got, err := uc.Execute(context.Background(), admin, ApproveCurriculumInput{ID: 7})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, entities.StatusApproved, got.Status())
	require.NotNil(t, got.ApprovedBy())
	assert.Equal(t, admin, *got.ApprovedBy())
	assert.Equal(t, 1, repo.updateCalls)

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "curriculum.approved", ev.Action)
	assert.Equal(t, int64(admin), ev.Fields["actor_user_id"])
	assert.Equal(t, int64(7), ev.Fields["curriculum_id"])
}

func TestApproveCurriculumUseCase_NonPendingStatusRejected(t *testing.T) {
	const author = int64(42)
	const admin = int64(99)
	cases := []struct {
		name   string
		status entities.CurriculumStatus
	}{
		{"draft", entities.StatusDraft},
		{"approved (already)", entities.StatusApproved},
		{"archived", entities.StatusArchived},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := reconstituted(t, 1, author, tc.status)
			repo := &fakeApproveRepo{loaded: c}
			audit := &recordingAuditSink{}
			uc := NewApproveCurriculumUseCase(repo, audit, time.Now)

			_, err := uc.Execute(context.Background(), admin, ApproveCurriculumInput{ID: 1})
			assert.True(t, errors.Is(err, entities.ErrCannotApprove),
				"expected ErrCannotApprove, got %v", err)
			assert.Zero(t, repo.updateCalls)

			require.Len(t, audit.events, 1)
			assert.Equal(t, "curriculum.approve_denied", audit.events[0].Action)
			assert.Equal(t, "not_pending", audit.events[0].Fields["reason"])
		})
	}
}

func TestApproveCurriculumUseCase_NotFoundDeniedAndAudited(t *testing.T) {
	repo := &fakeApproveRepo{loadErr: repositories.ErrCurriculumNotFound}
	audit := &recordingAuditSink{}
	uc := NewApproveCurriculumUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), 99, ApproveCurriculumInput{ID: 999})
	assert.True(t, errors.Is(err, repositories.ErrCurriculumNotFound))

	require.Len(t, audit.events, 1)
	assert.Equal(t, "curriculum.approve_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestApproveCurriculumUseCase_TransportErrorsPropagateWithoutAudit(t *testing.T) {
	t.Run("load transport error", func(t *testing.T) {
		transport := errors.New("conn refused on get")
		repo := &fakeApproveRepo{loadErr: transport}
		audit := &recordingAuditSink{}
		uc := NewApproveCurriculumUseCase(repo, audit, time.Now)

		_, err := uc.Execute(context.Background(), 99, ApproveCurriculumInput{ID: 1})
		assert.ErrorIs(t, err, transport)
		assert.Empty(t, audit.events)
	})

	t.Run("update transport error", func(t *testing.T) {
		const author = int64(42)
		transport := errors.New("conn refused on update")
		c := reconstituted(t, 1, author, entities.StatusPendingApproval)
		repo := &fakeApproveRepo{loaded: c, updateErr: transport}
		audit := &recordingAuditSink{}
		uc := NewApproveCurriculumUseCase(repo, audit, time.Now)

		_, err := uc.Execute(context.Background(), 99, ApproveCurriculumInput{ID: 1})
		assert.ErrorIs(t, err, transport)
		assert.Equal(t, 1, repo.updateCalls)
		assert.Empty(t, audit.events)
	})
}

func TestApproveCurriculumUseCase_NilSinkTolerated(t *testing.T) {
	const author = int64(42)
	c := reconstituted(t, 1, author, entities.StatusPendingApproval)
	repo := &fakeApproveRepo{loaded: c}
	uc := NewApproveCurriculumUseCase(repo, nil, time.Now)

	_, err := uc.Execute(context.Background(), 99, ApproveCurriculumInput{ID: 1})
	require.NoError(t, err)
}
