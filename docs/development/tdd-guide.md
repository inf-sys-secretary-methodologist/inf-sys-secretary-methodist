# Test-Driven Development (TDD) Guide

## Оглавление
- [Почему TDD?](#почему-tdd)
- [Red-Green-Refactor Cycle](#red-green-refactor-cycle)
- [TDD для Domain-Driven Design](#tdd-для-domain-driven-design)
- [Outside-in TDD](#outside-in-tdd)
- [Test Doubles Strategy](#test-doubles-strategy)
- [Best Practices](#best-practices)

---

## Почему TDD?

### Обоснование выбора TDD для проекта

Test-Driven Development выбран как основная методология разработки по следующим причинам:

#### ✅ Преимущества для нашего проекта:

1. **Сложная бизнес-логика**
   - Workflow согласования с 4 шагами и SLA
   - RBAC с матрицей разрешений 5x5x8
   - Schedule conflict detection
   - Plagiarism checking
   - **TDD гарантирует**: Все edge cases покрыты тестами ДО написания кода

2. **Clean Architecture + DDD**
   - Доменная логика изолирована от infrastructure
   - Aggregates с инвариантами требуют тщательного тестирования
   - **TDD помогает**: Проектировать чистые domain interfaces

3. **Высокие требования к качеству**
   - Учебная система: ошибки влияют на расписания, оценки студентов
   - Target: 80%+ code coverage для domain layer
   - **TDD обеспечивает**: Coverage естественным образом

4. **Refactoring confidence**
   - Модульный монолит → микросервисы в будущем
   - Постоянная эволюция требований
   - **TDD дает**: Safety net при рефакторинге

#### ⚖️ Trade-offs:

| Преимущество | Недостаток | Решение |
|--------------|------------|---------|
| Высокое покрытие тестами | Больше времени на тесты | Автоматизация, test utilities |
| Чистый design | Learning curve для новичков | Pair programming, code reviews |
| Быстрая обратная связь | CI может быть медленным | Parallel test execution |
| Уверенность при рефакторинге | Over-testing (tests become brittle) | Focus on behavior, not implementation |

---

## Red-Green-Refactor Cycle

### Основной цикл TDD

```
┌──────────────┐
│    RED       │  1. Write failing test
│  (Write test)│     - Test describes desired behavior
└──────┬───────┘     - Test MUST fail (verify it catches bugs)
       │
       ▼
┌──────────────┐
│   GREEN      │  2. Make test pass
│ (Make it work)│    - Write MINIMAL code to pass test
└──────┬───────┘     - Don't optimize yet
       │
       ▼
┌──────────────┐
│  REFACTOR    │  3. Improve code
│ (Make it clean)│   - Remove duplication
└──────┬───────┘     - Improve readability
       │             - Tests still pass
       │
       └─────────────┐
                     │
                     ▼
            ┌────────────────┐
            │  Next feature  │
            └────────────────┘
```

### Пример: Создание Document Aggregate

#### ❶ RED: Write failing test

```go
// internal/modules/documents/domain/entities/document_test.go

func TestDocument_Create_ValidInputs_Success(t *testing.T) {
    // Arrange
    title := "Учебный план по математике"
    content := []byte(`{"program": "Математика", "credits": 120}`)
    templateID := TemplateID("tpl_curriculum_001")
    createdBy := UserID("user_123")

    // Act
    doc, err := NewDocument(title, content, templateID, createdBy)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, doc)
    assert.NotEmpty(t, doc.ID)
    assert.Equal(t, title, doc.Title)
    assert.Equal(t, content, doc.Content)
    assert.Equal(t, DocumentStatusDraft, doc.Status)
    assert.Len(t, doc.Versions, 1) // первая версия создается автоматически
}
```

**Результат**: ❌ Test fails (функция `NewDocument` не существует)

#### ❷ GREEN: Make it pass

```go
// internal/modules/documents/domain/entities/document.go

type Document struct {
    ID        DocumentID
    Title     string
    Content   []byte
    TemplateID TemplateID
    Status    DocumentStatus
    Versions  []Version
    CreatedBy UserID
    CreatedAt time.Time
}

func NewDocument(title string, content []byte, templateID TemplateID, createdBy UserID) (*Document, error) {
    now := time.Now()

    doc := &Document{
        ID:        DocumentID(uuid.New().String()),
        Title:     title,
        Content:   content,
        TemplateID: templateID,
        Status:    DocumentStatusDraft,
        CreatedBy: createdBy,
        CreatedAt: now,
        Versions: []Version{
            {
                ID:        VersionID(uuid.New().String()),
                Content:   content,
                CreatedBy: createdBy,
                CreatedAt: now,
            },
        },
    }

    return doc, nil
}
```

**Результат**: ✅ Test passes

#### ❸ REFACTOR: Improve code

**Проблема**: Нет валидации, отсутствуют инварианты

```go
func TestDocument_Create_EmptyTitle_ReturnsError(t *testing.T) {
    // Act
    doc, err := NewDocument("", content, templateID, createdBy)

    // Assert
    assert.Error(t, err)
    assert.Nil(t, doc)
    assert.Equal(t, ErrEmptyTitle, err)
}

func TestDocument_Create_NilContent_ReturnsError(t *testing.T) {
    // Act
    doc, err := NewDocument("Title", nil, templateID, createdBy)

    // Assert
    assert.Error(t, err)
    assert.Equal(t, ErrNilContent, err)
}
```

**Refactored code**:
```go
var (
    ErrEmptyTitle = errors.New("document title cannot be empty")
    ErrNilContent = errors.New("document content cannot be nil")
)

func NewDocument(title string, content []byte, templateID TemplateID, createdBy UserID) (*Document, error) {
    // Validate invariants
    if strings.TrimSpace(title) == "" {
        return nil, ErrEmptyTitle
    }
    if content == nil {
        return nil, ErrNilContent
    }

    now := time.Now()

    doc := &Document{
        ID:        DocumentID(uuid.New().String()),
        Title:     strings.TrimSpace(title),
        Content:   content,
        TemplateID: templateID,
        Status:    DocumentStatusDraft,
        CreatedBy: createdBy,
        CreatedAt: now,
        UpdatedAt: now,
        Versions:  []Version{createInitialVersion(content, createdBy, now)},
    }

    return doc, nil
}

func createInitialVersion(content []byte, createdBy UserID, now time.Time) Version {
    return Version{
        ID:        VersionID(uuid.New().String()),
        Content:   content,
        CreatedBy: createdBy,
        CreatedAt: now,
    }
}
```

**Результат**: ✅ All tests pass, code is cleaner

---

## TDD для Domain-Driven Design

### Testing Aggregates

#### Правила тестирования Aggregates:

1. **Test invariants**: Каждое правило домена = отдельный тест
2. **Test state transitions**: Каждый переход статуса
3. **Test business rules**: Все if/else в domain logic
4. **Don't test infrastructure**: Repositories, DB - это integration tests

#### Пример: Workflow Aggregate

```go
func TestWorkflowInstance_ApproveStep_ValidApprover_Success(t *testing.T) {
    // Arrange
    workflow := createTestWorkflow(t)
    approver := &entities.UserContext{
        UserID: "user_methodist_001",
        Role:   domain.RoleMethodist,  // RequiredRole for step 0
    }
    comments := "Учебный план соответствует стандартам"

    // Act
    err := workflow.ApproveCurrentStep(approver, comments)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 1, workflow.CurrentStep) // moved to next step
    assert.Equal(t, StepStatusApproved, workflow.Steps[0].Status)
    assert.Equal(t, comments, workflow.Steps[0].Comments)
    assert.NotNil(t, workflow.Steps[0].CompletedAt)
}

func TestWorkflowInstance_ApproveStep_WrongRole_ReturnsError(t *testing.T) {
    // Arrange
    workflow := createTestWorkflow(t)
    approver := &entities.UserContext{
        UserID: "user_student_001",
        Role:   domain.RoleStudent,  // ❌ Student cannot approve
    }

    // Act
    err := workflow.ApproveCurrentStep(approver, "")

    // Assert
    assert.Error(t, err)
    assert.Equal(t, ErrInsufficientPermissions, err)
    assert.Equal(t, 0, workflow.CurrentStep) // не изменился
    assert.Equal(t, StepStatusPending, workflow.Steps[0].Status)
}

func TestWorkflowInstance_ApproveStep_SLAExceeded_Escalates(t *testing.T) {
    // Arrange
    workflow := createTestWorkflow(t)
    workflow.Steps[0].SLA = 1 * time.Hour
    workflow.Steps[0].StartedAt = time.Now().Add(-2 * time.Hour) // 2 hours ago

    approver := &entities.UserContext{
        UserID: "user_methodist_001",
        Role:   domain.RoleMethodist,
    }

    // Act
    err := workflow.ApproveCurrentStep(approver, "")

    // Assert
    assert.Error(t, err)
    assert.Equal(t, ErrSLAExceeded, err)
    assert.Equal(t, WorkflowStatusEscalated, workflow.Status)
}
```

### Testing Domain Services

```go
func TestAuthorizationService_CheckPermission_FullAccess_ReturnsTrue(t *testing.T) {
    // Arrange
    svc := services.NewAuthorizationService()
    userCtx := &entities.UserContext{
        UserID: "user_admin_001",
        Role:   domain.RoleSystemAdmin,
    }

    // Act
    hasPermission := svc.CheckPermission(
        userCtx,
        domain.ResourceDocuments,
        domain.ActionDelete,
        nil,
    )

    // Assert
    assert.True(t, hasPermission)
}

func TestAuthorizationService_CheckPermission_LimitedAccess_ChecksScope(t *testing.T) {
    // Arrange
    svc := services.NewAuthorizationService()
    userCtx := &entities.UserContext{
        UserID:    "user_methodist_001",
        Role:      domain.RoleMethodist,
        FacultyID: ptr.String("faculty_math"),
    }
    resourceScope := &services.Scope{
        FacultyID: ptr.String("faculty_math"),
    }

    // Act
    hasPermission := svc.CheckPermission(
        userCtx,
        domain.ResourceDocuments,
        domain.ActionRead,
        resourceScope,
    )

    // Assert
    assert.True(t, hasPermission)
}

func TestAuthorizationService_CheckPermission_DifferentFaculty_ReturnsFalse(t *testing.T) {
    // Arrange
    svc := services.NewAuthorizationService()
    userCtx := &entities.UserContext{
        UserID:    "user_methodist_001",
        Role:      domain.RoleMethodist,
        FacultyID: ptr.String("faculty_math"),
    }
    resourceScope := &services.Scope{
        FacultyID: ptr.String("faculty_physics"), // ❌ different faculty
    }

    // Act
    hasPermission := svc.CheckPermission(
        userCtx,
        domain.ResourceDocuments,
        domain.ActionRead,
        resourceScope,
    )

    // Assert
    assert.False(t, hasPermission)
}
```

---

## Outside-in TDD

### Подход Outside-in

Начинаем с **acceptance test** (внешнее поведение), затем углубляемся в **unit tests** (внутренняя логика).

```
Outside (высокий уровень)                    Inside (низкий уровень)
┌───────────────────┐                        ┌───────────────────┐
│ Acceptance Test   │ ──────────────────────▶│ Unit Tests        │
│ (End-to-End)      │                        │ (Domain logic)    │
│                   │                        │                   │
│ HTTP Request      │                        │ Aggregate         │
│ → JSON Response   │                        │ Domain Service    │
└───────────────────┘                        │ Value Object      │
                                             └───────────────────┘
```

#### Пример: Document Creation Flow

**❶ Acceptance Test (E2E)**

```go
// tests/e2e/document_creation_test.go

func TestCreateDocument_E2E_ValidInput_Success(t *testing.T) {
    // Arrange
    app := setupTestApp(t)
    defer app.Cleanup()

    token := app.Login("methodist@example.com", "password123")

    payload := map[string]interface{}{
        "title":       "Учебный план по математике",
        "type":        "curriculum",
        "template_id": "tpl_curriculum_001",
        "content": map[string]interface{}{
            "program":  "Математика",
            "credits":  120,
            "duration": "4 года",
        },
    }

    // Act
    resp := app.POST("/api/v1/documents", payload, token)

    // Assert
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var doc map[string]interface{}
    json.Unmarshal(resp.Body, &doc)

    assert.NotEmpty(t, doc["id"])
    assert.Equal(t, "Учебный план по математике", doc["title"])
    assert.Equal(t, "draft", doc["status"])
}
```

**Тест падает** ❌ → Начинаем писать implementation

**❷ Controller Layer Test**

```go
// internal/modules/documents/interfaces/http/document_handler_test.go

func TestDocumentHandler_Create_ValidInput_ReturnsCreated(t *testing.T) {
    // Arrange
    mockUseCase := new(mocks.MockDocumentUseCase)
    handler := NewDocumentHandler(mockUseCase)

    reqBody := `{
        "title": "Учебный план по математике",
        "type": "curriculum",
        "template_id": "tpl_curriculum_001",
        "content": {"program": "Математика"}
    }`

    // Mock expected call
    expectedDoc := &entities.Document{
        ID:     DocumentID("doc_123"),
        Title:  "Учебный план по математике",
        Status: DocumentStatusDraft,
    }
    mockUseCase.On("CreateDocument", mock.Anything, mock.Anything).Return(expectedDoc, nil)

    // Act
    req := httptest.NewRequest("POST", "/api/v1/documents", strings.NewReader(reqBody))
    w := httptest.NewRecorder()
    handler.Create(w, req)

    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    mockUseCase.AssertExpectations(t)
}
```

**❸ Use Case Test**

```go
// internal/modules/documents/application/usecases/create_document_test.go

func TestCreateDocumentUseCase_Execute_ValidInput_Success(t *testing.T) {
    // Arrange
    mockRepo := new(mocks.MockDocumentRepository)
    mockEventBus := new(mocks.MockEventBus)
    useCase := NewCreateDocumentUseCase(mockRepo, mockEventBus)

    input := CreateDocumentInput{
        Title:      "Учебный план по математике",
        Type:       "curriculum",
        TemplateID: "tpl_curriculum_001",
        Content:    map[string]interface{}{"program": "Математика"},
        CreatedBy:  "user_123",
    }

    mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.Document")).Return(nil)
    mockEventBus.On("Publish", "document.created", mock.Anything).Return(nil)

    // Act
    doc, err := useCase.Execute(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, doc)
    assert.NotEmpty(t, doc.ID)
    mockRepo.AssertExpectations(t)
    mockEventBus.AssertExpectations(t)
}
```

**❹ Domain Entity Test (already covered above)**

```go
func TestDocument_Create_ValidInputs_Success(t *testing.T) {
    // ... (см. выше)
}
```

---

## Test Doubles Strategy

### Когда использовать Mocks vs Stubs vs Fakes

| Test Double | Когда использовать | Пример |
|-------------|-------------------|--------|
| **Stub** | Предоставляет canned ответы, не проверяем вызовы | Stub TimeProvider возвращает фиксированное время |
| **Mock** | Проверяем, что метод вызван с правильными параметрами | Mock EventBus.Publish вызван с DocumentCreated |
| **Fake** | Реальная in-memory реализация для тестов | Fake Repository хранит данные в map[string]Entity |
| **Spy** | Записывает вызовы для последующей проверки | Spy Logger собирает все log entries |

#### Пример: Stub

```go
type StubTimeProvider struct {
    FixedTime time.Time
}

func (s *StubTimeProvider) Now() time.Time {
    return s.FixedTime
}

func TestDocument_CreatedAt_UsesTimeProvider(t *testing.T) {
    // Arrange
    fixedTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
    timeProvider := &StubTimeProvider{FixedTime: fixedTime}

    // Act
    doc := NewDocumentWithTimeProvider("Title", content, template, user, timeProvider)

    // Assert
    assert.Equal(t, fixedTime, doc.CreatedAt)
}
```

#### Пример: Mock

```go
type MockEventBus struct {
    mock.Mock
}

func (m *MockEventBus) Publish(topic string, event interface{}) error {
    args := m.Called(topic, event)
    return args.Error(0)
}

func TestDocumentService_CreateDocument_PublishesEvent(t *testing.T) {
    // Arrange
    mockEventBus := new(MockEventBus)
    service := NewDocumentService(repo, mockEventBus)

    // Expect specific call
    mockEventBus.On("Publish", "document.created", mock.MatchedBy(func(evt DocumentCreated) bool {
        return evt.DocumentID == "doc_123" && evt.Type == "curriculum"
    })).Return(nil)

    // Act
    service.CreateDocument(doc)

    // Assert: verify mock expectations
    mockEventBus.AssertExpectations(t)
    mockEventBus.AssertCalled(t, "Publish", "document.created", mock.Anything)
}
```

#### Пример: Fake

```go
type FakeDocumentRepository struct {
    documents map[DocumentID]*entities.Document
    mu        sync.RWMutex
}

func NewFakeDocumentRepository() *FakeDocumentRepository {
    return &FakeDocumentRepository{
        documents: make(map[DocumentID]*entities.Document),
    }
}

func (r *FakeDocumentRepository) Save(ctx context.Context, doc *entities.Document) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.documents[doc.ID] = doc
    return nil
}

func (r *FakeDocumentRepository) GetByID(ctx context.Context, id DocumentID) (*entities.Document, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    doc, exists := r.documents[id]
    if !exists {
        return nil, ErrDocumentNotFound
    }
    return doc, nil
}

// Use in tests
func TestDocumentService_CreateAndRetrieve(t *testing.T) {
    // Arrange
    repo := NewFakeDocumentRepository()
    service := NewDocumentService(repo, eventBus)

    // Act
    doc := createTestDocument(t)
    service.CreateDocument(doc)

    retrieved, err := service.GetDocument(doc.ID)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, doc.ID, retrieved.ID)
}
```

### Рекомендации по Test Doubles

✅ **DO**:
- Используйте Stubs для deterministic behavior (время, random)
- Используйте Mocks для проверки side effects (события, логи)
- Используйте Fakes для простых in-memory хранилищ
- Изолируйте domain logic от infrastructure

❌ **DON'T**:
- Не mocкайте domain entities (тестируйте их напрямую)
- Не переиспользуйте mocks между тестами (создавайте новые)
- Не используйте mocks, если можно использовать реальный объект
- Не mocкайте все подряд (over-mocking делает тесты хрупкими)

---

## Best Practices

### 1. Naming Conventions

```go
// ✅ GOOD: Описательные имена
func TestDocument_UpdateContent_ApprovedStatus_ReturnsError(t *testing.T) {}
func TestWorkflowInstance_ApproveStep_WrongRole_ReturnsInsufficientPermissionsError(t *testing.T) {}

// ❌ BAD: Неясные имена
func TestDocument1(t *testing.T) {}
func TestUpdate(t *testing.T) {}
```

**Паттерн**: `Test<Unit>_<Method>_<Scenario>_<ExpectedBehavior>`

### 2. Arrange-Act-Assert (AAA)

```go
func TestDocument_UpdateTitle_ValidInput_Success(t *testing.T) {
    // Arrange: Setup test data
    doc := createTestDocument(t)
    newTitle := "Updated Title"

    // Act: Execute behavior
    err := doc.UpdateTitle(newTitle)

    // Assert: Verify results
    assert.NoError(t, err)
    assert.Equal(t, newTitle, doc.Title)
}
```

### 3. Table-Driven Tests

Для множества сценариев с разными inputs:

```go
func TestDocument_Validate(t *testing.T) {
    tests := []struct {
        name        string
        title       string
        content     []byte
        templateID  TemplateID
        expectedErr error
    }{
        {
            name:        "valid curriculum",
            title:       "Учебный план",
            content:     []byte(`{"program": "Math"}`),
            templateID:  "tpl_curriculum",
            expectedErr: nil,
        },
        {
            name:        "empty title",
            title:       "",
            content:     []byte(`{}`),
            templateID:  "tpl_curriculum",
            expectedErr: ErrEmptyTitle,
        },
        {
            name:        "nil content",
            title:       "Title",
            content:     nil,
            templateID:  "tpl_curriculum",
            expectedErr: ErrNilContent,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Act
            _, err := NewDocument(tt.title, tt.content, tt.templateID, userID)

            // Assert
            if tt.expectedErr != nil {
                assert.Equal(t, tt.expectedErr, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 4. Test Fixtures и Builders

```go
// Test Builder Pattern
type DocumentBuilder struct {
    doc *entities.Document
}

func NewDocumentBuilder() *DocumentBuilder {
    return &DocumentBuilder{
        doc: &entities.Document{
            ID:     DocumentID("doc_test"),
            Title:  "Test Document",
            Status: DocumentStatusDraft,
        },
    }
}

func (b *DocumentBuilder) WithTitle(title string) *DocumentBuilder {
    b.doc.Title = title
    return b
}

func (b *DocumentBuilder) WithStatus(status DocumentStatus) *DocumentBuilder {
    b.doc.Status = status
    return b
}

func (b *DocumentBuilder) Build() *entities.Document {
    return b.doc
}

// Usage in tests
func TestWorkflow_StartApproval_DraftDocument_Success(t *testing.T) {
    // Arrange
    doc := NewDocumentBuilder().
        WithTitle("Test").
        WithStatus(DocumentStatusDraft).
        Build()

    // Act & Assert
    workflow, err := StartApprovalWorkflow(doc)
    assert.NoError(t, err)
}
```

### 5. Focus on Behavior, Not Implementation

```go
// ❌ BAD: Testing implementation details
func TestDocument_UpdateTitle_CallsSetTitle(t *testing.T) {
    doc := &Document{}
    doc.SetTitle("New Title")  // testing private method
    assert.Equal(t, "New Title", doc.title)  // accessing private field
}

// ✅ GOOD: Testing behavior
func TestDocument_UpdateTitle_TitleIsUpdated(t *testing.T) {
    doc := NewDocument("Old Title", ...)
    err := doc.UpdateTitle("New Title")

    assert.NoError(t, err)
    assert.Equal(t, "New Title", doc.GetTitle())  // public API
}
```

### 6. Avoid Test Interdependence

```go
// ❌ BAD: Tests depend on execution order
var globalDoc *Document

func TestCreateDocument(t *testing.T) {
    globalDoc = NewDocument(...)
}

func TestUpdateDocument(t *testing.T) {
    globalDoc.UpdateTitle("...")  // depends on previous test
}

// ✅ GOOD: Each test is independent
func TestCreateDocument(t *testing.T) {
    doc := NewDocument(...)
    assert.NotNil(t, doc)
}

func TestUpdateDocument(t *testing.T) {
    doc := createTestDocument(t)  // fresh instance
    err := doc.UpdateTitle("...")
    assert.NoError(t, err)
}
```

---

## Frontend TDD (React + TypeScript)

### Testing React Components

```typescript
// components/DocumentCard.test.tsx

describe('DocumentCard', () => {
  it('renders document title and status', () => {
    // Arrange
    const document = {
      id: 'doc_123',
      title: 'Учебный план по математике',
      status: 'draft' as const,
      createdAt: '2025-01-15T10:00:00Z',
    };

    // Act
    render(<DocumentCard document={document} />);

    // Assert
    expect(screen.getByText('Учебный план по математике')).toBeInTheDocument();
    expect(screen.getByText('Черновик')).toBeInTheDocument();
  });

  it('calls onDelete when delete button clicked', async () => {
    // Arrange
    const onDelete = vi.fn();
    const document = { id: 'doc_123', title: 'Test', status: 'draft' as const };

    // Act
    render(<DocumentCard document={document} onDelete={onDelete} />);
    await userEvent.click(screen.getByRole('button', { name: /delete/i }));

    // Assert
    expect(onDelete).toHaveBeenCalledWith('doc_123');
  });
});
```

---

## Метрики TDD

### Измерение эффективности TDD

| Метрика | Target | Как измерять |
|---------|--------|--------------|
| **Code Coverage** | 80%+ для domain layer | `go test -cover ./...` |
| **Test Execution Time** | < 10 секунд для unit tests | `go test -v ./...` |
| **Mutation Score** | 70%+ | `go-mutesting` |
| **Test/Code Ratio** | 1:1 - 2:1 | Lines of test code / Lines of production code |
| **Defect Escape Rate** | < 5% | Bugs found in production / Total bugs |

---

📅 **Актуальность документа**
**Последнее обновление**: 2025-01-15
**Версия проекта**: 0.2.0
**Статус**: Актуальный
