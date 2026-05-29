package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// approvedWPWithRevisionAt rebuilds an approved РПД carrying one revision
// with the given id + status, so revision-transition use cases can
// address it by id.
func approvedWPWithRevisionAt(t *testing.T, wpID, authorID, revID int64, revStatus domain.RevisionStatus) *entities.WorkProgram {
	t.Helper()
	now := time.Now().UTC()
	approver := int64(99)
	rev := entities.ReconstituteRevision(entities.ReconstituteRevisionInput{
		ID: revID, WorkProgramID: wpID, RevisionNumber: 1, ChangeType: domain.RevisionChangeTypeLiterature,
		ChangeSummary: "Обновлена литература", Status: revStatus, AuthorID: authorID,
		CreatedAt: now, UpdatedAt: now,
	})
	return entities.ReconstituteWorkProgram(entities.ReconstituteWorkProgramInput{
		ID: wpID, DisciplineID: 7, SpecialtyCode: "09.03.01", ApplicableFromYear: 2026,
		Title: "Базы данных", Status: domain.StatusApproved, AuthorID: authorID,
		ApproverID: &approver, ApprovedAt: &now, Version: 2, CreatedAt: now, UpdatedAt: now,
		Revisions: []*entities.Revision{rev},
	})
}

func TestNewSubmitRevisionUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil repo")
		}
	}()
	NewSubmitRevisionUseCase(nil, nil)
}

func TestSubmitRevisionUseCase_HappyPath(t *testing.T) {
	const authorID = int64(7)
	cases := []struct {
		name    string
		actorID int64
		role    string
	}{
		{"author_teacher", authorID, "teacher"},
		{"system_admin_override", 999, "system_admin"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := approvedWPWithRevisionAt(t, 100, authorID, 500, domain.RevisionStatusDraft)
			repo := &fakeTransitionRepo{wp: wp}
			audit := &recordingAuditSink{}
			uc := NewSubmitRevisionUseCase(repo, audit)

			got, err := uc.Execute(context.Background(), tc.actorID, tc.role,
				SubmitRevisionInput{WorkProgramID: 100, RevisionID: 500})
			require.NoError(t, err)
			require.NotNil(t, got)

			revs := got.Revisions()
			require.Len(t, revs, 1)
			assert.Equal(t, domain.RevisionStatusPendingApproval, revs[0].Status())
			assert.Equal(t, 1, repo.updateCalls)
			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.revision_submitted", audit.events[0].Action)
			assert.Equal(t, int64(500), audit.events[0].Fields["revision_id"])
		})
	}
}

func TestSubmitRevisionUseCase_NotFoundPropagates(t *testing.T) {
	repo := &fakeTransitionRepo{getErr: repositories.ErrWorkProgramNotFound}
	audit := &recordingAuditSink{}
	uc := NewSubmitRevisionUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher",
		SubmitRevisionInput{WorkProgramID: 100, RevisionID: 500})
	assert.ErrorIs(t, err, repositories.ErrWorkProgramNotFound)
	assert.Zero(t, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.revision_submit_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestSubmitRevisionUseCase_NonAuthorForbidden(t *testing.T) {
	const authorID = int64(7)
	wp := approvedWPWithRevisionAt(t, 100, authorID, 500, domain.RevisionStatusDraft)
	repo := &fakeTransitionRepo{wp: wp}
	audit := &recordingAuditSink{}
	uc := NewSubmitRevisionUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 99, "teacher",
		SubmitRevisionInput{WorkProgramID: 100, RevisionID: 500})
	assert.True(t, errors.Is(err, domain.ErrWorkProgramScopeForbidden), "got %v", err)
	assert.Zero(t, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.revision_submit_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

func TestSubmitRevisionUseCase_RevisionNotFound(t *testing.T) {
	const authorID = int64(7)
	wp := approvedWPWithRevisionAt(t, 100, authorID, 500, domain.RevisionStatusDraft)
	repo := &fakeTransitionRepo{wp: wp}
	audit := &recordingAuditSink{}
	uc := NewSubmitRevisionUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), authorID, "teacher",
		SubmitRevisionInput{WorkProgramID: 100, RevisionID: 999})
	assert.True(t, errors.Is(err, domain.ErrRevisionNotFound), "got %v", err)
	assert.Zero(t, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.revision_submit_denied", audit.events[0].Action)
	assert.Equal(t, "revision_not_found", audit.events[0].Fields["reason"])
}

func TestSubmitRevisionUseCase_WrongStatusRejected(t *testing.T) {
	const authorID = int64(7)
	wp := approvedWPWithRevisionAt(t, 100, authorID, 500, domain.RevisionStatusPendingApproval)
	repo := &fakeTransitionRepo{wp: wp}
	audit := &recordingAuditSink{}
	uc := NewSubmitRevisionUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), authorID, "teacher",
		SubmitRevisionInput{WorkProgramID: 100, RevisionID: 500})
	assert.True(t, errors.Is(err, domain.ErrInvalidStatusTransition), "got %v", err)
	assert.Zero(t, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.revision_submit_denied", audit.events[0].Action)
	assert.Equal(t, "not_submittable", audit.events[0].Fields["reason"])
}

func TestSubmitRevisionUseCase_UpdateErrorPropagates(t *testing.T) {
	const authorID = int64(7)
	wp := approvedWPWithRevisionAt(t, 100, authorID, 500, domain.RevisionStatusDraft)
	repo := &fakeTransitionRepo{wp: wp, updateErr: repositories.ErrWorkProgramVersionConflict}
	uc := NewSubmitRevisionUseCase(repo, &recordingAuditSink{})

	_, err := uc.Execute(context.Background(), authorID, "teacher",
		SubmitRevisionInput{WorkProgramID: 100, RevisionID: 500})
	assert.ErrorIs(t, err, repositories.ErrWorkProgramVersionConflict)
}

func TestSubmitRevisionUseCase_NilSinkTolerated(t *testing.T) {
	const authorID = int64(7)
	wp := approvedWPWithRevisionAt(t, 100, authorID, 500, domain.RevisionStatusDraft)
	repo := &fakeTransitionRepo{wp: wp}
	uc := NewSubmitRevisionUseCase(repo, nil)

	got, err := uc.Execute(context.Background(), authorID, "teacher",
		SubmitRevisionInput{WorkProgramID: 100, RevisionID: 500})
	require.NoError(t, err)
	assert.NotNil(t, got)
}
