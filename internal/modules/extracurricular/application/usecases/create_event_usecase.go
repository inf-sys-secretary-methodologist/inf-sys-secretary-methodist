package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
)

// CreateEventInput is the public DTO for CreateEvent.
type CreateEventInput struct {
	Title          string
	Description    string
	Category       entities.Category
	TargetAudience entities.TargetAudience
	Location       string
	StartAt        time.Time
	EndAt          time.Time
	MaxCapacity    *int
}

// createEventRepo is the narrow port — CreateEvent only needs Save.
type createEventRepo interface {
	Save(ctx context.Context, e *entities.ExtracurricularEvent) error
}

// CreateEventUseCase persists a fresh event after authorization.
type CreateEventUseCase struct {
	repo  createEventRepo
	audit AuditSink
	clock func() time.Time
}

// NewCreateEventUseCase wires the use case. Nil repo panics so
// misconfiguration surfaces at DI time, not on first request.
func NewCreateEventUseCase(repo createEventRepo, audit AuditSink, clock func() time.Time) *CreateEventUseCase {
	if repo == nil {
		panic("extracurricular: NewCreateEventUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &CreateEventUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the use case:
//  1. AuthorizeEventCreate (role gate; out-of-context → 'forbidden' denial)
//  2. NewExtracurricularEvent (invariant gate; ErrInvalidEvent → 'invalid')
//  3. repo.Save
//  4. audit "extracurricular.event_created" on success
//
// Pair 5 RED stub returns runtime error until GREEN impl.
func (uc *CreateEventUseCase) Execute(ctx context.Context, actorID int64, actorRole string, isAdmin bool, in CreateEventInput) (*entities.ExtracurricularEvent, error) {
	_ = actorID
	_ = actorRole
	_ = isAdmin
	_ = in
	_ = ctx
	return nil, errors.New("not implemented (Pair 5 RED stub)")
}
