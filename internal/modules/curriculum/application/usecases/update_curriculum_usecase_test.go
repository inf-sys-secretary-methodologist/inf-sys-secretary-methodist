package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// fakeUpdateRepo combines GetByID + Update for the use case under test.
type fakeUpdateRepo struct {
	loaded       *entities.Curriculum
	loadErr      error
	updateErr    error
	updateCalls  int
	updatedEntry *entities.Curriculum
}

func (f *fakeUpdateRepo) GetByID(_ context.Context, _ int64) (*entities.Curriculum, error) {
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	return f.loaded, nil
}

func (f *fakeUpdateRepo) Update(_ context.Context, c *entities.Curriculum) error {
	f.updateCalls++
	f.updatedEntry = c
	return f.updateErr
}

func reconstituted(t *testing.T, id int64, createdBy int64, status entities.CurriculumStatus) *entities.Curriculum {
	t.Helper()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var ab *int64
	var aat *time.Time
	if status == entities.StatusApproved {
		v := int64(99)
		ab = &v
		t := now.Add(48 * time.Hour)
		aat = &t
	}
	return entities.ReconstituteCurriculum(
		id, "Original", "ORIG-2026", "Original Specialty", 2026, "orig desc",
		status, createdBy, ab, aat, now, now,
	)
}

func TestNewUpdateCurriculumUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewUpdateCurriculumUseCase(nil, ...) did not panic")
		}
	}()
	NewUpdateCurriculumUseCase(nil, &recordingAuditSink{}, time.Now)
}

func TestUpdateCurriculumUseCase_AuthorMethodistUpdatesOwnDraft(t *testing.T) {
	const author = int64(42)
	repo := &fakeUpdateRepo{loaded: reconstituted(t, 7, author, entities.StatusDraft)}
	audit := &recordingAuditSink{}
	frozenNow := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	uc := NewUpdateCurriculumUseCase(repo, audit, func() time.Time { return frozenNow })

	got, err := uc.Execute(context.Background(), author, false, UpdateCurriculumInput{
		ID:          7,
		Title:       "New Title",
		Code:        "NEW-2026",
		Specialty:   "New Specialty",
		Year:        2027,
		Description: "new desc",
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "New Title", got.Title())
	assert.Equal(t, "NEW-2026", got.Code())
	assert.Equal(t, 2027, got.Year())
	assert.Equal(t, frozenNow, got.UpdatedAt())
	assert.Equal(t, 1, repo.updateCalls)

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "curriculum.updated", ev.Action)
	assert.Equal(t, "curriculum", ev.Resource)
	assert.Equal(t, author, ev.Fields["actor_user_id"])
	assert.Equal(t, int64(7), ev.Fields["curriculum_id"])
	assert.Equal(t, "NEW-2026", ev.Fields["code"])
}

func TestUpdateCurriculumUseCase_StrangerMethodistOnForeignDraftRejected(t *testing.T) {
	const author = int64(42)
	const stranger = int64(7)
	repo := &fakeUpdateRepo{loaded: reconstituted(t, 1, author, entities.StatusDraft)}
	audit := &recordingAuditSink{}
	uc := NewUpdateCurriculumUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), stranger, false, UpdateCurriculumInput{
		ID: 1, Title: "T", Code: "C", Specialty: "S", Year: 2026,
	})
	assert.True(t, errors.Is(err, entities.ErrCurriculumScopeForbidden))
	assert.Zero(t, repo.updateCalls, "Update must not be called on authz failure")

	require.Len(t, audit.events, 1)
	assert.Equal(t, "curriculum.update_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

func TestUpdateCurriculumUseCase_AdminOverridesOwnership(t *testing.T) {
	const author = int64(42)
	const admin = int64(99)
	repo := &fakeUpdateRepo{loaded: reconstituted(t, 1, author, entities.StatusDraft)}
	audit := &recordingAuditSink{}
	uc := NewUpdateCurriculumUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), admin, true, UpdateCurriculumInput{
		ID: 1, Title: "T", Code: "C", Specialty: "S", Year: 2026,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "curriculum.updated", audit.events[0].Action)
	assert.Equal(t, admin, audit.events[0].Fields["actor_user_id"])
}

func TestUpdateCurriculumUseCase_NonDraftStatusRejected(t *testing.T) {
	const author = int64(42)
	cases := []struct {
		name   string
		status entities.CurriculumStatus
	}{
		{"pending_approval", entities.StatusPendingApproval},
		{"approved", entities.StatusApproved},
		{"archived", entities.StatusArchived},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeUpdateRepo{loaded: reconstituted(t, 1, author, tc.status)}
			audit := &recordingAuditSink{}
			uc := NewUpdateCurriculumUseCase(repo, audit, time.Now)

			_, err := uc.Execute(context.Background(), author, false, UpdateCurriculumInput{
				ID: 1, Title: "T", Code: "C", Specialty: "S", Year: 2026,
			})
			assert.True(t, errors.Is(err, entities.ErrCannotEditApproved),
				"expected ErrCannotEditApproved, got %v", err)
			assert.Zero(t, repo.updateCalls)

			require.Len(t, audit.events, 1)
			assert.Equal(t, "curriculum.update_denied", audit.events[0].Action)
			assert.Equal(t, "not_editable", audit.events[0].Fields["reason"])
		})
	}
}

func TestUpdateCurriculumUseCase_InvariantViolationRejected(t *testing.T) {
	const author = int64(42)
	repo := &fakeUpdateRepo{loaded: reconstituted(t, 1, author, entities.StatusDraft)}
	audit := &recordingAuditSink{}
	uc := NewUpdateCurriculumUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), author, false, UpdateCurriculumInput{
		ID: 1, Title: "", Code: "C", Specialty: "S", Year: 2026,
	})
	assert.True(t, errors.Is(err, entities.ErrInvalidCurriculum))
	assert.Zero(t, repo.updateCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "curriculum.update_denied", audit.events[0].Action)
	assert.Equal(t, "invalid", audit.events[0].Fields["reason"])
}

func TestUpdateCurriculumUseCase_CodeConflictDeniedAndAudited(t *testing.T) {
	const author = int64(42)
	repo := &fakeUpdateRepo{
		loaded:    reconstituted(t, 1, author, entities.StatusDraft),
		updateErr: repositories.ErrCurriculumCodeExists,
	}
	audit := &recordingAuditSink{}
	uc := NewUpdateCurriculumUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), author, false, UpdateCurriculumInput{
		ID: 1, Title: "T", Code: "DUP-2026", Specialty: "S", Year: 2026,
	})
	assert.True(t, errors.Is(err, repositories.ErrCurriculumCodeExists))
	assert.Equal(t, 1, repo.updateCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "curriculum.update_denied", audit.events[0].Action)
	assert.Equal(t, "code_conflict", audit.events[0].Fields["reason"])
	assert.Equal(t, "DUP-2026", audit.events[0].Fields["code"])
}

func TestUpdateCurriculumUseCase_NotFoundDeniedAndAudited(t *testing.T) {
	repo := &fakeUpdateRepo{loadErr: repositories.ErrCurriculumNotFound}
	audit := &recordingAuditSink{}
	uc := NewUpdateCurriculumUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), 7, false, UpdateCurriculumInput{
		ID: 999, Title: "T", Code: "C", Specialty: "S", Year: 2026,
	})
	assert.True(t, errors.Is(err, repositories.ErrCurriculumNotFound))
	assert.Zero(t, repo.updateCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "curriculum.update_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestUpdateCurriculumUseCase_TransportLoadErrorPropagatesWithoutAudit(t *testing.T) {
	transport := errors.New("conn refused on get")
	repo := &fakeUpdateRepo{loadErr: transport}
	audit := &recordingAuditSink{}
	uc := NewUpdateCurriculumUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), 7, false, UpdateCurriculumInput{
		ID: 1, Title: "T", Code: "C", Specialty: "S", Year: 2026,
	})
	assert.ErrorIs(t, err, transport)
	assert.Empty(t, audit.events,
		"transport error during load must not produce an audit event")
}

func TestUpdateCurriculumUseCase_TransportUpdateErrorPropagatesWithoutSuccessAudit(t *testing.T) {
	const author = int64(42)
	transport := errors.New("conn refused on update")
	repo := &fakeUpdateRepo{
		loaded:    reconstituted(t, 1, author, entities.StatusDraft),
		updateErr: transport,
	}
	audit := &recordingAuditSink{}
	uc := NewUpdateCurriculumUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), author, false, UpdateCurriculumInput{
		ID: 1, Title: "T", Code: "OK-2026", Specialty: "S", Year: 2026,
	})
	assert.ErrorIs(t, err, transport)
	assert.Equal(t, 1, repo.updateCalls)
	assert.Empty(t, audit.events,
		"transport error during update must not produce updated/denied audit")
}

func TestUpdateCurriculumUseCase_NilSinkTolerated(t *testing.T) {
	const author = int64(42)
	repo := &fakeUpdateRepo{loaded: reconstituted(t, 1, author, entities.StatusDraft)}
	uc := NewUpdateCurriculumUseCase(repo, nil, time.Now)

	_, err := uc.Execute(context.Background(), author, false, UpdateCurriculumInput{
		ID: 1, Title: "T", Code: "C", Specialty: "S", Year: 2026,
	})
	require.NoError(t, err)
}
