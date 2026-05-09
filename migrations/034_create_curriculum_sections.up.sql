-- ============================================================================
-- CURRICULUM_SECTIONS — Section aggregate (раздел учебного плана)
-- ============================================================================
-- Establishes the Section aggregate for the curriculum bounded context per
-- ADR-1 (plan 2026-05-09-v0128-section-aggregate.md): independent AR with
-- FK to curricula, no own status (lifecycle inherits curriculum.status per
-- ADR-2). Optimistic locking foundation per ADR-3 (version column,
-- repository UPDATE uses WHERE id = ? AND version = ?).
--
-- v0.128.1 will add curriculum_section_items as a sibling table referencing
-- this one (FK with CASCADE) — items hold the rich academic detail (hours,
-- credits, control_form). v0.128.2+ uses optimistic locking here for
-- bulk-edit conflict detection.
-- ============================================================================
--
-- Lifecycle inheritance (ADR-2):
--   curriculum.status = 'draft'           → sections editable
--   curriculum.status = 'pending_approval' → sections frozen
--   curriculum.status = 'approved'         → sections frozen
--   curriculum.status = 'archived'         → sections frozen
-- The freeze is enforced at the use-case layer (Section.AuthorizeEdit takes
-- curStatus + curCreatedBy primitives); the schema does not encode it as a
-- CHECK because that would require a denormalized status column.
-- ============================================================================
--
-- Hard-delete (ADR-4): undo through UI confirm dialog; audit_sink captures
-- forensic events. No soft-delete column.
--
-- No UNIQUE constraint on (curriculum_id, order_index): bulk reorder with a
-- UNIQUE would require deferred-constraint dance. Stable ORDER BY
-- (order_index, created_at, id) gives deterministic display even with
-- duplicates; reorder UI sets explicit ordering.
-- ============================================================================

CREATE TABLE IF NOT EXISTS curriculum_sections (
    id BIGSERIAL PRIMARY KEY,
    curriculum_id BIGINT NOT NULL REFERENCES curricula(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    order_index INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Defense in depth: domain Section.NewSection / UpdateBasics enforce
    -- the same invariants on writes; CHECKs catch direct SQL bypass and
    -- Reconstitute paths from rows that somehow drifted.
    CONSTRAINT chk_curriculum_sections_title_nonempty CHECK (length(trim(title)) > 0),
    CONSTRAINT chk_curriculum_sections_description_length
        CHECK (description IS NULL OR length(description) <= 4096),
    CONSTRAINT chk_curriculum_sections_order_index_nonneg CHECK (order_index >= 0),
    CONSTRAINT chk_curriculum_sections_version_nonneg CHECK (version >= 0)
);

CREATE INDEX IF NOT EXISTS idx_curriculum_sections_curriculum_id
    ON curriculum_sections(curriculum_id);

COMMENT ON TABLE curriculum_sections IS 'Раздел учебного плана — aggregate root, parent of curriculum_section_items (v0.128.1+)';
COMMENT ON COLUMN curriculum_sections.curriculum_id IS 'FK to curricula.id; CASCADE delete propagates section + item cleanup';
COMMENT ON COLUMN curriculum_sections.title IS 'Section title (e.g. "Базовая часть", "Вариативная часть")';
COMMENT ON COLUMN curriculum_sections.order_index IS 'Display ordering hint (≥ 0); duplicates allowed, ORDER BY (order_index, created_at, id) is deterministic';
COMMENT ON COLUMN curriculum_sections.version IS 'Optimistic-locking version (ADR-3) — repository UPDATE uses WHERE id = ? AND version = ?, increments on success';

-- Reuse the shared updated_at trigger function defined в migration 021
-- (update_attendance_updated_at). Same approach as curricula in
-- migration 031 — single source of truth for NOW() semantics.
CREATE TRIGGER tr_curriculum_sections_updated_at
    BEFORE UPDATE ON curriculum_sections
    FOR EACH ROW
    EXECUTE FUNCTION update_attendance_updated_at();
