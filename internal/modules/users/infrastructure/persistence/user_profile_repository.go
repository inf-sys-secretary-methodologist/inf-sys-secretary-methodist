// Package persistence implements repository interfaces for the users module.
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/database"
)

// UserProfileRepositoryPG implements PostgreSQL user profile repository.
type UserProfileRepositoryPG struct {
	db *sql.DB
}

// NewUserProfileRepositoryPG creates a new PostgreSQL user profile repository.
func NewUserProfileRepositoryPG(db *sql.DB) repositories.UserProfileRepository {
	return &UserProfileRepositoryPG{db: db}
}

// GetProfileByID retrieves user profile with organizational info.
func (r *UserProfileRepositoryPG) GetProfileByID(ctx context.Context, userID int64) (*entities.UserWithOrg, error) {
	user := &entities.UserWithOrg{}
	query := `
		SELECT
			u.id, u.email, u.name, u.role, u.status,
			COALESCE(up.phone, '') as phone,
			COALESCE(up.avatar, '') as avatar,
			up.department_id,
			COALESCE(d.name, '') as department_name,
			up.position_id,
			COALESCE(p.name, '') as position_name,
			u.created_at, u.updated_at
		FROM users u
		LEFT JOIN user_profiles up ON u.id = up.user_id
		LEFT JOIN org_departments d ON up.department_id = d.id
		LEFT JOIN positions p ON up.position_id = p.id
		WHERE u.id = $1
	`

	var departmentID, positionID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.Status,
		&user.Phone,
		&user.Avatar,
		&departmentID,
		&user.DepartmentName,
		&positionID,
		&user.PositionName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}

	if departmentID.Valid {
		user.DepartmentID = &departmentID.Int64
	}
	if positionID.Valid {
		user.PositionID = &positionID.Int64
	}

	return user, nil
}

// UpdateProfile updates user's organizational information.
func (r *UserProfileRepositoryPG) UpdateProfile(ctx context.Context, userID int64, departmentID, positionID *int64, phone, avatar, bio string) error {
	now := time.Now()

	// Use upsert to create or update profile
	query := `
		INSERT INTO user_profiles (user_id, department_id, position_id, phone, avatar, bio, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			department_id = EXCLUDED.department_id,
			position_id = EXCLUDED.position_id,
			phone = EXCLUDED.phone,
			avatar = EXCLUDED.avatar,
			bio = EXCLUDED.bio,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		userID,
		departmentID,
		positionID,
		phone,
		avatar,
		bio,
		now,
	)

	if err != nil {
		return database.MapPostgresError(err)
	}
	return nil
}

// ListUsersWithOrg retrieves paginated list of users with their organizational info.
func (r *UserProfileRepositoryPG) ListUsersWithOrg(ctx context.Context, filter *repositories.UserFilter, limit, offset int) ([]*entities.UserWithOrg, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT
			u.id, u.email, u.name, u.role, u.status,
			COALESCE(up.phone, '') as phone,
			COALESCE(up.avatar, '') as avatar,
			up.department_id,
			COALESCE(d.name, '') as department_name,
			up.position_id,
			COALESCE(p.name, '') as position_name,
			u.created_at, u.updated_at
		FROM users u
		LEFT JOIN user_profiles up ON u.id = up.user_id
		LEFT JOIN org_departments d ON up.department_id = d.id
		LEFT JOIN positions p ON up.position_id = p.id
	`

	conditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter != nil {
		if filter.DepartmentID != nil {
			conditions = append(conditions, fmt.Sprintf("up.department_id = $%d", argIndex))
			args = append(args, *filter.DepartmentID)
			argIndex++
		}
		if filter.PositionID != nil {
			conditions = append(conditions, fmt.Sprintf("up.position_id = $%d", argIndex))
			args = append(args, *filter.PositionID)
			argIndex++
		}
		if filter.Role != "" {
			conditions = append(conditions, fmt.Sprintf("u.role = $%d", argIndex))
			args = append(args, filter.Role)
			argIndex++
		}
		if filter.Status != "" {
			conditions = append(conditions, fmt.Sprintf("u.status = $%d", argIndex))
			args = append(args, filter.Status)
			argIndex++
		}
		if filter.Search != "" {
			conditions = append(conditions, fmt.Sprintf("(u.name ILIKE $%d OR u.email ILIKE $%d)", argIndex, argIndex))
			args = append(args, "%"+filter.Search+"%")
			argIndex++
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	defer rows.Close()

	users := []*entities.UserWithOrg{}
	for rows.Next() {
		user := &entities.UserWithOrg{}
		var departmentID, positionID sql.NullInt64

		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Role,
			&user.Status,
			&user.Phone,
			&user.Avatar,
			&departmentID,
			&user.DepartmentName,
			&positionID,
			&user.PositionName,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, database.MapPostgresError(err)
		}

		if departmentID.Valid {
			user.DepartmentID = &departmentID.Int64
		}
		if positionID.Valid {
			user.PositionID = &positionID.Int64
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, database.MapPostgresError(err)
	}

	return users, nil
}

// CountUsers returns total count of users matching the filter.
func (r *UserProfileRepositoryPG) CountUsers(ctx context.Context, filter *repositories.UserFilter) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM users u
		LEFT JOIN user_profiles up ON u.id = up.user_id
	`

	conditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter != nil {
		if filter.DepartmentID != nil {
			conditions = append(conditions, fmt.Sprintf("up.department_id = $%d", argIndex))
			args = append(args, *filter.DepartmentID)
			argIndex++
		}
		if filter.PositionID != nil {
			conditions = append(conditions, fmt.Sprintf("up.position_id = $%d", argIndex))
			args = append(args, *filter.PositionID)
			argIndex++
		}
		if filter.Role != "" {
			conditions = append(conditions, fmt.Sprintf("u.role = $%d", argIndex))
			args = append(args, filter.Role)
			argIndex++
		}
		if filter.Status != "" {
			conditions = append(conditions, fmt.Sprintf("u.status = $%d", argIndex))
			args = append(args, filter.Status)
			argIndex++
		}
		if filter.Search != "" {
			conditions = append(conditions, fmt.Sprintf("(u.name ILIKE $%d OR u.email ILIKE $%d)", argIndex, argIndex))
			args = append(args, "%"+filter.Search+"%")
			argIndex++
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, database.MapPostgresError(err)
	}
	return count, nil
}

// GetUsersByDepartment retrieves all users in a specific department.
func (r *UserProfileRepositoryPG) GetUsersByDepartment(ctx context.Context, departmentID int64) ([]*entities.UserWithOrg, error) {
	filter := &repositories.UserFilter{DepartmentID: &departmentID}
	return r.ListUsersWithOrg(ctx, filter, 1000, 0)
}

// GetUsersByPosition retrieves all users with a specific position.
func (r *UserProfileRepositoryPG) GetUsersByPosition(ctx context.Context, positionID int64) ([]*entities.UserWithOrg, error) {
	filter := &repositories.UserFilter{PositionID: &positionID}
	return r.ListUsersWithOrg(ctx, filter, 1000, 0)
}

// BulkUpdateDepartment moves multiple users to a new department.
func (r *UserProfileRepositoryPG) BulkUpdateDepartment(ctx context.Context, userIDs []int64, departmentID *int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	now := time.Now()

	// Build placeholders for user IDs
	placeholders := make([]string, len(userIDs))
	args := []interface{}{departmentID, now}
	for i, id := range userIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+3)
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		UPDATE user_profiles
		SET department_id = $1, updated_at = $2
		WHERE user_id IN (%s)
	`, strings.Join(placeholders, ", "))

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return database.MapPostgresError(err)
	}
	return nil
}

// BulkUpdatePosition assigns multiple users to a new position.
func (r *UserProfileRepositoryPG) BulkUpdatePosition(ctx context.Context, userIDs []int64, positionID *int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	now := time.Now()

	// Build placeholders for user IDs
	placeholders := make([]string, len(userIDs))
	args := []interface{}{positionID, now}
	for i, id := range userIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+3)
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		UPDATE user_profiles
		SET position_id = $1, updated_at = $2
		WHERE user_id IN (%s)
	`, strings.Join(placeholders, ", "))

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return database.MapPostgresError(err)
	}
	return nil
}
