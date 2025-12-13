# Telegram Bot Integration

## Обзор

Интеграция с Telegram позволяет пользователям получать уведомления системы через Telegram бота. Поддерживаются два режима работы:
- **Polling** - для локальной разработки (без необходимости публичного URL)
- **Webhook** - для production окружения

## Настройка

### 1. Создание бота

1. Откройте [@BotFather](https://t.me/BotFather) в Telegram
2. Отправьте команду `/newbot`
3. Следуйте инструкциям для создания бота
4. Сохраните полученный токен

### 2. Переменные окружения

```env
# Telegram Bot Token (обязательно)
TELEGRAM_BOT_TOKEN=your_bot_token_here

# Режим работы: polling или webhook (опционально, по умолчанию polling)
TELEGRAM_MODE=polling

# Webhook URL (только для режима webhook)
TELEGRAM_WEBHOOK_URL=https://your-domain.com/api/v1/notifications/telegram/webhook
```

### 3. Docker Compose

В `compose.yml` добавьте переменную в сервис backend:

```yaml
backend:
  environment:
    - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
```

## Архитектура

### Компоненты

```
┌─────────────────────────────────────────────────────────────┐
│                        Backend                               │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────┐    ┌─────────────────────────────┐   │
│  │ TelegramHandler  │    │ TelegramWebhookHandler      │   │
│  │ (REST API)       │    │ (Incoming messages)         │   │
│  └────────┬─────────┘    └──────────────┬──────────────┘   │
│           │                              │                   │
│           ▼                              ▼                   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │          TelegramVerificationService                  │   │
│  │  - Генерация кодов верификации                       │   │
│  │  - Привязка аккаунтов                                │   │
│  │  - Отключение аккаунтов                              │   │
│  └──────────────────────────────────────────────────────┘   │
│           │                              │                   │
│           ▼                              ▼                   │
│  ┌──────────────────┐    ┌─────────────────────────────┐   │
│  │ TelegramRepo     │    │ TelegramService (Client)    │   │
│  │ (PostgreSQL)     │    │ (Sending messages)          │   │
│  └──────────────────┘    └─────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Таблицы БД

#### telegram_connections
Хранит связи между пользователями системы и Telegram аккаунтами.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | BIGSERIAL | Primary key |
| user_id | BIGINT | ID пользователя в системе |
| telegram_user_id | BIGINT | ID пользователя в Telegram |
| telegram_chat_id | BIGINT | ID чата для отправки сообщений |
| telegram_username | TEXT | Username в Telegram |
| telegram_first_name | TEXT | Имя в Telegram |
| is_active | BOOLEAN | Активно ли подключение |
| created_at | TIMESTAMP | Дата создания |
| updated_at | TIMESTAMP | Дата обновления |

#### telegram_verification_codes
Хранит временные коды для верификации.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | BIGSERIAL | Primary key |
| user_id | BIGINT | ID пользователя в системе |
| code | TEXT | Код верификации (6 символов) |
| expires_at | TIMESTAMP | Время истечения кода |
| used | BOOLEAN | Использован ли код |
| created_at | TIMESTAMP | Дата создания |

## API Endpoints

### GET /api/v1/notifications/telegram/status
Получить статус подключения Telegram.

**Response:**
```json
{
  "connected": true,
  "username": "john_doe",
  "first_name": "John",
  "connected_at": "2025-12-13T10:00:00Z"
}
```

### POST /api/v1/notifications/telegram/generate-code
Сгенерировать код для привязки аккаунта.

**Response:**
```json
{
  "code": "ABC123",
  "expires_at": "2025-12-13T10:05:00Z",
  "bot_link": "https://t.me/your_bot?start=ABC123"
}
```

### DELETE /api/v1/notifications/telegram/disconnect
Отключить Telegram аккаунт.

**Response:**
```json
{
  "success": true
}
```

### POST /api/v1/notifications/telegram/webhook
Webhook для входящих сообщений от Telegram (только для режима webhook).

## Процесс привязки аккаунта

1. Пользователь открывает настройки уведомлений в веб-интерфейсе
2. Нажимает "Получить код привязки"
3. Система генерирует уникальный 6-символьный код (действителен 5 минут)
4. Пользователь переходит в бота и отправляет код
5. Бот верифицирует код и создаёт связь между аккаунтами
6. Пользователь получает подтверждение в Telegram и на сайте

## Отправка уведомлений

При создании уведомления в системе, оно автоматически отправляется в Telegram, если:
1. У пользователя привязан Telegram аккаунт
2. Подключение активно
3. В настройках включены Telegram-уведомления

```go
// Пример использования в коде
func (uc *NotificationUseCase) Create(ctx context.Context, input *dto.CreateNotificationInput) (*dto.NotificationOutput, error) {
    notification := input.ToEntity()

    if err := uc.notificationRepo.Create(ctx, notification); err != nil {
        return nil, err
    }

    // Автоматическая отправка в Telegram
    uc.sendToTelegram(ctx, notification)

    return dto.ToOutput(notification), nil
}
```

## Форматирование сообщений

Уведомления в Telegram форматируются с использованием HTML:

```
🔔 <b>Заголовок уведомления</b>

Текст сообщения

<i>Приоритет: Высокий</i>
```

## Безопасность

- Коды верификации истекают через 5 минут
- Каждый код можно использовать только один раз
- Связь между аккаунтами уникальна (один Telegram на одного пользователя)
- Все запросы к API требуют аутентификации

---

**Последнее обновление**: 2025-12-13
