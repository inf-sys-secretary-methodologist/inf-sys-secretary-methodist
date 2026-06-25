package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// listDebtsRepo is the narrow read port ListDebts needs.
type listDebtsRepo interface {
	List(ctx context.Context, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error)
}

// ListDebtsUseCase lists the debt registry for staff and teachers.
// Staff (admin/methodist/secretary) see every debt subject to the
// inbound filter; a teacher's view is forced to the disciplines they own
// (any client-supplied DisciplineIDs are overridden, closing
// cross-discipline enumeration). Students are denied here — they use
// ListMyDebtsUseCase via the /my endpoint.
type ListDebtsUseCase struct {
	repo         listDebtsRepo
	teacherScope TeacherScopeResolver
	audit        AuditSink
}

// NewListDebtsUseCase wires the use case. repo and teacherScope are
// required; audit may be nil.
func NewListDebtsUseCase(repo listDebtsRepo, teacherScope TeacherScopeResolver, audit AuditSink) *ListDebtsUseCase {
	if repo == nil || teacherScope == nil {
		panic("student_debts: NewListDebtsUseCase requires non-nil repo and teacherScope")
	}
	return &ListDebtsUseCase{repo: repo, teacherScope: teacherScope, audit: audit}
}

// Execute applies the role-scoped filter and lists matching debts.
func (uc *ListDebtsUseCase) Execute(ctx context.Context, actorID int64, actorRole string, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
	return repositories.StudentDebtListResult{}, errNotImplemented
}
