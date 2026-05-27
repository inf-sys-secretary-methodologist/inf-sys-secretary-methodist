package persistence

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

func newWPRepoMock(t *testing.T) (*WorkProgramRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewWorkProgramRepositoryPG(db), mock
}

func freshDraftWP(t *testing.T) *entities.WorkProgram {
	t.Helper()
	wp, err := entities.NewWorkProgram(entities.NewWorkProgramInput{
		DisciplineID:       42,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Annotation:         "Курс по основам СУБД",
		AuthorID:           7,
	})
	require.NoError(t, err)
	return wp
}

func TestWorkProgramRepositoryPG_Save_RootOnly_HappyPath(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	wp := freshDraftWP(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO work_programs")).
		WithArgs(
			int64(42),     // discipline_id
			"09.03.01",    // specialty_code
			2026,          // applicable_from_year
			"Базы данных", // title
			sqlmock.AnyArg(),
			"draft",  // status
			int64(7), // author_id
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			0,                // version
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)))
	mock.ExpectCommit()

	err := repo.Save(context.Background(), wp)
	require.NoError(t, err)
	assert.Equal(t, int64(100), wp.ID(), "Save should write the generated id back onto the aggregate")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkProgramRepositoryPG_Save_IdentityConflict_ReturnsErrIdentityExists(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	wp := freshDraftWP(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO work_programs")).
		WillReturnError(&pq.Error{
			Code:       "23505",
			Constraint: "uq_wp_discipline_specialty_cohort",
		})
	mock.ExpectRollback()

	err := repo.Save(context.Background(), wp)
	assert.ErrorIs(t, err, repositories.ErrWorkProgramIdentityExists,
		"unique violation on identity tuple must map to ErrWorkProgramIdentityExists")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkProgramRepositoryPG_Save_WithChildren_InsertsAll(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	wp := freshDraftWP(t)

	// Add one of each child type via the aggregate.
	goal, err := entities.NewGoal("Освоить SQL", 0)
	require.NoError(t, err)
	require.NoError(t, wp.AddGoal(goal))

	comp, err := entities.NewCompetence("ПК-3", "pk", "Разработка СУБД")
	require.NoError(t, err)
	require.NoError(t, wp.AddCompetence(comp))

	topic, err := entities.NewTopic(entities.NewTopicInput{
		Kind: "lecture", Title: "Введение", Hours: 4,
	})
	require.NoError(t, err)
	require.NoError(t, wp.AddTopic(topic))

	ass, err := entities.NewAssessmentCriterion(entities.NewAssessmentCriterionInput{
		Type: "current", Description: "Опрос", MaxScore: 5,
	})
	require.NoError(t, err)
	require.NoError(t, wp.AddAssessment(ass))

	ref, err := entities.NewReference(entities.NewReferenceInput{
		Kind: "main", Citation: "Дейт",
	})
	require.NoError(t, err)
	require.NoError(t, wp.AddReference(ref))
	// Revision intentionally not added — drafts cannot carry revisions
	// (ErrRevisionNotPermitted); covered separately when status flows
	// allow it. Here we cover the 5 always-allowed child kinds.

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO work_programs")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO work_program_goals")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO work_program_competences")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO work_program_topics")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO work_program_assessment")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO work_program_references")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.Save(context.Background(), wp)
	require.NoError(t, err)
	assert.Equal(t, int64(100), wp.ID())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkProgramRepositoryPG_Save_ChildInsertFailure_RollsBack(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	wp := freshDraftWP(t)

	goal, err := entities.NewGoal("Освоить SQL", 0)
	require.NoError(t, err)
	require.NoError(t, wp.AddGoal(goal))

	childErr := errors.New("simulated child insert failure")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO work_programs")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO work_program_goals")).
		WillReturnError(childErr)
	mock.ExpectRollback()

	err = repo.Save(context.Background(), wp)
	require.Error(t, err)
	assert.ErrorIs(t, err, childErr, "child insert failure must surface and roll back the tx")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByID ---

func wpRootRow(id int64, status string, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "discipline_id", "specialty_code", "applicable_from_year",
		"title", "annotation", "status", "author_id",
		"approver_id", "approved_at", "reject_reason", "version",
		"created_at", "updated_at",
	}).AddRow(
		id, int64(42), "09.03.01", 2026,
		"Базы данных", sql.NullString{String: "СУБД", Valid: true}, status, int64(7),
		sql.NullInt64{}, sql.NullTime{}, sql.NullString{}, 0,
		now, now,
	)
}

func TestWorkProgramRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_programs WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	got, err := repo.GetByID(context.Background(), 999)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, repositories.ErrWorkProgramNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkProgramRepositoryPG_GetByID_RootOnly_PopulatesFields(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("FROM work_programs WHERE id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(wpRootRow(100, "draft", now))
	// All 6 child SELECTs return empty rows.
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_program_goals WHERE work_program_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_program_id", "text", "order_index", "created_at"}))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_program_competences WHERE work_program_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_program_id", "code", "type", "description", "created_at"}))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_program_topics WHERE work_program_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_program_id", "kind", "title", "hours", "week_number", "learning_outcomes", "order_index"}))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_program_assessment WHERE work_program_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_program_id", "type", "description", "max_score", "example_questions"}))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_program_references WHERE work_program_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_program_id", "kind", "citation", "year", "isbn", "url", "order_index"}))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_program_revisions WHERE work_program_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_program_id", "revision_number", "change_type", "change_summary", "status", "author_id", "approver_id", "approved_at", "reject_reason", "diff_payload", "created_at", "updated_at"}))

	got, err := repo.GetByID(context.Background(), 100)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(100), got.ID())
	assert.Equal(t, int64(42), got.DisciplineID())
	assert.Equal(t, "09.03.01", got.SpecialtyCode())
	assert.Equal(t, 2026, got.ApplicableFromYear())
	assert.Equal(t, "Базы данных", got.Title())
	assert.Equal(t, "СУБД", got.Annotation())
	assert.Equal(t, int64(7), got.AuthorID())
	assert.Empty(t, got.Goals())
	assert.Empty(t, got.Competences())
	assert.Empty(t, got.Topics())
	assert.Empty(t, got.Assessments())
	assert.Empty(t, got.References())
	assert.Empty(t, got.Revisions())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkProgramRepositoryPG_GetByID_HydratesAllChildKinds(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("FROM work_programs WHERE id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(wpRootRow(100, "draft", now))

	mock.ExpectQuery(regexp.QuoteMeta("FROM work_program_goals WHERE work_program_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_program_id", "text", "order_index", "created_at"}).
			AddRow(int64(1), int64(100), "Освоить SQL", 0, now))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_program_competences WHERE work_program_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_program_id", "code", "type", "description", "created_at"}).
			AddRow(int64(2), int64(100), "ПК-3", "pk", "Разработка СУБД", now))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_program_topics WHERE work_program_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_program_id", "kind", "title", "hours", "week_number", "learning_outcomes", "order_index"}).
			AddRow(int64(3), int64(100), "lecture", "Введение", 4, sql.NullInt32{}, sql.NullString{}, 0))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_program_assessment WHERE work_program_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_program_id", "type", "description", "max_score", "example_questions"}).
			AddRow(int64(4), int64(100), "current", "Опрос", 5, pq.Array([]string{"Q1"})))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_program_references WHERE work_program_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_program_id", "kind", "citation", "year", "isbn", "url", "order_index"}).
			AddRow(int64(5), int64(100), "main", "Дейт", sql.NullInt32{}, sql.NullString{}, sql.NullString{}, 0))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_program_revisions WHERE work_program_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "work_program_id", "revision_number", "change_type", "change_summary", "status", "author_id", "approver_id", "approved_at", "reject_reason", "diff_payload", "created_at", "updated_at"}).
			AddRow(int64(6), int64(100), 1, "other", "правки", "draft", int64(7), sql.NullInt64{}, sql.NullTime{}, sql.NullString{}, []byte(nil), now, now))

	got, err := repo.GetByID(context.Background(), 100)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Len(t, got.Goals(), 1)
	assert.Len(t, got.Competences(), 1)
	assert.Len(t, got.Topics(), 1)
	assert.Len(t, got.Assessments(), 1)
	assert.Len(t, got.References(), 1)
	assert.Len(t, got.Revisions(), 1)
	assert.Equal(t, "Освоить SQL", got.Goals()[0].Text())
	assert.Equal(t, "ПК-3", got.Competences()[0].Code())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Update ---

// approvedWP returns a Reconstituted WP suitable for Update testing
// (carries ID + non-zero version so optimistic-lock arg verification
// has known values).
func approvedWP(t *testing.T) *entities.WorkProgram {
	t.Helper()
	approver := int64(99)
	approvedAt := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	return entities.ReconstituteWorkProgram(entities.ReconstituteWorkProgramInput{
		ID:                 100,
		DisciplineID:       42,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Annotation:         "СУБД",
		Status:             "approved",
		AuthorID:           7,
		ApproverID:         &approver,
		ApprovedAt:         &approvedAt,
		Version:            3,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
}

func TestWorkProgramRepositoryPG_Update_HappyPath_BumpsVersion(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	wp := approvedWP(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE work_programs SET")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Children: 6 DELETE statements before reinsert pass — empty
	// collections still issue the DELETE so reinsert is symmetric.
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM work_program_goals")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM work_program_competences")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM work_program_topics")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM work_program_assessment")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM work_program_references")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM work_program_revisions")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.Update(context.Background(), wp)
	require.NoError(t, err)
	assert.Equal(t, 4, wp.Version(), "Update must bump version on the entity to mirror the row state")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkProgramRepositoryPG_Update_StaleVersion_ReturnsVersionConflict(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	wp := approvedWP(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE work_programs SET")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	// Follow-up existence check finds the row → version conflict.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM work_programs WHERE id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}).AddRow(1))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), wp)
	assert.ErrorIs(t, err, repositories.ErrWorkProgramVersionConflict)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkProgramRepositoryPG_Update_RowMissing_ReturnsNotFound(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	wp := approvedWP(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE work_programs SET")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM work_programs WHERE id = $1")).
		WithArgs(int64(100)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := repo.Update(context.Background(), wp)
	assert.ErrorIs(t, err, repositories.ErrWorkProgramNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Delete ---

func TestWorkProgramRepositoryPG_Delete_HappyPath(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM work_programs WHERE id = $1")).
		WithArgs(int64(100)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 100)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkProgramRepositoryPG_Delete_NotFound(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM work_programs WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	assert.ErrorIs(t, err, repositories.ErrWorkProgramNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- List ---

func wpListRow(id int64, status string) []driver.Value {
	return []driver.Value{
		id, int64(42), "09.03.01", 2026, "Базы данных", status, int64(7), 0,
	}
}

func wpListColumnsTest() []string {
	return []string{
		"id", "discipline_id", "specialty_code", "applicable_from_year",
		"title", "status", "author_id", "version",
	}
}

func TestWorkProgramRepositoryPG_List_EmptyFilter_ReturnsAllRows(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM work_programs")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_programs")).
		WillReturnRows(sqlmock.NewRows(wpListColumnsTest()).
			AddRow(wpListRow(1, "draft")...).
			AddRow(wpListRow(2, "approved")...))

	got, err := repo.List(context.Background(), repositories.WorkProgramListFilter{Limit: 20})
	require.NoError(t, err)
	assert.Equal(t, 2, got.Total)
	require.Len(t, got.Items, 2)
	assert.Equal(t, int64(1), got.Items[0].ID)
	assert.Equal(t, "draft", string(got.Items[0].Status))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkProgramRepositoryPG_List_FilterByAuthor_PassesAuthorIDArg(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	authorID := int64(7)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM work_programs")).
		WithArgs(
			"", sql.NullInt64{}, "", sql.NullInt32{},
			sql.NullInt64{Int64: 7, Valid: true},
		).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_programs")).
		WithArgs(
			"", sql.NullInt64{}, "", sql.NullInt32{},
			sql.NullInt64{Int64: 7, Valid: true},
			20, 0,
		).
		WillReturnRows(sqlmock.NewRows(wpListColumnsTest()).
			AddRow(wpListRow(1, "draft")...))

	got, err := repo.List(context.Background(), repositories.WorkProgramListFilter{
		AuthorID: &authorID,
		Limit:    20,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, got.Total)
	assert.Len(t, got.Items, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkProgramRepositoryPG_List_EmptyResult_NotAnError(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM work_programs")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_programs")).
		WillReturnRows(sqlmock.NewRows(wpListColumnsTest()))

	got, err := repo.List(context.Background(), repositories.WorkProgramListFilter{Limit: 20})
	require.NoError(t, err)
	assert.Equal(t, 0, got.Total)
	assert.Empty(t, got.Items)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkProgramRepositoryPG_List_FilterByStatusAndSpecialty(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	status := domain.StatusApproved
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM work_programs")).
		WithArgs(
			"approved", sql.NullInt64{}, "09.03.01", sql.NullInt32{}, sql.NullInt64{},
		).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(regexp.QuoteMeta("FROM work_programs")).
		WithArgs(
			"approved", sql.NullInt64{}, "09.03.01", sql.NullInt32{}, sql.NullInt64{},
			10, 20,
		).
		WillReturnRows(sqlmock.NewRows(wpListColumnsTest()).
			AddRow(wpListRow(100, "approved")...))

	got, err := repo.List(context.Background(), repositories.WorkProgramListFilter{
		Status:        &status,
		SpecialtyCode: "09.03.01",
		Limit:         10,
		Offset:        20,
	})
	require.NoError(t, err)
	assert.Equal(t, 3, got.Total)
	assert.Len(t, got.Items, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkProgramRepositoryPG_Save_BeginTxFailure_Surfaces(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	wp := freshDraftWP(t)

	beginErr := errors.New("simulated begin failure")
	mock.ExpectBegin().WillReturnError(beginErr)

	err := repo.Save(context.Background(), wp)
	require.Error(t, err)
	assert.ErrorIs(t, err, beginErr, "BeginTx failure must surface to caller")
	assert.NoError(t, mock.ExpectationsWereMet())
}
