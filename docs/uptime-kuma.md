# Uptime Kuma - Status Page & Monitoring

Self-hosted мониторинг сервисов с публичной status page.

## Запуск

```bash
# Запуск с профилем monitoring
docker compose --profile monitoring up -d uptime-kuma

# Проверка статуса
docker compose logs -f uptime-kuma
```

## Доступ

| Среда | URL |
|-------|-----|
| Локально | http://localhost:3002 |
| Production | https://status.5-129-253-112.sslip.io |

## Настройка Caddy (на сервере)

Подключиться к серверу и добавить в `/etc/caddy/Caddyfile`:

```bash
ssh -p 2222 deploy@5.129.253.112
sudo nano /etc/caddy/Caddyfile
```

Добавить блок:

```caddyfile
status.5-129-253-112.sslip.io {
    reverse_proxy localhost:3002
}
```

Применить изменения:

```bash
sudo caddy validate --config /etc/caddy/Caddyfile
sudo systemctl reload caddy
```

## Первоначальная настройка

### 1. Создание аккаунта

1. Открыть https://status.5-129-253-112.sslip.io (или http://localhost:3002 локально)
2. Создать admin аккаунт:
   - Username: `admin`
   - Password: (сгенерировать надёжный пароль)

### 2. Настройка Telegram уведомлений

1. **Settings** → **Notifications** → **Setup Notification**
2. **Notification Type:** Telegram
3. **Bot Token:** (использовать существующий бот из SERVER_ADMIN.md или создать нового через @BotFather)
4. **Chat ID:** (ID чата для уведомлений)
5. **Test** → **Save**

### 3. Добавление мониторов

Создать следующие мониторы через кнопку **Add New Monitor**:

| Название | Тип | URL/Host:Port | Интервал | Описание |
|----------|-----|---------------|----------|----------|
| Frontend | HTTP(s) | `https://5-129-253-112.sslip.io` | 60s | Главная страница |
| API Health | HTTP(s) | `https://api.5-129-253-112.sslip.io/api/health` | 60s | Health endpoint бэкенда |
| API Live | HTTP(s) | `https://api.5-129-253-112.sslip.io/api/live` | 60s | Liveness probe |
| Grafana | HTTP(s) | `https://grafana.5-129-253-112.sslip.io` | 120s | Мониторинг дашборд |
| PostgreSQL | TCP Port | `postgres:5432` | 60s | База данных |
| Redis | TCP Port | `redis:6379` | 60s | Кэш |
| MinIO | TCP Port | `minio:9000` | 60s | S3 хранилище |

**Настройки для каждого монитора:**
- **Heartbeat Interval:** 60s (или как указано)
- **Retries:** 3
- **Heartbeat Retry Interval:** 20s
- **Notification:** включить Telegram

### 4. Создание публичной Status Page

1. **Status Pages** → **Add Status Page**
2. Настройки:
   - **Name:** `Информационная система секретаря-методиста`
   - **Slug:** `status` (URL будет /status)
   - **Description:** `Статус сервисов системы управления документооборотом`
   - **Theme:** Auto (light/dark)
   - **Show Powered By:** Off
3. **Добавить группы:**
   - **Группа "Веб-сервисы":** Frontend, API Health, Grafana
   - **Группа "Инфраструктура":** PostgreSQL, Redis, MinIO
4. **Save**
5. **Publish** → включить публичный доступ

## Проверка работы

```bash
# Проверить что контейнер запущен
docker compose ps uptime-kuma

# Проверить логи
docker compose logs uptime-kuma --tail=50

# Проверить доступность локально
curl -s http://localhost:3002 | head -20

# Проверить доступность публично (после настройки Caddy)
curl -s https://status.5-129-253-112.sslip.io | head -20
```

## Интеграция с Docker network

Uptime Kuma находится в той же Docker сети (`default`), что и остальные сервисы, поэтому может мониторить их по именам контейнеров:
- `postgres:5432`
- `redis:6379`
- `minio:9000`
- `backend:8080`

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| TZ | Europe/Moscow | Часовой пояс |

## Бэкап конфигурации

Данные Uptime Kuma хранятся в Docker volume `uptime_kuma_data`. Для бэкапа:

```bash
# Экспорт настроек через UI
# Settings → Backup → Export

# Или скопировать volume
docker run --rm -v inf-sys-secretary-methodist_uptime_kuma_data:/data -v $(pwd):/backup alpine tar cvf /backup/uptime-kuma-backup.tar /data
```

## Восстановление

```bash
# Импорт через UI
# Settings → Backup → Import

# Или восстановить volume
docker run --rm -v inf-sys-secretary-methodist_uptime_kuma_data:/data -v $(pwd):/backup alpine tar xvf /backup/uptime-kuma-backup.tar -C /
```

## Документация

- [GitHub](https://github.com/louislam/uptime-kuma)
- [Wiki](https://github.com/louislam/uptime-kuma/wiki)
- [API Documentation](https://github.com/louislam/uptime-kuma/wiki/API-Documentation)

---

**📅 Последнее обновление:** 2026-02-07
**Версия проекта:** 0.3.3
