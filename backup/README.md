# Backup & Restore System

Система резервного копирования для PostgreSQL и MinIO.

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

### Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `BACKUP_MODE` | `cron` | Режим работы: `cron`, `backup`, `restore-postgres`, `restore-minio`, `shell` |
| `BACKUP_SCHEDULE` | `0 2 * * *` | Cron расписание (по умолчанию: 2:00 каждый день) |
| `BACKUP_ON_START` | `false` | Выполнить бэкап при старте контейнера |
| `POSTGRES_BACKUP_RETENTION` | `7` | Хранить бэкапы PostgreSQL N дней |
| `MINIO_BACKUP_RETENTION` | `7` | Хранить бэкапы MinIO N дней |

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
| `SERVER_ID` | `production` | Идентификатор сервера (для multi-server) |

### Примеры расписаний

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

## Структура бэкапов

```
/backups/
├── postgres/
│   ├── postgres_20250120_020000.sql.gz
│   ├── postgres_20250121_020000.sql.gz
│   └── ...
└── minio/
    ├── minio_20250120_020000.tar.gz
    ├── minio_20250121_020000.tar.gz
    └── ...
```

## Доступ к бэкапам

### Просмотр списка бэкапов

```bash
# Войти в контейнер
docker compose run --rm backup /bin/bash

# Внутри контейнера
ls -lh /backups/postgres/
ls -lh /backups/minio/
```

### Копирование на хост

```bash
# Получить ID volume
docker volume inspect inf-sys-secretary-methodist_backup_data

# Скопировать бэкап
docker compose run --rm backup cat /backups/postgres/postgres_20250120_020000.sql.gz > ./backup.sql.gz
```

## Мониторинг

### Просмотр логов

```bash
# Логи cron
docker compose logs -f backup

# Логи внутри контейнера
docker compose exec backup cat /var/log/backup.log
```

## Безопасность

- Бэкапы хранятся в Docker volume `backup_data`
- Пароли передаются через переменные окружения
- Рекомендуется настроить репликацию бэкапов на внешнее хранилище

## Восстановление после сбоя

1. Убедитесь что сервисы postgres и minio запущены
2. Выберите нужный бэкап
3. Выполните восстановление с `RESTORE_CONFIRM=true`
4. Перезапустите backend для переподключения

```bash
# Пример полного восстановления
docker compose up -d postgres minio
docker compose run --rm -e RESTORE_CONFIRM=true backup /scripts/restore-postgres.sh
docker compose run --rm -e RESTORE_CONFIRM=true backup /scripts/restore-minio.sh
docker compose restart backend
```

## Remote Sync (Offsite Backup)

Для надёжности бэкапы автоматически синхронизируются на внешнее S3-совместимое хранилище.

### Поддерживаемые провайдеры

#### AWS S3

```bash
REMOTE_SYNC_ENABLED=true
REMOTE_S3_ENDPOINT=https://s3.amazonaws.com
REMOTE_S3_ACCESS_KEY=AKIA...
REMOTE_S3_SECRET_KEY=...
REMOTE_S3_BUCKET=my-backups
REMOTE_S3_REGION=eu-central-1
```

#### Backblaze B2

```bash
REMOTE_SYNC_ENABLED=true
REMOTE_S3_ENDPOINT=https://s3.eu-central-003.backblazeb2.com
REMOTE_S3_ACCESS_KEY=your-key-id
REMOTE_S3_SECRET_KEY=your-application-key
REMOTE_S3_BUCKET=my-backups
```

#### Yandex Object Storage

```bash
REMOTE_SYNC_ENABLED=true
REMOTE_S3_ENDPOINT=https://storage.yandexcloud.net
REMOTE_S3_ACCESS_KEY=your-access-key
REMOTE_S3_SECRET_KEY=your-secret-key
REMOTE_S3_BUCKET=my-backups
REMOTE_S3_REGION=ru-central1
```

#### Selectel S3

```bash
REMOTE_SYNC_ENABLED=true
REMOTE_S3_ENDPOINT=https://s3.selcdn.ru
REMOTE_S3_ACCESS_KEY=your-access-key
REMOTE_S3_SECRET_KEY=your-secret-key
REMOTE_S3_BUCKET=my-backups
```

#### Self-hosted MinIO

```bash
REMOTE_SYNC_ENABLED=true
REMOTE_S3_ENDPOINT=https://minio.backup-server.com
REMOTE_S3_ACCESS_KEY=minioadmin
REMOTE_S3_SECRET_KEY=minioadmin
REMOTE_S3_BUCKET=offsite-backups
```

### Структура на удалённом хранилище

```
s3://my-backups/
└── backups/
    └── production/           # SERVER_ID
        ├── postgres/
        │   └── postgres_*.sql.gz
        └── minio/
            └── minio_*.tar.gz
```

### Multi-server setup

Для нескольких серверов используйте разные `SERVER_ID`:

```bash
# Сервер 1
SERVER_ID=prod-web-1

# Сервер 2
SERVER_ID=prod-web-2
```

### Ручная синхронизация

```bash
# Синхронизировать бэкапы на удалённое хранилище
docker compose run --rm backup /scripts/sync-to-remote.sh
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
│  PostgreSQL ────┼────────►│  Backups        │
│  MinIO     ─────┤         │                 │
│                 │         └─────────────────┘
│  Local Backups  │
│  (Docker Vol)   │
└─────────────────┘
```
