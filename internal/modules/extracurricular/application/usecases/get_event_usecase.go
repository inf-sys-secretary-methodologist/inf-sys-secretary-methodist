package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/repositories"
)

type getEventRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.ExtracurricularEvent, error)
}

// GetEventUseCase fetches one event by id и applies audience filter
// для non-admin callers. Audience mismatch returns ErrEventNotFound
// to hide existence per security best practice.
type GetEventUseCase struct {
	repo getEventRepo
}

// NewGetEventUseCase wires the read-side use case.
func NewGetEventUseCase(repo getEventRepo) *GetEventUseCase {
	if repo == nil {
		panic("extracurricular: NewGetEventUseCase requires non-nil repo")
	}
	return &GetEventUseCase{repo: repo}
}

// Execute fetches the event и applies audience filter.
func (uc *GetEventUseCase) Execute(ctx context.Context, actorRole string, isAdmin bool, eventID int64) (*entities.ExtracurricularEvent, error) {
	e, err := uc.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if isAdmin {
		return e, nil
	}
	if !entities.CanViewEvent(actorRole, e.TargetAudience()) {
		return nil, repositories.ErrEventNotFound
	}
	return e, nil
}
