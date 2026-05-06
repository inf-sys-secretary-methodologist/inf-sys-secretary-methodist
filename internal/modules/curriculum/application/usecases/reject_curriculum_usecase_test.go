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

type fakeRejectRepo struct {
	loaded      *entities.Curriculum
	loadErr     error
	updateErr   error
	updateCalls int
}

func (f *fakeRejectRepo) GetByID(_ context.Context, _ int64) (*entities.Curriculum, error) {
	return f.loaded, f.loadErr
}

func (f *fakeRejectRepo) Update(_ context.Context, _ *entities.Curriculum) error {
	f.updateCalls++
	return f.updateErr
}

func TestNewRejectCurriculumUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewRejectCurriculumUseCase(nil, ...) did not panic")
		}
	}()
	NewRejectCurriculumUseCase(nil, &recordingAuditSink{}, time.Now)
}

func TestRejectCurriculumUseCase_HappyPathRecordsReasonInAuditOnly(t *testing.T) {
	const author = int64(42)
	const admin = int64(99)
	c := reconstituted(t, 7, author, entities.StatusPendingApproval)
	repo := &fakeRejectRepo{loaded: c}
	audit := &recordingAuditSink{}
	uc := NewRejectCurriculumUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), admin, RejectCurriculumInput{
		ID:     7,
		Reason: "Не соответствует ФГОС, переделать раздел дисциплин",
	})
	require.NoError(t, err)
	assert.Equal(t, entities.StatusDraft, c.Status(),
		"after Reject the curriculum returns to draft so the methodist may revise")
	assert.Equal(t, 1, repo.updateCalls)

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "curriculum.rejected", ev.Action)
	assert.Equal(t, int64(admin), ev.Fields["actor_user_id"])
	assert.Equal(t, int64(7), ev.Fields["curriculum_id"])
	assert.Equal(t, "Не соответствует ФГОС, переделать раздел дисциплин",
		ev.Fields["reason"],
		"reason flows verbatim into the audit field — entity does NOT persist it")
}

func TestRejectCurriculumUseCase_NonPendingStatusRejected(t *testing.T) {
	const author = int64(42)
	const admin = int64(99)
	cases := []struct {
		name   string
		status entities.CurriculumStatus
	}{
		{"draft", entities.StatusDraft},
		{"approved", entities.StatusApproved},
		{"archived", entities.StatusArchived},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := reconstituted(t, 1, author, tc.status)
			repo := &fakeRejectRepo{loaded: c}
			audit := &recordingAuditSink{}
			uc := NewRejectCurriculumUseCase(repo, audit, time.Now)

			_, err := uc.Execute(context.Background(), admin, RejectCurriculumInput{
				ID: 1, Reason: "any",
			})
			assert.True(t, errors.Is(err, entities.ErrCannotReject),
				"expected ErrCannotReject, got %v", err)
			assert.Zero(t, repo.updateCalls)

			require.Len(t, audit.events, 1)
			assert.Equal(t, "curriculum.reject_denied", audit.events[0].Action)
			assert.Equal(t, "not_pending", audit.events[0].Fields["reason"])
		})
	}
}

func TestRejectCurriculumUseCase_NotFoundDeniedAndAudited(t *testing.T) {
	repo := &fakeRejectRepo{loadErr: repositories.ErrCurriculumNotFound}
	audit := &recordingAuditSink{}
	uc := NewRejectCurriculumUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), 99, RejectCurriculumInput{ID: 999, Reason: "any"})
	assert.True(t, errors.Is(err, repositories.ErrCurriculumNotFound))

	require.Len(t, audit.events, 1)
	assert.Equal(t, "curriculum.reject_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestRejectCurriculumUseCase_TransportErrorsPropagateWithoutAudit(t *testing.T) {
	t.Run("load transport error", func(t *testing.T) {
		transport := errors.New("conn refused on get")
		repo := &fakeRejectRepo{loadErr: transport}
		audit := &recordingAuditSink{}
		uc := NewRejectCurriculumUseCase(repo, audit, time.Now)

		_, err := uc.Execute(context.Background(), 99, RejectCurriculumInput{ID: 1, Reason: "x"})
		assert.ErrorIs(t, err, transport)
		assert.Empty(t, audit.events)
	})

	t.Run("update transport error", func(t *testing.T) {
		const author = int64(42)
		transport := errors.New("conn refused on update")
		c := reconstituted(t, 1, author, entities.StatusPendingApproval)
		repo := &fakeRejectRepo{loaded: c, updateErr: transport}
		audit := &recordingAuditSink{}
		uc := NewRejectCurriculumUseCase(repo, audit, time.Now)

		_, err := uc.Execute(context.Background(), 99, RejectCurriculumInput{ID: 1, Reason: "x"})
		assert.ErrorIs(t, err, transport)
		assert.Equal(t, 1, repo.updateCalls)
		assert.Empty(t, audit.events)
	})
}

func TestRejectCurriculumUseCase_NilSinkTolerated(t *testing.T) {
	const author = int64(42)
	c := reconstituted(t, 1, author, entities.StatusPendingApproval)
	repo := &fakeRejectRepo{loaded: c}
	uc := NewRejectCurriculumUseCase(repo, nil, time.Now)

	_, err := uc.Execute(context.Background(), 99, RejectCurriculumInput{ID: 1, Reason: "x"})
	require.NoError(t, err)
}
