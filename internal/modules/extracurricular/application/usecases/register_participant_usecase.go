package usecases

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
)

type registerParticipantRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.ExtracurricularEvent, error)
	AddParticipant(ctx context.Context, eventID, userID int64, registeredAt time.Time) error
}

// RegisterParticipantUseCase adds caller as participant к the event
// (self-register only; teacher-on-behalf out-of-scope для backend slice).
type RegisterParticipantUseCase struct {
	repo  registerParticipantRepo
	audit AuditSink
	clock func() time.Time
}

// NewRegisterParticipantUseCase wires the use case.
func NewRegisterParticipantUseCase(repo registerParticipantRepo, audit AuditSink, clock func() time.Time) *RegisterParticipantUseCase {
	if repo == nil {
		panic("extracurricular: NewRegisterParticipantUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &RegisterParticipantUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute loads the event, runs the aggregate Register method к check
// status + capacity + duplicate, then persists the participant row.
func (uc *RegisterParticipantUseCase) Execute(ctx context.Context, actorID int64, eventID int64) error {
	e, err := uc.repo.GetByID(ctx, eventID)
	if err != nil {
		return err
	}
	now := uc.clock()
	if err := e.Register(actorID, now); err != nil {
		emitAudit(uc.audit, ctx, "extracurricular.register_denied",
			denialFields(actorID, eventID, err.Error(), "register_blocked"))
		return err
	}
	if err := uc.repo.AddParticipant(ctx, eventID, actorID, now); err != nil {
		return err
	}
	emitAudit(uc.audit, ctx, "extracurricular.participant_registered", actionFields(actorID, eventID))
	return nil
}
