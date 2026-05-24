-- ============================================================================
-- EXTRACURRICULAR EVENTS — B3 greenfield bounded context (v0.162.0)
-- ============================================================================
-- ADR-1 (plan docs/plans/2026-05-24-b3-extracurricular.md):
--   ExtracurricularEvent — aggregate root; Participant — inner entity in
--   separate table (extracurricular_participants) with FK back. Capacity
--   invariant `len(participants) <= max_capacity` enforced в domain;
--   CHECK constraints in this migration provide defense-in-depth.
--
-- ADR-4 (schema): see plan doc.
-- ADR-5 (optimistic locking): version column on events table.
-- ============================================================================

CREATE TABLE IF NOT EXISTS extracurricular_events (
    id              BIGSERIAL PRIMARY KEY,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    category        VARCHAR(32)  NOT NULL,
    target_audience VARCHAR(32)  NOT NULL,
    status          VARCHAR(32)  NOT NULL DEFAULT 'draft',
    location        VARCHAR(255),
    start_at        TIMESTAMPTZ  NOT NULL,
    end_at          TIMESTAMPTZ  NOT NULL,
    max_capacity    INT,
    organizer_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    version         INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_extracurricular_title_nonempty
        CHECK (length(trim(title)) > 0),
    CONSTRAINT chk_extracurricular_description_length
        CHECK (description IS NULL OR length(description) <= 4096),
    CONSTRAINT chk_extracurricular_location_length
        CHECK (location IS NULL OR length(location) <= 255),
    CONSTRAINT chk_extracurricular_category
        CHECK (category IN ('cultural','sports','recreational','educational','other')),
    CONSTRAINT chk_extracurricular_audience
        CHECK (target_audience IN ('all','students','teachers','staff')),
    CONSTRAINT chk_extracurricular_status
        CHECK (status IN ('draft','published','canceled','completed')),
    CONSTRAINT chk_extracurricular_time_order
        CHECK (start_at < end_at),
    CONSTRAINT chk_extracurricular_capacity_nonneg
        CHECK (max_capacity IS NULL OR max_capacity >= 0),
    CONSTRAINT chk_extracurricular_version_nonneg
        CHECK (version >= 0)
);

CREATE INDEX IF NOT EXISTS idx_extracurricular_events_organizer_id
    ON extracurricular_events (organizer_id);
CREATE INDEX IF NOT EXISTS idx_extracurricular_events_status_start_at
    ON extracurricular_events (status, start_at);
CREATE INDEX IF NOT EXISTS idx_extracurricular_events_target_audience
    ON extracurricular_events (target_audience);

COMMENT ON COLUMN extracurricular_events.version IS
    'Optimistic-locking version (v0.162.0 #319) — repository UPDATE uses WHERE id = ? AND version = ?, increments on success per ADR-5';

COMMENT ON COLUMN extracurricular_events.max_capacity IS
    'NULL = unlimited; 0 = view-only event (no registration); positive = participant cap';

CREATE TABLE IF NOT EXISTS extracurricular_participants (
    id            BIGSERIAL PRIMARY KEY,
    event_id      BIGINT NOT NULL REFERENCES extracurricular_events(id) ON DELETE CASCADE,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_extracurricular_participants_event_user
        UNIQUE (event_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_extracurricular_participants_event_id
    ON extracurricular_participants (event_id);
CREATE INDEX IF NOT EXISTS idx_extracurricular_participants_user_id
    ON extracurricular_participants (user_id);

COMMENT ON TABLE extracurricular_participants IS
    'Inner entity of the ExtracurricularEvent aggregate (B3 v0.162.0 #319) — registration rows. UNIQUE (event_id, user_id) pins the domain invariant "no double-registration"';
