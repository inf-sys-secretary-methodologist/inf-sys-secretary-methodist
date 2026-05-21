package persistence

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/crypto"
)

func newUserRepoMock(t *testing.T) (*UserRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return &UserRepositoryPG{db: db}, mock
}

var userCols = []string{"id", "email", "password", "name", "role", "status", "mfa_secret", "mfa_enabled", "created_at", "updated_at"}

func TestNewUserRepositoryPG(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	repo := NewUserRepositoryPG(db)
	assert.NotNil(t, repo)
}

// --- Create ---

func TestUserRepo_Create_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	user := entities.NewUser("test@test.com", "hash", "Test", domain.RoleTeacher)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs(user.Email, user.Password, user.Name, user.Role, user.Status, user.CreatedAt, user.UpdatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
}

func TestUserRepo_Create_Error(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	user := entities.NewUser("test@test.com", "hash", "Test", domain.RoleTeacher)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users")).
		WillReturnError(fmt.Errorf("db error"))
	err := repo.Create(context.Background(), user)
	assert.Error(t, err)
}

// --- Save ---

func TestUserRepo_Save_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	user := &entities.User{ID: 1, Email: "test@test.com", Password: "hash", Name: "Test", Role: domain.RoleTeacher, Status: entities.UserStatusActive, UpdatedAt: time.Now()}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).
		WithArgs(user.Email, user.Password, user.Name, user.Role, user.Status, sql.NullString{}, false, user.UpdatedAt, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.Save(context.Background(), user)
	require.NoError(t, err)
}

func TestUserRepo_Save_NotFound(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	user := &entities.User{ID: 999, Email: "test@test.com", UpdatedAt: time.Now()}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.Save(context.Background(), user)
	assert.Error(t, err)
}

func TestUserRepo_Save_DBError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	user := &entities.User{ID: 1, UpdatedAt: time.Now()}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).
		WillReturnError(fmt.Errorf("db error"))
	err := repo.Save(context.Background(), user)
	assert.Error(t, err)
}

func TestUserRepo_Save_RowsAffectedError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	user := &entities.User{ID: 1, UpdatedAt: time.Now()}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))
	err := repo.Save(context.Background(), user)
	assert.Error(t, err)
}

// --- GetByID ---

func TestUserRepo_GetByID_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(userCols).AddRow(int64(1), "test@test.com", "hash", "Test", domain.RoleTeacher, entities.UserStatusActive, nil, false, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(int64(1)).WillReturnRows(rows)
	user, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "test@test.com", user.Email)
	assert.Equal(t, "Test", user.Name)
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByID(context.Background(), 999)
	assert.Error(t, err)
}

func TestUserRepo_GetByID_DBError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("connection error"))
	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

// --- GetByEmail ---

func TestUserRepo_GetByEmail_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(userCols).AddRow(int64(1), "test@test.com", "hash", "Test", domain.RoleTeacher, entities.UserStatusActive, nil, false, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs("test@test.com").WillReturnRows(rows)
	user, err := repo.GetByEmail(context.Background(), "test@test.com")
	require.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs("nonexistent@test.com").WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByEmail(context.Background(), "nonexistent@test.com")
	assert.Error(t, err)
}

// --- GetByEmailForAuth ---

func TestUserRepo_GetByEmailForAuth_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(userCols).AddRow(int64(1), "test@test.com", "hash", "Test", domain.RoleTeacher, entities.UserStatusActive, nil, false, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs("test@test.com").WillReturnRows(rows)
	user, err := repo.GetByEmailForAuth(context.Background(), "test@test.com")
	require.NoError(t, err)
	assert.Equal(t, "hash", user.Password)
}

// --- Delete ---

func TestUserRepo_Delete_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestUserRepo_Delete_NotFound(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.Delete(context.Background(), 999)
	assert.Error(t, err)
}

func TestUserRepo_Delete_DBError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("db error"))
	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestUserRepo_Delete_RowsAffectedError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))
	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

// --- List ---

func TestUserRepo_List_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(userCols).
		AddRow(int64(1), "a@test.com", "hash", "A", domain.RoleTeacher, entities.UserStatusActive, nil, false, now, now).
		AddRow(int64(2), "b@test.com", "hash", "B", domain.RoleStudent, entities.UserStatusActive, nil, false, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(10, 0).WillReturnRows(rows)
	users, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserRepo_List_Empty(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	rows := sqlmock.NewRows(userCols)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(10, 0).WillReturnRows(rows)
	users, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestUserRepo_List_DBError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WillReturnError(fmt.Errorf("db error"))
	_, err := repo.List(context.Background(), 10, 0)
	assert.Error(t, err)
}

func TestUserRepo_List_DefaultLimit(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	rows := sqlmock.NewRows(userCols)
	// limit <= 0 should be set to 10
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(10, 0).WillReturnRows(rows)
	_, err := repo.List(context.Background(), 0, 0)
	require.NoError(t, err)
}

func TestUserRepo_List_MaxLimit(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	rows := sqlmock.NewRows(userCols)
	// limit > 100 should be clamped to 100
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(100, 0).WillReturnRows(rows)
	_, err := repo.List(context.Background(), 200, 0)
	require.NoError(t, err)
}

func TestUserRepo_List_NegativeOffset(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	rows := sqlmock.NewRows(userCols)
	// negative offset should be set to 0
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(10, 0).WillReturnRows(rows)
	_, err := repo.List(context.Background(), 10, -5)
	require.NoError(t, err)
}

func TestUserRepo_List_ScanError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	// Return row with wrong number of columns to trigger scan error
	rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(10, 0).WillReturnRows(rows)
	_, err := repo.List(context.Background(), 10, 0)
	assert.Error(t, err)
}

// --- GetByIDForAuth (delegate to GetByID at PG layer — cache-free) ---

func TestUserRepo_GetByIDForAuth_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(userCols).AddRow(int64(7), "auth@test.com", "hash", "Auth User", domain.RoleAcademicSecretary, entities.UserStatusActive, nil, false, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(int64(7)).WillReturnRows(rows)
	user, err := repo.GetByIDForAuth(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, "hash", user.Password, "GetByIDForAuth must return password for auth flow")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByIDForAuth_NotFound(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByIDForAuth(context.Background(), 999)
	assert.Error(t, err)
}

// --- GetByEmailForAuth additional branches ---

func TestUserRepo_GetByEmailForAuth_NotFound(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs("missing@test.com").WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByEmailForAuth(context.Background(), "missing@test.com")
	assert.Error(t, err)
}

// --- scanUserByQuery: MFA secret parse error branch (corrupt persisted secret) ---

func TestUserRepo_GetByID_InvalidPersistedMFASecret(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	now := time.Now()
	// Persisted MFA secret with wrong length triggers entities.NewMFASecret error
	// inside scanUserByQuery, which wraps with "invalid persisted MFA secret".
	rows := sqlmock.NewRows(userCols).
		AddRow(int64(1), "test@test.com", "hash", "T", domain.RoleTeacher, entities.UserStatusActive, "TOOSHORT", false, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(int64(1)).WillReturnRows(rows)
	_, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid persisted MFA secret")
}

// --- List: rows.Err() after iteration + invalid MFA in row ---

func TestUserRepo_List_RowsErrAfterIteration(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	now := time.Now()
	// First row scans OK, then rows.Err() returns failure (driver-level).
	rows := sqlmock.NewRows(userCols).
		AddRow(int64(1), "a@test.com", "hash", "A", domain.RoleTeacher, entities.UserStatusActive, nil, false, now, now).
		RowError(0, fmt.Errorf("connection reset during iteration"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(10, 0).WillReturnRows(rows)
	_, err := repo.List(context.Background(), 10, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection reset")
}

func TestUserRepo_List_InvalidPersistedMFASecret(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(userCols).
		AddRow(int64(1), "a@test.com", "hash", "A", domain.RoleTeacher, entities.UserStatusActive, "BROKEN", false, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, created_at, updated_at")).
		WithArgs(10, 0).WillReturnRows(rows)
	_, err := repo.List(context.Background(), 10, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid persisted MFA secret")
}

// userColsV0159 mirrors userCols but appends the mfa_secret_encrypted
// boolean column introduced by migration 045 (v0.159.0 ADR-4). The
// GREEN pair updates all production SELECT statements to include the
// column so this slice is the canonical row shape going forward.
var userColsV0159 = []string{
	"id", "email", "password", "name", "role", "status",
	"mfa_secret", "mfa_enabled", "mfa_secret_encrypted",
	"created_at", "updated_at",
}

// genTestKEK returns a deterministic-looking but random 32-byte KEK
// for use across the ADR-4b sqlmock tests.
func genTestKEK(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return key
}

// TestUserRepo_MFASecretEncryptedAtRest pins v0.159.0 ADR-4b: when a
// KEK is wired via WithMFASecretKEK, Save encrypts the MFA secret
// before INSERT/UPDATE and marks mfa_secret_encrypted=TRUE; scan paths
// decrypt rows whose encrypted flag is TRUE and treat rows with FALSE
// as legacy plaintext (lazy migration). Issue #279.
func TestUserRepo_MFASecretEncryptedAtRest(t *testing.T) {
	const plaintextSecret = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"
	key := genTestKEK(t)

	t.Run("Save with KEK writes ciphertext + encrypted=true", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		repo = repo.WithMFASecretKEK(key)

		secret, err := entities.NewMFASecret(plaintextSecret)
		require.NoError(t, err)
		now := time.Now()
		user := &entities.User{
			ID: 1, Email: "x@x", Password: "h", Name: "X",
			Role: domain.RoleTeacher, Status: entities.UserStatusActive,
			MFASecret: &secret, MFAEnabled: true,
			UpdatedAt: now,
		}

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).
			WithArgs(
				user.Email, user.Password, user.Name, user.Role, user.Status,
				sqlmock.AnyArg(), // ciphertext (non-deterministic nonce)
				true,             // mfa_enabled
				true,             // mfa_secret_encrypted = TRUE
				user.UpdatedAt, user.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.Save(context.Background(), user)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet(), "Save with KEK must include mfa_secret_encrypted=true and ciphertext arg")
	})

	t.Run("Scan with KEK + encrypted=true row returns decrypted plaintext", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		repo = repo.WithMFASecretKEK(key)

		// Seed the row with a ciphertext produced by the same KEK so
		// the scan path must decrypt it back to the canonical plaintext.
		ct, err := crypto.EncryptString(plaintextSecret, key)
		require.NoError(t, err)
		now := time.Now()
		rows := sqlmock.NewRows(userColsV0159).
			AddRow(int64(42), "x@x", "h", "X", domain.RoleTeacher, entities.UserStatusActive,
				ct, true, true, now, now)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, mfa_secret_encrypted, created_at, updated_at")).
			WithArgs(int64(42)).WillReturnRows(rows)

		got, err := repo.GetByID(context.Background(), 42)
		require.NoError(t, err)
		require.NotNil(t, got.MFASecret)
		assert.Equal(t, plaintextSecret, got.MFASecret.String(), "MFA secret on the entity must be the decrypted plaintext")
	})

	t.Run("Scan with KEK + encrypted=false row preserves legacy plaintext (lazy migration)", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		repo = repo.WithMFASecretKEK(key)

		now := time.Now()
		rows := sqlmock.NewRows(userColsV0159).
			AddRow(int64(7), "y@y", "h", "Y", domain.RoleTeacher, entities.UserStatusActive,
				plaintextSecret, true, false, now, now)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, mfa_secret, mfa_enabled, mfa_secret_encrypted, created_at, updated_at")).
			WithArgs(int64(7)).WillReturnRows(rows)

		got, err := repo.GetByID(context.Background(), 7)
		require.NoError(t, err)
		require.NotNil(t, got.MFASecret)
		assert.Equal(t, plaintextSecret, got.MFASecret.String(), "legacy plaintext row must scan through unchanged")
	})

	t.Run("Save without KEK falls back to plaintext + encrypted=false", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		// No WithMFASecretKEK call — KEK stays nil

		secret, err := entities.NewMFASecret(plaintextSecret)
		require.NoError(t, err)
		now := time.Now()
		user := &entities.User{
			ID: 1, Email: "x@x", Password: "h", Name: "X",
			Role: domain.RoleTeacher, Status: entities.UserStatusActive,
			MFASecret: &secret, MFAEnabled: true,
			UpdatedAt: now,
		}

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).
			WithArgs(
				user.Email, user.Password, user.Name, user.Role, user.Status,
				sql.NullString{String: plaintextSecret, Valid: true},
				true,  // mfa_enabled
				false, // mfa_secret_encrypted = FALSE (no KEK)
				user.UpdatedAt, user.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.Save(context.Background(), user)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet(), "Save without KEK must store plaintext and encrypted=false")
	})
}
