-- v0.152.0 — documents workflow Phase 5 Archive (issue #233)
-- Extends migration 042 с archive-terminal audit:
--   - archived_by: who archived the executed document (admin actor);
--   - archived_at: timestamp of executed → archived transition.
-- Resubmit (rejected → draft cycle restart) не добавляет audit columns —
-- operation reverts state by nullifying existing rejected_* audit fields;
-- no new persisted information beyond UpdatedAt bump.
-- Both columns nullable; pre-v0.152.0 rows keep NULL via *int64 / *time.Time pointers.

ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS archived_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMP WITH TIME ZONE NULL;
