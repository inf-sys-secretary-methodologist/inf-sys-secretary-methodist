-- ============================================================================
-- LESSON_SLOTS — каталог пар (сетка звонков)
-- ============================================================================
-- Initiative: #139 (автогенерация расписания), Slice 1.
--
-- Дискретная сетка учебного дня: пара №1 08:30-10:00, №2 10:10-11:40 и т.д.
-- Автогенератор расписания (CSP-солвер) ставит занятия в эти фиксированные
-- слоты — домен CSP = {день недели × номер пары × аудитория}. Справочник
-- общий для всего учреждения, редактируется администратором/методистом.
--
-- Время хранится строкой "HH:MM" (VARCHAR(5)), а не PG TIME, чтобы формат
-- один-в-один совпадал с доменной сущностью LessonSlot (PG TIME отдал бы
-- "HH:MM:SS"). Строковое сравнение для zero-padded HH:MM монотонно, поэтому
-- CHECK (time_end > time_start) корректно проверяет порядок.
-- ============================================================================

CREATE TABLE IF NOT EXISTS lesson_slots (
    id         BIGSERIAL   PRIMARY KEY,
    number     INT         NOT NULL UNIQUE CHECK (number > 0),
    time_start VARCHAR(5)  NOT NULL CHECK (time_start ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
    time_end   VARCHAR(5)  NOT NULL CHECK (time_end   ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT lesson_slots_time_order CHECK (time_end > time_start)
);

-- Дефолтная сетка из 6 пар по 90 минут с переменами (можно править в UI).
INSERT INTO lesson_slots (number, time_start, time_end) VALUES
    (1, '08:30', '10:00'),
    (2, '10:10', '11:40'),
    (3, '12:20', '13:50'),
    (4, '14:00', '15:30'),
    (5, '15:40', '17:10'),
    (6, '17:20', '18:50')
ON CONFLICT (number) DO NOTHING;
