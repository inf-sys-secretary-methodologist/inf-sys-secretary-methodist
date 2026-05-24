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

type createEventRepo interface {
	Save(ctx context.Context, e *entities.ExtracurricularEvent) error
}

// CreateEventUseCase persists a fresh event after authorization
// against the actor's role.
type CreateEventUseCase struct {
	repo  createEventRepo
	audit AuditSink
	clock func() time.Time
}

// NewCreateEventUseCase wires the use case.
func NewCreateEventUseCase(repo createEventRepo, audit AuditSink, clock func() time.Time) *CreateEventUseCase {
	if repo == nil {
		panic("extracurricular: NewCreateEventUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &CreateEventUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the use case end-to-end:
//
//  1. AuthorizeEventCreate (role gate)
//  2. NewExtracurricularEvent (invariant gate)
//  3. repo.Save
//  4. emit extracurricular.event_created audit
func (uc *CreateEventUseCase) Execute(ctx context.Context, actorID int64, actorRole string, isAdmin bool, in CreateEventInput) (*entities.ExtracurricularEvent, error) {
	if err := entities.AuthorizeEventCreate(actorRole, isAdmin); err != nil {
		emitAudit(uc.audit, ctx, "extracurricular.event_create_denied",
			denialFields(actorID, 0, "role not allowed", "forbidden"))
		return nil, err
	}
	e, err := entities.NewExtracurricularEvent(entities.NewExtracurricularEventParams{
		Title:          in.Title,
		Description:    in.Description,
		Category:       in.Category,
		TargetAudience: in.TargetAudience,
		Location:       in.Location,
		StartAt:        in.StartAt,
		EndAt:          in.EndAt,
		MaxCapacity:    in.MaxCapacity,
		OrganizerID:    actorID,
		Now:            uc.clock(),
	})
	if err != nil {
		if errors.Is(err, entities.ErrInvalidEvent) {
			emitAudit(uc.audit, ctx, "extracurricular.event_create_denied",
				denialFields(actorID, 0, err.Error(), "invalid"))
		}
		return nil, err
	}
	if err := uc.repo.Save(ctx, e); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "extracurricular.event_created", actionFields(actorID, e.ID))
	return e, nil
}
