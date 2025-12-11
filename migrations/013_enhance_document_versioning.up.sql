-- ============================================================================
-- ENHANCED DOCUMENT VERSIONING - Расширенное версионирование документов
-- Issue #12: Implement document version control
-- ============================================================================

-- Расширяем таблицу document_versions для хранения полного снимка документа
ALTER TABLE document_versions
    ADD COLUMN IF NOT EXISTS title VARCHAR(500),
    ADD COLUMN IF NOT EXISTS subject TEXT,
    ADD COLUMN IF NOT EXISTS status VARCHAR(50),
    ADD COLUMN IF NOT EXISTS mime_type VARCHAR(100),
    ADD COLUMN IF NOT EXISTS metadata JSONB,
    ADD COLUMN IF NOT EXISTS storage_key VARCHAR(1000);

-- Комментарии к новым колонкам
COMMENT ON COLUMN document_versions.title IS 'Заголовок документа на момент создания версии';
COMMENT ON COLUMN document_versions.subject IS 'Тема документа на момент создания версии';
COMMENT ON COLUMN document_versions.status IS 'Статус документа на момент создания версии';
COMMENT ON COLUMN document_versions.mime_type IS 'MIME тип файла версии';
COMMENT ON COLUMN document_versions.metadata IS 'Дополнительные метаданные версии';
COMMENT ON COLUMN document_versions.storage_key IS 'Ключ файла версии в MinIO хранилище';

-- Индекс для быстрого поиска версий документа
CREATE INDEX IF NOT EXISTS idx_document_versions_created_at ON document_versions(created_at DESC);

-- Таблица для сравнения версий (кэш diff-ов)
CREATE TABLE IF NOT EXISTS document_version_diffs (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    from_version INT NOT NULL,
    to_version INT NOT NULL,

    -- Что изменилось
    changed_fields TEXT[] NOT NULL DEFAULT '{}',  -- Массив названий измененных полей
    diff_data JSONB,                               -- Подробные различия {field: {from: ..., to: ...}}

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_document_version_diff UNIQUE(document_id, from_version, to_version),
    CONSTRAINT different_versions CHECK (from_version != to_version)
);

CREATE INDEX IF NOT EXISTS idx_doc_version_diffs_document_id ON document_version_diffs(document_id);

COMMENT ON TABLE document_version_diffs IS 'Кэш сравнений между версиями документов';
COMMENT ON COLUMN document_version_diffs.changed_fields IS 'Список измененных полей между версиями';
COMMENT ON COLUMN document_version_diffs.diff_data IS 'JSON с детальными изменениями {field: {from: old_value, to: new_value}}';

-- Функция для автоматического создания версии при обновлении документа
CREATE OR REPLACE FUNCTION create_document_version_on_update()
RETURNS TRIGGER AS $$
DECLARE
    max_version INT;
BEGIN
    -- Создаем версию только если есть изменения в ключевых полях
    IF OLD.title IS DISTINCT FROM NEW.title
       OR OLD.subject IS DISTINCT FROM NEW.subject
       OR OLD.content IS DISTINCT FROM NEW.content
       OR OLD.file_path IS DISTINCT FROM NEW.file_path
       OR OLD.status IS DISTINCT FROM NEW.status THEN

        -- Найдем максимальный номер версии для этого документа
        SELECT COALESCE(MAX(version), 0) INTO max_version
        FROM document_versions
        WHERE document_id = OLD.id;

        -- Используем max_version + 1 для новой версии (сохраняем OLD состояние)
        INSERT INTO document_versions (
            document_id, version, title, subject, content,
            file_name, file_path, file_size, mime_type, storage_key,
            status, metadata, changed_by, change_description, created_at
        ) VALUES (
            OLD.id, max_version + 1, OLD.title, OLD.subject, OLD.content,
            OLD.file_name, OLD.file_path, OLD.file_size, OLD.mime_type, OLD.file_path,
            OLD.status, OLD.metadata,
            COALESCE(current_setting('app.current_user_id', true)::BIGINT, OLD.author_id),
            'Автоматическая версия при обновлении',
            NOW()
        );

        -- Устанавливаем версию документа на max_version + 2
        NEW.version := max_version + 2;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Создаем триггер (если не существует)
DROP TRIGGER IF EXISTS tr_document_version_on_update ON documents;
CREATE TRIGGER tr_document_version_on_update
    BEFORE UPDATE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION create_document_version_on_update();
