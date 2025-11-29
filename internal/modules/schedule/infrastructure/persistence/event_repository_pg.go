// Package persistence provides database implementations for schedule repositories.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
	"github.com/lib/pq"
)

// EventRepositoryPG implements EventRepository using PostgreSQL
type EventRepositoryPG struct {
	db *sql.DB
}

// NewEventRepositoryPG creates a new PostgreSQL event repository
func NewEventRepositoryPG(db *sql.DB) *EventRepositoryPG {
	return &EventRepositoryPG{db: db}
}

// Create inserts a new event
func (r *EventRepositoryPG) Create(ctx context.Context, event *entities.Event) error {
	var recurrenceJSON, metadataJSON []byte
	var err error

	if event.RecurrenceRule != nil {
		recurrenceJSON, err = json.Marshal(event.RecurrenceRule)
		if err != nil {
			return fmt.Errorf("failed to marshal recurrence rule: %w", err)
		}
	}

	if event.Metadata != nil {
		metadataJSON, err = json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO events (
			title, description, event_type, status,
			start_time, end_time, all_day, timezone, location,
			organizer_id, is_recurring, recurrence_rule,
			parent_event_id, recurrence_id, color, priority,
			metadata, external_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		) RETURNING id`

	err = r.db.QueryRowContext(ctx, query,
		event.Title, event.Description, event.EventType, event.Status,
		event.StartTime, event.EndTime, event.AllDay, event.Timezone, event.Location,
		event.OrganizerID, event.IsRecurring, recurrenceJSON,
		event.ParentEventID, event.RecurrenceID, event.Color, event.Priority,
		metadataJSON, event.ExternalID, event.CreatedAt, event.UpdatedAt,
	).Scan(&event.ID)

	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}
	return nil
}

// Update updates an existing event
func (r *EventRepositoryPG) Update(ctx context.Context, event *entities.Event) error {
	event.UpdatedAt = time.Now()

	var recurrenceJSON, metadataJSON []byte
	var err error

	if event.RecurrenceRule != nil {
		recurrenceJSON, err = json.Marshal(event.RecurrenceRule)
		if err != nil {
			return fmt.Errorf("failed to marshal recurrence rule: %w", err)
		}
	}

	if event.Metadata != nil {
		metadataJSON, err = json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		UPDATE events SET
			title = $1, description = $2, event_type = $3, status = $4,
			start_time = $5, end_time = $6, all_day = $7, timezone = $8, location = $9,
			is_recurring = $10, recurrence_rule = $11, color = $12, priority = $13,
			metadata = $14, external_id = $15, updated_at = $16
		WHERE id = $17 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query,
		event.Title, event.Description, event.EventType, event.Status,
		event.StartTime, event.EndTime, event.AllDay, event.Timezone, event.Location,
		event.IsRecurring, recurrenceJSON, event.Color, event.Priority,
		metadataJSON, event.ExternalID, event.UpdatedAt, event.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("event not found")
	}
	return nil
}

// Delete permanently deletes an event
func (r *EventRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

// SoftDelete marks an event as deleted
func (r *EventRepositoryPG) SoftDelete(ctx context.Context, id int64) error {
	now := time.Now()
	query := `UPDATE events SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete event: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("event not found")
	}
	return nil
}

// GetByID retrieves an event by ID
func (r *EventRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Event, error) {
	query := `
		SELECT id, title, description, event_type, status,
			start_time, end_time, all_day, timezone, location,
			organizer_id, is_recurring, recurrence_rule,
			parent_event_id, recurrence_id, color, priority,
			metadata, external_id, created_at, updated_at, deleted_at
		FROM events WHERE id = $1`

	event := &entities.Event{}
	var recurrenceJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID, &event.Title, &event.Description, &event.EventType, &event.Status,
		&event.StartTime, &event.EndTime, &event.AllDay, &event.Timezone, &event.Location,
		&event.OrganizerID, &event.IsRecurring, &recurrenceJSON,
		&event.ParentEventID, &event.RecurrenceID, &event.Color, &event.Priority,
		&metadataJSON, &event.ExternalID, &event.CreatedAt, &event.UpdatedAt, &event.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("event not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if len(recurrenceJSON) > 0 {
		event.RecurrenceRule = &entities.RecurrenceRule{}
		if err := json.Unmarshal(recurrenceJSON, event.RecurrenceRule); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recurrence rule: %w", err)
		}
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return event, nil
}

// List retrieves events with filtering and pagination
func (r *EventRepositoryPG) List(ctx context.Context, filter repositories.EventFilter) ([]*entities.Event, int64, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if !filter.IncludeDeleted {
		conditions = append(conditions, "deleted_at IS NULL")
	}

	if filter.OrganizerID != nil {
		conditions = append(conditions, fmt.Sprintf("organizer_id = $%d", argNum))
		args = append(args, *filter.OrganizerID)
		argNum++
	}

	if filter.EventType != nil {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", argNum))
		args = append(args, *filter.EventType)
		argNum++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *filter.Status)
		argNum++
	}

	if filter.StartFrom != nil {
		conditions = append(conditions, fmt.Sprintf("start_time >= $%d", argNum))
		args = append(args, *filter.StartFrom)
		argNum++
	}

	if filter.StartTo != nil {
		conditions = append(conditions, fmt.Sprintf("start_time <= $%d", argNum))
		args = append(args, *filter.StartTo)
		argNum++
	}

	if filter.IsRecurring != nil {
		conditions = append(conditions, fmt.Sprintf("is_recurring = $%d", argNum))
		args = append(args, *filter.IsRecurring)
		argNum++
	}

	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+*filter.SearchQuery+"%")
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM events %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Get events
	orderBy := "start_time ASC"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}

	limit := 20
	if filter.Limit > 0 {
		limit = filter.Limit
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, event_type, status,
			start_time, end_time, all_day, timezone, location,
			organizer_id, is_recurring, recurrence_rule,
			parent_event_id, recurrence_id, color, priority,
			metadata, external_id, created_at, updated_at, deleted_at
		FROM events %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argNum, argNum+1)

	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	events, err := r.scanEvents(rows)
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// GetByDateRange retrieves events in a date range
func (r *EventRepositoryPG) GetByDateRange(ctx context.Context, start, end time.Time, userID *int64) ([]*entities.Event, error) {
	query := `
		SELECT e.id, e.title, e.description, e.event_type, e.status,
			e.start_time, e.end_time, e.all_day, e.timezone, e.location,
			e.organizer_id, e.is_recurring, e.recurrence_rule,
			e.parent_event_id, e.recurrence_id, e.color, e.priority,
			e.metadata, e.external_id, e.created_at, e.updated_at, e.deleted_at
		FROM events e
		WHERE e.deleted_at IS NULL
			AND ((e.start_time >= $1 AND e.start_time <= $2)
				OR (e.end_time >= $1 AND e.end_time <= $2)
				OR (e.start_time <= $1 AND (e.end_time >= $2 OR e.end_time IS NULL)))`

	var args []interface{}
	args = append(args, start, end)

	if userID != nil {
		query += ` AND (e.organizer_id = $3 OR EXISTS (
			SELECT 1 FROM event_participants ep WHERE ep.event_id = e.id AND ep.user_id = $3
		))`
		args = append(args, *userID)
	}

	query += ` ORDER BY e.start_time ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by date range: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetByOrganizer retrieves events by organizer
func (r *EventRepositoryPG) GetByOrganizer(ctx context.Context, organizerID int64, limit, offset int) ([]*entities.Event, error) {
	query := `
		SELECT id, title, description, event_type, status,
			start_time, end_time, all_day, timezone, location,
			organizer_id, is_recurring, recurrence_rule,
			parent_event_id, recurrence_id, color, priority,
			metadata, external_id, created_at, updated_at, deleted_at
		FROM events
		WHERE organizer_id = $1 AND deleted_at IS NULL
		ORDER BY start_time DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, organizerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by organizer: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetByParticipant retrieves events where user is a participant
func (r *EventRepositoryPG) GetByParticipant(ctx context.Context, userID int64, limit, offset int) ([]*entities.Event, error) {
	query := `
		SELECT e.id, e.title, e.description, e.event_type, e.status,
			e.start_time, e.end_time, e.all_day, e.timezone, e.location,
			e.organizer_id, e.is_recurring, e.recurrence_rule,
			e.parent_event_id, e.recurrence_id, e.color, e.priority,
			e.metadata, e.external_id, e.created_at, e.updated_at, e.deleted_at
		FROM events e
		INNER JOIN event_participants ep ON e.id = ep.event_id
		WHERE ep.user_id = $1 AND e.deleted_at IS NULL
		ORDER BY e.start_time DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by participant: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetUpcoming retrieves upcoming events for a user
func (r *EventRepositoryPG) GetUpcoming(ctx context.Context, userID int64, limit int) ([]*entities.Event, error) {
	query := `
		SELECT DISTINCT e.id, e.title, e.description, e.event_type, e.status,
			e.start_time, e.end_time, e.all_day, e.timezone, e.location,
			e.organizer_id, e.is_recurring, e.recurrence_rule,
			e.parent_event_id, e.recurrence_id, e.color, e.priority,
			e.metadata, e.external_id, e.created_at, e.updated_at, e.deleted_at
		FROM events e
		LEFT JOIN event_participants ep ON e.id = ep.event_id
		WHERE e.deleted_at IS NULL
			AND e.start_time >= NOW()
			AND (e.organizer_id = $1 OR ep.user_id = $1)
			AND e.status IN ('scheduled', 'ongoing')
		ORDER BY e.start_time ASC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming events: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetRecurringEvents retrieves all recurring events
func (r *EventRepositoryPG) GetRecurringEvents(ctx context.Context) ([]*entities.Event, error) {
	query := `
		SELECT id, title, description, event_type, status,
			start_time, end_time, all_day, timezone, location,
			organizer_id, is_recurring, recurrence_rule,
			parent_event_id, recurrence_id, color, priority,
			metadata, external_id, created_at, updated_at, deleted_at
		FROM events
		WHERE is_recurring = true AND parent_event_id IS NULL AND deleted_at IS NULL
		ORDER BY start_time ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring events: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetRecurrenceInstances retrieves instances of a recurring event
func (r *EventRepositoryPG) GetRecurrenceInstances(ctx context.Context, parentEventID int64, start, end time.Time) ([]*entities.Event, error) {
	query := `
		SELECT id, title, description, event_type, status,
			start_time, end_time, all_day, timezone, location,
			organizer_id, is_recurring, recurrence_rule,
			parent_event_id, recurrence_id, color, priority,
			metadata, external_id, created_at, updated_at, deleted_at
		FROM events
		WHERE parent_event_id = $1
			AND start_time >= $2
			AND start_time <= $3
			AND deleted_at IS NULL
		ORDER BY start_time ASC`

	rows, err := r.db.QueryContext(ctx, query, parentEventID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurrence instances: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// CreateRecurrenceInstance creates an instance of a recurring event
func (r *EventRepositoryPG) CreateRecurrenceInstance(ctx context.Context, event *entities.Event) error {
	return r.Create(ctx, event)
}

// GetRecurrenceExceptions retrieves exception dates for a recurring event
func (r *EventRepositoryPG) GetRecurrenceExceptions(ctx context.Context, parentEventID int64) ([]time.Time, error) {
	query := `SELECT exception_date FROM event_recurrence_exceptions WHERE event_id = $1 ORDER BY exception_date`

	rows, err := r.db.QueryContext(ctx, query, parentEventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurrence exceptions: %w", err)
	}
	defer rows.Close()

	var exceptions []time.Time
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("failed to scan exception date: %w", err)
		}
		exceptions = append(exceptions, t)
	}

	return exceptions, nil
}

// AddRecurrenceException adds an exception date for a recurring event
func (r *EventRepositoryPG) AddRecurrenceException(ctx context.Context, parentEventID int64, exceptionDate time.Time) error {
	query := `INSERT INTO event_recurrence_exceptions (event_id, exception_date, created_at) VALUES ($1, $2, NOW())`
	_, err := r.db.ExecContext(ctx, query, parentEventID, exceptionDate)
	if err != nil {
		return fmt.Errorf("failed to add recurrence exception: %w", err)
	}
	return nil
}

// scanEvents scans multiple events from rows
func (r *EventRepositoryPG) scanEvents(rows *sql.Rows) ([]*entities.Event, error) {
	var events []*entities.Event

	for rows.Next() {
		event := &entities.Event{}
		var recurrenceJSON, metadataJSON []byte

		err := rows.Scan(
			&event.ID, &event.Title, &event.Description, &event.EventType, &event.Status,
			&event.StartTime, &event.EndTime, &event.AllDay, &event.Timezone, &event.Location,
			&event.OrganizerID, &event.IsRecurring, &recurrenceJSON,
			&event.ParentEventID, &event.RecurrenceID, &event.Color, &event.Priority,
			&metadataJSON, &event.ExternalID, &event.CreatedAt, &event.UpdatedAt, &event.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if len(recurrenceJSON) > 0 {
			event.RecurrenceRule = &entities.RecurrenceRule{}
			if err := json.Unmarshal(recurrenceJSON, event.RecurrenceRule); err != nil {
				return nil, fmt.Errorf("failed to unmarshal recurrence rule: %w", err)
			}
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		events = append(events, event)
	}

	return events, nil
}

// EventParticipantRepositoryPG implements EventParticipantRepository
type EventParticipantRepositoryPG struct {
	db *sql.DB
}

// NewEventParticipantRepositoryPG creates a new participant repository
func NewEventParticipantRepositoryPG(db *sql.DB) *EventParticipantRepositoryPG {
	return &EventParticipantRepositoryPG{db: db}
}

// Create inserts a new participant
func (r *EventParticipantRepositoryPG) Create(ctx context.Context, p *entities.EventParticipant) error {
	query := `
		INSERT INTO event_participants (event_id, user_id, response_status, role, created_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		p.EventID, p.UserID, p.ResponseStatus, p.Role, p.CreatedAt,
	).Scan(&p.ID)

	if err != nil {
		return fmt.Errorf("failed to create participant: %w", err)
	}
	return nil
}

// Update updates a participant
func (r *EventParticipantRepositoryPG) Update(ctx context.Context, p *entities.EventParticipant) error {
	query := `
		UPDATE event_participants SET
			response_status = $1, role = $2, notified_at = $3, responded_at = $4
		WHERE id = $5`

	_, err := r.db.ExecContext(ctx, query, p.ResponseStatus, p.Role, p.NotifiedAt, p.RespondedAt, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update participant: %w", err)
	}
	return nil
}

// Delete removes a participant
func (r *EventParticipantRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM event_participants WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete participant: %w", err)
	}
	return nil
}

// GetByID retrieves a participant by ID
func (r *EventParticipantRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.EventParticipant, error) {
	query := `
		SELECT id, event_id, user_id, response_status, role, notified_at, responded_at, created_at
		FROM event_participants WHERE id = $1`

	p := &entities.EventParticipant{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.EventID, &p.UserID, &p.ResponseStatus, &p.Role, &p.NotifiedAt, &p.RespondedAt, &p.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("participant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}
	return p, nil
}

// GetByEventID retrieves all participants for an event
func (r *EventParticipantRepositoryPG) GetByEventID(ctx context.Context, eventID int64) ([]*entities.EventParticipant, error) {
	query := `
		SELECT id, event_id, user_id, response_status, role, notified_at, responded_at, created_at
		FROM event_participants WHERE event_id = $1 ORDER BY created_at`

	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	defer rows.Close()

	return r.scanParticipants(rows)
}

// GetByUserID retrieves all participations for a user
func (r *EventParticipantRepositoryPG) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entities.EventParticipant, error) {
	query := `
		SELECT id, event_id, user_id, response_status, role, notified_at, responded_at, created_at
		FROM event_participants WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get participations: %w", err)
	}
	defer rows.Close()

	return r.scanParticipants(rows)
}

// GetByEventAndUser retrieves a specific participation
func (r *EventParticipantRepositoryPG) GetByEventAndUser(ctx context.Context, eventID, userID int64) (*entities.EventParticipant, error) {
	query := `
		SELECT id, event_id, user_id, response_status, role, notified_at, responded_at, created_at
		FROM event_participants WHERE event_id = $1 AND user_id = $2`

	p := &entities.EventParticipant{}
	err := r.db.QueryRowContext(ctx, query, eventID, userID).Scan(
		&p.ID, &p.EventID, &p.UserID, &p.ResponseStatus, &p.Role, &p.NotifiedAt, &p.RespondedAt, &p.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("participant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}
	return p, nil
}

// AddParticipants adds multiple participants to an event
func (r *EventParticipantRepositoryPG) AddParticipants(ctx context.Context, eventID int64, userIDs []int64, role entities.ParticipantRole) error {
	if len(userIDs) == 0 {
		return nil
	}

	query := `
		INSERT INTO event_participants (event_id, user_id, response_status, role, created_at)
		VALUES ($1, unnest($2::bigint[]), $3, $4, NOW())
		ON CONFLICT (event_id, user_id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query, eventID, pq.Array(userIDs), entities.ParticipantStatusPending, role)
	if err != nil {
		return fmt.Errorf("failed to add participants: %w", err)
	}
	return nil
}

// RemoveParticipants removes specific participants from an event
func (r *EventParticipantRepositoryPG) RemoveParticipants(ctx context.Context, eventID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	query := `DELETE FROM event_participants WHERE event_id = $1 AND user_id = ANY($2)`
	_, err := r.db.ExecContext(ctx, query, eventID, pq.Array(userIDs))
	if err != nil {
		return fmt.Errorf("failed to remove participants: %w", err)
	}
	return nil
}

// RemoveAllParticipants removes all participants from an event
func (r *EventParticipantRepositoryPG) RemoveAllParticipants(ctx context.Context, eventID int64) error {
	query := `DELETE FROM event_participants WHERE event_id = $1`
	_, err := r.db.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to remove all participants: %w", err)
	}
	return nil
}

// UpdateStatus updates participant response status
func (r *EventParticipantRepositoryPG) UpdateStatus(ctx context.Context, eventID, userID int64, status entities.ParticipantStatus) error {
	now := time.Now()
	query := `UPDATE event_participants SET response_status = $1, responded_at = $2 WHERE event_id = $3 AND user_id = $4`
	result, err := r.db.ExecContext(ctx, query, status, now, eventID, userID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("participant not found")
	}
	return nil
}

// GetPendingInvitations retrieves pending invitations for a user
func (r *EventParticipantRepositoryPG) GetPendingInvitations(ctx context.Context, userID int64) ([]*entities.EventParticipant, error) {
	query := `
		SELECT ep.id, ep.event_id, ep.user_id, ep.response_status, ep.role, ep.notified_at, ep.responded_at, ep.created_at
		FROM event_participants ep
		INNER JOIN events e ON ep.event_id = e.id
		WHERE ep.user_id = $1 AND ep.response_status = 'pending'
			AND e.deleted_at IS NULL AND e.start_time >= NOW()
		ORDER BY e.start_time ASC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending invitations: %w", err)
	}
	defer rows.Close()

	return r.scanParticipants(rows)
}

func (r *EventParticipantRepositoryPG) scanParticipants(rows *sql.Rows) ([]*entities.EventParticipant, error) {
	var participants []*entities.EventParticipant

	for rows.Next() {
		p := &entities.EventParticipant{}
		err := rows.Scan(
			&p.ID, &p.EventID, &p.UserID, &p.ResponseStatus, &p.Role, &p.NotifiedAt, &p.RespondedAt, &p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}
		participants = append(participants, p)
	}

	return participants, nil
}

// EventReminderRepositoryPG implements EventReminderRepository
type EventReminderRepositoryPG struct {
	db *sql.DB
}

// NewEventReminderRepositoryPG creates a new reminder repository
func NewEventReminderRepositoryPG(db *sql.DB) *EventReminderRepositoryPG {
	return &EventReminderRepositoryPG{db: db}
}

// Create inserts a new reminder
func (r *EventReminderRepositoryPG) Create(ctx context.Context, reminder *entities.EventReminder) error {
	query := `
		INSERT INTO event_reminders (event_id, user_id, reminder_type, minutes_before, is_sent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		reminder.EventID, reminder.UserID, reminder.ReminderType, reminder.MinutesBefore, reminder.IsSent, reminder.CreatedAt,
	).Scan(&reminder.ID)

	if err != nil {
		return fmt.Errorf("failed to create reminder: %w", err)
	}
	return nil
}

// Update updates a reminder
func (r *EventReminderRepositoryPG) Update(ctx context.Context, reminder *entities.EventReminder) error {
	query := `
		UPDATE event_reminders SET
			reminder_type = $1, minutes_before = $2, is_sent = $3, sent_at = $4
		WHERE id = $5`

	_, err := r.db.ExecContext(ctx, query,
		reminder.ReminderType, reminder.MinutesBefore, reminder.IsSent, reminder.SentAt, reminder.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update reminder: %w", err)
	}
	return nil
}

// Delete removes a reminder
func (r *EventReminderRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM event_reminders WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}
	return nil
}

// GetByID retrieves a reminder by ID
func (r *EventReminderRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.EventReminder, error) {
	query := `
		SELECT id, event_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at
		FROM event_reminders WHERE id = $1`

	reminder := &entities.EventReminder{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&reminder.ID, &reminder.EventID, &reminder.UserID, &reminder.ReminderType,
		&reminder.MinutesBefore, &reminder.IsSent, &reminder.SentAt, &reminder.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("reminder not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get reminder: %w", err)
	}
	return reminder, nil
}

// GetByEventID retrieves all reminders for an event
func (r *EventReminderRepositoryPG) GetByEventID(ctx context.Context, eventID int64) ([]*entities.EventReminder, error) {
	query := `
		SELECT id, event_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at
		FROM event_reminders WHERE event_id = $1 ORDER BY minutes_before DESC`

	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reminders: %w", err)
	}
	defer rows.Close()

	return r.scanReminders(rows)
}

// GetByUserID retrieves all reminders for a user
func (r *EventReminderRepositoryPG) GetByUserID(ctx context.Context, userID int64) ([]*entities.EventReminder, error) {
	query := `
		SELECT id, event_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at
		FROM event_reminders WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reminders: %w", err)
	}
	defer rows.Close()

	return r.scanReminders(rows)
}

// GetByEventAndUser retrieves reminders for a specific event and user
func (r *EventReminderRepositoryPG) GetByEventAndUser(ctx context.Context, eventID, userID int64) ([]*entities.EventReminder, error) {
	query := `
		SELECT id, event_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at
		FROM event_reminders WHERE event_id = $1 AND user_id = $2 ORDER BY minutes_before DESC`

	rows, err := r.db.QueryContext(ctx, query, eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reminders: %w", err)
	}
	defer rows.Close()

	return r.scanReminders(rows)
}

// GetPendingReminders retrieves unsent reminders that need to be sent
func (r *EventReminderRepositoryPG) GetPendingReminders(ctx context.Context, beforeTime time.Time) ([]*entities.EventReminder, error) {
	query := `
		SELECT r.id, r.event_id, r.user_id, r.reminder_type, r.minutes_before, r.is_sent, r.sent_at, r.created_at
		FROM event_reminders r
		INNER JOIN events e ON r.event_id = e.id
		WHERE r.is_sent = false
			AND e.deleted_at IS NULL
			AND e.status IN ('scheduled', 'ongoing')
			AND (e.start_time - (r.minutes_before || ' minutes')::interval) <= $1
		ORDER BY e.start_time ASC`

	rows, err := r.db.QueryContext(ctx, query, beforeTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending reminders: %w", err)
	}
	defer rows.Close()

	return r.scanReminders(rows)
}

// MarkAsSent marks a reminder as sent
func (r *EventReminderRepositoryPG) MarkAsSent(ctx context.Context, id int64) error {
	now := time.Now()
	query := `UPDATE event_reminders SET is_sent = true, sent_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark reminder as sent: %w", err)
	}
	return nil
}

// MarkMultipleAsSent marks multiple reminders as sent
func (r *EventReminderRepositoryPG) MarkMultipleAsSent(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	query := `UPDATE event_reminders SET is_sent = true, sent_at = $1 WHERE id = ANY($2)`
	_, err := r.db.ExecContext(ctx, query, now, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("failed to mark reminders as sent: %w", err)
	}
	return nil
}

// DeleteByEventID removes all reminders for an event
func (r *EventReminderRepositoryPG) DeleteByEventID(ctx context.Context, eventID int64) error {
	query := `DELETE FROM event_reminders WHERE event_id = $1`
	_, err := r.db.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete reminders: %w", err)
	}
	return nil
}

// CreateDefault creates default reminders for an event
func (r *EventReminderRepositoryPG) CreateDefault(ctx context.Context, eventID, userID int64) error {
	// Default reminders: 15 min, 1 hour, 1 day before
	defaults := []int{15, 60, 1440}

	query := `
		INSERT INTO event_reminders (event_id, user_id, reminder_type, minutes_before, is_sent, created_at)
		VALUES ($1, $2, $3, unnest($4::int[]), false, NOW())`

	_, err := r.db.ExecContext(ctx, query, eventID, userID, entities.ReminderTypeInApp, pq.Array(defaults))
	if err != nil {
		return fmt.Errorf("failed to create default reminders: %w", err)
	}
	return nil
}

func (r *EventReminderRepositoryPG) scanReminders(rows *sql.Rows) ([]*entities.EventReminder, error) {
	var reminders []*entities.EventReminder

	for rows.Next() {
		reminder := &entities.EventReminder{}
		err := rows.Scan(
			&reminder.ID, &reminder.EventID, &reminder.UserID, &reminder.ReminderType,
			&reminder.MinutesBefore, &reminder.IsSent, &reminder.SentAt, &reminder.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}
		reminders = append(reminders, reminder)
	}

	return reminders, nil
}
