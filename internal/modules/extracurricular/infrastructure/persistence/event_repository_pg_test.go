package persistence

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/repositories"
)

func newEventRepoMock(t *testing.T) (*EventRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewEventRepositoryPG(db), mock
}

func freshEvent(t *testing.T) *entities.ExtracurricularEvent {
	t.Helper()
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	e, err := entities.NewExtracurricularEvent(entities.NewExtracurricularEventParams{
		Title:          "Концерт",
		Description:    "desc",
		Category:       entities.CategoryCultural,
		TargetAudience: entities.TargetAudienceAll,
		Location:       "Актовый зал",
		StartAt:        now.Add(48 * time.Hour),
		EndAt:          now.Add(50 * time.Hour),
		OrganizerID:    42,
		Now:            now,
	})
	require.NoError(t, err)
	return e
}

// ===== Save =====

func TestEventRepoPG_Save_HappyPath(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	e := freshEvent(t)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO extracurricular_events")).
		WithArgs("Концерт", sql.NullString{String: "desc", Valid: true},
			"cultural", "all", "draft",
			sql.NullString{String: "Актовый зал", Valid: true},
			e.StartAt(), e.EndAt(), sql.NullInt64{},
			int64(42), 0, e.CreatedAt(), e.UpdatedAt()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(99)))

	err := repo.Save(context.Background(), e)
	require.NoError(t, err)
	assert.Equal(t, int64(99), e.ID, "Save must populate generated id")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepoPG_Save_NullableFields(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	e, err := entities.NewExtracurricularEvent(entities.NewExtracurricularEventParams{
		Title:          "minimal",
		Description:    "",
		Category:       entities.CategorySports,
		TargetAudience: entities.TargetAudienceStudents,
		Location:       "",
		StartAt:        now.Add(48 * time.Hour),
		EndAt:          now.Add(50 * time.Hour),
		MaxCapacity:    nil,
		OrganizerID:    7,
		Now:            now,
	})
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO extracurricular_events")).
		WithArgs("minimal", sql.NullString{}, "sports", "students", "draft",
			sql.NullString{}, e.StartAt(), e.EndAt(), sql.NullInt64{},
			int64(7), 0, e.CreatedAt(), e.UpdatedAt()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)))

	err = repo.Save(context.Background(), e)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===== GetByID =====

func TestEventRepoPG_GetByID_HappyPathWithParticipants(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, description, category, target_audience, status")).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "category", "target_audience", "status",
			"location", "start_at", "end_at", "max_capacity", "organizer_id",
			"version", "created_at", "updated_at",
		}).AddRow(int64(99), "Концерт", sql.NullString{String: "desc", Valid: true},
			"cultural", "all", "published",
			sql.NullString{String: "loc", Valid: true},
			now, now.Add(2*time.Hour), sql.NullInt64{Int64: 50, Valid: true},
			int64(42), 1, now, now))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, registered_at FROM extracurricular_participants")).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "registered_at"}).
			AddRow(int64(101), now).
			AddRow(int64(102), now.Add(time.Minute)))

	e, err := repo.GetByID(context.Background(), 99)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.Equal(t, int64(99), e.ID)
	assert.Equal(t, "Концерт", e.Title())
	assert.Equal(t, entities.CategoryCultural, e.Category())
	assert.Equal(t, entities.StatusPublished, e.Status())
	require.NotNil(t, e.MaxCapacity())
	assert.Equal(t, 50, *e.MaxCapacity())
	parts := e.Participants()
	require.Len(t, parts, 2)
	assert.Equal(t, int64(101), parts[0].UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepoPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM extracurricular_events WHERE id = $1")).
		WithArgs(int64(404)).
		WillReturnError(sql.ErrNoRows)

	e, err := repo.GetByID(context.Background(), 404)
	assert.Nil(t, e)
	assert.True(t, errors.Is(err, repositories.ErrEventNotFound), "want ErrEventNotFound, got %v", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===== Update =====

func TestEventRepoPG_Update_HappyPath(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	e := freshEvent(t)
	e.ID = 99
	mock.ExpectExec(regexp.QuoteMeta("UPDATE extracurricular_events SET")).
		WithArgs("Концерт", sql.NullString{String: "desc", Valid: true},
			"cultural", "all", "draft",
			sql.NullString{String: "Актовый зал", Valid: true},
			e.StartAt(), e.EndAt(), sql.NullInt64{},
			e.UpdatedAt(), int64(99), 0).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), e)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepoPG_Update_VersionConflict(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	e := freshEvent(t)
	e.ID = 99
	mock.ExpectExec(regexp.QuoteMeta("UPDATE extracurricular_events SET")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM extracurricular_events WHERE id")).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}).AddRow(1))

	err := repo.Update(context.Background(), e)
	assert.True(t, errors.Is(err, repositories.ErrEventVersionConflict), "want ErrEventVersionConflict, got %v", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepoPG_Update_VanishedRow(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	e := freshEvent(t)
	e.ID = 99
	mock.ExpectExec(regexp.QuoteMeta("UPDATE extracurricular_events SET")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM extracurricular_events WHERE id")).
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)

	err := repo.Update(context.Background(), e)
	assert.True(t, errors.Is(err, repositories.ErrEventNotFound), "want ErrEventNotFound after vanished row, got %v", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===== Delete =====

func TestEventRepoPG_Delete_HappyPath(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM extracurricular_events WHERE id = $1")).
		WithArgs(int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 99))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepoPG_Delete_NotFound(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM extracurricular_events WHERE id = $1")).
		WithArgs(int64(404)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 404)
	assert.True(t, errors.Is(err, repositories.ErrEventNotFound))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===== List =====

func TestEventRepoPG_List_FilterByStatusAndAudience(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM extracurricular_events WHERE status = $1 AND target_audience = ANY($2)")).
		WithArgs("published", pq.Array([]string{"all", "students"})).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(regexp.QuoteMeta("WHERE status = $1 AND target_audience = ANY($2)")).
		WithArgs("published", pq.Array([]string{"all", "students"}), 50, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "category", "target_audience", "status", "location",
			"start_at", "end_at", "max_capacity", "organizer_id", "version",
			"created_at", "updated_at", "participant_count",
		}).AddRow(int64(1), "Event 1", "cultural", "all", "published",
			sql.NullString{}, now, now.Add(time.Hour), sql.NullInt64{},
			int64(42), 0, now, now, 3))

	result, err := repo.List(context.Background(), repositories.EventListFilter{
		Status:     "published",
		AudienceIn: []string{"all", "students"},
		Limit:      50,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	require.Len(t, result.Items, 1)
	assert.Equal(t, int64(1), result.Items[0].ID)
	assert.Equal(t, 3, result.Items[0].ParticipantCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepoPG_List_DefaultLimitWhenZero(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM extracurricular_events")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta("FROM extracurricular_events")).
		WithArgs(100, 0). // default limit=100
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "category", "target_audience", "status", "location",
			"start_at", "end_at", "max_capacity", "organizer_id", "version",
			"created_at", "updated_at", "participant_count",
		}))

	result, err := repo.List(context.Background(), repositories.EventListFilter{})
	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)
	assert.Empty(t, result.Items)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===== AddParticipant =====

func TestEventRepoPG_AddParticipant_HappyPath(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	at := time.Now()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO extracurricular_participants")).
		WithArgs(int64(99), int64(101), at).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, repo.AddParticipant(context.Background(), 99, 101, at))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepoPG_AddParticipant_UniqueViolation(t *testing.T) {
	// SQLSTATE 23505 → entities.ErrParticipantExists (domain sentinel
	// re-used by usecase для 409 mapping без re-translation).
	repo, mock := newEventRepoMock(t)
	at := time.Now()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO extracurricular_participants")).
		WithArgs(int64(99), int64(101), at).
		WillReturnError(&pq.Error{Code: "23505"})

	err := repo.AddParticipant(context.Background(), 99, 101, at)
	assert.True(t, errors.Is(err, entities.ErrParticipantExists),
		"want ErrParticipantExists on unique violation, got %v", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===== RemoveParticipant =====

func TestEventRepoPG_RemoveParticipant_HappyPath(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM extracurricular_participants WHERE event_id = $1 AND user_id = $2")).
		WithArgs(int64(99), int64(101)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.RemoveParticipant(context.Background(), 99, 101))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEventRepoPG_RemoveParticipant_NotFound(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM extracurricular_participants")).
		WithArgs(int64(99), int64(404)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.RemoveParticipant(context.Background(), 99, 404)
	assert.True(t, errors.Is(err, entities.ErrParticipantNotFound))
	assert.NoError(t, mock.ExpectationsWereMet())
}
