package domain

// ProjectFilter defines filtering options for project queries.
type ProjectFilter struct {
	OwnerID *int64
	Status  *ProjectStatus
	Search  *string
}
