-- Rollback for migration 048 — drop inner aggregate tables in reverse
-- creation order. Root table work_programs is owned by migration 047
-- and untouched here.

DROP TABLE IF EXISTS work_program_revisions  CASCADE;
DROP TABLE IF EXISTS work_program_references CASCADE;
DROP TABLE IF EXISTS work_program_assessment CASCADE;
DROP TABLE IF EXISTS work_program_topics     CASCADE;
DROP TABLE IF EXISTS work_program_competences CASCADE;
DROP TABLE IF EXISTS work_program_goals      CASCADE;
