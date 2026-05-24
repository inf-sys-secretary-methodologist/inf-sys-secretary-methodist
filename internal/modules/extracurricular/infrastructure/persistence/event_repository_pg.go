package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/repositories"
)

// Compile-time assertion: PG impl satisfies the wide port в the
// consuming application/usecases layer (DIP per CLAUDE.md gate).
var _ usecases.EventRepository = (*EventRepositoryPG)(nil)

// EventRepositoryPG persists ExtracurricularEvent aggregates against
// the extracurricular_events + extracurricular_participants tables
// (migration 046). Optimistic locking per plan ADR-5 — Update uses
// WHERE id = ? AND version = ? и disambiguates RowsAffected == 0 via
// a follow-up existence SELECT.
type EventRepositoryPG struct {
	db DBTX
}

// NewEventRepositoryPG constructs the repository. db may be *sql.DB
// (default DI) или *sql.Tx (future bulk path).
func NewEventRepositoryPG(db DBTX) *EventRepositoryPG {
	return &EventRepositoryPG{db: db}
}

// pqUniqueViolation is the SQLSTATE code for a unique-constraint
// violation. Mirror к curriculum's local copy — keeps the
// extracurricular bounded context free of a cross-module dependency
// on shared error mappers.
const pqUniqueViolation = "23505"

const eventSelectColumns = `id, title, description, category, target_audience, status,
	location, start_at, end_at, max_capacity, organizer_id, version, created_at, updated_at`

// Save inserts a new event row and writes the generated id back onto
// the entity. Participants slice is NOT persisted here — fresh events
// have no participants.
func (r *EventRepositoryPG) Save(ctx context.Context, e *entities.ExtracurricularEvent) error {
	const query = `
		INSERT INTO extracurricular_events (
			title, description, category, target_audience, status,
			location, start_at, end_at, max_capacity, organizer_id,
			version, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id`

	var newID int64
	err := r.db.QueryRowContext(ctx, query,
		e.Title(),
		nullableText(e.Description()),
		string(e.Category()),
		string(e.TargetAudience()),
		string(e.Status()),
		nullableText(e.Location()),
		e.StartAt(),
		e.EndAt(),
		nullableIntPtr(e.MaxCapacity()),
		e.OrganizerID(),
		e.Version(),
		e.CreatedAt(),
		e.UpdatedAt(),
	).Scan(&newID)
	if err != nil {
		return fmt.Errorf("extracurricular: save: %w", err)
	}
	e.ID = newID
	return nil
}

// GetByID returns the event together с its participants list. Two
// queries — events row + participants slice. Returns
// repositories.ErrEventNotFound on missing event row.
func (r *EventRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.ExtracurricularEvent, error) {
	const query = `SELECT ` + eventSelectColumns + ` FROM extracurricular_events WHERE id = $1`
	var (
		idv            int64
		title          string
		description    sql.NullString
		category       string
		targetAudience string
		status         string
		location       sql.NullString
		startAt        time.Time
		endAt          time.Time
		maxCapacity    sql.NullInt64
		organizerID    int64
		version        int
		createdAt      time.Time
		updatedAt      time.Time
	)
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&idv, &title, &description, &category, &targetAudience, &status,
		&location, &startAt, &endAt, &maxCapacity, &organizerID, &version,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrEventNotFound
		}
		return nil, fmt.Errorf("extracurricular: get by id: %w", err)
	}
	participants, err := r.loadParticipants(ctx, idv)
	if err != nil {
		return nil, err
	}
	var maxCapPtr *int
	if maxCapacity.Valid {
		v := int(maxCapacity.Int64)
		maxCapPtr = &v
	}
	return entities.ReconstituteExtracurricularEvent(
		idv, title, description.String,
		entities.Category(category), entities.TargetAudience(targetAudience),
		entities.Status(status),
		location.String, startAt, endAt, maxCapPtr, organizerID,
		participants, version, createdAt, updatedAt,
	), nil
}

func (r *EventRepositoryPG) loadParticipants(ctx context.Context, eventID int64) ([]entities.Participant, error) {
	const query = `SELECT user_id, registered_at FROM extracurricular_participants
		WHERE event_id = $1 ORDER BY registered_at ASC, user_id ASC`
	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("extracurricular: load participants: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []entities.Participant
	for rows.Next() {
		var userID int64
		var registeredAt time.Time
		if err := rows.Scan(&userID, &registeredAt); err != nil {
			return nil, fmt.Errorf("extracurricular: scan participant: %w", err)
		}
		out = append(out, entities.Participant{UserID: userID, RegisteredAt: registeredAt})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("extracurricular: iter participants: %w", err)
	}
	return out, nil
}

// Update writes event row back с optimistic lock (WHERE id = ? AND
// version = ?). RowsAffected == 0 → existence probe disambiguates
// version conflict vs vanished row.
func (r *EventRepositoryPG) Update(ctx context.Context, e *entities.ExtracurricularEvent) error {
	const query = `
		UPDATE extracurricular_events SET
			title = $1, description = $2, category = $3, target_audience = $4,
			status = $5, location = $6, start_at = $7, end_at = $8,
			max_capacity = $9, version = version + 1, updated_at = $10
		WHERE id = $11 AND version = $12`
	res, err := r.db.ExecContext(ctx, query,
		e.Title(), nullableText(e.Description()),
		string(e.Category()), string(e.TargetAudience()), string(e.Status()),
		nullableText(e.Location()), e.StartAt(), e.EndAt(),
		nullableIntPtr(e.MaxCapacity()), e.UpdatedAt(),
		e.ID, e.Version(),
	)
	if err != nil {
		return fmt.Errorf("extracurricular: update: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("extracurricular: update: rows affected: %w", err)
	}
	if n == 0 {
		return r.disambiguateAbsentUpdate(ctx, e.ID)
	}
	return nil
}

func (r *EventRepositoryPG) disambiguateAbsentUpdate(ctx context.Context, id int64) error {
	const probe = `SELECT 1 FROM extracurricular_events WHERE id = $1`
	var found int
	err := r.db.QueryRowContext(ctx, probe, id).Scan(&found)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repositories.ErrEventNotFound
		}
		return fmt.Errorf("extracurricular: disambiguate: %w", err)
	}
	return repositories.ErrEventVersionConflict
}

// Delete removes the event row by id. Participants cascade via FK
// ON DELETE CASCADE per migration 046. RowsAffected == 0 →
// ErrEventNotFound.
func (r *EventRepositoryPG) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM extracurricular_events WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("extracurricular: delete: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("extracurricular: delete: rows affected: %w", err)
	}
	if n == 0 {
		return repositories.ErrEventNotFound
	}
	return nil
}

// List returns a page of event summaries matching the filter + total
// count. ParticipantCount populated via correlated subquery (no N+1).
func (r *EventRepositoryPG) List(ctx context.Context, filter repositories.EventListFilter) (repositories.EventListResult, error) {
	where, args := buildEventWhere(filter)
	countQuery := `SELECT COUNT(*) FROM extracurricular_events ` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return repositories.EventListResult{}, fmt.Errorf("extracurricular: list count: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	offset := max(filter.Offset, 0)
	args = append(args, limit, offset)

	pageQuery := `SELECT id, title, category, target_audience, status, location,
		start_at, end_at, max_capacity, organizer_id, version, created_at, updated_at,
		(SELECT COUNT(*) FROM extracurricular_participants WHERE event_id = extracurricular_events.id) AS participant_count
		FROM extracurricular_events ` + where +
		fmt.Sprintf(` ORDER BY start_at ASC, id ASC LIMIT $%d OFFSET $%d`, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, pageQuery, args...)
	if err != nil {
		return repositories.EventListResult{}, fmt.Errorf("extracurricular: list page: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []repositories.EventSummary
	for rows.Next() {
		var (
			s             repositories.EventSummary
			location      sql.NullString
			maxCap        sql.NullInt64
			startAt       time.Time
			endAt         time.Time
			createdAt     time.Time
			updatedAt     time.Time
			participantCt int
		)
		if err := rows.Scan(&s.ID, &s.Title, &s.Category, &s.TargetAudience, &s.Status,
			&location, &startAt, &endAt, &maxCap, &s.OrganizerID, &s.Version,
			&createdAt, &updatedAt, &participantCt); err != nil {
			return repositories.EventListResult{}, fmt.Errorf("extracurricular: list scan: %w", err)
		}
		s.Location = location.String
		if maxCap.Valid {
			v := int(maxCap.Int64)
			s.MaxCapacity = &v
		}
		s.StartAt = startAt.Format(time.RFC3339)
		s.EndAt = endAt.Format(time.RFC3339)
		s.CreatedAt = createdAt.Format(time.RFC3339)
		s.UpdatedAt = updatedAt.Format(time.RFC3339)
		s.ParticipantCount = participantCt
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return repositories.EventListResult{}, fmt.Errorf("extracurricular: list iter: %w", err)
	}
	return repositories.EventListResult{Items: items, Total: total}, nil
}

// buildEventWhere assembles dynamic WHERE clause + arg slice from the
// filter. Returns clause string starting with "WHERE ..." (or empty
// if no filters) plus positional args. #nosec G201 — all clause
// fragments are static literal strings; only user values bind via $N.
func buildEventWhere(f repositories.EventListFilter) (string, []any) {
	var clauses []string
	var args []any
	idx := 1
	if f.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}
	if f.Category != "" {
		clauses = append(clauses, fmt.Sprintf("category = $%d", idx))
		args = append(args, f.Category)
		idx++
	}
	if len(f.AudienceIn) > 0 {
		clauses = append(clauses, fmt.Sprintf("target_audience = ANY($%d)", idx))
		args = append(args, pq.Array(f.AudienceIn))
		idx++
	}
	if f.OrganizerID > 0 {
		clauses = append(clauses, fmt.Sprintf("organizer_id = $%d", idx))
		args = append(args, f.OrganizerID)
		idx++
	}
	if f.FromDate != "" {
		clauses = append(clauses, fmt.Sprintf("start_at >= $%d", idx))
		args = append(args, f.FromDate)
		idx++
	}
	if f.ToDate != "" {
		clauses = append(clauses, fmt.Sprintf("start_at <= $%d", idx))
		args = append(args, f.ToDate)
	}
	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

// AddParticipant inserts a (event_id, user_id) row. UNIQUE constraint
// violation maps to entities.ErrParticipantExists.
func (r *EventRepositoryPG) AddParticipant(ctx context.Context, eventID, userID int64, registeredAt time.Time) error {
	const query = `INSERT INTO extracurricular_participants (event_id, user_id, registered_at)
		VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, eventID, userID, registeredAt)
	if err != nil {
		if isUniqueViolation(err) {
			return entities.ErrParticipantExists
		}
		return fmt.Errorf("extracurricular: add participant: %w", err)
	}
	return nil
}

// RemoveParticipant deletes the (event_id, user_id) row. 0 rows →
// entities.ErrParticipantNotFound.
func (r *EventRepositoryPG) RemoveParticipant(ctx context.Context, eventID, userID int64) error {
	const query = `DELETE FROM extracurricular_participants WHERE event_id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, query, eventID, userID)
	if err != nil {
		return fmt.Errorf("extracurricular: remove participant: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("extracurricular: remove participant: rows affected: %w", err)
	}
	if n == 0 {
		return entities.ErrParticipantNotFound
	}
	return nil
}

// nullableText maps empty Go string to SQL NULL so the description /
// location columns stay NULL (the migration leaves both nullable).
func nullableText(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// nullableIntPtr maps *int → sql.NullInt64 — nil = NULL.
func nullableIntPtr(p *int) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*p), Valid: true}
}

// isUniqueViolation reports whether err is a PostgreSQL unique
// violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == pqUniqueViolation
	}
	return false
}
