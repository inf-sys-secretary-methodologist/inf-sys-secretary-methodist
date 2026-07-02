-- ============================================================================
-- CALENDAR_FEED_TOKENS — секретные токены подписки на iCalendar-фид
-- ============================================================================
-- Initiative: #40 (интеграция с внешними календарями), PR2.
--
-- Каждый пользователь подписывается на своё расписание/события из внешнего
-- календаря (Google/Outlook/Apple) по секретному URL вида
--   /api/public/calendar/{token}/feed.ics
-- Токен — непрозрачная случайная строка (256 бит, 64 hex-символа). Он играет
-- роль пароля к read-only фиду, поэтому у пользователя ровно один активный
-- токен (UNIQUE user_id): ротация заменяет строку и инвалидирует старый URL.
--
-- Токен хранится в открытом виде (в отличие от паролей): фид отдаёт только те
-- данные расписания, которые пользователь и так вправе видеть, а URL нужно
-- показывать в интерфейсе многократно. ON DELETE CASCADE: удаление
-- пользователя уносит его токен.
-- ============================================================================

CREATE TABLE IF NOT EXISTS calendar_feed_tokens (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    BIGINT       NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(128) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT calendar_feed_tokens_token_nonempty
        CHECK (length(btrim(token)) > 0)
);
