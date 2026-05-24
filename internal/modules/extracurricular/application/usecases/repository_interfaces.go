// Package usecases hosts application services + repository ports for
// the extracurricular events module. Wide repository ports declared
// here (NOT в domain/) per CLAUDE.md DIP gate.
package usecases

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/repositories"
)

// EventRepository is the persistence port для ExtracurricularEvent
// aggregates. Sentinel contract per repositories package:
//
//	repositories.ErrEventNotFound       — missing row
//	repositories.ErrEventVersionConflict — stale Update
//
// Implementations should accept DBTX (not *sql.DB) so the same struct
// can run в either single-connection или transactional mode (mirror
// curriculum DBTX pattern).
type EventRepository interface {
	// Save inserts a new event row and writes the generated id back
	// onto the entity. version starts at 0.
	Save(ctx context.Context, e *entities.ExtracurricularEvent) error

	// GetByID returns the event with the given id together with its
	// participants list (eager-loaded for the aggregate's capacity +
	// duplicate invariants to hold в Register/Unregister flows).
	// repositories.ErrEventNotFound on missing row.
	GetByID(ctx context.Context, id int64) (*entities.ExtracurricularEvent, error)

	// Update writes the (already-mutated) event row back с optimistic
	// lock on version. RowsAffected == 0 followed by an existence
	// SELECT disambiguates:
	//   row exists → repositories.ErrEventVersionConflict (HTTP 409)
	//   row gone   → repositories.ErrEventNotFound (HTTP 404)
	// Participants are NOT touched by Update — use AddParticipant /
	// RemoveParticipant for inner-entity changes.
	Update(ctx context.Context, e *entities.ExtracurricularEvent) error

	// Delete removes the event row by id (participants cascade via FK
	// ON DELETE CASCADE per migration 046). Returns
	// repositories.ErrEventNotFound if the id did not exist.
	Delete(ctx context.Context, id int64) error

	// List returns a page of event summaries matching the filter +
	// total count (ignoring Limit/Offset). Empty result is not an
	// error. ParticipantCount populated via subquery — separate query
	// per event would be N+1.
	List(ctx context.Context, filter repositories.EventListFilter) (repositories.EventListResult, error)

	// AddParticipant inserts a participant row для (eventID, userID).
	// UNIQUE constraint violation surfaces via pgErrCode 23505 →
	// entities.ErrParticipantExists is wrapped by usecase before
	// reaching handler; repository wraps generic DB errors.
	AddParticipant(ctx context.Context, eventID, userID int64, registeredAt time.Time) error

	// RemoveParticipant deletes the participant row для (eventID,
	// userID). Returns entities.ErrParticipantNotFound if no such row.
	RemoveParticipant(ctx context.Context, eventID, userID int64) error
}
