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

// Backfill coverage for the optimistic-lock path on every use case
// that calls repo.Update. ErrWorkProgramVersionConflict is a
// repository-layer sentinel (see migration 048 + repo Update impl
// shipped in v0.176.0); these tests pin что use cases propagate the
// sentinel naked (no audit) so handlers can map к HTTP 409 and the
// UI can show a "stale view, please retry" message.
//
// Carry-forward from v0.178.0 code-review SHIP 9.50/10 — the Tests
// axis noted that the version-conflict pathway was covered only at
// the repo layer (sqlmock), not at the use-case layer where the
// handler-facing error surface is actually shaped.

func TestSubmitWorkProgramUseCase_VersionConflictPropagatesWithoutAudit(t *testing.T) {
	repo := &fakeTransitionRepo{
		wp:        newDraftWP(t, 100, 7),
		updateErr: repositories.ErrWorkProgramVersionConflict,
	}
	audit := &recordingAuditSink{}
	uc := NewSubmitWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", SubmitWorkProgramInput{ID: 100})
	require.Error(t, err)
	assert.True(t, errors.Is(err, repositories.ErrWorkProgramVersionConflict),
		"version conflict must propagate via errors.Is for handler 409 mapping")
	assert.Empty(t, audit.events,
		"version conflict is a transport-layer concern, not a policy decision — no audit")
}

func TestApproveWorkProgramUseCase_VersionConflictPropagatesWithoutAudit(t *testing.T) {
	pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
	repo := &fakeTransitionRepo{
		wp:        pending,
		updateErr: repositories.ErrWorkProgramVersionConflict,
	}
	audit := &recordingAuditSink{}
	uc := NewApproveWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 42, "methodist", ApproveWorkProgramInput{ID: 100})
	require.Error(t, err)
	assert.True(t, errors.Is(err, repositories.ErrWorkProgramVersionConflict))
	assert.Empty(t, audit.events)
}

func TestRejectWorkProgramUseCase_VersionConflictPropagatesWithoutAudit(t *testing.T) {
	pending := reconstituteWPWithStatus(t, 100, 7, domain.StatusPendingApproval)
	repo := &fakeTransitionRepo{
		wp:        pending,
		updateErr: repositories.ErrWorkProgramVersionConflict,
	}
	audit := &recordingAuditSink{}
	uc := NewRejectWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 42, "methodist",
		RejectWorkProgramInput{ID: 100, Reason: "any"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, repositories.ErrWorkProgramVersionConflict))
	assert.Empty(t, audit.events)
}

func TestDiscardDraftWorkProgramUseCase_VersionConflictPropagatesWithoutAudit(t *testing.T) {
	repo := &fakeTransitionRepo{
		wp:        newDraftWP(t, 100, 7),
		updateErr: repositories.ErrWorkProgramVersionConflict,
	}
	audit := &recordingAuditSink{}
	uc := NewDiscardDraftWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", DiscardDraftWorkProgramInput{ID: 100})
	require.Error(t, err)
	assert.True(t, errors.Is(err, repositories.ErrWorkProgramVersionConflict))
	assert.Empty(t, audit.events)
}
