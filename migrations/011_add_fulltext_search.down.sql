-- Откат миграции полнотекстового поиска

-- Удаляем функцию подсветки
DROP FUNCTION IF EXISTS highlight_search_result(TEXT, TEXT, TEXT, TEXT);

-- Удаляем индексы
DROP INDEX IF EXISTS idx_documents_search_vector;
DROP INDEX IF EXISTS idx_documents_title_trgm;
DROP INDEX IF EXISTS idx_documents_subject_trgm;

-- Удаляем колонку search_vector
ALTER TABLE documents DROP COLUMN IF EXISTS search_vector;

-- Примечание: расширение pg_trgm не удаляем, т.к. оно может использоваться в других местах
