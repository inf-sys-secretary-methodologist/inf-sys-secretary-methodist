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

// fakeTransitionRepo is the shared narrow-port double for use cases
// that follow the load-mutate-persist pattern (Submit / DiscardDraft
// in PR 3a; Approve / Reject in PR 3b reuse the same shape).
type fakeTransitionRepo struct {
	wp          *entities.WorkProgram
	getErr      error
	updateErr   error
	getCalls    int
	updateCalls int
	updatedWP   *entities.WorkProgram
}

func (f *fakeTransitionRepo) GetByID(_ context.Context, _ int64) (*entities.WorkProgram, error) {
	f.getCalls++
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.wp, nil
}

func (f *fakeTransitionRepo) Update(_ context.Context, wp *entities.WorkProgram) error {
	f.updateCalls++
	f.updatedWP = wp
	return f.updateErr
}

// newDraftWP builds a draft WorkProgram via the canonical constructor,
// then attaches a synthetic id (mirrors the post-Save state the
// repository would produce).
func newDraftWP(t *testing.T, id, authorID int64) *entities.WorkProgram {
	t.Helper()
	wp, err := entities.NewWorkProgram(entities.NewWorkProgramInput{
		DisciplineID:       7,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		AuthorID:           authorID,
	})
	require.NoError(t, err)
	wp.SetID(id)
	return wp
}

// reconstituteWPWithStatus rebuilds a WorkProgram at the requested
// status — used for wrong-status tests where the canonical constructor
// always lands the aggregate in draft.
func reconstituteWPWithStatus(t *testing.T, id, authorID int64, status domain.Status) *entities.WorkProgram {
	t.Helper()
	now := time.Now().UTC()
	in := entities.ReconstituteWorkProgramInput{
		ID:                 id,
		DisciplineID:       7,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Status:             status,
		AuthorID:           authorID,
		Version:            1,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if status == domain.StatusApproved {
		approverID := int64(99)
		in.ApproverID = &approverID
		in.ApprovedAt = &now
	}
	return entities.ReconstituteWorkProgram(in)
}

func TestNewSubmitWorkProgramUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewSubmitWorkProgramUseCase(nil, ...) did not panic")
		}
	}()
	NewSubmitWorkProgramUseCase(nil, &recordingAuditSink{})
}

func TestSubmitWorkProgramUseCase_HappyPath_Author(t *testing.T) {
	repo := &fakeTransitionRepo{wp: newDraftWP(t, 100, 7)}
	audit := &recordingAuditSink{}
	uc := NewSubmitWorkProgramUseCase(repo, audit)

	wp, err := uc.Execute(context.Background(), 7, "teacher", SubmitWorkProgramInput{ID: 100})
	require.NoError(t, err)
	require.NotNil(t, wp)
	assert.Equal(t, domain.StatusPendingApproval, wp.Status(),
		"draft must transition to pending_approval after Submit")

	assert.Equal(t, 1, repo.getCalls)
	assert.Equal(t, 1, repo.updateCalls)

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "work_program.submitted", ev.Action)
	assert.Equal(t, "work_program", ev.Resource)
	assert.Equal(t, int64(7), ev.Fields["actor_user_id"])
	assert.Equal(t, int64(100), ev.Fields["work_program_id"])
	assert.Equal(t, "pending_approval", ev.Fields["status"])
}

func TestSubmitWorkProgramUseCase_HappyPath_SystemAdminOverride(t *testing.T) {
	repo := &fakeTransitionRepo{wp: newDraftWP(t, 100, 7)}
	uc := NewSubmitWorkProgramUseCase(repo, &recordingAuditSink{})

	// Actor id 999 != author id 7. Allowed via system_admin override.
	_, err := uc.Execute(context.Background(), 999, "system_admin", SubmitWorkProgramInput{ID: 100})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)
}

func TestSubmitWorkProgramUseCase_NotFoundAuditsDenialAndReturnsSentinel(t *testing.T) {
	repo := &fakeTransitionRepo{getErr: repositories.ErrWorkProgramNotFound}
	audit := &recordingAuditSink{}
	uc := NewSubmitWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", SubmitWorkProgramInput{ID: 100})
	assert.True(t, errors.Is(err, repositories.ErrWorkProgramNotFound),
		"expected ErrWorkProgramNotFound, got %v", err)
	assert.Zero(t, repo.updateCalls, "repo.Update must not be called when GetByID fails")

	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.submit_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestSubmitWorkProgramUseCase_NonAuthorAuditsDenialAndReturnsSentinel(t *testing.T) {
	repo := &fakeTransitionRepo{wp: newDraftWP(t, 100, 7)}
	audit := &recordingAuditSink{}
	uc := NewSubmitWorkProgramUseCase(repo, audit)

	// Actor id 8 is teacher but not the author (id 7). Override flag false.
	_, err := uc.Execute(context.Background(), 8, "teacher", SubmitWorkProgramInput{ID: 100})
	assert.True(t, errors.Is(err, domain.ErrWorkProgramScopeForbidden),
		"expected ErrWorkProgramScopeForbidden, got %v", err)
	assert.Zero(t, repo.updateCalls, "repo.Update must not be called on authz failure")

	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.submit_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

func TestSubmitWorkProgramUseCase_WrongStatusAuditsDenialAndReturnsSentinel(t *testing.T) {
	approved := reconstituteWPWithStatus(t, 100, 7, domain.StatusApproved)
	repo := &fakeTransitionRepo{wp: approved}
	audit := &recordingAuditSink{}
	uc := NewSubmitWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", SubmitWorkProgramInput{ID: 100})
	assert.True(t, errors.Is(err, domain.ErrInvalidStatusTransition),
		"expected ErrInvalidStatusTransition, got %v", err)
	assert.Zero(t, repo.updateCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "work_program.submit_denied", audit.events[0].Action)
	assert.Equal(t, "not_submittable", audit.events[0].Fields["reason"])
}

func TestSubmitWorkProgramUseCase_TransportErrorOnUpdatePropagatesWithoutAudit(t *testing.T) {
	repo := &fakeTransitionRepo{
		wp:        newDraftWP(t, 100, 7),
		updateErr: errors.New("conn refused"),
	}
	audit := &recordingAuditSink{}
	uc := NewSubmitWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", SubmitWorkProgramInput{ID: 100})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conn refused")
	assert.Empty(t, audit.events,
		"transport errors must not produce a submitted/denied audit event")
}

func TestSubmitWorkProgramUseCase_NilSinkIsTolerated(t *testing.T) {
	repo := &fakeTransitionRepo{wp: newDraftWP(t, 100, 7)}
	uc := NewSubmitWorkProgramUseCase(repo, nil)

	_, err := uc.Execute(context.Background(), 7, "teacher", SubmitWorkProgramInput{ID: 100})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)
}
