-- ============================================================================
-- FULL-TEXT SEARCH - Полнотекстовый поиск документов
-- ============================================================================

-- Добавляем колонку search_vector для хранения tsvector
-- Используем русскую конфигурацию для корректной обработки русского языка
ALTER TABLE documents ADD COLUMN IF NOT EXISTS search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('russian', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('russian', coalesce(subject, '')), 'B') ||
        setweight(to_tsvector('russian', coalesce(content, '')), 'C') ||
        setweight(to_tsvector('russian', coalesce(registration_number, '')), 'A')
    ) STORED;

-- Создаём GIN индекс для быстрого полнотекстового поиска
CREATE INDEX IF NOT EXISTS idx_documents_search_vector ON documents USING GIN (search_vector);

-- Также добавим индекс для триграмного поиска (fuzzy search)
-- Это позволит искать с опечатками
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_documents_title_trgm ON documents USING GIN (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_documents_subject_trgm ON documents USING GIN (subject gin_trgm_ops);

-- Функция для получения подсветки результатов поиска
CREATE OR REPLACE FUNCTION highlight_search_result(
    doc_title TEXT,
    doc_subject TEXT,
    doc_content TEXT,
    search_query TEXT
) RETURNS TABLE (
    highlighted_title TEXT,
    highlighted_subject TEXT,
    highlighted_content TEXT
) AS $$
BEGIN
    RETURN QUERY SELECT
        ts_headline('russian', coalesce(doc_title, ''), plainto_tsquery('russian', search_query),
            'StartSel=<mark>, StopSel=</mark>, MaxWords=50, MinWords=25, MaxFragments=3') AS highlighted_title,
        ts_headline('russian', coalesce(doc_subject, ''), plainto_tsquery('russian', search_query),
            'StartSel=<mark>, StopSel=</mark>, MaxWords=50, MinWords=25, MaxFragments=3') AS highlighted_subject,
        ts_headline('russian', coalesce(doc_content, ''), plainto_tsquery('russian', search_query),
            'StartSel=<mark>, StopSel=</mark>, MaxWords=100, MinWords=50, MaxFragments=3') AS highlighted_content;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Комментарии для документации
COMMENT ON COLUMN documents.search_vector IS 'tsvector для полнотекстового поиска (автоматически генерируется)';
COMMENT ON INDEX idx_documents_search_vector IS 'GIN индекс для полнотекстового поиска';
COMMENT ON FUNCTION highlight_search_result IS 'Функция для получения подсвеченных фрагментов результатов поиска';
