-- v0.150.0 — documents workflow Phase 3 Routing (issue #231)
-- Extends migration 040 с routing + visa audit trail:
--   - routed_by: who pushed the doc to routing (admin actor);
--   - routed_at: timestamp of registered → routing transition;
--   - visa_signed_by: who signed the visa (admin actor — single-step);
--   - visa_signed_at: timestamp of routing → execution transition.
-- All nullable; pre-v0.150.0 rows (which never traversed the gates) keep
-- clean JSON output via *int64 / *time.Time pointers on the entity.

ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS routed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS routed_at TIMESTAMP WITH TIME ZONE NULL,
    ADD COLUMN IF NOT EXISTS visa_signed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS visa_signed_at TIMESTAMP WITH TIME ZONE NULL;
