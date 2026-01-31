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

- **Локально:** http://localhost:3002
- **Публично:** https://status.YOUR_DOMAIN (после настройки Caddy)

## Настройка Caddy

Добавьте в Caddyfile на сервере:

```caddyfile
status.YOUR_DOMAIN {
    reverse_proxy localhost:3002
}
```

Перезагрузите Caddy:
```bash
sudo systemctl reload caddy
```

## Первоначальная настройка

1. Откройте http://localhost:3002
2. Создайте admin аккаунт
3. Добавьте мониторы:

### Рекомендуемые мониторы

| Сервис | Тип | URL/Host | Интервал |
|--------|-----|----------|----------|
| Frontend | HTTP(s) | https://YOUR_DOMAIN | 60s |
| API Health | HTTP(s) | https://YOUR_DOMAIN/api/health | 60s |
| PostgreSQL | TCP Port | postgres:5432 | 60s |
| Redis | TCP Port | redis:6379 | 60s |
| MinIO | TCP Port | minio:9000 | 60s |

### Настройка уведомлений

1. Settings → Notifications
2. Add Notification → Telegram
3. Введите Bot Token и Chat ID
4. Test → Save

## Status Page

1. Status Pages → Add Status Page
2. Выберите мониторы для отображения
3. Настройте публичный URL
4. Поделитесь ссылкой с пользователями

## Интеграция с Docker network

Для мониторинга внутренних сервисов по имени контейнера, uptime-kuma должен быть в той же сети:

```yaml
# compose.yml
uptime-kuma:
  networks:
    - default
```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| TZ | Europe/Moscow | Часовой пояс |

## Документация

- [GitHub](https://github.com/louislam/uptime-kuma)
- [Wiki](https://github.com/louislam/uptime-kuma/wiki)
