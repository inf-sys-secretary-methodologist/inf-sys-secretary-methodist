-- ============================================================================
-- REPORTS MODULE - Отчетность и аналитика
-- ============================================================================

-- Типы отчетов
CREATE TABLE IF NOT EXISTS report_types (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    category VARCHAR(100), -- "academic", "administrative", "financial", "methodical"
    template_path VARCHAR(500), -- путь к шаблону отчета
    output_format VARCHAR(50) DEFAULT 'pdf' CHECK (output_format IN ('pdf', 'xlsx', 'docx', 'html')),
    is_periodic BOOLEAN NOT NULL DEFAULT false, -- периодический ли отчет
    period_type VARCHAR(50) CHECK (period_type IN ('daily', 'weekly', 'monthly', 'quarterly', 'annual')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Параметры отчетов (для динамической генерации)
CREATE TABLE IF NOT EXISTS report_parameters (
    id BIGSERIAL PRIMARY KEY,
    report_type_id BIGINT NOT NULL REFERENCES report_types(id) ON DELETE CASCADE,
    parameter_name VARCHAR(100) NOT NULL,
    parameter_type VARCHAR(50) NOT NULL CHECK (parameter_type IN ('string', 'number', 'date', 'boolean', 'select', 'multiselect')),
    is_required BOOLEAN NOT NULL DEFAULT true,
    default_value TEXT,
    options JSONB, -- для select/multiselect типов
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_report_parameters_report_type_id ON report_parameters(report_type_id);

-- Отчеты
CREATE TABLE IF NOT EXISTS reports (
    id BIGSERIAL PRIMARY KEY,
    report_type_id BIGINT NOT NULL REFERENCES report_types(id) ON DELETE RESTRICT,

    -- Основная информация
    title VARCHAR(500) NOT NULL,
    description TEXT,

    -- Период отчета
    period_start DATE,
    period_end DATE,

    -- Автор
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- Статус
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN (
        'draft',        -- черновик
        'generating',   -- генерируется
        'ready',        -- готов
        'reviewing',    -- на проверке
        'approved',     -- утвержден
        'rejected',     -- отклонен
        'published'     -- опубликован
    )),

    -- Файлы
    file_name VARCHAR(500),
    file_path VARCHAR(1000),
    file_size BIGINT,
    mime_type VARCHAR(100),

    -- Параметры, с которыми был сгенерирован отчет
    parameters JSONB,

    -- Данные отчета (для хранения агрегированных метрик)
    data JSONB,

    -- Комментарии и замечания
    reviewer_comment TEXT,
    reviewed_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP,

    -- Публикация
    published_at TIMESTAMP,
    is_public BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для reports
CREATE INDEX idx_reports_report_type_id ON reports(report_type_id);
CREATE INDEX idx_reports_author_id ON reports(author_id);
CREATE INDEX idx_reports_status ON reports(status);
CREATE INDEX idx_reports_period ON reports(period_start, period_end);
CREATE INDEX idx_reports_created_at ON reports(created_at DESC);
CREATE INDEX idx_reports_data ON reports USING GIN (data);

-- Доступ к отчетам
CREATE TABLE IF NOT EXISTS report_access (
    id BIGSERIAL PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) CHECK (role IN ('admin', 'secretary', 'methodist', 'teacher', 'student')),
    permission VARCHAR(50) NOT NULL CHECK (permission IN ('read', 'write', 'approve', 'publish')),
    granted_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT report_access_user_or_role CHECK (user_id IS NOT NULL OR role IS NOT NULL)
);

CREATE INDEX idx_report_access_report_id ON report_access(report_id);
CREATE INDEX idx_report_access_user_id ON report_access(user_id);
CREATE INDEX idx_report_access_role ON report_access(role);

-- Комментарии к отчетам
CREATE TABLE IF NOT EXISTS report_comments (
    id BIGSERIAL PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_report_comments_report_id ON report_comments(report_id);
CREATE INDEX idx_report_comments_author_id ON report_comments(author_id);

-- История генерации отчетов
CREATE TABLE IF NOT EXISTS report_generation_log (
    id BIGSERIAL PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    status VARCHAR(50) NOT NULL CHECK (status IN ('started', 'completed', 'failed')),
    error_message TEXT,
    duration_seconds INT,
    records_processed INT
);

CREATE INDEX idx_report_generation_log_report_id ON report_generation_log(report_id);
CREATE INDEX idx_report_generation_log_status ON report_generation_log(status);

-- Подписки на отчеты (для автоматической рассылки)
CREATE TABLE IF NOT EXISTS report_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    report_type_id BIGINT NOT NULL REFERENCES report_types(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    delivery_method VARCHAR(50) NOT NULL DEFAULT 'email' CHECK (delivery_method IN ('email', 'notification', 'both')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(report_type_id, user_id)
);

CREATE INDEX idx_report_subscriptions_user_id ON report_subscriptions(user_id);
CREATE INDEX idx_report_subscriptions_is_active ON report_subscriptions(is_active);

-- Шаблоны отчетов
CREATE TABLE IF NOT EXISTS report_templates (
    id BIGSERIAL PRIMARY KEY,
    report_type_id BIGINT NOT NULL REFERENCES report_types(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    content TEXT NOT NULL, -- HTML/LaTeX шаблон
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_report_templates_report_type_id ON report_templates(report_type_id);
CREATE INDEX idx_report_templates_is_default ON report_templates(is_default);

-- История изменений отчета
CREATE TABLE IF NOT EXISTS report_history (
    id BIGSERIAL PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL, -- created, updated, approved, rejected, published
    details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_report_history_report_id ON report_history(report_id);
CREATE INDEX idx_report_history_user_id ON report_history(user_id);
CREATE INDEX idx_report_history_created_at ON report_history(created_at DESC);

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- Типы отчетов
INSERT INTO report_types (name, code, description, category, is_periodic, period_type, output_format) VALUES
    ('Отчет об успеваемости студентов', 'student_performance', 'Анализ успеваемости по группам и специальностям', 'academic', true, 'monthly', 'xlsx'),
    ('Отчет о посещаемости', 'attendance', 'Статистика посещаемости занятий', 'academic', true, 'weekly', 'pdf'),
    ('Нагрузка преподавателей', 'teacher_workload', 'Распределение учебной нагрузки', 'academic', true, 'monthly', 'xlsx'),
    ('Загрузка аудиторий', 'classroom_utilization', 'Использование аудиторного фонда', 'administrative', true, 'monthly', 'pdf'),
    ('Методическая работа кафедры', 'department_methodical_work', 'Отчет о методической деятельности', 'methodical', true, 'quarterly', 'docx'),
    ('Выполнение учебного плана', 'curriculum_execution', 'Контроль выполнения учебных планов', 'academic', true, 'annual', 'pdf'),
    ('Отчет о документообороте', 'document_flow', 'Статистика документооборота', 'administrative', true, 'monthly', 'xlsx'),
    ('Сводный отчет по задачам', 'tasks_summary', 'Выполнение поручений и задач', 'administrative', true, 'weekly', 'html')
ON CONFLICT (code) DO NOTHING;
