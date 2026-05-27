package repositories

import (
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
