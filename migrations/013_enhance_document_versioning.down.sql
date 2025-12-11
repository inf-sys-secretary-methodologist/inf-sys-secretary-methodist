-- ============================================================================
-- ROLLBACK: ENHANCED DOCUMENT VERSIONING
-- ============================================================================

-- Удаляем триггер
DROP TRIGGER IF EXISTS tr_document_version_on_update ON documents;

-- Удаляем функцию
DROP FUNCTION IF EXISTS create_document_version_on_update();

-- Удаляем таблицу diff-ов
DROP TABLE IF EXISTS document_version_diffs;

-- Удаляем добавленные колонки из document_versions
ALTER TABLE document_versions
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS subject,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS mime_type,
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS storage_key;

-- Удаляем индекс
DROP INDEX IF EXISTS idx_document_versions_created_at;
