package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

// ClassroomRepositoryPG implements ClassroomRepository using PostgreSQL.
type ClassroomRepositoryPG struct {
	db *sql.DB
}

// NewClassroomRepositoryPG creates a new ClassroomRepositoryPG.
func NewClassroomRepositoryPG(db *sql.DB) *ClassroomRepositoryPG {
	return &ClassroomRepositoryPG{db: db}
}

// GetByID retrieves a classroom by ID.
func (r *ClassroomRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Classroom, error) {
	query := `
		SELECT id, building, number, name, capacity, type, equipment,
			is_available, created_at, updated_at
		FROM classrooms WHERE id = $1`

	c := &entities.Classroom{}
	var equipmentJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Building, &c.Number, &c.Name, &c.Capacity, &c.Type,
		&equipmentJSON, &c.IsAvailable, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get classroom: %w", err)
	}

	if len(equipmentJSON) > 0 {
		if err := json.Unmarshal(equipmentJSON, &c.Equipment); err != nil {
			return nil, fmt.Errorf("failed to unmarshal equipment: %w", err)
		}
	}

	return c, nil
}

// List lists classrooms with filters.
func (r *ClassroomRepositoryPG) List(ctx context.Context, filter repositories.ClassroomFilter, limit, offset int) ([]*entities.Classroom, error) {
	whereClause, args := r.buildWhereClause(filter)

	query := `
		SELECT id, building, number, name, capacity, type, equipment,
			is_available, created_at, updated_at
		FROM classrooms` + whereClause + ` ORDER BY building, number`

	argNum := len(args) + 1
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list classrooms: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return r.scanClassrooms(rows)
}

// Count counts classrooms matching the filter.
func (r *ClassroomRepositoryPG) Count(ctx context.Context, filter repositories.ClassroomFilter) (int64, error) {
	whereClause, args := r.buildWhereClause(filter)

	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM classrooms"+whereClause, args...).Scan(&count)
	return count, err
}

func (r *ClassroomRepositoryPG) buildWhereClause(filter repositories.ClassroomFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.Building != nil {
		conditions = append(conditions, fmt.Sprintf("building = $%d", argNum))
		args = append(args, *filter.Building)
		argNum++
	}
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argNum))
		args = append(args, *filter.Type)
		argNum++
	}
	if filter.MinCapacity != nil {
		conditions = append(conditions, fmt.Sprintf("capacity >= $%d", argNum))
		args = append(args, *filter.MinCapacity)
		argNum++
	}
	if filter.IsAvailable != nil {
		conditions = append(conditions, fmt.Sprintf("is_available = $%d", argNum))
		args = append(args, *filter.IsAvailable)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

func (r *ClassroomRepositoryPG) scanClassrooms(rows *sql.Rows) ([]*entities.Classroom, error) {
	var classrooms []*entities.Classroom

	for rows.Next() {
		c := &entities.Classroom{}
		var equipmentJSON []byte

		err := rows.Scan(
			&c.ID, &c.Building, &c.Number, &c.Name, &c.Capacity, &c.Type,
			&equipmentJSON, &c.IsAvailable, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan classroom: %w", err)
		}

		if len(equipmentJSON) > 0 {
			if err := json.Unmarshal(equipmentJSON, &c.Equipment); err != nil {
				return nil, fmt.Errorf("failed to unmarshal equipment: %w", err)
			}
		}

		classrooms = append(classrooms, c)
	}

	return classrooms, rows.Err()
}
