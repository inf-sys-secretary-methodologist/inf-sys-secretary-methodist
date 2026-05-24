package usecases

import (
	"context"
	"errors"
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

// Execute registers actorID as participant в the event. Self-register
// only — teacher-on-behalf NOT в scope для backend slice (deferred).
// Pair 5 RED stub.
func (uc *RegisterParticipantUseCase) Execute(ctx context.Context, actorID int64, eventID int64) error {
	_ = ctx
	_ = actorID
	_ = eventID
	return errors.New("not implemented (Pair 5 RED stub)")
}
