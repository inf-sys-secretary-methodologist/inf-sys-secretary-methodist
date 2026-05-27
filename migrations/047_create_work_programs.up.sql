-- ============================================================================
-- WORK_PROGRAMS module — РПД (рабочая программа дисциплины)
-- ============================================================================
-- Initiative: docs/plans/2026-05-27-work-program-initiative.md
-- Phase: PR 1 of 9 — domain + migration.
--
-- ADR-1: WorkProgram is the aggregate root. Inner entities (Goal,
-- Competence, Topic, AssessmentCriterion, Reference, Revision) ship
-- with their own tables in the same migration so the schema is
-- consistent в one apply.
--
-- ADR-3: Identity = (discipline_id, specialty_code, applicable_from_year),
-- where applicable_from_year is the year of student cohort intake (not
-- the current academic year). Audit trail per Рособрнадзор 6-year
-- retention.
--
-- ADR-6: defense-in-depth — every domain invariant in
-- internal/modules/work_program/domain/entities/work_program.go is
-- mirrored as a CHECK constraint here. Domain enforces on writes;
-- CHECK catches direct SQL bypass and Reconstitute paths.
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 1) work_programs — aggregate root
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS work_programs (
    id                   BIGSERIAL    PRIMARY KEY,
    discipline_id        BIGINT       NOT NULL REFERENCES disciplines(id) ON DELETE RESTRICT,
    specialty_code       VARCHAR(20)  NOT NULL,
    applicable_from_year INT          NOT NULL,
    title                VARCHAR(500) NOT NULL,
    annotation           TEXT,
    status               VARCHAR(20)  NOT NULL DEFAULT 'draft',
    author_id            BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    approver_id          BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    approved_at          TIMESTAMPTZ,
    reject_reason        TEXT,
    version              INT          NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_wp_title_nonempty
        CHECK (length(trim(title)) > 0),
    CONSTRAINT chk_wp_specialty_code_nonempty
        CHECK (length(trim(specialty_code)) > 0),
    CONSTRAINT chk_wp_year_range
        CHECK (applicable_from_year BETWEEN 2000 AND 2100),
    CONSTRAINT chk_wp_annotation_length
        CHECK (annotation IS NULL OR length(annotation) <= 8192),
    CONSTRAINT chk_wp_version_nonneg
        CHECK (version >= 0),
    CONSTRAINT chk_wp_status_enum
        CHECK (status IN ('draft','pending_approval','approved','needs_revision','archived')),
    CONSTRAINT chk_wp_approved_consistency
        CHECK (
            (status = 'approved' AND approver_id IS NOT NULL AND approved_at IS NOT NULL)
            OR status <> 'approved'
        ),
    CONSTRAINT uq_wp_discipline_specialty_cohort
        UNIQUE (discipline_id, specialty_code, applicable_from_year)
);

CREATE INDEX IF NOT EXISTS idx_work_programs_discipline_id
    ON work_programs(discipline_id);
CREATE INDEX IF NOT EXISTS idx_work_programs_author_id
    ON work_programs(author_id);
CREATE INDEX IF NOT EXISTS idx_work_programs_status
    ON work_programs(status);
CREATE INDEX IF NOT EXISTS idx_work_programs_specialty_year
    ON work_programs(specialty_code, applicable_from_year);

COMMENT ON TABLE  work_programs IS 'Рабочая программа дисциплины (РПД), aggregate root per ADR-1';
COMMENT ON COLUMN work_programs.applicable_from_year IS 'Год набора студентов (cohort), не текущий academic_year per ADR-3';
COMMENT ON COLUMN work_programs.specialty_code IS 'Код направления (string, not FK) — archived РПД остаются interpret-able';
COMMENT ON COLUMN work_programs.status IS 'FSM per ADR-2: draft / pending_approval / approved / needs_revision / archived';
COMMENT ON COLUMN work_programs.version IS 'Optimistic-lock counter (lost-update guard), starts at 0';

-- Inner aggregate tables (work_program_goals / _competences / _topics /
-- _assessment / _references / _revisions) ship with PR 1b — see
-- docs/plans/2026-05-27-work-program-initiative.md ADR-6, ADR-10. This
-- migration intentionally lands only the root так чтобы PR 1a fits
-- the PR-Size gate; PR 1b adds children atomically.
