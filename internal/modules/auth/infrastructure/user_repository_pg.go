package persistence

import (
	"database/sql"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
)

type userRepositoryPG struct {
	db *sql.DB
}

func NewUserRepositoryPG(db *sql.DB) repositories.UserRepository {
	return &userRepositoryPG{db: db}
}

func (r *userRepositoryPG) Create(user *entities.User) error {
	query := `
		INSERT INTO users (id, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`
	_, err := r.db.Exec(query, user.ID, user.Email, user.Password, user.Role)
	return err
}

func (r *userRepositoryPG) Delete(userID string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = $1`, userID)
	return err
}

func (r *userRepositoryPG) GetByEmail(email string) (*entities.User, error) {
	user := &entities.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return user, err
}

func (r *userRepositoryPG) GetByID(userID string) (*entities.User, error) {
	user := &entities.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password, role, created_at, updated_at FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return user, err
}

func (r *userRepositoryPG) List(page int, limit int) ([]*entities.User, error) {
	offset := (page - 1) * limit
	rows, err := r.db.Query(
		`SELECT id, email, password, role, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*entities.User{}
	for rows.Next() {
		user := &entities.User{}
		if err := rows.Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// internal/modules/auth/infrastructure/persistence/user_repository_pg.go
func (r *userRepositoryPG) Save(user *entities.User) error {
	query := `
		UPDATE users
		SET email = $1, password = $2, role = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.Exec(query, user.Email, user.Password, user.Role, user.ID)
	return err
}
