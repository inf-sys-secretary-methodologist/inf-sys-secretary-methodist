package usecases

import (
	"context"

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

// Execute loads + authz + deletes.
func (uc *DeleteEventUseCase) Execute(ctx context.Context, actorID int64, actorRole string, isAdmin bool, eventID int64) error {
	e, err := uc.repo.GetByID(ctx, eventID)
	if err != nil {
		return err
	}
	if err := entities.AuthorizeEventEdit(actorID, e.OrganizerID(), actorRole, isAdmin); err != nil {
		emitAudit(uc.audit, ctx, "extracurricular.event_delete_denied",
			denialFields(actorID, eventID, "not organizer / not admin", "forbidden"))
		return err
	}
	if err := uc.repo.Delete(ctx, eventID); err != nil {
		return err
	}
	emitAudit(uc.audit, ctx, "extracurricular.event_deleted", actionFields(actorID, eventID))
	return nil
}
