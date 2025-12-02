-- ============================================================================
-- EVENTS MODULE - Календарь и события
-- ============================================================================

-- Основная таблица событий
CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('meeting', 'deadline', 'task', 'reminder', 'holiday', 'personal')),
    status VARCHAR(50) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'ongoing', 'completed', 'cancelled')),

    -- Время
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    all_day BOOLEAN NOT NULL DEFAULT false,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Europe/Moscow',

    -- Местоположение
    location VARCHAR(500),

    -- Организатор
    organizer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Повторяющиеся события
    is_recurring BOOLEAN NOT NULL DEFAULT false,
    recurrence_rule JSONB, -- {"frequency": "weekly", "interval": 1, "byDay": ["MO", "WE"], "until": "2024-12-31"}
    parent_event_id BIGINT REFERENCES events(id) ON DELETE CASCADE,
    recurrence_id VARCHAR(100), -- уникальный ID экземпляра в серии

    -- Внешний вид
    color VARCHAR(20),
    priority INT NOT NULL DEFAULT 0 CHECK (priority BETWEEN 0 AND 3),

    -- Метаданные
    metadata JSONB,
    external_id VARCHAR(255), -- для синхронизации с внешними календарями

    -- Временные метки
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Индексы для events
CREATE INDEX idx_events_organizer_id ON events(organizer_id);
CREATE INDEX idx_events_start_time ON events(start_time);
CREATE INDEX idx_events_end_time ON events(end_time);
CREATE INDEX idx_events_event_type ON events(event_type);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_is_recurring ON events(is_recurring);
CREATE INDEX idx_events_parent_event_id ON events(parent_event_id);
CREATE INDEX idx_events_deleted_at ON events(deleted_at);
CREATE INDEX idx_events_date_range ON events(start_time, end_time);

-- Участники событий
CREATE TABLE IF NOT EXISTS event_participants (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    response_status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (response_status IN ('pending', 'accepted', 'declined', 'tentative')),
    role VARCHAR(50) NOT NULL DEFAULT 'required' CHECK (role IN ('required', 'optional', 'organizer')),
    notified_at TIMESTAMP,
    responded_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, user_id)
);

-- Индексы для event_participants
CREATE INDEX idx_event_participants_event_id ON event_participants(event_id);
CREATE INDEX idx_event_participants_user_id ON event_participants(user_id);
CREATE INDEX idx_event_participants_response_status ON event_participants(response_status);

-- Напоминания о событиях
CREATE TABLE IF NOT EXISTS event_reminders (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reminder_type VARCHAR(50) NOT NULL DEFAULT 'in_app' CHECK (reminder_type IN ('email', 'push', 'in_app')),
    minutes_before INT NOT NULL DEFAULT 15,
    is_sent BOOLEAN NOT NULL DEFAULT false,
    sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для event_reminders
CREATE INDEX idx_event_reminders_event_id ON event_reminders(event_id);
CREATE INDEX idx_event_reminders_user_id ON event_reminders(user_id);
CREATE INDEX idx_event_reminders_is_sent ON event_reminders(is_sent);

-- Исключения из повторяющихся событий
CREATE TABLE IF NOT EXISTS event_recurrence_exceptions (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    exception_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, exception_date)
);

-- Индексы для event_recurrence_exceptions
CREATE INDEX idx_event_recurrence_exceptions_event_id ON event_recurrence_exceptions(event_id);
CREATE INDEX idx_event_recurrence_exceptions_date ON event_recurrence_exceptions(exception_date);
