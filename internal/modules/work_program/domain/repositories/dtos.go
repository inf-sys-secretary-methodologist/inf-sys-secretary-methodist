package repositories

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

// WorkProgramListFilter parameterizes List. All non-nil/non-empty
// fields combine with AND semantics. Empty filter returns every row
// (subject to Limit / Offset).
type WorkProgramListFilter struct {
	Status             *domain.Status // optional
	DisciplineID       *int64         // optional
	SpecialtyCode      string         // optional, empty = no filter
	ApplicableFromYear *int           // optional
	AuthorID           *int64         // optional, e.g. "my work programs"
	Limit              int            // pagination, > 0
	Offset             int            // pagination, ≥ 0
}

// WorkProgramListResult bundles the page items with the total count of
// matching rows (ignoring Limit / Offset) so the client can render
// pagination controls without a separate count query.
//
// Items carry root state only — child collections (Goals, Competences,
// etc.) are not hydrated in list responses to keep the list endpoint
// cheap. Callers needing full state should call GetByID.
type WorkProgramListResult struct {
	Items []ListItem
	Total int
}

// ListItem is the lightweight projection of a WorkProgram for list
// endpoints — root-only fields without inner aggregate slices.
type ListItem struct {
	ID                 int64
	DisciplineID       int64
	SpecialtyCode      string
	ApplicableFromYear int
	Title              string
	Status             domain.Status
	AuthorID           int64
	Version            int
}

// MinobrnaukiOrderListFilter parameterizes the MinobrnaukiOrderRepository
// List query (приказы Минобрнауки per ADR-11). Non-nil fields combine
// with AND semantics; an empty filter returns every order subject to
// Limit / Offset.
type MinobrnaukiOrderListFilter struct {
	ChangeScope *domain.MinobrnaukiOrderChangeScope // optional
	UploadedBy  *int64                              // optional, e.g. "orders I recorded"
	Limit       int                                 // pagination, > 0
	Offset      int                                 // pagination, ≥ 0
}

// MinobrnaukiOrderListResult bundles the page items with the total count
// of matching rows (ignoring Limit / Offset) so the client can render
// pagination controls without a separate count query.
type MinobrnaukiOrderListResult struct {
	Items []MinobrnaukiOrderListItem
	Total int
}

// MinobrnaukiOrderListItem is the read projection of a MinobrnaukiOrder
// for list endpoints. The order is a flat entity (no inner aggregates),
// so the projection carries every field; the affected-work-program set
// is a separate concern fetched via MinobrnaukiOrderRepository.FindAffected.
type MinobrnaukiOrderListItem struct {
	ID          int64
	OrderNumber string
	Title       string
	PublishedAt time.Time
	DocumentID  *int64
	ChangeScope domain.MinobrnaukiOrderChangeScope
	Summary     string
	UploadedBy  int64
	CreatedAt   time.Time
}
