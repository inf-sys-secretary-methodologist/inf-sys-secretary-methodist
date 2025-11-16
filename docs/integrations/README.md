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

## 🔮 Планируемые интеграции

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

### 1C Integration (Planned)
**Статус:** 📋 В планах

Интеграция с 1С для синхронизации данных.

**Планируемые возможности:**
- Синхронизация пользователей
- Экспорт отчетов
- Обмен документами

**Реализация:** Прямая интеграция через 1C API

---

## 📚 Дополнительные ресурсы

- [Composio Platform Documentation](https://docs.composio.dev/)
- [Gmail API Documentation](https://developers.google.com/gmail/api)
- [OAuth 2.0 Setup Guide](https://developers.google.com/identity/protocols/oauth2)

---

**📅 Актуальность документа**
**Последнее обновление**: 2025-01-15
**Версия проекта**: 0.1.0
**Статус**: Актуальный
