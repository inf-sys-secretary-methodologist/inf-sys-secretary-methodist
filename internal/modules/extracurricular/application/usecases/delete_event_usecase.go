package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
)

type deleteEventRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.ExtracurricularEvent, error)
	Delete(ctx context.Context, id int64) error
}

// DeleteEventUseCase removes the event row (participants cascade via
// FK ON DELETE CASCADE per migration 046) after authz + load.
type DeleteEventUseCase struct {
	repo  deleteEventRepo
	audit AuditSink
}

// NewDeleteEventUseCase wires the use case.
func NewDeleteEventUseCase(repo deleteEventRepo, audit AuditSink) *DeleteEventUseCase {
	if repo == nil {
		panic("extracurricular: NewDeleteEventUseCase requires non-nil repo")
	}
	return &DeleteEventUseCase{repo: repo, audit: audit}
}

// Execute deletes the event after loading it for authz check. Pair 5 RED stub.
func (uc *DeleteEventUseCase) Execute(ctx context.Context, actorID int64, actorRole string, isAdmin bool, eventID int64) error {
	_ = actorID
	_ = actorRole
	_ = isAdmin
	_ = eventID
	_ = ctx
	return errors.New("not implemented (Pair 5 RED stub)")
}
