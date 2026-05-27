package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// fakeGetRepo is a minimal getWorkProgramRepo test double.
type fakeGetRepo struct {
	wp       *entities.WorkProgram
	err      error
	getCalls int
}

func (f *fakeGetRepo) GetByID(_ context.Context, _ int64) (*entities.WorkProgram, error) {
	f.getCalls++
	if f.err != nil {
		return nil, f.err
	}
	return f.wp, nil
}

func TestNewGetWorkProgramUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewGetWorkProgramUseCase(nil, ...) did not panic")
		}
	}()
	NewGetWorkProgramUseCase(nil, &recordingAuditSink{})
}

func TestGetWorkProgramUseCase_NotFoundPropagatesWithoutAudit(t *testing.T) {
	repo := &fakeGetRepo{err: repositories.ErrWorkProgramNotFound}
	audit := &recordingAuditSink{}
	uc := NewGetWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", GetWorkProgramInput{ID: 100})
	assert.True(t, errors.Is(err, repositories.ErrWorkProgramNotFound),
		"expected ErrWorkProgramNotFound, got %v", err)
	assert.Empty(t, audit.events,
		"not-found reads must not produce audit noise (ID typos / race deletes are common)")
}

func TestGetWorkProgramUseCase_ViewRightsMatrix_Allowed(t *testing.T) {
	// Author is teacher id 7. WP at the requested status is loaded
	// and the actor inspects it under (actorID, actorRole, status).
	const authorID = int64(7)

	cases := []struct {
		name      string
		actorID   int64
		actorRole string
		status    domain.Status
	}{
		// system_admin sees everything
		{"system_admin_draft", 999, "system_admin", domain.StatusDraft},
		{"system_admin_approved", 999, "system_admin", domain.StatusApproved},
		// methodist (approver role) sees everything
		{"methodist_draft", 11, "methodist", domain.StatusDraft},
		{"methodist_pending", 11, "methodist", domain.StatusPendingApproval},
		{"methodist_approved", 11, "methodist", domain.StatusApproved},
		// academic_secretary cross-refs curriculum + WP — sees everything
		{"academic_secretary_draft", 12, "academic_secretary", domain.StatusDraft},
		{"academic_secretary_approved", 12, "academic_secretary", domain.StatusApproved},
		// teacher sees own at any status
		{"teacher_owner_draft", authorID, "teacher", domain.StatusDraft},
		{"teacher_owner_pending", authorID, "teacher", domain.StatusPendingApproval},
		{"teacher_owner_approved", authorID, "teacher", domain.StatusApproved},
		{"teacher_owner_needs_revision", authorID, "teacher", domain.StatusNeedsRevision},
		// teacher sees other authors' approved (collaborative cross-reference)
		{"teacher_other_approved", 99, "teacher", domain.StatusApproved},
		// student sees only approved (273-ФЗ ст. 29 mandatory openness)
		{"student_approved", 200, "student", domain.StatusApproved},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := reconstituteWPWithStatus(t, 100, authorID, tc.status)
			repo := &fakeGetRepo{wp: wp}
			audit := &recordingAuditSink{}
			uc := NewGetWorkProgramUseCase(repo, audit)

			got, err := uc.Execute(context.Background(), tc.actorID, tc.actorRole, GetWorkProgramInput{ID: 100})
			require.NoError(t, err, "(role=%s actor=%d) on status=%s must succeed",
				tc.actorRole, tc.actorID, tc.status)
			assert.Same(t, wp, got)
			assert.Empty(t, audit.events,
				"successful reads do not audit (only denials do)")
		})
	}
}

func TestGetWorkProgramUseCase_ViewRightsMatrix_Denied(t *testing.T) {
	const authorID = int64(7)

	cases := []struct {
		name      string
		actorID   int64
		actorRole string
		status    domain.Status
	}{
		// student denied non-approved statuses
		{"student_draft", 200, "student", domain.StatusDraft},
		{"student_pending", 200, "student", domain.StatusPendingApproval},
		{"student_needs_revision", 200, "student", domain.StatusNeedsRevision},
		{"student_archived", 200, "student", domain.StatusArchived},
		// teacher (non-author) denied non-approved
		{"teacher_other_draft", 99, "teacher", domain.StatusDraft},
		{"teacher_other_pending", 99, "teacher", domain.StatusPendingApproval},
		{"teacher_other_needs_revision", 99, "teacher", domain.StatusNeedsRevision},
		// unknown / empty role denied unconditionally
		{"unknown_role_approved", 1, "guest", domain.StatusApproved},
		{"empty_role_approved", 1, "", domain.StatusApproved},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := reconstituteWPWithStatus(t, 100, authorID, tc.status)
			repo := &fakeGetRepo{wp: wp}
			audit := &recordingAuditSink{}
			uc := NewGetWorkProgramUseCase(repo, audit)

			_, err := uc.Execute(context.Background(), tc.actorID, tc.actorRole, GetWorkProgramInput{ID: 100})
			assert.True(t, errors.Is(err, domain.ErrWorkProgramScopeForbidden),
				"(%s, status=%s) must return ErrWorkProgramScopeForbidden, got %v",
				tc.actorRole, tc.status, err)

			require.Len(t, audit.events, 1, "denied reads must emit one audit event")
			ev := audit.events[0]
			assert.Equal(t, "work_program.view_denied", ev.Action)
			assert.Equal(t, "work_program", ev.Resource)
			assert.Equal(t, "forbidden", ev.Fields["reason"])
			assert.Equal(t, tc.actorID, ev.Fields["actor_user_id"])
			assert.Equal(t, int64(100), ev.Fields["work_program_id"])
		})
	}
}

func TestGetWorkProgramUseCase_NilSinkIsTolerated(t *testing.T) {
	wp := reconstituteWPWithStatus(t, 100, 7, domain.StatusApproved)
	repo := &fakeGetRepo{wp: wp}
	uc := NewGetWorkProgramUseCase(repo, nil)

	got, err := uc.Execute(context.Background(), 200, "student", GetWorkProgramInput{ID: 100})
	require.NoError(t, err)
	assert.Same(t, wp, got)
}
