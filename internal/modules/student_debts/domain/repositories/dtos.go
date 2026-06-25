package repositories

import (
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

// StudentDebtListFilter parameterizes List. All non-nil/non-empty fields
// combine with AND semantics; an empty filter returns every row subject
// to Limit / Offset.
type StudentDebtListFilter struct {
	GroupName     string               // optional, empty = no filter
	Status        *entities.DebtStatus // optional
	Semester      *int                 // optional
	StudentUserID *int64               // optional, e.g. "my debts"
	// DisciplineIDs restricts results to debts linked to any of these
	// curriculum_section_items ids. Used for teacher scoping (the
	// disciplines a teacher owns). nil/empty disables the predicate; a
	// non-empty slice that matches no row yields an empty page (a teacher
	// with zero owned disciplines sees nothing, which is correct).
	DisciplineIDs []int64
	Limit         int // pagination, > 0
	Offset        int // pagination, ≥ 0
}

// StudentDebtListResult bundles the page items with the total count of
// matching rows (ignoring Limit / Offset) so the client can render
// pagination controls without a separate count query.
//
// Items carry root state only — resit attempts are not hydrated in list
// responses to keep the list endpoint cheap. Callers needing the full
// aggregate (attempt timeline) should call GetByID.
type StudentDebtListResult struct {
	Items []StudentDebtListItem
	Total int
}

// StudentDebtListItem is the lightweight read projection of a StudentDebt
// for list endpoints — root-only fields without the attempts slice.
type StudentDebtListItem struct {
	ID              int64
	StudentFullName string
	GroupName       string
	DisciplineName  string
	Semester        int
	ControlForm     entities.ControlForm
	StudentUserID   *int64
	Status          entities.DebtStatus
	Version         int
}
