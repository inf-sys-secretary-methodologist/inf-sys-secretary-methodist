-- ============================================================================
-- DOCUMENT PUBLIC LINKS - Публичные ссылки для документов
-- ============================================================================

-- Таблица публичных ссылок
CREATE TABLE IF NOT EXISTS document_public_links (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL UNIQUE, -- уникальный токен для доступа
    permission VARCHAR(50) NOT NULL DEFAULT 'read' CHECK (permission IN ('read', 'download')),
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP, -- срок действия ссылки (NULL = бессрочно)
    max_uses INT, -- максимальное количество использований (NULL = неограничено)
    use_count INT NOT NULL DEFAULT 0, -- счётчик использований
    password_hash VARCHAR(255), -- опциональный пароль для доступа
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы
CREATE INDEX idx_document_public_links_document_id ON document_public_links(document_id);
CREATE INDEX idx_document_public_links_token ON document_public_links(token);
CREATE INDEX idx_document_public_links_created_by ON document_public_links(created_by);
CREATE INDEX idx_document_public_links_expires_at ON document_public_links(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_document_public_links_is_active ON document_public_links(is_active) WHERE is_active = true;
