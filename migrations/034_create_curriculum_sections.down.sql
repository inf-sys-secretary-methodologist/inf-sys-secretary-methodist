-- Reverse of 034_create_curriculum_sections.up.sql.
-- Rollback drops the table; CASCADE handles index/trigger cleanup. The
-- dependent curriculum_section_items table (v0.128.1, migration 035) will
-- be dropped first by its own .down.sql before this rollback runs in
-- forward-rollback order.

DROP TABLE IF EXISTS curriculum_sections CASCADE;
