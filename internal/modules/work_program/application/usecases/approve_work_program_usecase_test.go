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

func TestNewApproveWorkProgramUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewApproveWorkProgramUseCase(nil, ...) did not panic")
		}
	}()
	NewApproveWorkProgramUseCase(nil, &recordingAuditSink{})
}

func TestApproveWorkProgramUseCase_HappyPath_Methodist(t *testing.T) {
	pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
	repo := &fakeTransitionRepo{wp: pending}
	audit := &recordingAuditSink{}
	uc := NewApproveWorkProgramUseCase(repo, audit)

	wp, err := uc.Execute(context.Background(), 42, "methodist", ApproveWorkProgramInput{ID: 100})
	require.NoError(t, err)
	require.NotNil(t, wp)
	assert.Equal(t, domain.StatusApproved, wp.Status())

	require.NotNil(t, wp.ApproverID(), "approver id must be populated")
	assert.Equal(t, int64(42), *wp.ApproverID())
	require.NotNil(t, wp.ApprovedAt(), "approved_at timestamp must be populated")

	assert.Equal(t, 1, repo.updateCalls)

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "work_program.approved", ev.Action)
	assert.Equal(t, "work_program", ev.Resource)
	assert.Equal(t, int64(42), ev.Fields["actor_user_id"])
	assert.Equal(t, int64(100), ev.Fields["work_program_id"])
	assert.Equal(t, "approved", ev.Fields["status"])
}

func TestApproveWorkProgramUseCase_HappyPath_SystemAdminOverride(t *testing.T) {
	pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
	repo := &fakeTransitionRepo{wp: pending}
	uc := NewApproveWorkProgramUseCase(repo, &recordingAuditSink{})

	_, err := uc.Execute(context.Background(), 99, "system_admin", ApproveWorkProgramInput{ID: 100})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)
}

func TestApproveWorkProgramUseCase_NotFoundAuditsDenial(t *testing.T) {
	repo := &fakeTransitionRepo{getErr: repositories.ErrWorkProgramNotFound}
	audit := &recordingAuditSink{}
	uc := NewApproveWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 42, "methodist", ApproveWorkProgramInput{ID: 100})
	assert.True(t, errors.Is(err, repositories.ErrWorkProgramNotFound))
	assert.Zero(t, repo.updateCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.approve_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestApproveWorkProgramUseCase_NonApproverRolesDenied(t *testing.T) {
	deniedRoles := []string{"teacher", "academic_secretary", "student", "", "unknown"}
	for _, role := range deniedRoles {
		t.Run(role, func(t *testing.T) {
			pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
			repo := &fakeTransitionRepo{wp: pending}
			audit := &recordingAuditSink{}
			uc := NewApproveWorkProgramUseCase(repo, audit)

			_, err := uc.Execute(context.Background(), 42, role, ApproveWorkProgramInput{ID: 100})
			assert.True(t, errors.Is(err, domain.ErrWorkProgramScopeForbidden),
				"role %q must be denied with scope-forbidden sentinel, got %v", role, err)
			assert.Zero(t, repo.updateCalls)

			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.approve_denied", audit.events[0].Action)
			assert.Equal(t, "forbidden_role", audit.events[0].Fields["reason"])
		})
	}
}

func TestApproveWorkProgramUseCase_WrongStatusAuditsDenial(t *testing.T) {
	// Approve only permitted from pending_approval.
	wrongStatuses := []domain.Status{
		domain.StatusDraft,
		domain.StatusApproved,
		domain.StatusNeedsRevision,
		domain.StatusArchived,
	}
	for _, status := range wrongStatuses {
		t.Run(string(status), func(t *testing.T) {
			wp := reconstituteWPWithStatus(t, 100, 7, status)
			repo := &fakeTransitionRepo{wp: wp}
			audit := &recordingAuditSink{}
			uc := NewApproveWorkProgramUseCase(repo, audit)

			_, err := uc.Execute(context.Background(), 42, "methodist", ApproveWorkProgramInput{ID: 100})
			assert.True(t, errors.Is(err, domain.ErrInvalidStatusTransition),
				"Approve from status=%s must be rejected", status)
			assert.Zero(t, repo.updateCalls)

			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.approve_denied", audit.events[0].Action)
			assert.Equal(t, "not_pending", audit.events[0].Fields["reason"])
		})
	}
}

func TestApproveWorkProgramUseCase_TransportErrorOnUpdatePropagatesWithoutAudit(t *testing.T) {
	pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
	repo := &fakeTransitionRepo{wp: pending, updateErr: errors.New("conn refused")}
	audit := &recordingAuditSink{}
	uc := NewApproveWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 42, "methodist", ApproveWorkProgramInput{ID: 100})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conn refused")
	assert.Empty(t, audit.events)
}

func TestApproveWorkProgramUseCase_NilSinkIsTolerated(t *testing.T) {
	pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
	repo := &fakeTransitionRepo{wp: pending}
	uc := NewApproveWorkProgramUseCase(repo, nil)

	_, err := uc.Execute(context.Background(), 42, "methodist", ApproveWorkProgramInput{ID: 100})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)
}
