package usecases

import (
	"context"
	"fmt"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// Pagination policy for the List use case. Centralized here (rather
// than in the handler) so any future caller — internal scheduler,
// batch job, alternate transport — inherits the same boundedness.
const (
	defaultListWPLimit = 50
	maxListWPLimit     = 200
)

// ListWorkProgramsInput mirrors WorkProgramListFilter at the use-case
// boundary so handlers don't import persistence DTOs directly. Status
// is wire-form (*string) — the use case converts to typed
// domain.Status before dispatching.
type ListWorkProgramsInput struct {
	Status             *string
	DisciplineID       *int64
	SpecialtyCode      string
	ApplicableFromYear *int
	AuthorID           *int64
	Limit              int
	Offset             int
}

// ListWorkProgramsResult is the public response shape. Items carry
// root-only state (per the persistence layer's ListItem projection)
// so list endpoints stay cheap; callers needing full child hydration
// follow up with GetWorkProgramUseCase.
type ListWorkProgramsResult struct {
	Items []repositories.ListItem
	Total int
}

// listWorkProgramsRepo is the narrow port: only List.
type listWorkProgramsRepo interface {
	List(ctx context.Context, filter repositories.WorkProgramListFilter) (repositories.WorkProgramListResult, error)
}

// ListWorkProgramsUseCase loads a role-scoped page of WorkPrograms.
// Row-level access policy per ADR-018 ADR-5 is applied here as filter
// overrides — the use case rewrites the inbound filter before dispatch
// rather than post-filtering repo results (cheaper, no over-fetch).
type ListWorkProgramsUseCase struct {
	repo  listWorkProgramsRepo
	audit AuditSink
}

// NewListWorkProgramsUseCase wires the use case. Repo is required.
func NewListWorkProgramsUseCase(repo listWorkProgramsRepo, audit AuditSink) *ListWorkProgramsUseCase {
	if repo == nil {
		panic("work_program: NewListWorkProgramsUseCase requires non-nil repo")
	}
	return &ListWorkProgramsUseCase{repo: repo, audit: audit}
}

// Execute runs the list flow:
//  1. Convert wire-form input → repository filter (status: *string → *domain.Status).
//  2. Apply role-driven filter overrides:
//     - system_admin / methodist / academic_secretary → pass-through
//     - teacher → AuthorID forced to actor id (closes "list someone
//     else's drafts" enumeration; cross-reference of approved WPs
//     from other authors goes through Get via deep link)
//     - student → Status forced to approved (273-ФЗ ст. 29 mandatory
//     openness covers approved only)
//     - unknown role → ErrWorkProgramScopeForbidden + audit
//  3. Apply pagination defaults / clamps (zero/negative limit → 50;
//     over-max → 200; negative offset → 0).
//  4. Dispatch to repo.List.
func (uc *ListWorkProgramsUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in ListWorkProgramsInput) (ListWorkProgramsResult, error) {
	filter := repositories.WorkProgramListFilter{
		DisciplineID:       in.DisciplineID,
		SpecialtyCode:      in.SpecialtyCode,
		ApplicableFromYear: in.ApplicableFromYear,
		AuthorID:           in.AuthorID,
	}
	if in.Status != nil {
		s := domain.Status(*in.Status)
		filter.Status = &s
	}

	switch authDomain.RoleType(actorRole) {
	case authDomain.RoleSystemAdmin, authDomain.RoleMethodist, authDomain.RoleAcademicSecretary:
		// Pass-through — these roles see every WP.
	case authDomain.RoleTeacher:
		actor := actorID
		filter.AuthorID = &actor
	case authDomain.RoleStudent:
		approved := domain.StatusApproved
		filter.Status = &approved
	default:
		emitAudit(uc.audit, ctx, "work_program.list_denied",
			denialFields(actorID, 0, "forbidden_role", ""))
		return ListWorkProgramsResult{}, fmt.Errorf("%w: role %q cannot list work programs",
			domain.ErrWorkProgramScopeForbidden, actorRole)
	}

	filter.Limit = in.Limit
	if filter.Limit <= 0 {
		filter.Limit = defaultListWPLimit
	}
	filter.Limit = min(filter.Limit, maxListWPLimit)
	filter.Offset = max(in.Offset, 0)

	res, err := uc.repo.List(ctx, filter)
	if err != nil {
		return ListWorkProgramsResult{}, err
	}
	return ListWorkProgramsResult{Items: res.Items, Total: res.Total}, nil
}
