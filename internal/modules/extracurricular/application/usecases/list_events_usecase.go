package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/repositories"
)

// ListEventsInput is the public DTO для ListEvents. Mirror к
// handler query params: optional status / category / organizer
// filter + date range + pagination.
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

// Execute applies role-aware audience visibility filter then delegates
// к repo.List. Admin sees all audiences (empty AudienceIn → no filter);
// non-admin restricted к the set visible per ADR-6.
func (uc *ListEventsUseCase) Execute(ctx context.Context, actorRole string, isAdmin bool, in ListEventsInput) (repositories.EventListResult, error) {
	filter := repositories.EventListFilter{
		Status:      in.Status,
		Category:    in.Category,
		OrganizerID: in.OrganizerID,
		FromDate:    in.FromDate,
		ToDate:      in.ToDate,
		Limit:       in.Limit,
		Offset:      in.Offset,
	}
	if !isAdmin {
		filter.AudienceIn = visibleAudiences(actorRole)
	}
	return uc.repo.List(ctx, filter)
}

// visibleAudiences returns the subset of TargetAudience values
// visible к the given role per ADR-6. Mirror к CanViewEvent matrix.
func visibleAudiences(role string) []string {
	switch role {
	case "student":
		return []string{string(entities.TargetAudienceAll), string(entities.TargetAudienceStudents)}
	case "teacher":
		return []string{string(entities.TargetAudienceAll), string(entities.TargetAudienceTeachers)}
	case "methodist", "academic_secretary", "system_admin":
		return []string{
			string(entities.TargetAudienceAll),
			string(entities.TargetAudienceStaff),
		}
	}
	// unknown role — only `all` visible (mirror CanViewEvent default-allow для all)
	return []string{string(entities.TargetAudienceAll)}
}
