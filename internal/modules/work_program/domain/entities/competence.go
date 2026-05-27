package entities

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

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

// NewCompetence — stub for RED commit.
func NewCompetence(_ string, _ domain.CompetenceType, _ string) (*Competence, error) {
	return nil, domain.ErrInvalidWorkProgram
}

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
