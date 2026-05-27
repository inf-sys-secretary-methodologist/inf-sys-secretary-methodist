package entities

import (
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

const (
	minReferenceYear = 1900
	maxReferenceYear = 2100
	maxReferenceURL  = 500
)

// NewReferenceInput collects constructor parameters for a Reference (литература/источник).
type NewReferenceInput struct {
	Kind       domain.ReferenceKind
	Citation   string
	Year       *int   // optional: [1900, 2100]
	ISBN       string // optional
	URL        string // optional, ≤ 500 chars
	OrderIndex int
}

// Reference — литература/источник, inner aggregate of WorkProgram per ADR-1.
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

// NewReference constructs a fresh Reference. Kind validated against
// canonical values, citation trim + non-empty, year (when set) in
// [1900, 2100], URL ≤ 500 chars, order_index ≥ 0. ISBN format is
// not validated — references могут содержать non-ISBN sources.
func NewReference(in NewReferenceInput) (*Reference, error) {
	trimmedCitation := strings.TrimSpace(in.Citation)
	trimmedISBN := strings.TrimSpace(in.ISBN)
	trimmedURL := strings.TrimSpace(in.URL)

	if !in.Kind.IsValid() {
		return nil, fmt.Errorf("%w: kind %q must be one of main/additional/electronic",
			domain.ErrInvalidWorkProgram, in.Kind)
	}
	if trimmedCitation == "" {
		return nil, fmt.Errorf("%w: citation is required", domain.ErrInvalidWorkProgram)
	}
	if in.Year != nil {
		if *in.Year < minReferenceYear || *in.Year > maxReferenceYear {
			return nil, fmt.Errorf("%w: year must be in [%d, %d]",
				domain.ErrInvalidWorkProgram, minReferenceYear, maxReferenceYear)
		}
	}
	if len(trimmedURL) > maxReferenceURL {
		return nil, fmt.Errorf("%w: url must be <= %d chars", domain.ErrInvalidWorkProgram, maxReferenceURL)
	}
	if in.OrderIndex < 0 {
		return nil, fmt.Errorf("%w: order_index must be non-negative", domain.ErrInvalidWorkProgram)
	}
	return &Reference{
		kind:       in.Kind,
		citation:   trimmedCitation,
		year:       in.Year,
		isbn:       trimmedISBN,
		url:        trimmedURL,
		orderIndex: in.OrderIndex,
	}, nil
}

// ReconstituteReferenceInput collects fields for repository hydration.
type ReconstituteReferenceInput struct {
	ID            int64
	WorkProgramID int64
	Kind          domain.ReferenceKind
	Citation      string
	Year          *int
	ISBN          string
	URL           string
	OrderIndex    int
}

// ReconstituteReference builds a Reference from persisted state.
// Skips invariant checks — DB CHECK constraints already validated.
func ReconstituteReference(in ReconstituteReferenceInput) *Reference {
	return &Reference{
		id:            in.ID,
		workProgramID: in.WorkProgramID,
		kind:          in.Kind,
		citation:      in.Citation,
		year:          in.Year,
		isbn:          in.ISBN,
		url:           in.URL,
		orderIndex:    in.OrderIndex,
	}
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
