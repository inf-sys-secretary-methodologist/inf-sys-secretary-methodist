package entities_test

import (
	"errors"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// approvedWPWithRevision reconstitutes an approved РПД carrying a single
// revision in the given status with a known id, so transition tests can
// address it by id (fresh NewRevision has id=0).
func approvedWPWithRevision(t *testing.T, revID int64, status domain.RevisionStatus) *entities.WorkProgram {
	t.Helper()
	approver := int64(99)
	approvedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	rev := entities.ReconstituteRevision(entities.ReconstituteRevisionInput{
		ID: revID, WorkProgramID: 7, RevisionNumber: 1, ChangeType: domain.RevisionChangeTypeOther,
		ChangeSummary: "правки", Status: status, AuthorID: 11,
		CreatedAt: approvedAt, UpdatedAt: approvedAt,
	})
	return entities.ReconstituteWorkProgram(entities.ReconstituteWorkProgramInput{
		ID: 7, DisciplineID: 42, SpecialtyCode: "09.03.01", ApplicableFromYear: 2026,
		Title: "Базы данных", Status: domain.StatusApproved, AuthorID: 11,
		ApproverID: &approver, ApprovedAt: &approvedAt, Version: 3,
		CreatedAt: approvedAt, UpdatedAt: approvedAt,
		Revisions: []*entities.Revision{rev},
	})
}

func TestWorkProgram_SubmitRevision_Draft_MovesToPendingApproval(t *testing.T) {
	wp := approvedWPWithRevision(t, 100, domain.RevisionStatusDraft)

	if err := wp.SubmitRevision(100); err != nil {
		t.Fatalf("SubmitRevision: unexpected error %v", err)
	}
	revs := wp.Revisions()
	if len(revs) != 1 {
		t.Fatalf("revisions len: got %d, want 1", len(revs))
	}
	if revs[0].Status() != domain.RevisionStatusPendingApproval {
		t.Errorf("revision status: got %q, want pending_approval", revs[0].Status())
	}
}

func TestWorkProgram_SubmitRevision_UnknownID_NotFound(t *testing.T) {
	wp := approvedWPWithRevision(t, 100, domain.RevisionStatusDraft)

	err := wp.SubmitRevision(999)
	if !errors.Is(err, domain.ErrRevisionNotFound) {
		t.Errorf("expected ErrRevisionNotFound for unknown revision id, got %v", err)
	}
}

func TestWorkProgram_SubmitRevision_WrongStatus_Rejected(t *testing.T) {
	wp := approvedWPWithRevision(t, 100, domain.RevisionStatusPendingApproval)

	err := wp.SubmitRevision(100)
	if !errors.Is(err, domain.ErrInvalidStatusTransition) {
		t.Errorf("submitting a non-draft revision must fail with ErrInvalidStatusTransition, got %v", err)
	}
}
