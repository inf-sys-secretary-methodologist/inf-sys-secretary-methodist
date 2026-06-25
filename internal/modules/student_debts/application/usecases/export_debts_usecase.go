package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// exportDebtsRepo is the narrow read port ExportDebts needs: the
// fully-hydrated registry snapshot for serialization.
type exportDebtsRepo interface {
	ListForExport(ctx context.Context, filter repositories.StudentDebtListFilter) ([]*entities.StudentDebt, error)
}

// ExportDebtsUseCase serializes the debt registry into a downloadable
// document for staff and teachers. Access mirrors ListDebts exactly (the
// shared resolveDebtReadScope gate): staff export everything matching the
// filter; a teacher's export is forced to the disciplines they own; a
// teacher owning no disciplines exports an empty document; students and
// unknown roles are denied.
type ExportDebtsUseCase struct {
	repo         exportDebtsRepo
	teacherScope TeacherScopeResolver
	exporter     DebtExporter
	audit        AuditSink
}

// NewExportDebtsUseCase wires the use case. repo, teacherScope and
// exporter are required; audit may be nil.
func NewExportDebtsUseCase(repo exportDebtsRepo, teacherScope TeacherScopeResolver, exporter DebtExporter, audit AuditSink) *ExportDebtsUseCase {
	if repo == nil || teacherScope == nil || exporter == nil {
		panic("student_debts: NewExportDebtsUseCase requires non-nil repo, teacherScope and exporter")
	}
	return &ExportDebtsUseCase{repo: repo, teacherScope: teacherScope, exporter: exporter, audit: audit}
}

// Execute authorizes the actor, fetches the role-scoped registry snapshot,
// serializes it via the exporter and returns the document bytes. A teacher
// owning no disciplines serializes an empty slice (a valid empty document)
// without hitting the repo. Denials return ErrDebtAccessForbidden with an
// export_denied audit and no repo/exporter traffic. Repo and exporter
// failures wrap through; the exporter is skipped when the repo fails.
func (uc *ExportDebtsUseCase) Execute(ctx context.Context, actorID int64, actorRole string, filter repositories.StudentDebtListFilter) ([]byte, error) {
	scoped, proceed, err := resolveDebtReadScope(ctx, uc.teacherScope, uc.audit, "student_debts.export_denied", actorID, actorRole, filter)
	if err != nil {
		return nil, err
	}

	var debts []*entities.StudentDebt
	if proceed {
		debts, err = uc.repo.ListForExport(ctx, scoped)
		if err != nil {
			return nil, fmt.Errorf("student_debts: export: list: %w", err)
		}
	}

	data, err := uc.exporter.Export(ctx, debts)
	if err != nil {
		return nil, fmt.Errorf("student_debts: export: serialize: %w", err)
	}

	emitAudit(uc.audit, ctx, "student_debts.exported", map[string]any{
		"actor_user_id": actorID,
		"count":         len(debts),
	})
	return data, nil
}
