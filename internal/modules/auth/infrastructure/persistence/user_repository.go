package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/crypto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/database"
)

// UserRepositoryPG implements PostgreSQL user repository
type UserRepositoryPG struct {
	db           *sql.DB
	mfaSecretKEK []byte // nil → no at-rest encryption (legacy / dev fallback)
}

// NewUserRepositoryPG creates a new PostgreSQL user repository.
// Returns the concrete *UserRepositoryPG so callers can chain
// the WithMFASecretKEK setter for at-rest encryption (v0.159.0
// ADR-4). The concrete type satisfies usecases.UserRepository
// structurally.
func NewUserRepositoryPG(db *sql.DB) *UserRepositoryPG {
	return &UserRepositoryPG{db: db}
}

// Compile-time guard that the concrete type still satisfies the
// usecase-facing interface — drift would break callers that take
// usecases.UserRepository (cmd/server/main.go, tests).
var _ usecases.UserRepository = (*UserRepositoryPG)(nil)

// WithMFASecretKEK wires the at-rest KEK (32-byte AES-256 key) used to
// encrypt users.mfa_secret on write and decrypt on read. Without a KEK,
// the repository preserves the legacy plaintext behavior (dev / test
// convenience); production deployments MUST attach a KEK so DB dumps do
// not expose TOTP shared secrets. Returns the receiver so callers can
// chain after NewUserRepositoryPG. Issue #279 ADR-4.
func (r *UserRepositoryPG) WithMFASecretKEK(key []byte) *UserRepositoryPG {
	r.mfaSecretKEK = key
	return r
}

// wrapMFASecret encodes the MFA secret for persistence: when a KEK is
// wired, returns (ciphertext, encrypted=true); otherwise pass-through
// (plaintext, encrypted=false). NULL when the user has no MFA secret.
func (r *UserRepositoryPG) wrapMFASecret(user *entities.User) (sql.NullString, bool, error) {
	if user.MFASecret == nil {
		return sql.NullString{}, false, nil
	}
	plaintext := user.MFASecret.String()
	if len(r.mfaSecretKEK) == 0 {
		return sql.NullString{String: plaintext, Valid: true}, false, nil
	}
	ct, err := crypto.EncryptString(plaintext, r.mfaSecretKEK)
	if err != nil {
		return sql.NullString{}, false, fmt.Errorf("repository: encrypt MFA secret for user %d: %w", user.ID, err)
	}
	return sql.NullString{String: ct, Valid: true}, true, nil
}

// unwrapMFASecret decodes a persisted (mfa_secret, encrypted) pair back
// to the typed VO. When encrypted=true, decrypts via the wired KEK;
// when encrypted=false (legacy plaintext rows), passes the value
// through so a deployment can migrate lazily by rewrapping on the next
// Save under the KEK.
func (r *UserRepositoryPG) unwrapMFASecret(stored sql.NullString, encrypted bool, userID int64) (*entities.MFASecret, error) {
	if !stored.Valid || stored.String == "" {
		return nil, nil
	}
	raw := stored.String
	if encrypted {
		if len(r.mfaSecretKEK) == 0 {
			return nil, fmt.Errorf("repository: user %d row marked encrypted but no KEK configured", userID)
		}
		plaintext, err := crypto.DecryptString(raw, r.mfaSecretKEK)
		if err != nil {
			return nil, fmt.Errorf("repository: decrypt MFA secret for user %d: %w", userID, err)
		}
		raw = plaintext
	}
	secret, err := entities.NewMFASecret(raw)
	if err != nil {
		return nil, fmt.Errorf("repository: invalid persisted MFA secret for user %d: %w", userID, err)
	}
	return &secret, nil
}

// Create creates a new user in the database
func (r *UserRepositoryPG) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (email, password, name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Password,
		user.Name,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return database.MapPostgresError(err)
	}
	return nil
}

// Save updates an existing user, including MFA enrollment fields.
// v0.159.0 ADR-4: when WithMFASecretKEK is wired, the MFA secret is
// AES-256-GCM encrypted before writing and the mfa_secret_encrypted
// column is set to TRUE. Without a KEK the legacy plaintext shape is
// preserved (encrypted=FALSE) so dev / test deployments stay
// compatible.
func (r *UserRepositoryPG) Save(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users
		SET email = $1, password = $2, name = $3, role = $4, status = $5,
		    mfa_secret = $6, mfa_enabled = $7, mfa_secret_encrypted = $8, updated_at = $9
		WHERE id = $10
	`
	mfaSecret, encrypted, err := r.wrapMFASecret(user)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Password,
		user.Name,
		user.Role,
		user.Status,
		mfaSecret,
		user.MFAEnabled,
		encrypted,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return database.MapPostgresError(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return database.MapPostgresError(err)
	}

	if rows == 0 {
		return database.MapPostgresError(sql.ErrNoRows)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepositoryPG) GetByID(ctx context.Context, userID int64) (*entities.User, error) {
	return r.scanUserByQuery(ctx, `
		SELECT id, email, password, name, role, status,
		       mfa_secret, mfa_enabled, mfa_secret_encrypted, created_at, updated_at
		FROM users
		WHERE id = $1
	`, userID)
}

// GetByEmail retrieves a user by email
func (r *UserRepositoryPG) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	return r.scanUserByQuery(ctx, `
		SELECT id, email, password, name, role, status,
		       mfa_secret, mfa_enabled, mfa_secret_encrypted, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email)
}

// GetByEmailForAuth retrieves a user by email for authentication
// In PG implementation, this is the same as GetByEmail (no caching at this level)
func (r *UserRepositoryPG) GetByEmailForAuth(ctx context.Context, email string) (*entities.User, error) {
	return r.GetByEmail(ctx, email)
}

// GetByIDForAuth retrieves a user by ID bypassing cache; PG layer is already
// cache-free, so this delegates to GetByID.
func (r *UserRepositoryPG) GetByIDForAuth(ctx context.Context, id int64) (*entities.User, error) {
	return r.GetByID(ctx, id)
}

// scanUserByQuery executes a single-row SELECT and decodes the MFA secret
// into the typed VO; centralized so all read paths share the same parsing.
func (r *UserRepositoryPG) scanUserByQuery(ctx context.Context, query string, arg any) (*entities.User, error) {
	user := &entities.User{}
	var mfaSecret sql.NullString
	var mfaEncrypted bool
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Role,
		&user.Status,
		&mfaSecret,
		&user.MFAEnabled,
		&mfaEncrypted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	secret, err := r.unwrapMFASecret(mfaSecret, mfaEncrypted, user.ID)
	if err != nil {
		return nil, err
	}
	user.MFASecret = secret
	return user, nil
}

// CountByRole returns the number of users carrying the given role.
// Used by the users module to enforce the last-system_admin guard
// before a destructive admin delete (#283 ADR-4 Tier 1).
func (r *UserRepositoryPG) CountByRole(ctx context.Context, role domain.RoleType) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = $1 AND deleted_at IS NULL`, string(role)).Scan(&count)
	if err != nil {
		return 0, database.MapPostgresError(err)
	}
	return count, nil
}

// Delete removes a user by ID
func (r *UserRepositoryPG) Delete(ctx context.Context, userID int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return database.MapPostgresError(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return database.MapPostgresError(err)
	}

	if rows == 0 {
		return database.MapPostgresError(sql.ErrNoRows)
	}

	return nil
}

// List retrieves a paginated list of users
func (r *UserRepositoryPG) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	query := `
		SELECT id, email, password, name, role, status,
		       mfa_secret, mfa_enabled, mfa_secret_encrypted, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail the operation
			_ = err
		}
	}()

	users := []*entities.User{}
	for rows.Next() {
		user := &entities.User{}
		var mfaSecret sql.NullString
		var mfaEncrypted bool
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.Name,
			&user.Role,
			&user.Status,
			&mfaSecret,
			&user.MFAEnabled,
			&mfaEncrypted,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, database.MapPostgresError(err)
		}
		secret, err := r.unwrapMFASecret(mfaSecret, mfaEncrypted, user.ID)
		if err != nil {
			return nil, err
		}
		user.MFASecret = secret
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, database.MapPostgresError(err)
	}

	return users, nil
}
