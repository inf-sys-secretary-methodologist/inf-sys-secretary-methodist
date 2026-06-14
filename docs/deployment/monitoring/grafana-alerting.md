# Grafana Alerting

Настройка алертов с отправкой в Telegram.

## Настроенные алерты

### Infrastructure
| Алерт | Условие | Severity |
|-------|---------|----------|
| High CPU Usage | CPU > 80% (5 мин) | warning |
| High Memory Usage | RAM > 85% (5 мин) | warning |
| High Disk Usage | Disk > 85% (5 мин) | warning |

### Application
| Алерт | Условие | Severity |
|-------|---------|----------|
| High API Error Rate | 5xx > 1% (5 мин) | critical |
| High API Latency | p95 > 2s (5 мин) | warning |

### Backup
| Алерт | Условие | Severity |
|-------|---------|----------|
| Backup Failed | success = 0 | critical |
| Backup Stale | age > 24h | warning |

## Настройка Telegram

### 1. Создание бота

1. Напишите @BotFather в Telegram
2. `/newbot` → введите имя
3. Сохраните Bot Token

### 2. Получение Chat ID

```bash
# Отправьте сообщение боту, затем:
curl "https://api.telegram.org/bot<BOT_TOKEN>/getUpdates" | jq '.result[0].message.chat.id'
```

### 3. Конфигурация

Добавьте в `.env` на сервере:

```bash
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
TELEGRAM_CHAT_ID=-100123456789
```

### 4. Применение

Перезапустите Grafana:
```bash
docker compose restart grafana
```

## Provisioning файлы

```
monitoring/grafana/provisioning/alerting/
├── contact-points.yml     # Telegram contact point
├── notification-policies.yml  # Routing rules
└── alert-rules.yml        # Alert definitions
```

## Ручная настройка (альтернатива)

Если provisioning не работает, настройте через UI:

1. **Alerting → Contact points → Add**
   - Name: telegram-alerts
   - Type: Telegram
   - Bot Token: ваш токен
   - Chat ID: ваш chat_id

2. **Alerting → Notification policies**
   - Default contact point: telegram-alerts

3. **Alerting → Alert rules → Create**
   - Создайте правила по шаблонам выше

## Тестирование

```bash
# Отправить тестовое уведомление
curl -X POST "http://localhost:3000/api/alertmanager/grafana/api/v2/alerts" \
  -H "Content-Type: application/json" \
  -u admin:admin \
  -d '[{"labels":{"alertname":"TestAlert"},"annotations":{"summary":"Test"}}]'
```

## Документация

- [Grafana Alerting](https://grafana.com/docs/grafana/latest/alerting/)
- [File Provisioning](https://grafana.com/docs/grafana/latest/alerting/set-up/provision-alerting-resources/file-provisioning/)
