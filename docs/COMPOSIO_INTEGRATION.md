# Composio Gmail Integration

Интеграция с Composio для отправки email через Gmail API.

## Что было сделано

1. ✅ Создан Composio HTTP client (`internal/shared/infrastructure/composio/client.go`)
2. ✅ Создан Email Service (`internal/modules/notifications/`)
3. ✅ Создан HTTP Handler для отправки писем
4. ✅ Добавлены переменные окружения в `.env`

## Настройка

### 1. Получить Composio API Key

1. Откройте <https://platform.composio.dev/>
1. Войдите в свой workspace: `daniilvdovin4_workspace/daniilvdovin4_workspace_first_project`
3. Перейдите в **Settings** → **API Keys**
4. Скопируйте API key (он начинается с `ak_`)

### 2. Настроить переменные окружения

Откройте файл `.env` и замените placeholder на реальный API key:

```bash
# Composio Configuration
COMPOSIO_API_KEY=ak_YOUR_REAL_API_KEY_HERE
COMPOSIO_MCP_CONFIG_ID=8c55f91a-6c5f-4cbb-8691-b497005abf29
```

### 3. Настроить OAuth для Gmail

Перед отправкой писем нужно подключить Gmail аккаунт через OAuth:

1. Откройте <https://platform.composio.dev/daniilvdovin4_workspace/daniilvdovin4_workspace_first_project/auth-configs>
2. Найдите созданный **Gmail Auth Config** (mcp_gmail-hwnmvk)
3. Кликните **Connect Account**
4. Пройдите процесс OAuth авторизации с вашим Gmail аккаунтом
5. После успешной авторизации получите **Entity ID** (это будет user ID для Composio)

## Использование

### API Endpoints

#### 1. Отправка произвольного email

```http
POST /api/notifications/send-email
Content-Type: application/json
Authorization: Bearer <your-jwt-token>

{
  "to": ["recipient@example.com"],
  "cc": ["cc@example.com"],
  "subject": "Тема письма",
  "body": "Текст письма",
  "is_html": false
}
```

#### 2. Отправка приветственного письма

```http
POST /api/notifications/send-welcome
Content-Type: application/json
Authorization: Bearer <your-jwt-token>

{
  "email": "newuser@example.com",
  "name": "Иван Петров"
}
```

### Пример использования в коде

```go
import (
	"context"
	emailServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/services"
)

// Создание email service
apiKey := os.Getenv("COMPOSIO_API_KEY")
entityID := "user-id-from-composio-oauth" // Получается после OAuth
emailService := emailServices.NewComposioEmailService(apiKey, entityID, auditLogger) // auditLogger для аудит-логов (можно nil)

// Отправка welcome email
err := emailService.SendWelcomeEmail(
	context.Background(),
	"user@example.com",
	"Иван Петров",
)
```

## Структура проекта

```
internal/
├── shared/
│   └── infrastructure/
│       └── composio/
│           └── client.go              # HTTP клиент для Composio API
└── modules/
    └── notifications/
        ├── domain/
        │   └── services/
        │       └── email_service.go   # Интерфейс EmailService
        ├── application/
        │   └── services/
        │       └── composio_email_service.go  # Реализация через Composio
        └── interfaces/
            └── http/
                └── handlers/
                    └── email_handler.go  # HTTP хендлеры
```

## Доступные Email действия

- ✅ **SendEmail** - Отправка произвольного email
- ✅ **SendWelcomeEmail** - Приветственное письмо для новых пользователей
- ✅ **SendPasswordResetEmail** - Письмо для сброса пароля
- ✅ **SendNotification** - Общие уведомления

## Следующие шаги

1. Добавить маршруты в `cmd/server/main.go`:
```go
// Email routes
emailService := emailServices.NewComposioEmailService(
	os.Getenv("COMPOSIO_API_KEY"),
	os.Getenv("COMPOSIO_ENTITY_ID"),
	auditLogger, // для аудит-логирования отправки email
)
emailHandler := emailHandlers.NewEmailHandler(emailService)

protectedGroup.POST("/notifications/send-email", emailHandler.SendEmail)
protectedGroup.POST("/notifications/send-welcome", emailHandler.SendWelcomeEmail)
```

1. Пересобрать Docker образ:
```bash
docker build -t inf-sys-secretary-methodist-backend:latest .
docker-compose down
docker-compose up -d
```

1. Протестировать отправку email через API

## Troubleshooting

### Ошибка: "user not authenticated"
- Убедитесь что прошли OAuth авторизацию в Composio
- Проверьте что Entity ID правильный

### Ошибка: "API request failed with status 401"
- Проверьте что API key правильный в `.env`
- Убедитесь что API key не истёк

### Ошибка: "failed to send email"
- Проверьте логи Composio: <https://platform.composio.dev/.../logs>
- Убедитесь что Gmail Auth Config активен

## Полезные ссылки

- [Composio Dashboard](https://platform.composio.dev/daniilvdovin4_workspace/daniilvdovin4_workspace_first_project)
- [Composio Docs](https://docs.composio.dev/)
- [Gmail API Docs](https://docs.composio.dev/toolkits/gmail)

---

**📅 Актуальность документа**
**Последнее обновление**: 2025-12-09
**Версия проекта**: 0.2.0
**Статус**: Актуальный
