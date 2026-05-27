-- ============================================================================
-- WORK_PROGRAMS module — inner aggregate tables
-- ============================================================================
-- Initiative: docs/plans/2026-05-27-work-program-initiative.md
-- Phase: PR 1b of 9 — inner aggregates (Goal/Competence/Topic/
-- AssessmentCriterion/Reference/Revision per ADR-1, ADR-10).
-- Root table work_programs created in migration 047.
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 2) work_program_goals — цели и задачи освоения дисциплины (ADR-1)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS work_program_goals (
    id              BIGSERIAL   PRIMARY KEY,
    work_program_id BIGINT      NOT NULL REFERENCES work_programs(id) ON DELETE CASCADE,
    text            TEXT        NOT NULL,
    order_index     INT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_wpg_text_nonempty   CHECK (length(trim(text)) > 0),
    CONSTRAINT chk_wpg_text_length     CHECK (length(text) <= 2048),
    CONSTRAINT chk_wpg_order_nonneg    CHECK (order_index >= 0)
);
CREATE INDEX IF NOT EXISTS idx_wp_goals_work_program_id ON work_program_goals(work_program_id);

-- ----------------------------------------------------------------------------
-- 3) work_program_competences — ПК/ОК/УК per ФГОС (ADR-1)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS work_program_competences (
    id              BIGSERIAL   PRIMARY KEY,
    work_program_id BIGINT      NOT NULL REFERENCES work_programs(id) ON DELETE CASCADE,
    code            VARCHAR(50) NOT NULL,
    type            VARCHAR(10) NOT NULL,
    description     TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_wpc_code_nonempty        CHECK (length(trim(code)) > 0),
    CONSTRAINT chk_wpc_type_enum            CHECK (type IN ('pk','ok','uk')),
    CONSTRAINT chk_wpc_description_nonempty CHECK (length(trim(description)) > 0),
    CONSTRAINT chk_wpc_description_length   CHECK (length(description) <= 2048),
    CONSTRAINT uq_wpc_program_code          UNIQUE (work_program_id, code)
);
CREATE INDEX IF NOT EXISTS idx_wp_competences_work_program_id ON work_program_competences(work_program_id);

-- ----------------------------------------------------------------------------
-- 4) work_program_topics — темы лекций/практик/лабораторных (ADR-1)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS work_program_topics (
    id                BIGSERIAL    PRIMARY KEY,
    work_program_id   BIGINT       NOT NULL REFERENCES work_programs(id) ON DELETE CASCADE,
    kind              VARCHAR(20)  NOT NULL,
    title             VARCHAR(500) NOT NULL,
    hours             INT          NOT NULL,
    week_number       INT,
    learning_outcomes TEXT,
    order_index       INT          NOT NULL DEFAULT 0,
    CONSTRAINT chk_wpt_kind_enum            CHECK (kind IN ('lecture','practice','lab','self_study')),
    CONSTRAINT chk_wpt_title_nonempty       CHECK (length(trim(title)) > 0),
    CONSTRAINT chk_wpt_hours_positive       CHECK (hours > 0),
    CONSTRAINT chk_wpt_week_range           CHECK (week_number IS NULL OR (week_number BETWEEN 1 AND 52)),
    CONSTRAINT chk_wpt_outcomes_length      CHECK (learning_outcomes IS NULL OR length(learning_outcomes) <= 2048),
    CONSTRAINT chk_wpt_order_nonneg         CHECK (order_index >= 0)
);
CREATE INDEX IF NOT EXISTS idx_wp_topics_work_program_id ON work_program_topics(work_program_id);

-- ----------------------------------------------------------------------------
-- 5) work_program_assessment — ФОС: current/intermediate/final controls (ADR-1)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS work_program_assessment (
    id                BIGSERIAL   PRIMARY KEY,
    work_program_id   BIGINT      NOT NULL REFERENCES work_programs(id) ON DELETE CASCADE,
    type              VARCHAR(20) NOT NULL,
    description       TEXT        NOT NULL,
    max_score         INT         NOT NULL,
    example_questions TEXT[],
    CONSTRAINT chk_wpa_type_enum            CHECK (type IN ('current','intermediate','final')),
    CONSTRAINT chk_wpa_description_nonempty CHECK (length(trim(description)) > 0),
    CONSTRAINT chk_wpa_score_range          CHECK (max_score > 0 AND max_score <= 100)
);
CREATE INDEX IF NOT EXISTS idx_wp_assessment_work_program_id ON work_program_assessment(work_program_id);

-- ----------------------------------------------------------------------------
-- 6) work_program_references — литература основная/дополнительная/электронная (ADR-1)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS work_program_references (
    id              BIGSERIAL    PRIMARY KEY,
    work_program_id BIGINT       NOT NULL REFERENCES work_programs(id) ON DELETE CASCADE,
    kind            VARCHAR(20)  NOT NULL,
    citation        TEXT         NOT NULL,
    year            INT,
    isbn            VARCHAR(20),
    url             VARCHAR(500),
    order_index     INT          NOT NULL DEFAULT 0,
    CONSTRAINT chk_wpr_kind_enum          CHECK (kind IN ('main','additional','electronic')),
    CONSTRAINT chk_wpr_citation_nonempty  CHECK (length(trim(citation)) > 0),
    CONSTRAINT chk_wpr_year_range         CHECK (year IS NULL OR (year BETWEEN 1900 AND 2100)),
    CONSTRAINT chk_wpr_order_nonneg       CHECK (order_index >= 0)
);
CREATE INDEX IF NOT EXISTS idx_wp_references_work_program_id ON work_program_references(work_program_id);

-- ----------------------------------------------------------------------------
-- 7) work_program_revisions — лист актуализации (ADR-10)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS work_program_revisions (
    id               BIGSERIAL   PRIMARY KEY,
    work_program_id  BIGINT      NOT NULL REFERENCES work_programs(id) ON DELETE CASCADE,
    revision_number  INT         NOT NULL,
    change_type      VARCHAR(20) NOT NULL,
    change_summary   TEXT        NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'draft',
    author_id        BIGINT      NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    approver_id      BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    approved_at      TIMESTAMPTZ,
    reject_reason    TEXT,
    diff_payload     JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_wprev_revision_positive    CHECK (revision_number > 0),
    CONSTRAINT chk_wprev_status_enum          CHECK (status IN ('draft','pending_approval','approved','rejected')),
    CONSTRAINT chk_wprev_change_type_enum     CHECK (change_type IN ('hours','semester','literature','assessment','other')),
    CONSTRAINT chk_wprev_summary_nonempty     CHECK (length(trim(change_summary)) > 0),
    CONSTRAINT chk_wprev_summary_length       CHECK (length(change_summary) <= 4096),
    CONSTRAINT chk_wprev_approved_consistency CHECK (
        (status = 'approved' AND approver_id IS NOT NULL AND approved_at IS NOT NULL)
        OR status <> 'approved'
    ),
    CONSTRAINT uq_wprev_program_revision      UNIQUE (work_program_id, revision_number)
);
CREATE INDEX IF NOT EXISTS idx_wp_revisions_work_program_id ON work_program_revisions(work_program_id);
CREATE INDEX IF NOT EXISTS idx_wp_revisions_status          ON work_program_revisions(status);

COMMENT ON TABLE work_program_revisions IS 'Лист актуализации РПД — минор-изменение без полного reapproval per ADR-10';
COMMENT ON COLUMN work_program_revisions.diff_payload IS 'Structured before/after JSON, опционально, для audit trail';
