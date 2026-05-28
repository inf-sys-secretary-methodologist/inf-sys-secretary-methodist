package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// GenerateDraftUseCase fills an empty draft РПД with LLM-generated
// content (goals / competences / topics / references) per ADR-7.
//
// Pipeline: role gate (teacher / methodist / system_admin — the same
// authoring set as Create) → per-user rate limit (LLM cost guard) →
// load aggregate → ownership (author or system_admin; a non-owner
// collapses to NotFound per the OWASP IDOR carry-forward) →
// frozen-status guard (draft / needs_revision only) → discipline
// enrichment from curriculum → generation → map the result into the
// aggregate through its constructors (a malformed row fails the whole
// call) → persist.
type GenerateDraftUseCase struct {
	repo        generateDraftRepo
	generator   DraftGenerator
	disciplines DisciplineInfoProvider
	limiter     GenerationRateLimiter
	audit       AuditSink
}

// NewGenerateDraftUseCase wires the use case. repo / generator /
// disciplines / limiter are required (a nil dependency should fail at
// DI wiring, not deep inside a request). A nil audit sink is tolerated.
func NewGenerateDraftUseCase(
	repo generateDraftRepo,
	generator DraftGenerator,
	disciplines DisciplineInfoProvider,
	limiter GenerationRateLimiter,
	audit AuditSink,
) *GenerateDraftUseCase {
	if repo == nil {
		panic("work_program: NewGenerateDraftUseCase requires non-nil repo")
	}
	if generator == nil {
		panic("work_program: NewGenerateDraftUseCase requires non-nil generator")
	}
	if disciplines == nil {
		panic("work_program: NewGenerateDraftUseCase requires non-nil disciplines")
	}
	if limiter == nil {
		panic("work_program: NewGenerateDraftUseCase requires non-nil limiter")
	}
	return &GenerateDraftUseCase{
		repo:        repo,
		generator:   generator,
		disciplines: disciplines,
		limiter:     limiter,
		audit:       audit,
	}
}

// Execute runs the generation pipeline end-to-end and returns the
// filled aggregate on success. Domain denials wrap a sentinel so
// handler error mapping resolves via errors.Is; transport / upstream
// errors propagate without an audit event (the audit log records
// policy decisions, not infrastructure outages).
func (uc *GenerateDraftUseCase) Execute(
	ctx context.Context,
	actorID int64,
	actorRole string,
	workProgramID int64,
) (*entities.WorkProgram, error) {
	if !isAllowedToCreateWorkProgram(actorRole) {
		emitAudit(uc.audit, ctx, "work_program.generate_denied",
			denialFields(actorID, workProgramID, "forbidden_role", ""))
		return nil, fmt.Errorf("%w: role %q cannot generate work program draft",
			domain.ErrWorkProgramScopeForbidden, actorRole)
	}

	allowed, err := uc.limiter.Allow(ctx, actorID)
	if err != nil {
		return nil, fmt.Errorf("generate draft: rate limit check: %w", err)
	}
	if !allowed {
		emitAudit(uc.audit, ctx, "work_program.generate_denied",
			denialFields(actorID, workProgramID, "rate_limited", ""))
		return nil, fmt.Errorf("%w: actor %d", domain.ErrGenerationRateLimited, actorID)
	}

	wp, err := uc.repo.GetByID(ctx, workProgramID)
	if err != nil {
		// Not-found propagates without audit (ID typos / race deletes
		// are common and non-forensic); other errors propagate as-is.
		return nil, err
	}

	if !isAuthorOrSystemAdmin(actorID, actorRole, wp.AuthorID()) {
		// Collapse 403→404 so a non-owner cannot probe existence
		// (OWASP IDOR carry-forward).
		emitAudit(uc.audit, ctx, "work_program.generate_denied",
			denialFields(actorID, workProgramID, "not_owner", wp.SpecialtyCode()))
		return nil, fmt.Errorf("%w: id %d", repositories.ErrWorkProgramNotFound, workProgramID)
	}

	if wp.Status() != domain.StatusDraft && wp.Status() != domain.StatusNeedsRevision {
		emitAudit(uc.audit, ctx, "work_program.generate_denied",
			denialFields(actorID, workProgramID, "frozen_status", wp.SpecialtyCode()))
		return nil, fmt.Errorf("%w: status %q", domain.ErrCannotEditFrozenStatus, wp.Status())
	}

	info, err := uc.disciplines.GetDisciplineInfo(ctx, wp.DisciplineID())
	if err != nil {
		return nil, fmt.Errorf("generate draft: discipline lookup: %w", err)
	}

	result, err := uc.generator.GenerateDraft(ctx, DraftRequest{
		DisciplineName:     info.Name,
		SpecialtyCode:      wp.SpecialtyCode(),
		ApplicableFromYear: wp.ApplicableFromYear(),
		HoursLecture:       info.HoursLecture,
		HoursPractice:      info.HoursPractice,
		HoursLab:           info.HoursLab,
		HoursSelfStudy:     info.HoursSelfStudy,
		ControlForm:        info.ControlForm,
		Annotation:         wp.Annotation(),
	})
	if err != nil {
		return nil, fmt.Errorf("generate draft: %w", err)
	}

	if err := applyDraft(wp, result); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, wp); err != nil {
		return nil, err
	}

	fields := successFields(actorID, wp.ID(), wp.SpecialtyCode(), string(wp.Status()))
	fields["goals"] = len(result.Goals)
	fields["competences"] = len(result.Competences)
	fields["topics"] = len(result.Topics)
	fields["references"] = len(result.References)
	emitAudit(uc.audit, ctx, "work_program.generated", fields)
	return wp, nil
}

// applyDraft maps a DraftResult onto the aggregate via the domain
// constructors + AddX methods. Any invalid row (bad enum, empty field,
// duplicate competence code) aborts the whole mapping — the aggregate
// is discarded unpersisted, so partial in-memory mutation is harmless.
// order_index is the slice position so generated collections keep their
// emitted ordering.
func applyDraft(wp *entities.WorkProgram, r DraftResult) error {
	for i, text := range r.Goals {
		g, err := entities.NewGoal(text, i)
		if err != nil {
			return err
		}
		if err := wp.AddGoal(g); err != nil {
			return err
		}
	}
	for _, c := range r.Competences {
		comp, err := entities.NewCompetence(c.Code, domain.CompetenceType(c.Type), c.Description)
		if err != nil {
			return err
		}
		if err := wp.AddCompetence(comp); err != nil {
			return err
		}
	}
	for i, tp := range r.Topics {
		topic, err := entities.NewTopic(entities.NewTopicInput{
			Kind:       domain.TopicKind(tp.Kind),
			Title:      tp.Title,
			Hours:      tp.Hours,
			OrderIndex: i,
		})
		if err != nil {
			return err
		}
		if err := wp.AddTopic(topic); err != nil {
			return err
		}
	}
	for i, ref := range r.References {
		reference, err := entities.NewReference(entities.NewReferenceInput{
			Kind:       domain.ReferenceKind(ref.Kind),
			Citation:   ref.Citation,
			OrderIndex: i,
		})
		if err != nil {
			return err
		}
		if err := wp.AddReference(reference); err != nil {
			return err
		}
	}
	return nil
}
