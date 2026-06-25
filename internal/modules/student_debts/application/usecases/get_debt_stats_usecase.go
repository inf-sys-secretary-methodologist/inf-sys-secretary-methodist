package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// statsRepo is the narrow read port GetDebtStats needs: the per-status
// aggregate for a filter.
type statsRepo interface {
	Stats(ctx context.Context, filter repositories.StudentDebtListFilter) (repositories.StudentDebtStats, error)
}

// GetDebtStatsUseCase returns the dashboard aggregate (debt counts per FSM
// state) for staff and teachers. Access mirrors ListDebts exactly (the
// shared resolveDebtReadScope gate): staff aggregate over everything
// matching the filter; a teacher's aggregate is forced to the disciplines
// they own; a teacher owning no disciplines gets a zero aggregate without
// touching the repo; students and unknown roles are denied. Like ListDebts
// (and unlike ExportDebts, which logs a data-egress event), a successful
// read emits no audit — only denials do.
type GetDebtStatsUseCase struct {
	repo         statsRepo
	teacherScope TeacherScopeResolver
	audit        AuditSink
}

// NewGetDebtStatsUseCase wires the use case. repo and teacherScope are
// required; audit may be nil.
func NewGetDebtStatsUseCase(repo statsRepo, teacherScope TeacherScopeResolver, audit AuditSink) *GetDebtStatsUseCase {
	if repo == nil || teacherScope == nil {
		panic("student_debts: NewGetDebtStatsUseCase requires non-nil repo and teacherScope")
	}
	return &GetDebtStatsUseCase{repo: repo, teacherScope: teacherScope, audit: audit}
}

// Execute authorizes the actor, applies the role-scoped filter and returns
// the per-status aggregate. A teacher owning no disciplines returns a zero
// aggregate without hitting the repo. Denials return ErrDebtAccessForbidden
// with a stats_denied audit and no repo traffic. The repo error passes
// through unwrapped (it already carries a student_debts: stats: prefix),
// mirroring ListDebts which likewise returns its repo error directly.
func (uc *GetDebtStatsUseCase) Execute(ctx context.Context, actorID int64, actorRole string, filter repositories.StudentDebtListFilter) (repositories.StudentDebtStats, error) {
	scoped, proceed, err := resolveDebtReadScope(ctx, uc.teacherScope, uc.audit, "student_debts.stats_denied", actorID, actorRole, filter)
	if err != nil {
		return repositories.StudentDebtStats{}, err
	}
	if !proceed {
		return repositories.StudentDebtStats{}, nil
	}
	return uc.repo.Stats(ctx, scoped)
}
