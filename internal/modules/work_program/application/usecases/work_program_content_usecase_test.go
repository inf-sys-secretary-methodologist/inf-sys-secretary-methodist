package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

const contentTeacherID = int64(5)

// draftWithGoal builds a draft РПД (author = contentTeacherID) carrying one
// reconstituted Goal with the given id, so Update/Remove can target it.
func draftWithGoal(t *testing.T, goalID int64) *entities.WorkProgram {
	t.Helper()
	wp := reconstituteWPWithStatus(t, 1, contentTeacherID, domain.StatusDraft)
	if err := wp.AddGoal(entities.ReconstituteGoal(entities.ReconstituteGoalInput{ID: goalID, WorkProgramID: 1, Text: "Цель", OrderIndex: 0})); err != nil {
		t.Fatalf("seed goal: %v", err)
	}
	return wp
}

func TestWorkProgramContentUseCase_PanicsOnNilRepo(t *testing.T) {
	assert.Panics(t, func() { NewWorkProgramContentUseCase(nil, nil) })
}

// The shared skeleton: load → author-scope gate → apply → persist. Exercised
// through AddGoal as the representative operation.
func TestWorkProgramContentUseCase_AuthorizationGate(t *testing.T) {
	cases := []struct {
		name      string
		actorID   int64
		actorRole string
		notFound  bool
		wantErrIs error
		wantSaved bool
	}{
		{"not found", contentTeacherID, "teacher", true, repositories.ErrWorkProgramNotFound, false},
		{"non-author teacher forbidden", 999, "teacher", false, domain.ErrWorkProgramScopeForbidden, false},
		{"author allowed", contentTeacherID, "teacher", false, nil, true},
		{"system_admin allowed", 42, "system_admin", false, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRevisionRepo{
				programs: map[int64]*entities.WorkProgram{1: reconstituteWPWithStatus(t, 1, contentTeacherID, domain.StatusDraft)},
			}
			if tc.notFound {
				repo.getErr = map[int64]error{1: repositories.ErrWorkProgramNotFound}
			}
			audit := &recordingAuditSink{}
			uc := NewWorkProgramContentUseCase(repo, audit)

			_, err := uc.AddGoal(context.Background(), tc.actorID, tc.actorRole, 1, "Новая цель", 0)
			if tc.wantErrIs != nil {
				require.ErrorIs(t, err, tc.wantErrIs)
			} else {
				require.NoError(t, err)
			}
			if tc.wantSaved {
				assert.Equal(t, []int64{1}, repo.updateCalls, "authorized edit persists")
			} else {
				assert.Empty(t, repo.updateCalls, "rejected edit must not persist")
			}
		})
	}
}

func TestWorkProgramContentUseCase_AddGoal_AppendsAndPersists(t *testing.T) {
	wp := reconstituteWPWithStatus(t, 1, contentTeacherID, domain.StatusDraft)
	repo := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{1: wp}}
	uc := NewWorkProgramContentUseCase(repo, &recordingAuditSink{})

	got, err := uc.AddGoal(context.Background(), contentTeacherID, "teacher", 1, "Освоить нормализацию", 2)
	require.NoError(t, err)
	require.Len(t, got.Goals(), 1)
	assert.Equal(t, "Освоить нормализацию", got.Goals()[0].Text())
	assert.Equal(t, []int64{1}, repo.updateCalls)
}

func TestWorkProgramContentUseCase_UpdateGoal_AppliesAndPersists(t *testing.T) {
	wp := draftWithGoal(t, 10)
	repo := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{1: wp}}
	uc := NewWorkProgramContentUseCase(repo, &recordingAuditSink{})

	got, err := uc.UpdateGoal(context.Background(), contentTeacherID, "teacher", 1, 10, "Обновлённая цель", 1)
	require.NoError(t, err)
	assert.Equal(t, "Обновлённая цель", got.Goals()[0].Text())
	assert.Equal(t, []int64{1}, repo.updateCalls)
}

func TestWorkProgramContentUseCase_RemoveGoal_DeletesAndPersists(t *testing.T) {
	wp := draftWithGoal(t, 10)
	repo := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{1: wp}}
	uc := NewWorkProgramContentUseCase(repo, &recordingAuditSink{})

	got, err := uc.RemoveGoal(context.Background(), contentTeacherID, "teacher", 1, 10)
	require.NoError(t, err)
	assert.Empty(t, got.Goals())
	assert.Equal(t, []int64{1}, repo.updateCalls)
}

// A domain invariant violation (empty goal text) propagates and is NOT
// persisted — the use case never silently swallows it.
func TestWorkProgramContentUseCase_DomainInvariantPropagates(t *testing.T) {
	wp := reconstituteWPWithStatus(t, 1, contentTeacherID, domain.StatusDraft)
	repo := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{1: wp}}
	uc := NewWorkProgramContentUseCase(repo, &recordingAuditSink{})

	_, err := uc.AddGoal(context.Background(), contentTeacherID, "teacher", 1, "   ", 0)
	require.ErrorIs(t, err, domain.ErrInvalidWorkProgram)
	assert.Empty(t, repo.updateCalls)
}

// A missing child id propagates ErrChildNotFound from the domain.
func TestWorkProgramContentUseCase_UpdateGoal_NotFoundPropagates(t *testing.T) {
	wp := draftWithGoal(t, 10)
	repo := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{1: wp}}
	uc := NewWorkProgramContentUseCase(repo, &recordingAuditSink{})

	_, err := uc.UpdateGoal(context.Background(), contentTeacherID, "teacher", 1, 99, "x", 0)
	require.ErrorIs(t, err, domain.ErrChildNotFound)
	assert.Empty(t, repo.updateCalls)
}

// Each collection's Add path delegates to the right domain method.
func TestWorkProgramContentUseCase_AddOtherCollections(t *testing.T) {
	t.Run("competence", func(t *testing.T) {
		wp := reconstituteWPWithStatus(t, 1, contentTeacherID, domain.StatusDraft)
		repo := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{1: wp}}
		uc := NewWorkProgramContentUseCase(repo, &recordingAuditSink{})
		got, err := uc.AddCompetence(context.Background(), contentTeacherID, "teacher", 1, "ПК-1", string(domain.CompetenceTypePK), "Разработка СУБД")
		require.NoError(t, err)
		assert.Len(t, got.Competences(), 1)
	})
	t.Run("topic", func(t *testing.T) {
		wp := reconstituteWPWithStatus(t, 1, contentTeacherID, domain.StatusDraft)
		repo := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{1: wp}}
		uc := NewWorkProgramContentUseCase(repo, &recordingAuditSink{})
		got, err := uc.AddTopic(context.Background(), contentTeacherID, "teacher", 1, TopicContentInput{Kind: string(domain.TopicKindLecture), Title: "Введение", Hours: 2, OrderIndex: 0})
		require.NoError(t, err)
		assert.Len(t, got.Topics(), 1)
	})
	t.Run("assessment", func(t *testing.T) {
		wp := reconstituteWPWithStatus(t, 1, contentTeacherID, domain.StatusDraft)
		repo := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{1: wp}}
		uc := NewWorkProgramContentUseCase(repo, &recordingAuditSink{})
		got, err := uc.AddAssessment(context.Background(), contentTeacherID, "teacher", 1, AssessmentContentInput{Type: string(domain.AssessmentTypeCurrent), Description: "Опрос", MaxScore: 5})
		require.NoError(t, err)
		assert.Len(t, got.Assessments(), 1)
	})
	t.Run("reference", func(t *testing.T) {
		wp := reconstituteWPWithStatus(t, 1, contentTeacherID, domain.StatusDraft)
		repo := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{1: wp}}
		uc := NewWorkProgramContentUseCase(repo, &recordingAuditSink{})
		got, err := uc.AddReference(context.Background(), contentTeacherID, "teacher", 1, ReferenceContentInput{Kind: string(domain.ReferenceKindMain), Citation: "Дейт. Введение в БД", OrderIndex: 0})
		require.NoError(t, err)
		assert.Len(t, got.References(), 1)
	})
}
