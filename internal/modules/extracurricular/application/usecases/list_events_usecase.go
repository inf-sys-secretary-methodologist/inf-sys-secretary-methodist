package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/repositories"
)

// ListEventsInput is the public DTO для ListEvents. Matches handler
// query params — Status/Category/FromDate/ToDate strings, OrganizerID
// optional, pagination Limit+Offset.
type ListEventsInput struct {
	Status      string
	Category    string
	OrganizerID int64
	FromDate    string
	ToDate      string
	Limit       int
	Offset      int
}

type listEventsRepo interface {
	List(ctx context.Context, filter repositories.EventListFilter) (repositories.EventListResult, error)
}

// ListEventsUseCase returns a paginated, audience-filtered slice of
// event summaries.
type ListEventsUseCase struct {
	repo listEventsRepo
}

// NewListEventsUseCase wires the read-side use case.
func NewListEventsUseCase(repo listEventsRepo) *ListEventsUseCase {
	if repo == nil {
		panic("extracurricular: NewListEventsUseCase requires non-nil repo")
	}
	return &ListEventsUseCase{repo: repo}
}

// Execute applies the audience visibility filter за каллер's role
// (admin = all audiences; per-role = restricted set) и delegates к
// repo.List. Pair 5 RED stub.
func (uc *ListEventsUseCase) Execute(ctx context.Context, actorRole string, isAdmin bool, in ListEventsInput) (repositories.EventListResult, error) {
	_ = ctx
	_ = actorRole
	_ = isAdmin
	_ = in
	return repositories.EventListResult{}, errors.New("not implemented (Pair 5 RED stub)")
}
