package usecases

import (
	"context"
	"errors"

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

// Execute is a RED stub — real implementation lands in the GREEN commit.
func (uc *GetMinobrnaukiOrderUseCase) Execute(_ context.Context, _ string, _ int64) (*entities.MinobrnaukiOrder, []int64, error) {
	_ = uc.repo
	return nil, nil, errors.New("work_program: get minobrnauki order not implemented (RED stub)")
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

// Execute is a RED stub — real implementation lands in the GREEN commit.
func (uc *ListMinobrnaukiOrdersUseCase) Execute(_ context.Context, _ string, _ repositories.MinobrnaukiOrderListFilter) (repositories.MinobrnaukiOrderListResult, error) {
	_ = uc.repo
	return repositories.MinobrnaukiOrderListResult{}, errors.New("work_program: list minobrnauki orders not implemented (RED stub)")
}
