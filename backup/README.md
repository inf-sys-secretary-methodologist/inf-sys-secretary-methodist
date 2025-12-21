# Backup & Restore System

Система резервного копирования для PostgreSQL и MinIO с поддержкой уведомлений, метрик и шифрования.

## Возможности

- 📦 Автоматическое резервное копирование PostgreSQL и MinIO
- 🔐 Шифрование бэкапов (GPG или age)
- 📊 Prometheus-совместимые метрики
- 🔔 Уведомления (Telegram, Email, Webhook)
- ☁️ Offsite sync на внешние S3-совместимые хранилища
- 🔄 Автоматическая ротация старых бэкапов
- ✅ CI/CD тесты восстановления

## Быстрый старт

### Запуск backup-сервиса (cron режим)

```bash
# Запуск с профилем backup
docker compose --profile backup up -d backup

# Проверка статуса
docker compose logs -f backup
```

### Ручной бэкап

```bash
# Полный бэкап (PostgreSQL + MinIO)
docker compose run --rm backup /scripts/backup-all.sh

# Только PostgreSQL
docker compose run --rm backup /scripts/backup-postgres.sh

# Только MinIO
docker compose run --rm backup /scripts/backup-minio.sh
```

### Восстановление

```bash
# Восстановление PostgreSQL (последний бэкап)
docker compose run --rm -e RESTORE_CONFIRM=true backup /scripts/restore-postgres.sh

# Восстановление PostgreSQL (конкретный файл)
docker compose run --rm -e RESTORE_CONFIRM=true backup /scripts/restore-postgres.sh /backups/postgres/postgres_20250120_020000.sql.gz

# Восстановление MinIO (последний бэкап)
docker compose run --rm -e RESTORE_CONFIRM=true backup /scripts/restore-minio.sh

# Восстановление MinIO (конкретный файл)
docker compose run --rm -e RESTORE_CONFIRM=true backup /scripts/restore-minio.sh /backups/minio/minio_20250120_020000.tar.gz
```

## Конфигурация

### Основные переменные

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `BACKUP_MODE` | `cron` | Режим работы: `cron`, `backup`, `restore-postgres`, `restore-minio`, `shell` |
| `BACKUP_SCHEDULE` | `0 2 * * *` | Cron расписание (по умолчанию: 2:00 каждый день) |
| `BACKUP_ON_START` | `false` | Выполнить бэкап при старте контейнера |
| `POSTGRES_BACKUP_RETENTION` | `7` | Хранить бэкапы PostgreSQL N дней |
| `MINIO_BACKUP_RETENTION` | `7` | Хранить бэкапы MinIO N дней |
| `SERVER_ID` | `production` | Идентификатор сервера (для multi-server) |

### Шифрование

Бэкапы можно шифровать для безопасного хранения. Поддерживаются два инструмента на выбор:

| Инструмент | Описание | Когда использовать |
|------------|----------|-------------------|
| **age** | Современный, простой | Рекомендуется для большинства случаев |
| **GPG** | Классический, с сертификатами | Если уже используется в компании |

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `BACKUP_ENCRYPTION_ENABLED` | `false` | Включить шифрование бэкапов |
| `BACKUP_ENCRYPTION_TYPE` | `age` | Тип шифрования: `age` или `gpg` |
| `BACKUP_AGE_PUBLIC_KEY` | - | Публичный ключ age для шифрования |
| `BACKUP_GPG_RECIPIENT` | - | Email/ID получателя GPG |

#### Вариант 1: age (рекомендуется)

[age](https://github.com/FiloSottile/age) — современный инструмент шифрования. Простой, без сложной настройки.

**Шаг 1: Генерация ключей**

```bash
# На своём компьютере (не на сервере!)
age-keygen -o backup-key.txt

# Вывод:
# Public key: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
```

Файл `backup-key.txt` содержит:
```
# created: 2025-01-20T12:00:00+03:00
# public key: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
AGE-SECRET-KEY-1QQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQ
```

**Шаг 2: Настройка сервера**

```bash
# В .env добавить ТОЛЬКО публичный ключ
BACKUP_ENCRYPTION_ENABLED=true
BACKUP_ENCRYPTION_TYPE=age
BACKUP_AGE_PUBLIC_KEY=age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
```

**Шаг 3: Хранение приватного ключа**

Приватный ключ (`backup-key.txt`) храните в безопасном месте:
- 1Password / Bitwarden / другой менеджер паролей
- Физический сейф (распечатка)
- Зашифрованный USB-накопитель

⚠️ **Без приватного ключа расшифровать бэкапы невозможно!**

**Расшифровка бэкапа:**

```bash
# Скачать зашифрованный бэкап
scp server:/backups/postgres/postgres_20250120_020000.sql.gz.age ./

# Расшифровать
age -d -i backup-key.txt -o backup.sql.gz backup.sql.gz.age

# Распаковать
gunzip backup.sql.gz
```

#### Вариант 2: GPG

[GPG](https://gnupg.org/) — классический стандарт шифрования с поддержкой сертификатов.

**Шаг 1: Генерация ключа**

```bash
# Генерация ключевой пары
gpg --full-generate-key

# Выбрать:
# - (1) RSA and RSA
# - 4096 bits
# - 2y (срок действия 2 года)
# - Имя: Backup Admin
# - Email: backup@company.com
# - Пароль: надёжный пароль
```

**Шаг 2: Экспорт публичного ключа (для сервера)**

```bash
# Экспорт публичного ключа
gpg --armor --export backup@company.com > backup-public.asc

# Скопировать на сервер
scp backup-public.asc server:/tmp/
```

**Шаг 3: Импорт ключа на сервере**

```bash
# На сервере (внутри контейнера backup)
docker compose run --rm backup sh

# Импорт публичного ключа
gpg --import /tmp/backup-public.asc

# Доверие ключу (выбрать 5 = ultimate)
gpg --edit-key backup@company.com trust quit
```

**Шаг 4: Настройка**

```bash
# В .env
BACKUP_ENCRYPTION_ENABLED=true
BACKUP_ENCRYPTION_TYPE=gpg
BACKUP_GPG_RECIPIENT=backup@company.com
```

**Расшифровка бэкапа:**

```bash
# На компьютере где есть приватный ключ
gpg --decrypt -o backup.sql.gz backup.sql.gz.gpg

# Введите пароль от ключа
gunzip backup.sql.gz
```

#### Какой выбрать?

| Критерий | age | GPG |
|----------|-----|-----|
| Простота настройки | ✅ 2 минуты | ❌ 10-15 минут |
| Нужен пароль при расшифровке | ❌ Нет | ✅ Да |
| Срок действия ключа | ♾️ Бессрочный | ⏰ Настраивается |
| Корпоративный PKI | ❌ Нет | ✅ Да |
| Размер в Docker образе | 3 MB | 10 MB |

**Рекомендация:** Используйте **age** если нет особых требований к PKI/сертификатам

### Уведомления

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `NOTIFY_ON_SUCCESS` | `false` | Уведомлять при успехе |
| `NOTIFY_ON_FAILURE` | `true` | Уведомлять при ошибке |
| `NOTIFY_LEVELS` | `error,warning,success` | Уровни для уведомлений |

#### Telegram

| Переменная | Описание |
|------------|----------|
| `NOTIFY_TELEGRAM_ENABLED` | Включить Telegram уведомления |
| `NOTIFY_TELEGRAM_BOT_TOKEN` | Token бота (@BotFather) |
| `NOTIFY_TELEGRAM_CHAT_ID` | ID чата/группы для уведомлений |

```bash
# Получение chat_id
# Отправьте сообщение боту, затем:
curl https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates | jq '.result[0].message.chat.id'
```

#### Email (SMTP)

| Переменная | Описание |
|------------|----------|
| `NOTIFY_EMAIL_ENABLED` | Включить email уведомления |
| `NOTIFY_EMAIL_SMTP_HOST` | SMTP сервер |
| `NOTIFY_EMAIL_SMTP_PORT` | Порт SMTP (по умолчанию 587) |
| `NOTIFY_EMAIL_FROM` | Email отправителя |
| `NOTIFY_EMAIL_TO` | Email получателя |
| `NOTIFY_EMAIL_USER` | Логин SMTP |
| `NOTIFY_EMAIL_PASSWORD` | Пароль SMTP |

#### Webhook

| Переменная | Описание |
|------------|----------|
| `NOTIFY_WEBHOOK_ENABLED` | Включить webhook уведомления |
| `NOTIFY_WEBHOOK_URL` | URL для POST запроса |
| `NOTIFY_WEBHOOK_SECRET` | Секрет для HMAC подписи (опционально) |

Формат webhook payload:
```json
{
  "level": "error",
  "title": "PostgreSQL Backup Failed",
  "message": "Connection refused",
  "timestamp": "2025-01-20 02:00:00",
  "server_id": "production",
  "hostname": "backup-service"
}
```

### Prometheus Метрики

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `METRICS_ENABLED` | `true` | Включить сбор метрик |
| `METRICS_DIR` | `/var/lib/node_exporter/textfile_collector` | Директория для метрик |

#### Доступные метрики

```prometheus
# Время последнего бэкапа
backup_last_run_timestamp_seconds{server_id="production",type="postgres"} 1705708800

# Время последнего успешного бэкапа
backup_last_success_timestamp_seconds{server_id="production",type="postgres"} 1705708800

# Статус последнего бэкапа (1 = успех, 0 = ошибка)
backup_last_run_success{server_id="production",type="postgres"} 1

# Длительность бэкапа в секундах
backup_duration_seconds{server_id="production",type="postgres"} 120

# Размер бэкапа в байтах
backup_size_bytes{server_id="production",type="postgres"} 1048576

# Время с последнего успешного бэкапа
backup_age_seconds{server_id="production",type="postgres"} 3600

# Счётчики
backup_total_count{server_id="production",type="postgres"} 100
backup_success_count{server_id="production",type="postgres"} 99
backup_failure_count{server_id="production",type="postgres"} 1

# Remote sync метрики
backup_remote_sync_last_run_success{server_id="production"} 1
backup_remote_sync_duration_seconds{server_id="production"} 30
```

#### Интеграция с node_exporter

```yaml
# docker-compose.yml
node_exporter:
  image: prom/node-exporter:latest
  volumes:
    - backup_metrics:/var/lib/node_exporter/textfile_collector:ro
  command:
    - '--collector.textfile.directory=/var/lib/node_exporter/textfile_collector'
```

#### Alertmanager правила

```yaml
# prometheus/alerts.yml
groups:
  - name: backup
    rules:
      - alert: BackupFailed
        expr: backup_last_run_success == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Backup failed on {{ $labels.server_id }}"

      - alert: BackupStale
        expr: backup_age_seconds > 86400
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "No successful backup for 24h on {{ $labels.server_id }}"

      - alert: BackupTooLarge
        expr: backup_size_bytes > 10737418240  # 10GB
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Backup size exceeds 10GB on {{ $labels.server_id }}"
```

### Remote Sync (Offsite Backup)

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `REMOTE_SYNC_ENABLED` | `false` | Включить синхронизацию на внешний S3 |
| `REMOTE_S3_ENDPOINT` | - | URL S3-совместимого хранилища |
| `REMOTE_S3_ACCESS_KEY` | - | Access Key для удалённого хранилища |
| `REMOTE_S3_SECRET_KEY` | - | Secret Key для удалённого хранилища |
| `REMOTE_S3_BUCKET` | - | Имя bucket для бэкапов |
| `REMOTE_S3_REGION` | `us-east-1` | Регион S3 |
| `REMOTE_S3_PATH` | `backups` | Путь внутри bucket |

#### Примеры провайдеров

<details>
<summary>AWS S3</summary>

```bash
REMOTE_SYNC_ENABLED=true
REMOTE_S3_ENDPOINT=https://s3.amazonaws.com
REMOTE_S3_ACCESS_KEY=AKIA...
REMOTE_S3_SECRET_KEY=...
REMOTE_S3_BUCKET=my-backups
REMOTE_S3_REGION=eu-central-1
```
</details>

<details>
<summary>Backblaze B2</summary>

```bash
REMOTE_SYNC_ENABLED=true
REMOTE_S3_ENDPOINT=https://s3.eu-central-003.backblazeb2.com
REMOTE_S3_ACCESS_KEY=your-key-id
REMOTE_S3_SECRET_KEY=your-application-key
REMOTE_S3_BUCKET=my-backups
```
</details>

<details>
<summary>Yandex Object Storage</summary>

```bash
REMOTE_SYNC_ENABLED=true
REMOTE_S3_ENDPOINT=https://storage.yandexcloud.net
REMOTE_S3_ACCESS_KEY=your-access-key
REMOTE_S3_SECRET_KEY=your-secret-key
REMOTE_S3_BUCKET=my-backups
REMOTE_S3_REGION=ru-central1
```
</details>

<details>
<summary>Selectel S3</summary>

```bash
REMOTE_SYNC_ENABLED=true
REMOTE_S3_ENDPOINT=https://s3.selcdn.ru
REMOTE_S3_ACCESS_KEY=your-access-key
REMOTE_S3_SECRET_KEY=your-secret-key
REMOTE_S3_BUCKET=my-backups
```
</details>

## Структура бэкапов

```
/backups/
├── postgres/
│   ├── postgres_20250120_020000.sql.gz      # Обычный бэкап
│   ├── postgres_20250120_020000.sql.gz.age  # Зашифрованный age
│   └── postgres_20250120_020000.sql.gz.gpg  # Зашифрованный GPG
└── minio/
    ├── minio_20250120_020000.tar.gz
    └── minio_20250120_020000.tar.gz.age

# Remote S3 структура
s3://my-backups/
└── backups/
    └── production/           # SERVER_ID
        ├── postgres/
        └── minio/
```

## Примеры расписаний

```bash
# Каждый день в 2:00
BACKUP_SCHEDULE="0 2 * * *"

# Каждые 6 часов
BACKUP_SCHEDULE="0 */6 * * *"

# Каждый понедельник в 3:00
BACKUP_SCHEDULE="0 3 * * 1"

# Каждый час
BACKUP_SCHEDULE="0 * * * *"
```

## Стратегия 3-2-1

Рекомендуемая стратегия резервного копирования:

- **3** копии данных (оригинал + локальный бэкап + offsite)
- **2** разных типа носителей (Docker volume + S3)
- **1** копия offsite (удалённое хранилище)

```
┌─────────────────┐         ┌─────────────────┐
│  Production     │         │  Remote S3      │
│  Server         │         │  (Offsite)      │
│                 │  sync   │                 │
│  PostgreSQL ────┼────────►│  Encrypted      │
│  MinIO     ─────┤         │  Backups        │
│                 │         └─────────────────┘
│  Local Backups  │
│  (Encrypted)    │
└─────────────────┘
```

## CI/CD Тестирование

Бэкапы автоматически тестируются в GitHub Actions:

- Создание бэкапа PostgreSQL
- Создание зашифрованного бэкапа
- Восстановление бэкапа
- Проверка целостности данных
- Создание бэкапа MinIO
- Проверка метрик

Workflow запускается:
- При изменениях в `backup/`
- Еженедельно по воскресеньям
- Вручную через workflow_dispatch

## Восстановление после сбоя

1. Убедитесь что сервисы postgres и minio запущены
2. Выберите нужный бэкап
3. При необходимости расшифруйте бэкап
4. Выполните восстановление с `RESTORE_CONFIRM=true`
5. Перезапустите backend для переподключения

```bash
# Пример полного восстановления
docker compose up -d postgres minio

# Расшифровка (если включено шифрование)
docker compose run --rm backup sh -c "age -d -i /path/to/key.txt -o /backups/postgres/latest.sql.gz /backups/postgres/latest.sql.gz.age"

# Восстановление
docker compose run --rm -e RESTORE_CONFIRM=true backup /scripts/restore-postgres.sh
docker compose run --rm -e RESTORE_CONFIRM=true backup /scripts/restore-minio.sh
docker compose restart backend
```

## Безопасность

- Бэкапы могут быть зашифрованы с age или GPG
- Пароли передаются через переменные окружения
- Webhook запросы могут быть подписаны HMAC
- Рекомендуется использовать offsite sync для критических данных
- CI/CD автоматически проверяет целостность бэкапов
