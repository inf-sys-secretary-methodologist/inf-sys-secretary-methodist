package persistence

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// Compile-time assertion that the PG impl satisfies the wide port
// declared in application/usecases (DIP). Catches signature drift at the
// impl's compile site rather than only at DI wiring.
var _ usecases.MinobrnaukiOrderRepository = (*MinobrnaukiOrderRepositoryPG)(nil)

// MinobrnaukiOrderRepositoryPG is the SQL implementation of
// MinobrnaukiOrderRepository (приказы Минобрнауки per ADR-11). Accepts
// DBTX (not *sql.DB) so the same struct works in single-connection mode
// and against `*sql.Tx`.
type MinobrnaukiOrderRepositoryPG struct {
	db DBTX
}

// NewMinobrnaukiOrderRepositoryPG constructs the repository. db can be
// `*sql.DB` (default DI) or `*sql.Tx` (future transactional paths).
func NewMinobrnaukiOrderRepositoryPG(db DBTX) *MinobrnaukiOrderRepositoryPG {
	return &MinobrnaukiOrderRepositoryPG{db: db}
}

// errMinobrnaukiOrderNotImplemented marks the RED-stub state — the real
// SQL implementation lands in the GREEN commit of PR 6a.
var errMinobrnaukiOrderNotImplemented = errors.New("minobrnauki_order: repository not implemented (RED stub)")

// Save is a RED stub — see GREEN commit.
func (r *MinobrnaukiOrderRepositoryPG) Save(_ context.Context, _ *entities.MinobrnaukiOrder, _ []int64) error {
	_ = r.db
	return errMinobrnaukiOrderNotImplemented
}

// GetByID is a RED stub — see GREEN commit.
func (r *MinobrnaukiOrderRepositoryPG) GetByID(_ context.Context, _ int64) (*entities.MinobrnaukiOrder, error) {
	_ = r.db
	return nil, errMinobrnaukiOrderNotImplemented
}

// List is a RED stub — see GREEN commit.
func (r *MinobrnaukiOrderRepositoryPG) List(_ context.Context, _ repositories.MinobrnaukiOrderListFilter) (repositories.MinobrnaukiOrderListResult, error) {
	_ = r.db
	return repositories.MinobrnaukiOrderListResult{}, errMinobrnaukiOrderNotImplemented
}

// FindAffected is a RED stub — see GREEN commit.
func (r *MinobrnaukiOrderRepositoryPG) FindAffected(_ context.Context, _ int64) ([]int64, error) {
	_ = r.db
	return nil, errMinobrnaukiOrderNotImplemented
}
