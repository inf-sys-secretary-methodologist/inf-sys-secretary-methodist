# n8n - Workflow Automation Platform

## Обзор

[n8n](https://n8n.io/) - self-hosted платформа автоматизации с 400+ интеграциями, визуальным редактором workflow и поддержкой AI.

## Быстрый старт

### Запуск n8n

```bash
# Добавить пароль в .env
echo "N8N_PASSWORD=your-secure-password" >> .env

# Запустить с профилем automation
docker compose --profile automation up -d n8n

# Проверить статус
docker compose ps n8n

# Открыть UI
open http://localhost:5678
```

### Первый вход

1. Откройте <http://localhost:5678>
2. Войдите: `admin` / `<N8N_PASSWORD из .env>`
3. Создайте первый workflow

---

## Примеры автоматизаций для проекта

### 1. Документ → Уведомление в Telegram

**Триггер:** Webhook от backend при создании документа

```
[Webhook] → [Set] → [Telegram]
```

**Backend интеграция:**
```go
// При создании документа вызвать n8n webhook
func (u *DocumentUseCase) Create(ctx context.Context, doc *Document) error {
    // ... создание документа ...

    // Отправить в n8n
    go u.notifyN8N("document_created", map[string]any{
        "id":    doc.ID,
        "title": doc.Title,
        "author": doc.AuthorID,
    })
    return nil
}
```

### 2. Пропуски → Алерт куратору

**Триггер:** Schedule (cron) - ежедневно в 18:00

```
[Schedule Trigger] → [HTTP Request: API] → [IF: пропусков > 3] → [Telegram]
```

**Workflow:**
1. Получить список студентов с пропусками за день
2. Фильтровать: `attendance_missed > 3`
3. Для каждого: отправить сообщение куратору

### 3. Дедлайн → Напоминание

**Триггер:** Schedule - каждый час

```
[Schedule] → [HTTP: API/events?deadline_in=24h] → [Loop] → [Telegram/Email]
```

### 4. Синхронизация с Google Calendar

**Триггер:** Webhook при создании события

```
[Webhook] → [Google Calendar: Create Event] → [Respond to Webhook]
```

### 5. AI-обработка документов

**Триггер:** Webhook при загрузке документа

```
[Webhook] → [HTTP: Download file] → [OpenAI: Extract info] → [HTTP: Update document metadata]
```

### 6. Автоотчёты по расписанию

**Триггер:** Schedule - каждый понедельник 9:00

```
[Schedule] → [HTTP: API/reports/weekly] → [Convert to PDF] → [Email/Telegram]
```

---

## Конфигурация

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `N8N_PORT` | Порт UI | `5678` |
| `N8N_PASSWORD` | Пароль admin | **обязательно** |
| `N8N_BASIC_AUTH_USER` | Логин | `admin` |
| `N8N_WEBHOOK_URL` | URL для webhooks | `http://localhost:5678` |
| `N8N_DB_TYPE` | Тип БД | `postgresdb` |
| `N8N_DB_NAME` | Имя БД для n8n | `n8n` |

### Production настройки

Для production добавьте в `.env`:

```bash
# Публичный URL для webhooks
N8N_WEBHOOK_URL=https://n8n.your-domain.com
N8N_HOST=n8n.your-domain.com
N8N_PROTOCOL=https

# Безопасный пароль
N8N_PASSWORD=<generated-secure-password>
```

### Caddy reverse proxy

Добавьте в `/etc/caddy/Caddyfile`:

```caddyfile
n8n.your-domain.com {
    reverse_proxy localhost:5678
}
```

---

## Создание webhook в backend

### Универсальный сервис уведомлений

```go
// internal/pkg/n8n/client.go
package n8n

import (
    "bytes"
    "encoding/json"
    "net/http"
    "os"
)

type Client struct {
    webhookURL string
    httpClient *http.Client
}

func NewClient() *Client {
    return &Client{
        webhookURL: os.Getenv("N8N_WEBHOOK_URL"),
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

// TriggerWorkflow отправляет событие в n8n webhook
func (c *Client) TriggerWorkflow(workflowID string, data map[string]any) error {
    url := c.webhookURL + "/webhook/" + workflowID

    body, _ := json.Marshal(data)
    resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("n8n webhook failed: %d", resp.StatusCode)
    }
    return nil
}
```

### Использование

```go
// В usecase
func (u *DocumentUseCase) Create(ctx context.Context, doc *Document) error {
    // Создание документа...

    // Async уведомление в n8n
    go func() {
        if err := u.n8nClient.TriggerWorkflow("document-created", map[string]any{
            "document_id": doc.ID,
            "title":       doc.Title,
            "author_id":   doc.AuthorID,
            "created_at":  doc.CreatedAt,
        }); err != nil {
            u.logger.Warn("n8n webhook failed", "error", err)
        }
    }()

    return nil
}
```

---

## Полезные nodes

### Интеграции с проектом

| Node | Использование |
|------|---------------|
| **Webhook** | Получение событий от backend |
| **HTTP Request** | Вызов API backend |
| **Telegram** | Отправка уведомлений |
| **Gmail** | Email уведомления |
| **Schedule Trigger** | Периодические задачи |
| **IF** | Условная логика |
| **Code** | Кастомный JavaScript |
| **OpenAI** | AI-обработка |
| **Google Calendar** | Синхронизация событий |

### AI возможности

n8n поддерживает AI nodes:

- **OpenAI** - GPT-4, DALL-E
- **Anthropic** - Claude
- **LangChain** - RAG pipelines
- **Vector Store** - Embeddings

---

## Мониторинг

### Health check

```bash
curl http://localhost:5678/healthz
```

### Логи

```bash
docker compose logs -f n8n
```

### Метрики

n8n экспортирует метрики Prometheus на `/metrics` (требует настройки).

---

## Troubleshooting

### База данных не создана

Создайте БД `n8n` в PostgreSQL:

```bash
docker compose exec postgres psql -U postgres -c "CREATE DATABASE n8n;"
```

### Webhook не работает

1. Проверьте `N8N_WEBHOOK_URL` в .env
2. Для production: настройте HTTPS через Caddy
3. Проверьте firewall/network

### Ошибка авторизации

1. Проверьте `N8N_PASSWORD` в .env
2. Перезапустите контейнер: `docker compose restart n8n`

---

## Ссылки

- [Официальная документация](https://docs.n8n.io/)
- [Библиотека workflow](https://n8n.io/workflows/)
- [Community nodes](https://www.npmjs.com/search?q=n8n-nodes)
- [GitHub](https://github.com/n8n-io/n8n)

---

## Go Backend Integration

### n8n Webhook Client

The project includes a Go client for triggering n8n workflows via webhooks:

**Location:** `internal/shared/infrastructure/n8n/client.go`

```go
// Initialize in main.go (already wired)
n8nClient := n8ninfra.NewClient(n8ninfra.Config{
    WebhookURL: cfg.N8N.WebhookURL,  // from N8N_WEBHOOK_URL env
    Enabled:    cfg.N8N.Enabled,       // from N8N_ENABLED env
}, logger)

// Trigger a workflow synchronously
err := n8nClient.TriggerWorkflow(ctx, "document-created", map[string]any{
    "document_id": doc.ID,
    "title":       doc.Title,
})

// Trigger asynchronously (fire-and-forget, logs errors)
n8nClient.TriggerAsync("document-created", map[string]any{
    "document_id": doc.ID,
    "title":       doc.Title,
})
```

### EventBus Integration

The `WebhookEventHandler` subscribes to domain events and forwards them to n8n:

**Location:** `internal/shared/infrastructure/n8n/event_handler.go`

```go
handler := n8ninfra.NewWebhookEventHandler(n8nClient, logger)
eventBus.Subscribe("document.created", handler)
eventBus.Subscribe("document.updated", handler)
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `N8N_ENABLED` | Enable webhook integration | `false` |
| `N8N_WEBHOOK_URL` | Base URL for n8n | `http://localhost:5678` |

---

## Pre-built Workflows

Three workflow JSON files are available in the `workflows/` directory:

| File | Trigger | Description |
|------|---------|-------------|
| `document-notification.json` | Webhook | Document created → Telegram notification |
| `absence-alert.json` | Schedule (hourly) | At-risk students → Curator Telegram alert |
| `deadline-reminder.json` | Schedule (daily 9AM) | Approaching deadlines → Assignee reminder |

### Importing Workflows

1. Open n8n UI → Workflows → Import from File
2. Select a JSON file from `workflows/`
3. Configure credentials (Telegram Bot, API Auth header)
4. Activate the workflow

---

## Caddy Reverse Proxy Setup

To expose n8n externally with HTTPS, add to your Caddyfile on the server:

```caddyfile
n8n.your-domain.com {
    reverse_proxy localhost:5678 {
        # WebSocket support for n8n editor
        header_up X-Forwarded-Proto {scheme}
    }
}
```

Then update `.env`:

```bash
N8N_WEBHOOK_URL=https://n8n.your-domain.com
N8N_HOST=n8n.your-domain.com
N8N_PROTOCOL=https
```

Reload Caddy:

```bash
sudo systemctl reload caddy
```

---

**Последнее обновление**: 2026-04-11
