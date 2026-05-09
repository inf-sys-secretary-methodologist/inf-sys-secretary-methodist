package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInvalidSection signals a violation of one of the Section
// construction invariants (empty/oversize title, oversize description,
// non-positive curriculum_id, negative order_index). Handlers map this
// sentinel to HTTP 422.
var ErrInvalidSection = errors.New("section: invalid section")

// ErrSectionScopeForbidden indicates that a user is not authorized to
// operate on a particular Section — typically because the user is a
// methodist who did not author the parent Curriculum. Admins
// (system_admin / academic_secretary) override this check via the
// isAdmin flag in AuthorizeEdit. Handlers map this sentinel to HTTP 403.
var ErrSectionScopeForbidden = errors.New("section: caller cannot operate on this section")

// ErrCannotEditSection indicates that the parent Curriculum is in a
// status that does not permit content edits to its sections (anything
// other than draft per ADR-2 lifecycle inheritance). Status transitions
// are Curriculum-aggregate concerns; sections inherit the gate.
// Handlers map this sentinel to HTTP 422 (the request is well-formed
// but conflicts with the curriculum's lifecycle).
var ErrCannotEditSection = errors.New("section: cannot edit, curriculum is not in editable status")

// Section is the aggregate root for раздел учебного плана — a top-level
// container within a Curriculum that groups DisciplineItem entities
// (v0.128.1+). Per ADR-1 (Beta aggregate boundary, plan
// 2026-05-09-v0128-section-aggregate.md) Section is an independent AR;
// it carries only the FK to its parent Curriculum, never a navigable
// reference. Lifecycle inheritance (ADR-2): Section has no own status;
// editability is determined by curriculum.status, enforced at the
// use-case layer via AuthorizeEdit (Pair 2).
//
// All write operations bump version (optimistic locking, ADR-3) so
// concurrent bulk-edit (B1b, v0.128.2) detects races without a
// pessimistic lock. The aggregate validates its own canonical form on
// every write; SQL CHECKs in migration 034 are defense-in-depth for
// the same invariants.
type Section struct {
	ID           int64
	curriculumID int64
	title        string
	description  string
	orderIndex   int
	version      int
	createdAt    time.Time
	updatedAt    time.Time
}

// NewSectionParams bundles the constructor inputs so call sites stay
// readable as more optional fields are added (CLAUDE.md ubiquitous
// language gate; mirrors NewCurriculumParams).
type NewSectionParams struct {
	CurriculumID int64
	Title        string
	Description  string
	OrderIndex   int
	Now          time.Time
}

// Section text-field bounds — chosen to fit comfortably within
// PostgreSQL VARCHAR(255) for title and within the 4096-char
// description column shared с Curriculum.description (migration 031).
// Mirrored exactly by the chk_curriculum_sections_* CHECKs в
// migration 034.
const (
	maxSectionTitleLen       = 255
	maxSectionDescriptionLen = 4096
)

// NewSection validates invariants and returns a fresh Section at
// version 0 (optimistic-locking baseline).
//
// Invariants (mirroring the SQL CHECK constraints in migration 034):
//   - curriculum_id > 0
//   - title trimmed-non-empty, ≤ 255 chars (rune count)
//   - description ≤ 4096 chars (rune count, after trim) — blank OK
//   - order_index ≥ 0
//
// Each violation wraps ErrInvalidSection with the offending field so
// errors.Is still resolves the sentinel for the 422 mapping in
// handlers, and the message identifies which field for the operator.
func NewSection(p NewSectionParams) (*Section, error) {
	if p.CurriculumID <= 0 {
		return nil, fmt.Errorf("%w: curriculum_id must be positive, got %d",
			ErrInvalidSection, p.CurriculumID)
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title must not be empty", ErrInvalidSection)
	}
	if len([]rune(title)) > maxSectionTitleLen {
		return nil, fmt.Errorf("%w: title length %d exceeds max %d",
			ErrInvalidSection, len([]rune(title)), maxSectionTitleLen)
	}
	description := strings.TrimSpace(p.Description)
	if len([]rune(description)) > maxSectionDescriptionLen {
		return nil, fmt.Errorf("%w: description length %d exceeds max %d",
			ErrInvalidSection, len([]rune(description)), maxSectionDescriptionLen)
	}
	if p.OrderIndex < 0 {
		return nil, fmt.Errorf("%w: order_index must be non-negative, got %d",
			ErrInvalidSection, p.OrderIndex)
	}
	return &Section{
		curriculumID: p.CurriculumID,
		title:        title,
		description:  description,
		orderIndex:   p.OrderIndex,
		version:      0,
		createdAt:    p.Now,
		updatedAt:    p.Now,
	}, nil
}

// ReconstituteSection rebuilds a Section from authoritative storage.
// It bypasses NewSection's invariant checks because the values are
// already canonical (the DB enforces the same CHECKs at write time).
// Used exclusively by repository implementations.
func ReconstituteSection(
	id, curriculumID int64,
	title, description string,
	orderIndex, version int,
	createdAt, updatedAt time.Time,
) *Section {
	return &Section{
		ID:           id,
		curriculumID: curriculumID,
		title:        title,
		description:  description,
		orderIndex:   orderIndex,
		version:      version,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// CurriculumID returns the FK to curricula.id.
func (s *Section) CurriculumID() int64 { return s.curriculumID }

// Title returns the section's human-readable title.
func (s *Section) Title() string { return s.title }

// Description returns the optional free-form description.
func (s *Section) Description() string { return s.description }

// OrderIndex returns the display ordering hint (≥ 0).
func (s *Section) OrderIndex() int { return s.orderIndex }

// Version returns the optimistic-locking version (≥ 0).
// Repository Update bumps this on each successful write.
func (s *Section) Version() int { return s.version }

// CreatedAt returns the creation timestamp.
func (s *Section) CreatedAt() time.Time { return s.createdAt }

// UpdatedAt returns the last-mutation timestamp.
func (s *Section) UpdatedAt() time.Time { return s.updatedAt }

// UpdateBasics — implementation lands в GREEN commit (Pair 2).
func (s *Section) UpdateBasics(title, description string, orderIndex int, now time.Time) error {
	_, _, _, _ = title, description, orderIndex, now
	return errors.New("section: UpdateBasics not implemented yet")
}

// AuthorizeEdit — implementation lands в GREEN commit (Pair 2).
func (s *Section) AuthorizeEdit(actorID int64, isAdmin bool, curStatus CurriculumStatus, curCreatedBy int64) error {
	_, _, _, _ = actorID, isAdmin, curStatus, curCreatedBy
	return errors.New("section: AuthorizeEdit not implemented yet")
}
