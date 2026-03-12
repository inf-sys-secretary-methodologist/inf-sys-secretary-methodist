package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/repositories"
)

// SyncLogRepositoryPg implements SyncLogRepository using PostgreSQL
type SyncLogRepositoryPg struct {
	db *sql.DB
}

// NewSyncLogRepositoryPg creates a new PostgreSQL sync log repository
func NewSyncLogRepositoryPg(db *sql.DB) repositories.SyncLogRepository {
	return &SyncLogRepositoryPg{db: db}
}

// Create creates a new sync log entry
func (r *SyncLogRepositoryPg) Create(ctx context.Context, log *entities.SyncLog) error {
	metadata, err := json.Marshal(log.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO sync_logs (
			entity_type, direction, status, started_at, completed_at,
			total_records, processed_count, success_count, error_count,
			conflict_count, error_message, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id`

	err = r.db.QueryRowContext(ctx, query,
		log.EntityType, log.Direction, log.Status, log.StartedAt, log.CompletedAt,
		log.TotalRecords, log.ProcessedCount, log.SuccessCount, log.ErrorCount,
		log.ConflictCount, log.ErrorMessage, metadata, log.CreatedAt, log.UpdatedAt,
	).Scan(&log.ID)

	if err != nil {
		return fmt.Errorf("failed to create sync log: %w", err)
	}

	return nil
}

// Update updates an existing sync log entry
func (r *SyncLogRepositoryPg) Update(ctx context.Context, log *entities.SyncLog) error {
	metadata, err := json.Marshal(log.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE sync_logs SET
			status = $1, completed_at = $2, total_records = $3,
			processed_count = $4, success_count = $5, error_count = $6,
			conflict_count = $7, error_message = $8, metadata = $9, updated_at = $10
		WHERE id = $11`

	result, err := r.db.ExecContext(ctx, query,
		log.Status, log.CompletedAt, log.TotalRecords,
		log.ProcessedCount, log.SuccessCount, log.ErrorCount,
		log.ConflictCount, log.ErrorMessage, metadata, time.Now(), log.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update sync log: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func scanSyncLogRow(row *sql.Row) (*entities.SyncLog, error) {
	var log entities.SyncLog
	var completedAt sql.NullTime
	var errorMessage sql.NullString
	var metadata []byte
	var entityType, direction, status string

	err := row.Scan(
		&log.ID, &entityType, &direction, &status,
		&log.StartedAt, &completedAt, &log.TotalRecords, &log.ProcessedCount,
		&log.SuccessCount, &log.ErrorCount, &log.ConflictCount,
		&errorMessage, &metadata, &log.CreatedAt, &log.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	log.EntityType = entities.SyncEntityType(entityType)
	log.Direction = entities.SyncDirection(direction)
	log.Status = entities.SyncStatus(status)

	if completedAt.Valid {
		log.CompletedAt = &completedAt.Time
	}
	if errorMessage.Valid {
		log.ErrorMessage = errorMessage.String
	}

	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &log.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	} else {
		log.Metadata = make(map[string]any)
	}

	return &log, nil
}

func scanSyncLogRows(rows *sql.Rows) ([]*entities.SyncLog, error) {
	var logs []*entities.SyncLog

	for rows.Next() {
		var log entities.SyncLog
		var completedAt sql.NullTime
		var errorMessage sql.NullString
		var metadata []byte
		var entityType, direction, status string

		err := rows.Scan(
			&log.ID, &entityType, &direction, &status,
			&log.StartedAt, &completedAt, &log.TotalRecords, &log.ProcessedCount,
			&log.SuccessCount, &log.ErrorCount, &log.ConflictCount,
			&errorMessage, &metadata, &log.CreatedAt, &log.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		log.EntityType = entities.SyncEntityType(entityType)
		log.Direction = entities.SyncDirection(direction)
		log.Status = entities.SyncStatus(status)

		if completedAt.Valid {
			log.CompletedAt = &completedAt.Time
		}
		if errorMessage.Valid {
			log.ErrorMessage = errorMessage.String
		}

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &log.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		} else {
			log.Metadata = make(map[string]any)
		}

		logs = append(logs, &log)
	}

	return logs, rows.Err()
}

// GetByID retrieves a sync log by ID
func (r *SyncLogRepositoryPg) GetByID(ctx context.Context, id int64) (*entities.SyncLog, error) {
	query := `
		SELECT id, entity_type, direction, status, started_at, completed_at,
		       total_records, processed_count, success_count, error_count,
		       conflict_count, error_message, metadata, created_at, updated_at
		FROM sync_logs WHERE id = $1`

	log, err := scanSyncLogRow(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sync log: %w", err)
	}

	return log, nil
}

// List retrieves sync logs with optional filtering
func (r *SyncLogRepositoryPg) List(ctx context.Context, filter entities.SyncLogFilter) ([]*entities.SyncLog, int64, error) {
	var conditions []string
	var args []any
	argNum := 1

	if filter.EntityType != nil {
		conditions = append(conditions, fmt.Sprintf("entity_type = $%d", argNum))
		args = append(args, *filter.EntityType)
		argNum++
	}
	if filter.Direction != nil {
		conditions = append(conditions, fmt.Sprintf("direction = $%d", argNum))
		args = append(args, *filter.Direction)
		argNum++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *filter.Status)
		argNum++
	}
	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("started_at >= $%d", argNum))
		args = append(args, *filter.StartDate)
		argNum++
	}
	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("started_at <= $%d", argNum))
		args = append(args, *filter.EndDate)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sync_logs %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count sync logs: %w", err)
	}

	// Get items
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(`
		SELECT id, entity_type, direction, status, started_at, completed_at,
		       total_records, processed_count, success_count, error_count,
		       conflict_count, error_message, metadata, created_at, updated_at
		FROM sync_logs %s
		ORDER BY started_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argNum, argNum+1)

	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list sync logs: %w", err)
	}
	defer rows.Close()

	logs, err := scanSyncLogRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetLatest retrieves the most recent sync log for an entity type
func (r *SyncLogRepositoryPg) GetLatest(ctx context.Context, entityType entities.SyncEntityType) (*entities.SyncLog, error) {
	query := `
		SELECT id, entity_type, direction, status, started_at, completed_at,
		       total_records, processed_count, success_count, error_count,
		       conflict_count, error_message, metadata, created_at, updated_at
		FROM sync_logs
		WHERE entity_type = $1
		ORDER BY started_at DESC
		LIMIT 1`

	log, err := scanSyncLogRow(r.db.QueryRowContext(ctx, query, entityType))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest sync log: %w", err)
	}

	return log, nil
}

// GetRunning retrieves all currently running sync operations
func (r *SyncLogRepositoryPg) GetRunning(ctx context.Context) ([]*entities.SyncLog, error) {
	query := `
		SELECT id, entity_type, direction, status, started_at, completed_at,
		       total_records, processed_count, success_count, error_count,
		       conflict_count, error_message, metadata, created_at, updated_at
		FROM sync_logs WHERE status = 'in_progress' ORDER BY started_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get running syncs: %w", err)
	}
	defer rows.Close()

	return scanSyncLogRows(rows)
}

// GetStats retrieves sync statistics
func (r *SyncLogRepositoryPg) GetStats(ctx context.Context, entityType *entities.SyncEntityType) (*entities.SyncStats, error) {
	var whereClause string
	var args []any

	if entityType != nil {
		whereClause = "WHERE entity_type = $1"
		args = append(args, *entityType)
	}

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total_syncs,
			COUNT(*) FILTER (WHERE status = 'completed') as successful_syncs,
			COUNT(*) FILTER (WHERE status = 'failed') as failed_syncs,
			COALESCE(SUM(total_records), 0) as total_records,
			COALESCE(SUM(conflict_count), 0) as total_conflicts,
			COALESCE(MAX(started_at), NOW()) as last_sync_at
		FROM sync_logs %s`, whereClause) // #nosec G201 -- dynamic WHERE clause from code logic, not user input

	var stats entities.SyncStats
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&stats.TotalSyncs,
		&stats.SuccessfulSyncs,
		&stats.FailedSyncs,
		&stats.TotalRecords,
		&stats.TotalConflicts,
		&stats.LastSyncAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get sync stats: %w", err)
	}

	return &stats, nil
}

// Delete deletes a sync log entry
func (r *SyncLogRepositoryPg) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM sync_logs WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete sync log: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteOlderThan deletes sync logs older than the specified number of days
func (r *SyncLogRepositoryPg) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days)

	result, err := r.db.ExecContext(ctx,
		"DELETE FROM sync_logs WHERE created_at < $1", cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old sync logs: %w", err)
	}

	return result.RowsAffected()
}
