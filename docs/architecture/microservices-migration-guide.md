# 🚀 Microservices Migration Guide

## 📋 Обзор

Пошаговое руководство по миграции модульного монолита в микросервисную архитектуру с минимальными рисками и максимальной безопасностью перехода.

## 🎯 Стратегия миграции

### Принципы миграции

#### 1. **Strangler Fig Pattern**
- Постепенное замещение функциональности
- Параллельная работа старой и новой системы
- Постепенное перенаправление трафика

#### 2. **Database Decomposition**
- Поэтапное разделение данных
- Избежание distributed transactions
- Saga pattern для консистентности

#### 3. **Event-Driven Migration**
- Event sourcing для синхронизации
- Асинхронная коммуникация между сервисами
- Компенсирующие транзакции

## 📈 Этапы миграции

### Phase 1: Подготовка модульного монолита

#### 1.1 Укрепление границ модулей
```go
// До: модули могут напрямую обращаться к данным других модулей
type DocumentService struct {
    documentRepo DocumentRepository
    userRepo     UserRepository     // Прямая зависимость
    taskRepo     TaskRepository     // Прямая зависимость
}

// После: использование только интерфейсов и событий
type DocumentService struct {
    documentRepo   DocumentRepository
    userService    UserServiceInterface    // Интерфейс
    taskService    TaskServiceInterface    // Интерфейс
    eventBus       EventBus               // События
}

// Интерфейсы для взаимодействия
type UserServiceInterface interface {
    GetUser(userID UserID) (*UserInfo, error)
    ValidatePermissions(userID UserID, action string) error
}

type TaskServiceInterface interface {
    CreateTaskFromDocument(documentID DocumentID, assigneeID UserID) error
}
```

#### 1.2 Разделение схем БД
```sql
-- Создаем отдельные схемы для каждого модуля
CREATE SCHEMA auth_module;
CREATE SCHEMA document_module;
CREATE SCHEMA workflow_module;
CREATE SCHEMA schedule_module;

-- Мигрируем таблицы в соответствующие схемы
ALTER TABLE users SET SCHEMA auth_module;
ALTER TABLE documents SET SCHEMA document_module;
ALTER TABLE workflows SET SCHEMA workflow_module;

-- Создаем view для кросс-модульных данных
CREATE VIEW public.user_info AS
SELECT id, name, email, role
FROM auth_module.users;
```

#### 1.3 Event-driven коммуникация
```go
// Замена прямых вызовов на события
// До:
func (s *DocumentService) CreateDocument(cmd CreateDocumentCommand) error {
    document := NewDocument(cmd)
    if err := s.documentRepo.Save(document); err != nil {
        return err
    }

    // Прямой вызов
    return s.taskService.CreateTaskFromDocument(document.ID, document.AuthorID)
}

// После:
func (s *DocumentService) CreateDocument(cmd CreateDocumentCommand) error {
    document := NewDocument(cmd)
    if err := s.documentRepo.Save(document); err != nil {
        return err
    }

    // Публикация события
    event := DocumentCreated{
        DocumentID: document.ID,
        AuthorID:   document.AuthorID,
        Title:      document.Title,
        CreatedAt:  document.CreatedAt,
    }

    return s.eventBus.Publish(event)
}

// Обработчик в task module
type TaskDocumentHandler struct {
    taskService TaskService
}

func (h *TaskDocumentHandler) HandleDocumentCreated(event DocumentCreated) error {
    return h.taskService.CreateTaskFromDocument(event.DocumentID, event.AuthorID)
}
```

### Phase 2: Выделение первого микросервиса

#### 2.1 Выбор модуля для выделения

**Критерии выбора:**
```yaml
selection_criteria:
  low_coupling: "Минимальные зависимости от других модулей"
  stable_api: "Стабильный и хорошо определенный API"
  team_ownership: "Четкое ownership командой"
  business_value: "Высокая бизнес-ценность от независимого развития"

recommended_order:
  1: "notification-service"     # Слабые зависимости, четкий API
  2: "file-service"            # Независимый функционал
  3: "auth-service"            # Стабильный и критичный
  4: "reporting-service"       # Сложные вычисления
  5: "integration-service"     # Внешние интеграции
```

#### 2.2 Создание Notification Service

```go
// Структура микросервиса
notification-service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── notification.go
│   │   ├── template.go
│   │   └── subscription.go
│   ├── application/
│   │   ├── usecases/
│   │   └── handlers/
│   ├── infrastructure/
│   │   ├── persistence/
│   │   ├── messaging/
│   │   └── external/
│   └── interfaces/
│       ├── http/
│       ├── grpc/
│       └── messaging/
├── api/
│   ├── proto/
│   │   └── notification.proto
│   └── openapi/
│       └── notification.yaml
├── deployments/
│   ├── docker/
│   └── k8s/
└── tests/
    ├── unit/
    ├── integration/
    └── e2e/
```

#### 2.3 API Gateway Pattern
```go
// API Gateway для маршрутизации
type APIGateway struct {
    monolithHandler    http.Handler
    notificationClient NotificationServiceClient
    router            *mux.Router
}

func (gw *APIGateway) setupRoutes() {
    // Маршруты для микросервиса
    gw.router.PathPrefix("/api/v1/notifications").
        Handler(gw.notificationServiceProxy())

    // Остальные маршруты к монолиту
    gw.router.PathPrefix("/api/v1/").
        Handler(gw.monolithHandler)
}

func (gw *APIGateway) notificationServiceProxy() http.Handler {
    return &httputil.ReverseProxy{
        Director: func(req *http.Request) {
            req.URL.Scheme = "http"
            req.URL.Host = "notification-service:8080"
            req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1/notifications")
        },
    }
}
```

#### 2.4 Синхронизация данных
```go
// Event-based синхронизация
type NotificationEventHandler struct {
    notificationClient NotificationServiceClient
}

func (h *NotificationEventHandler) HandleUserCreated(event UserCreated) error {
    // Синхронизируем данные пользователя в notification service
    return h.notificationClient.SyncUser(context.Background(), &SyncUserRequest{
        UserID: event.UserID,
        Email:  event.Email,
        Name:   event.Name,
        Role:   event.Role,
    })
}

func (h *NotificationEventHandler) HandleDocumentCreated(event DocumentCreated) error {
    // Отправляем уведомление через микросервис
    return h.notificationClient.SendNotification(context.Background(), &SendNotificationRequest{
        Type:      "document_created",
        UserID:    event.AuthorID,
        Data: map[string]string{
            "document_id": event.DocumentID,
            "title":       event.Title,
        },
    })
}
```

### Phase 3: Database per Service

#### 3.1 Создание отдельной БД для сервиса
```yaml
# docker-compose для dev окружения
version: '3.8'
services:
  # Основная БД монолита
  postgres-main:
    image: postgres:15
    environment:
      POSTGRES_DB: main_db
      POSTGRES_USER: main_user
      POSTGRES_PASSWORD: main_pass
    ports:
      - "5432:5432"

  # БД для notification service
  postgres-notifications:
    image: postgres:15
    environment:
      POSTGRES_DB: notifications_db
      POSTGRES_USER: notifications_user
      POSTGRES_PASSWORD: notifications_pass
    ports:
      - "5433:5432"

  notification-service:
    build: ./notification-service
    environment:
      DATABASE_URL: "postgres://notifications_user:notifications_pass@postgres-notifications:5432/notifications_db"
      KAFKA_BROKERS: "kafka:9092"
    depends_on:
      - postgres-notifications
      - kafka
```

#### 3.2 Миграция данных
```go
// Скрипт миграции данных
type DataMigrator struct {
    sourceDB *sql.DB
    targetDB *sql.DB
}

func (m *DataMigrator) MigrateNotificationData() error {
    // 1. Извлекаем данные из монолита
    users, err := m.extractUsers()
    if err != nil {
        return err
    }

    subscriptions, err := m.extractSubscriptions()
    if err != nil {
        return err
    }

    templates, err := m.extractNotificationTemplates()
    if err != nil {
        return err
    }

    // 2. Переносим в новую БД
    if err := m.insertUsers(users); err != nil {
        return err
    }

    if err := m.insertSubscriptions(subscriptions); err != nil {
        return err
    }

    if err := m.insertTemplates(templates); err != nil {
        return err
    }

    return nil
}

func (m *DataMigrator) extractUsers() ([]User, error) {
    query := `
        SELECT id, name, email, role, notification_preferences
        FROM users
        WHERE notification_preferences IS NOT NULL
    `

    rows, err := m.sourceDB.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.NotificationPreferences)
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    return users, nil
}
```

#### 3.3 Saga Pattern для distributed transactions
```go
// Saga для создания документа с уведомлениями
type CreateDocumentSaga struct {
    documentService    DocumentService
    notificationClient NotificationServiceClient
    auditService      AuditService
}

type CreateDocumentSagaStep struct {
    Name    string
    Execute func() error
    Compensate func() error
}

func (s *CreateDocumentSaga) Execute(cmd CreateDocumentCommand) error {
    var documentID DocumentID
    var notificationID string

    steps := []CreateDocumentSagaStep{
        {
            Name: "create_document",
            Execute: func() error {
                doc, err := s.documentService.CreateDocument(cmd)
                if err != nil {
                    return err
                }
                documentID = doc.ID
                return nil
            },
            Compensate: func() error {
                if documentID != "" {
                    return s.documentService.DeleteDocument(documentID)
                }
                return nil
            },
        },
        {
            Name: "send_notification",
            Execute: func() error {
                resp, err := s.notificationClient.SendNotification(context.Background(), &SendNotificationRequest{
                    Type:   "document_created",
                    UserID: cmd.AuthorID,
                    Data: map[string]string{
                        "document_id": documentID.String(),
                        "title":       cmd.Title,
                    },
                })
                if err != nil {
                    return err
                }
                notificationID = resp.NotificationID
                return nil
            },
            Compensate: func() error {
                if notificationID != "" {
                    return s.notificationClient.CancelNotification(context.Background(), &CancelNotificationRequest{
                        NotificationID: notificationID,
                    })
                }
                return nil
            },
        },
        {
            Name: "audit_log",
            Execute: func() error {
                return s.auditService.LogDocumentCreated(documentID, cmd.AuthorID)
            },
            Compensate: func() error {
                return s.auditService.RemoveAuditLog(documentID, "document_created")
            },
        },
    }

    // Выполняем шаги
    executedSteps := 0
    for i, step := range steps {
        if err := step.Execute(); err != nil {
            // Компенсируем выполненные шаги в обратном порядке
            for j := i - 1; j >= 0; j-- {
                if compensateErr := steps[j].Compensate(); compensateErr != nil {
                    log.Printf("Failed to compensate step %s: %v", steps[j].Name, compensateErr)
                }
            }
            return fmt.Errorf("saga failed at step %s: %w", step.Name, err)
        }
        executedSteps++
    }

    return nil
}
```

### Phase 4: Полная миграция остальных сервисов

#### 4.1 Приоритет миграции сервисов
```yaml
migration_phases:
  phase_2:
    services: ["file-service"]
    duration: "4 weeks"
    complexity: "Medium"
    dependencies: ["notification-service"]

  phase_3:
    services: ["auth-service"]
    duration: "6 weeks"
    complexity: "High"
    critical: true

  phase_4:
    services: ["user-service", "integration-service"]
    duration: "8 weeks"
    complexity: "High"
    dependencies: ["auth-service"]

  phase_5:
    services: ["document-service", "workflow-service"]
    duration: "10 weeks"
    complexity: "Very High"
    core_business: true

  phase_6:
    services: ["schedule-service", "task-service"]
    duration: "6 weeks"
    complexity: "Medium"

  phase_7:
    services: ["reporting-service"]
    duration: "4 weeks"
    complexity: "Medium"
    final: true
```

#### 4.2 Service Mesh Implementation
```yaml
# Istio configuration для микросервисов
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: document-service
spec:
  http:
  - match:
    - headers:
        version:
          exact: v2
    route:
    - destination:
        host: document-service-v2
        port:
          number: 8080
  - route:
    - destination:
        host: monolith
        port:
          number: 8080
      headers:
        request:
          add:
            x-service-route: document-legacy
---
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: document-service
spec:
  host: document-service-v2
  trafficPolicy:
    circuitBreaker:
      consecutive5xxErrors: 3
      interval: 30s
      baseEjectionTime: 30s
```

#### 4.3 Feature Flags для постепенного переключения
```go
// Feature flag service для управления миграцией
type FeatureFlagService struct {
    flags map[string]bool
    userSegments map[string][]string
}

func (f *FeatureFlagService) IsEnabled(flag string, userID string) bool {
    // Проверяем глобальный флаг
    if enabled, exists := f.flags[flag]; exists && enabled {
        return true
    }

    // Проверяем пользовательские сегменты
    if segments, exists := f.userSegments[flag]; exists {
        for _, segment := range segments {
            if f.isUserInSegment(userID, segment) {
                return true
            }
        }
    }

    return false
}

// Использование в API Gateway
func (gw *APIGateway) routeDocumentRequest(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromRequest(r)

    if gw.featureFlags.IsEnabled("document-service-v2", userID) {
        // Направляем в микросервис
        gw.documentServiceProxy.ServeHTTP(w, r)
    } else {
        // Направляем в монолит
        gw.monolithHandler.ServeHTTP(w, r)
    }
}
```

## 🛠️ Tooling и Infrastructure

### 1. **Observability Stack**

#### Monitoring
```yaml
# Prometheus configuration
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'monolith'
    static_configs:
      - targets: ['monolith:8080']

  - job_name: 'notification-service'
    static_configs:
      - targets: ['notification-service:8080']

  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
```

#### Tracing
```go
// Distributed tracing setup
func setupTracing() {
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://jaeger:14268/api/traces")))
    if err != nil {
        log.Fatal(err)
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("document-service"),
            semconv.ServiceVersionKey.String("v1.0.0"),
        )),
    )

    otel.SetTracerProvider(tp)
}

// В HTTP handlers
func (h *DocumentHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
    ctx, span := otel.Tracer("document-service").Start(r.Context(), "create_document")
    defer span.End()

    // Добавляем span ID в headers для downstream сервисов
    spanCtx := span.SpanContext()
    r.Header.Set("X-Trace-ID", spanCtx.TraceID().String())
    r.Header.Set("X-Span-ID", spanCtx.SpanID().String())

    // Бизнес-логика
    result, err := h.documentService.CreateDocument(ctx, request)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    span.SetAttributes(
        attribute.String("document.id", result.ID),
        attribute.String("document.type", result.Type),
    )

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

### 2. **CI/CD для микросервисов**

#### Multi-service pipeline
```yaml
# .github/workflows/microservices.yml
name: Microservices CI/CD

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  detect-changes:
    runs-on: ubuntu-latest
    outputs:
      services: ${{ steps.changes.outputs.services }}
    steps:
      - uses: actions/checkout@v3
      - uses: dorny/paths-filter@v2
        id: changes
        with:
          filters: |
            notification-service:
              - 'services/notification-service/**'
            document-service:
              - 'services/document-service/**'
            auth-service:
              - 'services/auth-service/**'
            shared:
              - 'shared/**'

  build-and-test:
    needs: detect-changes
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: ${{ fromJSON(needs.detect-changes.outputs.services) }}
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: 1.21

      - name: Test ${{ matrix.service }}
        run: |
          cd services/${{ matrix.service }}
          go test ./...

      - name: Build ${{ matrix.service }}
        run: |
          cd services/${{ matrix.service }}
          docker build -t ${{ matrix.service }}:${{ github.sha }} .

      - name: Run integration tests
        run: |
          cd services/${{ matrix.service }}
          docker-compose -f docker-compose.test.yml up --abort-on-container-exit

      - name: Push to registry
        if: github.ref == 'refs/heads/main'
        run: |
          echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
          docker push ${{ matrix.service }}:${{ github.sha }}

  deploy:
    needs: [detect-changes, build-and-test]
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    strategy:
      matrix:
        service: ${{ fromJSON(needs.detect-changes.outputs.services) }}
    steps:
      - name: Deploy ${{ matrix.service }}
        run: |
          kubectl set image deployment/${{ matrix.service }} \
            ${{ matrix.service }}=${{ matrix.service }}:${{ github.sha }}
          kubectl rollout status deployment/${{ matrix.service }}
```

### 3. **Configuration Management**

#### Service configuration
```go
// Centralized configuration
type ServiceConfig struct {
    Name        string        `yaml:"name"`
    Port        int           `yaml:"port"`
    Database    DatabaseConfig `yaml:"database"`
    Kafka       KafkaConfig   `yaml:"kafka"`
    Redis       RedisConfig   `yaml:"redis"`
    Auth        AuthConfig    `yaml:"auth"`
    Observability ObservabilityConfig `yaml:"observability"`
}

type DatabaseConfig struct {
    Host            string        `yaml:"host"`
    Port            int           `yaml:"port"`
    Database        string        `yaml:"database"`
    Username        string        `yaml:"username"`
    Password        string        `yaml:"password"`
    MaxOpenConns    int           `yaml:"max_open_conns"`
    MaxIdleConns    int           `yaml:"max_idle_conns"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// Configuration per environment
func LoadConfig(env string) (*ServiceConfig, error) {
    configFile := fmt.Sprintf("configs/%s.yaml", env)

    data, err := os.ReadFile(configFile)
    if err != nil {
        return nil, err
    }

    var config ServiceConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    // Override with environment variables
    if dbHost := os.Getenv("DATABASE_HOST"); dbHost != "" {
        config.Database.Host = dbHost
    }

    if dbPassword := os.Getenv("DATABASE_PASSWORD"); dbPassword != "" {
        config.Database.Password = dbPassword
    }

    return &config, nil
}
```

## 🔍 Testing Strategy

### 1. **Contract Testing**

#### Pact testing между сервисами
```go
// Consumer test (document-service тестирует notification-service)
func TestNotificationServiceContract(t *testing.T) {
    pact := dsl.Pact{
        Consumer: "document-service",
        Provider: "notification-service",
    }
    defer pact.Teardown()

    pact.AddInteraction().
        Given("user exists").
        UponReceiving("a request to send notification").
        WithRequest(dsl.Request{
            Method: "POST",
            Path:   dsl.String("/notifications"),
            Headers: dsl.MapMatcher{
                "Content-Type": dsl.String("application/json"),
            },
            Body: map[string]interface{}{
                "user_id": dsl.String("user-123"),
                "type":    dsl.String("document_created"),
                "data": map[string]interface{}{
                    "document_id": dsl.String("doc-456"),
                    "title":       dsl.String("Test Document"),
                },
            },
        }).
        WillRespondWith(dsl.Response{
            Status: 200,
            Headers: dsl.MapMatcher{
                "Content-Type": dsl.String("application/json"),
            },
            Body: map[string]interface{}{
                "notification_id": dsl.String("notif-789"),
                "status":         dsl.String("sent"),
            },
        })

    // Выполняем тест
    err := pact.Verify(func() error {
        client := NewNotificationClient("http://localhost:8080")
        resp, err := client.SendNotification(SendNotificationRequest{
            UserID: "user-123",
            Type:   "document_created",
            Data: map[string]string{
                "document_id": "doc-456",
                "title":       "Test Document",
            },
        })

        if err != nil {
            return err
        }

        if resp.Status != "sent" {
            return fmt.Errorf("expected status 'sent', got '%s'", resp.Status)
        }

        return nil
    })

    assert.NoError(t, err)
}
```

### 2. **End-to-End Testing**

#### E2E тесты для критических сценариев
```go
// E2E test для создания и одобрения документа
func TestDocumentApprovalWorkflow(t *testing.T) {
    // Setup test environment
    testEnv := setupTestEnvironment(t)
    defer testEnv.Cleanup()

    // Create test users
    author := testEnv.CreateUser("author@test.com", "Author", "methodist")
    approver := testEnv.CreateUser("approver@test.com", "Approver", "admin")

    // Step 1: Create document
    createReq := CreateDocumentRequest{
        Title:   "Test Service Note",
        Content: "This is a test document",
        Type:    "service_note",
    }

    createResp := testEnv.APIClient.CreateDocument(author.Token, createReq)
    assert.Equal(t, 201, createResp.StatusCode)

    var document Document
    json.Unmarshal(createResp.Body, &document)

    // Step 2: Verify notification was sent
    notifications := testEnv.GetNotifications(author.ID)
    assert.Len(t, notifications, 1)
    assert.Equal(t, "document_created", notifications[0].Type)

    // Step 3: Verify workflow was started
    workflow := testEnv.GetWorkflow(document.ID)
    assert.Equal(t, "pending", workflow.Status)
    assert.Len(t, workflow.Steps, 1)

    // Step 4: Approve document
    approveReq := ApproveDocumentRequest{
        DocumentID: document.ID,
        Comments:   "Approved for testing",
    }

    approveResp := testEnv.APIClient.ApproveDocument(approver.Token, approveReq)
    assert.Equal(t, 200, approveResp.StatusCode)

    // Step 5: Verify document status changed
    updatedDocument := testEnv.GetDocument(document.ID)
    assert.Equal(t, "approved", updatedDocument.Status)

    // Step 6: Verify approval notification sent
    approvalNotifications := testEnv.GetNotifications(author.ID)
    assert.Len(t, approvalNotifications, 2) // create + approve
    assert.Equal(t, "document_approved", approvalNotifications[1].Type)

    // Step 7: Verify workflow completed
    completedWorkflow := testEnv.GetWorkflow(document.ID)
    assert.Equal(t, "completed", completedWorkflow.Status)
}
```

## 📊 Migration Metrics

### Success Metrics
```yaml
technical_metrics:
  service_independence:
    target: ">95%"
    measure: "percentage of API calls not crossing service boundaries"

  response_time:
    target: "<200ms p95"
    measure: "API response time including service-to-service calls"

  availability:
    target: "99.9%"
    measure: "service uptime during migration"

  error_rate:
    target: "<0.1%"
    measure: "percentage of failed requests"

business_metrics:
  feature_velocity:
    target: "maintain current velocity"
    measure: "story points delivered per sprint"

  deployment_frequency:
    target: "increase by 50%"
    measure: "deployments per week per service"

  lead_time:
    target: "reduce by 30%"
    measure: "time from code commit to production"

  mean_time_to_recovery:
    target: "reduce by 60%"
    measure: "time to resolve production issues"
```

### Risk Mitigation

#### Rollback Strategy
```go
// Circuit breaker для возврата к монолиту
type ServiceCircuitBreaker struct {
    name           string
    failureThreshold int
    resetTimeout     time.Duration
    failures        int
    lastFailTime     time.Time
    state           CircuitState
}

type CircuitState int

const (
    Closed CircuitState = iota
    Open
    HalfOpen
)

func (cb *ServiceCircuitBreaker) Call(fn func() error) error {
    if cb.state == Open {
        if time.Since(cb.lastFailTime) > cb.resetTimeout {
            cb.state = HalfOpen
            cb.failures = 0
        } else {
            return ErrCircuitBreakerOpen
        }
    }

    err := fn()

    if err != nil {
        cb.onFailure()
        return err
    }

    cb.onSuccess()
    return nil
}

func (cb *ServiceCircuitBreaker) onFailure() {
    cb.failures++
    cb.lastFailTime = time.Now()

    if cb.failures >= cb.failureThreshold {
        cb.state = Open
    }
}

func (cb *ServiceCircuitBreaker) onSuccess() {
    cb.failures = 0
    cb.state = Closed
}

// Использование в API Gateway
func (gw *APIGateway) callDocumentService(req *http.Request) (*http.Response, error) {
    var resp *http.Response
    var err error

    err = gw.documentServiceCircuitBreaker.Call(func() error {
        resp, err = gw.documentServiceClient.Do(req)
        return err
    })

    if err == ErrCircuitBreakerOpen {
        // Fallback к монолиту
        return gw.monolithClient.Do(req)
    }

    return resp, err
}
```

## 🎯 Migration Checklist

### Pre-migration
- [ ] Модули четко разделены с минимальными зависимостями
- [ ] Event-driven коммуникация реализована
- [ ] Раздельные схемы БД созданы
- [ ] Monitoring и observability настроены
- [ ] Rollback стратегия определена

### During migration
- [ ] API Gateway настроен
- [ ] Feature flags реализованы
- [ ] Circuit breakers установлены
- [ ] Contract tests написаны
- [ ] Canary deployment готов

### Post-migration
- [ ] Service mesh настроен
- [ ] Distributed tracing работает
- [ ] Security policies обновлены
- [ ] Performance baselines установлены
- [ ] Team ownership определен

Эта стратегия миграции обеспечивает:
- ✅ Минимальные риски при переходе
- ✅ Постепенное и безопасное внедрение
- ✅ Возможность отката на любом этапе
- ✅ Сохранение работоспособности системы
- ✅ Готовность к масштабированию
---

**📅 Актуальность документа**  
**Последнее обновление**: 2025-01-15  
**Версия проекта**: 0.2.0  
**Статус**: Актуальный
