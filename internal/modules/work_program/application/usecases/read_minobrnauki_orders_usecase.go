package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// getMinobrnaukiOrderRepo is the narrow port the single-order read use
// case needs: the order row plus its affected-work-program set.
type getMinobrnaukiOrderRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.MinobrnaukiOrder, error)
	FindAffected(ctx context.Context, orderID int64) ([]int64, error)
}

// GetMinobrnaukiOrderUseCase reads one приказ Минобрнауки together with
// the ids of the work programs it affects. Read gate per ADR-11: every
// non-student staff role may view orders; students are denied.
type GetMinobrnaukiOrderUseCase struct {
	repo getMinobrnaukiOrderRepo
}

// NewGetMinobrnaukiOrderUseCase wires the use case (repo required).
func NewGetMinobrnaukiOrderUseCase(repo getMinobrnaukiOrderRepo) *GetMinobrnaukiOrderUseCase {
	if repo == nil {
		panic("work_program: NewGetMinobrnaukiOrderUseCase requires non-nil repo")
	}
	return &GetMinobrnaukiOrderUseCase{repo: repo}
}

// Execute applies the read gate, loads the order, then loads its
// affected-work-program set. Returns ErrMinobrnaukiOrderScopeForbidden
// for students, or repositories.ErrMinobrnaukiOrderNotFound (propagated
// from the repo) when the id does not exist. FindAffected runs only
// after GetByID succeeds, so a missing order short-circuits cleanly.
func (uc *GetMinobrnaukiOrderUseCase) Execute(ctx context.Context, actorRole string, id int64) (*entities.MinobrnaukiOrder, []int64, error) {
	if !isAllowedToViewMinobrnaukiOrders(actorRole) {
		return nil, nil, fmt.Errorf("%w: role %q cannot view minobrnauki orders", domain.ErrMinobrnaukiOrderScopeForbidden, actorRole)
	}

	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	affected, err := uc.repo.FindAffected(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return order, affected, nil
}

// listMinobrnaukiOrdersRepo is the narrow port the list read use case needs.
type listMinobrnaukiOrdersRepo interface {
	List(ctx context.Context, filter repositories.MinobrnaukiOrderListFilter) (repositories.MinobrnaukiOrderListResult, error)
}

// ListMinobrnaukiOrdersUseCase returns a page of orders. Orders are not
// author-scoped (unlike WorkPrograms), so there is no row-level filter
// rewrite — the read gate is a flat non-student check.
type ListMinobrnaukiOrdersUseCase struct {
	repo listMinobrnaukiOrdersRepo
}

// NewListMinobrnaukiOrdersUseCase wires the use case (repo required).
func NewListMinobrnaukiOrdersUseCase(repo listMinobrnaukiOrdersRepo) *ListMinobrnaukiOrdersUseCase {
	if repo == nil {
		panic("work_program: NewListMinobrnaukiOrdersUseCase requires non-nil repo")
	}
	return &ListMinobrnaukiOrdersUseCase{repo: repo}
}

// Execute applies the read gate then delegates to the repo. Orders are
// not author-scoped, so the filter passes through unchanged (no role
// rewrite, unlike WorkProgram List). Returns
// ErrMinobrnaukiOrderScopeForbidden for students.
func (uc *ListMinobrnaukiOrdersUseCase) Execute(ctx context.Context, actorRole string, filter repositories.MinobrnaukiOrderListFilter) (repositories.MinobrnaukiOrderListResult, error) {
	if !isAllowedToViewMinobrnaukiOrders(actorRole) {
		return repositories.MinobrnaukiOrderListResult{}, fmt.Errorf("%w: role %q cannot list minobrnauki orders", domain.ErrMinobrnaukiOrderScopeForbidden, actorRole)
	}
	return uc.repo.List(ctx, filter)
}
