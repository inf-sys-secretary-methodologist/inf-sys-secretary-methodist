package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
)

type getEventRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.ExtracurricularEvent, error)
}

// GetEventUseCase fetches one event by id и applies audience filter
// для non-admin callers.
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

// Execute fetches the event и filters via CanViewEvent + isAdmin
// для non-admin callers. Returns ErrEventNotFound для audience
// mismatch — hides existence per security best practice. Pair 5 RED stub.
func (uc *GetEventUseCase) Execute(ctx context.Context, actorRole string, isAdmin bool, eventID int64) (*entities.ExtracurricularEvent, error) {
	_ = ctx
	_ = actorRole
	_ = isAdmin
	_ = eventID
	return nil, errors.New("not implemented (Pair 5 RED stub)")
}
