# Module Interaction Guide: Modular Monolith Architecture

**Дата актуальности**: 2025-12-23
**Статус**: Актуально
**Версия**: 1.1

## Содержание

- [Введение](#введение)
- [Архитектурные принципы](#архитектурные-принципы)
- [Bounded Contexts и их границы](#bounded-contexts-и-их-границы)
- [Паттерны взаимодействия модулей](#паттерны-взаимодействия-модулей)
  - [Синхронное взаимодействие](#синхронное-взаимодействие)
  - [Асинхронное взаимодействие](#асинхронное-взаимодействие)
- [Event-Driven Architecture](#event-driven-architecture)
  - [Domain Events Catalog](#domain-events-catalog)
  - [Event Versioning](#event-versioning)
  - [Event Sourcing (для аудита)](#event-sourcing-для-аудита)
- [Saga Patterns](#saga-patterns)
  - [Orchestration Saga](#orchestration-saga)
  - [Choreography Saga](#choreography-saga)
- [Eventual Consistency](#eventual-consistency)
- [Shared Kernel](#shared-kernel)
- [Anti-Corruption Layer](#anti-corruption-layer)
- [Resilience Patterns](#resilience-patterns)
  - [Circuit Breaker](#circuit-breaker)
  - [Retry with Exponential Backoff](#retry-with-exponential-backoff)
  - [Bulkhead Pattern](#bulkhead-pattern)
- [Observability](#observability)
- [Migration to Microservices](#migration-to-microservices)
- [Best Practices](#best-practices)

---

## Введение

Информационная система секретаря-методиста построена как **модульный монолит** - архитектурный паттерн, который объединяет преимущества монолитной архитектуры (простота развертывания, отладки) с модульностью микросервисов (изолированные bounded contexts, готовность к миграции).

### Что такое модульный монолит?

```
Традиционный монолит:
┌─────────────────────────────────────┐
│  Все в одном проекте без границ    │
│  User ────> Document ────> Auth    │
│  Тесная связанность                │
└─────────────────────────────────────┘

Микросервисы:
┌──────────┐    ┌──────────┐    ┌──────────┐
│  Auth    │───>│ Document │───>│  Files   │
│  Service │    │ Service  │    │ Service  │
└──────────┘    └──────────┘    └──────────┘
Сетевые вызовы, сложное развертывание

Модульный монолит:
┌───────────────────────────────────────────┐
│  Один процесс, но четкие границы модулей  │
│  ┌──────┐    ┌──────┐    ┌──────┐        │
│  │ Auth │───>│ Docs │───>│Files │        │
│  └──────┘    └──────┘    └──────┘        │
│  Взаимодействие через события             │
└───────────────────────────────────────────┘
```

### Преимущества нашего подхода

| Аспект | Модульный монолит | Монолит | Микросервисы |
|--------|-------------------|---------|--------------|
| Развертывание | ✅ Один бинарник | ✅ Один бинарник | ❌ Много сервисов |
| Отладка | ✅ Локально | ✅ Просто | ❌ Распределенная трассировка |
| Транзакции | ✅ ACID в БД | ✅ ACID | ❌ Eventual consistency |
| Модульность | ✅ Четкие границы | ❌ Нет границ | ✅ Физические границы |
| Готовность к миграции | ✅ Легко выделить модуль | ❌ Сложно | ✅ Уже микросервисы |
| Производительность | ✅ In-process calls | ✅ In-process | ❌ Network overhead |

---

## Архитектурные принципы

### 1. Bounded Context как единица модульности

Каждый модуль = один Bounded Context из DDD:

```
internal/modules/
├── auth/           # Authentication & Authorization Context
├── documents/      # Document Management Context
├── workflow/       # Workflow & Approvals Context
├── schedule/       # Schedule Management Context
├── tasks/          # Task Assignment Context
├── reporting/      # Reporting & Analytics Context
├── notifications/  # Notification Context
├── files/          # File Storage Context
├── integration/    # 1C Integration Context
└── users/          # User Management Context
```

### 2. Dependency Rule

```
Зависимости только в одном направлении:

┌─────────────────────────────────────────┐
│           Presentation Layer            │
│  (HTTP Handlers, gRPC Services)         │
└───────────┬─────────────────────────────┘
            │ ↓ (только вниз)
┌───────────┴─────────────────────────────┐
│          Application Layer              │
│  (Use Cases, Application Services)      │
└───────────┬─────────────────────────────┘
            │ ↓ (только вниз)
┌───────────┴─────────────────────────────┐
│            Domain Layer                 │
│  (Entities, Value Objects, Aggregates)  │
└───────────┬─────────────────────────────┘
            │ ↓ (только вниз)
┌───────────┴─────────────────────────────┐
│        Infrastructure Layer             │
│  (Repositories, External Services)      │
└─────────────────────────────────────────┘

Важно: Domain слой НЕ зависит от Infrastructure!
```

### 3. No Direct Module Dependencies

❌ **Неправильно:**

```go
// internal/modules/notifications/application/service.go
import "github.com/.../internal/modules/documents/domain/entities"

func (s *NotificationService) NotifyDocumentCreated(doc *entities.Document) {
    // Прямая зависимость от модуля documents
}
```

✅ **Правильно:**

```go
// internal/modules/notifications/application/service.go
type DocumentCreatedEvent struct {
    DocumentID int64
    Title      string
    CreatedBy  int64
}

func (s *NotificationService) HandleDocumentCreated(event DocumentCreatedEvent) {
    // Взаимодействие через события
}
```

### 4. Shared Kernel (минимальный)

Общий код вынесен в `internal/shared/`:

```
internal/shared/
├── domain/
│   └── events/          # Базовые интерфейсы событий
├── infrastructure/
│   ├── eventbus/        # Event Bus реализация
│   ├── logging/         # Логирование
│   └── telemetry/       # Метрики, трассировка
└── testing/
    └── fixtures/        # Тестовые данные
```

**Правило:** Shared Kernel должен быть минимальным. Если код используется только в 1-2 модулях - не выносить в shared.

---

## Bounded Contexts и их границы

### Карта взаимодействий (Context Map)

```
┌──────────────────────────────────────────────────────────────┐
│                    Frontend (Next.js)                         │
└─────────┬────────────────────────────────────────────────────┘
          │ HTTP/REST
          ▼
┌─────────────────────────────────────────────────────────────┐
│                   API Gateway Layer                          │
└─┬──────┬──────┬──────┬──────┬──────┬──────┬──────┬──────┬──┘
  │      │      │      │      │      │      │      │      │
  ▼      ▼      ▼      ▼      ▼      ▼      ▼      ▼      ▼
┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐
│Auth│ │Docs│ │Work│ │Sched│ │Task│ │Repo│ │Noti│ │File│ │User│
└──┬─┘ └──┬─┘ └──┬─┘ └──┬─┘ └──┬─┘ └──┬─┘ └──┬─┘ └──┬─┘ └──┬─┘
   │      │      │      │      │      │      │      │      │
   └──────┴──────┴──────┴──────┴──────┴──────┴──────┴──────┘
                         │
                         ▼
              ┌──────────────────┐
              │   Event Bus      │
              │   (Kafka/Redis)  │
              └──────────────────┘
                         │
                         ▼
              ┌──────────────────┐
              │   Integration    │
              │   (1C System)    │
              └──────────────────┘
```

### Типы взаимодействий между Contexts

1. **Partnership** (партнерство)
   - `Documents ↔ Workflow`: Тесная интеграция через события
   - `Workflow ↔ Notifications`: Уведомления об этапах согласования

2. **Customer-Supplier** (заказчик-поставщик)
   - `Documents → Files`: Документы используют файловый сервис
   - `Tasks → Schedule`: Задания связаны с расписанием

3. **Conformist** (конформист)
   - `Integration → 1C`: Адаптация к внешнему API 1С

4. **Anti-Corruption Layer** (ACL)
   - `Integration`: ACL между нашей системой и 1С

---

## Паттерны взаимодействия модулей

### Синхронное взаимодействие

Используется только когда **необходим немедленный ответ**.

#### 1. Direct Function Call (в рамках одного процесса)

```go
// internal/modules/documents/application/usecases/document_usecase.go
type DocumentUseCase struct {
    docRepo      repositories.DocumentRepository
    authService  *auth.AuthorizationService // ✅ Разрешено: синхронный вызов
}

func (uc *DocumentUseCase) CreateDocument(ctx context.Context, req CreateDocumentRequest) error {
    // Проверка прав доступа (синхронно)
    hasPermission := uc.authService.CheckPermission(
        ctx,
        req.UserID,
        domain.ResourceDocument,
        domain.ActionCreate,
    )

    if !hasPermission {
        return ErrPermissionDenied
    }

    // Создать документ
    doc := entities.NewDocument(req.Title, req.Content, req.TemplateID, req.CreatedBy)
    return uc.docRepo.Create(ctx, doc)
}
```

**Когда использовать:**
- ✅ Авторизация (нужен немедленный ответ)
- ✅ Валидация (синхронная проверка)
- ✅ Чтение справочников (если закэшировано)

**Когда НЕ использовать:**
- ❌ Операции записи в другой модуль
- ❌ Отправка уведомлений
- ❌ Вызовы внешних API

#### 2. Query через Shared Read Model

```go
// internal/shared/readmodel/document_read_model.go
type DocumentReadModel interface {
    GetDocumentSummary(ctx context.Context, docID int64) (*DocumentSummary, error)
    ListDocumentsByUser(ctx context.Context, userID int64) ([]*DocumentSummary, error)
}

// Реализация в documents модуле
// internal/modules/documents/infrastructure/readmodel/document_read_model_impl.go
type DocumentReadModelImpl struct {
    db *sql.DB
}

func (rm *DocumentReadModelImpl) GetDocumentSummary(ctx context.Context, docID int64) (*DocumentSummary, error) {
    // Легковесный запрос (только нужные поля)
    query := `
        SELECT id, title, status, created_by, created_at
        FROM documents
        WHERE id = $1
    `
    // ...
}

// Использование в другом модуле
// internal/modules/notifications/application/service.go
type NotificationService struct {
    docReadModel readmodel.DocumentReadModel
}

func (s *NotificationService) CreateNotification(ctx context.Context, docID int64) error {
    // Получить только необходимую информацию
    summary, err := s.docReadModel.GetDocumentSummary(ctx, docID)
    if err != nil {
        return err
    }

    // Создать уведомление
    notification := fmt.Sprintf("Document '%s' was approved", summary.Title)
    // ...
}
```

### Асинхронное взаимодействие

Основной способ взаимодействия между модулями.

#### Event-Driven Communication

```go
// Публикация события (Producer)
// internal/modules/documents/application/usecases/document_usecase.go
func (uc *DocumentUseCase) CreateDocument(ctx context.Context, req CreateDocumentRequest) error {
    // 1. Создать документ
    doc := entities.NewDocument(req.Title, req.Content, req.TemplateID, req.CreatedBy)
    if err := uc.docRepo.Create(ctx, doc); err != nil {
        return err
    }

    // 2. Опубликовать событие (асинхронно)
    event := events.DocumentCreatedEvent{
        EventID:    uuid.New().String(),
        Timestamp:  time.Now(),
        DocumentID: doc.ID,
        Title:      doc.Title,
        CreatedBy:  doc.CreatedBy,
    }

    return uc.eventBus.Publish(ctx, "document.created", event)
}

// Подписка на событие (Consumer)
// internal/modules/notifications/application/eventhandlers/document_event_handler.go
type DocumentEventHandler struct {
    notificationService *NotificationService
}

func (h *DocumentEventHandler) HandleDocumentCreated(ctx context.Context, event events.DocumentCreatedEvent) error {
    // Отправить уведомление создателю
    return h.notificationService.NotifyDocumentCreated(ctx, event.DocumentID, event.CreatedBy)
}

// Регистрация обработчика
// internal/modules/notifications/module.go
func (m *NotificationModule) RegisterEventHandlers(eventBus *eventbus.EventBus) {
    eventBus.Subscribe("document.created", m.eventHandler.HandleDocumentCreated)
    eventBus.Subscribe("document.approved", m.eventHandler.HandleDocumentApproved)
}
```

---

## Event-Driven Architecture

### Domain Events Catalog

Полный каталог всех событий в системе:

#### Authentication Context Events

| Событие | Описание | Payload | Подписчики |
|---------|----------|---------|------------|
| `user.registered` | Новый пользователь зарегистрирован | `{ user_id, email, role }` | Notifications, Audit |
| `user.login` | Успешный вход | `{ user_id, ip_address, timestamp }` | Audit, Security |
| `user.logout` | Выход из системы | `{ user_id, timestamp }` | Audit |
| `user.password_changed` | Смена пароля | `{ user_id, timestamp }` | Notifications, Security |
| `user.blocked` | Пользователь заблокирован | `{ user_id, reason, blocked_by }` | Notifications, Audit |
| `user.unblocked` | Пользователь разблокирован | `{ user_id, unblocked_by }` | Notifications |

#### Document Management Context Events

| Событие | Описание | Payload | Подписчики |
|---------|----------|---------|------------|
| `document.created` | Документ создан | `{ doc_id, title, created_by, template_id }` | Workflow, Notifications, Audit |
| `document.updated` | Документ обновлен | `{ doc_id, updated_by, changes }` | Notifications, Audit |
| `document.deleted` | Документ удален | `{ doc_id, deleted_by, reason }` | Files, Notifications, Audit |
| `document.published` | Документ опубликован | `{ doc_id, published_by }` | Notifications, Reporting |
| `document.archived` | Документ архивирован | `{ doc_id, archived_by }` | Notifications |
| `document.shared` | Документ расшарен (Issue #13) | `{ doc_id, target_user_id, target_role, permission, shared_by, expires_at }` | Notifications, Audit |
| `document.permission_revoked` | Права отозваны (Issue #13) | `{ doc_id, permission_id, revoked_by }` | Notifications, Audit |
| `document.public_link_created` | Создана публичная ссылка (Issue #13) | `{ doc_id, token, permission, created_by, expires_at, max_uses }` | Audit |
| `document.public_link_deactivated` | Ссылка деактивирована (Issue #13) | `{ doc_id, token, deactivated_by, reason }` | Audit |

#### Workflow Context Events

| Событие | Описание | Payload | Подписчики |
|---------|----------|---------|------------|
| `workflow.started` | Согласование начато | `{ workflow_id, doc_id, started_by }` | Notifications, Audit |
| `workflow.step_completed` | Шаг согласования завершен | `{ workflow_id, step_id, approver_id, decision }` | Notifications, Documents |
| `workflow.approved` | Документ согласован | `{ workflow_id, doc_id, approved_by, approved_at }` | Documents, Notifications, Integration, Audit |
| `workflow.rejected` | Документ отклонен | `{ workflow_id, doc_id, rejected_by, reason }` | Documents, Notifications, Audit |
| `workflow.cancelled` | Согласование отменено | `{ workflow_id, doc_id, cancelled_by, reason }` | Notifications, Audit |

#### Schedule Context Events

| Событие | Описание | Payload | Подписчики |
|---------|----------|---------|------------|
| `schedule.created` | Расписание создано | `{ schedule_id, course_id, semester, created_by }` | Tasks, Notifications, Audit |
| `schedule.updated` | Расписание обновлено | `{ schedule_id, changes, updated_by }` | Tasks, Notifications, Audit |
| `schedule.published` | Расписание опубликовано | `{ schedule_id, published_by }` | Notifications, Integration |
| `lesson.created` | Занятие добавлено | `{ lesson_id, schedule_id, teacher_id, group_id, datetime }` | Notifications |
| `lesson.cancelled` | Занятие отменено | `{ lesson_id, reason, cancelled_by }` | Notifications, Reporting |

#### Task Assignment Context Events

| Событие | Описание | Payload | Подписчики |
|---------|----------|---------|------------|
| `task.created` | Задание создано | `{ task_id, title, assignee_id, due_date, created_by }` | Notifications, Audit |
| `task.assigned` | Задание назначено | `{ task_id, assignee_id, assigned_by }` | Notifications |
| `task.started` | Работа над заданием начата | `{ task_id, started_by, started_at }` | Notifications, Reporting |
| `task.completed` | Задание выполнено | `{ task_id, completed_by, completed_at }` | Notifications, Reporting, Audit |
| `task.overdue` | Задание просрочено | `{ task_id, assignee_id, due_date }` | Notifications, Reporting |

#### Reporting Context Events

| Событие | Описание | Payload | Подписчики |
|---------|----------|---------|------------|
| `report.generated` | Отчет сгенерирован | `{ report_id, type, generated_by, period }` | Notifications, Files, Audit |
| `report.scheduled` | Отчет запланирован | `{ report_id, schedule, recipients }` | Notifications |
| `custom_report.created` | Пользовательский отчёт создан | `{ report_id, name, data_source, created_by }` | Audit |
| `custom_report.updated` | Пользовательский отчёт обновлён | `{ report_id, name, updated_by }` | Audit |
| `custom_report.deleted` | Пользовательский отчёт удалён | `{ report_id, deleted_by }` | Audit |
| `custom_report.executed` | Пользовательский отчёт выполнен | `{ report_id, rows_count, executed_by }` | Audit, Reporting |
| `custom_report.exported` | Пользовательский отчёт экспортирован | `{ report_id, format, exported_by }` | Audit, Files |

#### Notifications Context Events

| Событие | Описание | Payload | Подписчики |
|---------|----------|---------|------------|
| `notification.sent` | Уведомление отправлено | `{ notification_id, user_id, channel, sent_at }` | Audit |
| `notification.failed` | Ошибка отправки | `{ notification_id, user_id, error }` | Audit, Alerting |

#### File Storage Context Events

| Событие | Описание | Payload | Подписчики |
|---------|----------|---------|------------|
| `file.uploaded` | Файл загружен | `{ file_id, name, size, uploaded_by }` | Documents, Audit |
| `file.deleted` | Файл удален | `{ file_id, deleted_by }` | Audit |
| `file.virus_detected` | Обнаружен вирус | `{ file_id, uploaded_by }` | Notifications, Security |

#### Integration Context Events

| Событие | Описание | Payload | Подписчики |
|---------|----------|---------|------------|
| `integration.1c.synced` | Данные синхронизированы с 1С | `{ entity_type, entity_id, synced_at }` | Audit |
| `integration.1c.failed` | Ошибка синхронизации | `{ entity_type, entity_id, error }` | Notifications, Alerting |

### Event Structure (стандарт)

Все события следуют единой структуре:

```go
// internal/shared/domain/events/base_event.go
type BaseEvent struct {
    EventID   string    `json:"event_id"`   // UUID события
    EventType string    `json:"event_type"` // Тип события (document.created)
    Timestamp time.Time `json:"timestamp"`  // Время возникновения
    Version   string    `json:"version"`    // Версия схемы события
}

// Конкретное событие
type DocumentCreatedEvent struct {
    BaseEvent
    DocumentID int64  `json:"document_id"`
    Title      string `json:"title"`
    CreatedBy  int64  `json:"created_by"`
    TemplateID *int64 `json:"template_id,omitempty"`
}
```

### Event Versioning

События могут изменяться со временем. Используем версионирование:

```go
// v1 (старая версия)
type DocumentCreatedEventV1 struct {
    BaseEvent
    DocumentID int64  `json:"document_id"`
    Title      string `json:"title"`
}

// v2 (новая версия - добавлено поле created_by)
type DocumentCreatedEventV2 struct {
    BaseEvent
    DocumentID int64  `json:"document_id"`
    Title      string `json:"title"`
    CreatedBy  int64  `json:"created_by"` // новое поле
}

// Обработчик поддерживает обе версии
func (h *DocumentEventHandler) HandleDocumentCreated(ctx context.Context, eventData []byte) error {
    var baseEvent events.BaseEvent
    json.Unmarshal(eventData, &baseEvent)

    switch baseEvent.Version {
    case "v1":
        var event DocumentCreatedEventV1
        json.Unmarshal(eventData, &event)
        return h.handleV1(ctx, event)

    case "v2":
        var event DocumentCreatedEventV2
        json.Unmarshal(eventData, &event)
        return h.handleV2(ctx, event)

    default:
        return fmt.Errorf("unsupported event version: %s", baseEvent.Version)
    }
}
```

### Event Sourcing (для аудита)

Для критичных операций используем Event Sourcing:

```go
// internal/modules/workflow/infrastructure/eventsourcing/workflow_event_store.go
type WorkflowEventStore struct {
    db *sql.DB
}

// Сохранение всех событий workflow
func (es *WorkflowEventStore) AppendEvent(ctx context.Context, event WorkflowEvent) error {
    query := `
        INSERT INTO workflow_events (
            workflow_id, event_type, event_data, version, occurred_at
        ) VALUES ($1, $2, $3, $4, $5)
    `

    eventData, _ := json.Marshal(event)

    _, err := es.db.ExecContext(ctx, query,
        event.WorkflowID,
        event.EventType,
        eventData,
        event.Version,
        event.Timestamp,
    )

    return err
}

// Восстановление состояния Workflow из событий
func (es *WorkflowEventStore) ReplayWorkflow(ctx context.Context, workflowID int64) (*Workflow, error) {
    query := `
        SELECT event_data
        FROM workflow_events
        WHERE workflow_id = $1
        ORDER BY occurred_at ASC
    `

    rows, err := es.db.QueryContext(ctx, query, workflowID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    workflow := &Workflow{ID: workflowID}

    // Применить все события по порядку
    for rows.Next() {
        var eventData []byte
        rows.Scan(&eventData)

        var event WorkflowEvent
        json.Unmarshal(eventData, &event)

        workflow.Apply(event) // Применить событие к агрегату
    }

    return workflow, nil
}
```

---

## Saga Patterns

Для long-running транзакций между модулями используем Saga паттерн.

### Orchestration Saga

Централизованный координатор управляет всеми шагами.

**Пример: Согласование документа**

```go
// internal/modules/workflow/application/saga/approval_saga.go
type ApprovalSagaOrchestrator struct {
    workflowRepo    repositories.WorkflowRepository
    documentClient  *documents.DocumentClient
    notificationSvc *notifications.NotificationService
    integrationSvc  *integration.IntegrationService
    eventBus        *eventbus.EventBus
}

func (s *ApprovalSagaOrchestrator) ExecuteApproval(ctx context.Context, workflowID int64) error {
    // Шаг 1: Начать workflow
    workflow, err := s.workflowRepo.GetByID(ctx, workflowID)
    if err != nil {
        return err
    }

    workflow.Start()
    s.workflowRepo.Save(ctx, workflow)

    // Шаг 2: Уведомить первого согласующего
    if err := s.notificationSvc.NotifyApprover(ctx, workflow.CurrentStepApproverID()); err != nil {
        return s.rollback(ctx, workflow, "notification_failed")
    }

    // Ожидание решения от согласующего (через события)
    // ...

    // Шаг 3: Обновить статус документа
    if err := s.documentClient.UpdateStatus(ctx, workflow.DocumentID, "approved"); err != nil {
        return s.rollback(ctx, workflow, "document_update_failed")
    }

    // Шаг 4: Синхронизировать с 1С
    if err := s.integrationSvc.SyncDocument(ctx, workflow.DocumentID); err != nil {
        // НЕ откатываем - синхронизация 1С не критична
        log.Error("Failed to sync with 1C", "error", err)
    }

    // Шаг 5: Финальное уведомление
    s.notificationSvc.NotifyApprovalCompleted(ctx, workflow.DocumentID)

    return nil
}

func (s *ApprovalSagaOrchestrator) rollback(ctx context.Context, workflow *Workflow, reason string) error {
    workflow.Cancel(reason)
    s.workflowRepo.Save(ctx, workflow)

    // Откатить изменения в документе
    s.documentClient.UpdateStatus(ctx, workflow.DocumentID, "draft")

    // Уведомить о отмене
    s.notificationSvc.NotifyWorkflowCancelled(ctx, workflow.ID, reason)

    return fmt.Errorf("saga rollback: %s", reason)
}
```

**Диаграмма Orchestration Saga:**

```
┌────────────────────────────────────────────────────┐
│        Approval Saga Orchestrator                  │
│                                                    │
│  1. Start Workflow      ──────────────> Workflow  │
│  2. Notify Approver     ──────────────> Notif.    │
│  3. Wait for Decision   <──────────────┐          │
│  4. Update Document     ──────────────> Documents │
│  5. Sync with 1C        ──────────────> Integration│
│  6. Final Notification  ──────────────> Notif.    │
│                                                    │
│  Rollback if any step fails                       │
└────────────────────────────────────────────────────┘
```

### Choreography Saga

Децентрализованный подход - каждый сервис реагирует на события.

**Пример: Создание документа и автоматическое согласование**

```go
// Шаг 1: Documents модуль публикует событие
// internal/modules/documents/application/usecases/document_usecase.go
func (uc *DocumentUseCase) CreateDocument(ctx context.Context, req CreateDocumentRequest) error {
    doc := entities.NewDocument(req.Title, req.Content, req.TemplateID, req.CreatedBy)
    uc.docRepo.Create(ctx, doc)

    // Публикуем событие
    uc.eventBus.Publish(ctx, "document.created", DocumentCreatedEvent{
        DocumentID: doc.ID,
        CreatedBy:  doc.CreatedBy,
    })

    return nil
}

// Шаг 2: Workflow модуль слушает событие и начинает согласование
// internal/modules/workflow/application/eventhandlers/document_event_handler.go
func (h *DocumentEventHandler) HandleDocumentCreated(ctx context.Context, event DocumentCreatedEvent) error {
    // Определить нужен ли workflow для этого документа
    needsApproval := h.workflowService.DocumentNeedsApproval(ctx, event.DocumentID)
    if !needsApproval {
        return nil
    }

    // Создать workflow
    workflow := h.workflowService.CreateWorkflow(ctx, event.DocumentID)

    // Публикуем событие
    h.eventBus.Publish(ctx, "workflow.started", WorkflowStartedEvent{
        WorkflowID: workflow.ID,
        DocumentID: event.DocumentID,
    })

    return nil
}

// Шаг 3: Notifications модуль слушает событие и отправляет уведомления
// internal/modules/notifications/application/eventhandlers/workflow_event_handler.go
func (h *WorkflowEventHandler) HandleWorkflowStarted(ctx context.Context, event WorkflowStartedEvent) error {
    workflow, _ := h.workflowClient.GetWorkflow(ctx, event.WorkflowID)

    // Уведомить первого согласующего
    return h.notificationService.NotifyApprover(ctx, workflow.CurrentStepApproverID())
}
```

**Диаграмма Choreography Saga:**

```
Documents           Event Bus           Workflow          Notifications
    │                   │                   │                   │
    │  create doc       │                   │                   │
    ├──────────────────>│                   │                   │
    │                   │                   │                   │
    │ document.created  │                   │                   │
    │──────────────────>│                   │                   │
    │                   │  document.created │                   │
    │                   ├──────────────────>│                   │
    │                   │                   │ create workflow   │
    │                   │                   ├──────────────────>│
    │                   │ workflow.started  │                   │
    │                   │<──────────────────┤                   │
    │                   │  workflow.started │                   │
    │                   ├──────────────────────────────────────>│
    │                   │                   │     notify user   │
```

### Когда использовать Orchestration vs Choreography?

| Критерий | Orchestration | Choreography |
|----------|---------------|--------------|
| Сложность логики | ✅ Высокая (много шагов, условия) | ❌ Низкая (простые цепочки) |
| Контроль | ✅ Централизованный | ❌ Распределенный |
| Отладка | ✅ Проще (один координатор) | ❌ Сложнее (нужно следить за событиями) |
| Связанность | ❌ Координатор знает о всех сервисах | ✅ Сервисы не знают друг о друге |
| Масштабируемость | ❌ Координатор - bottleneck | ✅ Легко масштабировать |

**Наш выбор:**
- **Orchestration**: Согласование документов (сложная логика, много шагов)
- **Choreography**: Уведомления, аудит (простые реакции на события)

---

## Eventual Consistency

В модульном монолите некоторые данные могут быть **eventually consistent** (в конечном итоге согласованные).

### Пример: Статистика документов

```go
// Documents модуль хранит документы
// Reporting модуль хранит агрегированную статистику

// 1. Документ создан
// internal/modules/documents/application/usecases/document_usecase.go
func (uc *DocumentUseCase) CreateDocument(ctx context.Context, req CreateDocumentRequest) error {
    doc := entities.NewDocument(req.Title, req.Content, req.TemplateID, req.CreatedBy)

    // Сохранить в основной БД (ACID гарантии)
    if err := uc.docRepo.Create(ctx, doc); err != nil {
        return err
    }

    // Опубликовать событие (асинхронно)
    uc.eventBus.Publish(ctx, "document.created", DocumentCreatedEvent{
        DocumentID: doc.ID,
        CreatedBy:  doc.CreatedBy,
    })

    return nil // Возвращаем успех ДО обновления статистики
}

// 2. Reporting модуль обновляет статистику (асинхронно)
// internal/modules/reporting/application/eventhandlers/document_event_handler.go
func (h *DocumentEventHandler) HandleDocumentCreated(ctx context.Context, event DocumentCreatedEvent) error {
    // Обновить счетчик документов
    return h.statsRepo.IncrementUserDocumentCount(ctx, event.CreatedBy)
}
```

**Временная шкала:**

```
T0: User создает документ
T1: Документ сохранен в БД (ACID)
T2: HTTP ответ 201 Created возвращается клиенту
    ↓
T3: Событие document.created опубликовано
    ↓
T4: Reporting модуль получает событие
    ↓
T5: Статистика обновлена

Между T2 и T5 - окно eventual consistency
Статистика может быть устаревшей на несколько миллисекунд
```

### Обработка eventual consistency в UI

```tsx
// frontend/app/documents/page.tsx
export default function DocumentsPage() {
    const [documents, setDocuments] = useState<Document[]>([])
    const [stats, setStats] = useState({ total: 0 })

    useEffect(() => {
        // Получить документы
        fetch('/api/documents').then(res => res.json()).then(setDocuments)

        // Получить статистику
        fetch('/api/reporting/stats').then(res => res.json()).then(setStats)
    }, [])

    return (
        <div>
            {/* Показываем актуальное количество из списка */}
            <p>Total documents: {documents.length}</p>

            {/* Статистика может быть чуть устаревшей */}
            <p>Total (from cache): {stats.total}</p>
        </div>
    )
}
```

---

## Shared Kernel

Минимальный набор общего кода между модулями.

### Что входит в Shared Kernel?

```
internal/shared/
├── domain/
│   ├── events/
│   │   ├── base_event.go           # Базовая структура события
│   │   └── event_handler.go        # Интерфейс обработчика
│   └── errors/
│       └── domain_errors.go        # Общие domain ошибки
│
├── infrastructure/
│   ├── eventbus/
│   │   ├── event_bus.go            # Event Bus интерфейс
│   │   ├── kafka_event_bus.go      # Kafka реализация
│   │   └── redis_event_bus.go      # Redis Pub/Sub реализация
│   │
│   ├── logging/
│   │   ├── logger.go               # Structured logging
│   │   ├── audit_logger.go         # Аудит логи
│   │   └── security_logger.go      # Security логи
│   │
│   ├── telemetry/
│   │   ├── metrics.go              # Prometheus метрики
│   │   └── tracing.go              # OpenTelemetry трассировка
│   │
│   └── database/
│       ├── transaction.go          # Транзакции
│       └── retry.go                # Retry логика
│
└── testing/
    ├── fixtures/
    │   └── user_fixtures.go        # Тестовые данные
    └── testcontainers/
        └── postgres.go             # PostgreSQL для тестов
```

### Event Bus (ключевой компонент Shared Kernel)

```go
// internal/shared/infrastructure/eventbus/event_bus.go
type EventBus interface {
    // Публикация события
    Publish(ctx context.Context, topic string, event interface{}) error

    // Подписка на событие
    Subscribe(topic string, handler EventHandler) error

    // Закрыть соединения
    Close() error
}

type EventHandler func(ctx context.Context, event interface{}) error

// Kafka реализация
// internal/shared/infrastructure/eventbus/kafka_event_bus.go
type KafkaEventBus struct {
    writer   *kafka.Writer
    readers  map[string]*kafka.Reader
    handlers map[string][]EventHandler
}

func (bus *KafkaEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
    eventData, err := json.Marshal(event)
    if err != nil {
        return err
    }

    return bus.writer.WriteMessages(ctx, kafka.Message{
        Topic: topic,
        Value: eventData,
    })
}

func (bus *KafkaEventBus) Subscribe(topic string, handler EventHandler) error {
    bus.handlers[topic] = append(bus.handlers[topic], handler)

    // Создать Kafka reader если еще не создан
    if _, exists := bus.readers[topic]; !exists {
        bus.readers[topic] = kafka.NewReader(kafka.ReaderConfig{
            Brokers: []string{"kafka:9092"},
            Topic:   topic,
            GroupID: "app-consumer-group",
        })

        // Запустить consumer в фоне
        go bus.consumeMessages(topic)
    }

    return nil
}

func (bus *KafkaEventBus) consumeMessages(topic string) {
    reader := bus.readers[topic]

    for {
        msg, err := reader.ReadMessage(context.Background())
        if err != nil {
            log.Error("Failed to read message", "error", err)
            continue
        }

        // Десериализовать событие
        var event interface{}
        json.Unmarshal(msg.Value, &event)

        // Вызвать все обработчики
        for _, handler := range bus.handlers[topic] {
            if err := handler(context.Background(), event); err != nil {
                log.Error("Handler failed", "topic", topic, "error", err)
            }
        }
    }
}
```

---

## Anti-Corruption Layer

Для интеграции с внешней системой 1С используем ACL.

```go
// internal/modules/integration/domain/adapters/onec_adapter.go
type OneCAdapter struct {
    client *onec.Client
}

// Маппинг внешней модели 1С в нашу доменную модель
func (a *OneCAdapter) GetStudent(ctx context.Context, studentID string) (*domain.Student, error) {
    // Вызов внешнего API 1С
    onecStudent, err := a.client.GetStudent(ctx, studentID)
    if err != nil {
        return nil, err
    }

    // Трансформация внешней модели в нашу
    return &domain.Student{
        ID:        convertOneCID(onecStudent.ID),
        FullName:  onecStudent.FIO, // 1С использует "FIO", мы "FullName"
        GroupCode: onecStudent.GroupCode,
        Email:     onecStudent.Email,
    }, nil
}

// Маппинг нашей модели в модель 1С
func (a *OneCAdapter) SyncDocument(ctx context.Context, doc *domain.Document) error {
    // Трансформация нашей модели в модель 1С
    onecDoc := &onec.Document{
        ID:          convertToOneCID(doc.ID),
        Title:       doc.Title,
        Content:     doc.Content,
        Status:      mapStatusToOneC(doc.Status), // Статусы могут отличаться
        CreatedDate: doc.CreatedAt.Format("2006-01-02"), // 1С ждет строку
    }

    return a.client.CreateDocument(ctx, onecDoc)
}

// Маппинг статусов
func mapStatusToOneC(status domain.DocumentStatus) string {
    mapping := map[domain.DocumentStatus]string{
        domain.StatusDraft:    "Черновик",     // 1С использует русские названия
        domain.StatusPending:  "На согласовании",
        domain.StatusApproved: "Утверждено",
        domain.StatusRejected: "Отклонено",
    }
    return mapping[status]
}
```

---

## Resilience Patterns

### Circuit Breaker

Защита от каскадных сбоев при интеграции с внешними сервисами.

```go
// internal/shared/infrastructure/resilience/circuit_breaker.go
type CircuitBreaker struct {
    maxFailures  int
    resetTimeout time.Duration
    state        State
    failures     int
    lastFailure  time.Time
    mu           sync.Mutex
}

type State int

const (
    StateClosed State = iota // Все работает
    StateOpen                 // Отключено (слишком много ошибок)
    StateHalfOpen            // Пробуем восстановиться
)

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    // Проверить состояние
    if cb.state == StateOpen {
        // Проверить, прошел ли timeout
        if time.Since(cb.lastFailure) > cb.resetTimeout {
            cb.state = StateHalfOpen
            cb.failures = 0
        } else {
            return ErrCircuitBreakerOpen
        }
    }

    // Выполнить функцию
    err := fn()

    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()

        if cb.failures >= cb.maxFailures {
            cb.state = StateOpen
        }

        return err
    }

    // Успех - вернуться в нормальное состояние
    cb.failures = 0
    cb.state = StateClosed

    return nil
}

// Использование для интеграции с 1С
// internal/modules/integration/application/service.go
type IntegrationService struct {
    onecAdapter     *adapters.OneCAdapter
    circuitBreaker  *resilience.CircuitBreaker
}

func (s *IntegrationService) SyncDocument(ctx context.Context, docID int64) error {
    doc, _ := s.docRepo.GetByID(ctx, docID)

    // Защита Circuit Breaker
    return s.circuitBreaker.Execute(func() error {
        return s.onecAdapter.SyncDocument(ctx, doc)
    })
}
```

### Retry with Exponential Backoff

Автоматические повторные попытки с увеличивающейся задержкой.

```go
// internal/shared/infrastructure/resilience/retry.go
func RetryWithBackoff(
    ctx context.Context,
    maxAttempts int,
    initialDelay time.Duration,
    fn func() error,
) error {
    var err error
    delay := initialDelay

    for attempt := 1; attempt <= maxAttempts; attempt++ {
        err = fn()

        if err == nil {
            return nil // Успех
        }

        // Последняя попытка - не ждать
        if attempt == maxAttempts {
            break
        }

        // Экспоненциальная задержка
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            delay *= 2 // 1s, 2s, 4s, 8s, ...
        }
    }

    return fmt.Errorf("failed after %d attempts: %w", maxAttempts, err)
}

// Использование
func (s *IntegrationService) SyncWithRetry(ctx context.Context, docID int64) error {
    return resilience.RetryWithBackoff(ctx, 3, time.Second, func() error {
        return s.SyncDocument(ctx, docID)
    })
}
```

### Bulkhead Pattern

Изоляция пулов ресурсов для предотвращения каскадных сбоев.

```go
// internal/shared/infrastructure/resilience/bulkhead.go
type Bulkhead struct {
    semaphore chan struct{}
}

func NewBulkhead(maxConcurrent int) *Bulkhead {
    return &Bulkhead{
        semaphore: make(chan struct{}, maxConcurrent),
    }
}

func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
    // Попытка получить слот
    select {
    case b.semaphore <- struct{}{}:
        defer func() { <-b.semaphore }()
        return fn()

    case <-ctx.Done():
        return ctx.Err()

    default:
        return ErrBulkheadFull
    }
}

// Использование: ограничение одновременных вызовов 1С
var onecBulkhead = resilience.NewBulkhead(10) // Максимум 10 одновременных вызовов

func (s *IntegrationService) SyncDocument(ctx context.Context, docID int64) error {
    return onecBulkhead.Execute(ctx, func() error {
        return s.onecAdapter.SyncDocument(ctx, doc)
    })
}
```

---

## Observability

### Distributed Tracing

```go
// internal/shared/infrastructure/telemetry/tracing.go
import "go.opentelemetry.io/otel"

func TraceOperation(ctx context.Context, operationName string, fn func(context.Context) error) error {
    tracer := otel.Tracer("inf-sys")

    ctx, span := tracer.Start(ctx, operationName)
    defer span.End()

    err := fn(ctx)

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }

    return err
}

// Использование
func (uc *DocumentUseCase) CreateDocument(ctx context.Context, req CreateDocumentRequest) error {
    return telemetry.TraceOperation(ctx, "DocumentUseCase.CreateDocument", func(ctx context.Context) error {
        // Создание документа
        doc := entities.NewDocument(req.Title, req.Content, req.TemplateID, req.CreatedBy)

        // Сохранение в БД (автоматически создаст child span)
        return uc.docRepo.Create(ctx, doc)
    })
}
```

**Пример трассировки:**

```
CreateDocument (200ms)
├── Validate permissions (5ms)
├── Create document entity (1ms)
├── Save to database (180ms)
│   ├── INSERT documents (150ms)
│   └── INSERT document_history (30ms)
└── Publish event (14ms)
    └── Kafka write (12ms)
```

### Metrics

```go
// internal/shared/infrastructure/telemetry/metrics.go
var (
    HttpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request latency",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint", "status"},
    )

    EventsPublished = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "events_published_total",
            Help: "Total number of events published",
        },
        []string{"event_type"},
    )

    EventsConsumed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "events_consumed_total",
            Help: "Total number of events consumed",
        },
        []string{"event_type", "status"},
    )
)

// Middleware для HTTP метрик
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        recorder := &statusRecorder{ResponseWriter: w, status: 200}
        next.ServeHTTP(recorder, r)

        duration := time.Since(start).Seconds()

        HttpRequestDuration.WithLabelValues(
            r.Method,
            r.URL.Path,
            strconv.Itoa(recorder.status),
        ).Observe(duration)
    })
}
```

---

## Migration to Microservices

Модульный монолит готов к миграции на микросервисы.

### Этапы миграции

**Этап 1: Модульный монолит (текущее состояние)**

```
┌──────────────────────────────────────────────┐
│         Modular Monolith                     │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐       │
│  │ Auth │ │ Docs │ │ Work │ │ Noti │       │
│  └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘       │
│     │        │        │        │            │
│     └────────┴────────┴────────┘            │
│              Event Bus                       │
│                                              │
│     ┌────────────────────────┐              │
│     │   PostgreSQL (shared)  │              │
│     └────────────────────────┘              │
└──────────────────────────────────────────────┘
```

**Этап 2: Выделить первый микросервис (Notifications)**

```
┌─────────────────────────────────────┐
│      Modular Monolith               │
│  ┌──────┐ ┌──────┐ ┌──────┐        │
│  │ Auth │ │ Docs │ │ Work │        │
│  └──┬───┘ └──┬───┘ └──┬───┘        │
│     │        │        │             │
│     └────────┴────────┘             │
│          Event Bus (Kafka)          │
└──────────┬──────────────────────────┘
           │
           │ Kafka
           ▼
┌──────────────────────────┐
│  Notifications Service   │
│  (отдельный процесс)     │
│                          │
│  ┌────────────────────┐  │
│  │  PostgreSQL (own)  │  │
│  └────────────────────┘  │
└──────────────────────────┘
```

**Этап 3: Выделить остальные сервисы**

```
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│  Auth    │  │ Documents│  │ Workflow │  │  Notif.  │
│  Service │  │ Service  │  │ Service  │  │ Service  │
└────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘
     │             │             │             │
     └─────────────┴─────────────┴─────────────┘
                    │
                    ▼
            ┌───────────────┐
            │  Kafka / NATS │
            └───────────────┘
```

### Что нужно изменить при миграции

1. **Event Bus: in-process → Kafka**

```go
// Было (in-process)
eventBus.Publish(ctx, "document.created", event)

// Стало (Kafka)
kafkaProducer.WriteMessages(ctx, kafka.Message{
    Topic: "document.created",
    Value: json.Marshal(event),
})
```

1. **Database: shared → per-service**

```go
// Было (один PostgreSQL)
db := sql.Open("postgres", "host=db port=5432 dbname=inf_sys")

// Стало (отдельная БД для каждого сервиса)
// Documents Service
docsDB := sql.Open("postgres", "host=docs-db port=5432 dbname=documents")

// Workflow Service
workflowDB := sql.Open("postgres", "host=workflow-db port=5432 dbname=workflow")
```

1. **Synchronous calls → API calls**

```go
// Было (прямой вызов)
hasPermission := authService.CheckPermission(ctx, userID, resource, action)

// Стало (HTTP API)
resp, _ := http.Get(fmt.Sprintf("http://auth-service/api/permissions/check?user_id=%d", userID))
var result struct { HasPermission bool }
json.NewDecoder(resp.Body).Decode(&result)
hasPermission := result.HasPermission
```

---

## Best Practices

### ✅ DO

1. **Используйте события для взаимодействия между модулями**
   ```go
   // ✅ Правильно
   eventBus.Publish(ctx, "document.created", event)
   ```

2. **Держите Bounded Contexts изолированными**
   ```go
   // ✅ Правильно - нет импортов из других модулей
   import "internal/modules/documents/domain/entities"
   ```

3. **Версионируйте события**
   ```go
   // ✅ Правильно
   type DocumentCreatedEventV2 struct {
       BaseEvent
       Version string `json:"version"` // "v2"
   }
   ```

4. **Используйте idempotent consumers**
   ```go
   // ✅ Правильно - проверка дубликатов
   func (h *Handler) HandleEvent(ctx context.Context, event Event) error {
       if h.processed[event.EventID] {
           return nil // Уже обработано
       }
       // Обработка...
       h.processed[event.EventID] = true
   }
   ```

5. **Мониторьте события**
   ```go
   // ✅ Правильно
   EventsPublished.WithLabelValues(event.Type).Inc()
   ```

### ❌ DON'T

1. **Не делайте прямые вызовы между модулями для записи**
   ```go
   // ❌ Неправильно
   documentService.UpdateStatus(ctx, docID, "approved")

   // ✅ Правильно
   eventBus.Publish(ctx, "workflow.approved", event)
   ```

2. **Не импортируйте domain entities из других модулей**
   ```go
   // ❌ Неправильно
   import "internal/modules/documents/domain/entities"

   // ✅ Правильно - используйте DTO
   type DocumentDTO struct {
       ID    int64
       Title string
   }
   ```

3. **Не блокируйте HTTP handlers на ожидание событий**
   ```go
   // ❌ Неправильно
   func CreateDocument(w http.ResponseWriter, r *http.Request) {
       doc := createDocument()
       publishEvent("document.created", doc)
       waitForWorkflowStarted() // ❌ Блокирование
   }

   // ✅ Правильно
   func CreateDocument(w http.ResponseWriter, r *http.Request) {
       doc := createDocument()
       publishEvent("document.created", doc)
       return 201 // Немедленный ответ
   }
   ```

4. **Не используйте shared database между модулями**
   ```go
   // ❌ Неправильно
   // documents модуль читает из таблицы users напрямую
   SELECT * FROM users WHERE id = $1

   // ✅ Правильно
   // documents модуль получает user через событие или read model
   userDTO := userReadModel.GetUser(ctx, userID)
   ```

---

## Заключение

Модульный монолит обеспечивает:

1. ✅ **Четкие границы модулей** - каждый Bounded Context изолирован
2. ✅ **Event-Driven взаимодействие** - слабая связанность через события
3. ✅ **Готовность к миграции** - легко выделить модуль в микросервис
4. ✅ **Простота развертывания** - один бинарник для всей системы
5. ✅ **Resilience** - Circuit Breaker, Retry, Bulkhead для внешних интеграций
6. ✅ **Observability** - трассировка, метрики, структурированные логи

**Основной принцип**: Думай как о микросервисах, развертывай как монолит.
