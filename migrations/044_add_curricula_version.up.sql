-- ============================================================================
-- CURRICULA — add version column для optimistic locking (v0.157.0 #269 ADR-2)
-- ============================================================================
-- Mirrors migration 034 (curriculum_sections) + migration 035
-- (curriculum_section_items) which were authored с the version column from
-- day one. The curricula aggregate root was migrated в 031 без the column,
-- leaving Update vulnerable к lost-update races: methodist A and B both
-- load @v0, edit different fields, both write — last writer wins, first
-- methodist's change silently lost.
--
-- DEFAULT 0 backfills existing rows на the baseline version; subsequent
-- repository UPDATE statements add `AND version = ?` clause + bump version
-- on success. rows=0 with row-still-existing → ErrCurriculumVersionConflict.
-- ============================================================================

ALTER TABLE curricula
    ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 0;

-- Idempotent CHECK constraint addition: ADD CONSTRAINT itself is not
-- IF NOT EXISTS in PG 17, so wrap в DO block + EXCEPTION на
-- duplicate_object к support partial-prior-apply scenarios (column
-- added but constraint missing). Drop-add via DO/EXCEPTION mirrors
-- the safe pattern used downstream of pg_dump restores.
DO $$
BEGIN
    ALTER TABLE curricula
        ADD CONSTRAINT chk_curricula_version_nonneg CHECK (version >= 0);
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON COLUMN curricula.version IS 'Optimistic-locking version (v0.157.0 #269) — repository UPDATE uses WHERE id = ? AND version = ?, increments on success';
