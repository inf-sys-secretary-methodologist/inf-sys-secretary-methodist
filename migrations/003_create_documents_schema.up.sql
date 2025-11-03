-- ============================================================================
-- DOCUMENTS MODULE - Управление документами с полным жизненным циклом
-- ============================================================================

-- Типы документов (служебные записки, приказы, письма, протоколы и т.д.)
CREATE TABLE IF NOT EXISTS document_types (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL UNIQUE, -- служебная_записка, приказ, письмо, протокол и т.д.
    description TEXT,
    template_path VARCHAR(500), -- путь к шаблону документа
    requires_approval BOOLEAN NOT NULL DEFAULT false, -- требует ли согласования
    requires_registration BOOLEAN NOT NULL DEFAULT true, -- требует ли регистрации
    numbering_pattern VARCHAR(100), -- паттерн нумерации (например: "№{seq}/{year}-СЗ")
    retention_period INT, -- срок хранения в месяцах
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Категории документов (для группировки)
CREATE TABLE IF NOT EXISTS document_categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id BIGINT REFERENCES document_categories(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_document_categories_parent_id ON document_categories(parent_id);

-- Основная таблица документов
CREATE TABLE IF NOT EXISTS documents (
    id BIGSERIAL PRIMARY KEY,
    document_type_id BIGINT NOT NULL REFERENCES document_types(id) ON DELETE RESTRICT,
    category_id BIGINT REFERENCES document_categories(id) ON DELETE SET NULL,

    -- Регистрационные данные
    registration_number VARCHAR(100), -- регистрационный номер
    registration_date DATE, -- дата регистрации

    -- Основная информация
    title VARCHAR(500) NOT NULL, -- заголовок документа
    subject TEXT, -- тема/о чем (для служебных записок - "О приобретении монитора")
    content TEXT, -- текст документа

    -- Реквизиты автора
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    author_department VARCHAR(255), -- подразделение автора
    author_position VARCHAR(255), -- должность автора

    -- Реквизиты адресата (для служебных записок, писем)
    recipient_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    recipient_department VARCHAR(255),
    recipient_position VARCHAR(255),
    recipient_external TEXT, -- для внешних адресатов

    -- Статус и workflow
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN (
        'draft',        -- черновик
        'registered',   -- зарегистрирован
        'routing',      -- на маршрутизации
        'approval',     -- на согласовании
        'approved',     -- согласован
        'rejected',     -- отклонен
        'execution',    -- на исполнении
        'executed',     -- исполнен
        'archived'      -- архивирован
    )),

    -- Файлы
    file_name VARCHAR(500),
    file_path VARCHAR(1000),
    file_size BIGINT,
    mime_type VARCHAR(100),

    -- Версионирование
    version INT NOT NULL DEFAULT 1,
    parent_document_id BIGINT REFERENCES documents(id) ON DELETE SET NULL, -- для версий

    -- Сроки
    deadline DATE, -- срок исполнения
    execution_date DATE, -- дата исполнения

    -- Метаданные
    metadata JSONB, -- дополнительные поля (зависят от типа документа)
    is_public BOOLEAN NOT NULL DEFAULT false,
    importance VARCHAR(50) DEFAULT 'normal' CHECK (importance IN ('low', 'normal', 'high', 'urgent')),

    -- Аудит
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP -- soft delete
);

-- Индексы для documents
CREATE INDEX idx_documents_document_type_id ON documents(document_type_id);
CREATE INDEX idx_documents_category_id ON documents(category_id);
CREATE INDEX idx_documents_author_id ON documents(author_id);
CREATE INDEX idx_documents_recipient_id ON documents(recipient_id);
CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_documents_registration_number ON documents(registration_number);
CREATE INDEX idx_documents_registration_date ON documents(registration_date DESC);
CREATE INDEX idx_documents_deadline ON documents(deadline);
CREATE INDEX idx_documents_created_at ON documents(created_at DESC);
CREATE INDEX idx_documents_metadata ON documents USING GIN (metadata);
CREATE INDEX idx_documents_deleted_at ON documents(deleted_at) WHERE deleted_at IS NULL;

-- История версий документа
CREATE TABLE IF NOT EXISTS document_versions (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    version INT NOT NULL,
    content TEXT,
    file_name VARCHAR(500),
    file_path VARCHAR(1000),
    file_size BIGINT,
    changed_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    change_description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(document_id, version)
);

CREATE INDEX idx_document_versions_document_id ON document_versions(document_id);

-- Workflow - маршруты согласования
CREATE TABLE IF NOT EXISTS document_routes (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    step_number INT NOT NULL, -- порядковый номер шага
    approver_id BIGINT REFERENCES users(id) ON DELETE SET NULL, -- согласующий
    approver_role VARCHAR(50), -- или роль (если не конкретный пользователь)
    action VARCHAR(50) NOT NULL CHECK (action IN ('approve', 'review', 'sign', 'notify')),
    is_parallel BOOLEAN NOT NULL DEFAULT false, -- параллельное или последовательное согласование
    is_required BOOLEAN NOT NULL DEFAULT true,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'skipped')),
    comment TEXT,
    processed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_document_routes_document_id ON document_routes(document_id);
CREATE INDEX idx_document_routes_approver_id ON document_routes(approver_id);
CREATE INDEX idx_document_routes_status ON document_routes(status);

-- Права доступа к документу
CREATE TABLE IF NOT EXISTS document_permissions (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) CHECK (role IN ('admin', 'secretary', 'methodist', 'teacher', 'student')),
    permission VARCHAR(50) NOT NULL CHECK (permission IN ('read', 'write', 'delete', 'admin')),
    granted_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP, -- срок действия прав
    CONSTRAINT document_permissions_user_or_role CHECK (user_id IS NOT NULL OR role IS NOT NULL)
);

CREATE INDEX idx_document_permissions_document_id ON document_permissions(document_id);
CREATE INDEX idx_document_permissions_user_id ON document_permissions(user_id);
CREATE INDEX idx_document_permissions_role ON document_permissions(role);

-- Связи между документами (ответы, приложения, ссылки)
CREATE TABLE IF NOT EXISTS document_relations (
    id BIGSERIAL PRIMARY KEY,
    source_document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    target_document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    relation_type VARCHAR(50) NOT NULL CHECK (relation_type IN (
        'reply',        -- ответ на документ
        'attachment',   -- приложение к документу
        'reference',    -- ссылка на документ
        'supersedes',   -- заменяет документ
        'amendment'     -- дополнение к документу
    )),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT document_relations_not_self CHECK (source_document_id != target_document_id)
);

CREATE INDEX idx_document_relations_source ON document_relations(source_document_id);
CREATE INDEX idx_document_relations_target ON document_relations(target_document_id);

-- Теги для документов
CREATE TABLE IF NOT EXISTS document_tags (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    color VARCHAR(7), -- hex color для UI
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS document_tag_relations (
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tag_id BIGINT NOT NULL REFERENCES document_tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (document_id, tag_id)
);

CREATE INDEX idx_document_tag_relations_document_id ON document_tag_relations(document_id);
CREATE INDEX idx_document_tag_relations_tag_id ON document_tag_relations(tag_id);

-- История действий с документом (audit log)
CREATE TABLE IF NOT EXISTS document_history (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL, -- created, updated, registered, approved, rejected, etc.
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_document_history_document_id ON document_history(document_id);
CREATE INDEX idx_document_history_user_id ON document_history(user_id);
CREATE INDEX idx_document_history_action ON document_history(action);
CREATE INDEX idx_document_history_created_at ON document_history(created_at DESC);

-- ============================================================================
-- SEED DATA - Начальные данные
-- ============================================================================

-- Типы документов
INSERT INTO document_types (name, code, description, requires_approval, numbering_pattern) VALUES
    ('Служебная записка', 'memo', 'Информационно-справочный документ для внутреннего обмена', false, '№{seq}/{year}-СЗ'),
    ('Приказ по основной деятельности', 'order_main', 'Распорядительный документ по основным вопросам', true, '№{seq}-к'),
    ('Приказ по личному составу', 'order_hr', 'Распорядительный документ по кадровым вопросам', true, '№{seq}-л/с'),
    ('Приказ по АХД', 'order_admin', 'Распорядительный документ по административно-хозяйственной деятельности', true, '№{seq}-ахд'),
    ('Распоряжение', 'directive', 'Правовой акт по оперативным вопросам', false, '№{seq}-р'),
    ('Деловое письмо', 'business_letter', 'Официальная переписка', false, '№{seq}'),
    ('Протокол заседания', 'protocol', 'Документ для фиксации решений коллегиальных органов', false, '№{seq}'),
    ('Договор', 'contract', 'Юридический документ', true, '№{seq}'),
    ('Должностная инструкция', 'job_instruction', 'Регламент обязанностей сотрудника', true, NULL)
ON CONFLICT (code) DO NOTHING;

-- Категории документов
INSERT INTO document_categories (name, description) VALUES
    ('Учебная деятельность', 'Документы, связанные с учебным процессом'),
    ('Кадровые вопросы', 'Приказы и распоряжения по персоналу'),
    ('Административные', 'АХД и организационные вопросы'),
    ('Методическая работа', 'Методические материалы и рекомендации'),
    ('Финансовые', 'Договоры, сметы, финансовые документы'),
    ('Архив', 'Архивные документы')
ON CONFLICT DO NOTHING;
