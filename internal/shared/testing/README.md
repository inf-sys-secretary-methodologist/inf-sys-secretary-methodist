# Testing Infrastructure

Инфраструктура для тестирования Go backend приложения.

## Структура

```
testing/
├── fixtures/          # Тестовые данные и builders
│   └── user_fixtures.go
├── helpers/           # Вспомогательные функции для тестов
│   └── test_helpers.go
├── mocks/            # Mock реализации интерфейсов
│   └── user_repository_mock.go
└── suite/            # Base test suites
    └── integration_suite.go
```

## Типы тестов

### Unit Tests

Unit тесты не требуют внешних зависимостей (БД, сеть) и используют моки.

**Пример:**

```go
func TestAuthUseCase_Register(t *testing.T) {
    mockRepo := &mocks.MockUserRepository{}
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    usecase := usecases.NewAuthUseCase(mockRepo, []byte("secret"), []byte("refresh"), nil, nil)

    err := usecase.Register(context.Background(), dto.RegisterInput{
        Email:    "test@example.com",
        Password: "Test123456",
        Role:     "student",
    })

    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### Integration Tests

Integration тесты используют реальную БД и testify/suite.

**Пример:**

```go
type UserRepositoryTestSuite struct {
    suite.IntegrationSuite
    repo *infrastructure.UserRepositoryPG
}

func (s *UserRepositoryTestSuite) SetupSuite() {
    s.IntegrationSuite.SetupSuite()
    s.repo = infrastructure.NewUserRepositoryPG(s.DB)
}

func (s *UserRepositoryTestSuite) TearDownTest() {
    s.TruncateTables("users")
}

func (s *UserRepositoryTestSuite) TestCreate() {
    ctx := helpers.TestContext(s.T())
    user := fixtures.NewUserBuilder().
        WithEmail("test@example.com").
        Build()

    err := s.repo.Create(ctx, user)
    s.NoError(err)
    s.NotZero(user.ID)
}

func TestUserRepositoryTestSuite(t *testing.T) {
    suite.Run(t, new(UserRepositoryTestSuite))
}
```

## Fixtures

Используйте builders для создания тестовых данных:

```go
// Predefined fixtures
user := fixtures.AdminUser()
user := fixtures.StudentUser()

// Custom fixture
user := fixtures.NewUserBuilder().
    WithEmail("custom@example.com").
    WithRole(entities.RoleTeacher).
    WithPassword("CustomPass123").
    Build()
```

## Helpers

### TestContext

Создает context с таймаутом для тестов:

```go
ctx := helpers.TestContext(t)
```

### Database Helpers

```go
// Setup test database
db := helpers.SetupTestDB(t)

// Cleanup tables
helpers.CleanupTestDB(t, db, "users", "sessions")

// Truncate tables
helpers.TruncateTestDB(t, db, "users", "sessions")

// Execute query
helpers.MustExec(t, db, "INSERT INTO users ...")
```

## Запуск тестов

```bash
# Все тесты
just test

# Только unit тесты (без БД)
just test-unit

# Только integration тесты
just test-integration

# Coverage report
just test-coverage

# HTML coverage report
just test-coverage-html

# Конкретный пакет
just test-package internal/modules/auth/application/usecases
```

## Настройка тестовой БД

```bash
# Создать тестовую БД
just setup-test-db

# Удалить тестовую БД
just drop-test-db
```

## Best Practices

### 1. Используйте Table-Driven Tests

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid email", "test@example.com", false},
        {"invalid email", "notanemail", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2. Cleanup после тестов

```go
func (s *UserRepositoryTestSuite) TearDownTest() {
    s.TruncateTables("users", "sessions")
}
```

### 3. Используйте fixtures вместо hardcoded данных

❌ Плохо:
```go
user := &entities.User{
    Email: "test@example.com",
    Password: "$2a$14$...",
    Role: entities.RoleStudent,
}
```

✅ Хорошо:
```go
user := fixtures.NewUserBuilder().
    WithEmail("test@example.com").
    Build()
```

### 4. Моки для unit тестов, реальные зависимости для integration

- **Unit тесты**: используйте `mocks.MockUserRepository`
- **Integration тесты**: используйте реальные implementations

### 5. Изолируйте тесты

Каждый тест должен быть независимым и не влиять на другие.

## Coverage цели

- **Unit тесты**: минимум 80% coverage
- **Integration тесты**: критические пути (auth, CRUD операции)
- **Total coverage**: минимум 70%
