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

func validCreateRevisionInput(wpID int64) CreateRevisionInput {
	return CreateRevisionInput{
		WorkProgramID: wpID,
		ChangeType:    "literature",
		ChangeSummary: "Обновлён список основной литературы по приказу",
	}
}

func TestCreateRevisionUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil repo")
		}
	}()
	NewCreateRevisionUseCase(nil, nil)
}

func TestCreateRevisionUseCase_HappyPath(t *testing.T) {
	const authorID = int64(7)
	cases := []struct {
		name    string
		actorID int64
		role    string
		status  domain.Status
	}{
		{"author_teacher_on_approved", authorID, "teacher", domain.StatusApproved},
		{"author_teacher_on_needs_revision", authorID, "teacher", domain.StatusNeedsRevision},
		{"system_admin_override", 999, "system_admin", domain.StatusApproved},
		{"methodist_author_on_approved", authorID, "methodist", domain.StatusApproved},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := reconstituteWPWithStatus(t, 100, authorID, tc.status)
			repo := &fakeTransitionRepo{wp: wp}
			audit := &recordingAuditSink{}
			uc := NewCreateRevisionUseCase(repo, audit)

			got, err := uc.Execute(context.Background(), tc.actorID, tc.role, validCreateRevisionInput(100))
			require.NoError(t, err)
			require.NotNil(t, got)

			revs := got.Revisions()
			require.Len(t, revs, 1, "a revision must be appended")
			assert.Equal(t, 1, revs[0].RevisionNumber())
			assert.Equal(t, domain.RevisionStatusDraft, revs[0].Status())
			assert.Equal(t, domain.RevisionChangeTypeLiterature, revs[0].ChangeType())
			assert.Equal(t, tc.actorID, revs[0].AuthorID(), "revision author is the acting user")

			assert.Equal(t, 1, repo.updateCalls)
			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.revision_created", audit.events[0].Action)
			assert.Equal(t, tc.actorID, audit.events[0].Fields["actor_user_id"])
			assert.Equal(t, 1, audit.events[0].Fields["revision_number"])
		})
	}
}

func TestCreateRevisionUseCase_NotFoundPropagates(t *testing.T) {
	repo := &fakeTransitionRepo{getErr: repositories.ErrWorkProgramNotFound}
	audit := &recordingAuditSink{}
	uc := NewCreateRevisionUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", validCreateRevisionInput(100))
	assert.ErrorIs(t, err, repositories.ErrWorkProgramNotFound)
	assert.Zero(t, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.revision_create_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestCreateRevisionUseCase_NonAuthorForbidden(t *testing.T) {
	const authorID = int64(7)
	cases := []struct {
		name    string
		actorID int64
		role    string
	}{
		{"other_teacher", 99, "teacher"},
		{"methodist_not_author", 11, "methodist"},
		{"student", 7, "student"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// student case shares actorID with author to prove the role —
			// not the id — is what the non-author branch must reject.
			wp := reconstituteWPWithStatus(t, 100, authorID, domain.StatusApproved)
			if tc.role == "student" {
				wp = reconstituteWPWithStatus(t, 100, 1234, domain.StatusApproved)
			}
			repo := &fakeTransitionRepo{wp: wp}
			audit := &recordingAuditSink{}
			uc := NewCreateRevisionUseCase(repo, audit)

			_, err := uc.Execute(context.Background(), tc.actorID, tc.role, validCreateRevisionInput(100))
			assert.True(t, errors.Is(err, domain.ErrWorkProgramScopeForbidden),
				"non-author must be forbidden, got %v", err)
			assert.Zero(t, repo.updateCalls)
			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.revision_create_denied", audit.events[0].Action)
			assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
		})
	}
}

func TestCreateRevisionUseCase_InvalidInputRejected(t *testing.T) {
	const authorID = int64(7)
	cases := []struct {
		name string
		in   CreateRevisionInput
	}{
		{"bad_change_type", CreateRevisionInput{WorkProgramID: 100, ChangeType: "nonsense", ChangeSummary: "x"}},
		{"empty_summary", CreateRevisionInput{WorkProgramID: 100, ChangeType: "other", ChangeSummary: "   "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := reconstituteWPWithStatus(t, 100, authorID, domain.StatusApproved)
			repo := &fakeTransitionRepo{wp: wp}
			audit := &recordingAuditSink{}
			uc := NewCreateRevisionUseCase(repo, audit)

			_, err := uc.Execute(context.Background(), authorID, "teacher", tc.in)
			assert.True(t, errors.Is(err, domain.ErrInvalidWorkProgram), "got %v", err)
			assert.Zero(t, repo.updateCalls)
			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.revision_create_denied", audit.events[0].Action)
			assert.Equal(t, "invalid", audit.events[0].Fields["reason"])
		})
	}
}

func TestCreateRevisionUseCase_ParentNotRevisable(t *testing.T) {
	const authorID = int64(7)
	for _, st := range []domain.Status{domain.StatusDraft, domain.StatusPendingApproval, domain.StatusArchived} {
		t.Run(string(st), func(t *testing.T) {
			wp := reconstituteWPWithStatus(t, 100, authorID, st)
			repo := &fakeTransitionRepo{wp: wp}
			audit := &recordingAuditSink{}
			uc := NewCreateRevisionUseCase(repo, audit)

			_, err := uc.Execute(context.Background(), authorID, "teacher", validCreateRevisionInput(100))
			assert.True(t, errors.Is(err, domain.ErrRevisionNotPermitted),
				"revisions only on approved/needs_revision, got %v", err)
			assert.Zero(t, repo.updateCalls)
			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.revision_create_denied", audit.events[0].Action)
			assert.Equal(t, "not_permitted", audit.events[0].Fields["reason"])
		})
	}
}

func TestCreateRevisionUseCase_UpdateErrorPropagates(t *testing.T) {
	const authorID = int64(7)
	wp := reconstituteWPWithStatus(t, 100, authorID, domain.StatusApproved)
	repo := &fakeTransitionRepo{wp: wp, updateErr: repositories.ErrWorkProgramVersionConflict}
	uc := NewCreateRevisionUseCase(repo, &recordingAuditSink{})

	_, err := uc.Execute(context.Background(), authorID, "teacher", validCreateRevisionInput(100))
	assert.ErrorIs(t, err, repositories.ErrWorkProgramVersionConflict)
}

func TestCreateRevisionUseCase_NilSinkTolerated(t *testing.T) {
	const authorID = int64(7)
	wp := reconstituteWPWithStatus(t, 100, authorID, domain.StatusApproved)
	repo := &fakeTransitionRepo{wp: wp}
	uc := NewCreateRevisionUseCase(repo, nil)

	got, err := uc.Execute(context.Background(), authorID, "teacher", validCreateRevisionInput(100))
	require.NoError(t, err)
	assert.NotNil(t, got)
}
