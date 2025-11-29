# 📧 Composio Gmail Integration

## 📋 Обзор

Интеграция с Gmail через платформу **Composio** для автоматической отправки email уведомлений пользователям системы.

## 🎯 Что такое Composio

**Composio** - это платформа интеграций для AI-агентов и приложений, предоставляющая унифицированный API для работы с различными сервисами (Gmail, Slack, Calendar, Jira и т.д.).

### Архитектура решения:

```
Ваше приложение → Composio API → Gmail API → Получатель
```

### Компоненты:
- **Composio API**: Унифицированный REST API
- **OAuth Manager**: Управление авторизацией через UI
- **Connected Accounts**: Привязка Gmail аккаунтов к Entity ID
- **Action Executor**: Выполнение действий (отправка email)

---

## 🆚 Сравнение подходов к отправке email

### 1. SMTP (классический подход)

**Архитектура:**
```
Приложение → SMTP сервер Gmail → Получатель
```

**Преимущества:**
- ✅ Простая настройка
- ✅ Нет зависимостей от third-party
- ✅ Бесплатный
- ✅ Прямой протокол

**Недостатки:**
- ❌ Требует App Password (небезопасно)
- ❌ Лимиты: 500 писем/день для обычных аккаунтов
- ❌ Только отправка email (нет доступа к labels, threads, drafts)
- ❌ Сложно управлять OAuth

**Когда использовать:**
- Простые приложения с <100 писем/день
- Бюджет ограничен
- Не нужны дополнительные функции Gmail

### 2. Gmail API (прямой подход)

**Архитектура:**
```
Приложение → OAuth → Gmail API → Получатель
```

**Преимущества:**
- ✅ Высокие лимиты (1,000,000,000 quota units/день ≈ 100,000+ писем)
- ✅ Полный доступ к Gmail (labels, threads, drafts, filters)
- ✅ OAuth 2.0 безопасность
- ✅ Бесплатный
- ✅ Нет промежуточных сервисов

**Недостатки:**
- ❌ Сложная настройка OAuth flow
- ❌ Нужно управлять refresh tokens
- ❌ Требует Google Cloud Project
- ❌ Сложная обработка ошибок API

**Когда использовать:**
- Нужна полная функциональность Gmail
- Критична latency (нет дополнительных hops)
- Не хотите зависимости от third-party
- Есть ресурсы на поддержку OAuth

### 3. Composio + Gmail API (наш подход)

**Архитектура:**
```
Приложение → Composio API → Gmail API → Получатель
```

**Преимущества:**
- ✅ OAuth управляется через Composio UI (проще)
- ✅ Те же лимиты что и Gmail API (100,000+ писем/день)
- ✅ Унифицированный API для разных сервисов
- ✅ Легко добавить Outlook, Slack, Discord
- ✅ Мониторинг и логи из коробки
- ✅ Автоматическое управление refresh tokens
- ✅ Webhook поддержка

**Недостатки:**
- ❌ Зависимость от third-party (Single Point of Failure)
- ❌ Дополнительная latency (~100-200ms)
- ❌ Платная подписка при больших объемах
- ❌ Vendor lock-in

**Когда использовать:**
- ✅ Планируется интеграция с несколькими сервисами (Gmail + Slack + Calendar)
- ✅ Нужно >500 писем/день
- ✅ Хотите простую OAuth настройку
- ✅ Нужен мониторинг и аналитика
- ✅ Enterprise-level решение

### 4. SendGrid / AWS SES (Email Service Providers)

**Преимущества:**
- ✅ Unlimited emails (платно)
- ✅ Высокая надежность
- ✅ Email analytics
- ✅ Template management

**Недостатки:**
- ❌ Платная подписка с первого письма
- ❌ Только email (нет интеграций со Slack, Calendar)
- ❌ Нужна настройка DNS (SPF, DKIM)

---

## 🛠️ Настройка Composio Gmail Integration

### Шаг 1: Создание Google Cloud Project

1. Перейдите в [Google Cloud Console](https://console.cloud.google.com/)
2. Создайте новый проект или выберите существующий
3. Включите Gmail API:
   - APIs & Services → Enable APIs and Services
   - Поиск "Gmail API" → Enable

### Шаг 2: Настройка OAuth Credentials

1. **Создание OAuth 2.0 Client:**
   ```
   APIs & Services → Credentials → Create Credentials → OAuth client ID
   ```

2. **Настройка OAuth consent screen:**
   - User Type: External (для тестирования) или Internal (для G Suite)
   - App name: "Secretary Methodist System"
   - User support email: ваш email
   - Developer contact: ваш email
   - Scopes: `https://www.googleapis.com/auth/gmail.send`

3. **Создание OAuth Client ID:**
   - Application type: Web application
   - Name: "Composio Gmail Integration"
   - Authorized redirect URIs:
     ```
     https://backend.composio.dev/api/v1/auth-apps/add
     ```

4. **Сохраните credentials:**
   ```json
   {
     "client_id": "451773640106-403d63dukqff5qgvusjpub5u6nfhtgr9.apps.googleusercontent.com",
     "client_secret": "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"
   }
   ```

### Шаг 3: Настройка Composio

1. **Создание Auth Config:**
   - Перейдите на [Composio Dashboard](https://app.composio.dev/)
   - Apps → Gmail → Custom OAuth
   - Paste Client ID и Client Secret
   - Redirect URI: `https://backend.composio.dev/api/v1/auth-apps/add`

2. **Создание Connected Account:**
   - Authorize через OAuth flow
   - Выберите Gmail аккаунт (например, `daniilvdovin4@gmail.com`)
   - Дайте разрешения на `gmail.send` scope
   - Получите Connected Account ID (например, `ca_18Xl76uC4wFV`)

3. **Получение Entity ID:**
   - Создайте Entity для вашего проекта
   - Скопируйте Entity ID (например, `pg-test-5210c210-e67a-49e4-a0f4-cf993514d6f2`)

### Шаг 4: Конфигурация в проекте

Добавьте в `.env`:

```bash
# Composio Configuration
COMPOSIO_API_KEY=your-composio-api-key
COMPOSIO_ENTITY_ID=pg-test-5210c210-e67a-49e4-a0f4-cf993514d6f2
```

Добавьте в `internal/shared/infrastructure/config/config.go`:

```go
type ComposioConfig struct {
    APIKey   string `mapstructure:"api_key"`
    EntityID string `mapstructure:"entity_id"`
}

type Config struct {
    // ... existing fields
    Composio ComposioConfig `mapstructure:"composio"`
}
```

---

## 📊 Архитектура реализации

### Компоненты системы:

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                       │
├─────────────────────────────────────────────────────────────┤
│ AuthHandler (Register) → Sends Welcome Email                │
│ EmailHandler → Manual Email Sending                         │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   Domain Layer (Services)                   │
├─────────────────────────────────────────────────────────────┤
│ EmailService Interface:                                     │
│  - SendEmail(ctx, req)                                      │
│  - SendWelcomeEmail(ctx, email, name)                       │
│  - SendPasswordResetEmail(ctx, email, token)                │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│           Infrastructure Layer (Implementation)             │
├─────────────────────────────────────────────────────────────┤
│ ComposioEmailService:                                       │
│  - composioClient *composio.Client                          │
│  - entityID string                                          │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                  Composio API Client                        │
├─────────────────────────────────────────────────────────────┤
│ Client:                                                     │
│  - ExecuteAction(ctx, actionID, req)                        │
│  - SendEmail(ctx, entityID, emailReq)                       │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│               Composio Platform (External)                  │
├─────────────────────────────────────────────────────────────┤
│ POST /api/v2/actions/GMAIL_SEND_EMAIL/execute              │
│  - Manages OAuth tokens                                     │
│  - Calls Gmail API                                          │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                    Gmail API (Google)                       │
├─────────────────────────────────────────────────────────────┤
│ POST https://gmail.googleapis.com/gmail/v1/users/me/messages/send │
└─────────────────────────────────────────────────────────────┘
```

### Ключевые файлы:

| Файл | Описание |
|------|----------|
| `internal/shared/infrastructure/composio/client.go` | HTTP клиент для Composio API |
| `internal/modules/notifications/domain/services/email_service.go` | Interface определение |
| `internal/modules/notifications/application/services/composio_email_service.go` | Реализация через Composio |
| `internal/modules/notifications/interfaces/http/handlers/email_handler.go` | HTTP handlers для API |
| `internal/modules/auth/interfaces/http/handlers/auth_handler.go` | Автоматическая отправка при регистрации |
| `cmd/server/main.go` | Инициализация и dependency injection |

---

## 🚀 Использование

### 1. Автоматическая отправка Welcome Email при регистрации

При успешной регистрации пользователя автоматически отправляется приветственное письмо:

```go
// internal/modules/auth/interfaces/http/handlers/auth_handler.go

func (h *AuthHandler) Register(c *gin.Context) {
    // ... registration logic

    // Send welcome email if email service is available
    if h.emailService != nil {
        // Use background context with timeout for async email sending
        emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        go func() {
            defer cancel()
            if err := h.emailService.SendWelcomeEmail(emailCtx, input.Email, input.Name); err != nil {
                log.Printf("[AuthHandler] Failed to send welcome email to %s: %v", input.Email, err)
            } else {
                log.Printf("[AuthHandler] Welcome email sent successfully to %s", input.Email)
            }
        }()
    }

    // ... auto-login logic
}
```

**Почему используется `context.Background()`:**
- Request context (`c.Request.Context()`) отменяется когда HTTP запрос завершается
- Email отправка асинхронная и может занять несколько секунд
- `context.Background()` с timeout гарантирует, что email отправится даже после возврата HTTP ответа

**Пример Welcome Email:**
```html
Subject: Добро пожаловать в Secretary Methodist System!

Привет, Даниил!

Спасибо за регистрацию в нашей системе.

С уважением,
Команда Secretary Methodist System
```

### 2. Ручная отправка email через API

**Endpoint:** `POST /api/notifications/send-email`

**Headers:**
```http
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request Body:**
```json
{
  "to": ["recipient@example.com"],
  "cc": ["cc@example.com"],
  "bcc": ["bcc@example.com"],
  "subject": "Важное уведомление",
  "body": "Текст письма или HTML содержимое",
  "is_html": true
}
```

**Response (Success):**
```json
{
  "success": true,
  "data": {
    "message": "Email sent successfully"
  }
}
```

**Response (Error):**
```json
{
  "success": false,
  "error": {
    "code": "EMAIL_SEND_FAILED",
    "message": "Failed to send email"
  }
}
```

### 3. Отправка Welcome Email вручную

**Endpoint:** `POST /api/notifications/send-welcome`

**Request Body:**
```json
{
  "email": "user@example.com",
  "name": "Иван Петров"
}
```

### 4. Использование в коде (Domain Layer)

```go
// Получение сервиса через DI
emailService := container.GetEmailService()

// Отправка кастомного email
req := &services.SendEmailRequest{
    To:      []string{"user@example.com"},
    Subject: "Напоминание о дедлайне",
    Body:    "Ваше задание должно быть сдано до 20.01.2025",
    IsHTML:  false,
}

err := emailService.SendEmail(ctx, req)
if err != nil {
    log.Printf("Failed to send email: %v", err)
}
```

---

## 📈 Метрики и мониторинг

### Логирование

Все email операции логируются:

```
[AuthHandler] Welcome email sent successfully to zarovdaniil95@gmail.com
[Composio] Sending request to https://backend.composio.dev/api/v2/actions/GMAIL_SEND_EMAIL/execute
[Composio] Response status: 200
[Composio] Response body: {"executionId":"...","successful":true,"data":{"message_id":"19a7dff1525892be"}}
```

### Метрики для мониторинга

Рекомендуется отслеживать:

| Метрика | Описание | Threshold |
|---------|----------|-----------|
| Email Send Success Rate | % успешных отправок | >95% |
| Email Send Latency | Время отправки email | <2s |
| Composio API Error Rate | % ошибок Composio API | <5% |
| Gmail API Quota Usage | Использование quota | <80% |
| Daily Email Volume | Кол-во писем/день | <100,000 |

### Prometheus метрики (TODO)

```go
var (
    emailsSentTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "emails_sent_total",
            Help: "Total number of emails sent",
        },
        []string{"status", "type"},
    )

    emailSendDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "email_send_duration_seconds",
            Help: "Email send duration in seconds",
        },
        []string{"type"},
    )
)
```

---

## 🔒 Безопасность

### OAuth 2.0 Security

- **Scopes:** Только `gmail.send` - минимальные права
- **Token Refresh:** Composio автоматически обновляет access tokens
- **Token Storage:** Токены хранятся на Composio, не в нашей БД

### Input Validation

Все email inputs проходят валидацию и sanitization:

```go
// Sanitize email addresses
input.To[i] = h.sanitizer.SanitizeEmail(input.To[i])

// Validate
if err := h.validator.Validate(input); err != nil {
    return BadRequest(err.Error())
}
```

### Rate Limiting

Protected endpoints имеют rate limiting:

```go
// Auth rate limiter: 60 req/min + burst 10
protectedGroup.Use(authRateLimiter.RateLimitMiddleware())
```

### Secrets Management

API Key и Entity ID хранятся в:
- Development: `.env` файл
- Production: Kubernetes Secrets или HashiCorp Vault

```bash
# НЕ коммитить в Git!
COMPOSIO_API_KEY=your-secret-key
COMPOSIO_ENTITY_ID=your-entity-id
```

---

## 🧪 Тестирование

### Unit Tests

```go
func TestComposioEmailService_SendWelcomeEmail(t *testing.T) {
    // Mock Composio client
    mockClient := &MockComposioClient{}
    service := NewComposioEmailService(mockClient, "test-entity-id", nil) // nil auditLogger для тестов

    // Test
    err := service.SendWelcomeEmail(context.Background(), "test@example.com", "Test User")

    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, 1, mockClient.SendEmailCallCount())
}
```

### Integration Tests

```bash
# Отправка тестового email
curl -X POST http://localhost:8080/api/notifications/send-welcome \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "name": "Test User"
  }'
```

### E2E Tests

```bash
# Регистрация пользователя и проверка получения welcome email
npm run test:e2e -- --spec "user-registration.spec.ts"
```

---

## 🔧 Troubleshooting

### Проблема: Email не отправляется

**Проверка 1: Конфигурация**
```bash
# Проверьте переменные окружения
echo $COMPOSIO_API_KEY
echo $COMPOSIO_ENTITY_ID
```

**Проверка 2: Connected Account**
```bash
# Проверьте статус Connected Account в Composio Dashboard
# https://app.composio.dev/apps/gmail
```

**Проверка 3: Логи**
```bash
# Проверьте логи backend
docker logs backend-dev | grep -i "email\|composio"
```

### Проблема: Composio API возвращает 400

**Причина:** Отсутствует поле `appName` в запросе

**Решение:**
```go
req := &ExecuteActionRequest{
    EntityID: entityID,
    AppName:  "gmail",  // Обязательное поле!
    Input:    input,
}
```

### Проблема: OAuth authorization failed

**Причина:** Неверный redirect URI в Google Cloud Console

**Решение:**
1. Проверьте redirect URI в Composio Auth Config
2. Добавьте точно такой же URI в Google Cloud Console
3. Пересоздайте Connected Account

### Проблема: Rate limit exceeded

**Симптомы:**
```
429 Too Many Requests
X-RateLimit-Remaining: 0
```

**Решение:**
- Проверьте Gmail API quota в Google Cloud Console
- Убедитесь что не превышен лимит 1,000,000,000 quota units/день
- Для большего лимита запросите повышение quota

---

## 📚 Best Practices

### 1. Асинхронная отправка

Всегда отправляйте email асинхронно чтобы не блокировать HTTP запросы:

```go
go func() {
    emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := emailService.SendEmail(emailCtx, req); err != nil {
        log.Printf("Failed to send email: %v", err)
    }
}()
```

### 2. Error Handling

Не возвращайте ошибки email отправки пользователю:

```go
// ❌ Плохо - пользователь видит ошибку email
if err := emailService.SendEmail(ctx, req); err != nil {
    return c.JSON(500, gin.H{"error": "Email send failed"})
}

// ✅ Хорошо - email отправляется асинхронно, ошибка логируется
go func() {
    if err := emailService.SendEmail(ctx, req); err != nil {
        log.Printf("Email send failed: %v", err)
    }
}()
return c.JSON(200, gin.H{"message": "Registration successful"})
```

### 3. Template Management

Используйте шаблоны для email:

```go
type EmailTemplate struct {
    Subject string
    Body    string
}

func GetWelcomeEmailTemplate(userName string) *EmailTemplate {
    return &EmailTemplate{
        Subject: "Добро пожаловать!",
        Body: fmt.Sprintf(`
            <html>
            <body>
                <h1>Привет, %s!</h1>
                <p>Спасибо за регистрацию.</p>
            </body>
            </html>
        `, userName),
    }
}
```

### 4. Retry Logic

Добавьте retry для временных сбоев:

```go
func sendEmailWithRetry(ctx context.Context, service EmailService, req *SendEmailRequest) error {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        err := service.SendEmail(ctx, req)
        if err == nil {
            return nil
        }

        if isTemporaryError(err) && i < maxRetries-1 {
            time.Sleep(time.Second * time.Duration(i+1))
            continue
        }

        return err
    }
    return nil
}
```

### 5. Email Queue (TODO)

Для больших объемов используйте очередь (Redis/Kafka):

```go
// Добавить в очередь вместо прямой отправки
emailQueue.Enqueue(&EmailJob{
    To:      "user@example.com",
    Subject: "Welcome",
    Body:    "...",
})

// Worker обрабатывает очередь
for job := range emailQueue.Dequeue() {
    emailService.SendEmail(ctx, job)
}
```

---

## 🚀 Будущие улучшения

### Планируемые фичи:

1. **Email Templates System**
   - Шаблоны в БД
   - Template engine (HTML + variables)
   - Preview перед отправкой

2. **Email Queue с Redis**
   - Асинхронная обработка
   - Retry mechanism
   - Priority queue

3. **Email Analytics**
   - Tracking открытий (open rate)
   - Tracking кликов (click rate)
   - Delivery status

4. **Multi-provider Support**
   - Fallback на SendGrid если Composio недоступен
   - A/B testing разных провайдеров

5. **Дополнительные интеграции через Composio**
   - Slack notifications
   - Calendar events
   - Jira task creation

---

## 📊 Сравнительная таблица решений

| Критерий | SMTP | Gmail API | Composio | SendGrid |
|----------|------|-----------|----------|----------|
| **Стоимость** | Free | Free | Free/Paid | Paid |
| **Лимиты** | 500/день | 100k+/день | 100k+/день | Unlimited |
| **Сложность настройки** | Низкая | Высокая | Средняя | Средняя |
| **OAuth управление** | Нет | Ручное | Автоматическое | Нет |
| **Latency** | Низкая | Низкая | Средняя | Низкая |
| **Vendor lock-in** | Нет | Нет | Да | Да |
| **Multi-service** | Нет | Нет | Да | Нет |
| **Мониторинг** | Нет | Нет | Да | Да |
| **Production-ready** | Нет | Да | Да | Да |

---

## 🎓 Выводы

### Используйте Composio когда:
- ✅ Планируете интеграции с несколькими сервисами (Gmail + Slack + Calendar)
- ✅ Нужно отправлять >500 писем/день
- ✅ Хотите простую OAuth настройку через UI
- ✅ Нужен enterprise-level мониторинг
- ✅ Бюджет позволяет платить за масштабирование

### НЕ используйте Composio когда:
- ❌ Простое приложение с <100 писем/день → используйте SMTP
- ❌ Критична latency → используйте прямой Gmail API
- ❌ Не хотите third-party зависимости → используйте собственный SMTP
- ❌ Очень ограниченный бюджет → SMTP бесплатный

### Для secretary-methodist проекта:

**Composio подходит потому что:**
1. Планируется добавить Calendar API для автоматического создания встреч
2. Возможно понадобится Slack для уведомлений преподавателей
3. Нужен мониторинг email отправок для compliance
4. Объем может вырасти >500 писем/день при росте пользователей
5. OAuth через UI проще для non-technical администраторов

**Альтернатива для MVP:**
- SMTP для первой версии (проще и дешевле)
- Миграция на Composio при росте до 500+ писем/день

---

**📅 Актуальность документа**
**Последнее обновление**: 2025-11-29
**Версия проекта**: 0.1.0
**Статус**: Актуальный

**📧 Контакты**
По вопросам интеграции: [Issues](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues)
