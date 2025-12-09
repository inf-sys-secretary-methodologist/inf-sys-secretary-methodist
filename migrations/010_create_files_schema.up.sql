-- ============================================================================
-- FILES MODULE - Файловое хранилище (2 таблицы)
-- ============================================================================

-- Метаданные файлов
CREATE TABLE file_metadata (
    id BIGSERIAL PRIMARY KEY,
    original_name VARCHAR(500) NOT NULL,           -- Оригинальное имя файла
    storage_key VARCHAR(1000) NOT NULL UNIQUE,     -- Ключ в S3/MinIO хранилище
    size BIGINT NOT NULL,                          -- Размер файла в байтах
    mime_type VARCHAR(100) NOT NULL,               -- MIME тип файла
    checksum VARCHAR(64) NOT NULL,                 -- SHA-256 хеш файла
    uploaded_by BIGINT NOT NULL REFERENCES users(id), -- Кто загрузил файл
    document_id BIGINT REFERENCES documents(id),   -- Связь с документом (опционально)
    task_id BIGINT REFERENCES tasks(id),           -- Связь с задачей (опционально)
    announcement_id BIGINT REFERENCES announcements(id), -- Связь с объявлением (опционально)
    is_temporary BOOLEAN NOT NULL DEFAULT true,    -- Временный файл (до прикрепления)
    expires_at TIMESTAMP,                          -- Срок жизни временного файла
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP                           -- Мягкое удаление
);

-- Индексы для file_metadata
CREATE INDEX idx_file_metadata_uploaded_by ON file_metadata(uploaded_by);
CREATE INDEX idx_file_metadata_document_id ON file_metadata(document_id) WHERE document_id IS NOT NULL;
CREATE INDEX idx_file_metadata_task_id ON file_metadata(task_id) WHERE task_id IS NOT NULL;
CREATE INDEX idx_file_metadata_announcement_id ON file_metadata(announcement_id) WHERE announcement_id IS NOT NULL;
CREATE INDEX idx_file_metadata_is_temporary ON file_metadata(is_temporary) WHERE is_temporary = true;
CREATE INDEX idx_file_metadata_expires_at ON file_metadata(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_file_metadata_deleted_at ON file_metadata(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_file_metadata_created_at ON file_metadata(created_at);
CREATE INDEX idx_file_metadata_checksum ON file_metadata(checksum);

COMMENT ON TABLE file_metadata IS 'Метаданные файлов в хранилище MinIO/S3';
COMMENT ON COLUMN file_metadata.original_name IS 'Оригинальное имя файла при загрузке';
COMMENT ON COLUMN file_metadata.storage_key IS 'Уникальный ключ файла в S3/MinIO';
COMMENT ON COLUMN file_metadata.checksum IS 'SHA-256 хеш для проверки целостности';
COMMENT ON COLUMN file_metadata.is_temporary IS 'true = временный файл, ожидающий прикрепления';
COMMENT ON COLUMN file_metadata.expires_at IS 'Срок жизни временного файла для автоудаления';

-- Версии файлов (для версионирования документов)
CREATE TABLE file_versions (
    id BIGSERIAL PRIMARY KEY,
    file_metadata_id BIGINT NOT NULL REFERENCES file_metadata(id) ON DELETE CASCADE,
    version_number INT NOT NULL,                   -- Номер версии
    storage_key VARCHAR(1000) NOT NULL,            -- Ключ версии в S3/MinIO
    size BIGINT NOT NULL,                          -- Размер версии в байтах
    checksum VARCHAR(64) NOT NULL,                 -- SHA-256 хеш версии
    comment VARCHAR(500),                          -- Комментарий к версии
    created_by BIGINT NOT NULL REFERENCES users(id), -- Кто создал версию
    created_at TIMESTAMP NOT NULL DEFAULT now(),

    CONSTRAINT uq_file_versions_file_version UNIQUE(file_metadata_id, version_number)
);

-- Индексы для file_versions
CREATE INDEX idx_file_versions_file_metadata_id ON file_versions(file_metadata_id);
CREATE INDEX idx_file_versions_created_by ON file_versions(created_by);
CREATE INDEX idx_file_versions_created_at ON file_versions(created_at);

COMMENT ON TABLE file_versions IS 'Версии файлов для поддержки версионирования документов';
COMMENT ON COLUMN file_versions.version_number IS 'Номер версии, начиная с 1';
COMMENT ON COLUMN file_versions.comment IS 'Описание изменений в версии';
