# 🔗 Integrations

Документация по интеграциям системы с внешними сервисами.

## 📧 Email Integrations

### Composio Gmail Integration
**Статус:** ✅ Активно используется

Интеграция с Gmail через платформу Composio для автоматической отправки email уведомлений.

**Основные возможности:**
- Автоматическая отправка Welcome Email при регистрации
- Ручная отправка email через API
- HTML и plain text форматы
- Support для CC, BCC
- Лимиты: 100,000+ писем/день

**Документация:**
- [Полное руководство](./composio-gmail.md)
- [API Reference](../api/api-documentation.md#-notification-service-api)
- [Environment Configuration](../deployment/environment.md)

**Quick Start:**
```bash
# 1. Настройте переменные окружения
export COMPOSIO_API_KEY="your-api-key"
export COMPOSIO_ENTITY_ID="your-entity-id"

# 2. Отправьте тестовый email
curl -X POST http://localhost:8080/api/notifications/send-welcome \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "name": "Test User"}'
```

**Используемые технологии:**
- Composio Platform API
- Gmail API (OAuth 2.0)
- Go HTTP Client

---

## 📱 Telegram Integration

### Telegram Bot
**Статус:** ✅ Активно используется

Интеграция с Telegram для отправки push-уведомлений через бота.

**Основные возможности:**
- Привязка Telegram аккаунта к профилю пользователя
- Автоматическая отправка уведомлений в Telegram
- Поддержка polling (для разработки) и webhook (для production)
- HTML форматирование сообщений

**Документация:**
- [Полное руководство](./telegram-bot.md)
- [Environment Configuration](../deployment/environment.md)

**Quick Start:**
```bash
# 1. Настройте переменные окружения
export TELEGRAM_BOT_TOKEN="your-bot-token"

# 2. Привяжите аккаунт через UI
# Настройки -> Уведомления -> Telegram
```

**Используемые технологии:**
- Telegram Bot API
- Polling / Webhook режимы
- Go HTTP Client

---

## 🏢 Data Integrations

### 1C Integration
**Статус:** ✅ Реализовано

Интеграция с 1С для синхронизации данных сотрудников и студентов.

**Реализованные возможности:**
- Синхронизация сотрудников из 1С
- Синхронизация студентов из 1С
- Автоматическое разрешение конфликтов (стратегии: source_wins, target_wins, manual)
- Просмотр логов синхронизации с фильтрацией
- Управление конфликтами через UI
- Поддержка параметров синхронизации (force, dry_run)

**Документация:**
- [API Reference](../api/api-documentation.md#-integration-api)
- [Environment Configuration](../deployment/environment.md)

**Quick Start:**
```bash
# 1. Настройте переменные окружения
export INTEGRATION_1C_ENABLED=true
export INTEGRATION_1C_BASE_URL="https://your-1c-server/api"
export INTEGRATION_1C_API_KEY="your-api-key"

# 2. Выполните синхронизацию
curl -X POST http://localhost:8080/api/integration/sync/employees \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"force": false}'
```

**Миграции:**
- `016_integration_schema.up.sql` - Создание таблиц:
  - `sync_logs` - Логи синхронизации
  - `sync_conflicts` - Конфликты при синхронизации
  - `external_employees` - Внешние сотрудники из 1С
  - `external_students` - Внешние студенты из 1С

**Используемые технологии:**
- Go HTTP Client с retry logic
- PostgreSQL для хранения состояния синхронизации
- React UI с фильтрацией и пагинацией

**Frontend UI:**
- Страница интеграции: `/integration`
- Просмотр логов синхронизации
- Управление конфликтами с разрешением
- Статистика синхронизации в реальном времени

---

## 🔮 Планируемые интеграции

### AI Assistant (Planned)
**Статус:** 📋 В планах

AI-ассистент на базе LLM для автоматизации рутинных задач через естественный язык.

**Планируемые возможности:**
- Обработка запросов на естественном языке
- Автоматическое создание задач и документов
- Отправка уведомлений по запросу
- Генерация отчётов

**Варианты реализации:**
- Claude API + MCP (рекомендуется)
- OpenAI + Composio SDK
- Ollama (локально, бесплатно)

**Документация:** [Полный план внедрения](./ai-assistant-roadmap.md)

---

### Slack Integration (Planned)
**Статус:** 📋 В планах

Интеграция со Slack для отправки уведомлений в каналы.

**Планируемые возможности:**
- Уведомления о новых документах
- Напоминания о дедлайнах
- Статусы workflows

**Реализация:** Через Composio Slack API

---

### Google Calendar Integration (Planned)
**Статус:** 📋 В планах

Автоматическое создание событий в Google Calendar.

**Планируемые возможности:**
- Синхронизация расписания
- Напоминания о встречах
- Уведомления о дедлайнах

**Реализация:** Через Composio Calendar API

---

## 📚 Дополнительные ресурсы

- [Composio Platform Documentation](https://docs.composio.dev/)
- [Gmail API Documentation](https://developers.google.com/gmail/api)
- [OAuth 2.0 Setup Guide](https://developers.google.com/identity/protocols/oauth2)
- [1C REST API Documentation](https://v8.1c.ru/platforma/http-servisy/)

---

**📅 Актуальность документа**
**Последнее обновление**: 2025-12-17
**Версия проекта**: 0.3.0
**Статус**: Актуальный
