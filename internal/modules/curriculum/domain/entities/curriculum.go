package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInvalidCurriculum signals a violation of one of the Curriculum
// construction invariants (empty title/code/specialty, year out of
// range, description too long, non-positive created_by). Handlers map
// this sentinel to HTTP 422.
var ErrInvalidCurriculum = errors.New("curriculum: invalid curriculum")

// Year-range invariants — chosen wide enough to cover both legacy
// archived curricula (pre-2010 not realistic in this institution but
// kept permissive) and forward-looking programmes. Mirrored exactly by
// chk_curricula_year_range in migration 031.
const (
	minYear           = 2000
	maxYear           = 2100
	maxDescriptionLen = 4096
)

// Curriculum is the aggregate root for a single академический учебный
// план: a programme published by a methodist that the administrator
// will eventually approve. Disciplines (child entities, v0.117.0)
// belong to it. The aggregate validates its own canonical form on
// every write; the SQL CHECKs in migration 031 are defense-in-depth
// for the same invariants.
type Curriculum struct {
	ID          int64
	title       string
	code        string
	specialty   string
	year        int
	description string
	status      CurriculumStatus
	createdBy   int64
	approvedBy  *int64
	approvedAt  *time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

// NewCurriculumParams bundles the constructor inputs so call sites
// stay readable when more optional fields are added (attachments,
// reviewers, ...). Mirrors NewAssignmentParams from assignments.
type NewCurriculumParams struct {
	Title       string
	Code        string
	Specialty   string
	Year        int
	Description string
	CreatedBy   int64
	Now         time.Time
}

// NewCurriculum validates invariants and returns a fresh Curriculum
// in draft status with no approval recorded.
//
// Invariants (mirroring the SQL CHECK constraints in migration 031):
//   - title trimmed-non-empty
//   - code trimmed-non-empty
//   - specialty trimmed-non-empty
//   - year ∈ [2000, 2100]
//   - description (after trim) ≤ 4096 chars
//   - created_by > 0
//
// Each violation wraps ErrInvalidCurriculum with the offending field
// so errors.Is still resolves the sentinel for the 422 mapping in
// handlers.
func NewCurriculum(p NewCurriculumParams) (*Curriculum, error) {
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title must not be empty", ErrInvalidCurriculum)
	}
	code := strings.TrimSpace(p.Code)
	if code == "" {
		return nil, fmt.Errorf("%w: code must not be empty", ErrInvalidCurriculum)
	}
	specialty := strings.TrimSpace(p.Specialty)
	if specialty == "" {
		return nil, fmt.Errorf("%w: specialty must not be empty", ErrInvalidCurriculum)
	}
	if p.Year < minYear || p.Year > maxYear {
		return nil, fmt.Errorf("%w: year %d outside [%d, %d]",
			ErrInvalidCurriculum, p.Year, minYear, maxYear)
	}
	description := strings.TrimSpace(p.Description)
	if len([]rune(description)) > maxDescriptionLen {
		return nil, fmt.Errorf("%w: description length %d exceeds max %d",
			ErrInvalidCurriculum, len([]rune(description)), maxDescriptionLen)
	}
	if p.CreatedBy <= 0 {
		return nil, fmt.Errorf("%w: created_by must be positive, got %d",
			ErrInvalidCurriculum, p.CreatedBy)
	}
	return &Curriculum{
		title:       title,
		code:        code,
		specialty:   specialty,
		year:        p.Year,
		description: description,
		status:      StatusDraft,
		createdBy:   p.CreatedBy,
		createdAt:   p.Now,
		updatedAt:   p.Now,
	}, nil
}

// ReconstituteCurriculum rebuilds a Curriculum from authoritative
// storage. It bypasses NewCurriculum's invariant checks because the
// values are already canonical (the DB enforces the same CHECKs at
// write time). Used exclusively by repository implementations.
//
// approvedBy / approvedAt are wired through as pointers so the
// existing nullable columns in migration 031 round-trip naturally.
func ReconstituteCurriculum(
	id int64, title, code, specialty string, year int, description string,
	status CurriculumStatus, createdBy int64,
	approvedBy *int64, approvedAt *time.Time,
	createdAt, updatedAt time.Time,
) *Curriculum {
	return &Curriculum{
		ID:          id,
		title:       title,
		code:        code,
		specialty:   specialty,
		year:        year,
		description: description,
		status:      status,
		createdBy:   createdBy,
		approvedBy:  approvedBy,
		approvedAt:  approvedAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// Title returns the curriculum's human-readable title.
func (c *Curriculum) Title() string { return c.title }

// Code returns the unique code identifier.
func (c *Curriculum) Code() string { return c.code }

// Specialty returns the academic specialty (направление подготовки).
func (c *Curriculum) Specialty() string { return c.specialty }

// Year returns the year of programme start (e.g. 2026 for the
// 2026/2027 учебный год).
func (c *Curriculum) Year() int { return c.year }

// Description returns the optional free-form description.
func (c *Curriculum) Description() string { return c.description }

// Status returns the current lifecycle state.
func (c *Curriculum) Status() CurriculumStatus { return c.status }

// CreatedBy returns the methodist user id that authored this curriculum.
func (c *Curriculum) CreatedBy() int64 { return c.createdBy }

// ApprovedBy returns the administrator id that approved the
// curriculum, or nil if it has not yet been approved.
func (c *Curriculum) ApprovedBy() *int64 { return c.approvedBy }

// ApprovedAt returns the timestamp at which the curriculum reached
// the approved state, or nil if it has not yet been approved.
func (c *Curriculum) ApprovedAt() *time.Time { return c.approvedAt }

// CreatedAt returns the creation timestamp.
func (c *Curriculum) CreatedAt() time.Time { return c.createdAt }

// UpdatedAt returns the last-mutation timestamp.
func (c *Curriculum) UpdatedAt() time.Time { return c.updatedAt }
