-- v0.151.0 — documents workflow Phase 4 Execution (issue #232)
-- Extends migration 041 с executor assignment + execution-completion audit:
--   - executor_assigned_to: who is assigned to execute the document (admin actor);
--   - executor_assigned_at: timestamp of executor assignment;
--   - executor_due_date: optional execution deadline (NULL if no hard deadline);
--   - executed_by: who marked the document as executed (admin actor);
--   - executed_at: timestamp of execution → executed transition.
-- All nullable; pre-v0.151.0 rows (which never traversed the gates) keep
-- clean JSON output via *int64 / *time.Time pointers on the entity.

ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS executor_assigned_to BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS executor_assigned_at TIMESTAMP WITH TIME ZONE NULL,
    ADD COLUMN IF NOT EXISTS executor_due_date DATE NULL,
    ADD COLUMN IF NOT EXISTS executed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS executed_at TIMESTAMP WITH TIME ZONE NULL;
