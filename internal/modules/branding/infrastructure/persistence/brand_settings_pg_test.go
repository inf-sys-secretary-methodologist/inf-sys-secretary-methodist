package persistence

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/entities"
)

func newRepoMock(t *testing.T) (*BrandSettingsRepositoryPG, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	repo := NewBrandSettingsRepositoryPG(db)
	return repo, mock, func() { _ = db.Close() }
}

func TestGet_ReturnsHydratedRow(t *testing.T) {
	repo, mock, cleanup := newRepoMock(t)
	defer cleanup()

	now := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"app_name", "tagline", "logo_url", "favicon_url",
		"primary_color", "secondary_color", "updated_at",
	}).AddRow(
		"Loaded Brand",
		"loaded tagline",
		"https://loaded.example/logo.png",
		"https://loaded.example/favicon.ico",
		"#112233",
		"#445566",
		now,
	)
	mock.ExpectQuery(`SELECT app_name, tagline, logo_url, favicon_url, primary_color, secondary_color, updated_at FROM brand_settings WHERE id = 1`).
		WillReturnRows(rows)

	bs, err := repo.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, bs)
	assert.Equal(t, "Loaded Brand", bs.AppName())
	assert.Equal(t, "loaded tagline", bs.Tagline())
	assert.Equal(t, "https://loaded.example/logo.png", bs.LogoURL())
	assert.Equal(t, "https://loaded.example/favicon.ico", bs.FaviconURL())
	assert.Equal(t, "#112233", bs.PrimaryColor())
	assert.Equal(t, "#445566", bs.SecondaryColor())
	assert.Equal(t, now, bs.UpdatedAt())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGet_MissingRow_ReturnsErrBrandSettingsMissing(t *testing.T) {
	repo, mock, cleanup := newRepoMock(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"app_name", "tagline", "logo_url", "favicon_url",
		"primary_color", "secondary_color", "updated_at",
	})
	mock.ExpectQuery(`SELECT app_name, tagline, logo_url, favicon_url, primary_color, secondary_color, updated_at FROM brand_settings WHERE id = 1`).
		WillReturnRows(rows)

	_, err := repo.Get(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrBrandSettingsMissing),
		"missing seed row → ErrBrandSettingsMissing, got %v", err)
}

func TestUpdate_HappyPath_PinsAllSixFields(t *testing.T) {
	repo, mock, cleanup := newRepoMock(t)
	defer cleanup()

	now := time.Date(2026, 5, 14, 11, 0, 0, 0, time.UTC)
	bs, err := entities.NewBrandSettings(
		"My System",
		"Welcome to my system",
		"https://example.com/logo.png",
		"https://example.com/favicon.ico",
		"#ABCDEF",
		"#012345",
		now,
	)
	require.NoError(t, err)

	mock.ExpectExec(`UPDATE brand_settings SET app_name = $1, tagline = $2, logo_url = $3, favicon_url = $4, primary_color = $5, secondary_color = $6, updated_at = $7 WHERE id = 1`).
		WithArgs(
			"My System",
			"Welcome to my system",
			"https://example.com/logo.png",
			"https://example.com/favicon.ico",
			"#ABCDEF",
			"#012345",
			now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(context.Background(), bs)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_RowsAffectedZero_ReturnsErrBrandSettingsMissing(t *testing.T) {
	repo, mock, cleanup := newRepoMock(t)
	defer cleanup()

	now := time.Date(2026, 5, 14, 11, 0, 0, 0, time.UTC)
	bs, err := entities.NewBrandSettings("ok", "", "", "", "", "", now)
	require.NoError(t, err)

	mock.ExpectExec(`UPDATE brand_settings SET app_name = $1, tagline = $2, logo_url = $3, favicon_url = $4, primary_color = $5, secondary_color = $6, updated_at = $7 WHERE id = 1`).
		WithArgs("ok", "", "", "", "", "", now).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(context.Background(), bs)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrBrandSettingsMissing),
		"RowsAffected=0 → ErrBrandSettingsMissing, got %v", err)
}

func TestGet_DBError_WrappedNotSentinel(t *testing.T) {
	repo, mock, cleanup := newRepoMock(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT app_name, tagline, logo_url, favicon_url, primary_color, secondary_color, updated_at FROM brand_settings WHERE id = 1`).
		WillReturnError(fmt.Errorf("connection closed"))

	_, err := repo.Get(context.Background())
	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrBrandSettingsMissing),
		"non-NoRows DB error must NOT collapse to ErrBrandSettingsMissing")
	assert.Contains(t, err.Error(), "failed to read settings")
}

func TestUpdate_ExecError_Wrapped(t *testing.T) {
	repo, mock, cleanup := newRepoMock(t)
	defer cleanup()

	now := time.Date(2026, 5, 14, 11, 0, 0, 0, time.UTC)
	bs, err := entities.NewBrandSettings("ok", "", "", "", "", "", now)
	require.NoError(t, err)

	mock.ExpectExec(`UPDATE brand_settings SET app_name = $1, tagline = $2, logo_url = $3, favicon_url = $4, primary_color = $5, secondary_color = $6, updated_at = $7 WHERE id = 1`).
		WithArgs("ok", "", "", "", "", "", now).
		WillReturnError(fmt.Errorf("constraint violation"))

	err = repo.Update(context.Background(), bs)
	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrBrandSettingsMissing))
	assert.Contains(t, err.Error(), "failed to update settings")
}

func TestUpdate_RowsAffectedError_Wrapped(t *testing.T) {
	repo, mock, cleanup := newRepoMock(t)
	defer cleanup()

	now := time.Date(2026, 5, 14, 11, 0, 0, 0, time.UTC)
	bs, err := entities.NewBrandSettings("ok", "", "", "", "", "", now)
	require.NoError(t, err)

	failingResult := sqlmock.NewErrorResult(fmt.Errorf("driver does not support RowsAffected"))
	mock.ExpectExec(`UPDATE brand_settings SET app_name = $1, tagline = $2, logo_url = $3, favicon_url = $4, primary_color = $5, secondary_color = $6, updated_at = $7 WHERE id = 1`).
		WithArgs("ok", "", "", "", "", "", now).
		WillReturnResult(failingResult)

	err = repo.Update(context.Background(), bs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to inspect update result")
}
