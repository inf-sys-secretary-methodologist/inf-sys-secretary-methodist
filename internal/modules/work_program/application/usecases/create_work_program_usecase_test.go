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

// fakeCreateRepo is a minimal createWorkProgramRepo test double.
type fakeCreateRepo struct {
	saveCalls  int
	saved      *entities.WorkProgram
	saveErr    error
	idAssigned int64
}

func (f *fakeCreateRepo) Save(_ context.Context, wp *entities.WorkProgram) error {
	f.saveCalls++
	f.saved = wp
	if f.saveErr != nil {
		return f.saveErr
	}
	if f.idAssigned > 0 {
		wp.SetID(f.idAssigned)
	}
	return nil
}

func validCreateInput() CreateWorkProgramInput {
	return CreateWorkProgramInput{
		DisciplineID:       7,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Annotation:         "Курс по реляционным БД",
	}
}

func TestNewCreateWorkProgramUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewCreateWorkProgramUseCase(nil, ...) did not panic")
		}
	}()
	NewCreateWorkProgramUseCase(nil, &recordingAuditSink{})
}

func TestCreateWorkProgramUseCase_HappyPath_Teacher(t *testing.T) {
	repo := &fakeCreateRepo{idAssigned: 42}
	audit := &recordingAuditSink{}
	uc := NewCreateWorkProgramUseCase(repo, audit)

	wp, err := uc.Execute(context.Background(), 7, "teacher", validCreateInput())
	require.NoError(t, err)
	require.NotNil(t, wp)
	assert.Equal(t, int64(42), wp.ID())
	assert.Equal(t, int64(7), wp.AuthorID(), "AuthorID derives from actorID")
	assert.Equal(t, domain.StatusDraft, wp.Status())
	assert.Equal(t, "09.03.01", wp.SpecialtyCode())
	assert.Equal(t, 2026, wp.ApplicableFromYear())

	require.Equal(t, 1, repo.saveCalls)
	require.NotNil(t, repo.saved)
	assert.Equal(t, "Базы данных", repo.saved.Title())

	require.Len(t, audit.events, 1, "one audit event expected")
	ev := audit.events[0]
	assert.Equal(t, "work_program.created", ev.Action)
	assert.Equal(t, "work_program", ev.Resource)
	assert.Equal(t, int64(7), ev.Fields["actor_user_id"])
	assert.Equal(t, int64(42), ev.Fields["work_program_id"])
	assert.Equal(t, "09.03.01", ev.Fields["specialty_code"])
}

func TestCreateWorkProgramUseCase_AllowedRoles(t *testing.T) {
	allowedRoles := []string{"teacher", "methodist", "system_admin"}
	for _, role := range allowedRoles {
		t.Run(role, func(t *testing.T) {
			repo := &fakeCreateRepo{idAssigned: 1}
			uc := NewCreateWorkProgramUseCase(repo, &recordingAuditSink{})
			_, err := uc.Execute(context.Background(), 1, role, validCreateInput())
			require.NoError(t, err, "role %q must be allowed to create WP", role)
			assert.Equal(t, 1, repo.saveCalls)
		})
	}
}

func TestCreateWorkProgramUseCase_DeniedRoles(t *testing.T) {
	deniedRoles := []string{"student", "academic_secretary", ""}
	for _, role := range deniedRoles {
		t.Run(role, func(t *testing.T) {
			repo := &fakeCreateRepo{}
			audit := &recordingAuditSink{}
			uc := NewCreateWorkProgramUseCase(repo, audit)

			_, err := uc.Execute(context.Background(), 1, role, validCreateInput())
			assert.True(t, errors.Is(err, domain.ErrWorkProgramScopeForbidden),
				"role %q must be denied with scope-forbidden sentinel, got %v", role, err)
			assert.Zero(t, repo.saveCalls, "repo.Save must not be called on denied role")

			require.Len(t, audit.events, 1)
			assert.Equal(t, "work_program.create_denied", audit.events[0].Action)
			assert.Equal(t, "forbidden_role", audit.events[0].Fields["reason"])
		})
	}
}

func TestCreateWorkProgramUseCase_InvalidInputAuditsDenialAndReturnsSentinel(t *testing.T) {
	repo := &fakeCreateRepo{}
	audit := &recordingAuditSink{}
	uc := NewCreateWorkProgramUseCase(repo, audit)

	// Empty title violates NewWorkProgram invariant.
	in := validCreateInput()
	in.Title = ""

	_, err := uc.Execute(context.Background(), 7, "teacher", in)
	assert.True(t, errors.Is(err, domain.ErrInvalidWorkProgram),
		"expected ErrInvalidWorkProgram, got %v", err)
	assert.Zero(t, repo.saveCalls, "repo.Save must not be called on invariant failure")

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "work_program.create_denied", ev.Action)
	assert.Equal(t, "invalid", ev.Fields["reason"])
	assert.Equal(t, int64(7), ev.Fields["actor_user_id"])
}

func TestCreateWorkProgramUseCase_IdentityConflictAuditsDenialAndReturnsSentinel(t *testing.T) {
	repo := &fakeCreateRepo{saveErr: repositories.ErrWorkProgramIdentityExists}
	audit := &recordingAuditSink{}
	uc := NewCreateWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", validCreateInput())
	assert.True(t, errors.Is(err, repositories.ErrWorkProgramIdentityExists),
		"expected ErrWorkProgramIdentityExists, got %v", err)
	assert.Equal(t, 1, repo.saveCalls)

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "work_program.create_denied", ev.Action)
	assert.Equal(t, "identity_conflict", ev.Fields["reason"])
	assert.Equal(t, "09.03.01", ev.Fields["specialty_code"])
}

func TestCreateWorkProgramUseCase_TransportErrorPropagatesWithoutSuccessAudit(t *testing.T) {
	repo := &fakeCreateRepo{saveErr: errors.New("conn refused")}
	audit := &recordingAuditSink{}
	uc := NewCreateWorkProgramUseCase(repo, audit)

	_, err := uc.Execute(context.Background(), 7, "teacher", validCreateInput())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conn refused")
	assert.Empty(t, audit.events,
		"transport errors must not produce a created/denied audit event")
}

func TestCreateWorkProgramUseCase_NilSinkIsTolerated(t *testing.T) {
	repo := &fakeCreateRepo{idAssigned: 1}
	uc := NewCreateWorkProgramUseCase(repo, nil)

	wp, err := uc.Execute(context.Background(), 7, "teacher", validCreateInput())
	require.NoError(t, err)
	assert.Equal(t, int64(1), wp.ID())
}
