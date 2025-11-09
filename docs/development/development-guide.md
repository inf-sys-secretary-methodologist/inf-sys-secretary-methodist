# 💻 Development Guide

Объединенное руководство по разработке, включающее все необходимые практики и инструменты.

## 📋 Содержание

1. [Начало работы](#-начало-работы)
2. [Локальная разработка](#-локальная-разработка)
3. [Стандарты кодирования](#-стандарты-кодирования)
4. [Git Workflow](#-git-workflow) & [Git Terminal Guide](git-terminal-guide.md)
5. [Pull Request Process](pull-request-guide.md)
6. [Тестирование](#-тестирование)
7. [CI/CD](#-cicd)
8. [Управление спринтами](#-управление-спринтами)

## 🚀 Начало работы

### Требования к окружению

#### Backend (Go)
- **Go**: версия 1.25+
- **PostgreSQL**: 17+
- **Redis**: 7+
- **Docker**: для контейнеризации
- **Make**: для автоматизации

#### Frontend (Next.js)
- **Node.js**: версия 25+ (Current)
- **npm/yarn**: пакетный менеджер
- **TypeScript**: для type safety

### Настройка проекта

```bash
# Клонирование репозитория
git clone https://github.com/org/inf-sys-secretary-methodist.git
cd inf-sys-secretary-methodist

# Установка зависимостей backend
cd backend
go mod download
go mod verify

# Установка зависимостей frontend
cd ../frontend
npm install

# Копирование конфигурации
cp .env.example .env.local
```

### Docker Setup

**ВАЖНО**: Для безопасности используем переопределения конфигурации через `compose.override.yml`.

```bash
# Первичная настройка (только один раз)
cp compose.override.yml.example compose.override.yml

# Отредактируйте compose.override.yml и установите безопасные пароли
# Для локальной разработки можно использовать дефолтные значения из примера

# Запуск всех сервисов
docker compose up -d

# Только БД и Redis
docker compose up -d postgres redis

# Проверка статуса
docker compose ps

# Логи
docker compose logs -f backend
```

**Структура файлов:**
- `compose.yml` - базовая конфигурация (в гите)
- `compose.override.yml.example` - пример локальных настроек (в гите)
- `compose.override.yml` - ваши локальные настройки (НЕ в гите, добавлен в .gitignore)

**Важные переменные окружения:**
- `POSTGRES_PASSWORD` - пароль PostgreSQL (ОБЯЗАТЕЛЬНО установить)
- `JWT_SECRET` - секрет для JWT токенов (ОБЯЗАТЕЛЬНО установить)
- `JWT_REFRESH_SECRET` - секрет для refresh токенов (ОБЯЗАТЕЛЬНО установить)
- `REDIS_PASSWORD` - пароль Redis (опционально для dev, обязательно для prod)

### Database Setup

```bash
# Создание миграций
make migrate-create name=create_users_table

# Применение миграций
make migrate-up

# Откат миграций
make migrate-down

# Сброс БД (development only)
make db-reset
```

## 💻 Локальная разработка

### Backend Development

#### Структура проекта
```
backend/
├── cmd/
│   └── server/           # Entry point
├── internal/
│   ├── modules/          # Модульная архитектура
│   │   ├── auth/
│   │   ├── documents/
│   │   ├── users/
│   │   └── workflow/
│   ├── shared/           # Общие компоненты
│   └── infrastructure/   # Инфраструктурный слой
├── api/                  # API documentation
├── configs/              # Конфигурационные файлы
├── migrations/           # Database migrations
├── scripts/              # Utility scripts
└── tests/                # Интеграционные тесты
```

#### Запуск backend

```bash
# Development режим с hot reload
make dev

# Production build
make build
make run

# С отладкой
make debug

# Тестирование
make test
make test-coverage
```

### Frontend Development

#### Структура проекта
```
frontend/
├── src/
│   ├── app/              # Next.js 15 App Router
│   ├── components/       # React компоненты
│   │   ├── ui/          # UI компоненты
│   │   ├── forms/       # Формы
│   │   └── layouts/     # Layouts
│   ├── lib/             # Utilities и helpers
│   ├── stores/          # Zustand stores
│   ├── types/           # TypeScript типы
│   └── styles/          # CSS и Tailwind
├── public/              # Статические файлы
└── tests/               # Frontend тесты
```

#### Запуск frontend

```bash
# Development server
npm run dev

# Build для production
npm run build
npm run start

# Тестирование
npm run test
npm run test:e2e

# Линтинг
npm run lint
npm run lint:fix
```

## ⚙️ Стандарты кодирования

### Go Code Standards

#### Naming Conventions
```go
// Константы - ALL_CAPS с underscores
const (
    MAX_RETRY_ATTEMPTS = 3
    DEFAULT_TIMEOUT    = 30 * time.Second
)

// Интерфейсы - существительные с суффиксом -er
type DocumentProcessor interface {
    Process(doc Document) error
}

type UserRepository interface {
    Save(user User) error
    GetByID(id string) (*User, error)
}

// Структуры - PascalCase
type CreateDocumentCommand struct {
    Title     string
    Content   string
    AuthorID  UserID
}

// Методы - PascalCase
func (s *DocumentService) CreateDocument(cmd CreateDocumentCommand) error {
    // implementation
}

// Пакеты - lowercase, одно слово
package documents // ✅
package userservice // ❌ лучше users
```

#### Error Handling
```go
// Определение доменных ошибок
var (
    ErrDocumentNotFound = errors.New("document not found")
    ErrUnauthorized     = errors.New("unauthorized access")
    ErrValidationFailed = errors.New("validation failed")
)

// Wrapped errors для контекста
func (s *DocumentService) GetDocument(id DocumentID) (*Document, error) {
    doc, err := s.repo.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get document %s: %w", id, err)
    }
    return doc, nil
}

// Проверка типов ошибок
if errors.Is(err, ErrDocumentNotFound) {
    return http.StatusNotFound
}
```

#### Testing Standards
```go
// Названия тестов - Test<Function>_<Scenario>_<Expected>
func TestDocumentService_CreateDocument_ValidInput_Success(t *testing.T) {
    // Arrange
    service := setupDocumentService(t)
    cmd := CreateDocumentCommand{
        Title:    "Test Document",
        Content:  "Test content",
        AuthorID: "user-123",
    }

    // Act
    result, err := service.CreateDocument(cmd)

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, result.ID)
    assert.Equal(t, cmd.Title, result.Title)
}

// Table-driven tests для множественных сценариев
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"invalid format", "invalid-email", true},
        {"empty email", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### TypeScript/React Standards

#### Component Structure
```typescript
// Props интерфейс
interface DocumentCardProps {
  document: Document;
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
  className?: string;
}

// Functional component с TypeScript
export const DocumentCard: React.FC<DocumentCardProps> = ({
  document,
  onEdit,
  onDelete,
  className = '',
}) => {
  const handleEdit = useCallback(() => {
    onEdit?.(document.id);
  }, [document.id, onEdit]);

  const handleDelete = useCallback(() => {
    onDelete?.(document.id);
  }, [document.id, onDelete]);

  return (
    <Card className={`document-card ${className}`}>
      <CardHeader>
        <CardTitle>{document.title}</CardTitle>
        <CardDescription>
          Created by {document.authorName} on {formatDate(document.createdAt)}
        </CardDescription>
      </CardHeader>

      <CardContent>
        <p className="text-sm text-gray-600">
          {document.content.substring(0, 150)}...
        </p>

        <div className="flex items-center justify-between mt-4">
          <Badge variant={getStatusVariant(document.status)}>
            {document.status}
          </Badge>

          <div className="flex gap-2">
            {onEdit && (
              <Button size="sm" variant="outline" onClick={handleEdit}>
                Edit
              </Button>
            )}
            {onDelete && (
              <Button size="sm" variant="destructive" onClick={handleDelete}>
                Delete
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
```

#### State Management (Zustand)
```typescript
// Store interface
interface DocumentState {
  documents: Document[];
  currentDocument: Document | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchDocuments: () => Promise<void>;
  createDocument: (data: CreateDocumentData) => Promise<void>;
  updateDocument: (id: string, data: UpdateDocumentData) => Promise<void>;
  deleteDocument: (id: string) => Promise<void>;
  setCurrentDocument: (document: Document | null) => void;
}

// Store implementation
export const useDocumentStore = create<DocumentState>((set, get) => ({
  documents: [],
  currentDocument: null,
  isLoading: false,
  error: null,

  fetchDocuments: async () => {
    set({ isLoading: true, error: null });

    try {
      const response = await api.get('/documents');
      set({ documents: response.data, isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Unknown error',
        isLoading: false
      });
    }
  },

  createDocument: async (data) => {
    set({ isLoading: true, error: null });

    try {
      const response = await api.post('/documents', data);
      const newDocument = response.data;

      set(state => ({
        documents: [...state.documents, newDocument],
        isLoading: false
      }));
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to create document',
        isLoading: false
      });
    }
  },

  setCurrentDocument: (document) => set({ currentDocument: document }),
}));
```

## 📝 Git Workflow

> 📘 **Подробное руководство**: См. [Git Terminal Guide](git-terminal-guide.md) для полного списка команд и примеров

### Branch Strategy

#### GitFlow для продакшена
```bash
# Основные ветки
main        # Production-ready код
develop     # Integration branch для разработки

# Feature branches
feature/PROJ-123-document-creation
feature/PROJ-124-user-authentication

# Release branches
release/v1.2.0

# Hotfix branches
hotfix/v1.1.1-critical-security-fix
```

### Commit Message Convention

```bash
# Формат: <type>(<scope>): <description>
#
# <body>
#
# <footer>

# Примеры:
feat(documents): add document creation functionality

docs(api): update authentication endpoints documentation

fix(auth): resolve JWT token validation issue

BREAKING CHANGE: remove deprecated v1 API endpoints

# Types:
feat      # Новая функциональность
fix       # Исправление бага
docs      # Документация
style     # Форматирование кода
refactor  # Рефакторинг без изменения функциональности
test      # Добавление тестов
chore     # Обновление сборки, зависимостей
```

### Pull Request Process

#### Pull Request Process
Для детального процесса создания PR см. [Pull Request Guide](pull-request-guide.md), который включает:
- Шаблоны PR
- Code review процесс
- Naming conventions
- Чеклисты

## 🧪 Тестирование

### Testing Pyramid

#### Unit Tests (70%)
```bash
# Backend unit tests
make test-unit

# Frontend unit tests
npm run test:unit

# Coverage reports
make test-coverage
npm run test:coverage
```

#### Integration Tests (20%)
```bash
# API integration tests
make test-integration

# Database integration tests
make test-db

# Frontend integration tests
npm run test:integration
```

#### E2E Tests (10%)
```bash
# Playwright E2E tests
npm run test:e2e

# Cypress E2E tests (alternative)
npm run cypress:run
```

### Test Data Management

#### Fixtures и Factories
```go
// Test factories
func CreateTestUser(t *testing.T, overrides ...func(*User)) *User {
    user := &User{
        ID:        uuid.New().String(),
        Name:      "Test User",
        Email:     "test@example.com",
        Role:      RoleUser,
        Status:    UserStatusActive,
        CreatedAt: time.Now(),
    }

    for _, override := range overrides {
        override(user)
    }

    return user
}

// Usage
user := CreateTestUser(t, func(u *User) {
    u.Role = RoleAdmin
    u.Email = "admin@example.com"
})
```

## 🔄 CI/CD

### GitHub Actions Pipeline

#### Main workflow
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.25'

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '25'

      - name: Run backend tests
        run: |
          cd backend
          go test ./... -v -race -coverprofile=coverage.out

      - name: Run frontend tests
        run: |
          cd frontend
          npm ci
          npm run test:ci

      - name: Upload coverage
        uses: codecov/codecov-action@v3

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    steps:
      - uses: actions/checkout@v3

      - name: Build and push Docker images
        run: |
          docker build -t inf-sys-backend:${{ github.sha }} ./backend
          docker build -t inf-sys-frontend:${{ github.sha }} ./frontend

          echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin

          docker push inf-sys-backend:${{ github.sha }}
          docker push inf-sys-frontend:${{ github.sha }}

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    steps:
      - name: Deploy to staging
        run: |
          # Deployment logic here
          echo "Deploying to staging environment"
```

## 📊 Управление спринтами

### Sprint Planning

#### Процесс планирования
1. **Product Backlog Review** (30 мин)
2. **Sprint Goal Definition** (30 мин)
3. **Story Selection** (60 мин)
4. **Task Breakdown** (90 мин)
5. **Capacity Planning** (30 мин)

#### Definition of Ready
- [ ] User Story написана и понятна
- [ ] Acceptance criteria определены
- [ ] Story points оценены
- [ ] Dependencies выявлены
- [ ] UI/UX макеты готовы (если нужно)
- [ ] Technical approach согласован

#### Definition of Done
- [ ] Код написан и соответствует стандартам
- [ ] Unit тесты покрывают >80% кода
- [ ] Integration тесты пройдены
- [ ] Code review выполнен
- [ ] Документация API обновлена
- [ ] Acceptance criteria выполнены
- [ ] Deployed в staging и протестировано
- [ ] Product Owner approval получен

### Daily Standups

#### Формат (15 минут)
1. **Что сделал вчера?**
2. **Что планирую сегодня?**
3. **Есть ли блокеры?**

#### GitHub Projects Integration
```yaml
automation_rules:
  - when: "PR opened"
    then: "move to In Review"

  - when: "PR merged"
    then: "move to Done"

  - when: "Issue assigned"
    then: "move to In Progress"
```

### Sprint Review & Retrospective

#### Sprint Review (2 часа)
1. **Demo** выполненных задач (90 мин)
2. **Feedback** от stakeholders (20 мин)
3. **Next sprint** preview (10 мин)

#### Retrospective (1.5 часа)
**Format: Start-Stop-Continue**
- **Start**: Что начать делать?
- **Stop**: Что перестать делать?
- **Continue**: Что продолжить делать?

## 🛠️ Инструменты и настройка

### IDE Configuration

#### VS Code Settings
```json
{
  "go.lintTool": "golangci-lint",
  "go.testFlags": ["-v", "-race"],
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "typescript.preferences.importModuleSpecifier": "relative",
  "eslint.autoFixOnSave": true
}
```

### Makefile Commands
```makefile
# Development commands
.PHONY: dev build test clean

dev:
	air -c .air.toml

build:
	go build -o bin/server cmd/server/main.go

test:
	go test ./... -v -race

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

lint:
	golangci-lint run

# Database commands
migrate-up:
	migrate -path migrations -database $(DATABASE_URL) up

migrate-down:
	migrate -path migrations -database $(DATABASE_URL) down

db-reset:
	migrate -path migrations -database $(DATABASE_URL) drop
	migrate -path migrations -database $(DATABASE_URL) up

# Docker commands
docker-build:
	docker build -t inf-sys-backend .

docker-run:
	docker-compose up -d
```

Это руководство обеспечивает:
- ✅ Единую точку входа для разработчиков
- ✅ Четкие стандарты и процессы
- ✅ Автоматизацию рутинных задач
- ✅ Качественную разработку и тестирование
- ✅ Эффективное управление проектом
---

**📅 Актуальность документа**  
**Последнее обновление**: 2025-01-15  
**Версия проекта**: 0.1.0  
**Статус**: Актуальный

