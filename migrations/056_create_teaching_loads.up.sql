-- ============================================================================
-- TEACHING_LOADS — планируемая учебная нагрузка
-- ============================================================================
-- Initiative: #139 (автогенерация расписания), Slice 2.
--
-- Строка нагрузки = «группа изучает дисциплину у преподавателя N пар в неделю
-- (все/нечётные/чётные недели), занятие такого-то типа». Это источник истины
-- для автогенератора: он разворачивает каждую строку в N расставляемых пар.
-- Ведёт методист. Уникальность (semester, group, discipline, lesson_type)
-- не даёт задвоить одну и ту же нагрузку в семестре.
-- ============================================================================

CREATE TABLE IF NOT EXISTS teaching_loads (
    id             BIGSERIAL   PRIMARY KEY,
    semester_id    BIGINT      NOT NULL REFERENCES semesters(id) ON DELETE CASCADE,
    group_id       BIGINT      NOT NULL REFERENCES student_groups(id) ON DELETE CASCADE,
    discipline_id  BIGINT      NOT NULL REFERENCES disciplines(id) ON DELETE CASCADE,
    teacher_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    lesson_type_id BIGINT      NOT NULL REFERENCES lesson_types(id) ON DELETE RESTRICT,
    pairs_per_week INT         NOT NULL CHECK (pairs_per_week > 0 AND pairs_per_week <= 20),
    week_type      VARCHAR(20) NOT NULL CHECK (week_type IN ('all', 'odd', 'even')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT teaching_loads_unique UNIQUE (semester_id, group_id, discipline_id, lesson_type_id)
);

CREATE INDEX IF NOT EXISTS idx_teaching_loads_semester ON teaching_loads (semester_id);
CREATE INDEX IF NOT EXISTS idx_teaching_loads_group ON teaching_loads (group_id);
CREATE INDEX IF NOT EXISTS idx_teaching_loads_teacher ON teaching_loads (teacher_id);
