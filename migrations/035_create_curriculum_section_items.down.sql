-- Reverse of 035_create_curriculum_section_items.up.sql.
-- Rollback drops the table; CASCADE handles index/trigger cleanup.

DROP TABLE IF EXISTS curriculum_section_items CASCADE;
