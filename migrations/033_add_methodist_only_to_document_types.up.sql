-- v0.126.0 Slot D row #11: scope-filter document_types templates so that
-- teachers and students do not see methodist-only paperwork forms in the
-- /documents/templates list. The flag is set by methodist / system_admin
-- when authoring or editing a template; default FALSE keeps every existing
-- template visible to every role (backwards compatible).
ALTER TABLE document_types
    ADD COLUMN IF NOT EXISTS methodist_only BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN document_types.methodist_only IS
    'When TRUE the template is hidden from teacher and student in the templates list (v0.126.0)';
