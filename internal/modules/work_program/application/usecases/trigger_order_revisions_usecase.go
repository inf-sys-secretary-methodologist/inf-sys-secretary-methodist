package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// revisionTargetRepo is the narrow WorkProgram persistence port the
// trigger use case needs: load an affected program and write its status
// change back. A subset of the wide WorkProgramRepository so trigger
// tests stay free of Save / List / Delete wiring they do not exercise.
type revisionTargetRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// RevisionDelegation is the structured payload handed to the
// RevisionTaskDelegator — IDs only, no UI strings, so the adapter at the
// DI seam owns the task title/description wording (Clean Architecture:
// no user-facing text in the use case).
type RevisionDelegation struct {
	CreatorID          int64 // actor who recorded the order (task creator)
	TeacherID          int64 // РПД author → task assignee
	WorkProgramID      int64
	MinobrnaukiOrderID int64
	OrderNumber        string
}

// RevisionTaskDelegator delegates an РПД-revision task to the
// discipline's teacher. Implemented by an adapter over the tasks module
// at the DI seam, so cross-module wiring stays out of this package
// (DIP — the consumer owns the port).
type RevisionTaskDelegator interface {
	DelegateRevision(ctx context.Context, d RevisionDelegation) error
}

// TriggerOrderRevisionsResult summarizes a trigger run for the caller's
// forensic / logging needs.
type TriggerOrderRevisionsResult struct {
	Marked    int // approved → needs_revision and persisted
	Skipped   int // not approved (draft / pending / already needs_revision)
	Delegated int // teacher revision tasks created
	Failures  int // per-program load / update / delegate errors
}

// TriggerOrderRevisionsUseCase drives every РПД affected by a recorded
// приказ Минобрнауки into needs_revision and delegates a revision task to
// each program's author (the discipline's teacher) per ADR-11. It is the
// real trigger for WorkProgram.MarkNeedsRevision (dormant since ADR-8).
type TriggerOrderRevisionsUseCase struct {
	repo      revisionTargetRepo
	delegator RevisionTaskDelegator
	audit     AuditSink
}

// NewTriggerOrderRevisionsUseCase wires the use case. repo and delegator
// are required (non-nil) so a missing dependency fails at DI wiring, not
// deep in the call stack. Nil audit sink is tolerated.
func NewTriggerOrderRevisionsUseCase(repo revisionTargetRepo, delegator RevisionTaskDelegator, audit AuditSink) *TriggerOrderRevisionsUseCase {
	if repo == nil || delegator == nil {
		panic("work_program: NewTriggerOrderRevisionsUseCase requires non-nil repo and delegator")
	}
	return &TriggerOrderRevisionsUseCase{repo: repo, delegator: delegator, audit: audit}
}

// Execute drives each affected РПД into needs_revision and delegates a
// teacher task. STUB — real behavior lands in the GREEN commit.
func (uc *TriggerOrderRevisionsUseCase) Execute(ctx context.Context, actorID, orderID int64, orderNumber string, affectedWorkProgramIDs []int64) (TriggerOrderRevisionsResult, error) {
	_ = uc.repo
	_ = uc.delegator
	_ = uc.audit
	return TriggerOrderRevisionsResult{}, nil
}
