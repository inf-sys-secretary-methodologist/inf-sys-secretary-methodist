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

# Имя бота (без @, используется в ссылках)
TELEGRAM_BOT_NAME=your_bot_name

# Режим работы: polling или webhook (опционально, по умолчанию polling)
TELEGRAM_MODE=polling

# Webhook URL (только для режима webhook)
TELEGRAM_WEBHOOK_URL=https://your-domain.com/api/telegram/webhook

# Webhook Secret (опционально, для дополнительной безопасности)
TELEGRAM_WEBHOOK_SECRET=your_secret_here
```

### 3. Docker Compose

В `compose.yml` добавьте переменные в сервис backend:

```yaml
backend:
  environment:
    - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
    - TELEGRAM_BOT_NAME=${TELEGRAM_BOT_NAME}
    - TELEGRAM_MODE=${TELEGRAM_MODE:-polling}
```

## Архитектура

### Компоненты

```
┌─────────────────────────────────────────────────────────────────────┐
│                            Backend                                   │
├─────────────────────────────────────────────────────────────────────┤
│  ┌────────────────────┐       ┌──────────────────────────────┐     │
│  │  TelegramHandler   │       │  TelegramWebhookHandler      │     │
│  │  (REST API)        │       │  (Incoming messages)         │     │
│  │                    │       │  - /start command            │     │
│  │  - generate-code   │       │  - verification codes        │     │
│  │  - status          │       │  - /help, /status commands   │     │
│  │  - disconnect      │       │                              │     │
│  └─────────┬──────────┘       └───────────────┬──────────────┘     │
│            │                                   │                    │
│            ▼                                   ▼                    │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │               TelegramVerificationService                     │  │
│  │  - GenerateVerificationCode() - криптобезопасный 8-hex код   │  │
│  │  - VerifyCode() - привязка аккаунта                          │  │
│  │  - GetConnection() - проверка статуса                        │  │
│  │  - DisconnectTelegram() - отключение                         │  │
│  │  - CleanupExpiredCodes() - очистка истекших кодов            │  │
│  └──────────────────────────────────────────────────────────────┘  │
│            │                                   │                    │
│            ▼                                   ▼                    │
│  ┌────────────────────┐       ┌──────────────────────────────┐     │
│  │  TelegramRepo      │       │  ComposioTelegramService     │     │
│  │  (PostgreSQL)      │       │  (Sending messages)          │     │
│  │                    │       │  - SendMessage()             │     │
│  │  - connections     │       │  - SendNotification()        │     │
│  │  - codes           │       │  - HTML formatting           │     │
│  └────────────────────┘       └──────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────────┘
```

### Таблицы БД

#### user_telegram_connections
Хранит связи между пользователями системы и Telegram аккаунтами.

| Колонка | Тип | Описание |
|---------|-----|----------|
| user_id | BIGINT | PRIMARY KEY, ID пользователя в системе |
| telegram_chat_id | BIGINT | ID чата для отправки сообщений |
| telegram_username | VARCHAR(255) | Username в Telegram |
| telegram_first_name | VARCHAR(255) | Имя в Telegram |
| is_active | BOOLEAN | Активно ли подключение (default: true) |
| connected_at | TIMESTAMP | Дата подключения |
| updated_at | TIMESTAMP | Дата обновления |

**Индексы:**
- `idx_telegram_connections_chat_id` (telegram_chat_id)

#### telegram_verification_codes
Хранит временные коды для верификации.

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | BIGSERIAL | PRIMARY key |
| user_id | BIGINT | ID пользователя в системе |
| code | VARCHAR(8) | UNIQUE, код верификации (8 hex символов) |
| expires_at | TIMESTAMP | Время истечения кода (15 минут) |
| used_at | TIMESTAMP | Время использования (NULL если не использован) |
| created_at | TIMESTAMP | Дата создания |

**Индексы:**
- `idx_telegram_verification_codes_code` WHERE used_at IS NULL
- `idx_telegram_verification_codes_user_id`
- `idx_telegram_verification_codes_expires` WHERE used_at IS NULL

**Автоочистка:**
```sql
-- PL/pgSQL функция для периодической очистки
CREATE OR REPLACE FUNCTION cleanup_expired_telegram_codes()
RETURNS INTEGER AS $$
DECLARE deleted_count INTEGER;
BEGIN
  DELETE FROM telegram_verification_codes
  WHERE expires_at < NOW() - INTERVAL '1 hour'
  RETURNING COUNT(*) INTO deleted_count;
  RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
```

## API Endpoints

### POST `/api/telegram/verification-code`
Сгенерировать код для привязки Telegram аккаунта.

**Headers:** `Authorization: Bearer <JWT_TOKEN>`

**Response (200):**
```json
{
  "success": true,
  "data": {
    "code": "a1b2c3d4",
    "expires_at": "2025-12-13T10:15:00Z",
    "bot_name": "your_bot",
    "bot_link": "https://t.me/your_bot?start=a1b2c3d4"
  }
}
```

**Примечание:** Если у пользователя уже есть действительный (не истёкший, не использованный) код, он будет возвращён вместо генерации нового.

### GET `/api/telegram/status`
Получить статус подключения Telegram.

**Headers:** `Authorization: Bearer <JWT_TOKEN>`

**Response (200) - подключен:**
```json
{
  "success": true,
  "data": {
    "connected": true,
    "telegram_username": "john_doe",
    "telegram_first_name": "John",
    "connected_at": "2025-12-13T10:00:00Z"
  }
}
```

**Response (200) - не подключен:**
```json
{
  "success": true,
  "data": {
    "connected": false
  }
}
```

### POST `/api/telegram/disconnect`
Отключить Telegram аккаунт.

**Headers:** `Authorization: Bearer <JWT_TOKEN>`

**Response (200):**
```json
{
  "success": true,
  "data": {
    "message": "Telegram disconnected successfully"
  }
}
```

**Побочные эффекты:**
- Удаляется запись из `user_telegram_connections`
- `telegram_enabled` в preferences устанавливается в `false`

### POST `/api/telegram/webhook`
Webhook для входящих сообщений от Telegram Bot API.

**Headers:** (опционально) `X-Telegram-Bot-Api-Secret-Token: <WEBHOOK_SECRET>`

**Request Body:** Telegram Update object

**Response:** Всегда 200 OK (требование Telegram API)

## Обработка команд бота

### Команда `/start`

**С кодом:** `/start a1b2c3d4`
- Проверяет валидность кода (не истёк, не использован)
- Проверяет, что chat_id не привязан к другому пользователю
- Создаёт TelegramConnection
- Отмечает код как использованный
- Включает Telegram в preferences пользователя
- Отправляет приветственное сообщение

**Без кода:** `/start`
- Отправляет инструкции по привязке аккаунта

### Команда `/help`
Отправляет справку по доступным командам.

### Команда `/status`
Показывает статус подключения к системе.

### 8-значный hex код
Если пользователь отправляет 8-значный hex код напрямую (без /start), бот обрабатывает его как код верификации.

## Процесс привязки аккаунта

```
┌──────────────────┐                    ┌──────────────────┐
│   Веб-интерфейс  │                    │   Telegram Bot   │
└────────┬─────────┘                    └────────┬─────────┘
         │                                       │
         │  1. Открыть настройки                 │
         │     уведомлений                       │
         │                                       │
         │  2. Нажать "Подключить Telegram"      │
         │                                       │
         │  3. POST /api/telegram/verification-code
         │     ──────────────────────────────►   │
         │     ◄────────────────────────────────│
         │     {code, bot_link, expires_at}     │
         │                                       │
         │  4. Показать код и QR-код/ссылку     │
         │                                       │
         │                                       │  5. Пользователь переходит
         │                                       │     по ссылке или сканирует QR
         │                                       │
         │                                       │  6. /start a1b2c3d4
         │                                       │     ◄────────────────────
         │                                       │
         │                                       │  7. Верификация кода
         │                                       │     Создание connection
         │                                       │     Включение в preferences
         │                                       │
         │                                       │  8. "Аккаунт успешно привязан!"
         │                                       │     ────────────────────►
         │                                       │
         │  9. Обновление UI                     │
         │     (статус: подключен)               │
         │                                       │
```

## Отправка уведомлений

При создании уведомления в системе, оно автоматически отправляется в Telegram, если:
1. У пользователя привязан Telegram аккаунт (`user_telegram_connections`)
2. Подключение активно (`is_active = true`)
3. В настройках включены Telegram-уведомления (`telegram_enabled = true`)

### Код отправки

```go
// NotificationUseCase автоматически отправляет в Telegram
func (uc *NotificationUseCase) Create(ctx context.Context, input *dto.CreateNotificationInput) (*dto.NotificationOutput, error) {
    notification := input.ToEntity()

    if err := uc.notificationRepo.Create(ctx, notification); err != nil {
        return nil, err
    }

    // Асинхронная отправка в Telegram
    go uc.sendToTelegram(context.Background(), notification)

    return dto.ToOutput(notification), nil
}

func (uc *NotificationUseCase) sendToTelegram(ctx context.Context, notification *entities.Notification) {
    // 1. Проверить наличие подключения
    conn, err := uc.telegramRepo.GetByUserID(ctx, notification.UserID)
    if err != nil || conn == nil || !conn.IsActive {
        return
    }

    // 2. Проверить preferences
    prefs, err := uc.preferencesRepo.GetByUserID(ctx, notification.UserID)
    if err != nil || prefs == nil || !prefs.TelegramEnabled {
        return
    }

    // 3. Отправить через Composio API
    uc.telegramService.SendNotification(
        ctx,
        conn.TelegramChatID,
        notification.Title,
        notification.Message,
        string(notification.Priority),
    )
}
```

## Форматирование сообщений

Уведомления форматируются с HTML разметкой и эмодзи по приоритету:

| Приоритет | Эмодзи | Пример |
|-----------|--------|--------|
| urgent | 🚨 | `🚨 <b>Срочное уведомление</b>` |
| high | ⚠️ | `⚠️ <b>Важное уведомление</b>` |
| normal | 📬 | `📬 <b>Новое уведомление</b>` |
| low | ℹ️ | `ℹ️ <b>Информация</b>` |

**Формат сообщения:**
```html
🚨 <b>Заголовок уведомления</b>

Текст сообщения с подробностями

<i>Приоритет: Срочный</i>
```

## Безопасность

### Коды верификации
- **Длина:** 8 символов (hex)
- **Генерация:** `crypto/rand` (криптографически безопасный)
- **Срок действия:** 15 минут
- **Одноразовые:** Код можно использовать только один раз
- **Уникальность:** Один Telegram аккаунт привязывается только к одному пользователю

### Webhook безопасность
- Опциональный `X-Telegram-Bot-Api-Secret-Token` заголовок
- Всегда возвращаем 200 OK (требование Telegram)
- Асинхронная обработка в goroutine

### Аутентификация API
- Все endpoints (кроме webhook) требуют JWT токен
- Пользователь может управлять только своим подключением

## Frontend компоненты

### TelegramLinkCard
Компонент для управления подключением Telegram в настройках уведомлений.

**Возможности:**
- Отображение статуса подключения
- Генерация кода верификации
- Показ QR-кода или прямой ссылки
- Отключение аккаунта

### Настройки уведомлений
В разделе `/settings/notifications`:
- Переключатель "Telegram уведомления"
- Карточка подключения (TelegramLinkCard)

## Troubleshooting

### Код не принимается
- Проверьте, не истёк ли код (15 минут)
- Убедитесь, что код не был использован ранее
- Попробуйте сгенерировать новый код

### Уведомления не приходят
1. Проверьте статус подключения в `/settings/notifications`
2. Убедитесь, что `telegram_enabled = true` в preferences
3. Проверьте логи backend на наличие ошибок

### Webhook не работает
1. Проверьте доступность URL из интернета
2. Убедитесь, что SSL сертификат валидный
3. Проверьте `TELEGRAM_WEBHOOK_URL` в переменных окружения

---

**📅 Актуальность документа**
**Последнее обновление**: 2025-12-13
**Версия проекта**: 0.2.0
**Статус**: Актуальный
