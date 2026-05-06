-- ============================================================================
-- CURRICULA - Academic curriculum (учебный план) bounded context
-- ============================================================================
-- Establishes the Curriculum bounded context: methodists author curricula
-- (учебные планы), administrators approve them. v0.116.0 introduces only the
-- aggregate root table; child entity Discipline (дисциплина) and the approval
-- workflow land in v0.117.0. The status column and approval audit fields are
-- present from day one with nullable defaults so v0.117.0 can ship a code-only
-- approve flow with no further DDL hop.
-- ============================================================================
--
-- Status lifecycle (v0.117.0 will exercise transitions):
--   draft ─SubmitForApproval→ pending_approval ─Approve→ approved ─Archive→ archived
--                                              ─Reject──→ draft
--                                                            ─Archive→ archived
--
-- Editing is permitted only in 'draft' (Curriculum.UpdateBasics enforces it
-- at the domain layer; the schema does not — partial updates of approved
-- curricula are blocked at the use-case boundary, not via DB CHECK, so that
-- v0.117.0 Reject (approved → draft revisit, hypothetical) keeps a clean
-- migration). The approval-audit consistency CHECK below pins the
-- "status=approved implies approved_by/at populated" invariant.
-- ============================================================================

CREATE TABLE IF NOT EXISTS curricula (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    specialty VARCHAR(255) NOT NULL,
    year INT NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    approved_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_curricula_title_nonempty CHECK (length(trim(title)) > 0),
    CONSTRAINT chk_curricula_code_nonempty CHECK (length(trim(code)) > 0),
    CONSTRAINT chk_curricula_specialty_nonempty CHECK (length(trim(specialty)) > 0),
    CONSTRAINT chk_curricula_year_range CHECK (year >= 2000 AND year <= 2100),
    CONSTRAINT chk_curricula_description_length CHECK (description IS NULL OR length(description) <= 4096),
    CONSTRAINT chk_curricula_status_enum CHECK (status IN ('draft', 'pending_approval', 'approved', 'archived')),
    -- Defense in depth: status=approved must carry the actor + timestamp.
    -- The domain Approve() method enforces this on writes; the CHECK catches
    -- direct SQL bypass and Reconstitute paths. Mirrors the
    -- chk_submissions_graded_consistency / chk_submissions_returned_consistency
    -- pattern from migrations 029/030.
    CONSTRAINT chk_curricula_approved_consistency CHECK (
        (status = 'approved' AND approved_by IS NOT NULL AND approved_at IS NOT NULL)
        OR (status <> 'approved')
    )
);

CREATE INDEX IF NOT EXISTS idx_curricula_status ON curricula(status);
CREATE INDEX IF NOT EXISTS idx_curricula_year ON curricula(year);
CREATE INDEX IF NOT EXISTS idx_curricula_specialty ON curricula(specialty);
CREATE INDEX IF NOT EXISTS idx_curricula_created_by ON curricula(created_by);
CREATE INDEX IF NOT EXISTS idx_curricula_approved_by ON curricula(approved_by);

COMMENT ON TABLE curricula IS 'Academic curriculum (учебный план) — methodist-authored, admin-approved';
COMMENT ON COLUMN curricula.title IS 'Human-readable title (e.g. "ИВТ-2026 / 4 года")';
COMMENT ON COLUMN curricula.code IS 'Unique identifier per programme (e.g. ФГОС specialty code + year)';
COMMENT ON COLUMN curricula.specialty IS 'Specialty name (направление подготовки) per ФГОС';
COMMENT ON COLUMN curricula.year IS 'Academic year of programme start (2026 means 2026/2027 учебный год)';
COMMENT ON COLUMN curricula.status IS 'Lifecycle state: draft / pending_approval / approved / archived';
COMMENT ON COLUMN curricula.created_by IS 'Methodist who authored the curriculum';
COMMENT ON COLUMN curricula.approved_by IS 'Administrator who approved the curriculum (NULL until approved)';
COMMENT ON COLUMN curricula.approved_at IS 'When the curriculum reached approved status (NULL until approved)';

-- Reuse the shared updated_at trigger function defined in migration 021
-- (update_attendance_updated_at). Same approach as assignments/submissions
-- in migration 029 — a single source of truth keeps NOW() semantics
-- consistent across modules.
CREATE TRIGGER tr_curricula_updated_at
    BEFORE UPDATE ON curricula
    FOR EACH ROW
    EXECUTE FUNCTION update_attendance_updated_at();
