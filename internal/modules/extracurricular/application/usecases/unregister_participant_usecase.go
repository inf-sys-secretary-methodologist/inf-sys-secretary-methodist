package usecases

import (
	"context"
	"errors"

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

// Execute removes actorID's participant entry от the event.
// Pair 5 RED stub.
func (uc *UnregisterParticipantUseCase) Execute(ctx context.Context, actorID int64, eventID int64) error {
	_ = ctx
	_ = actorID
	_ = eventID
	return errors.New("not implemented (Pair 5 RED stub)")
}
