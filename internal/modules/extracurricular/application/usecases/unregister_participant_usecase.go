package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
)

type unregisterParticipantRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.ExtracurricularEvent, error)
	RemoveParticipant(ctx context.Context, eventID, userID int64) error
}

// UnregisterParticipantUseCase removes caller's participant entry от
// the event.
type UnregisterParticipantUseCase struct {
	repo  unregisterParticipantRepo
	audit AuditSink
}

// NewUnregisterParticipantUseCase wires the use case.
func NewUnregisterParticipantUseCase(repo unregisterParticipantRepo, audit AuditSink) *UnregisterParticipantUseCase {
	if repo == nil {
		panic("extracurricular: NewUnregisterParticipantUseCase requires non-nil repo")
	}
	return &UnregisterParticipantUseCase{repo: repo, audit: audit}
}

// Execute loads the event (ErrEventNotFound surfaces для 404 mapping),
// then deletes the participant row.
func (uc *UnregisterParticipantUseCase) Execute(ctx context.Context, actorID int64, eventID int64) error {
	if _, err := uc.repo.GetByID(ctx, eventID); err != nil {
		return err
	}
	if err := uc.repo.RemoveParticipant(ctx, eventID, actorID); err != nil {
		return err
	}
	emitAudit(uc.audit, ctx, "extracurricular.participant_unregistered", actionFields(actorID, eventID))
	return nil
}
