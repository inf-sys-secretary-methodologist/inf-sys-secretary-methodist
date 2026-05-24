package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
)

// UpdateEventInput bundles update fields. Full-replace semantics
// (no pointer tri-state) — UI sends complete state.
type UpdateEventInput struct {
	ID             int64
	Title          string
	Description    string
	Category       entities.Category
	TargetAudience entities.TargetAudience
	Location       string
	StartAt        time.Time
	EndAt          time.Time
	MaxCapacity    *int
}

type updateEventRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.ExtracurricularEvent, error)
	Update(ctx context.Context, e *entities.ExtracurricularEvent) error
}

// UpdateEventUseCase applies a content edit к an existing event after
// loading + authz + invariant gates. Optimistic lock conflicts
// surface as repositories.ErrEventVersionConflict для handler 409.
type UpdateEventUseCase struct {
	repo     updateEventRepo
	audit    AuditSink
	notifier EventNotifier
	clock    func() time.Time
}

// NewUpdateEventUseCase wires the use case. Nil notifier defaults
// к noopNotifier (production wiring lands в v0.163.0 frontend slice).
func NewUpdateEventUseCase(repo updateEventRepo, audit AuditSink, notifier EventNotifier, clock func() time.Time) *UpdateEventUseCase {
	if repo == nil {
		panic("extracurricular: NewUpdateEventUseCase requires non-nil repo")
	}
	if notifier == nil {
		notifier = noopNotifier{}
	}
	if clock == nil {
		clock = time.Now
	}
	return &UpdateEventUseCase{repo: repo, audit: audit, notifier: notifier, clock: clock}
}

// Execute loads + authz + UpdateBasics + persists.
func (uc *UpdateEventUseCase) Execute(ctx context.Context, actorID int64, actorRole string, isAdmin bool, in UpdateEventInput) (*entities.ExtracurricularEvent, error) {
	e, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if err := entities.AuthorizeEventEdit(actorID, e.OrganizerID(), actorRole, isAdmin); err != nil {
		emitAudit(uc.audit, ctx, "extracurricular.event_update_denied",
			denialFields(actorID, in.ID, "not organizer / not admin", "forbidden"))
		return nil, err
	}
	if err := e.UpdateBasics(entities.UpdateEventBasicsParams{
		Title:          in.Title,
		Description:    in.Description,
		Category:       in.Category,
		TargetAudience: in.TargetAudience,
		Location:       in.Location,
		StartAt:        in.StartAt,
		EndAt:          in.EndAt,
		MaxCapacity:    in.MaxCapacity,
		Now:            uc.clock(),
	}); err != nil {
		if errors.Is(err, entities.ErrInvalidEvent) {
			emitAudit(uc.audit, ctx, "extracurricular.event_update_denied",
				denialFields(actorID, in.ID, err.Error(), "invalid"))
		}
		return nil, err
	}
	if err := uc.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "extracurricular.event_updated", actionFields(actorID, e.ID))
	uc.notifier.NotifyEventUpdated(ctx, e.ID, e.Title(), string(e.TargetAudience()))
	return e, nil
}
