-- ============================================================================
-- MINOBRNAUKI_ORDERS — приказы Минобрнауки (ADR-11)
-- ============================================================================
-- Initiative: docs/plans/2026-05-27-work-program-initiative.md
-- Phase: PR 6a of 9 — MinobrnaukiOrder domain + migration + persistence.
--
-- ADR-11: приказ Минобрнауки is the real-world trigger for РПД revisions
-- (изменения ФГОС / методических требований / перечня компетенций),
-- physically arriving as a PDF/Word. A methodist records the order; it
-- then drives affected work programs into needs_revision (use cases land
-- in PR 6b). The order is an immutable artifact — corrections are made by
-- recording a new order (no in-place edits).
--
-- ADR-6 defense-in-depth: every domain invariant in
-- internal/modules/work_program/domain/entities/minobrnauki_order.go is
-- mirrored as a CHECK constraint here. Domain enforces on writes; CHECK
-- catches direct SQL bypass and Reconstitute paths.
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 1) minobrnauki_orders — the recorded order artifact
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS minobrnauki_orders (
    id            BIGSERIAL    PRIMARY KEY,
    order_number  VARCHAR(100) NOT NULL,
    title         TEXT         NOT NULL,
    published_at  DATE         NOT NULL,
    document_id   BIGINT       REFERENCES documents(id) ON DELETE SET NULL,
    change_scope  VARCHAR(20)  NOT NULL,
    summary       TEXT,
    uploaded_by   BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_mo_order_number_nonempty
        CHECK (length(trim(order_number)) > 0),
    CONSTRAINT chk_mo_title_nonempty
        CHECK (length(trim(title)) > 0),
    CONSTRAINT chk_mo_title_length
        CHECK (length(title) <= 1024),
    CONSTRAINT chk_mo_change_scope_enum
        CHECK (change_scope IN ('minor','major')),
    CONSTRAINT chk_mo_summary_length
        CHECK (summary IS NULL OR length(summary) <= 4096),
    -- Mirrors the entity's published_at IsZero() invariant: DATE NOT NULL
    -- blocks NULL but not Go's zero time.Time (0001-01-01), which is a valid
    -- DATE. The 1900 floor catches that and any nonsensical Reconstitute /
    -- direct-SQL path (Минобрнауки orders are modern artifacts).
    CONSTRAINT chk_mo_published_at_sane
        CHECK (published_at > DATE '1900-01-01')
);

CREATE INDEX IF NOT EXISTS idx_minobrnauki_orders_uploaded_by
    ON minobrnauki_orders(uploaded_by);
CREATE INDEX IF NOT EXISTS idx_minobrnauki_orders_published_at
    ON minobrnauki_orders(published_at);
CREATE INDEX IF NOT EXISTS idx_minobrnauki_orders_change_scope
    ON minobrnauki_orders(change_scope);

COMMENT ON TABLE  minobrnauki_orders IS 'Приказ Минобрнауки — внешний регуляторный триггер ревизии РПД per ADR-11';
COMMENT ON COLUMN minobrnauki_orders.change_scope IS 'minor → лист актуализации; major → новая редакция РПД (ADR-11)';
COMMENT ON COLUMN minobrnauki_orders.document_id IS 'Опц. FK на загруженный PDF/Word приказа в documents module';
COMMENT ON COLUMN minobrnauki_orders.summary IS 'Опц. человекочитаемая выжимка изменений (заносит методист)';

-- ----------------------------------------------------------------------------
-- 2) minobrnauki_order_affected — M:N приказ ↔ затронутые РПД (ADR-11)
-- ----------------------------------------------------------------------------
-- Методист отмечает, какие work programs затрагивает приказ (опционально:
-- может записать приказ сначала, отметить затронутые РПД позже).
CREATE TABLE IF NOT EXISTS minobrnauki_order_affected (
    order_id        BIGINT NOT NULL REFERENCES minobrnauki_orders(id) ON DELETE CASCADE,
    work_program_id BIGINT NOT NULL REFERENCES work_programs(id)      ON DELETE CASCADE,
    PRIMARY KEY (order_id, work_program_id)
);

CREATE INDEX IF NOT EXISTS idx_mo_affected_work_program_id
    ON minobrnauki_order_affected(work_program_id);

COMMENT ON TABLE minobrnauki_order_affected IS 'M:N связь приказ ↔ затронутые РПД per ADR-11';

-- ----------------------------------------------------------------------------
-- 3) work_program_revisions.triggered_by_order_id — audit trail (ADR-11)
-- ----------------------------------------------------------------------------
-- Extend the revision table (created in 048) with the audit FK so
-- Рособрнадзор can trace «по какому приказу изменена РПД». ON DELETE SET
-- NULL: removing an order must not cascade-delete the legally significant
-- revision document.
ALTER TABLE work_program_revisions
    ADD COLUMN IF NOT EXISTS triggered_by_order_id BIGINT
        REFERENCES minobrnauki_orders(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_wp_revisions_triggered_by_order_id
    ON work_program_revisions(triggered_by_order_id);

COMMENT ON COLUMN work_program_revisions.triggered_by_order_id IS 'Опц. FK на приказ Минобрнауки, инициировавший ревизию per ADR-11';
