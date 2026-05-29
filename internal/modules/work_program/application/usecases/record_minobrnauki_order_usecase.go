package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// RecordMinobrnaukiOrderInput is the public request DTO for recording a
// приказ Минобрнауки (ADR-11). The actor (→ UploadedBy) and actor role
// are supplied as separate Execute arguments, not in this struct, so
// handlers wire the JWT subject + role explicitly rather than through a
// struct deserialised from untrusted JSON.
type RecordMinobrnaukiOrderInput struct {
	OrderNumber            string
	Title                  string
	PublishedAt            time.Time
	DocumentID             *int64
	ChangeScope            string // "minor" / "major"; mapped to the domain enum
	Summary                string
	AffectedWorkProgramIDs []int64
}

// recordMinobrnaukiOrderRepo is the narrow persistence port this use
// case needs — Save only. Defining it here (not the wide
// MinobrnaukiOrderRepository) keeps use-case tests free of GetByID /
// List / FindAffected wiring they do not exercise.
type recordMinobrnaukiOrderRepo interface {
	Save(ctx context.Context, order *entities.MinobrnaukiOrder, affectedWorkProgramIDs []int64) error
}

// RecordMinobrnaukiOrderUseCase records a Минобрнауки order artifact and
// emits the matching audit event. Role gate per ADR-11: only methodist /
// academic_secretary / system_admin may record an order; teacher and
// student are denied.
type RecordMinobrnaukiOrderUseCase struct {
	repo  recordMinobrnaukiOrderRepo
	audit AuditSink
}

// NewRecordMinobrnaukiOrderUseCase wires the use case. repo is required
// (non-nil) so a missing dependency fails at DI wiring, not deep in the
// call stack. Nil audit sink is tolerated (tests may opt out).
func NewRecordMinobrnaukiOrderUseCase(repo recordMinobrnaukiOrderRepo, audit AuditSink) *RecordMinobrnaukiOrderUseCase {
	if repo == nil {
		panic("work_program: NewRecordMinobrnaukiOrderUseCase requires non-nil repo")
	}
	return &RecordMinobrnaukiOrderUseCase{repo: repo, audit: audit}
}

// Execute is a RED stub — real implementation lands in the GREEN commit.
func (uc *RecordMinobrnaukiOrderUseCase) Execute(_ context.Context, _ int64, _ string, _ RecordMinobrnaukiOrderInput) (*entities.MinobrnaukiOrder, error) {
	_ = uc.repo
	_ = uc.audit
	return nil, errors.New("work_program: record minobrnauki order not implemented (RED stub)")
}
