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

func TestNewRejectWorkProgramUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewRejectWorkProgramUseCase(nil, ...) did not panic")
		}
	}()
	NewRejectWorkProgramUseCase(nil, &recordingAuditSink{})
}

func TestRejectWorkProgramUseCase_HappyPath_Methodist(t *testing.T) {
	pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
	repo := &fakeTransitionRepo{wp: pending}
	audit := &recordingAuditSink{}
	uc := NewRejectWorkProgramUseCase(repo, audit)

	const reason = "Расчасовка не соответствует учебному плану"
	wp, err := uc.Execute(context.Background(), 42, "methodist",
		RejectWorkProgramInput{ID: 100, Reason: reason})
	require.NoError(t, err)
	require.NotNil(t, wp)
	assert.Equal(t, domain.StatusDraft, wp.Status(),
		"pending_approval → draft after Reject")
	assert.Equal(t, reason, wp.RejectReason())

	assert.Equal(t, 1, repo.updateCalls)

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "work_program.rejected", ev.Action)
	assert.Equal(t, "work_program", ev.Resource)
	assert.Equal(t, int64(42), ev.Fields["actor_user_id"])
	assert.Equal(t, int64(100), ev.Fields["work_program_id"])
	assert.Equal(t, reason, ev.Fields["reject_reason"])
}

func TestRejectWorkProgramUseCase_HappyPath_SystemAdminOverride(t *testing.T) {
	pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
	repo := &fakeTransitionRepo{wp: pending}
	uc := NewRejectWorkProgramUseCase(repo, &recordingAuditSink{})

	_, err := uc.Execute(context.Background(), 99, "system_admin",
		RejectWorkProgramInput{ID: 100, Reason: "structural mismatch"})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)
}

func TestRejectWorkProgramUseCase_NotFoundAuditsDenial(t *testing.T) {
	repo := &fakeTransitionRepo{getErr: repositories.ErrWorkProgramNotFound}
	audit := &recordingAuditSink{}
	uc := NewRejectWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 42, "methodist",
		RejectWorkProgramInput{ID: 100, Reason: "irrelevant"})
	assert.True(t, errors.Is(err, repositories.ErrWorkProgramNotFound))
	assert.Zero(t, repo.updateCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.reject_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestRejectWorkProgramUseCase_NonApproverRolesDenied(t *testing.T) {
	deniedRoles := []string{"teacher", "academic_secretary", "student", "", "unknown"}
	for _, role := range deniedRoles {
		t.Run(role, func(t *testing.T) {
			pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
			repo := &fakeTransitionRepo{wp: pending}
			audit := &recordingAuditSink{}
			uc := NewRejectWorkProgramUseCase(repo, audit)

			_, err := uc.Execute(context.Background(), 42, role,
				RejectWorkProgramInput{ID: 100, Reason: "any"})
			assert.True(t, errors.Is(err, domain.ErrWorkProgramScopeForbidden),
				"role %q must be denied, got %v", role, err)
			assert.Zero(t, repo.updateCalls)

			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.reject_denied", audit.events[0].Action)
			assert.Equal(t, "forbidden_role", audit.events[0].Fields["reason"])
		})
	}
}

func TestRejectWorkProgramUseCase_EmptyReasonAuditsDenial(t *testing.T) {
	emptyReasons := []string{"", "   ", "\t\n"}
	for _, reason := range emptyReasons {
		t.Run("empty="+reason, func(t *testing.T) {
			pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
			repo := &fakeTransitionRepo{wp: pending}
			audit := &recordingAuditSink{}
			uc := NewRejectWorkProgramUseCase(repo, audit)

			_, err := uc.Execute(context.Background(), 42, "methodist",
				RejectWorkProgramInput{ID: 100, Reason: reason})
			assert.True(t, errors.Is(err, domain.ErrRejectReasonRequired),
				"empty/whitespace reason must be rejected, got %v", err)
			assert.Zero(t, repo.updateCalls)

			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.reject_denied", audit.events[0].Action)
			assert.Equal(t, "empty_reason", audit.events[0].Fields["reason"])
		})
	}
}

func TestRejectWorkProgramUseCase_WrongStatusAuditsDenial(t *testing.T) {
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
			uc := NewRejectWorkProgramUseCase(repo, audit)

			_, err := uc.Execute(context.Background(), 42, "methodist",
				RejectWorkProgramInput{ID: 100, Reason: "any"})
			assert.True(t, errors.Is(err, domain.ErrInvalidStatusTransition),
				"Reject from status=%s must be rejected", status)
			assert.Zero(t, repo.updateCalls)

			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.reject_denied", audit.events[0].Action)
			assert.Equal(t, "not_pending", audit.events[0].Fields["reason"])
		})
	}
}

func TestRejectWorkProgramUseCase_TransportErrorOnUpdatePropagatesWithoutAudit(t *testing.T) {
	pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
	repo := &fakeTransitionRepo{wp: pending, updateErr: errors.New("conn refused")}
	audit := &recordingAuditSink{}
	uc := NewRejectWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 42, "methodist",
		RejectWorkProgramInput{ID: 100, Reason: "any"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conn refused")
	assert.Empty(t, audit.events)
}

func TestRejectWorkProgramUseCase_NilSinkIsTolerated(t *testing.T) {
	pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
	repo := &fakeTransitionRepo{wp: pending}
	uc := NewRejectWorkProgramUseCase(repo, nil)

	_, err := uc.Execute(context.Background(), 42, "methodist",
		RejectWorkProgramInput{ID: 100, Reason: "any"})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)
}
