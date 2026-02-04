-- Rollback: Remove template columns from document_types

DROP INDEX IF EXISTS idx_document_types_has_template;

ALTER TABLE document_types
    DROP COLUMN IF EXISTS template_content,
    DROP COLUMN IF EXISTS template_variables;
