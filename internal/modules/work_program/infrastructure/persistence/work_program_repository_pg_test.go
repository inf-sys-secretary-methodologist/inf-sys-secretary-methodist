package persistence

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
