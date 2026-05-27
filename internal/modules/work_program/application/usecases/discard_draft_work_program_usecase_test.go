package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

func TestNewDiscardDraftWorkProgramUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewDiscardDraftWorkProgramUseCase(nil, ...) did not panic")
		}
	}()
	NewDiscardDraftWorkProgramUseCase(nil, &recordingAuditSink{})
}

func TestDiscardDraftWorkProgramUseCase_HappyPath_Author(t *testing.T) {
	repo := &fakeTransitionRepo{wp: newDraftWP(t, 100, 7)}
	audit := &recordingAuditSink{}
	uc := NewDiscardDraftWorkProgramUseCase(repo, audit)

	wp, err := uc.Execute(context.Background(), 7, "teacher", DiscardDraftWorkProgramInput{ID: 100})
	require.NoError(t, err)
	require.NotNil(t, wp)
	assert.Equal(t, domain.StatusArchived, wp.Status(),
		"draft must transition to archived after DiscardDraft")

	assert.Equal(t, 1, repo.getCalls)
	assert.Equal(t, 1, repo.updateCalls)

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "work_program.discarded", ev.Action)
	assert.Equal(t, "work_program", ev.Resource)
	assert.Equal(t, int64(7), ev.Fields["actor_user_id"])
	assert.Equal(t, int64(100), ev.Fields["work_program_id"])
	assert.Equal(t, "archived", ev.Fields["status"])
}

func TestDiscardDraftWorkProgramUseCase_HappyPath_SystemAdminOverride(t *testing.T) {
	repo := &fakeTransitionRepo{wp: newDraftWP(t, 100, 7)}
	uc := NewDiscardDraftWorkProgramUseCase(repo, &recordingAuditSink{})

	_, err := uc.Execute(context.Background(), 999, "system_admin", DiscardDraftWorkProgramInput{ID: 100})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)
}

func TestDiscardDraftWorkProgramUseCase_NotFoundAuditsDenial(t *testing.T) {
	repo := &fakeTransitionRepo{getErr: repositories.ErrWorkProgramNotFound}
	audit := &recordingAuditSink{}
	uc := NewDiscardDraftWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", DiscardDraftWorkProgramInput{ID: 100})
	assert.True(t, errors.Is(err, repositories.ErrWorkProgramNotFound))
	assert.Zero(t, repo.updateCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.discard_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestDiscardDraftWorkProgramUseCase_NonAuthorAuditsDenial(t *testing.T) {
	repo := &fakeTransitionRepo{wp: newDraftWP(t, 100, 7)}
	audit := &recordingAuditSink{}
	uc := NewDiscardDraftWorkProgramUseCase(repo, audit)

	// Another teacher (actor 8 != author 7) cannot discard.
	_, err := uc.Execute(context.Background(), 8, "teacher", DiscardDraftWorkProgramInput{ID: 100})
	assert.True(t, errors.Is(err, domain.ErrWorkProgramScopeForbidden))
	assert.Zero(t, repo.updateCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.discard_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

func TestDiscardDraftWorkProgramUseCase_NonDraftStatusAuditsDenial(t *testing.T) {
	approved := reconstituteWPWithStatus(t, 100, 7, domain.StatusApproved)
	repo := &fakeTransitionRepo{wp: approved}
	audit := &recordingAuditSink{}
	uc := NewDiscardDraftWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", DiscardDraftWorkProgramInput{ID: 100})
	assert.True(t, errors.Is(err, domain.ErrInvalidStatusTransition),
		"DiscardDraft only allowed from draft per ADR-2 FSM")
	assert.Zero(t, repo.updateCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.discard_denied", audit.events[0].Action)
	assert.Equal(t, "not_draft", audit.events[0].Fields["reason"])
}

func TestDiscardDraftWorkProgramUseCase_TransportErrorOnUpdatePropagatesWithoutAudit(t *testing.T) {
	repo := &fakeTransitionRepo{
		wp:        newDraftWP(t, 100, 7),
		updateErr: errors.New("conn refused"),
	}
	audit := &recordingAuditSink{}
	uc := NewDiscardDraftWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", DiscardDraftWorkProgramInput{ID: 100})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conn refused")
	assert.Empty(t, audit.events)
}

func TestDiscardDraftWorkProgramUseCase_NilSinkIsTolerated(t *testing.T) {
	repo := &fakeTransitionRepo{wp: newDraftWP(t, 100, 7)}
	uc := NewDiscardDraftWorkProgramUseCase(repo, nil)

	_, err := uc.Execute(context.Background(), 7, "teacher", DiscardDraftWorkProgramInput{ID: 100})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)
}
