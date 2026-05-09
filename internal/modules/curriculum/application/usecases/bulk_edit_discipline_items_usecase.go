package usecases

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// ===== Public DTOs =====

// BulkEditDisciplineItemsInput is the public DTO для bulk-edit endpoint
// (v0.128.3 ADR-11 — single-section per request, combined operations).
type BulkEditDisciplineItemsInput struct {
	SectionID int64
	Creates   []BulkCreateItem
	Updates   []BulkUpdateItem
	Deletes   []int64
}

// BulkCreateItem carries fields needed для one fresh DisciplineItem.
// SectionID inherited from BulkEditDisciplineItemsInput.SectionID — не
// in the per-item DTO (single-section invariant).
type BulkCreateItem struct {
	Title         string
	HoursLectures int
	HoursPractice int
	HoursLab      int
	HoursSelf     int
	ControlForm   entities.ControlForm
	Credits       int
	Semester      int
	OrderIndex    int
}

// BulkUpdateItem identifies a target item by ID + carries new field
// values. Version intentionally NOT в DTO — repo loads fresh entity
// which carries server-side version; optimistic-lock SQL guards the
// race.
type BulkUpdateItem struct {
	ID            int64
	Title         string
	HoursLectures int
	HoursPractice int
	HoursLab      int
	HoursSelf     int
	ControlForm   entities.ControlForm
	Credits       int
	Semester      int
	OrderIndex    int
}

// BulkEditConflict describes an optimistic-lock race на a single
// update target. Bulk usecase collects ALL conflicts (not short-
// circuit) — UI shows entire stale set in one render для merge.
type BulkEditConflict struct {
	ID              int64
	ExpectedVersion int
	CurrentVersion  int
}

// BulkEditDisciplineItemsResult bundles per-operation outcomes.
// Conflicts non-empty means the whole tx was rolled back и nothing
// applied — caller surfaces 409.
type BulkEditDisciplineItemsResult struct {
	Created   []*entities.DisciplineItem
	Updated   []*entities.DisciplineItem
	Deleted   []int64
	Conflicts []BulkEditConflict
}

// ===== Sentinel errors =====

// ErrEmptyBulkInput signals zero-operation bulk request (creates +
// updates + deletes all empty). 422.
var ErrEmptyBulkInput = errors.New("bulk_edit: empty input — must contain creates, updates, or deletes")

// ErrCrossSectionBulkEdit signals that an Update or Delete target's
// section_id does not match the request path's :sectionID. Single-
// section invariant (ADR-11). 422.
var ErrCrossSectionBulkEdit = errors.New("bulk_edit: target item belongs to different section")

// ErrBulkSectionDeleted signals FK CASCADE race — section disappeared
// между tx open and operation. 409 SECTION_DELETED.
var ErrBulkSectionDeleted = errors.New("bulk_edit: section was deleted concurrently")

// ===== Narrow ports =====

// bulkEditUnitOfWork is the narrow port на UoW Begin (full UoW interface
// excludes anything else).
type bulkEditUnitOfWork interface {
	Begin(ctx context.Context, opts *sql.TxOptions) (repositories.BulkDisciplineItemsTx, error)
}

// ===== Use case =====

// BulkEditDisciplineItemsUseCase applies creates+updates+deletes
// atomically within a single transaction (v0.128.3 ADR-10).
type BulkEditDisciplineItemsUseCase struct {
	uow   bulkEditUnitOfWork
	audit AuditSink
	clock func() time.Time
}

// NewBulkEditDisciplineItemsUseCase wires the use case. Panics on nil
// uow — failure-closed DI (audit nil OK; emit helper handles nil sink).
func NewBulkEditDisciplineItemsUseCase(uow bulkEditUnitOfWork, audit AuditSink, clock func() time.Time) *BulkEditDisciplineItemsUseCase {
	if uow == nil {
		panic("bulk_edit: NewBulkEditDisciplineItemsUseCase requires non-nil uow")
	}
	if clock == nil {
		clock = time.Now
	}
	return &BulkEditDisciplineItemsUseCase{uow: uow, audit: audit, clock: clock}
}

// Execute runs bulk-edit. Pair 2 RED stub — returns not-implemented.
// Pair 2 GREEN replaces с full path:
//  1. Validate non-empty input.
//  2. Begin tx (RepeatableRead per ADR-12).
//  3. defer Rollback (idempotent if Commit fires first).
//  4. Load section + curriculum, AuthorizeDisciplineItemEdit.
//  5. Apply creates / updates / deletes within tx.
//  6. Commit; emit success audit.
func (uc *BulkEditDisciplineItemsUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, in BulkEditDisciplineItemsInput) (*BulkEditDisciplineItemsResult, error) {
	_ = ctx
	_ = actorID
	_ = isAdmin
	_ = in
	return nil, errors.New("bulk_edit: not implemented (Pair 2 RED stub)")
}
