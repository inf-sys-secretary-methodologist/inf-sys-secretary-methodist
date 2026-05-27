package entities

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

const maxCompetenceDescriptionLen = 2048

// Competence — ФГОС-derived компетенция (ПК/ОК/УК), inner aggregate
// of WorkProgram per ADR-1. Identity within parent = code (UNIQUE
// enforced at DB level in migration 048).
type Competence struct {
	id            int64
	workProgramID int64
	code          string
	ctype         domain.CompetenceType
	description   string
	createdAt     time.Time
}

// NewCompetence constructs a fresh Competence. code and description
// are trimmed; ctype is validated against the three FGOS classifications.
// description is rune-count-bounded (UTF-8 aware).
func NewCompetence(code string, ctype domain.CompetenceType, description string) (*Competence, error) {
	trimmedCode := strings.TrimSpace(code)
	trimmedDesc := strings.TrimSpace(description)

	if trimmedCode == "" {
		return nil, fmt.Errorf("%w: code is required", domain.ErrInvalidWorkProgram)
	}
	if !ctype.IsValid() {
		return nil, fmt.Errorf("%w: type %q must be one of pk/ok/uk", domain.ErrInvalidWorkProgram, ctype)
	}
	if trimmedDesc == "" {
		return nil, fmt.Errorf("%w: description is required", domain.ErrInvalidWorkProgram)
	}
	if utf8.RuneCountInString(trimmedDesc) > maxCompetenceDescriptionLen {
		return nil, fmt.Errorf("%w: description must be <= %d runes",
			domain.ErrInvalidWorkProgram, maxCompetenceDescriptionLen)
	}
	return &Competence{
		code:        trimmedCode,
		ctype:       ctype,
		description: trimmedDesc,
		createdAt:   time.Now().UTC(),
	}, nil
}

// ReconstituteCompetenceInput collects fields for repository hydration.
type ReconstituteCompetenceInput struct {
	ID            int64
	WorkProgramID int64
	Code          string
	Type          domain.CompetenceType
	Description   string
	CreatedAt     time.Time
}

// ReconstituteCompetence builds a Competence from persisted state.
// Skips invariant checks. RED stub returns nil.
func ReconstituteCompetence(_ ReconstituteCompetenceInput) *Competence { return nil }

// ID returns the persistent identifier.
func (c *Competence) ID() int64 { return c.id }

// WorkProgramID returns the parent aggregate identifier.
func (c *Competence) WorkProgramID() int64 { return c.workProgramID }

// Code returns the FGOS code (e.g. "ПК-3", "УК-7"), trimmed.
func (c *Competence) Code() string { return c.code }

// Type returns the typed classification (PK/OK/UK).
func (c *Competence) Type() domain.CompetenceType { return c.ctype }

// Description returns the human-readable competence formulation.
func (c *Competence) Description() string { return c.description }

// CreatedAt returns the creation timestamp.
func (c *Competence) CreatedAt() time.Time { return c.createdAt }
