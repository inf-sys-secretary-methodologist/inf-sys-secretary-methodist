-- Rollback 051: restore the original curriculum_section_items(id) reference.

ALTER TABLE student_debts
    DROP CONSTRAINT IF EXISTS student_debts_discipline_id_fkey;

ALTER TABLE student_debts
    ADD CONSTRAINT student_debts_discipline_id_fkey
    FOREIGN KEY (discipline_id) REFERENCES curriculum_section_items(id) ON DELETE SET NULL;
