-- ============================================================================
-- USERS MODULE - Организационная структура и профили пользователей
-- ============================================================================

-- Организационные подразделения (отличаются от академических кафедр в schedule module)
CREATE TABLE IF NOT EXISTS org_departments (
    id BIGSERIAL PRIMARY KEY,

    -- Основная информация
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) NOT NULL UNIQUE,
    description VARCHAR(500),

    -- Иерархия
    parent_id BIGINT REFERENCES org_departments(id) ON DELETE SET NULL,

    -- Руководитель подразделения
    head_id BIGINT REFERENCES users(id) ON DELETE SET NULL,

    -- Статус
    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для org_departments
CREATE INDEX idx_org_departments_code ON org_departments(code);
CREATE INDEX idx_org_departments_parent_id ON org_departments(parent_id);
CREATE INDEX idx_org_departments_head_id ON org_departments(head_id);
CREATE INDEX idx_org_departments_is_active ON org_departments(is_active) WHERE is_active = true;
CREATE INDEX idx_org_departments_name ON org_departments(name);

-- Должности в организации
CREATE TABLE IF NOT EXISTS positions (
    id BIGSERIAL PRIMARY KEY,

    -- Основная информация
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) NOT NULL UNIQUE,
    description VARCHAR(500),

    -- Уровень должности (для иерархии и сортировки)
    level INT NOT NULL DEFAULT 0 CHECK (level >= 0 AND level <= 100),

    -- Статус
    is_active BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для positions
CREATE INDEX idx_positions_code ON positions(code);
CREATE INDEX idx_positions_level ON positions(level);
CREATE INDEX idx_positions_is_active ON positions(is_active) WHERE is_active = true;
CREATE INDEX idx_positions_name ON positions(name);

-- Расширенные профили пользователей
CREATE TABLE IF NOT EXISTS user_profiles (
    id BIGSERIAL PRIMARY KEY,

    -- Связь с пользователем (1:1)
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- Организационная привязка
    department_id BIGINT REFERENCES org_departments(id) ON DELETE SET NULL,
    position_id BIGINT REFERENCES positions(id) ON DELETE SET NULL,

    -- Контактная информация
    phone VARCHAR(20),
    avatar VARCHAR(500),
    bio VARCHAR(500),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для user_profiles
CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_user_profiles_department_id ON user_profiles(department_id);
CREATE INDEX idx_user_profiles_position_id ON user_profiles(position_id);

-- Базовые должности
INSERT INTO positions (name, code, description, level) VALUES
    ('Системный администратор', 'SYSADMIN', 'Администратор информационной системы', 100),
    ('Методист', 'METHODIST', 'Специалист по методической работе', 80),
    ('Учёный секретарь', 'ACADSEC', 'Ученый секретарь организации', 90),
    ('Преподаватель', 'TEACHER', 'Преподаватель учебных дисциплин', 50),
    ('Студент', 'STUDENT', 'Обучающийся', 10)
ON CONFLICT (code) DO NOTHING;

-- Базовые организационные подразделения
INSERT INTO org_departments (name, code, description) VALUES
    ('Администрация', 'ADMIN', 'Административное подразделение'),
    ('Учебный отдел', 'EDU', 'Отдел учебной работы'),
    ('Методический отдел', 'METHOD', 'Отдел методической работы'),
    ('IT отдел', 'IT', 'Отдел информационных технологий')
ON CONFLICT (code) DO NOTHING;

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_users_module_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_org_departments_updated_at
    BEFORE UPDATE ON org_departments
    FOR EACH ROW
    EXECUTE FUNCTION update_users_module_updated_at();

CREATE TRIGGER trigger_positions_updated_at
    BEFORE UPDATE ON positions
    FOR EACH ROW
    EXECUTE FUNCTION update_users_module_updated_at();

CREATE TRIGGER trigger_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_users_module_updated_at();

-- Комментарии к таблицам
COMMENT ON TABLE org_departments IS 'Организационные подразделения с поддержкой иерархии (отличается от академических кафедр)';
COMMENT ON TABLE positions IS 'Должности в организации';
COMMENT ON TABLE user_profiles IS 'Расширенные профили пользователей с организационной информацией';
