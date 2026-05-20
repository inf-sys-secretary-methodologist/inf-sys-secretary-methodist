-- Rollback v0.157.0 #269 ADR-2 — drop curricula.version column.
ALTER TABLE curricula DROP CONSTRAINT IF EXISTS chk_curricula_version_nonneg;
ALTER TABLE curricula DROP COLUMN IF EXISTS version;
