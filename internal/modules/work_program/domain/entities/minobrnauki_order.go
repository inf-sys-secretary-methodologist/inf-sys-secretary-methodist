package entities

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

// NewMinobrnaukiOrderInput collects constructor parameters for a
// MinobrnaukiOrder (приказ Минобрнауки) per ADR-11.
type NewMinobrnaukiOrderInput struct {
	OrderNumber string
	Title       string
	PublishedAt time.Time
	DocumentID  *int64 // optional FK to the uploaded PDF in the documents module
	ChangeScope domain.MinobrnaukiOrderChangeScope
	Summary     string // optional human-readable digest
	UploadedBy  int64
}

// MinobrnaukiOrder — приказ Минобрнауки, the external regulatory trigger
// for РПД revisions per ADR-11.
//
// STUB (PR 6a RED): fields + validation land in GREEN.
type MinobrnaukiOrder struct{}

// NewMinobrnaukiOrder constructs a MinobrnaukiOrder, validating invariants.
//
// STUB (PR 6a RED).
func NewMinobrnaukiOrder(in NewMinobrnaukiOrderInput) (*MinobrnaukiOrder, error) {
	_ = in
	return &MinobrnaukiOrder{}, nil
}

// ReconstituteMinobrnaukiOrderInput collects fields for repository hydration.
type ReconstituteMinobrnaukiOrderInput struct {
	ID          int64
	OrderNumber string
	Title       string
	PublishedAt time.Time
	DocumentID  *int64
	ChangeScope domain.MinobrnaukiOrderChangeScope
	Summary     string
	UploadedBy  int64
	CreatedAt   time.Time
}

// ReconstituteMinobrnaukiOrder builds a MinobrnaukiOrder from persisted state.
//
// STUB (PR 6a RED).
func ReconstituteMinobrnaukiOrder(in ReconstituteMinobrnaukiOrderInput) *MinobrnaukiOrder {
	_ = in
	return &MinobrnaukiOrder{}
}

// ID returns the persistent identifier.
func (o *MinobrnaukiOrder) ID() int64 { return 0 }

// OrderNumber returns the official order number.
func (o *MinobrnaukiOrder) OrderNumber() string { return "" }

// Title returns the order title.
func (o *MinobrnaukiOrder) Title() string { return "" }

// PublishedAt returns the publication date.
func (o *MinobrnaukiOrder) PublishedAt() time.Time { return time.Time{} }

// DocumentID returns the optional FK to the uploaded document, or nil.
func (o *MinobrnaukiOrder) DocumentID() *int64 { return nil }

// ChangeScope returns the minor/major scope.
func (o *MinobrnaukiOrder) ChangeScope() domain.MinobrnaukiOrderChangeScope { return "" }

// Summary returns the optional digest.
func (o *MinobrnaukiOrder) Summary() string { return "" }

// UploadedBy returns the methodist/secretary who recorded the order.
func (o *MinobrnaukiOrder) UploadedBy() int64 { return 0 }

// CreatedAt returns the creation timestamp.
func (o *MinobrnaukiOrder) CreatedAt() time.Time { return time.Time{} }
