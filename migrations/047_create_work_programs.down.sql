-- Rollback for migration 047 — drop work_programs (PR 1a scope).
-- Inner tables (work_program_goals / ...) ship with PR 1b and have
-- their own down migration.

DROP TABLE IF EXISTS work_programs CASCADE;
