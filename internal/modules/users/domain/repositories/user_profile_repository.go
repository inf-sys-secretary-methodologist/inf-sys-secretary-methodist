// Package repositories holds value-type query DTOs consumed by the
// users repository ports. The interfaces themselves moved к
// internal/modules/users/application/usecases per Clean Architecture
// DIP (CLAUDE.md gate) in v0.160.1; this package now contains only
// the filter shapes — mirror curriculum's domain/repositories
// (sentinels + query DTOs).
package repositories

// UserFilter contains filter options for listing users.
type UserFilter struct {
	DepartmentID *int64
	PositionID   *int64
	Role         string
	Status       string
	Search       string // search by name or email
}
