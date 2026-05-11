// Package auditlog hosts the admin-only read API for the audit_logs
// table. The write path (AuditLogger + AuditLogRepositoryPG) lives in
// internal/shared/infrastructure/logging — this package consumes its
// AuditLogReader port via DIP and exposes a thin use case + HTTP
// handler pair gated by RequireRole(system_admin).
//
// "shared/admin" houses cross-cutting administrative features that do
// not belong to any single bounded context (audit-log, backup/restore,
// global settings). New admin features should add a subpackage here
// rather than introducing top-level admin modules.
package auditlog

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// DefaultLimit is the page size used when a caller omits ?limit= or
// passes a non-positive value.
const DefaultLimit = 50

// MaxLimit caps the per-page size regardless of request input. At ~1
// KiB JSON per row the cap keeps a single response well under any
// reasonable reverse-proxy buffer ceiling.
const MaxLimit = 200

// ListInput is the validated, parsed input to AdminAuditLogUseCase.List.
// Time fields are pointers so the empty/unset case is unambiguous —
// an explicit zero-value time would silently filter all rows older
// than 0001-01-01 (i.e. everything).
type ListInput struct {
	Action   string
	Resource string
	UserID   *int64
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

// ErrInvalidTimeRange signals From >= To — the half-open [from, to)
// reader contract requires strict inequality to return any rows.
var ErrInvalidTimeRange = errors.New("audit-log list: from must be strictly before to")

// AdminAuditLogUseCase is the application-layer collaborator that
// clamps pagination, validates the time range, and delegates to the
// reader port. Intentionally thin — authorization is the route-level
// RequireRole(system_admin) gate and persistence shape passes through
// without transformation.
type AdminAuditLogUseCase struct {
	reader logging.AuditLogReader
}

// NewAdminAuditLogUseCase wires the use case against an
// AuditLogReader. Panics on a nil reader so misconfigured DI fails
// at construction rather than on the first request.
func NewAdminAuditLogUseCase(reader logging.AuditLogReader) *AdminAuditLogUseCase {
	if reader == nil {
		panic("auditlog: nil AuditLogReader")
	}
	return &AdminAuditLogUseCase{reader: reader}
}

// List clamps pagination, validates the time range, and forwards the
// filter to the reader. Negative offsets clamp to zero (rather than
// 400) because they are reachable from URL-template arithmetic on
// the client and a silent floor is friendlier than a hard reject.
func (uc *AdminAuditLogUseCase) List(ctx context.Context, in ListInput) (logging.AuditLogListResult, error) {
	if in.From != nil && in.To != nil && !in.From.Before(*in.To) {
		return logging.AuditLogListResult{}, ErrInvalidTimeRange
	}

	limit := in.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	offset := max(in.Offset, 0)

	return uc.reader.List(ctx, logging.AuditLogFilter{
		Action:   in.Action,
		Resource: in.Resource,
		UserID:   in.UserID,
		From:     in.From,
		To:       in.To,
		Limit:    limit,
		Offset:   offset,
	})
}
