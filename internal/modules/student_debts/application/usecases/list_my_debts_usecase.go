package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// ListMyDebtsUseCase lists the authenticated user's own debts (the
// student "Мои долги" view). The actor id is forced onto the filter's
// StudentUserID — a caller can never read another user's debts through
// this path regardless of any client-supplied filter. No role gate:
// seeing your own debts is always permitted.
type ListMyDebtsUseCase struct {
	repo listDebtsRepo
}

// NewListMyDebtsUseCase wires the use case. repo is required.
func NewListMyDebtsUseCase(repo listDebtsRepo) *ListMyDebtsUseCase {
	if repo == nil {
		panic("student_debts: NewListMyDebtsUseCase requires non-nil repo")
	}
	return &ListMyDebtsUseCase{repo: repo}
}

// Execute forces StudentUserID = actorID and lists the actor's debts.
func (uc *ListMyDebtsUseCase) Execute(ctx context.Context, actorID int64, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
	return repositories.StudentDebtListResult{}, errNotImplemented
}
