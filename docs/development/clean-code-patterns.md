# 🧩 Clean Code Patterns & Best Practices

## 📋 Обзор

Руководство по применению паттернов проектирования, принципов SOLID и лучших практик в контексте модульной архитектуры для обеспечения высокого качества кода и готовности к переходу на микросервисы.

## 🎯 Принципы SOLID

### 1. **Single Responsibility Principle (SRP)**

#### ❌ Плохо
```go
// Нарушение SRP: класс делает слишком много
type UserService struct {
    db           *sql.DB
    emailService *EmailService
    logger       *Logger
}

func (s *UserService) CreateUser(user User) error {
    // Валидация
    if user.Email == "" {
        return errors.New("email required")
    }

    // Хеширование пароля
    hashedPassword := s.hashPassword(user.Password)
    user.Password = hashedPassword

    // Сохранение в БД
    query := "INSERT INTO users..."
    _, err := s.db.Exec(query, user.Name, user.Email, user.Password)
    if err != nil {
        return err
    }

    // Отправка email
    err = s.emailService.SendWelcomeEmail(user.Email)
    if err != nil {
        s.logger.Error("Failed to send welcome email", err)
    }

    // Логирование
    s.logger.Info("User created", user.ID)

    return nil
}
```

#### ✅ Хорошо
```go
// Разделение ответственностей
type UserRepository interface {
    Save(user User) error
    GetByID(id string) (*User, error)
    GetByEmail(email string) (*User, error)
}

type UserValidator struct{}

func (v *UserValidator) Validate(user User) error {
    if user.Email == "" {
        return ErrEmailRequired
    }
    if !isValidEmail(user.Email) {
        return ErrInvalidEmail
    }
    return nil
}

type PasswordHasher struct{}

func (h *PasswordHasher) Hash(password string) (string, error) {
    return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

type CreateUserUseCase struct {
    userRepo      UserRepository
    validator     UserValidator
    passwordHasher PasswordHasher
    eventBus      EventBus
}

func (uc *CreateUserUseCase) Execute(cmd CreateUserCommand) error {
    // Валидация
    user := User{
        Name:     cmd.Name,
        Email:    cmd.Email,
        Password: cmd.Password,
    }

    if err := uc.validator.Validate(user); err != nil {
        return err
    }

    // Хеширование пароля
    hashedPassword, err := uc.passwordHasher.Hash(user.Password)
    if err != nil {
        return err
    }
    user.Password = hashedPassword

    // Сохранение
    if err := uc.userRepo.Save(user); err != nil {
        return err
    }

    // Публикация события для отправки email
    event := UserCreated{
        UserID: user.ID,
        Email:  user.Email,
        Name:   user.Name,
    }
    return uc.eventBus.Publish(event)
}

// Отдельный обработчик для отправки email
type WelcomeEmailHandler struct {
    emailService EmailService
}

func (h *WelcomeEmailHandler) Handle(event UserCreated) error {
    return h.emailService.SendWelcomeEmail(event.Email, event.Name)
}
```

### 2. **Open/Closed Principle (OCP)**

#### ✅ Расширяемость через интерфейсы
```go
// Базовый интерфейс для уведомлений
type NotificationSender interface {
    Send(message Message, recipient string) error
}

// Реализации
type EmailSender struct {
    smtpClient SMTPClient
}

func (s *EmailSender) Send(message Message, recipient string) error {
    return s.smtpClient.SendEmail(recipient, message.Subject, message.Body)
}

type SMSSender struct {
    smsProvider SMSProvider
}

func (s *SMSSender) Send(message Message, recipient string) error {
    return s.smsProvider.SendSMS(recipient, message.Body)
}

type TelegramSender struct {
    botAPI TelegramBotAPI
}

func (s *TelegramSender) Send(message Message, recipient string) error {
    return s.botAPI.SendMessage(recipient, message.Body)
}

// Сервис уведомлений использует стратегию
type NotificationService struct {
    senders map[string]NotificationSender
}

func (s *NotificationService) SendNotification(channel string, message Message, recipient string) error {
    sender, exists := s.senders[channel]
    if !exists {
        return ErrUnsupportedChannel
    }
    return sender.Send(message, recipient)
}

// Легко добавить новый канал без изменения существующего кода
func (s *NotificationService) RegisterSender(channel string, sender NotificationSender) {
    s.senders[channel] = sender
}
```

### 3. **Liskov Substitution Principle (LSP)**

#### ✅ Правильное наследование поведения
```go
// Базовый интерфейс
type DocumentProcessor interface {
    Process(document Document) (*ProcessResult, error)
    Validate(document Document) error
}

// Все реализации должны соблюдать контракт
type PDFProcessor struct{}

func (p *PDFProcessor) Process(document Document) (*ProcessResult, error) {
    if err := p.Validate(document); err != nil {
        return nil, err
    }
    // PDF-специфичная обработка
    return &ProcessResult{Format: "PDF", ProcessedAt: time.Now()}, nil
}

func (p *PDFProcessor) Validate(document Document) error {
    if document.Type != "PDF" {
        return ErrInvalidDocumentType
    }
    return nil
}

type WordProcessor struct{}

func (p *WordProcessor) Process(document Document) (*ProcessResult, error) {
    if err := p.Validate(document); err != nil {
        return nil, err
    }
    // Word-специфичная обработка
    return &ProcessResult{Format: "DOCX", ProcessedAt: time.Now()}, nil
}

func (p *WordProcessor) Validate(document Document) error {
    if document.Type != "DOCX" {
        return ErrInvalidDocumentType
    }
    return nil
}

// Клиентский код может использовать любую реализацию
type DocumentService struct {
    processors map[string]DocumentProcessor
}

func (s *DocumentService) ProcessDocument(document Document) (*ProcessResult, error) {
    processor, exists := s.processors[document.Type]
    if !exists {
        return nil, ErrUnsupportedDocumentType
    }
    return processor.Process(document) // LSP соблюден
}
```

### 4. **Interface Segregation Principle (ISP)**

#### ❌ Плохо - толстый интерфейс
```go
type UserManager interface {
    CreateUser(user User) error
    UpdateUser(user User) error
    DeleteUser(id string) error
    GetUser(id string) (*User, error)
    SendWelcomeEmail(userID string) error
    GenerateReport(userID string) (*Report, error)
    BackupUserData(userID string) error
}
```

#### ✅ Хорошо - разделенные интерфейсы
```go
// Основные операции с пользователями
type UserRepository interface {
    Save(user User) error
    GetByID(id string) (*User, error)
    Delete(id string) error
}

// Операции поиска
type UserFinder interface {
    FindByEmail(email string) (*User, error)
    FindByRole(role string) ([]*User, error)
    Search(criteria SearchCriteria) ([]*User, error)
}

// Уведомления
type UserNotifier interface {
    SendWelcomeEmail(userID string) error
    SendPasswordReset(userID string) error
}

// Отчеты
type UserReporter interface {
    GenerateActivityReport(userID string) (*Report, error)
    GenerateUsageReport(userID string) (*Report, error)
}

// Клиенты используют только нужные интерфейсы
type CreateUserUseCase struct {
    userRepo UserRepository // только основные операции
    notifier UserNotifier   // только уведомления
}

type UserSearchService struct {
    userFinder UserFinder // только поиск
}
```

### 5. **Dependency Inversion Principle (DIP)**

#### ❌ Плохо - зависимость от конкретных типов
```go
type OrderService struct {
    mysqlRepo  *MySQLOrderRepository // зависимость от конкретной реализации
    emailSvc   *SMTPEmailService     // зависимость от конкретной реализации
}

func (s *OrderService) CreateOrder(order Order) error {
    // Жестко привязан к MySQL и SMTP
    err := s.mysqlRepo.Save(order)
    if err != nil {
        return err
    }
    return s.emailSvc.SendOrderConfirmation(order.CustomerEmail)
}
```

#### ✅ Хорошо - зависимость от абстракций
```go
// Абстракции
type OrderRepository interface {
    Save(order Order) error
    GetByID(id string) (*Order, error)
}

type EmailService interface {
    SendOrderConfirmation(email string, order Order) error
}

// Сервис зависит от абстракций
type OrderService struct {
    orderRepo    OrderRepository
    emailService EmailService
}

func (s *OrderService) CreateOrder(order Order) error {
    if err := s.orderRepo.Save(order); err != nil {
        return err
    }
    return s.emailService.SendOrderConfirmation(order.CustomerEmail, order)
}

// Конкретные реализации
type PostgreSQLOrderRepository struct {
    db *sql.DB
}

func (r *PostgreSQLOrderRepository) Save(order Order) error {
    // PostgreSQL-специфичная реализация
    return nil
}

type SendGridEmailService struct {
    apiKey string
}

func (s *SendGridEmailService) SendOrderConfirmation(email string, order Order) error {
    // SendGrid-специфичная реализация
    return nil
}

// DI контейнер
func NewOrderService(orderRepo OrderRepository, emailService EmailService) *OrderService {
    return &OrderService{
        orderRepo:    orderRepo,
        emailService: emailService,
    }
}
```

## 🏗️ Architectural Patterns

### 1. **Repository Pattern**

#### ✅ Правильная реализация
```go
// Доменная сущность
type Document struct {
    ID          DocumentID
    Title       string
    Content     string
    AuthorID    UserID
    Status      DocumentStatus
    CreatedAt   time.Time
    UpdatedAt   time.Time
    domainEvents []DomainEvent
}

// Интерфейс репозитория в доменном слое
type DocumentRepository interface {
    Save(document *Document) error
    GetByID(id DocumentID) (*Document, error)
    GetByAuthor(authorID UserID) ([]*Document, error)
    GetByStatus(status DocumentStatus) ([]*Document, error)
    Delete(id DocumentID) error
}

// Реализация в инфраструктурном слое
type PostgreSQLDocumentRepository struct {
    db     *sql.DB
    mapper DocumentMapper
}

func (r *PostgreSQLDocumentRepository) Save(document *Document) error {
    data := r.mapper.ToData(document)

    query := `
        INSERT INTO documents (id, title, content, author_id, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (id) DO UPDATE SET
            title = EXCLUDED.title,
            content = EXCLUDED.content,
            status = EXCLUDED.status,
            updated_at = EXCLUDED.updated_at
    `

    _, err := r.db.Exec(query,
        data.ID, data.Title, data.Content,
        data.AuthorID, data.Status,
        data.CreatedAt, data.UpdatedAt,
    )

    return err
}

func (r *PostgreSQLDocumentRepository) GetByID(id DocumentID) (*Document, error) {
    query := `
        SELECT id, title, content, author_id, status, created_at, updated_at
        FROM documents
        WHERE id = $1
    `

    var data DocumentData
    err := r.db.QueryRow(query, id.String()).Scan(
        &data.ID, &data.Title, &data.Content,
        &data.AuthorID, &data.Status,
        &data.CreatedAt, &data.UpdatedAt,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrDocumentNotFound
        }
        return nil, err
    }

    return r.mapper.ToDomain(data), nil
}

// Маппер для преобразования между доменными объектами и данными
type DocumentMapper struct{}

func (m *DocumentMapper) ToDomain(data DocumentData) *Document {
    return &Document{
        ID:        DocumentID(data.ID),
        Title:     data.Title,
        Content:   data.Content,
        AuthorID:  UserID(data.AuthorID),
        Status:    DocumentStatus(data.Status),
        CreatedAt: data.CreatedAt,
        UpdatedAt: data.UpdatedAt,
    }
}

func (m *DocumentMapper) ToData(document *Document) DocumentData {
    return DocumentData{
        ID:        document.ID.String(),
        Title:     document.Title,
        Content:   document.Content,
        AuthorID:  document.AuthorID.String(),
        Status:    string(document.Status),
        CreatedAt: document.CreatedAt,
        UpdatedAt: document.UpdatedAt,
    }
}
```

### 2. **Unit of Work Pattern**

#### ✅ Транзакционная обработка
```go
// Интерфейс Unit of Work
type UnitOfWork interface {
    Begin() error
    Commit() error
    Rollback() error
    Execute(fn func() error) error
}

// Реализация для PostgreSQL
type PostgreSQLUnitOfWork struct {
    db *sql.DB
    tx *sql.Tx
}

func (uow *PostgreSQLUnitOfWork) Begin() error {
    tx, err := uow.db.Begin()
    if err != nil {
        return err
    }
    uow.tx = tx
    return nil
}

func (uow *PostgreSQLUnitOfWork) Commit() error {
    if uow.tx == nil {
        return ErrNoActiveTransaction
    }
    return uow.tx.Commit()
}

func (uow *PostgreSQLUnitOfWork) Rollback() error {
    if uow.tx == nil {
        return ErrNoActiveTransaction
    }
    return uow.tx.Rollback()
}

func (uow *PostgreSQLUnitOfWork) Execute(fn func() error) error {
    if err := uow.Begin(); err != nil {
        return err
    }

    defer func() {
        if r := recover(); r != nil {
            uow.Rollback()
            panic(r)
        }
    }()

    if err := fn(); err != nil {
        uow.Rollback()
        return err
    }

    return uow.Commit()
}

// Использование в Use Case
type ApproveDocumentUseCase struct {
    documentRepo DocumentRepository
    workflowRepo WorkflowRepository
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

        // Проверяем права
        user, err := uc.userRepo.GetByID(cmd.ApproverID)
        if err != nil {
            return err
        }

        if !user.CanApprove(document) {
            return ErrInsufficientPermissions
        }

        // Одобряем документ
        if err := document.Approve(cmd.ApproverID); err != nil {
            return err
        }

        // Сохраняем документ
        if err := uc.documentRepo.Save(document); err != nil {
            return err
        }

        // Обновляем workflow
        workflow, err := uc.workflowRepo.GetByDocumentID(cmd.DocumentID)
        if err != nil {
            return err
        }

        if err := workflow.CompleteStep(cmd.ApproverID); err != nil {
            return err
        }

        if err := uc.workflowRepo.Save(workflow); err != nil {
            return err
        }

        // Публикуем события
        for _, event := range document.GetDomainEvents() {
            if err := uc.eventBus.Publish(event); err != nil {
                return err
            }
        }

        return nil
        // Если любая операция провалится, вся транзакция откатится
    })
}
```

### 3. **CQRS (Command Query Responsibility Segregation)**

#### ✅ Разделение команд и запросов
```go
// Commands (изменяют состояние)
type CreateDocumentCommand struct {
    Title     string
    Content   string
    AuthorID  UserID
    Type      DocumentType
}

type ApproveDocumentCommand struct {
    DocumentID DocumentID
    ApproverID UserID
    Comments   string
}

// Command Handlers
type CreateDocumentHandler struct {
    documentRepo DocumentRepository
    eventBus     EventBus
    unitOfWork   UnitOfWork
}

func (h *CreateDocumentHandler) Handle(cmd CreateDocumentCommand) error {
    return h.unitOfWork.Execute(func() error {
        document := NewDocument(
            NewDocumentID(),
            cmd.Title,
            cmd.Content,
            cmd.AuthorID,
            cmd.Type,
        )

        if err := h.documentRepo.Save(document); err != nil {
            return err
        }

        // Публикуем событие
        event := DocumentCreated{
            DocumentID: document.ID,
            AuthorID:   document.AuthorID,
            Title:      document.Title,
            CreatedAt:  document.CreatedAt,
        }

        return h.eventBus.Publish(event)
    })
}

// Queries (не изменяют состояние)
type GetDocumentQuery struct {
    DocumentID DocumentID
}

type SearchDocumentsQuery struct {
    AuthorID   *UserID
    Status     *DocumentStatus
    DateFrom   *time.Time
    DateTo     *time.Time
    Keyword    string
    PageSize   int
    PageNumber int
}

// Query Handlers (могут использовать read-optimized модели)
type DocumentQueryHandler struct {
    readRepo DocumentReadRepository
}

func (h *DocumentQueryHandler) GetDocument(query GetDocumentQuery) (*DocumentView, error) {
    return h.readRepo.GetDocumentView(query.DocumentID)
}

func (h *DocumentQueryHandler) SearchDocuments(query SearchDocumentsQuery) (*DocumentListView, error) {
    return h.readRepo.SearchDocuments(DocumentSearchCriteria{
        AuthorID:   query.AuthorID,
        Status:     query.Status,
        DateFrom:   query.DateFrom,
        DateTo:     query.DateTo,
        Keyword:    query.Keyword,
        PageSize:   query.PageSize,
        PageNumber: query.PageNumber,
    })
}

// Read Models (оптимизированы для чтения)
type DocumentView struct {
    ID           string    `json:"id"`
    Title        string    `json:"title"`
    AuthorName   string    `json:"author_name"`
    Status       string    `json:"status"`
    CreatedAt    time.Time `json:"created_at"`
    ApprovalSteps []ApprovalStepView `json:"approval_steps"`
}

type DocumentListView struct {
    Documents  []DocumentSummaryView `json:"documents"`
    TotalCount int                  `json:"total_count"`
    PageSize   int                  `json:"page_size"`
    PageNumber int                  `json:"page_number"`
}

// Отдельный репозиторий для чтения (может использовать денормализованные таблицы)
type DocumentReadRepository interface {
    GetDocumentView(id DocumentID) (*DocumentView, error)
    SearchDocuments(criteria DocumentSearchCriteria) (*DocumentListView, error)
}
```

### 4. **Event Sourcing Pattern**

#### ✅ Сохранение событий вместо состояния
```go
// Доменные события
type DomainEvent interface {
    GetAggregateID() string
    GetEventType() string
    GetOccurredAt() time.Time
    GetVersion() int
}

type DocumentCreated struct {
    AggregateID string    `json:"aggregate_id"`
    Title       string    `json:"title"`
    AuthorID    string    `json:"author_id"`
    Type        string    `json:"type"`
    OccurredAt  time.Time `json:"occurred_at"`
    Version     int       `json:"version"`
}

func (e DocumentCreated) GetAggregateID() string { return e.AggregateID }
func (e DocumentCreated) GetEventType() string   { return "DocumentCreated" }
func (e DocumentCreated) GetOccurredAt() time.Time { return e.OccurredAt }
func (e DocumentCreated) GetVersion() int        { return e.Version }

type DocumentApproved struct {
    AggregateID string    `json:"aggregate_id"`
    ApproverID  string    `json:"approver_id"`
    Comments    string    `json:"comments"`
    OccurredAt  time.Time `json:"occurred_at"`
    Version     int       `json:"version"`
}

// Event Store
type EventStore interface {
    SaveEvents(aggregateID string, events []DomainEvent, expectedVersion int) error
    GetEvents(aggregateID string) ([]DomainEvent, error)
    GetEventsFromVersion(aggregateID string, version int) ([]DomainEvent, error)
}

// Aggregate Root с Event Sourcing
type Document struct {
    id           DocumentID
    version      int
    uncommittedEvents []DomainEvent

    // Текущее состояние (восстанавливается из событий)
    title        string
    content      string
    authorID     UserID
    status       DocumentStatus
    createdAt    time.Time
    updatedAt    time.Time
}

// Восстановление состояния из событий
func (d *Document) LoadFromHistory(events []DomainEvent) {
    for _, event := range events {
        d.Apply(event)
        d.version = event.GetVersion()
    }
}

// Применение событий к агрегату
func (d *Document) Apply(event DomainEvent) {
    switch e := event.(type) {
    case DocumentCreated:
        d.id = DocumentID(e.AggregateID)
        d.title = e.Title
        d.authorID = UserID(e.AuthorID)
        d.status = DocumentStatusDraft
        d.createdAt = e.OccurredAt
        d.updatedAt = e.OccurredAt

    case DocumentApproved:
        d.status = DocumentStatusApproved
        d.updatedAt = e.OccurredAt

    case DocumentRejected:
        d.status = DocumentStatusRejected
        d.updatedAt = e.OccurredAt
    }
}

// Добавление новых событий
func (d *Document) Approve(approverID UserID, comments string) error {
    if d.status != DocumentStatusPending {
        return ErrDocumentNotPending
    }

    event := DocumentApproved{
        AggregateID: d.id.String(),
        ApproverID:  approverID.String(),
        Comments:    comments,
        OccurredAt:  time.Now(),
        Version:     d.version + 1,
    }

    d.Apply(event)
    d.uncommittedEvents = append(d.uncommittedEvents, event)

    return nil
}

// Repository для Event Sourcing
type EventSourcingDocumentRepository struct {
    eventStore EventStore
}

func (r *EventSourcingDocumentRepository) Save(document *Document) error {
    events := document.GetUncommittedEvents()
    if len(events) == 0 {
        return nil
    }

    err := r.eventStore.SaveEvents(
        document.ID.String(),
        events,
        document.version - len(events), // expected version
    )

    if err != nil {
        return err
    }

    document.MarkEventsAsCommitted()
    return nil
}

func (r *EventSourcingDocumentRepository) GetByID(id DocumentID) (*Document, error) {
    events, err := r.eventStore.GetEvents(id.String())
    if err != nil {
        return nil, err
    }

    if len(events) == 0 {
        return nil, ErrDocumentNotFound
    }

    document := &Document{}
    document.LoadFromHistory(events)

    return document, nil
}
```

## 🔧 Design Patterns

### 1. **Factory Pattern**

#### ✅ Создание документов разных типов
```go
// Интерфейс документа
type Document interface {
    GetID() DocumentID
    GetType() DocumentType
    Validate() error
    Process() error
}

// Конкретные типы документов
type ServiceNote struct {
    ID       DocumentID
    Title    string
    Content  string
    AuthorID UserID
}

func (d *ServiceNote) GetID() DocumentID { return d.ID }
func (d *ServiceNote) GetType() DocumentType { return DocumentTypeServiceNote }

func (d *ServiceNote) Validate() error {
    if d.Title == "" {
        return ErrTitleRequired
    }
    return nil
}

func (d *ServiceNote) Process() error {
    // Специфичная обработка служебной записки
    return nil
}

type Order struct {
    ID       DocumentID
    Number   string
    Content  string
    AuthorID UserID
}

func (d *Order) GetID() DocumentID { return d.ID }
func (d *Order) GetType() DocumentType { return DocumentTypeOrder }

func (d *Order) Validate() error {
    if d.Number == "" {
        return ErrOrderNumberRequired
    }
    return nil
}

func (d *Order) Process() error {
    // Специфичная обработка приказа
    return nil
}

// Фабрика документов
type DocumentFactory interface {
    CreateDocument(docType DocumentType, data map[string]interface{}) (Document, error)
}

type ConcreteDocumentFactory struct{}

func (f *ConcreteDocumentFactory) CreateDocument(docType DocumentType, data map[string]interface{}) (Document, error) {
    switch docType {
    case DocumentTypeServiceNote:
        return &ServiceNote{
            ID:       NewDocumentID(),
            Title:    data["title"].(string),
            Content:  data["content"].(string),
            AuthorID: UserID(data["author_id"].(string)),
        }, nil

    case DocumentTypeOrder:
        return &Order{
            ID:       NewDocumentID(),
            Number:   data["number"].(string),
            Content:  data["content"].(string),
            AuthorID: UserID(data["author_id"].(string)),
        }, nil

    default:
        return nil, ErrUnsupportedDocumentType
    }
}

// Использование
type CreateDocumentUseCase struct {
    factory      DocumentFactory
    documentRepo DocumentRepository
}

func (uc *CreateDocumentUseCase) Execute(cmd CreateDocumentCommand) error {
    document, err := uc.factory.CreateDocument(cmd.Type, map[string]interface{}{
        "title":     cmd.Title,
        "content":   cmd.Content,
        "author_id": cmd.AuthorID.String(),
        "number":    cmd.Number,
    })
    if err != nil {
        return err
    }

    if err := document.Validate(); err != nil {
        return err
    }

    if err := document.Process(); err != nil {
        return err
    }

    return uc.documentRepo.Save(document)
}
```

### 2. **Strategy Pattern**

#### ✅ Стратегии обработки файлов
```go
// Стратегия обработки файла
type FileProcessor interface {
    Process(file File) (*ProcessResult, error)
    CanProcess(fileType string) bool
}

// Конкретные стратегии
type PDFProcessor struct{}

func (p *PDFProcessor) CanProcess(fileType string) bool {
    return fileType == "application/pdf"
}

func (p *PDFProcessor) Process(file File) (*ProcessResult, error) {
    // PDF-специфичная обработка
    return &ProcessResult{
        ThumbnailURL: generatePDFThumbnail(file),
        TextContent:  extractPDFText(file),
        PageCount:    getPDFPageCount(file),
    }, nil
}

type ImageProcessor struct{}

func (p *ImageProcessor) CanProcess(fileType string) bool {
    return strings.HasPrefix(fileType, "image/")
}

func (p *ImageProcessor) Process(file File) (*ProcessResult, error) {
    // Обработка изображений
    return &ProcessResult{
        ThumbnailURL: generateImageThumbnail(file),
        Dimensions:   getImageDimensions(file),
        ColorSpace:   getColorSpace(file),
    }, nil
}

type WordProcessor struct{}

func (p *WordProcessor) CanProcess(fileType string) bool {
    return fileType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
}

func (p *WordProcessor) Process(file File) (*ProcessResult, error) {
    // Обработка Word документов
    return &ProcessResult{
        ThumbnailURL: generateWordThumbnail(file),
        TextContent:  extractWordText(file),
        WordCount:    getWordCount(file),
    }, nil
}

// Контекст использования стратегий
type FileProcessingService struct {
    processors []FileProcessor
}

func (s *FileProcessingService) RegisterProcessor(processor FileProcessor) {
    s.processors = append(s.processors, processor)
}

func (s *FileProcessingService) ProcessFile(file File) (*ProcessResult, error) {
    for _, processor := range s.processors {
        if processor.CanProcess(file.MimeType) {
            return processor.Process(file)
        }
    }

    // Стратегия по умолчанию
    return &ProcessResult{
        FileName: file.Name,
        FileSize: file.Size,
    }, nil
}

// Инициализация
func NewFileProcessingService() *FileProcessingService {
    service := &FileProcessingService{}

    // Регистрируем стратегии
    service.RegisterProcessor(&PDFProcessor{})
    service.RegisterProcessor(&ImageProcessor{})
    service.RegisterProcessor(&WordProcessor{})

    return service
}
```

### 3. **Observer Pattern (Event-Driven)**

#### ✅ Система событий для модулей
```go
// Система событий
type EventBus interface {
    Subscribe(eventType string, handler EventHandler) error
    Unsubscribe(eventType string, handler EventHandler) error
    Publish(event DomainEvent) error
}

type EventHandler interface {
    Handle(event DomainEvent) error
    GetHandlerName() string
}

// Реализация Event Bus
type InMemoryEventBus struct {
    handlers map[string][]EventHandler
    mutex    sync.RWMutex
}

func NewInMemoryEventBus() *InMemoryEventBus {
    return &InMemoryEventBus{
        handlers: make(map[string][]EventHandler),
    }
}

func (bus *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
    bus.mutex.Lock()
    defer bus.mutex.Unlock()

    bus.handlers[eventType] = append(bus.handlers[eventType], handler)
    return nil
}

func (bus *InMemoryEventBus) Publish(event DomainEvent) error {
    bus.mutex.RLock()
    handlers := bus.handlers[event.GetEventType()]
    bus.mutex.RUnlock()

    for _, handler := range handlers {
        if err := handler.Handle(event); err != nil {
            // Логируем ошибку, но продолжаем обработку других хэндлеров
            log.Printf("Error handling event %s with handler %s: %v",
                event.GetEventType(), handler.GetHandlerName(), err)
        }
    }

    return nil
}

// Конкретные обработчики событий
type EmailNotificationHandler struct {
    emailService EmailService
}

func (h *EmailNotificationHandler) GetHandlerName() string {
    return "EmailNotificationHandler"
}

func (h *EmailNotificationHandler) Handle(event DomainEvent) error {
    switch e := event.(type) {
    case DocumentCreated:
        return h.handleDocumentCreated(e)
    case DocumentApproved:
        return h.handleDocumentApproved(e)
    case UserRegistered:
        return h.handleUserRegistered(e)
    }
    return nil
}

func (h *EmailNotificationHandler) handleDocumentCreated(event DocumentCreated) error {
    // Отправляем уведомление о создании документа
    return h.emailService.SendDocumentCreatedNotification(
        event.AuthorID,
        event.Title,
    )
}

type WorkflowHandler struct {
    workflowService WorkflowService
}

func (h *WorkflowHandler) GetHandlerName() string {
    return "WorkflowHandler"
}

func (h *WorkflowHandler) Handle(event DomainEvent) error {
    switch e := event.(type) {
    case DocumentCreated:
        return h.startApprovalWorkflow(e)
    case DocumentApproved:
        return h.continueWorkflow(e)
    }
    return nil
}

func (h *WorkflowHandler) startApprovalWorkflow(event DocumentCreated) error {
    return h.workflowService.StartDocumentApprovalWorkflow(
        event.AggregateID,
        event.AuthorID,
    )
}

// Регистрация обработчиков
func SetupEventHandlers(eventBus EventBus, container *DIContainer) {
    // Email уведомления
    emailHandler := &EmailNotificationHandler{
        emailService: container.EmailService,
    }
    eventBus.Subscribe("DocumentCreated", emailHandler)
    eventBus.Subscribe("DocumentApproved", emailHandler)
    eventBus.Subscribe("UserRegistered", emailHandler)

    // Workflow
    workflowHandler := &WorkflowHandler{
        workflowService: container.WorkflowService,
    }
    eventBus.Subscribe("DocumentCreated", workflowHandler)
    eventBus.Subscribe("DocumentApproved", workflowHandler)

    // Аудит
    auditHandler := &AuditHandler{
        auditService: container.AuditService,
    }
    eventBus.Subscribe("DocumentCreated", auditHandler)
    eventBus.Subscribe("DocumentApproved", auditHandler)
    eventBus.Subscribe("DocumentDeleted", auditHandler)
}
```

### 4. **Decorator Pattern**

#### ✅ Middleware для обработки запросов
```go
// Базовый интерфейс обработчика
type RequestHandler interface {
    Handle(ctx context.Context, request Request) (Response, error)
}

// Базовый обработчик
type BaseDocumentHandler struct {
    documentService DocumentService
}

func (h *BaseDocumentHandler) Handle(ctx context.Context, request Request) (Response, error) {
    // Основная логика обработки
    return h.documentService.ProcessRequest(ctx, request)
}

// Декораторы (middleware)
type LoggingDecorator struct {
    handler RequestHandler
    logger  Logger
}

func (d *LoggingDecorator) Handle(ctx context.Context, request Request) (Response, error) {
    start := time.Now()

    d.logger.Info("Processing request", map[string]interface{}{
        "request_id": request.GetID(),
        "user_id":    request.GetUserID(),
        "action":     request.GetAction(),
    })

    response, err := d.handler.Handle(ctx, request)

    duration := time.Since(start)

    if err != nil {
        d.logger.Error("Request failed", map[string]interface{}{
            "request_id": request.GetID(),
            "error":      err.Error(),
            "duration":   duration,
        })
    } else {
        d.logger.Info("Request completed", map[string]interface{}{
            "request_id": request.GetID(),
            "duration":   duration,
        })
    }

    return response, err
}

type AuthorizationDecorator struct {
    handler      RequestHandler
    authService  AuthService
}

func (d *AuthorizationDecorator) Handle(ctx context.Context, request Request) (Response, error) {
    // Проверяем авторизацию
    if !d.authService.IsAuthorized(ctx, request.GetUserID(), request.GetAction()) {
        return nil, ErrUnauthorized
    }

    return d.handler.Handle(ctx, request)
}

type ValidationDecorator struct {
    handler   RequestHandler
    validator RequestValidator
}

func (d *ValidationDecorator) Handle(ctx context.Context, request Request) (Response, error) {
    // Валидируем запрос
    if err := d.validator.Validate(request); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    return d.handler.Handle(ctx, request)
}

type MetricsDecorator struct {
    handler         RequestHandler
    metricsCollector MetricsCollector
}

func (d *MetricsDecorator) Handle(ctx context.Context, request Request) (Response, error) {
    start := time.Now()

    response, err := d.handler.Handle(ctx, request)

    duration := time.Since(start)

    d.metricsCollector.RecordRequest(
        request.GetAction(),
        duration,
        err != nil,
    )

    return response, err
}

// Построение цепочки декораторов
func BuildRequestHandler(
    documentService DocumentService,
    authService AuthService,
    validator RequestValidator,
    logger Logger,
    metrics MetricsCollector,
) RequestHandler {
    // Базовый обработчик
    handler := &BaseDocumentHandler{
        documentService: documentService,
    }

    // Оборачиваем в декораторы (порядок важен!)
    handler = &ValidationDecorator{
        handler:   handler,
        validator: validator,
    }

    handler = &AuthorizationDecorator{
        handler:     handler,
        authService: authService,
    }

    handler = &LoggingDecorator{
        handler: handler,
        logger:  logger,
    }

    handler = &MetricsDecorator{
        handler:         handler,
        metricsCollector: metrics,
    }

    return handler
}
```

## 🧪 Testing Patterns

### 1. **Test Doubles (Mocks, Stubs, Fakes)**

#### ✅ Правильное использование test doubles
```go
// Интерфейс для мокирования
type UserRepository interface {
    Save(user User) error
    GetByID(id UserID) (*User, error)
    GetByEmail(email string) (*User, error)
}

// Fake implementation для интеграционных тестов
type FakeUserRepository struct {
    users map[string]*User
    mutex sync.RWMutex
}

func NewFakeUserRepository() *FakeUserRepository {
    return &FakeUserRepository{
        users: make(map[string]*User),
    }
}

func (r *FakeUserRepository) Save(user User) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()

    r.users[user.ID.String()] = &user
    return nil
}

func (r *FakeUserRepository) GetByID(id UserID) (*User, error) {
    r.mutex.RLock()
    defer r.mutex.RUnlock()

    user, exists := r.users[id.String()]
    if !exists {
        return nil, ErrUserNotFound
    }

    return user, nil
}

func (r *FakeUserRepository) GetByEmail(email string) (*User, error) {
    r.mutex.RLock()
    defer r.mutex.RUnlock()

    for _, user := range r.users {
        if user.Email == email {
            return user, nil
        }
    }

    return nil, ErrUserNotFound
}

// Mock для unit тестов (с помощью testify/mock)
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Save(user User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepository) GetByID(id UserID) (*User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*User, error) {
    args := m.Called(email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

// Unit тест с mock
func TestCreateUserUseCase_Execute_Success(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    mockEventBus := &MockEventBus{}
    mockHasher := &MockPasswordHasher{}

    useCase := NewCreateUserUseCase(mockRepo, mockEventBus, mockHasher)

    hashedPassword := "hashed_password"
    mockHasher.On("Hash", "password").Return(hashedPassword, nil)
    mockRepo.On("Save", mock.MatchedBy(func(user User) bool {
        return user.Email == "test@example.com" && user.Password == hashedPassword
    })).Return(nil)
    mockEventBus.On("Publish", mock.AnythingOfType("UserCreated")).Return(nil)

    cmd := CreateUserCommand{
        Name:     "Test User",
        Email:    "test@example.com",
        Password: "password",
    }

    // Act
    err := useCase.Execute(cmd)

    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
    mockEventBus.AssertExpectations(t)
    mockHasher.AssertExpectations(t)
}

// Интеграционный тест с fake
func TestCreateUserUseCase_Integration(t *testing.T) {
    // Arrange
    fakeRepo := NewFakeUserRepository()
    fakeEventBus := NewFakeEventBus()
    realHasher := &BCryptPasswordHasher{}

    useCase := NewCreateUserUseCase(fakeRepo, fakeEventBus, realHasher)

    cmd := CreateUserCommand{
        Name:     "Test User",
        Email:    "test@example.com",
        Password: "password",
    }

    // Act
    err := useCase.Execute(cmd)

    // Assert
    assert.NoError(t, err)

    // Проверяем, что пользователь был сохранен
    user, err := fakeRepo.GetByEmail("test@example.com")
    assert.NoError(t, err)
    assert.Equal(t, "Test User", user.Name)
    assert.NotEqual(t, "password", user.Password) // Пароль должен быть захеширован

    // Проверяем, что событие было опубликовано
    events := fakeEventBus.GetPublishedEvents()
    assert.Len(t, events, 1)
    assert.IsType(t, UserCreated{}, events[0])
}
```

### 2. **Test Builders**

#### ✅ Удобное создание тестовых данных
```go
// Builder для создания тестовых пользователей
type UserBuilder struct {
    user User
}

func NewUserBuilder() *UserBuilder {
    return &UserBuilder{
        user: User{
            ID:        NewUserID(),
            Name:      "Test User",
            Email:     "test@example.com",
            Password:  "hashed_password",
            Role:      RoleUser,
            Status:    UserStatusActive,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
    }
}

func (b *UserBuilder) WithID(id UserID) *UserBuilder {
    b.user.ID = id
    return b
}

func (b *UserBuilder) WithName(name string) *UserBuilder {
    b.user.Name = name
    return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    b.user.Email = email
    return b
}

func (b *UserBuilder) WithRole(role UserRole) *UserBuilder {
    b.user.Role = role
    return b
}

func (b *UserBuilder) WithStatus(status UserStatus) *UserBuilder {
    b.user.Status = status
    return b
}

func (b *UserBuilder) AsAdmin() *UserBuilder {
    b.user.Role = RoleAdmin
    return b
}

func (b *UserBuilder) AsInactive() *UserBuilder {
    b.user.Status = UserStatusInactive
    return b
}

func (b *UserBuilder) Build() User {
    return b.user
}

// Использование в тестах
func TestSomeFunction(t *testing.T) {
    // Простое создание
    user := NewUserBuilder().Build()

    // С кастомными параметрами
    admin := NewUserBuilder().
        WithName("Admin User").
        WithEmail("admin@example.com").
        AsAdmin().
        Build()

    // Для конкретного случая
    inactiveUser := NewUserBuilder().
        WithEmail("inactive@example.com").
        AsInactive().
        Build()
}

// Builder для документов
type DocumentBuilder struct {
    document Document
}

func NewDocumentBuilder() *DocumentBuilder {
    return &DocumentBuilder{
        document: Document{
            ID:        NewDocumentID(),
            Title:     "Test Document",
            Content:   "Test content",
            AuthorID:  NewUserID(),
            Status:    DocumentStatusDraft,
            Type:      DocumentTypeServiceNote,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
    }
}

func (b *DocumentBuilder) WithTitle(title string) *DocumentBuilder {
    b.document.Title = title
    return b
}

func (b *DocumentBuilder) WithAuthor(authorID UserID) *DocumentBuilder {
    b.document.AuthorID = authorID
    return b
}

func (b *DocumentBuilder) WithStatus(status DocumentStatus) *DocumentBuilder {
    b.document.Status = status
    return b
}

func (b *DocumentBuilder) WithType(docType DocumentType) *DocumentBuilder {
    b.document.Type = docType
    return b
}

func (b *DocumentBuilder) AsPending() *DocumentBuilder {
    b.document.Status = DocumentStatusPending
    return b
}

func (b *DocumentBuilder) AsApproved() *DocumentBuilder {
    b.document.Status = DocumentStatusApproved
    return b
}

func (b *DocumentBuilder) AsOrder() *DocumentBuilder {
    b.document.Type = DocumentTypeOrder
    return b
}

func (b *DocumentBuilder) Build() Document {
    return b.document
}
```

### 3. **Table-Driven Tests**

#### ✅ Тестирование множества сценариев
```go
func TestDocumentValidator_Validate(t *testing.T) {
    validator := NewDocumentValidator()

    tests := []struct {
        name        string
        document    Document
        expectedErr error
    }{
        {
            name: "valid document",
            document: NewDocumentBuilder().
                WithTitle("Valid Title").
                WithContent("Valid content").
                Build(),
            expectedErr: nil,
        },
        {
            name: "empty title",
            document: NewDocumentBuilder().
                WithTitle("").
                Build(),
            expectedErr: ErrTitleRequired,
        },
        {
            name: "title too long",
            document: NewDocumentBuilder().
                WithTitle(strings.Repeat("a", 256)).
                Build(),
            expectedErr: ErrTitleTooLong,
        },
        {
            name: "empty content",
            document: NewDocumentBuilder().
                WithContent("").
                Build(),
            expectedErr: ErrContentRequired,
        },
        {
            name: "content too short",
            document: NewDocumentBuilder().
                WithContent("ab").
                Build(),
            expectedErr: ErrContentTooShort,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.Validate(tt.document)

            if tt.expectedErr == nil {
                assert.NoError(t, err)
            } else {
                assert.ErrorIs(t, err, tt.expectedErr)
            }
        })
    }
}

func TestUserPermissions_CanApprove(t *testing.T) {
    tests := []struct {
        name     string
        user     User
        document Document
        expected bool
    }{
        {
            name: "admin can approve any document",
            user: NewUserBuilder().AsAdmin().Build(),
            document: NewDocumentBuilder().Build(),
            expected: true,
        },
        {
            name: "user cannot approve own document",
            user: NewUserBuilder().WithID(UserID("user-1")).Build(),
            document: NewDocumentBuilder().WithAuthor(UserID("user-1")).Build(),
            expected: false,
        },
        {
            name: "methodist can approve service notes",
            user: NewUserBuilder().WithRole(RoleMethodist).Build(),
            document: NewDocumentBuilder().WithType(DocumentTypeServiceNote).Build(),
            expected: true,
        },
        {
            name: "methodist cannot approve orders",
            user: NewUserBuilder().WithRole(RoleMethodist).Build(),
            document: NewDocumentBuilder().AsOrder().Build(),
            expected: false,
        },
        {
            name: "secretary can approve orders",
            user: NewUserBuilder().WithRole(RoleSecretary).Build(),
            document: NewDocumentBuilder().AsOrder().Build(),
            expected: true,
        },
        {
            name: "regular user cannot approve documents",
            user: NewUserBuilder().WithRole(RoleUser).Build(),
            document: NewDocumentBuilder().Build(),
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.user.CanApprove(tt.document)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## 📏 Performance Patterns

### 1. **Caching Patterns**

#### ✅ Repository with caching
```go
// Кеширующий декоратор для репозитория
type CachedUserRepository struct {
    baseRepo UserRepository
    cache    Cache
    ttl      time.Duration
}

func NewCachedUserRepository(baseRepo UserRepository, cache Cache, ttl time.Duration) *CachedUserRepository {
    return &CachedUserRepository{
        baseRepo: baseRepo,
        cache:    cache,
        ttl:      ttl,
    }
}

func (r *CachedUserRepository) GetByID(id UserID) (*User, error) {
    cacheKey := fmt.Sprintf("user:%s", id.String())

    // Пытаемся получить из кеша
    if cached, err := r.cache.Get(cacheKey); err == nil {
        var user User
        if err := json.Unmarshal(cached, &user); err == nil {
            return &user, nil
        }
    }

    // Получаем из основного репозитория
    user, err := r.baseRepo.GetByID(id)
    if err != nil {
        return nil, err
    }

    // Сохраняем в кеш
    if data, err := json.Marshal(user); err == nil {
        r.cache.Set(cacheKey, data, r.ttl)
    }

    return user, nil
}

func (r *CachedUserRepository) Save(user User) error {
    if err := r.baseRepo.Save(user); err != nil {
        return err
    }

    // Инвалидируем кеш
    cacheKey := fmt.Sprintf("user:%s", user.ID.String())
    r.cache.Delete(cacheKey)

    return nil
}
```

### 2. **Connection Pooling**

#### ✅ Database connection management
```go
type DatabaseConfig struct {
    Host            string
    Port            int
    Database        string
    Username        string
    Password        string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}

func NewDatabaseConnection(config DatabaseConfig) (*sql.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        config.Host, config.Port, config.Username, config.Password, config.Database)

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }

    // Настройка пула соединений
    db.SetMaxOpenConns(config.MaxOpenConns)
    db.SetMaxIdleConns(config.MaxIdleConns)
    db.SetConnMaxLifetime(config.ConnMaxLifetime)
    db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

    // Проверка соединения
    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}
```

Эти паттерны и практики обеспечивают:
- ✅ Высокое качество кода
- ✅ Легкость тестирования и поддержки
- ✅ Готовность к масштабированию
- ✅ Соответствие принципам SOLID
- ✅ Производительность и надежность
---

**📅 Актуальность документа**  
**Последнее обновление**: 2025-01-15  
**Версия проекта**: 0.2.0  
**Статус**: Актуальный

