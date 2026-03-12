package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/repositories"
)

// SyncConflictRepositoryPg implements SyncConflictRepository using PostgreSQL
type SyncConflictRepositoryPg struct {
	db *sql.DB
}

// NewSyncConflictRepositoryPg creates a new PostgreSQL sync conflict repository
func NewSyncConflictRepositoryPg(db *sql.DB) repositories.SyncConflictRepository {
	return &SyncConflictRepositoryPg{db: db}
}

func scanConflictRow(row *sql.Row) (*entities.SyncConflict, error) {
	var conflict entities.SyncConflict
	var entityType, resolution string
	var localData, externalData, resolvedData []byte
	var conflictFields pq.StringArray
	var resolvedBy sql.NullInt64
	var resolvedAt sql.NullTime
	var notes sql.NullString

	err := row.Scan(
		&conflict.ID, &conflict.SyncLogID, &entityType, &conflict.EntityID,
		&localData, &externalData, &conflict.ConflictType, &conflictFields,
		&resolution, &resolvedBy, &resolvedAt, &resolvedData, &notes,
		&conflict.CreatedAt, &conflict.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	conflict.EntityType = entities.SyncEntityType(entityType)
	conflict.Resolution = entities.ConflictResolution(resolution)
	conflict.ConflictFields = []string(conflictFields)

	if len(localData) > 0 {
		conflict.LocalData = string(localData)
	}
	if len(externalData) > 0 {
		conflict.ExternalData = string(externalData)
	}
	if resolvedBy.Valid {
		conflict.ResolvedBy = &resolvedBy.Int64
	}
	if resolvedAt.Valid {
		conflict.ResolvedAt = &resolvedAt.Time
	}
	if len(resolvedData) > 0 {
		conflict.ResolvedData = string(resolvedData)
	}
	if notes.Valid {
		conflict.Notes = notes.String
	}

	return &conflict, nil
}

func scanConflictRows(rows *sql.Rows) ([]*entities.SyncConflict, error) {
	var conflicts []*entities.SyncConflict

	for rows.Next() {
		var conflict entities.SyncConflict
		var entityType, resolution string
		var localData, externalData, resolvedData []byte
		var conflictFields pq.StringArray
		var resolvedBy sql.NullInt64
		var resolvedAt sql.NullTime
		var notes sql.NullString

		err := rows.Scan(
			&conflict.ID, &conflict.SyncLogID, &entityType, &conflict.EntityID,
			&localData, &externalData, &conflict.ConflictType, &conflictFields,
			&resolution, &resolvedBy, &resolvedAt, &resolvedData, &notes,
			&conflict.CreatedAt, &conflict.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		conflict.EntityType = entities.SyncEntityType(entityType)
		conflict.Resolution = entities.ConflictResolution(resolution)
		conflict.ConflictFields = []string(conflictFields)

		if len(localData) > 0 {
			conflict.LocalData = string(localData)
		}
		if len(externalData) > 0 {
			conflict.ExternalData = string(externalData)
		}
		if resolvedBy.Valid {
			conflict.ResolvedBy = &resolvedBy.Int64
		}
		if resolvedAt.Valid {
			conflict.ResolvedAt = &resolvedAt.Time
		}
		if len(resolvedData) > 0 {
			conflict.ResolvedData = string(resolvedData)
		}
		if notes.Valid {
			conflict.Notes = notes.String
		}

		conflicts = append(conflicts, &conflict)
	}

	return conflicts, rows.Err()
}

const conflictSelectFields = `id, sync_log_id, entity_type, entity_id, local_data, external_data,
	conflict_type, conflict_fields, resolution, resolved_by, resolved_at, resolved_data,
	notes, created_at, updated_at`

// Create creates a new sync conflict record
func (r *SyncConflictRepositoryPg) Create(ctx context.Context, conflict *entities.SyncConflict) error {
	query := `
		INSERT INTO sync_conflicts (
			sync_log_id, entity_type, entity_id, local_data, external_data,
			conflict_type, conflict_fields, resolution, resolved_by, resolved_at,
			resolved_data, notes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		conflict.SyncLogID, conflict.EntityType, conflict.EntityID,
		nullBytes(conflict.LocalData), nullBytes(conflict.ExternalData),
		conflict.ConflictType, pq.Array(conflict.ConflictFields),
		conflict.Resolution, nullInt64(conflict.ResolvedBy),
		nullTime(conflict.ResolvedAt), nullBytes(conflict.ResolvedData),
		nullString(conflict.Notes), conflict.CreatedAt, conflict.UpdatedAt,
	).Scan(&conflict.ID)

	if err != nil {
		return fmt.Errorf("failed to create sync conflict: %w", err)
	}

	return nil
}

// Update updates an existing sync conflict record
func (r *SyncConflictRepositoryPg) Update(ctx context.Context, conflict *entities.SyncConflict) error {
	query := `
		UPDATE sync_conflicts SET
			local_data = $1, external_data = $2, conflict_type = $3,
			conflict_fields = $4, resolution = $5, resolved_by = $6,
			resolved_at = $7, resolved_data = $8, notes = $9, updated_at = $10
		WHERE id = $11`

	result, err := r.db.ExecContext(ctx, query,
		nullBytes(conflict.LocalData), nullBytes(conflict.ExternalData),
		conflict.ConflictType, pq.Array(conflict.ConflictFields),
		conflict.Resolution, nullInt64(conflict.ResolvedBy),
		nullTime(conflict.ResolvedAt), nullBytes(conflict.ResolvedData),
		nullString(conflict.Notes), time.Now(), conflict.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update sync conflict: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetByID retrieves a sync conflict by ID
func (r *SyncConflictRepositoryPg) GetByID(ctx context.Context, id int64) (*entities.SyncConflict, error) {
	query := fmt.Sprintf(`SELECT %s FROM sync_conflicts WHERE id = $1`, conflictSelectFields)
	conflict, err := scanConflictRow(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sync conflict: %w", err)
	}

	return conflict, nil
}

// List retrieves sync conflicts with optional filtering
func (r *SyncConflictRepositoryPg) List(ctx context.Context, filter entities.SyncConflictFilter) ([]*entities.SyncConflict, int64, error) {
	var conditions []string
	var args []any
	argNum := 1

	if filter.SyncLogID != nil {
		conditions = append(conditions, fmt.Sprintf("sync_log_id = $%d", argNum))
		args = append(args, *filter.SyncLogID)
		argNum++
	}
	if filter.EntityType != nil {
		conditions = append(conditions, fmt.Sprintf("entity_type = $%d", argNum))
		args = append(args, *filter.EntityType)
		argNum++
	}
	if filter.Resolution != nil {
		conditions = append(conditions, fmt.Sprintf("resolution = $%d", argNum))
		args = append(args, *filter.Resolution)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sync_conflicts %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count sync conflicts: %w", err)
	}

	// Get items
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(`
		SELECT %s FROM sync_conflicts %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		conflictSelectFields, whereClause, argNum, argNum+1)

	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list sync conflicts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	conflicts, err := scanConflictRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return conflicts, total, nil
}

// GetBySyncLogID retrieves all conflicts for a specific sync log
func (r *SyncConflictRepositoryPg) GetBySyncLogID(ctx context.Context, syncLogID int64) ([]*entities.SyncConflict, error) {
	query := fmt.Sprintf(`SELECT %s FROM sync_conflicts WHERE sync_log_id = $1 ORDER BY created_at`, conflictSelectFields)
	rows, err := r.db.QueryContext(ctx, query, syncLogID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conflicts by sync log ID: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanConflictRows(rows)
}

// GetPending retrieves all pending (unresolved) conflicts
func (r *SyncConflictRepositoryPg) GetPending(ctx context.Context, limit, offset int) ([]*entities.SyncConflict, int64, error) {
	pending := entities.ConflictResolutionPending
	return r.List(ctx, entities.SyncConflictFilter{
		Resolution: &pending,
		Limit:      limit,
		Offset:     offset,
	})
}

// GetPendingByEntityType retrieves pending conflicts for a specific entity type
func (r *SyncConflictRepositoryPg) GetPendingByEntityType(ctx context.Context, entityType entities.SyncEntityType) ([]*entities.SyncConflict, error) {
	query := fmt.Sprintf(`SELECT %s FROM sync_conflicts WHERE entity_type = $1 AND resolution = 'pending' ORDER BY created_at`, conflictSelectFields)
	rows, err := r.db.QueryContext(ctx, query, entityType)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending conflicts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanConflictRows(rows)
}

// Resolve resolves a conflict with the specified resolution
func (r *SyncConflictRepositoryPg) Resolve(ctx context.Context, id int64, resolution entities.ConflictResolution, userID int64, resolvedData string) error {
	query := `
		UPDATE sync_conflicts SET
			resolution = $1, resolved_by = $2, resolved_at = $3,
			resolved_data = $4, updated_at = $5
		WHERE id = $6`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		resolution, userID, now, nullBytes(resolvedData), now, id,
	)

	if err != nil {
		return fmt.Errorf("failed to resolve conflict: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// BulkResolve resolves multiple conflicts with the same resolution
func (r *SyncConflictRepositoryPg) BulkResolve(ctx context.Context, ids []int64, resolution entities.ConflictResolution, userID int64) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
		UPDATE sync_conflicts SET
			resolution = $1, resolved_by = $2, resolved_at = $3, updated_at = $4
		WHERE id = ANY($5)`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		resolution, userID, now, now, pq.Array(ids),
	)

	if err != nil {
		return fmt.Errorf("failed to bulk resolve conflicts: %w", err)
	}

	return nil
}

// Delete deletes a sync conflict record
func (r *SyncConflictRepositoryPg) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM sync_conflicts WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete sync conflict: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteBySyncLogID deletes all conflicts for a specific sync log
func (r *SyncConflictRepositoryPg) DeleteBySyncLogID(ctx context.Context, syncLogID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sync_conflicts WHERE sync_log_id = $1", syncLogID)
	if err != nil {
		return fmt.Errorf("failed to delete conflicts by sync log ID: %w", err)
	}
	return nil
}

// GetStats retrieves conflict statistics
func (r *SyncConflictRepositoryPg) GetStats(ctx context.Context) (*entities.ConflictStats, error) {
	stats := &entities.ConflictStats{
		ByEntityType: make(map[entities.SyncEntityType]int64),
	}

	// Get overall stats
	query := `
		SELECT
			COUNT(*) as total_conflicts,
			COUNT(*) FILTER (WHERE resolution = 'pending') as pending_conflicts,
			COUNT(*) FILTER (WHERE resolution != 'pending') as resolved_conflicts
		FROM sync_conflicts`

	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalConflicts,
		&stats.PendingConflicts,
		&stats.ResolvedConflicts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get conflict stats: %w", err)
	}

	// Get by entity type
	typeQuery := `
		SELECT entity_type, COUNT(*) as count
		FROM sync_conflicts
		WHERE resolution = 'pending'
		GROUP BY entity_type`

	rows, err := r.db.QueryContext(ctx, typeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get conflicts by entity type: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var entityType string
		var count int64
		if err := rows.Scan(&entityType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan entity type stats: %w", err)
		}
		stats.ByEntityType[entities.SyncEntityType(entityType)] = count
	}

	return stats, nil
}

// Helper function
func nullBytes(s string) []byte {
	if s == "" {
		return nil
	}

	// Try to parse as JSON, if valid return as-is, otherwise wrap in quotes
	var js json.RawMessage
	if json.Unmarshal([]byte(s), &js) == nil {
		return []byte(s)
	}

	// If not valid JSON, encode as JSON string
	b, _ := json.Marshal(s)
	return b
}
