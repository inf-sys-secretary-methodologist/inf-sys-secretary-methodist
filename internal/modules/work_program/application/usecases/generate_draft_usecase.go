package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// GenerateDraftUseCase fills an empty draft РПД with LLM-generated
// content (goals / competences / topics / references). STUB — the
// behavior lands in the GREEN commit; this compiles so the RED test
// suite can run and fail.
type GenerateDraftUseCase struct{}

// NewGenerateDraftUseCase is the STUB constructor (no wiring yet).
func NewGenerateDraftUseCase(
	repo generateDraftRepo,
	generator DraftGenerator,
	disciplines DisciplineInfoProvider,
	limiter GenerationRateLimiter,
	audit AuditSink,
) *GenerateDraftUseCase {
	return &GenerateDraftUseCase{}
}

// Execute is the STUB entrypoint — always returns not-implemented so
// every behavioral assertion in the test suite is RED.
func (uc *GenerateDraftUseCase) Execute(
	ctx context.Context,
	actorID int64,
	actorRole string,
	workProgramID int64,
) (*entities.WorkProgram, error) {
	return nil, errors.New("work_program: GenerateDraft not implemented")
}
