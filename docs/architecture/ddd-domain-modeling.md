# Domain-Driven Design: Domain Modeling Guide

## Оглавление
- [Введение](#введение)
- [Bounded Contexts](#bounded-contexts)
- [Ubiquitous Language](#ubiquitous-language)
- [Aggregates](#aggregates)
- [Domain Events](#domain-events)
- [Anti-Corruption Layer](#anti-corruption-layer)
- [Процесс моделирования](#процесс-моделирования)

---

## Введение

### Почему DDD?

Информационная система секретаря-методиста - это сложная доменная область с множеством бизнес-правил, workflow'ов и согласований. **Domain-Driven Design (DDD)** выбран как архитектурный подход по следующим причинам:

#### ✅ Преимущества для нашего проекта:

1. **Сложная бизнес-логика**
   - Workflow согласования документов (до 4 шагов)
   - RBAC с 5 ролями и матрицей разрешений
   - Автоматизация расписаний и заданий
   - Интеграция с 1С

2. **Ubiquitous Language**
   - Единый язык между разработчиками и заказчиком (методистами, секретарями)
   - Код отражает реальные бизнес-процессы
   - Сокращение недопонимания требований

3. **Модульная архитектура**
   - 10 изолированных Bounded Contexts
   - Готовность к миграции на микросервисы
   - Независимое развитие модулей

4. **Event-Driven Architecture**
   - Естественное моделирование событий предметной области
   - Async коммуникация между модулями
   - Audit trail из коробки

#### ⚖️ Trade-offs:

| Преимущество | Недостаток | Решение |
|--------------|------------|---------|
| Чистая доменная логика | Больше кода (слои) | Code generators, templates |
| Изолированные модули | Дублирование моделей | Shared Kernel для общих концепций |
| Event Sourcing | Сложность debugging | Comprehensive logging, Event Store UI |
| Строгие границы | Сложнее cross-module queries | CQRS + read models |

---

## Bounded Contexts

### Определение Bounded Contexts

**Bounded Context** - это явная граница, внутри которой конкретная доменная модель определена и применима. В нашей системе 10 Bounded Contexts:

### 1. Authentication Context

**Ответственность**: Аутентификация, авторизация, управление сессиями

**Ubiquitous Language**:
- **User** (Пользователь): Субъект с учетными данными
- **Session** (Сессия): Активное подключение пользователя
- **Role** (Роль): SystemAdmin, Methodist, AcademicSecretary, Teacher, Student
- **Permission** (Разрешение): Право на действие над ресурсом
- **AccessLevel** (Уровень доступа): Denied, Limited, Own, Full

**Core Aggregates**:
- `User` (Root: UserID)
- `Session` (Root: SessionID)
- `Role` (Root: RoleType)

**Key Domain Events**:
```go
UserRegistered {
    UserID    string
    Email     string
    Role      RoleType
    Timestamp time.Time
}

UserLoggedIn {
    UserID    string
    SessionID string
    IP        string
    Timestamp time.Time
}

RoleAssigned {
    UserID    string
    Role      RoleType
    AssignedBy string
    Timestamp time.Time
}
```

**External Dependencies**:
- OAuth providers (Google, Azure AD)
- Redis (session store)

---

### 2. Document Management Context

**Ответственность**: CRUD документов, версионирование, поиск, метаданные

**Ubiquitous Language**:
- **Curriculum** (Учебный план): Официальный план обучения программы
- **Report** (Отчет): Периодический отчет о прогрессе
- **Schedule** (Расписание): Временная таблица занятий
- **Assignment** (Задание): Учебная задача для студентов
- **Template** (Шаблон): Переиспользуемая структура документа
- **Version** (Версия): Снимок документа в конкретный момент времени
- **Metadata** (Метаданные): Теги, категории, автор, дата создания

**Core Aggregates**:
```go
// Document Aggregate Root
type Document struct {
    ID          DocumentID
    Type        DocumentType  // Curriculum, Report, Schedule, Assignment
    Title       string
    Content     []byte        // JSON structure based on template
    TemplateID  TemplateID
    Versions    []Version     // Версии документа
    Metadata    Metadata
    Status      DocumentStatus
    CreatedBy   UserID
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Invariants:
// - Нельзя удалить документ в статусе "Approved"
// - Версия создается при каждом изменении content
// - Template должен существовать при создании
```

**Domain Events**:
```go
DocumentCreated {
    DocumentID   string
    Type         string
    CreatedBy    string
    TemplateID   string
    Timestamp    time.Time
}

DocumentVersioned {
    DocumentID   string
    VersionID    string
    PrevVersion  string
    ChangedBy    string
    Timestamp    time.Time
}

DocumentStatusChanged {
    DocumentID   string
    OldStatus    string
    NewStatus    string
    ChangedBy    string
    Timestamp    time.Time
}

// Sharing Events (Issue #13)
DocumentShared {
    DocumentID   string
    TargetUserID string    // nullable - если шаринг по роли
    TargetRole   string    // nullable - если шаринг по пользователю
    Permission   string    // read, write, delete, admin
    SharedBy     string
    ExpiresAt    time.Time // nullable
    Timestamp    time.Time
}

DocumentPermissionRevoked {
    DocumentID   string
    PermissionID string
    RevokedBy    string
    Timestamp    time.Time
}

PublicLinkCreated {
    DocumentID   string
    Token        string
    Permission   string    // read, download
    CreatedBy    string
    ExpiresAt    time.Time // nullable
    MaxUses      int       // nullable
    HasPassword  bool
    Timestamp    time.Time
}

PublicLinkDeactivated {
    DocumentID   string
    Token        string
    DeactivatedBy string
    Reason       string    // manual, expired, max_uses_reached
    Timestamp    time.Time
}
```

**Sharing Entities (Issue #13)**:
```go
// DocumentPermission - права доступа к документу
type DocumentPermission struct {
    ID          PermissionID
    DocumentID  DocumentID
    UserID      UserID       // nullable - если шаринг по роли
    Role        string       // nullable - если шаринг по пользователю
    Permission  PermissionType // read, write, delete, admin
    GrantedBy   UserID
    ExpiresAt   time.Time    // nullable - бессрочно
    CreatedAt   time.Time
}

// PermissionType - типы прав доступа
type PermissionType string
const (
    PermissionRead   PermissionType = "read"
    PermissionWrite  PermissionType = "write"
    PermissionDelete PermissionType = "delete"
    PermissionAdmin  PermissionType = "admin"
)

// PublicLink - публичная ссылка на документ
type PublicLink struct {
    ID           PublicLinkID
    DocumentID   DocumentID
    Token        string       // уникальный токен для доступа
    Permission   PermissionType // read, download
    CreatedBy    UserID
    ExpiresAt    time.Time    // nullable
    MaxUses      int          // nullable - неограничено
    UseCount     int
    PasswordHash string       // nullable - опциональный пароль
    IsActive     bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// Invariants:
// - UserID или Role обязательно (одно из двух)
// - Token должен быть уникальным
// - UseCount не может превышать MaxUses
// - Нельзя создать публичную ссылку на Draft документ
```

**Business Rules**:
1. **Версионирование**: При изменении Content создается новая Version
2. **Валидация по типу**: Curriculum требует полей "Program", "Credits", "Duration"
3. **Права доступа**: Только автор или Admin может удалить Draft
4. **Template consistency**: Нельзя изменить TemplateID после создания
5. **Шаринг документов** (Issue #13):
   - Автор документа может шарить свой документ любому пользователю/роли
   - Admin может шарить любой документ
   - Публичные ссылки могут создавать только пользователи с правами на документ
   - Публичная ссылка не может быть создана для документа в статусе Draft
   - При истечении срока действия права автоматически отзываются
   - При достижении лимита использований публичная ссылка деактивируется

---

### 3. Workflow Context

**Ответственность**: Согласование документов, маршрутизация, эскалация, SLA

**Ubiquitous Language**:
- **Workflow Instance** (Экземпляр процесса): Конкретное прохождение согласования
- **Step** (Шаг): Этап согласования (Draft → Review → Approve → Published)
- **Approver** (Согласующий): Роль или конкретный пользователь с правом approve
- **Rejection** (Отклонение): Возврат документа на доработку
- **Escalation** (Эскалация): Автоматическая передача при просрочке SLA
- **SLA** (Service Level Agreement): Максимальное время на шаг

**Core Aggregates**:
```go
type WorkflowInstance struct {
    ID           WorkflowInstanceID
    DocumentID   DocumentID
    DocumentType DocumentType
    CurrentStep  WorkflowStep
    Steps        []WorkflowStep
    Status       WorkflowStatus  // InProgress, Approved, Rejected, Escalated
    CreatedBy    UserID
    CreatedAt    time.Time
    CompletedAt  *time.Time
}

type WorkflowStep struct {
    StepNumber   int
    Name         string
    RequiredRole RoleType
    AssignedTo   *UserID       // null = any user with role
    Status       StepStatus    // Pending, InReview, Approved, Rejected
    SLA          time.Duration
    StartedAt    *time.Time
    CompletedAt  *time.Time
    Comments     string
}

// Invariants:
// - Нельзя skip шаг
// - Только пользователь с RequiredRole может approve
// - После Approved нельзя вернуться в InProgress
```

**Workflow Patterns**:

#### Curriculum Approval (4 шага):
```
Draft → Methodist Review → Admin Approval → Published
         (5 days SLA)       (3 days SLA)
```

#### Report Approval (2 шага):
```
Draft → Methodist Approval → Published
         (2 days SLA)
```

**Domain Events**:
```go
WorkflowStarted {
    WorkflowID   string
    DocumentID   string
    DocumentType string
    InitiatedBy  string
    Timestamp    time.Time
}

WorkflowStepApproved {
    WorkflowID   string
    StepNumber   int
    ApprovedBy   string
    Comments     string
    Timestamp    time.Time
}

WorkflowStepRejected {
    WorkflowID   string
    StepNumber   int
    RejectedBy   string
    Reason       string
    Timestamp    time.Time
}

WorkflowEscalated {
    WorkflowID   string
    StepNumber   int
    Reason       string  // SLA_EXCEEDED, APPROVER_UNAVAILABLE
    EscalatedTo  string
    Timestamp    time.Time
}

WorkflowCompleted {
    WorkflowID   string
    FinalStatus  string
    CompletedBy  string
    Duration     int64   // milliseconds
    Timestamp    time.Time
}
```

**Business Rules**:
1. **Escalation**: Если шаг не завершен за SLA → auto-escalate to next role in hierarchy
2. **Rejection handling**: При Rejected возврат на Draft, все следующие шаги сбрасываются
3. **Approval authority**: Methodist может approve Methodist Review, но не Admin Approval

---

### 4. Schedule Context

**Ответственность**: Генерация расписаний, управление занятиями, конфликты

**Ubiquitous Language**:
- **Timetable** (Расписание): Набор занятий на период (семестр, неделя)
- **Lesson** (Занятие): Единица расписания (дата, время, группа, преподаватель, аудитория)
- **Conflict** (Конфликт): Пересечение ресурсов (преподаватель занят, аудитория занята)
- **Recurrence** (Повторение): Паттерн регулярных занятий (каждый понедельник 9:00)
- **Classroom** (Аудитория): Физическое помещение с capacity
- **TimeSlot** (Временной слот): Стандартный период (пара 1: 9:00-10:30)

**Core Aggregates**:
```go
type Timetable struct {
    ID        TimetableID
    Semester  string
    StartDate time.Time
    EndDate   time.Time
    Lessons   []Lesson
    Status    TimetableStatus
}

type Lesson struct {
    ID          LessonID
    Date        time.Time
    TimeSlot    TimeSlot
    SubjectID   string
    TeacherID   UserID
    GroupID     string
    ClassroomID string
    Recurrence  *RecurrencePattern
}

// Invariants:
// - Один преподаватель не может вести два занятия одновременно
// - Одна аудитория не может использоваться дважды одновременно
// - Группа не может иметь два занятия одновременно
// - Capacity аудитории >= количество студентов группы
```

**Domain Events**:
```go
TimetableGenerated {
    TimetableID string
    Semester    string
    LessonCount int
    GeneratedBy string
    Timestamp   time.Time
}

LessonScheduled {
    LessonID    string
    TimetableID string
    Date        string
    TeacherID   string
    ClassroomID string
    Timestamp   time.Time
}

ScheduleConflictDetected {
    LessonID     string
    ConflictType string  // TEACHER_BUSY, CLASSROOM_OCCUPIED, GROUP_BUSY
    ConflictWith string
    Timestamp    time.Time
}

LessonCanceled {
    LessonID  string
    Reason    string
    CanceledBy string
    Timestamp  time.Time
}
```

---

### 5. Tasks Context

**Ответственность**: Создание заданий, сдача работ, оценивание

**Ubiquitous Language**:
- **Task** (Задание): Учебная задача с deadline
- **Submission** (Сдача): Результат выполнения задания студентом
- **Grading** (Оценивание): Проверка и выставление оценки
- **Rubric** (Критерии): Матрица оценивания
- **Attachment** (Вложение): Файл с решением
- **Plagiarism Check** (Проверка на плагиат): Автоматическая проверка

**Core Aggregates**:
```go
type Task struct {
    ID          TaskID
    Title       string
    Description string
    SubjectID   string
    TeacherID   UserID
    DueDate     time.Time
    MaxScore    int
    Rubric      Rubric
    Submissions []Submission
    Status      TaskStatus
}

type Submission struct {
    ID             SubmissionID
    TaskID         TaskID
    StudentID      UserID
    Attachments    []Attachment
    SubmittedAt    time.Time
    Grade          *int
    Feedback       string
    PlagiarismScore *float64
    Status         SubmissionStatus
}

// Invariants:
// - Нельзя submit после DueDate (кроме Late submission с penalty)
// - Только автор Task может grade
// - Grade от 0 до MaxScore
```

**Domain Events**:
```go
TaskCreated {
    TaskID    string
    Title     string
    TeacherID string
    DueDate   string
    Timestamp time.Time
}

SubmissionReceived {
    SubmissionID string
    TaskID       string
    StudentID    string
    IsLate       bool
    Timestamp    time.Time
}

SubmissionGraded {
    SubmissionID string
    TaskID       string
    StudentID    string
    Grade        int
    GradedBy     string
    Timestamp    time.Time
}

PlagiarismDetected {
    SubmissionID    string
    StudentID       string
    PlagiarismScore float64
    SimilarTo       []string
    Timestamp       time.Time
}
```

---

## Ubiquitous Language

### Глоссарий терминов по Bounded Contexts

#### General Terms (Общие для всех контекстов)

| Термин (Русский) | Термин (English) | Определение | Контекст |
|------------------|------------------|-------------|----------|
| Пользователь | User | Любой субъект с учетными данными в системе | All |
| Роль | Role | Набор разрешений и ответственностей | Auth |
| Разрешение | Permission | Право на действие над ресурсом | Auth |
| Событие | Event | Факт, произошедший в системе | All |
| Aggregate | Aggregate | Кластер объектов, обрабатываемых как единое целое | All |

#### Document Management Terms

| Термин | English | Определение |
|--------|---------|-------------|
| Учебный план | Curriculum | Официальный план обучения образовательной программы |
| Отчет | Report | Периодический отчет о прогрессе (методиста, студента, кафедры) |
| Расписание | Schedule/Timetable | Временная таблица занятий на семестр/неделю |
| Задание | Assignment/Task | Учебная задача для выполнения студентами |
| Шаблон | Template | Переиспользуемая структура документа с полями |
| Версия | Version | Снимок состояния документа в конкретный момент |
| Метаданные | Metadata | Теги, категории, автор, дата создания документа |

#### Workflow Terms

| Термин | English | Определение |
|--------|---------|-------------|
| Согласование | Approval | Процесс получения разрешения от уполномоченных лиц |
| Маршрут | Workflow Route | Последовательность шагов согласования |
| Шаг | Step | Этап процесса согласования (Review, Approve, Publish) |
| Согласующий | Approver | Роль или пользователь с правом approve на шаге |
| Отклонение | Rejection | Возврат документа на доработку с комментариями |
| Эскалация | Escalation | Автоматическая передача при просрочке SLA |
| SLA | Service Level Agreement | Максимальное время на выполнение шага |

#### Schedule Terms

| Термин | English | Определение |
|--------|---------|-------------|
| Занятие | Lesson | Единица расписания (предмет, группа, преподаватель, время, место) |
| Пара | TimeSlot | Стандартный период занятия (1 пара: 9:00-10:30) |
| Аудитория | Classroom | Физическое помещение с вместимостью |
| Конфликт | Conflict | Пересечение ресурсов (преподаватель/аудитория заняты) |
| Повторение | Recurrence | Паттерн регулярных занятий (еженедельно, по четным неделям) |

#### Tasks Terms

| Термин | English | Определение |
|--------|---------|-------------|
| Сдача | Submission | Результат выполнения задания студентом |
| Оценивание | Grading | Проверка работы и выставление оценки |
| Критерии | Rubric | Матрица оценивания с весами критериев |
| Плагиат | Plagiarism | Заимствование чужой работы без ссылки |
| Просрочка | Late Submission | Сдача после deadline с возможным penalty |

---

## Aggregates

### Принципы проектирования Aggregates

#### 1. Aggregate Root

**Aggregate Root** - единственная точка входа в Aggregate. Все изменения идут через Root.

**Пример**: `Document` Aggregate
```go
// ✅ CORRECT: Изменение через Root
document.UpdateTitle("New Title")
document.AddVersion(version)

// ❌ WRONG: Прямое изменение вложенных объектов
document.Versions[0].Content = newContent  // bypasses invariants!
```

#### 2. Transactional Consistency

Aggregate - граница транзакционной консистентности. Одна транзакция БД = один Aggregate.

**Пример**: При создании Workflow нельзя в той же транзакции обновить Document
```go
// ✅ CORRECT: Разные транзакции
tx1: CreateDocument(doc)
tx2: StartWorkflow(workflowInstance)  // через Domain Event

// ❌ WRONG: Две Aggregates в одной транзакции
tx: {
    CreateDocument(doc)
    StartWorkflow(workflowInstance)  // coupling!
}
```

#### 3. Eventual Consistency

Между Aggregates - eventual consistency через Domain Events.

**Пример**: Document создан → событие → Workflow запущен асинхронно
```go
// Publisher (Document Context)
func (s *DocumentService) CreateDocument(doc *Document) error {
    if err := s.repo.Save(doc); err != nil {
        return err
    }

    // Publish event
    event := DocumentCreated{
        DocumentID: doc.ID,
        Type:       doc.Type,
        CreatedBy:  doc.CreatedBy,
    }
    s.eventBus.Publish("document.created", event)
    return nil
}

// Subscriber (Workflow Context)
func (h *WorkflowEventHandler) OnDocumentCreated(event DocumentCreated) {
    // Auto-start workflow based on document type
    workflow := CreateWorkflowForDocumentType(event.DocumentID, event.Type)
    h.repo.Save(workflow)
}
```

#### 4. Small Aggregates

Держим Aggregates маленькими для снижения contention и повышения concurrency.

**Пример**: Вместо большого `Timetable` Aggregate с 1000 `Lesson`
```go
// ❌ WRONG: Огромный Aggregate
type Timetable struct {
    ID      string
    Lessons []Lesson  // 1000 lessons → huge transaction, lock contention
}

// ✅ CORRECT: Маленькие Aggregates
type Timetable struct {
    ID       string
    Semester string
    // Lessons stored as separate Aggregates
}

type Lesson struct {
    ID          string
    TimetableID string  // reference
    Date        time.Time
    // ...
}
```

---

### Aggregates по контекстам

#### Document Aggregate

**Root**: `Document`
**Entities**: `Version`, `Metadata`
**Value Objects**: `DocumentID`, `TemplateID`, `DocumentStatus`

**Invariants**:
1. `Title` не может быть пустым
2. При изменении `Content` создается новая `Version`
3. `Status` переходы: Draft → UnderReview → Approved → Published
4. Нельзя удалить документ в статусе Approved

```go
type Document struct {
    id        DocumentID      // private field
    Type      DocumentType
    Title     string
    Content   []byte
    Versions  []Version
    Status    DocumentStatus
    // ...
}

// Public method enforcing invariants
func (d *Document) UpdateContent(newContent []byte, updatedBy UserID) error {
    if d.Status == StatusApproved || d.Status == StatusPublished {
        return ErrCannotEditApprovedDocument
    }

    // Create version before changing
    version := Version{
        ID:        generateVersionID(),
        Content:   d.Content,  // old content
        CreatedBy: updatedBy,
        CreatedAt: time.Now(),
    }
    d.Versions = append(d.Versions, version)

    // Update content
    d.Content = newContent
    d.UpdatedAt = time.Now()

    return nil
}
```

#### Workflow Aggregate

**Root**: `WorkflowInstance`
**Entities**: `WorkflowStep`
**Value Objects**: `WorkflowInstanceID`, `WorkflowStatus`, `SLA`

**Invariants**:
1. Шаги выполняются строго по порядку
2. Только пользователь с `RequiredRole` может approve шаг
3. После `Approved` нельзя вернуться в `InProgress`
4. SLA tracking: если шаг не завершен за SLA → auto-escalate

```go
type WorkflowInstance struct {
    id          WorkflowInstanceID
    CurrentStep int
    Steps       []WorkflowStep
    Status      WorkflowStatus
    // ...
}

func (w *WorkflowInstance) ApproveCurrentStep(approver UserContext, comments string) error {
    step := &w.Steps[w.CurrentStep]

    // Check role
    if !approver.HasRole(step.RequiredRole) {
        return ErrInsufficientPermissions
    }

    // Check SLA
    if time.Since(step.StartedAt) > step.SLA {
        w.escalate()
        return ErrSLAExceeded
    }

    // Approve step
    step.Status = StepStatusApproved
    step.CompletedAt = time.Now()
    step.Comments = comments

    // Move to next step or complete
    if w.CurrentStep < len(w.Steps)-1 {
        w.CurrentStep++
        w.Steps[w.CurrentStep].Status = StepStatusInReview
        w.Steps[w.CurrentStep].StartedAt = time.Now()
    } else {
        w.Status = WorkflowStatusCompleted
        w.CompletedAt = time.Now()
    }

    return nil
}
```

---

## Domain Events

### Event Catalog

Полный список всех доменных событий системы.

#### Authentication Context Events

| Event | Publisher | Subscribers | Payload | Guarantees |
|-------|-----------|-------------|---------|------------|
| `UserRegistered` | Auth Service | Notification, Audit | UserID, Email, Role | At-least-once |
| `UserLoggedIn` | Auth Service | Audit, Analytics | UserID, SessionID, IP | At-least-once |
| `UserLoggedOut` | Auth Service | Session Manager | UserID, SessionID | At-least-once |
| `RoleAssigned` | User Service | Auth Cache, Audit | UserID, Role, AssignedBy | Exactly-once |
| `PermissionChanged` | RBAC Service | Auth Cache | RoleType, Permissions | Exactly-once |

#### Document Context Events

| Event | Publisher | Subscribers | Payload | Guarantees |
|-------|-----------|-------------|---------|------------|
| `DocumentCreated` | Document Service | Workflow, Search, Notification | DocumentID, Type, CreatedBy | At-least-once |
| `DocumentUpdated` | Document Service | Search, Notification | DocumentID, ChangedFields | At-least-once |
| `DocumentVersioned` | Document Service | Audit, Backup | DocumentID, VersionID, PrevVersion | Exactly-once |
| `DocumentDeleted` | Document Service | Search, Files, Workflow | DocumentID, DeletedBy | Exactly-once |
| `DocumentStatusChanged` | Document Service | Workflow, Notification | DocumentID, OldStatus, NewStatus | At-least-once |

#### Workflow Context Events

| Event | Publisher | Subscribers | Payload | Guarantees |
|-------|-----------|-------------|---------|------------|
| `WorkflowStarted` | Workflow Service | Document, Notification | WorkflowID, DocumentID, InitiatedBy | At-least-once |
| `WorkflowStepApproved` | Workflow Service | Document, Notification, Audit | WorkflowID, StepNumber, ApprovedBy | Exactly-once |
| `WorkflowStepRejected` | Workflow Service | Document, Notification | WorkflowID, StepNumber, RejectedBy, Reason | Exactly-once |
| `WorkflowEscalated` | Workflow Service | Notification, Admin Dashboard | WorkflowID, StepNumber, EscalatedTo | At-least-once |
| `WorkflowCompleted` | Workflow Service | Document, Analytics, Notification | WorkflowID, FinalStatus, Duration | Exactly-once |

#### Schedule Context Events

| Event | Publisher | Subscribers | Payload | Guarantees |
|-------|-----------|-------------|---------|------------|
| `TimetableGenerated` | Schedule Service | Notification, Calendar Sync | TimetableID, Semester, LessonCount | At-least-once |
| `LessonScheduled` | Schedule Service | Calendar, Notification | LessonID, Date, TeacherID, GroupID | At-least-once |
| `ScheduleConflictDetected` | Schedule Service | Admin Dashboard, Notification | LessonID, ConflictType, ConflictWith | At-least-once |
| `LessonCanceled` | Schedule Service | Notification, Calendar Sync | LessonID, Reason, CanceledBy | Exactly-once |

#### Tasks Context Events

| Event | Publisher | Subscribers | Payload | Guarantees |
|-------|-----------|-------------|---------|------------|
| `TaskCreated` | Task Service | Notification, Calendar | TaskID, Title, DueDate, TeacherID | At-least-once |
| `SubmissionReceived` | Submission Service | Plagiarism Check, Notification | SubmissionID, TaskID, StudentID, IsLate | At-least-once |
| `SubmissionGraded` | Grading Service | Notification, Analytics | SubmissionID, Grade, GradedBy | Exactly-once |
| `PlagiarismDetected` | Plagiarism Service | Teacher Dashboard, Notification | SubmissionID, PlagiarismScore, SimilarTo | At-least-once |
| `DeadlineApproaching` | Task Scheduler | Notification | TaskID, StudentsWithoutSubmission, HoursLeft | At-least-once |

---

### Event Versioning

#### Стратегия версионирования

Используем **Schema Evolution** с обратной совместимостью:

```go
// Version 1
type DocumentCreatedV1 struct {
    DocumentID string
    Type       string
    CreatedBy  string
    Timestamp  time.Time
}

// Version 2 (added TemplateID)
type DocumentCreatedV2 struct {
    DocumentID string
    Type       string
    TemplateID string  // NEW FIELD
    CreatedBy  string
    Timestamp  time.Time
}

// Upcaster: V1 → V2
func UpcastDocumentCreatedV1ToV2(v1 DocumentCreatedV1) DocumentCreatedV2 {
    return DocumentCreatedV2{
        DocumentID: v1.DocumentID,
        Type:       v1.Type,
        TemplateID: "default",  // default value for old events
        CreatedBy:  v1.CreatedBy,
        Timestamp:  v1.Timestamp,
    }
}
```

**Best Practices**:
1. Всегда добавляем поле `EventVersion: int` в payload
2. Никогда не удаляем поля (только deprecate)
3. Новые поля должны быть optional или иметь default value
4. Храним upcasters для миграции старых событий

---

## Anti-Corruption Layer

### Интеграция с 1С

**Проблема**: 1С имеет свою доменную модель, несовместимую с нашей.

**Решение**: Anti-Corruption Layer (ACL) изолирует нашу систему от внешней.

#### Архитектура ACL

```
┌─────────────────────────────────────────────────────────┐
│ Integration Context (Bounded Context)                   │
│                                                           │
│  ┌──────────────┐         ┌────────────────────────┐    │
│  │ 1C Adapter   │────────▶│ ACL (Translator)       │    │
│  │ (REST Client)│         │                        │    │
│  └──────────────┘         │ ┌────────────────────┐ │    │
│                           │ │ Domain Translator  │ │    │
│                           │ │ 1C Model → Our     │ │    │
│                           │ │ Our Model → 1C     │ │    │
│                           │ └────────────────────┘ │    │
│                           │                        │    │
│                           │ ┌────────────────────┐ │    │
│                           │ │ Event Mapper       │ │    │
│                           │ │ 1C Events → Domain │ │    │
│                           │ └────────────────────┘ │    │
│                           └────────────────────────┘    │
│                                     │                    │
│                                     ▼                    │
│                           ┌────────────────────────┐    │
│                           │ Integration Repository │    │
│                           │ (stores mappings)      │    │
│                           └────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
                                     │
                                     ▼
                           ┌────────────────────────┐
                           │ Document Context       │
                           │ Workflow Context       │
                           │ (наша доменная модель) │
                           └────────────────────────┘
```

#### Пример Domain Translator

```go
// 1C Model (внешняя система)
type OneC_Document struct {
    DocID       string
    DocType     int  // 1=Curriculum, 2=Report, etc.
    Name        string
    Content     string  // XML string
    AuthorLogin string
    CreateDate  string  // "2025-01-15 10:30:00"
}

// Our Domain Model
type Document struct {
    ID        DocumentID
    Type      DocumentType
    Title     string
    Content   []byte  // JSON
    CreatedBy UserID
    CreatedAt time.Time
}

// ACL Translator
type OneCDocumentTranslator struct {
    userMapper UserMappingService  // maps 1C login → our UserID
}

func (t *OneCDocumentTranslator) ToDomain(oneC OneC_Document) (*Document, error) {
    // Translate Type
    docType, err := t.translateDocType(oneC.DocType)
    if err != nil {
        return nil, err
    }

    // Translate Content (XML → JSON)
    jsonContent, err := xmlToJSON(oneC.Content)
    if err != nil {
        return nil, err
    }

    // Map User
    userID, err := t.userMapper.GetUserIDByLogin(oneC.AuthorLogin)
    if err != nil {
        return nil, err
    }

    // Parse Date
    createdAt, err := time.Parse("2006-01-02 15:04:05", oneC.CreateDate)
    if err != nil {
        return nil, err
    }

    return &Document{
        ID:        DocumentID(oneC.DocID),
        Type:      docType,
        Title:     oneC.Name,
        Content:   jsonContent,
        CreatedBy: userID,
        CreatedAt: createdAt,
    }, nil
}

func (t *OneCDocumentTranslator) translateDocType(oneCType int) (DocumentType, error) {
    switch oneCType {
    case 1:
        return DocumentTypeCurriculum, nil
    case 2:
        return DocumentTypeReport, nil
    case 3:
        return DocumentTypeSchedule, nil
    default:
        return "", fmt.Errorf("unknown 1C document type: %d", oneCType)
    }
}
```

---

## Процесс моделирования

### Event Storming

**Event Storming** - collaborative workshop для discovery доменной модели.

#### Шаги Event Storming:

1. **Big Picture Event Storming** (4-8 часов)
   - Участники: Domain experts (методисты, секретари), разработчики, архитектор
   - Цель: Найти все Domain Events в системе
   - Output: Timeline событий на доске

2. **Process Modeling** (2-4 часа на процесс)
   - Группируем события по бизнес-процессам
   - Находим Commands, Aggregates, Policies
   - Output: Workflow диаграммы

3. **Software Design** (1-2 часа на Bounded Context)
   - Определяем границы Bounded Contexts
   - Проектируем Aggregates
   - Output: Context Map, Aggregate diagrams

#### Пример: Event Storming для Document Approval

```
[Domain Expert narrative]
"Методист создает учебный план → система генерирует документ →
отправляется на согласование администратору → администратор проверяет →
если все ок, утверждает → документ публикуется → уведомления рассылаются"

[Events discovered]
🟠 DocumentCreated
🟠 WorkflowStarted
🟠 NotificationSent (to admin)
🟠 WorkflowStepReviewed
🟠 WorkflowStepApproved
🟠 DocumentPublished
🟠 NotificationSent (to methodist + stakeholders)

[Aggregates identified]
📦 Document Aggregate
📦 Workflow Aggregate
📦 Notification Aggregate

[Commands]
🔵 CreateDocument
🔵 StartWorkflow
🔵 ApproveWorkflowStep
🔵 PublishDocument
🔵 SendNotification

[Policies (automated)]
💠 "When DocumentCreated → StartWorkflow (if type requires approval)"
💠 "When WorkflowStepApproved AND is last step → PublishDocument"
💠 "When DocumentPublished → SendNotification to stakeholders"
```

---

### Context Mapping

**Context Map** показывает отношения между Bounded Contexts.

#### Типы отношений:

1. **Partnership** (Партнерство): Две команды синхронизируют изменения
2. **Shared Kernel** (Общее ядро): Общие модели, используемые обоими контекстами
3. **Customer-Supplier** (Заказчик-Поставщик): Supplier предоставляет API для Customer
4. **Conformist** (Конформист): Downstream принимает модель Upstream без изменений
5. **Anti-Corruption Layer** (ACL): Downstream изолируется от Upstream через переводчик

#### Context Map нашей системы:

```
┌───────────────────────┐
│ Authentication        │
│ Context               │
│ (Upstream)            │
└───────────────────────┘
           │
           │ Customer-Supplier (REST API)
           │
           ▼
┌───────────────────────┐       ┌───────────────────────┐
│ Document Management   │──────▶│ Workflow Context      │
│ Context               │       │                       │
└───────────────────────┘       └───────────────────────┘
           │                              │
           │                              │
           │ Domain Events                │ Domain Events
           │                              │
           ▼                              ▼
┌───────────────────────┐       ┌───────────────────────┐
│ Search/Index Context  │       │ Notification Context  │
│ (read model)          │       │                       │
└───────────────────────┘       └───────────────────────┘

┌───────────────────────┐
│ Integration Context   │◀───── ACL
│                       │
└───────────────────────┘
           │
           │ Anti-Corruption Layer
           │
           ▼
┌───────────────────────┐
│ 1С (External System)  │
│                       │
└───────────────────────┘

┌───────────────────────┐       ┌───────────────────────┐
│ Schedule Context      │       │ Tasks Context         │
│                       │       │                       │
└───────────────────────┘       └───────────────────────┘
           │                              │
           └──────────────┬───────────────┘
                          │
                          │ Shared Kernel
                          │ (User, Role, Permission)
                          │
                          ▼
                 ┌───────────────────────┐
                 │ Shared Kernel         │
                 │ (common models)       │
                 └───────────────────────┘
```

---

## Рекомендации

### Best Practices

1. **Начинайте с событий**: Event Storming помогает найти доменную модель
2. **Говорите на языке бизнеса**: Ubiquitous Language в коде и документации
3. **Держите Aggregates маленькими**: Один Aggregate = одна транзакция
4. **Eventual Consistency**: Используйте Domain Events для связи между Aggregates
5. **Защищайте границы**: ACL для внешних систем (1С)
6. **Версионируйте события**: Event Schema Evolution для backward compatibility
7. **Документируйте решения**: ADR (Architecture Decision Records) для важных выборов

### Анти-паттерны (чего избегать)

❌ **Anemic Domain Model**: Domain объекты без логики (только getters/setters)
❌ **God Aggregate**: Огромный Aggregate с 1000+ entities
❌ **Bi-directional associations**: Циклические зависимости между Aggregates
❌ **Shared database**: Два Bounded Contexts читают/пишут одни таблицы
❌ **Distributed transactions**: 2PC/3PC между Aggregates
❌ **Leaky abstractions**: Domain Events содержат infrastructure детали (Kafka offsets)

---

📅 **Актуальность документа**
**Последнее обновление**: 2025-12-11
**Версия проекта**: 0.2.0
**Статус**: Актуальный
