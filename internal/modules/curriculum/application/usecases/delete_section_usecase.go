package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// deleteSectionRepo is the narrow port for the delete use case:
// fetch (for authorization scope) + delete.
type deleteSectionRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.Section, error)
	Delete(ctx context.Context, id int64) error
}

// deleteSectionCurriculumLookup is the cross-aggregate read port for
// authorization (curriculum status + author primitives).
type deleteSectionCurriculumLookup interface {
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)
}

// DeleteSectionUseCase removes a section after authorization. CASCADE
// in migration 034 handles eventual child-item cleanup (v0.128.1+).
// Hard-delete per ADR-4: undo via UI confirm dialog; audit trail
// captures forensics.
type DeleteSectionUseCase struct {
	repo           deleteSectionRepo
	curriculumRepo deleteSectionCurriculumLookup
	audit          AuditSink
}

// NewDeleteSectionUseCase wires the use case.
func NewDeleteSectionUseCase(
	repo deleteSectionRepo,
	curriculumRepo deleteSectionCurriculumLookup,
	audit AuditSink,
) *DeleteSectionUseCase {
	if repo == nil {
		panic("section: NewDeleteSectionUseCase requires non-nil repo")
	}
	if curriculumRepo == nil {
		panic("section: NewDeleteSectionUseCase requires non-nil curriculumRepo")
	}
	return &DeleteSectionUseCase{repo: repo, curriculumRepo: curriculumRepo, audit: audit}
}

// Execute — implementation lands в GREEN commit (Pair 4).
func (uc *DeleteSectionUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, sectionID int64) error {
	_ = isAdmin
	emitSectionAudit(uc.audit, ctx, "section.unimplemented",
		sectionDenialFields(actorID, sectionID, 0, "stub"))
	return errors.New("section: DeleteSectionUseCase.Execute not implemented yet")
}
