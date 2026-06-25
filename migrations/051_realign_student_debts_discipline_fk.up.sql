-- Migration 051: realign student_debts.discipline_id to disciplines(id).
--
-- Migration 050 made student_debts.discipline_id reference
-- curriculum_section_items(id). That is inconsistent with the rest of the
-- system: work_program (047) and schedule_lessons (004) both reference
-- disciplines(id) as the canonical "discipline" entity. Teacher scoping for
-- the debt registry resolves a teacher's disciplines through
-- schedule_lessons (teacher_id -> discipline_id, which is disciplines(id));
-- for that scope to share one id space with student_debts.discipline_id,
-- this column must reference disciplines(id) too.
--
-- discipline_id is a best-effort, nullable link that no code populates yet,
-- so realigning the FK target needs no data migration. The inline FK from
-- migration 050 carries PostgreSQL's default name student_debts_discipline_id_fkey.
--
-- This is a versioned (run-once) migration: the DROP ... IF EXISTS guards a
-- name drift, but re-running the whole file would fail on the second
-- ADD CONSTRAINT (by design — golang-migrate applies each version once).

ALTER TABLE student_debts
    DROP CONSTRAINT IF EXISTS student_debts_discipline_id_fkey;

ALTER TABLE student_debts
    ADD CONSTRAINT student_debts_discipline_id_fkey
    FOREIGN KEY (discipline_id) REFERENCES disciplines(id) ON DELETE SET NULL;
