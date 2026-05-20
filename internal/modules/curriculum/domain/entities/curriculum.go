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

// ErrCurriculumScopeForbidden indicates that a user is not authorized
// to operate on a particular Curriculum — typically because the user
// is not the author (per v0.158.0+ the author role is the academic
// secretary). Admins override this check (see AuthorizeEdit). Handlers
// map this sentinel to HTTP 403.
var ErrCurriculumScopeForbidden = errors.New("curriculum: caller cannot operate on this curriculum")

// ErrCannotEditApproved indicates that a Curriculum is not in a state
// that permits content edits. Only draft curricula are editable;
// pending_approval, approved and archived curricula are frozen.
// Status transitions (Approve / Reject / Archive) are separate
// domain methods landing in v0.117.0 — they do NOT go through
// UpdateBasics. Handlers map this sentinel to HTTP 422 (the request
// is well-formed but conflicts with the curriculum's lifecycle).
var ErrCannotEditApproved = errors.New("curriculum: cannot edit non-draft curriculum")

// ErrCannotSubmit signals an attempt to submit a curriculum for
// approval from a status other than draft. Handlers map this
// sentinel to HTTP 422 (NOT_DRAFT).
var ErrCannotSubmit = errors.New("curriculum: cannot submit, status must be draft")

// ErrCannotApprove signals an attempt to approve a curriculum that
// is not currently awaiting approval. Handlers map this sentinel to
// HTTP 422 (NOT_PENDING).
var ErrCannotApprove = errors.New("curriculum: cannot approve, status must be pending_approval")

// ErrCannotReject signals an attempt to reject a curriculum that is
// not currently awaiting approval. Handlers map this sentinel to
// HTTP 422 (NOT_PENDING).
var ErrCannotReject = errors.New("curriculum: cannot reject, status must be pending_approval")

// Year-range invariants — chosen wide enough to cover both legacy
// archived curricula (pre-2010 not realistic in this institution but
// kept permissive) and forward-looking programs. Mirrored exactly by
// chk_curricula_year_range in migration 031.
const (
	minYear           = 2000
	maxYear           = 2100
	maxDescriptionLen = 4096
)

// Curriculum is the aggregate root for a single академический учебный
// план: a program authored by the academic secretary (v0.158.0+) that
// the methodist will eventually approve. Sections + Discipline items
// (child aggregates, v0.117.0+ / v0.128.0+) belong to it. The aggregate
// validates its own canonical form on every write; the SQL CHECKs in
// migration 031 are defense-in-depth for the same invariants.
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
	// v0.157.0 #269 ADR-2 — optimistic-locking version. Starts at 0
	// for fresh aggregates; persistence layer's WHERE id = ? AND
	// version = ? guards lost-update races.
	version int
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
	version int,
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
		version:     version,
	}
}

// Version returns the optimistic-locking version (≥ 0). Persistence
// uses this in the UPDATE WHERE clause; on success the row's stored
// version becomes version+1 и BumpCurriculumVersion lifts the entity's
// in-memory copy to match.
//
// v0.157.0 #269 ADR-2.
func (c *Curriculum) Version() int { return c.version }

// BumpCurriculumVersion is the cross-package helper для repository
// implementations к sync the entity's in-memory version after a
// successful UPDATE. Lives в the entities package (where the field
// is private) и is exported deliberately — Section uses a different
// approach (rebuild via *s = *ReconstituteSection(...) inside the
// persistence-package private helper), but Curriculum has fewer fields
// + no curriculumID-relative invariants к re-validate on rebuild, so
// a focused in-place increment is cleaner. Reviewers wanting к narrow
// the surface should make the field public OR introduce a Setter
// returning a new entity; this helper is the minimum-blast-radius
// equivalent for now. v0.157.0 #269 ADR-2.
func BumpCurriculumVersion(c *Curriculum) {
	c.version++
}

// Title returns the curriculum's human-readable title.
func (c *Curriculum) Title() string { return c.title }

// Code returns the unique code identifier.
func (c *Curriculum) Code() string { return c.code }

// Specialty returns the academic specialty (направление подготовки).
func (c *Curriculum) Specialty() string { return c.specialty }

// Year returns the year of program start (e.g. 2026 for the
// 2026/2027 учебный год).
func (c *Curriculum) Year() int { return c.year }

// Description returns the optional free-form description.
func (c *Curriculum) Description() string { return c.description }

// Status returns the current lifecycle state.
func (c *Curriculum) Status() CurriculumStatus { return c.status }

// CreatedBy returns the user id that authored this curriculum (per
// v0.158.0+ the author is the academic secretary).
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

// UpdateBasics applies a content edit (title / code / specialty /
// year / description) to a draft curriculum. The method is atomic:
// if any invariant fails or the status is not editable, the entity
// is left untouched and the appropriate sentinel is returned.
//
// Status gate fires before invariant validation so callers learn
// "you can't edit this" before "your input is invalid" — the former
// is a workflow problem (resolve via Reject in v0.117.0); the
// latter is fixable by re-submitting cleaner input.
//
// Authorization (who may call this for which curriculum) lives in
// AuthorizeEdit and runs in the use case layer; UpdateBasics
// itself trusts the caller has already passed authorization.
func (c *Curriculum) UpdateBasics(
	title, code, specialty string,
	year int,
	description string,
	now time.Time,
) error {
	if !c.status.CanEdit() {
		return fmt.Errorf("%w: status %q is not editable",
			ErrCannotEditApproved, string(c.status))
	}
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return fmt.Errorf("%w: title must not be empty", ErrInvalidCurriculum)
	}
	trimmedCode := strings.TrimSpace(code)
	if trimmedCode == "" {
		return fmt.Errorf("%w: code must not be empty", ErrInvalidCurriculum)
	}
	trimmedSpecialty := strings.TrimSpace(specialty)
	if trimmedSpecialty == "" {
		return fmt.Errorf("%w: specialty must not be empty", ErrInvalidCurriculum)
	}
	if year < minYear || year > maxYear {
		return fmt.Errorf("%w: year %d outside [%d, %d]",
			ErrInvalidCurriculum, year, minYear, maxYear)
	}
	trimmedDescription := strings.TrimSpace(description)
	if len([]rune(trimmedDescription)) > maxDescriptionLen {
		return fmt.Errorf("%w: description length %d exceeds max %d",
			ErrInvalidCurriculum, len([]rune(trimmedDescription)), maxDescriptionLen)
	}
	// All validation passed — apply mutations atomically.
	c.title = trimmedTitle
	c.code = trimmedCode
	c.specialty = trimmedSpecialty
	c.year = year
	c.description = trimmedDescription
	c.updatedAt = now
	return nil
}

// SubmitForApproval transitions a draft curriculum into the
// pending_approval state. The status check is the only invariant
// the entity enforces; identity policy (who may submit — author,
// admin, neither) is the use case's responsibility.
//
// Approval audit fields (approvedBy / approvedAt) stay untouched
// on Submit — they are populated only by Approve. updatedAt bumps
// to the caller-supplied 'now'.
//
// Atomic: any error leaves the entity untouched.
func (c *Curriculum) SubmitForApproval(now time.Time) error {
	if c.status != StatusDraft {
		return fmt.Errorf("%w: status %q", ErrCannotSubmit, string(c.status))
	}
	c.status = StatusPendingApproval
	c.updatedAt = now
	return nil
}

// Approve transitions a pending_approval curriculum into the
// approved state and records the admin's identity + timestamp on
// the entity. The use case enforces admin-only access via the
// route-level RequireRole(SystemAdmin) middleware plus the
// handler whitelist; the entity enforces only the status
// invariant and a non-zero adminID guard (defense in depth
// against a silent admin scenario where the JWT subject was lost
// upstream).
//
// Atomic: any error leaves the entity untouched.
func (c *Curriculum) Approve(adminID int64, now time.Time) error {
	if adminID <= 0 {
		return fmt.Errorf("%w: admin id must be positive, got %d",
			ErrCannotApprove, adminID)
	}
	if c.status != StatusPendingApproval {
		return fmt.Errorf("%w: status %q", ErrCannotApprove, string(c.status))
	}
	c.status = StatusApproved
	c.approvedBy = &adminID
	at := now
	c.approvedAt = &at
	c.updatedAt = now
	return nil
}

// Reject transitions a pending_approval curriculum back to draft.
// The author (academic secretary per v0.158.0+) may revise the content
// (UpdateBasics is unblocked by status === draft) and then re-submit.
//
// Reject reason is intentionally not part of the entity contract —
// it lives only in the audit log (ADR-3) so a future "rework after
// rejection" cycle doesn't carry persistent rejection context. If
// permanent reason storage becomes a product requirement, add via
// migration without changing this method's shape.
//
// Atomic: any error leaves the entity untouched.
func (c *Curriculum) Reject(now time.Time) error {
	if c.status != StatusPendingApproval {
		return fmt.Errorf("%w: status %q", ErrCannotReject, string(c.status))
	}
	c.status = StatusDraft
	c.updatedAt = now
	return nil
}

// AuthorizeEdit returns nil if the caller may modify this
// curriculum's content via UpdateBasics, or one of the two domain
// sentinels otherwise.
//
// The status gate fires BEFORE the ownership / admin checks:
// approved and pending_approval curricula are frozen for everyone.
// Status transitions (Approve / Reject / Archive) are separate
// domain methods that land in v0.117.0 — UpdateBasics is purely
// content-mutation (title / code / specialty / year / description)
// and runs only against draft curricula.
//
// When isAdmin is true the ownership check is skipped — admins may
// edit any draft. Methodists may edit only their own drafts.
func (c *Curriculum) AuthorizeEdit(actorID int64, isAdmin bool) error {
	if !c.status.CanEdit() {
		return fmt.Errorf("%w: status %q is not editable",
			ErrCannotEditApproved, string(c.status))
	}
	if isAdmin {
		return nil
	}
	if actorID == c.createdBy && actorID > 0 {
		return nil
	}
	return fmt.Errorf("%w: actor %d is not the author (%d)",
		ErrCurriculumScopeForbidden, actorID, c.createdBy)
}
