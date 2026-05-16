-- v0.148.0 — documents workflow gates closure (issue #227)
-- Adds the columns the Reject path writes: rejection reason + admin
-- who pressed reject + timestamp. Submit/Approve transitions reuse
-- existing columns (status + updated_at).
--
-- Backward-compatible (NULL allowed for legacy rows that never
-- entered the rejection path).

ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS rejected_reason TEXT NULL,
    ADD COLUMN IF NOT EXISTS rejected_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ NULL;

-- Audit trail for approval — mirror к curriculum.approved_by /
-- .approved_at columns. Methodist-cluster роли могут approve;
-- nullable так как pre-v0.148.0 rows никогда не проходили approval gate.

ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS approved_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ NULL;

-- Submit audit columns (who pushed draft → approval queue + when).
-- Author always invokes submit, so submitted_by may differ от author
-- only when an admin force-submits (Phase 2). For now это просто
-- forensic audit trail.

ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS submitted_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ NULL;
