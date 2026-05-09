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

// bulkEditDenialFields composes canonical {actor_user_id, section_id,
// curriculum_id, reason} field shape для bulk_edit_denied audit events.
// Mirror к disciplineItemDenialFields but без per-item axis (bulk-edit
// audit is single event per operation per ADR-13).
func bulkEditDenialFields(actorID, sectionID, curriculumID int64, reason string) map[string]any {
	return map[string]any{
		"actor_user_id": actorID,
		"section_id":    sectionID,
		"curriculum_id": curriculumID,
		"reason":        reason,
	}
}

// Execute runs bulk-edit:
//
//  1. Validate non-empty input → ErrEmptyBulkInput pre-tx (audit denied).
//  2. Begin tx (Repeatable Read per ADR-12) + defer Rollback (idempotent).
//  3. Load section через tx.Sections().GetByID; ErrSectionNotFound → audit
//     denied.
//  4. Load curriculum через tx.Curricula().GetByID — errors propagate без
//     audit (orphaned-section operational anomaly, не policy event).
//  5. AuthorizeDisciplineItemEdit cross-aggregate primitives → audit
//     denied with reason 'forbidden' / 'not_editable'.
//  6. Apply creates: NewDisciplineItem invariant gate → audit denied
//     'invalid'; tx.Items().Save() persists.
//  7. Updates + Deletes (Pair 3 + Pair 4) — placeholder.
//  8. Commit; emit success audit с per-op counts.
//
// Pair 2 GREEN: creates only. Updates / Deletes branches will land
// в Pair 3 / Pair 4 and merge into this Execute body.
func (uc *BulkEditDisciplineItemsUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, in BulkEditDisciplineItemsInput) (*BulkEditDisciplineItemsResult, error) {
	if len(in.Creates) == 0 && len(in.Updates) == 0 && len(in.Deletes) == 0 {
		emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.bulk_edit_denied",
			bulkEditDenialFields(actorID, in.SectionID, 0, "empty_input"))
		return nil, ErrEmptyBulkInput
	}

	tx, err := uc.uow.Begin(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	section, err := tx.Sections().GetByID(ctx, in.SectionID)
	if err != nil {
		if errors.Is(err, repositories.ErrSectionNotFound) {
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.bulk_edit_denied",
				bulkEditDenialFields(actorID, in.SectionID, 0, "section_not_found"))
		}
		return nil, err
	}

	cur, err := tx.Curricula().GetByID(ctx, section.CurriculumID())
	if err != nil {
		return nil, err
	}

	if err := entities.AuthorizeDisciplineItemEdit(actorID, isAdmin, cur.Status(), cur.CreatedBy()); err != nil {
		switch {
		case errors.Is(err, entities.ErrCannotEditDisciplineItem):
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.bulk_edit_denied",
				bulkEditDenialFields(actorID, in.SectionID, section.CurriculumID(), "not_editable"))
		case errors.Is(err, entities.ErrDisciplineItemScopeForbidden):
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.bulk_edit_denied",
				bulkEditDenialFields(actorID, in.SectionID, section.CurriculumID(), "forbidden"))
		}
		return nil, err
	}

	result := &BulkEditDisciplineItemsResult{}
	for _, c := range in.Creates {
		d, err := entities.NewDisciplineItem(entities.NewDisciplineItemParams{
			SectionID:     in.SectionID,
			Title:         c.Title,
			HoursLectures: c.HoursLectures,
			HoursPractice: c.HoursPractice,
			HoursLab:      c.HoursLab,
			HoursSelf:     c.HoursSelf,
			ControlForm:   c.ControlForm,
			Credits:       c.Credits,
			Semester:      c.Semester,
			OrderIndex:    c.OrderIndex,
			Now:           uc.clock(),
		})
		if err != nil {
			if errors.Is(err, entities.ErrInvalidDisciplineItem) {
				emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.bulk_edit_denied",
					bulkEditDenialFields(actorID, in.SectionID, section.CurriculumID(), "invalid"))
			}
			return nil, err
		}
		if err := tx.Items().Save(ctx, d); err != nil {
			return nil, err
		}
		result.Created = append(result.Created, d)
	}

	// Pair 3 (updates) + Pair 4 (deletes) extend here.

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.bulk_edited", map[string]any{
		"actor_user_id": actorID,
		"section_id":    in.SectionID,
		"curriculum_id": section.CurriculumID(),
		"created_count": len(result.Created),
		"updated_count": len(result.Updated),
		"deleted_count": len(result.Deleted),
	})
	return result, nil
}
