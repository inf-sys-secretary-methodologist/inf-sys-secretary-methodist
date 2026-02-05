# 🏗️ Модульная архитектура

## 📋 Обзор архитектуры

Система построена на принципах модульного монолита с использованием Domain-Driven Design (DDD) для обеспечения масштабируемости, поддержки и возможности плавного перехода к микросервисной архитектуре.

## 🎯 Принципы архитектуры

### 1. **Domain-Driven Design (DDD)**
- Четкое разделение доменов и субдоменов
- Bounded Contexts для каждого модуля
- Ubiquitous Language в рамках каждого контекста
- Доменные сервисы, сущности и value objects

### 2. **Clean Architecture**
- Зависимости направлены от внешних слоев к внутренним
- Бизнес-логика независима от фреймворков и БД
- Инверсия зависимостей через интерфейсы
- Тестируемость на всех уровнях

### 3. **Модульность**
- Слабая связанность между модулями
- Высокая связность внутри модулей
- Четкие интерфейсы между модулями
- Независимое развитие и тестирование модулей

## 🧩 Модульная структура

> **Статус реализации**: На данный момент реализованы 12 модулей из 14. Остальные существуют как placeholder'ы для будущей разработки.

| Модуль | Статус | Файлов | Описание |
|--------|--------|--------|----------|
| auth | ✅ Реализован | 18 | JWT аутентификация, регистрация, роли |
| documents | ✅ Реализован | 30 | CRUD документов, версии, шаринг, теги |
| notifications | ✅ Реализован | 27 | In-app, Email (Composio), Telegram |
| schedule | ✅ Реализован | 7 | Расписание занятий |
| users | ✅ Реализован | 20 | CRUD пользователей, профили, аватары |
| files | ✅ Реализован | 12 | MinIO/S3 хранилище файлов |
| messaging | ✅ Реализован | 15 | Внутренняя переписка, групповые чаты, WebSocket |
| workflow | 📋 Планируется | 0 | Согласование документов |
| tasks | 📋 Планируется | 0 | Задания для студентов |
| reporting | ✅ Реализован | 15 | Пользовательские отчёты, экспорт в PDF/Excel/CSV |
| integration | ✅ Реализован | 27 | Интеграция с 1С (импорт сотрудников/студентов) |
| analytics | ✅ Реализован | 12 | Предиктивная аналитика студентов (risk score) |
| dashboard | ✅ Реализован | 8 | KPI статистика, тренды, активность |
| announcements | ✅ Реализован | 10 | Новости и объявления |

### Core Modules (Основные модули)

#### 1. **Authentication Module** 🔐 ✅ Реализован
**Bounded Context**: Identity & Access Management
```
internal/modules/auth/
├── domain/
│   ├── entities/
│   │   ├── user.go
│   │   ├── role.go
│   │   └── session.go
│   ├── repositories/
│   │   └── user_repository.go
│   ├── services/
│   │   ├── auth_service.go
│   │   └── token_service.go
│   └── events/
│       └── user_events.go
├── application/
│   ├── usecases/
│   │   ├── login_user.go
│   │   ├── register_user.go
│   │   └── refresh_token.go
│   ├── commands/
│   │   └── auth_commands.go
│   └── queries/
│       └── auth_queries.go
├── infrastructure/
│   ├── persistence/
│   │   ├── postgres/
│   │   │   └── user_repository.go
│   │   └── redis/
│   │       └── session_store.go
│   ├── external/
│   │   ├── oauth_provider.go
│   │   └── email_service.go
│   └── security/
│       ├── jwt_manager.go
│       └── password_hasher.go
└── interfaces/
    ├── http/
    │   ├── handlers/
    │   │   └── auth_handler.go
    │   ├── middleware/
    │   │   └── auth_middleware.go
    │   └── dto/
    │       └── auth_dto.go
    └── grpc/
        └── auth_service.proto
```

**Ответственности**:
- Локальная аутентификация email/password
- JWT токены (access + refresh)
- Управление сессиями через Redis
- Базовые роли (user, admin)

#### 2. **User Management Module** 👥 ✅ Реализован
**Bounded Context**: User Profile & Organization

Модуль управления пользователями:
- CRUD операции для профилей, подразделений, должностей
- Загрузка и кадрирование аватаров
- Импорт/экспорт пользователей через CSV
- Организационная структура

```
internal/modules/users/
├── domain/
│   ├── entities/
│   │   ├── user_profile.go
│   │   ├── department.go
│   │   └── position.go
│   └── repositories/
│       ├── user_profile_repository.go
│       ├── department_repository.go
│       └── position_repository.go
├── application/
│   ├── dto/
│   │   ├── user_dto.go
│   │   ├── department_dto.go
│   │   ├── position_dto.go
│   │   └── csv_dto.go
│   └── usecases/
│       ├── user_usecase.go
│       ├── department_usecase.go
│       └── position_usecase.go
├── infrastructure/
│   └── persistence/
│       ├── user_profile_repository.go
│       ├── department_repository.go
│       └── position_repository.go
└── interfaces/
    └── http/
        └── handlers/
            ├── user_handler.go
            ├── avatar_handler.go
            ├── department_handler.go
            └── position_handler.go
```

#### 3. **Document Management Module** 📄 ✅ Реализован
**Bounded Context**: Document Lifecycle & Content
```
internal/modules/documents/
├── domain/
│   ├── entities/
│   │   ├── document.go
│   │   ├── template.go
│   │   ├── version.go
│   │   └── metadata.go
│   ├── repositories/
│   │   ├── document_repository.go
│   │   └── template_repository.go
│   ├── services/
│   │   ├── document_service.go
│   │   ├── versioning_service.go
│   │   └── search_service.go
│   └── value_objects/
│       ├── document_type.go
│       └── status.go
├── application/
│   ├── usecases/
│   │   ├── create_document.go
│   │   ├── approve_document.go
│   │   ├── search_documents.go
│   │   └── archive_document.go
│   ├── commands/
│   │   └── document_commands.go
│   └── queries/
│       └── document_queries.go
├── infrastructure/
│   ├── persistence/
│   │   ├── postgres/
│   │   │   └── document_repository.go
│   │   └── elasticsearch/
│   │       └── search_repository.go
│   ├── storage/
│   │   ├── minio/
│   │   │   └── file_storage.go
│   │   └── local/
│   │       └── file_system.go
│   └── external/
│       └── pdf_generator.go
└── interfaces/
    ├── http/
    │   ├── handlers/
    │   └── dto/
    └── events/
        └── document_events.go
```

#### 4. **Workflow Module** 🔄 📋 Планируется
**Bounded Context**: Business Process & Approval
```
internal/modules/workflow/
├── domain/
│   ├── entities/
│   │   ├── workflow.go
│   │   ├── step.go
│   │   ├── approval.go
│   │   └── route.go
│   ├── repositories/
│   │   └── workflow_repository.go
│   ├── services/
│   │   ├── workflow_engine.go
│   │   ├── approval_service.go
│   │   └── routing_service.go
│   └── value_objects/
│       ├── approval_status.go
│       └── route_type.go
├── application/
│   ├── usecases/
│   │   ├── start_workflow.go
│   │   ├── approve_step.go
│   │   ├── reject_step.go
│   │   └── escalate_workflow.go
│   └── handlers/
│       └── workflow_handler.go
├── infrastructure/
│   ├── persistence/
│   │   └── postgres/
│   │       └── workflow_repository.go
│   └── external/
│       └── notification_client.go
└── interfaces/
    ├── http/
    │   └── handlers/
    └── events/
        └── workflow_events.go
```

### Business Modules (Бизнес-модули)

#### 5. **Schedule Module** 📅 ✅ Реализован
**Bounded Context**: Academic Planning & Resources
```
internal/modules/schedule/
├── domain/
│   ├── entities/
│   │   ├── schedule.go
│   │   ├── lesson.go
│   │   ├── room.go
│   │   ├── group.go
│   │   └── teacher.go
│   ├── repositories/
│   │   ├── schedule_repository.go
│   │   └── resource_repository.go
│   ├── services/
│   │   ├── scheduling_service.go
│   │   ├── conflict_detector.go
│   │   └── optimizer.go
│   └── value_objects/
│       ├── time_slot.go
│       └── room_capacity.go
├── application/
│   ├── usecases/
│   │   ├── create_schedule.go
│   │   ├── detect_conflicts.go
│   │   ├── optimize_schedule.go
│   │   └── export_schedule.go
│   └── algorithms/
│       ├── genetic_algorithm.go
│       └── constraint_solver.go
├── infrastructure/
│   ├── persistence/
│   │   └── postgres/
│   └── external/
│       └── calendar_export.go
└── interfaces/
    ├── http/
    └── events/
```

#### 6. **Reporting Module** 📊 ✅ Реализован
**Bounded Context**: Analytics & Business Intelligence

Модуль пользовательских отчётов (Custom Report Builder):
- Конструктор отчётов с drag-and-drop интерфейсом
- Выбор полей и фильтров из различных источников данных
- Группировка и сортировка данных
- Экспорт в PDF, Excel (XLSX), CSV
- Публичные и приватные отчёты
- Динамическое выполнение SQL запросов

**Источники данных (DataSource)**:
- `documents` - Документы
- `users` - Пользователи
- `events` - События
- `tasks` - Задачи
- `students` - Студенты

```
internal/modules/reporting/
├── domain/
│   ├── entities/
│   │   ├── custom_report.go      # Пользовательский отчёт
│   │   ├── report_field.go       # Поле отчёта
│   │   └── report.go             # Базовый отчёт
│   └── repositories/
│       ├── custom_report_repository.go
│       └── report_repository.go
├── application/
│   ├── dto/
│   │   ├── custom_report_dto.go  # DTO для пользовательских отчётов
│   │   └── report_dto.go
│   └── usecases/
│       ├── custom_report_usecase.go      # Бизнес-логика
│       ├── custom_report_usecase_test.go # Unit тесты
│       └── report_usecase.go
├── infrastructure/
│   ├── persistence/
│   │   ├── custom_report_repository_pg.go
│   │   └── report_repository_pg.go
│   └── query/
│       └── dynamic_query_builder.go  # Динамический SQL builder
└── interfaces/
    └── http/
        └── handlers/
            ├── custom_report_handler.go      # REST API
            ├── custom_report_handler_test.go # E2E тесты
            └── report_handler.go
```

**API Endpoints**:
- `GET /api/custom-reports` - Список отчётов
- `POST /api/custom-reports` - Создание отчёта
- `GET /api/custom-reports/:id` - Получение отчёта
- `PUT /api/custom-reports/:id` - Обновление отчёта
- `DELETE /api/custom-reports/:id` - Удаление отчёта
- `POST /api/custom-reports/:id/execute` - Выполнение отчёта
- `POST /api/custom-reports/:id/export` - Экспорт в PDF/Excel/CSV
- `GET /api/custom-reports/my` - Мои отчёты
- `GET /api/custom-reports/public` - Публичные отчёты
- `GET /api/custom-reports/available-fields/:dataSource` - Доступные поля

#### 7. **Task Management Module** ☑️ 📋 Планируется
**Bounded Context**: Task Tracking & Assignment
```
internal/modules/tasks/
├── domain/
│   ├── entities/
│   │   ├── task.go
│   │   ├── assignment.go
│   │   └── reminder.go
│   ├── repositories/
│   │   └── task_repository.go
│   ├── services/
│   │   ├── task_service.go
│   │   ├── assignment_service.go
│   │   └── reminder_service.go
│   └── value_objects/
│       ├── priority.go
│       └── due_date.go
├── application/
│   ├── usecases/
│   │   ├── create_task.go
│   │   ├── assign_task.go
│   │   ├── complete_task.go
│   │   └── send_reminders.go
│   └── handlers/
│       └── task_handler.go
├── infrastructure/
│   ├── persistence/
│   │   └── postgres/
│   └── external/
│       └── notification_client.go
└── interfaces/
    ├── http/
    └── cron/
        └── reminder_job.go
```

### Supporting Modules (Поддерживающие модули)

#### 8. **Notification Module** 📧 ✅ Реализован
**Bounded Context**: Communication & Alerts

Модуль поддерживает три канала уведомлений:
- **In-app**: Уведомления внутри приложения с фильтрацией по типам
- **Email**: Интеграция с Gmail через Composio API
- **Telegram**: Push-уведомления через Telegram Bot (polling/webhook)

```
internal/modules/notifications/
├── domain/
│   ├── entities/
│   │   ├── notification.go          # In-app уведомления
│   │   ├── preferences.go           # Настройки пользователя
│   │   └── telegram_verification.go # Коды верификации Telegram
│   ├── repositories/
│   │   ├── notification_repository.go
│   │   ├── preferences_repository.go
│   │   └── telegram_repository.go
│   └── services/
│       ├── email_service.go         # Interface
│       ├── telegram_service.go      # Interface
│       └── slack_service.go         # Interface (planned)
├── application/
│   ├── dto/
│   │   ├── notification_dto.go
│   │   └── preferences_dto.go
│   ├── services/
│   │   ├── composio_email_service.go
│   │   ├── composio_telegram_service.go
│   │   ├── composio_slack_service.go
│   │   └── telegram_verification_service.go
│   └── usecases/
│       ├── notification_usecase.go
│       ├── notification_usecase_test.go
│       └── preferences_usecase.go
├── infrastructure/
│   ├── persistence/
│   │   ├── notification_repository_pg.go
│   │   ├── preferences_repository_pg.go
│   │   └── telegram_repository_pg.go
│   └── scheduler/
│       └── reminder_scheduler.go
└── interfaces/
    └── http/
        ├── handlers/
        │   └── email_handler.go
        ├── notification_handler.go
        ├── preferences_handler.go
        ├── telegram_handler.go
        └── telegram_webhook_handler.go
```

**Telegram Integration**: См. [Telegram Bot Integration](../../integrations/telegram-bot.md)

#### 9. **File Storage Module** 📁 ✅ Реализован
**Bounded Context**: File Management & Processing

Модуль управления файлами с поддержкой MinIO/S3 хранилища:
- Upload/Download файлов
- Привязка файлов к документам
- Версионирование файлов
- Проверка MIME-типов и размера

```
internal/modules/files/
├── domain/
│   ├── entities/
│   │   └── file.go
│   └── repositories/
│       └── file_repository.go
├── application/
│   ├── dto/
│   │   └── file_dto.go
│   ├── services/
│   │   └── minio_storage_service.go
│   └── usecases/
│       └── file_usecase.go
├── infrastructure/
│   └── persistence/
│       └── file_repository_pg.go
└── interfaces/
    └── http/
        └── file_handler.go
```

**Старая планируемая структура (для справки):**
```
internal/modules/files/ (planned)
├── domain/
│   ├── entities/
│   │   ├── file.go
│   │   ├── folder.go
│   │   └── access_control.go
│   ├── repositories/
│   │   └── file_repository.go
│   ├── services/
│   │   ├── storage_service.go
│   │   ├── conversion_service.go
│   │   ├── preview_service.go
│   │   └── virus_scanner.go
│   └── value_objects/
│       ├── file_type.go
│       └── permissions.go
├── application/
│   ├── usecases/
│   │   ├── upload_file.go
│   │   ├── convert_file.go
│   │   ├── generate_preview.go
│   │   └── scan_file.go
│   └── handlers/
│       └── file_handler.go
├── infrastructure/
│   ├── storage/
│   │   ├── minio/
│   │   ├── s3/
│   │   └── local/
│   ├── processing/
│   │   ├── imagemagick/
│   │   ├── libreoffice/
│   │   └── clamav/
│   └── persistence/
│       └── postgres/
└── interfaces/
    ├── http/
    └── api/
        └── storage_api.go
```

#### 10. **Integration Module** 🔗 ✅ Реализован
**Bounded Context**: External System Integration

Модуль интеграции с внешними системами (1С):
- Импорт сотрудников из 1С через OData
- Импорт студентов из 1С через OData
- Логирование операций синхронизации
- Автоматическое разрешение конфликтов (source_wins, target_wins, manual)
- Frontend UI для управления интеграцией

```
internal/modules/integration/
├── domain/
│   ├── entities/
│   │   ├── external_employee.go
│   │   ├── external_student.go
│   │   ├── sync_log.go
│   │   ├── sync_status.go
│   │   └── sync_conflict.go
│   └── repositories/
│       ├── external_employee_repository.go
│       ├── external_student_repository.go
│       ├── sync_log_repository.go
│       └── sync_conflict_repository.go
├── application/
│   ├── dto/
│   │   ├── employee_dto.go
│   │   ├── student_dto.go
│   │   ├── sync_dto.go
│   │   └── conflict_dto.go
│   └── usecases/
│       ├── employee_usecase.go
│       ├── student_usecase.go
│       ├── sync_usecase.go
│       └── conflict_usecase.go
├── infrastructure/
│   ├── persistence/
│   │   ├── external_employee_repository_pg.go
│   │   ├── external_student_repository_pg.go
│   │   ├── sync_log_repository_pg.go
│   │   └── sync_conflict_repository_pg.go
│   └── odata/
│       └── client.go
├── interfaces/
│   └── http/
│       ├── employee_handler.go
│       ├── student_handler.go
│       ├── sync_handler.go
│       └── conflict_handler.go
└── module.go
```

**См. также**: [1C Integration](../../integrations/README.md#1c-integration)

## 🔄 Межмодульное взаимодействие

### 1. **Event-Driven Architecture**
```go
// Пример доменного события
type DocumentCreated struct {
    DocumentID   string    `json:"document_id"`
    AuthorID     string    `json:"author_id"`
    DocumentType string    `json:"document_type"`
    CreatedAt    time.Time `json:"created_at"`
}

// Event Bus для межмодульной коммуникации
type EventBus interface {
    Publish(event DomainEvent) error
    Subscribe(eventType string, handler EventHandler) error
}
```

### 2. **Shared Kernel**
```
internal/shared/
├── domain/
│   ├── common/
│   │   ├── aggregate_root.go
│   │   ├── entity.go
│   │   ├── value_object.go
│   │   └── domain_event.go
│   ├── errors/
│   │   ├── domain_errors.go
│   │   └── validation_errors.go
│   └── events/
│       ├── event_bus.go
│       └── event_store.go
├── infrastructure/
│   ├── database/
│   │   ├── transaction_manager.go
│   │   └── unit_of_work.go
│   ├── logging/
│   │   └── logger.go
│   ├── metrics/
│   │   └── metrics_collector.go
│   └── config/
│       └── config_manager.go
└── application/
    ├── middleware/
    │   ├── auth_middleware.go
    │   ├── logging_middleware.go
    │   └── metrics_middleware.go
    └── contracts/
        ├── repositories.go
        └── services.go
```

### 3. **API Gateway Pattern**
```
internal/gateway/
├── router/
│   ├── routes.go
│   └── middleware.go
├── handlers/
│   ├── auth_proxy.go
│   ├── document_proxy.go
│   └── user_proxy.go
├── middleware/
│   ├── rate_limiting.go
│   ├── request_validation.go
│   └── response_transformation.go
└── config/
    └── routing_config.go
```

## 🏛️ Слоистая архитектура

### 1. **Domain Layer (Доменный слой)**
```go
// Пример доменной сущности
type Document struct {
    id          DocumentID
    title       string
    content     string
    authorID    UserID
    status      DocumentStatus
    createdAt   time.Time
    updatedAt   time.Time
    domainEvents []DomainEvent
}

// Доменные методы содержат бизнес-логику
func (d *Document) Approve(approverID UserID) error {
    if d.status != DocumentStatusPending {
        return ErrDocumentNotPending
    }

    d.status = DocumentStatusApproved
    d.updatedAt = time.Now()

    // Публикуем доменное событие
    d.AddDomainEvent(DocumentApproved{
        DocumentID: d.id,
        ApproverID: approverID,
        ApprovedAt: d.updatedAt,
    })

    return nil
}
```

### 2. **Application Layer (Слой приложения)**
```go
// Use Case пример
type ApproveDocumentUseCase struct {
    documentRepo DocumentRepository
    userRepo     UserRepository
    eventBus     EventBus
    unitOfWork   UnitOfWork
}

func (uc *ApproveDocumentUseCase) Execute(cmd ApproveDocumentCommand) error {
    return uc.unitOfWork.Execute(func() error {
        // Получаем документ
        document, err := uc.documentRepo.GetByID(cmd.DocumentID)
        if err != nil {
            return err
        }

        // Проверяем права пользователя
        user, err := uc.userRepo.GetByID(cmd.ApproverID)
        if err != nil {
            return err
        }

        if !user.CanApprove(document) {
            return ErrInsufficientPermissions
        }

        // Выполняем бизнес-операцию
        if err := document.Approve(cmd.ApproverID); err != nil {
            return err
        }

        // Сохраняем изменения
        if err := uc.documentRepo.Save(document); err != nil {
            return err
        }

        // Публикуем события
        for _, event := range document.GetDomainEvents() {
            if err := uc.eventBus.Publish(event); err != nil {
                return err
            }
        }

        return nil
    })
}
```

### 3. **Infrastructure Layer (Инфраструктурный слой)**
```go
// Реализация репозитория
type PostgresDocumentRepository struct {
    db *sql.DB
}

func (r *PostgresDocumentRepository) GetByID(id DocumentID) (*Document, error) {
    query := `
        SELECT id, title, content, author_id, status, created_at, updated_at
        FROM documents
        WHERE id = $1
    `

    var doc Document
    err := r.db.QueryRow(query, id.String()).Scan(
        &doc.id,
        &doc.title,
        &doc.content,
        &doc.authorID,
        &doc.status,
        &doc.createdAt,
        &doc.updatedAt,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrDocumentNotFound
        }
        return nil, err
    }

    return &doc, nil
}
```

### 4. **Interface Layer (Слой интерфейсов)**
```go
// HTTP Handler
type DocumentHandler struct {
    approveUseCase ApproveDocumentUseCase
}

func (h *DocumentHandler) ApproveDocument(w http.ResponseWriter, r *http.Request) {
    var req ApproveDocumentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    cmd := ApproveDocumentCommand{
        DocumentID: DocumentID(req.DocumentID),
        ApproverID: UserID(getUserIDFromContext(r.Context())),
    }

    if err := h.approveUseCase.Execute(cmd); err != nil {
        handleError(w, err)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "approved",
    })
}
```

## 🔧 Dependency Injection

### Container Configuration
```go
// DI Container
type Container struct {
    // Repositories
    userRepo     UserRepository
    documentRepo DocumentRepository
    workflowRepo WorkflowRepository

    // Services
    authService      AuthService
    workflowEngine   WorkflowEngine
    notificationSvc  NotificationService

    // Use Cases
    loginUseCase         LoginUseCase
    createDocumentUseCase CreateDocumentUseCase
    approveDocumentUseCase ApproveDocumentUseCase

    // Infrastructure
    db       *sql.DB
    redis    *redis.Client
    eventBus EventBus
}

func NewContainer(config Config) *Container {
    container := &Container{}

    // Infrastructure
    container.db = setupDatabase(config.Database)
    container.redis = setupRedis(config.Redis)
    container.eventBus = events.NewEventBus()

    // Repositories
    container.userRepo = postgres.NewUserRepository(container.db)
    container.documentRepo = postgres.NewDocumentRepository(container.db)
    container.workflowRepo = postgres.NewWorkflowRepository(container.db)

    // Services
    container.authService = auth.NewService(container.userRepo, container.redis)
    container.workflowEngine = workflow.NewEngine(container.workflowRepo, container.eventBus)
    container.notificationSvc = notifications.NewService(config.SMTP)

    // Use Cases
    container.loginUseCase = auth.NewLoginUseCase(
        container.authService,
        container.userRepo,
    )

    container.createDocumentUseCase = documents.NewCreateDocumentUseCase(
        container.documentRepo,
        container.workflowEngine,
        container.eventBus,
    )

    container.approveDocumentUseCase = documents.NewApproveDocumentUseCase(
        container.documentRepo,
        container.userRepo,
        container.eventBus,
        container.db, // UnitOfWork
    )

    return container
}
```

## 🧪 Тестовая архитектура

### Unit Tests
```go
// Доменные тесты
func TestDocument_Approve_Success(t *testing.T) {
    // Arrange
    doc := &Document{
        id:     DocumentID("doc-1"),
        status: DocumentStatusPending,
    }
    approverID := UserID("user-1")

    // Act
    err := doc.Approve(approverID)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, DocumentStatusApproved, doc.status)
    assert.Len(t, doc.GetDomainEvents(), 1)

    event := doc.GetDomainEvents()[0].(DocumentApproved)
    assert.Equal(t, approverID, event.ApproverID)
}
```

### Integration Tests
```go
// Тест use case с моками
func TestApproveDocumentUseCase_Execute_Success(t *testing.T) {
    // Arrange
    mockDocRepo := &mocks.DocumentRepository{}
    mockUserRepo := &mocks.UserRepository{}
    mockEventBus := &mocks.EventBus{}
    mockUnitOfWork := &mocks.UnitOfWork{}

    useCase := NewApproveDocumentUseCase(
        mockDocRepo, mockUserRepo, mockEventBus, mockUnitOfWork,
    )

    document := &Document{id: "doc-1", status: DocumentStatusPending}
    user := &User{id: "user-1", role: RoleApprover}

    mockDocRepo.On("GetByID", "doc-1").Return(document, nil)
    mockUserRepo.On("GetByID", "user-1").Return(user, nil)
    mockDocRepo.On("Save", document).Return(nil)
    mockEventBus.On("Publish", mock.Anything).Return(nil)
    mockUnitOfWork.On("Execute", mock.Anything).Return(nil)

    cmd := ApproveDocumentCommand{
        DocumentID: "doc-1",
        ApproverID: "user-1",
    }

    // Act
    err := useCase.Execute(cmd)

    // Assert
    assert.NoError(t, err)
    mockDocRepo.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    mockEventBus.AssertExpectations(t)
}
```

## 📈 Миграция к микросервисам

### Этапы миграции

#### Этап 1: Выделение модулей
- Четкое разделение bounded contexts
- Минимизация зависимостей между модулями
- Использование event-driven communication

#### Этап 2: Database per Module
- Разделение схем БД по модулям
- Saga pattern для транзакций между модулями
- Event sourcing для критичных данных

#### Этап 3: Network Communication
- Замена in-process вызовов на HTTP/gRPC
- Circuit breaker pattern
- Service discovery

#### Этап 4: Independent Deployment
- Отдельные Docker images для каждого модуля
- CI/CD pipeline per module
- Feature toggles для плавного перехода

### Готовность к микросервисам
```yaml
readiness_indicators:
  module_independence: ">95%"
  test_coverage: ">80%"
  event_driven_communication: "100%"
  database_separation: "per_module"
  api_stability: "versioned"
  monitoring: "comprehensive"
  team_structure: "aligned_with_modules"
```

## 🔍 Мониторинг и обслуживание

### Metrics
```go
// Метрики модуля
type ModuleMetrics struct {
    RequestsTotal    prometheus.CounterVec
    RequestDuration  prometheus.HistogramVec
    ErrorsTotal      prometheus.CounterVec
    ActiveConnections prometheus.Gauge
}

func (m *ModuleMetrics) RecordRequest(module, operation string, duration time.Duration, err error) {
    m.RequestsTotal.WithLabelValues(module, operation).Inc()
    m.RequestDuration.WithLabelValues(module, operation).Observe(duration.Seconds())

    if err != nil {
        m.ErrorsTotal.WithLabelValues(module, operation, err.Error()).Inc()
    }
}
```

### Health Checks
```go
// Health check для модуля
type HealthChecker struct {
    db    *sql.DB
    redis *redis.Client
}

func (h *HealthChecker) Check(ctx context.Context) error {
    // Проверка БД
    if err := h.db.PingContext(ctx); err != nil {
        return fmt.Errorf("database unhealthy: %w", err)
    }

    // Проверка Redis
    if err := h.redis.Ping(ctx).Err(); err != nil {
        return fmt.Errorf("redis unhealthy: %w", err)
    }

    return nil
}
```

## 🛡️ Безопасность

### Security by Design
- Принцип минимальных привилегий
- Валидация на границах модулей
- Аудит всех доменных операций
- Шифрование sensitive данных

### Security Middleware
```go
func SecurityMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Rate limiting
        if !rateLimiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        // Input validation
        if err := validateRequest(r); err != nil {
            http.Error(w, "Invalid input", http.StatusBadRequest)
            return
        }

        // Security headers
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")

        next.ServeHTTP(w, r)
    })
}
```

Эта модульная архитектура обеспечивает:
- ✅ Четкое разделение ответственностей
- ✅ Высокую тестируемость
- ✅ Простоту поддержки и развития
- ✅ Готовность к переходу на микросервисы
- ✅ Соответствие принципам SOLID и DDD
---

**📅 Актуальность документа**
**Последнее обновление**: 2026-02-05
**Версия проекта**: 0.3.3
**Статус**: Актуальный

