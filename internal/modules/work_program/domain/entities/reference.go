package entities

import (
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

// NewReferenceInput collects constructor parameters for a Reference
// (литература/источник).
type NewReferenceInput struct {
	Kind       domain.ReferenceKind
	Citation   string
	Year       *int   // optional: [1900, 2100]
	ISBN       string // optional
	URL        string // optional, ≤ 500 chars
	OrderIndex int
}

// Reference — литература/источник для дисциплины. Inner aggregate of
// WorkProgram per ADR-1.
type Reference struct {
	id            int64
	workProgramID int64
	kind          domain.ReferenceKind
	citation      string
	year          *int
	isbn          string
	url           string
	orderIndex    int
}

// NewReference — stub for RED commit.
func NewReference(_ NewReferenceInput) (*Reference, error) {
	return nil, domain.ErrInvalidWorkProgram
}

// ID returns the persistent identifier.
func (r *Reference) ID() int64 { return r.id }

// WorkProgramID returns the parent aggregate identifier.
func (r *Reference) WorkProgramID() int64 { return r.workProgramID }

// Kind returns the typed kind (main/additional/electronic).
func (r *Reference) Kind() domain.ReferenceKind { return r.kind }

// Citation returns the formatted citation (e.g. ГОСТ Р 7.0.5).
func (r *Reference) Citation() string { return r.citation }

// Year returns the publication year, nil if unknown.
func (r *Reference) Year() *int { return r.year }

// ISBN returns the ISBN, empty string if none.
func (r *Reference) ISBN() string { return r.isbn }

// URL returns the canonical URL for electronic references, empty if none.
func (r *Reference) URL() string { return r.url }

// OrderIndex returns the display ordering hint (≥ 0).
func (r *Reference) OrderIndex() int { return r.orderIndex }
