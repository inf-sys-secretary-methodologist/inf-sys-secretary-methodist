package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
)

// UpdateEventInput bundles update fields. Pointer-less; full replace
// semantics (UI sends full state) — simpler than tri-state pointers
// для greenfield. MaxCapacity nil = explicit "unlimited".
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

// Execute runs UpdateBasics on the loaded aggregate с authz + optimistic
// lock conflict translation. Pair 5 RED stub.
func (uc *UpdateEventUseCase) Execute(ctx context.Context, actorID int64, actorRole string, isAdmin bool, in UpdateEventInput) (*entities.ExtracurricularEvent, error) {
	_ = actorID
	_ = actorRole
	_ = isAdmin
	_ = in
	_ = ctx
	return nil, errors.New("not implemented (Pair 5 RED stub)")
}
