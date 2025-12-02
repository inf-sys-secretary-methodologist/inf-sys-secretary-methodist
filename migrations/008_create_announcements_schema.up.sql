-- ============================================================================
-- ANNOUNCEMENTS MODULE - Система объявлений и новостей
-- ============================================================================

-- Объявления
CREATE TABLE IF NOT EXISTS announcements (
    id BIGSERIAL PRIMARY KEY,

    -- Основная информация
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    summary VARCHAR(1000),

    -- Автор
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- Статус и приоритет
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    priority VARCHAR(50) NOT NULL DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),

    -- Целевая аудитория
    target_audience VARCHAR(50) NOT NULL DEFAULT 'all' CHECK (target_audience IN ('all', 'students', 'teachers', 'staff', 'admins')),

    -- Даты публикации
    publish_at TIMESTAMP,
    expire_at TIMESTAMP,

    -- Закрепление и просмотры
    is_pinned BOOLEAN NOT NULL DEFAULT false,
    view_count BIGINT NOT NULL DEFAULT 0,

    -- Метаданные
    tags TEXT[],
    metadata JSONB,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для announcements
CREATE INDEX idx_announcements_author_id ON announcements(author_id);
CREATE INDEX idx_announcements_status ON announcements(status);
CREATE INDEX idx_announcements_priority ON announcements(priority);
CREATE INDEX idx_announcements_target_audience ON announcements(target_audience);
CREATE INDEX idx_announcements_publish_at ON announcements(publish_at);
CREATE INDEX idx_announcements_expire_at ON announcements(expire_at);
CREATE INDEX idx_announcements_is_pinned ON announcements(is_pinned) WHERE is_pinned = true;
CREATE INDEX idx_announcements_tags ON announcements USING GIN (tags);
CREATE INDEX idx_announcements_created_at ON announcements(created_at DESC);

-- Составной индекс для публичных объявлений
CREATE INDEX idx_announcements_published ON announcements(status, publish_at, expire_at)
    WHERE status = 'published';

-- Вложенные файлы к объявлениям
CREATE TABLE IF NOT EXISTS announcement_attachments (
    id BIGSERIAL PRIMARY KEY,
    announcement_id BIGINT NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    file_name VARCHAR(500) NOT NULL,
    file_path VARCHAR(1000) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100),
    uploaded_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_announcement_attachments_announcement_id ON announcement_attachments(announcement_id);

-- Комментарий к таблице
COMMENT ON TABLE announcements IS 'Объявления и новости организации';
COMMENT ON COLUMN announcements.status IS 'Статус: draft - черновик, published - опубликовано, archived - в архиве';
COMMENT ON COLUMN announcements.priority IS 'Приоритет: low - низкий, normal - обычный, high - высокий, urgent - срочный';
COMMENT ON COLUMN announcements.target_audience IS 'Целевая аудитория: all - все, students - студенты, teachers - преподаватели, staff - сотрудники, admins - администраторы';
COMMENT ON COLUMN announcements.publish_at IS 'Дата и время публикации (для отложенной публикации)';
COMMENT ON COLUMN announcements.expire_at IS 'Дата и время окончания показа';
COMMENT ON COLUMN announcements.is_pinned IS 'Закреплено ли объявление вверху списка';
