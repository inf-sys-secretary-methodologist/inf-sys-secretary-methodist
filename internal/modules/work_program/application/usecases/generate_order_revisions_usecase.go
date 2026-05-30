package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// AI bulk-revision (ADR-12): a methodist triggers LLM-assisted generation
// of draft листы актуализации (revisions) for every РПД affected by a
// recorded приказ Минобрнауки. The LLM drafts; a human (the РПД author)
// submits and the methodist approves via the existing revision flow
// (slice 10) — NOT a silent auto-apply, because a РПД is legally
// significant (Рособрнадзор) and LLM output must be reviewed.

// RevisionDraftRequest is the context handed to a RevisionDraftGenerator:
// the recorded order (number / title / summary / published year) plus the
// affected РПД's identity, so the LLM can propose a targeted актуализация.
type RevisionDraftRequest struct {
	OrderNumber        string
	OrderTitle         string
	OrderSummary       string
	PublishedYear      int
	WorkProgramTitle   string
	SpecialtyCode      string
	ApplicableFromYear int
}

// RevisionProposal is the generator's output for one РПД — the categorized
// change + a human summary + an optional structured before/after diff blob.
// Mapped into a domain Revision via NewRevision (a bad change_type fails the
// constructor rather than silently bypassing the invariant).
type RevisionProposal struct {
	ChangeType    string
	ChangeSummary string
	DiffPayload   []byte
}

// RevisionDraftGenerator is the outbound LLM port for bulk revision
// drafting (ADR-12). The concrete adapter (OpenRouter, reusing the slice 5
// stack) is wired in main.go (PR 11b); tests substitute a deterministic
// fake. Consumer-owned port (DIP).
type RevisionDraftGenerator interface {
	GenerateRevision(ctx context.Context, req RevisionDraftRequest) (RevisionProposal, error)
}

// orderRevisionSourceRepo loads the order plus its affected-РПД set.
type orderRevisionSourceRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.MinobrnaukiOrder, error)
	FindAffected(ctx context.Context, orderID int64) ([]int64, error)
}

// orderRevisionTargetRepo loads and persists one affected РПД.
type orderRevisionTargetRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// GenerateOrderRevisionsResult summarizes a bulk-generation run.
type GenerateOrderRevisionsResult struct {
	Generated int // draft revisions created + persisted
	Skipped   int // РПД not in approved/needs_revision (no edition to revise)
	Failures  int // per-РПД load / generate / add / persist errors
}

// GenerateOrderRevisionsUseCase generates a draft Revision for every
// affected РПД of a recorded order, via the LLM. Methodist-triggered.
type GenerateOrderRevisionsUseCase struct {
	orders  orderRevisionSourceRepo
	targets orderRevisionTargetRepo
	gen     RevisionDraftGenerator
	limiter GenerationRateLimiter // optional (nil tolerated)
	audit   AuditSink             // optional (nil tolerated)
}

// NewGenerateOrderRevisionsUseCase wires the use case. orders, targets and
// gen are required; limiter and audit are optional.
func NewGenerateOrderRevisionsUseCase(
	orders orderRevisionSourceRepo,
	targets orderRevisionTargetRepo,
	gen RevisionDraftGenerator,
	limiter GenerationRateLimiter,
	audit AuditSink,
) *GenerateOrderRevisionsUseCase {
	if orders == nil || targets == nil || gen == nil {
		panic("work_program: NewGenerateOrderRevisionsUseCase requires non-nil orders, targets and gen")
	}
	return &GenerateOrderRevisionsUseCase{
		orders:  orders,
		targets: targets,
		gen:     gen,
		limiter: limiter,
		audit:   audit,
	}
}

// Execute — RED stub, implemented in the GREEN step.
func (uc *GenerateOrderRevisionsUseCase) Execute(
	_ context.Context, _ int64, _ string, _ int64,
) (GenerateOrderRevisionsResult, error) {
	_ = uc.orders
	_ = uc.targets
	_ = uc.gen
	_ = uc.limiter
	_ = uc.audit
	return GenerateOrderRevisionsResult{}, nil
}
