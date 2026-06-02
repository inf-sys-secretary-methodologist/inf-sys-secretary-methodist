package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/entities"
)

// selectCols mirrors the column order used by SELECT in announcement queries.
var selectCols = []string{
	"id", "title", "content", "summary", "author_id", "status", "priority",
	"target_audience", "publish_at", "expire_at", "is_pinned", "view_count",
	"tags", "metadata", "created_at", "updated_at",
}

// attachmentCols mirrors the column order used by SELECT on attachments.
var attachmentCols = []string{
	"id", "announcement_id", "file_name", "file_path", "file_size",
	"mime_type", "uploaded_by", "created_at",
}

func newAnnouncementRepoMock(t *testing.T) (*AnnouncementRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewAnnouncementRepositoryPG(db), mock
}

func sampleAnnouncementRow(id int64, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(selectCols).AddRow(
		id, "Title", "Content", nil, int64(1), "published", "normal",
		"all", nil, nil, false, int64(0),
		pq.StringArray{"news"}, []byte(`{"k":"v"}`), now, now,
	)
}

// --- Create ---

func TestAnnouncementRepositoryPG_Create_Success(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()
	announcement := &entities.Announcement{
		Title: "Title", Content: "Content", AuthorID: 1,
		Status: domain.AnnouncementStatusDraft, Priority: domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll, IsPinned: false, ViewCount: 0,
		Tags: []string{"news"}, Metadata: json.RawMessage(`{}`),
		CreatedAt: now, UpdatedAt: now,
	}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO announcements")).
		WithArgs(
			"Title", "Content", (*string)(nil), int64(1),
			domain.AnnouncementStatusDraft, domain.AnnouncementPriorityNormal,
			domain.TargetAudienceAll, (*time.Time)(nil), (*time.Time)(nil),
			false, int64(0), pq.Array([]string{"news"}), json.RawMessage(`{}`),
			now, now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(42)))

	err := repo.Create(context.Background(), announcement)
	require.NoError(t, err)
	assert.Equal(t, int64(42), announcement.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_Create_DBError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()
	announcement := &entities.Announcement{
		Title: "T", Content: "C", AuthorID: 1,
		Status: domain.AnnouncementStatusDraft, Priority: domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll, Tags: []string{}, Metadata: json.RawMessage(`{}`),
		CreatedAt: now, UpdatedAt: now,
	}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO announcements")).
		WillReturnError(errors.New("db down"))

	err := repo.Create(context.Background(), announcement)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

// --- Save ---

func TestAnnouncementRepositoryPG_Save_Success(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()
	announcement := &entities.Announcement{
		ID: 42, Title: "Updated", Content: "Updated body", AuthorID: 1,
		Status: domain.AnnouncementStatusPublished, Priority: domain.AnnouncementPriorityHigh,
		TargetAudience: domain.TargetAudienceStudents, IsPinned: true,
		Tags: []string{"urgent"}, Metadata: json.RawMessage(`{}`), UpdatedAt: now,
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE announcements SET")).
		WithArgs(
			"Updated", "Updated body", (*string)(nil),
			domain.AnnouncementStatusPublished, domain.AnnouncementPriorityHigh,
			domain.TargetAudienceStudents, (*time.Time)(nil), (*time.Time)(nil),
			true, pq.Array([]string{"urgent"}), json.RawMessage(`{}`), now, int64(42),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Save(context.Background(), announcement))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_Save_DBError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	announcement := &entities.Announcement{
		ID: 42, Status: domain.AnnouncementStatusDraft, Priority: domain.AnnouncementPriorityNormal,
		TargetAudience: domain.TargetAudienceAll, Tags: []string{}, Metadata: json.RawMessage(`{}`),
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE announcements SET")).
		WillReturnError(errors.New("constraint violation"))

	err := repo.Save(context.Background(), announcement)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "constraint violation")
}

// --- GetByID ---

func TestAnnouncementRepositoryPG_GetByID_Found(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("FROM announcements WHERE id = $1")).
		WithArgs(int64(42)).
		WillReturnRows(sampleAnnouncementRow(42, now))

	result, err := repo.GetByID(context.Background(), 42)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(42), result.ID)
	assert.Equal(t, "Title", result.Title)
	assert.Equal(t, []string{"news"}, result.Tags)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcements WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestAnnouncementRepositoryPG_GetByID_QueryError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcements WHERE id = $1")).
		WithArgs(int64(42)).
		WillReturnError(errors.New("conn closed"))

	result, err := repo.GetByID(context.Background(), 42)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "conn closed")
}

// --- Delete ---

func TestAnnouncementRepositoryPG_Delete_Success(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM announcements WHERE id = $1")).
		WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 42))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_Delete_DBError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM announcements WHERE id = $1")).
		WithArgs(int64(42)).
		WillReturnError(errors.New("fk violation"))

	err := repo.Delete(context.Background(), 42)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fk violation")
}

// --- List & Count (table-driven over filter dimensions) ---

func TestAnnouncementRepositoryPG_List_NoFilter(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()

	rows := sqlmock.NewRows(selectCols).
		AddRow(int64(1), "A", "ca", nil, int64(1), "draft", "normal", "all", nil, nil, false, int64(0), pq.StringArray{}, []byte("{}"), now, now).
		AddRow(int64(2), "B", "cb", nil, int64(1), "published", "high", "all", nil, nil, true, int64(5), pq.StringArray{"news"}, []byte("{}"), now, now)

	mock.ExpectQuery(regexp.QuoteMeta("FROM announcements ORDER BY is_pinned DESC, publish_at DESC NULLS LAST, created_at DESC LIMIT $1 OFFSET $2")).
		WithArgs(10, 0).
		WillReturnRows(rows)

	result, err := repo.List(context.Background(), usecases.AnnouncementFilter{}, 10, 0)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_List_AllScalarFilters(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()
	authorID := int64(7)
	status := domain.AnnouncementStatusPublished
	priority := domain.AnnouncementPriorityHigh
	audience := domain.TargetAudienceStudents
	pinned := true
	search := "midterm"

	rows := sqlmock.NewRows(selectCols).AddRow(
		int64(1), "T", "C", nil, authorID, status, priority,
		audience, nil, nil, pinned, int64(0), pq.StringArray{"alpha"}, []byte("{}"), now, now,
	)

	// All scalar conditions: AuthorID($1), Status($2), Priority($3), TargetAudience($4),
	// IsPinned($5), Search($6,$7), Tags($8), LIMIT $9 OFFSET $10
	mock.ExpectQuery(regexp.QuoteMeta(
		"WHERE author_id = $1 AND status = $2 AND priority = $3 AND "+
			"(target_audience = $4 OR target_audience = 'all') AND is_pinned = $5 AND "+
			"(title ILIKE $6 OR content ILIKE $7) AND tags && $8 "+
			"ORDER BY is_pinned DESC, publish_at DESC NULLS LAST, created_at DESC LIMIT $9 OFFSET $10",
	)).
		WithArgs(authorID, status, priority, audience, pinned, "%midterm%", "%midterm%", pq.Array([]string{"alpha"}), 25, 0).
		WillReturnRows(rows)

	result, err := repo.List(context.Background(), usecases.AnnouncementFilter{
		AuthorID:       &authorID,
		Status:         &status,
		Priority:       &priority,
		TargetAudience: &audience,
		IsPinned:       &pinned,
		Search:         &search,
		Tags:           []string{"alpha"},
	}, 25, 0)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_List_IsExpiredBranches(t *testing.T) {
	cases := []struct {
		name     string
		expired  bool
		fragment string
	}{
		{"expired=true", true, "expire_at IS NOT NULL AND expire_at < NOW()"},
		{"expired=false", false, "(expire_at IS NULL OR expire_at >= NOW())"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newAnnouncementRepoMock(t)
			expired := tc.expired
			mock.ExpectQuery(regexp.QuoteMeta("WHERE "+tc.fragment)).
				WithArgs(10, 0).
				WillReturnRows(sqlmock.NewRows(selectCols))

			result, err := repo.List(context.Background(), usecases.AnnouncementFilter{IsExpired: &expired}, 10, 0)
			require.NoError(t, err)
			assert.Empty(t, result)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAnnouncementRepositoryPG_List_QueryError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcements")).
		WillReturnError(errors.New("query failed"))

	result, err := repo.List(context.Background(), usecases.AnnouncementFilter{}, 10, 0)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestAnnouncementRepositoryPG_List_ScanError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	// Wrong column count → scan failure.
	badRows := sqlmock.NewRows([]string{"id"}).AddRow(int64(1))
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcements")).
		WillReturnRows(badRows)

	result, err := repo.List(context.Background(), usecases.AnnouncementFilter{}, 10, 0)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestAnnouncementRepositoryPG_Count_NoFilter(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM announcements")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(99)))

	count, err := repo.Count(context.Background(), usecases.AnnouncementFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(99), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_Count_WithStatusFilter(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	status := domain.AnnouncementStatusPublished

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM announcements WHERE status = $1")).
		WithArgs(status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(7)))

	count, err := repo.Count(context.Background(), usecases.AnnouncementFilter{Status: &status})
	require.NoError(t, err)
	assert.Equal(t, int64(7), count)
}

func TestAnnouncementRepositoryPG_Count_QueryError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM announcements")).
		WillReturnError(errors.New("count failed"))

	count, err := repo.Count(context.Background(), usecases.AnnouncementFilter{})
	require.Error(t, err)
	assert.Equal(t, int64(0), count)
}

// --- GetByAuthor ---

func TestAnnouncementRepositoryPG_GetByAuthor_Success(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()
	authorID := int64(7)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE author_id = $1 ORDER BY is_pinned DESC, publish_at DESC NULLS LAST, created_at DESC LIMIT $2 OFFSET $3")).
		WithArgs(authorID, 5, 0).
		WillReturnRows(sampleAnnouncementRow(1, now))

	result, err := repo.GetByAuthor(context.Background(), authorID, 5, 0)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetPublished ---

func TestAnnouncementRepositoryPG_GetPublished_Success(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()
	audience := domain.TargetAudienceTeachers

	mock.ExpectQuery(regexp.QuoteMeta(
		"WHERE status = $1 AND (target_audience = $2 OR target_audience = 'all') "+
			"AND (publish_at IS NULL OR publish_at <= NOW()) AND (expire_at IS NULL OR expire_at > NOW()) "+
			"ORDER BY is_pinned DESC, publish_at DESC NULLS LAST, created_at DESC LIMIT $3 OFFSET $4",
	)).
		WithArgs(domain.AnnouncementStatusPublished, audience, 10, 0).
		WillReturnRows(sampleAnnouncementRow(1, now))

	result, err := repo.GetPublished(context.Background(), audience, 10, 0)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_GetPublished_QueryError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE status = $1")).
		WillReturnError(errors.New("conn lost"))

	result, err := repo.GetPublished(context.Background(), domain.TargetAudienceAll, 10, 0)
	require.Error(t, err)
	assert.Nil(t, result)
}

// --- GetPinned ---

func TestAnnouncementRepositoryPG_GetPinned_Success(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()
	audiences := []domain.TargetAudience{domain.TargetAudienceAll, domain.TargetAudienceStudents}

	mock.ExpectQuery(regexp.QuoteMeta(
		"WHERE is_pinned = true AND status = 'published' "+
			"AND target_audience = ANY($1) "+
			"AND (publish_at IS NULL OR publish_at <= NOW()) AND (expire_at IS NULL OR expire_at > NOW()) "+
			"ORDER BY priority DESC, created_at DESC LIMIT $2",
	)).
		WithArgs(pq.StringArray{"all", "students"}, 5).
		WillReturnRows(sampleAnnouncementRow(1, now))

	result, err := repo.GetPinned(context.Background(), audiences, 5)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_GetPinned_QueryError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_pinned = true")).
		WillReturnError(errors.New("timeout"))

	result, err := repo.GetPinned(context.Background(), []domain.TargetAudience{domain.TargetAudienceAll}, 5)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestAnnouncementRepositoryPG_GetPinned_EmptyAudiences(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	// Empty audiences slice → zero rows without touching the DB.
	// The repo refuses к build the query, defending against a caller
	// that passed an empty list (e.g. unknown role mishandled upstream).

	result, err := repo.GetPinned(context.Background(), nil, 5)
	require.NoError(t, err)
	assert.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetRecent ---

func TestAnnouncementRepositoryPG_GetRecent_Success(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()
	audiences := []domain.TargetAudience{domain.TargetAudienceAll, domain.TargetAudienceTeachers}

	mock.ExpectQuery(regexp.QuoteMeta(
		"WHERE status = 'published' "+
			"AND target_audience = ANY($1) "+
			"AND (publish_at IS NULL OR publish_at <= NOW()) AND (expire_at IS NULL OR expire_at > NOW()) "+
			"ORDER BY publish_at DESC NULLS LAST, created_at DESC LIMIT $2",
	)).
		WithArgs(pq.StringArray{"all", "teachers"}, 20).
		WillReturnRows(sampleAnnouncementRow(1, now))

	result, err := repo.GetRecent(context.Background(), audiences, 20)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_GetRecent_QueryError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE status = 'published'")).
		WillReturnError(errors.New("dead"))

	result, err := repo.GetRecent(context.Background(), []domain.TargetAudience{domain.TargetAudienceAll}, 20)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestAnnouncementRepositoryPG_GetRecent_EmptyAudiences(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)

	result, err := repo.GetRecent(context.Background(), nil, 20)
	require.NoError(t, err)
	assert.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByIDForAudience ---

func TestAnnouncementRepositoryPG_GetByIDForAudience_Found(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()
	audiences := []domain.TargetAudience{domain.TargetAudienceAll, domain.TargetAudienceStudents}

	mock.ExpectQuery(regexp.QuoteMeta(
		"FROM announcements WHERE id = $1 AND target_audience = ANY($2)",
	)).
		WithArgs(int64(42), pq.StringArray{"all", "students"}).
		WillReturnRows(sampleAnnouncementRow(42, now))

	result, err := repo.GetByIDForAudience(context.Background(), 42, audiences)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(42), result.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_GetByIDForAudience_OutsideAudience(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	audiences := []domain.TargetAudience{domain.TargetAudienceStudents}

	mock.ExpectQuery(regexp.QuoteMeta(
		"FROM announcements WHERE id = $1 AND target_audience = ANY($2)",
	)).
		WithArgs(int64(42), pq.StringArray{"students"}).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByIDForAudience(context.Background(), 42, audiences)
	require.NoError(t, err)
	assert.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_GetByIDForAudience_EmptyAudiences(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)

	result, err := repo.GetByIDForAudience(context.Background(), 42, nil)
	require.NoError(t, err)
	assert.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_GetByIDForAudience_QueryError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcements WHERE id = $1")).
		WithArgs(int64(42), pq.StringArray{"all"}).
		WillReturnError(errors.New("db down"))

	result, err := repo.GetByIDForAudience(context.Background(), 42, []domain.TargetAudience{domain.TargetAudienceAll})
	require.Error(t, err)
	assert.Nil(t, result)
}

// --- IncrementViewCount ---

func TestAnnouncementRepositoryPG_IncrementViewCount_Success(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE announcements SET view_count = view_count + 1 WHERE id = $1")).
		WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.IncrementViewCount(context.Background(), 42))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_IncrementViewCount_DBError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE announcements SET view_count")).
		WillReturnError(errors.New("lock timeout"))

	err := repo.IncrementViewCount(context.Background(), 42)
	require.Error(t, err)
}

// --- AddAttachment ---

func TestAnnouncementRepositoryPG_AddAttachment_Success(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()
	att := &entities.AnnouncementAttachment{
		AnnouncementID: 42, FileName: "report.pdf", FilePath: "/uploads/report.pdf",
		FileSize: 1024, MimeType: "application/pdf", UploadedBy: 1, CreatedAt: now,
	}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO announcement_attachments")).
		WithArgs(int64(42), "report.pdf", "/uploads/report.pdf", int64(1024), "application/pdf", int64(1), now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(7)))

	require.NoError(t, repo.AddAttachment(context.Background(), att))
	assert.Equal(t, int64(7), att.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_AddAttachment_DBError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	att := &entities.AnnouncementAttachment{AnnouncementID: 42}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO announcement_attachments")).
		WillReturnError(errors.New("fk fail"))

	err := repo.AddAttachment(context.Background(), att)
	require.Error(t, err)
}

// --- RemoveAttachment ---

func TestAnnouncementRepositoryPG_RemoveAttachment_Success(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM announcement_attachments WHERE id = $1")).
		WithArgs(int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.RemoveAttachment(context.Background(), 7))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_RemoveAttachment_DBError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM announcement_attachments")).
		WillReturnError(errors.New("no row"))

	err := repo.RemoveAttachment(context.Background(), 7)
	require.Error(t, err)
}

// --- GetAttachments ---

func TestAnnouncementRepositoryPG_GetAttachments_Success(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()

	rows := sqlmock.NewRows(attachmentCols).
		AddRow(int64(1), int64(42), "a.pdf", "/p/a.pdf", int64(100), "application/pdf", int64(1), now).
		AddRow(int64(2), int64(42), "b.jpg", "/p/b.jpg", int64(200), "image/jpeg", int64(1), now)

	mock.ExpectQuery(regexp.QuoteMeta("FROM announcement_attachments WHERE announcement_id = $1")).
		WithArgs(int64(42)).
		WillReturnRows(rows)

	result, err := repo.GetAttachments(context.Background(), 42)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "a.pdf", result[0].FileName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_GetAttachments_Empty(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcement_attachments WHERE announcement_id = $1")).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows(attachmentCols))

	result, err := repo.GetAttachments(context.Background(), 99)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestAnnouncementRepositoryPG_GetAttachments_QueryError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcement_attachments WHERE announcement_id = $1")).
		WillReturnError(errors.New("conn closed"))

	result, err := repo.GetAttachments(context.Background(), 42)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestAnnouncementRepositoryPG_GetAttachments_ScanError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	// Wrong column count → scan failure.
	badRows := sqlmock.NewRows([]string{"id"}).AddRow(int64(1))
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcement_attachments WHERE announcement_id = $1")).
		WithArgs(int64(42)).
		WillReturnRows(badRows)

	result, err := repo.GetAttachments(context.Background(), 42)
	require.Error(t, err)
	assert.Nil(t, result)
}

// --- GetAttachmentByID ---

func TestAnnouncementRepositoryPG_GetAttachmentByID_Found(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()

	rows := sqlmock.NewRows(attachmentCols).AddRow(
		int64(7), int64(42), "doc.pdf", "/p/doc.pdf", int64(500), "application/pdf", int64(1), now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcement_attachments\n\t\t\tWHERE id = $1")).
		WithArgs(int64(7)).
		WillReturnRows(rows)

	result, err := repo.GetAttachmentByID(context.Background(), 7)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "doc.pdf", result.FileName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_GetAttachmentByID_NotFound(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcement_attachments")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetAttachmentByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestAnnouncementRepositoryPG_GetAttachmentByID_QueryError(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcement_attachments")).
		WithArgs(int64(7)).
		WillReturnError(errors.New("conn closed"))

	result, err := repo.GetAttachmentByID(context.Background(), 7)
	require.Error(t, err)
	assert.Nil(t, result)
}

// --- NULL jsonb metadata regression (real PG returns NULL for unset
// metadata; scanning NULL into json.RawMessage fails unless the scan goes
// through a []byte intermediate). sqlmock reproduces the real driver path. ---

func TestAnnouncementRepositoryPG_GetByID_NullMetadata(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(selectCols).AddRow(
		int64(42), "Title", "Content", nil, int64(1), "published", "normal",
		"all", nil, nil, false, int64(0),
		pq.StringArray{}, nil, now, now, // metadata = NULL
	)
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcements WHERE id = $1")).
		WithArgs(int64(42)).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), 42)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Metadata)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAnnouncementRepositoryPG_List_NullMetadata(t *testing.T) {
	repo, mock := newAnnouncementRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(selectCols).AddRow(
		int64(1), "A", "ca", nil, int64(1), "published", "normal", "all",
		nil, nil, false, int64(0), pq.StringArray{}, nil, now, now, // metadata = NULL
	)
	mock.ExpectQuery(regexp.QuoteMeta("FROM announcements")).
		WillReturnRows(rows)

	result, err := repo.List(context.Background(), usecases.AnnouncementFilter{}, 10, 0)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Empty(t, result[0].Metadata)
	require.NoError(t, mock.ExpectationsWereMet())
}
