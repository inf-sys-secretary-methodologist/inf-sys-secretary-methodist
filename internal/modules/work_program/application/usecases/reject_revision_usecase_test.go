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

const sampleRejectReason = "Не соответствует приказу Минобрнауки"

func TestNewRejectRevisionUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil repo")
		}
	}()
	NewRejectRevisionUseCase(nil, nil)
}

func TestRejectRevisionUseCase_HappyPath(t *testing.T) {
	const authorID = int64(7)
	cases := []struct {
		name    string
		actorID int64
		role    string
	}{
		{"methodist", 55, "methodist"},
		{"system_admin", 999, "system_admin"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := approvedWPWithRevisionAt(t, 100, authorID, 500, domain.RevisionStatusPendingApproval)
			repo := &fakeTransitionRepo{wp: wp}
			audit := &recordingAuditSink{}
			uc := NewRejectRevisionUseCase(repo, audit)

			got, err := uc.Execute(context.Background(), tc.actorID, tc.role,
				RejectRevisionInput{WorkProgramID: 100, RevisionID: 500, Reason: sampleRejectReason})
			require.NoError(t, err)
			require.NotNil(t, got)

			rev := got.Revisions()[0]
			assert.Equal(t, domain.RevisionStatusRejected, rev.Status())
			assert.Equal(t, sampleRejectReason, rev.RejectReason())
			assert.Equal(t, 1, repo.updateCalls)
			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.revision_rejected", audit.events[0].Action)
			assert.Equal(t, int64(500), audit.events[0].Fields["revision_id"])
			assert.Equal(t, sampleRejectReason, audit.events[0].Fields["reject_reason"])
		})
	}
}

func TestRejectRevisionUseCase_ForbiddenRole(t *testing.T) {
	for _, role := range []string{"teacher", "academic_secretary", "student", ""} {
		t.Run(role, func(t *testing.T) {
			repo := &fakeTransitionRepo{}
			audit := &recordingAuditSink{}
			uc := NewRejectRevisionUseCase(repo, audit)

			_, err := uc.Execute(context.Background(), 7, role,
				RejectRevisionInput{WorkProgramID: 100, RevisionID: 500, Reason: sampleRejectReason})
			assert.True(t, errors.Is(err, domain.ErrWorkProgramScopeForbidden), "got %v", err)
			assert.Zero(t, repo.getCalls, "forbidden role must not hit repo")
			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.revision_reject_denied", audit.events[0].Action)
			assert.Equal(t, "forbidden_role", audit.events[0].Fields["reason"])
		})
	}
}

func TestRejectRevisionUseCase_NotFoundPropagates(t *testing.T) {
	repo := &fakeTransitionRepo{getErr: repositories.ErrWorkProgramNotFound}
	audit := &recordingAuditSink{}
	uc := NewRejectRevisionUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 55, "methodist",
		RejectRevisionInput{WorkProgramID: 100, RevisionID: 500, Reason: sampleRejectReason})
	assert.ErrorIs(t, err, repositories.ErrWorkProgramNotFound)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestRejectRevisionUseCase_RevisionNotFound(t *testing.T) {
	wp := approvedWPWithRevisionAt(t, 100, 7, 500, domain.RevisionStatusPendingApproval)
	repo := &fakeTransitionRepo{wp: wp}
	audit := &recordingAuditSink{}
	uc := NewRejectRevisionUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 55, "methodist",
		RejectRevisionInput{WorkProgramID: 100, RevisionID: 999, Reason: sampleRejectReason})
	assert.True(t, errors.Is(err, domain.ErrRevisionNotFound), "got %v", err)
	assert.Zero(t, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "revision_not_found", audit.events[0].Fields["reason"])
}

func TestRejectRevisionUseCase_WrongStatusRejected(t *testing.T) {
	wp := approvedWPWithRevisionAt(t, 100, 7, 500, domain.RevisionStatusDraft)
	repo := &fakeTransitionRepo{wp: wp}
	audit := &recordingAuditSink{}
	uc := NewRejectRevisionUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 55, "methodist",
		RejectRevisionInput{WorkProgramID: 100, RevisionID: 500, Reason: sampleRejectReason})
	assert.True(t, errors.Is(err, domain.ErrInvalidStatusTransition), "got %v", err)
	assert.Zero(t, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "not_pending", audit.events[0].Fields["reason"])
}

func TestRejectRevisionUseCase_EmptyReasonRejected(t *testing.T) {
	wp := approvedWPWithRevisionAt(t, 100, 7, 500, domain.RevisionStatusPendingApproval)
	repo := &fakeTransitionRepo{wp: wp}
	audit := &recordingAuditSink{}
	uc := NewRejectRevisionUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 55, "methodist",
		RejectRevisionInput{WorkProgramID: 100, RevisionID: 500, Reason: "   "})
	assert.True(t, errors.Is(err, domain.ErrRejectReasonRequired), "got %v", err)
	assert.Zero(t, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "empty_reason", audit.events[0].Fields["reason"])
}

func TestRejectRevisionUseCase_UpdateErrorPropagates(t *testing.T) {
	wp := approvedWPWithRevisionAt(t, 100, 7, 500, domain.RevisionStatusPendingApproval)
	repo := &fakeTransitionRepo{wp: wp, updateErr: repositories.ErrWorkProgramVersionConflict}
	uc := NewRejectRevisionUseCase(repo, &recordingAuditSink{})

	_, err := uc.Execute(context.Background(), 55, "methodist",
		RejectRevisionInput{WorkProgramID: 100, RevisionID: 500, Reason: sampleRejectReason})
	assert.ErrorIs(t, err, repositories.ErrWorkProgramVersionConflict)
}

func TestRejectRevisionUseCase_NilSinkTolerated(t *testing.T) {
	wp := approvedWPWithRevisionAt(t, 100, 7, 500, domain.RevisionStatusPendingApproval)
	repo := &fakeTransitionRepo{wp: wp}
	uc := NewRejectRevisionUseCase(repo, nil)

	got, err := uc.Execute(context.Background(), 55, "methodist",
		RejectRevisionInput{WorkProgramID: 100, RevisionID: 500, Reason: sampleRejectReason})
	require.NoError(t, err)
	assert.NotNil(t, got)
}
