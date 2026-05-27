package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// Pagination policy for the List use case.
const (
	defaultListWPLimit = 50
	maxListWPLimit     = 200
)

// ListWorkProgramsInput mirrors WorkProgramListFilter at the use-case
// boundary so handlers don't import persistence DTOs directly.
type ListWorkProgramsInput struct {
	Status             *string // wire-form: nil means no filter
	DisciplineID       *int64
	SpecialtyCode      string
	ApplicableFromYear *int
	AuthorID           *int64
	Limit              int
	Offset             int
}

// ListWorkProgramsResult is the public response shape.
type ListWorkProgramsResult struct {
	Items []repositories.ListItem
	Total int
}

// listWorkProgramsRepo is the narrow port: only List.
type listWorkProgramsRepo interface {
	List(ctx context.Context, filter repositories.WorkProgramListFilter) (repositories.WorkProgramListResult, error)
}

// ListWorkProgramsUseCase loads a role-scoped page of WorkPrograms.
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

// Execute is a stub. Real implementation lands in the matching GREEN commit.
func (uc *ListWorkProgramsUseCase) Execute(_ context.Context, _ int64, _ string, _ ListWorkProgramsInput) (ListWorkProgramsResult, error) {
	return ListWorkProgramsResult{}, errors.New("work_program: ListWorkProgramsUseCase not implemented yet (RED)")
}
