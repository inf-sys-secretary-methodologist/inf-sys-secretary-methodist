-- ============================================================================
-- CURRICULUM_SECTION_ITEMS — DisciplineItem aggregate (Layer 2)
-- ============================================================================
-- Layer 2 of two-level hierarchy Curriculum → Sections → DisciplineItems
-- per plan docs/plans/2026-05-09-v0128-section-aggregate.md ADR-1 Beta:
-- independent AR with FK to curriculum_sections (CASCADE), no own status
-- (lifecycle inherits curriculum.status via section per ADR-2). Optimistic
-- locking foundation per ADR-3 (version column).
--
-- v0.128.2 will add bulk-edit endpoint with transactional commit-or-rollback
-- on top of this table; optimistic-lock enforcement gates concurrent edits.
-- ============================================================================
--
-- Rich invariants vs Section (Layer 1):
--   - 4 hours columns (lectures / practice / lab / self) — separate так что
--     bulk-edit table view может render 4-column hours grid;
--   - control_form enum: zachet / exam / course_project / differential_zachet
--     (РФ academic standard, 4 forms);
--   - credits ≥ 0 (ECTS-style credit count);
--   - semester ∈ [1, 12] (covers bachelor 8 + master 4).
-- ============================================================================
--
-- Hard-delete (ADR-4): undo through UI confirm dialog; audit_sink captures
-- forensic events. No soft-delete column.
--
-- No UNIQUE constraint on (section_id, order_index) per ADR-4: bulk reorder
-- with UNIQUE = deferred-constraint headache. Stable ORDER BY
-- (order_index, created_at, id) gives deterministic display.
-- ============================================================================

CREATE TABLE IF NOT EXISTS curriculum_section_items (
    id BIGSERIAL PRIMARY KEY,
    section_id BIGINT NOT NULL REFERENCES curriculum_sections(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    hours_lectures INT NOT NULL DEFAULT 0,
    hours_practice INT NOT NULL DEFAULT 0,
    hours_lab INT NOT NULL DEFAULT 0,
    hours_self INT NOT NULL DEFAULT 0,
    control_form TEXT NOT NULL,
    credits INT NOT NULL DEFAULT 0,
    semester INT NOT NULL,
    order_index INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Defense in depth: domain DisciplineItem.NewDisciplineItem / UpdateBasics
    -- enforce the same invariants on writes; CHECKs catch direct SQL bypass
    -- and Reconstitute paths from rows that somehow drifted.
    CONSTRAINT chk_section_items_title_nonempty CHECK (length(trim(title)) > 0),
    CONSTRAINT chk_section_items_hours_lectures_nonneg CHECK (hours_lectures >= 0),
    CONSTRAINT chk_section_items_hours_practice_nonneg CHECK (hours_practice >= 0),
    CONSTRAINT chk_section_items_hours_lab_nonneg CHECK (hours_lab >= 0),
    CONSTRAINT chk_section_items_hours_self_nonneg CHECK (hours_self >= 0),
    CONSTRAINT chk_section_items_control_form_enum
        CHECK (control_form IN ('zachet', 'exam', 'course_project', 'differential_zachet')),
    CONSTRAINT chk_section_items_credits_nonneg CHECK (credits >= 0),
    CONSTRAINT chk_section_items_semester_range CHECK (semester >= 1 AND semester <= 12),
    CONSTRAINT chk_section_items_order_index_nonneg CHECK (order_index >= 0),
    CONSTRAINT chk_section_items_version_nonneg CHECK (version >= 0)
);

CREATE INDEX IF NOT EXISTS idx_section_items_section_id
    ON curriculum_section_items(section_id);

COMMENT ON TABLE curriculum_section_items IS 'Дисциплина учебного плана — Layer 2 AR within Section, holds rich academic detail (hours / credits / control_form / semester)';
COMMENT ON COLUMN curriculum_section_items.section_id IS 'FK to curriculum_sections.id; CASCADE delete propagates from section delete (and curriculum delete via section CASCADE chain)';
COMMENT ON COLUMN curriculum_section_items.title IS 'Discipline name (e.g. "Математический анализ", "Программирование")';
COMMENT ON COLUMN curriculum_section_items.control_form IS 'РФ academic control form: zachet / exam / course_project / differential_zachet';
COMMENT ON COLUMN curriculum_section_items.credits IS 'ECTS-style credit count (≥ 0)';
COMMENT ON COLUMN curriculum_section_items.semester IS 'Academic semester (1..12 — covers bachelor 8 + master 4)';
COMMENT ON COLUMN curriculum_section_items.order_index IS 'Display ordering hint (≥ 0); duplicates allowed, ORDER BY (order_index, created_at, id) is deterministic';
COMMENT ON COLUMN curriculum_section_items.version IS 'Optimistic-locking version (ADR-3) — repository UPDATE uses WHERE id = ? AND version = ?, increments on success';

-- Reuse the shared updated_at trigger function (migration 021). Same
-- approach as curricula (031) and curriculum_sections (034).
CREATE TRIGGER tr_curriculum_section_items_updated_at
    BEFORE UPDATE ON curriculum_section_items
    FOR EACH ROW
    EXECUTE FUNCTION update_attendance_updated_at();
