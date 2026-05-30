package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
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
	OrderText          string // full text extracted from the attached приказ document (slice 7); empty if none
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

// OrderDocumentTextProvider fetches the extracted plain-text content of the
// order's attached document (PDF/DOCX) so the LLM can ground its revision
// proposal on the real приказ rather than only the manual summary (slice 7).
// Consumer-owned port (DIP); the adapter bridging to the documents +
// text-extraction infrastructure is wired at the composition root. Optional
// collaborator — when nil (or the order has no document) generation works
// from the manual OrderSummary alone.
type OrderDocumentTextProvider interface {
	GetDocumentText(ctx context.Context, documentID int64) (string, error)
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
	limiter GenerationRateLimiter     // optional (nil tolerated)
	audit   AuditSink                 // optional (nil tolerated)
	docText OrderDocumentTextProvider // optional (nil tolerated): order document text for the LLM
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

// WithDocumentText attaches the optional provider that supplies the extracted
// text of the order's attached document to the LLM (slice 7). Chainable;
// nil leaves bulk-revision working from the manual OrderSummary alone.
func (uc *GenerateOrderRevisionsUseCase) WithDocumentText(p OrderDocumentTextProvider) *GenerateOrderRevisionsUseCase {
	uc.docText = p
	return uc
}

// Execute runs the bulk-revision generation for one order:
//  1. Authorize: only order-managing staff (methodist / academic_secretary
//     / system_admin) may trigger generation → otherwise
//     ErrMinobrnaukiOrderScopeForbidden.
//  2. Rate-limit the whole bulk run once per actor (optional limiter) —
//     guards the cost of N LLM calls → ErrGenerationRateLimited.
//  3. Load the order (ErrMinobrnaukiOrderNotFound propagates) + its
//     affected-РПД set.
//  4. For each affected РПД in a revisable status (approved /
//     needs_revision): ask the LLM for a proposal, map it into a draft
//     Revision authored by the РПД author (NOT the triggering methodist —
//     separation of duties), append it via AddRevision (parent-status +
//     monotonic-number gate), and persist. Non-revisable programs are
//     skipped; a generated revision stays in draft for the author to
//     submit and the methodist to approve via the existing flow.
//
// The pass is best-effort: a per-РПД load / generate / add / persist error
// is counted in Failures and never aborts the remaining programs.
func (uc *GenerateOrderRevisionsUseCase) Execute(
	ctx context.Context, actorID int64, actorRole string, orderID int64,
) (GenerateOrderRevisionsResult, error) {
	var res GenerateOrderRevisionsResult

	if !isAllowedToRecordMinobrnaukiOrder(actorRole) {
		emitOrderAudit(uc.audit, ctx, "minobrnauki_order.revision_generation_denied", map[string]any{
			"actor_user_id":        actorID,
			"minobrnauki_order_id": orderID,
			"reason":               "forbidden",
		})
		return res, fmt.Errorf("%w: role %q cannot generate order revisions",
			domain.ErrMinobrnaukiOrderScopeForbidden, actorRole)
	}

	if uc.limiter != nil {
		ok, err := uc.limiter.Allow(ctx, actorID)
		if err != nil {
			return res, err
		}
		if !ok {
			return res, fmt.Errorf("%w: actor %d exceeded the generation rate limit",
				domain.ErrGenerationRateLimited, actorID)
		}
	}

	order, err := uc.orders.GetByID(ctx, orderID)
	if err != nil {
		return res, err
	}
	affected, err := uc.orders.FindAffected(ctx, orderID)
	if err != nil {
		return res, err
	}

	// Fetch the order's attached-document text once (slice 7), reused for
	// every affected РПД's request. Best-effort: when no provider is wired,
	// the order has no document, or extraction fails, OrderText stays empty
	// and the LLM falls back to the manual OrderSummary — a missing document
	// must never block the bulk run.
	var orderText string
	if uc.docText != nil && order.DocumentID() != nil {
		if txt, derr := uc.docText.GetDocumentText(ctx, *order.DocumentID()); derr == nil {
			orderText = txt
		} else {
			// Best-effort still proceeds on the manual summary, but a swallowed
			// extraction error must stay observable — otherwise a systematic
			// failure (S3 down, corrupt file) looks like a clean run.
			emitOrderAudit(uc.audit, ctx, "minobrnauki_order.document_text_unavailable", map[string]any{
				"actor_user_id":        actorID,
				"minobrnauki_order_id": orderID,
				"document_id":          *order.DocumentID(),
				"reason":               "extraction_failed",
			})
		}
	}

	for _, wpID := range affected {
		wp, err := uc.targets.GetByID(ctx, wpID)
		if err != nil {
			res.Failures++
			continue
		}
		if wp.Status() != domain.StatusApproved && wp.Status() != domain.StatusNeedsRevision {
			res.Skipped++
			continue
		}

		proposal, err := uc.gen.GenerateRevision(ctx, RevisionDraftRequest{
			OrderNumber:        order.OrderNumber(),
			OrderTitle:         order.Title(),
			OrderSummary:       order.Summary(),
			OrderText:          orderText,
			PublishedYear:      order.PublishedAt().Year(),
			WorkProgramTitle:   wp.Title(),
			SpecialtyCode:      wp.SpecialtyCode(),
			ApplicableFromYear: wp.ApplicableFromYear(),
		})
		if err != nil {
			res.Failures++
			continue
		}

		rev, err := entities.NewRevision(entities.NewRevisionInput{
			WorkProgramID:  wp.ID(),
			RevisionNumber: wp.NextRevisionNumber(),
			ChangeType:     domain.RevisionChangeType(proposal.ChangeType),
			ChangeSummary:  proposal.ChangeSummary,
			AuthorID:       wp.AuthorID(),
			DiffPayload:    proposal.DiffPayload,
		})
		if err != nil {
			res.Failures++
			continue
		}
		if err := wp.AddRevision(rev); err != nil {
			res.Failures++
			continue
		}
		if err := uc.targets.Update(ctx, wp); err != nil {
			res.Failures++
			continue
		}
		res.Generated++
	}

	emitOrderAudit(uc.audit, ctx, "minobrnauki_order.revisions_generated", map[string]any{
		"actor_user_id":        actorID,
		"minobrnauki_order_id": orderID,
		"order_number":         order.OrderNumber(),
		"generated":            res.Generated,
		"skipped":              res.Skipped,
		"failures":             res.Failures,
	})
	return res, nil
}
