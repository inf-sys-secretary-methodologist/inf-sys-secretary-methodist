package entities

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

const (
	maxOrderNumberLen  = 100
	maxOrderTitleLen   = 1024
	maxOrderSummaryLen = 4096
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
// for РПД revisions per ADR-11. A methodist (or secretary with the right)
// records the order; it then drives affected work programs into
// needs_revision. The order itself is an immutable artifact once created
// (no in-place edits) — corrections are made by recording a new order.
type MinobrnaukiOrder struct {
	id          int64
	orderNumber string
	title       string
	publishedAt time.Time
	documentID  *int64
	changeScope domain.MinobrnaukiOrderChangeScope
	summary     string
	uploadedBy  int64
	createdAt   time.Time
}

// NewMinobrnaukiOrder constructs a MinobrnaukiOrder, validating invariants.
// All violations surface as ErrInvalidMinobrnaukiOrder with the offending
// field named. Summary is optional; DocumentID is optional but, when set,
// must be positive.
func NewMinobrnaukiOrder(in NewMinobrnaukiOrderInput) (*MinobrnaukiOrder, error) {
	orderNumber := strings.TrimSpace(in.OrderNumber)
	title := strings.TrimSpace(in.Title)
	summary := strings.TrimSpace(in.Summary)

	if orderNumber == "" {
		return nil, fmt.Errorf("%w: order_number is required", domain.ErrInvalidMinobrnaukiOrder)
	}
	if utf8.RuneCountInString(orderNumber) > maxOrderNumberLen {
		return nil, fmt.Errorf("%w: order_number must be <= %d runes", domain.ErrInvalidMinobrnaukiOrder, maxOrderNumberLen)
	}
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrInvalidMinobrnaukiOrder)
	}
	if utf8.RuneCountInString(title) > maxOrderTitleLen {
		return nil, fmt.Errorf("%w: title must be <= %d runes", domain.ErrInvalidMinobrnaukiOrder, maxOrderTitleLen)
	}
	if in.PublishedAt.IsZero() {
		return nil, fmt.Errorf("%w: published_at is required", domain.ErrInvalidMinobrnaukiOrder)
	}
	if !in.ChangeScope.IsValid() {
		return nil, fmt.Errorf("%w: change_scope %q must be minor or major", domain.ErrInvalidMinobrnaukiOrder, in.ChangeScope)
	}
	if utf8.RuneCountInString(summary) > maxOrderSummaryLen {
		return nil, fmt.Errorf("%w: summary must be <= %d runes", domain.ErrInvalidMinobrnaukiOrder, maxOrderSummaryLen)
	}
	if in.UploadedBy <= 0 {
		return nil, fmt.Errorf("%w: uploaded_by must be positive", domain.ErrInvalidMinobrnaukiOrder)
	}
	if in.DocumentID != nil && *in.DocumentID <= 0 {
		return nil, fmt.Errorf("%w: document_id must be positive when set", domain.ErrInvalidMinobrnaukiOrder)
	}

	return &MinobrnaukiOrder{
		orderNumber: orderNumber,
		title:       title,
		publishedAt: in.PublishedAt,
		documentID:  in.DocumentID,
		changeScope: in.ChangeScope,
		summary:     summary,
		uploadedBy:  in.UploadedBy,
		createdAt:   time.Now().UTC(),
	}, nil
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

// ReconstituteMinobrnaukiOrder builds a MinobrnaukiOrder from persisted
// state. Skips invariant checks — DB CHECK constraints and the original
// NewMinobrnaukiOrder call already validated.
func ReconstituteMinobrnaukiOrder(in ReconstituteMinobrnaukiOrderInput) *MinobrnaukiOrder {
	return &MinobrnaukiOrder{
		id:          in.ID,
		orderNumber: in.OrderNumber,
		title:       in.Title,
		publishedAt: in.PublishedAt,
		documentID:  in.DocumentID,
		changeScope: in.ChangeScope,
		summary:     in.Summary,
		uploadedBy:  in.UploadedBy,
		createdAt:   in.CreatedAt,
	}
}

// ID returns the persistent identifier.
func (o *MinobrnaukiOrder) ID() int64 { return o.id }

// SetID assigns the persistent identifier after a successful repository
// insert. Repository-only contract — domain callers construct via
// NewMinobrnaukiOrder (id stays 0 until the row is written).
func (o *MinobrnaukiOrder) SetID(id int64) { o.id = id }

// OrderNumber returns the official order number (trimmed, ≤ 100 runes).
func (o *MinobrnaukiOrder) OrderNumber() string { return o.orderNumber }

// Title returns the order title (trimmed, ≤ 1024 runes).
func (o *MinobrnaukiOrder) Title() string { return o.title }

// PublishedAt returns the publication date.
func (o *MinobrnaukiOrder) PublishedAt() time.Time { return o.publishedAt }

// DocumentID returns the optional FK to the uploaded document, or nil.
func (o *MinobrnaukiOrder) DocumentID() *int64 { return o.documentID }

// ChangeScope returns the minor/major scope.
func (o *MinobrnaukiOrder) ChangeScope() domain.MinobrnaukiOrderChangeScope { return o.changeScope }

// Summary returns the optional digest (trimmed, ≤ 4096 runes).
func (o *MinobrnaukiOrder) Summary() string { return o.summary }

// UploadedBy returns the methodist/secretary who recorded the order.
func (o *MinobrnaukiOrder) UploadedBy() int64 { return o.uploadedBy }

// CreatedAt returns the creation timestamp.
func (o *MinobrnaukiOrder) CreatedAt() time.Time { return o.createdAt }
