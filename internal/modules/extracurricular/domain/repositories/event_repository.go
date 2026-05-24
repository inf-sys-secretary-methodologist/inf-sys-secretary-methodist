// Package repositories declares sentinels + read-model DTOs for the
// extracurricular events bounded context. Repository interfaces live в
// internal/modules/extracurricular/application/usecases per DIP — этот
// package keeps only domain values referenced both by usecases и
// infrastructure без a circular import.
package repositories

import "errors"

// ErrEventNotFound signals that no extracurricular_events row exists
// for the given id. Handlers map к HTTP 404.
var ErrEventNotFound = errors.New("extracurricular_event: not found")

// ErrEventVersionConflict signals that an Update attempted to write
// against a stale version (optimistic lock per plan ADR-5). The row
// still exists. Handlers map к HTTP 409.
var ErrEventVersionConflict = errors.New("extracurricular_event: version conflict")

// EventListFilter narrows ListEvents query result. All fields optional;
// zero-value means "no filter on this dimension". Result is also
// audience-filtered after fetch via CanViewEvent for non-admin callers.
type EventListFilter struct {
	// Status filters by lifecycle state (e.g. only `published`).
	// Empty string = no filter.
	Status string

	// Category filters by event classification. Empty = no filter.
	Category string

	// AudienceIn restricts по target_audience IN (...). Empty slice =
	// no filter (admin queries). Non-admin handlers populate with the
	// audience set visible к caller's role per ADR-6.
	AudienceIn []string

	// OrganizerID filters by event organizer. Zero = no filter.
	OrganizerID int64

	// FromDate / ToDate range filter on start_at. Zero = no bound.
	FromDate string // ISO date YYYY-MM-DD or empty
	ToDate   string

	// Limit + Offset for pagination. Limit==0 → default (100). Offset
	// negative → 0.
	Limit  int
	Offset int
}

// EventSummary is the projection returned by ListEvents — omits
// participants slice (loaded only by GetByID) and description (loaded
// by GetByID для detail view). ParticipantCount поднимается из
// extracurricular_participants table via subquery.
type EventSummary struct {
	ID               int64
	Title            string
	Category         string
	TargetAudience   string
	Status           string
	Location         string
	StartAt          string // ISO 8601
	EndAt            string
	MaxCapacity      *int
	OrganizerID      int64
	ParticipantCount int
	Version          int
	CreatedAt        string
	UpdatedAt        string
}

// EventListResult bundles the page slice + total count for paginated
// queries — pagination needs both для frontend "X of Y" rendering.
type EventListResult struct {
	Items []EventSummary
	Total int
}
