package entities

import (
	"errors"
	"time"
)

// ErrInvalidSection signals a violation of one of the Section
// construction invariants (empty title, title too long, description
// too long, non-positive curriculum_id, negative order_index).
// Handlers map this sentinel to HTTP 422.
var ErrInvalidSection = errors.New("section: invalid section")

// Section is the aggregate root for раздел учебного плана — a top-level
// container within a Curriculum that groups DisciplineItem entities
// (v0.128.1+). Per ADR-1 (Beta aggregate boundary) Section is an
// independent AR; it does NOT carry a Curriculum reference beyond the
// FK column. Lifecycle inheritance (ADR-2): Section has no own status;
// editability is determined by curriculum.status, enforced in the
// use-case layer.
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

// NewSectionParams bundles the constructor inputs.
type NewSectionParams struct {
	CurriculumID int64
	Title        string
	Description  string
	OrderIndex   int
	Now          time.Time
}

// NewSection — implementation lands в GREEN commit (Pair 1).
func NewSection(p NewSectionParams) (*Section, error) {
	_ = p
	return nil, errors.New("section: NewSection not implemented yet")
}

// ReconstituteSection — implementation lands в GREEN commit (Pair 1).
func ReconstituteSection(
	id, curriculumID int64,
	title, description string,
	orderIndex, version int,
	createdAt, updatedAt time.Time,
) *Section {
	_, _, _, _, _, _, _, _ = id, curriculumID, title, description, orderIndex, version, createdAt, updatedAt
	return nil
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
func (s *Section) Version() int { return s.version }

// CreatedAt returns the creation timestamp.
func (s *Section) CreatedAt() time.Time { return s.createdAt }

// UpdatedAt returns the last-mutation timestamp.
func (s *Section) UpdatedAt() time.Time { return s.updatedAt }
