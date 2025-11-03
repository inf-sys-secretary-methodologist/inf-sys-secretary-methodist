-- ============================================================================
-- SCHEDULE MODULE - Управление расписанием
-- ============================================================================

-- Учебные года
CREATE TABLE IF NOT EXISTS academic_years (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE, -- "2023/2024"
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_academic_years_is_active ON academic_years(is_active);

-- Семестры
CREATE TABLE IF NOT EXISTS semesters (
    id BIGSERIAL PRIMARY KEY,
    academic_year_id BIGINT NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL, -- "Осенний семестр", "Весенний семестр"
    number INT NOT NULL CHECK (number IN (1, 2)),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(academic_year_id, number)
);

CREATE INDEX idx_semesters_academic_year_id ON semesters(academic_year_id);
CREATE INDEX idx_semesters_is_active ON semesters(is_active);

-- Факультеты/Институты
CREATE TABLE IF NOT EXISTS faculties (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    short_name VARCHAR(50),
    code VARCHAR(50) UNIQUE,
    dean_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_faculties_dean_id ON faculties(dean_id);

-- Кафедры
CREATE TABLE IF NOT EXISTS departments (
    id BIGSERIAL PRIMARY KEY,
    faculty_id BIGINT NOT NULL REFERENCES faculties(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    short_name VARCHAR(50),
    code VARCHAR(50),
    head_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(faculty_id, code)
);

CREATE INDEX idx_departments_faculty_id ON departments(faculty_id);
CREATE INDEX idx_departments_head_id ON departments(head_id);

-- Специальности/Направления подготовки
CREATE TABLE IF NOT EXISTS specialties (
    id BIGSERIAL PRIMARY KEY,
    faculty_id BIGINT NOT NULL REFERENCES faculties(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL, -- "38.03.01"
    qualification VARCHAR(100), -- "Бакалавр", "Магистр"
    duration_years INT NOT NULL, -- продолжительность обучения
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_specialties_faculty_id ON specialties(faculty_id);
CREATE INDEX idx_specialties_code ON specialties(code);

-- Группы студентов
CREATE TABLE IF NOT EXISTS student_groups (
    id BIGSERIAL PRIMARY KEY,
    specialty_id BIGINT NOT NULL REFERENCES specialties(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL UNIQUE, -- "ЭК-201"
    course INT NOT NULL CHECK (course BETWEEN 1 AND 6),
    curator_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    capacity INT NOT NULL DEFAULT 25,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_student_groups_specialty_id ON student_groups(specialty_id);
CREATE INDEX idx_student_groups_curator_id ON student_groups(curator_id);
CREATE INDEX idx_student_groups_course ON student_groups(course);

-- Дисциплины
CREATE TABLE IF NOT EXISTS disciplines (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50),
    department_id BIGINT REFERENCES departments(id) ON DELETE SET NULL,
    credits INT, -- зачетные единицы
    hours_total INT, -- всего часов
    hours_lectures INT,
    hours_practice INT,
    hours_labs INT,
    control_type VARCHAR(50) CHECK (control_type IN ('exam', 'test', 'coursework', 'practice')),
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_disciplines_department_id ON disciplines(department_id);
CREATE INDEX idx_disciplines_code ON disciplines(code);

-- Учебные планы
CREATE TABLE IF NOT EXISTS curriculum_plans (
    id BIGSERIAL PRIMARY KEY,
    specialty_id BIGINT NOT NULL REFERENCES specialties(id) ON DELETE CASCADE,
    academic_year_id BIGINT NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    approved_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    approved_at DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_curriculum_plans_specialty_id ON curriculum_plans(specialty_id);
CREATE INDEX idx_curriculum_plans_academic_year_id ON curriculum_plans(academic_year_id);

-- Дисциплины в учебном плане
CREATE TABLE IF NOT EXISTS curriculum_disciplines (
    id BIGSERIAL PRIMARY KEY,
    curriculum_plan_id BIGINT NOT NULL REFERENCES curriculum_plans(id) ON DELETE CASCADE,
    discipline_id BIGINT NOT NULL REFERENCES disciplines(id) ON DELETE CASCADE,
    semester_number INT NOT NULL CHECK (semester_number BETWEEN 1 AND 12),
    hours_per_week INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(curriculum_plan_id, discipline_id, semester_number)
);

CREATE INDEX idx_curriculum_disciplines_plan_id ON curriculum_disciplines(curriculum_plan_id);
CREATE INDEX idx_curriculum_disciplines_discipline_id ON curriculum_disciplines(discipline_id);

-- Аудитории
CREATE TABLE IF NOT EXISTS classrooms (
    id BIGSERIAL PRIMARY KEY,
    building VARCHAR(100) NOT NULL,
    number VARCHAR(50) NOT NULL,
    name VARCHAR(255), -- полное название
    capacity INT NOT NULL,
    type VARCHAR(50) CHECK (type IN ('lecture', 'practice', 'lab', 'computer', 'conference', 'sports')),
    equipment JSONB, -- оборудование (проектор, компьютеры и т.д.)
    is_available BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(building, number)
);

CREATE INDEX idx_classrooms_type ON classrooms(type);
CREATE INDEX idx_classrooms_capacity ON classrooms(capacity);
CREATE INDEX idx_classrooms_is_available ON classrooms(is_available);

-- Типы занятий
CREATE TABLE IF NOT EXISTS lesson_types (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE, -- "Лекция", "Практика", "Лабораторная"
    short_name VARCHAR(20) NOT NULL,
    color VARCHAR(7), -- для UI
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Расписание занятий
CREATE TABLE IF NOT EXISTS schedule_lessons (
    id BIGSERIAL PRIMARY KEY,
    semester_id BIGINT NOT NULL REFERENCES semesters(id) ON DELETE CASCADE,
    discipline_id BIGINT NOT NULL REFERENCES disciplines(id) ON DELETE CASCADE,
    lesson_type_id BIGINT NOT NULL REFERENCES lesson_types(id) ON DELETE RESTRICT,
    teacher_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    group_id BIGINT NOT NULL REFERENCES student_groups(id) ON DELETE CASCADE,
    classroom_id BIGINT NOT NULL REFERENCES classrooms(id) ON DELETE RESTRICT,

    -- Время
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 1 AND 7), -- 1=понедельник
    time_start TIME NOT NULL,
    time_end TIME NOT NULL,
    week_type VARCHAR(20) CHECK (week_type IN ('all', 'odd', 'even')), -- все недели, четные, нечетные

    -- Даты действия
    date_start DATE NOT NULL,
    date_end DATE NOT NULL,

    -- Дополнительная информация
    notes TEXT,
    is_cancelled BOOLEAN NOT NULL DEFAULT false,
    cancellation_reason TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для быстрого поиска занятий
CREATE INDEX idx_schedule_lessons_semester_id ON schedule_lessons(semester_id);
CREATE INDEX idx_schedule_lessons_discipline_id ON schedule_lessons(discipline_id);
CREATE INDEX idx_schedule_lessons_teacher_id ON schedule_lessons(teacher_id);
CREATE INDEX idx_schedule_lessons_group_id ON schedule_lessons(group_id);
CREATE INDEX idx_schedule_lessons_classroom_id ON schedule_lessons(classroom_id);
CREATE INDEX idx_schedule_lessons_day_of_week ON schedule_lessons(day_of_week);
CREATE INDEX idx_schedule_lessons_time_start ON schedule_lessons(time_start);
CREATE INDEX idx_schedule_lessons_date_range ON schedule_lessons(date_start, date_end);

-- Изменения в расписании (отмены, переносы)
CREATE TABLE IF NOT EXISTS schedule_changes (
    id BIGSERIAL PRIMARY KEY,
    lesson_id BIGINT NOT NULL REFERENCES schedule_lessons(id) ON DELETE CASCADE,
    change_type VARCHAR(50) NOT NULL CHECK (change_type IN ('cancelled', 'moved', 'replaced_teacher', 'replaced_classroom')),
    original_date DATE NOT NULL,
    new_date DATE,
    new_classroom_id BIGINT REFERENCES classrooms(id) ON DELETE SET NULL,
    new_teacher_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reason TEXT,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_schedule_changes_lesson_id ON schedule_changes(lesson_id);
CREATE INDEX idx_schedule_changes_original_date ON schedule_changes(original_date);

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- Типы занятий
INSERT INTO lesson_types (name, short_name, color) VALUES
    ('Лекция', 'Лек', '#3B82F6'),
    ('Практическое занятие', 'Практ', '#10B981'),
    ('Лабораторная работа', 'Лаб', '#F59E0B'),
    ('Семинар', 'Сем', '#8B5CF6'),
    ('Консультация', 'Конс', '#6B7280'),
    ('Экзамен', 'Экз', '#EF4444'),
    ('Зачет', 'Зач', '#EC4899')
ON CONFLICT (name) DO NOTHING;
