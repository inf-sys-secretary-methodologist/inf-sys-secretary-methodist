# Администрирование сервера Timeweb

## Подключение

```bash
# SSH подключение (порт 2222!)
ssh -p 2222 deploy@5.129.253.112

# Если нужен root доступ (после подключения как deploy)
sudo -i
```

## URLs

| Сервис | URL |
|--------|-----|
| Frontend | https://5-129-253-112.sslip.io |
| API | https://api.5-129-253-112.sslip.io |
| MinIO Console | https://minio.5-129-253-112.sslip.io |

---

## Docker Compose

```bash
# Перейти в директорию проекта
cd /root/inf-sys-secretary-methodist

# Статус всех контейнеров
docker compose ps

# Статус с подробностями
docker compose ps -a

# Запустить все сервисы
docker compose up -d

# Остановить все сервисы
docker compose down

# Перезапустить конкретный сервис
docker compose restart backend
docker compose restart frontend
docker compose restart postgres
docker compose restart redis
docker compose restart minio

# Пересобрать и перезапустить (после git pull)
docker compose up -d --build

# Пересобрать конкретный сервис
docker compose up -d --build backend
docker compose up -d --build frontend
```

## Логи контейнеров

```bash
# Логи всех сервисов (последние 100 строк)
docker compose logs --tail=100

# Логи конкретного сервиса
docker compose logs backend --tail=100
docker compose logs frontend --tail=100
docker compose logs postgres --tail=100

# Следить за логами в реальном времени
docker compose logs -f backend
docker compose logs -f frontend

# Логи за последний час
docker compose logs --since=1h backend
```

## База данных (PostgreSQL)

```bash
# Подключиться к PostgreSQL
docker compose exec postgres psql -U postgres -d inf_sys_db

# Выполнить SQL запрос
docker compose exec postgres psql -U postgres -d inf_sys_db -c "SELECT COUNT(*) FROM users;"

# Список таблиц
docker compose exec postgres psql -U postgres -d inf_sys_db -c "\dt"

# Дамп базы данных
docker compose exec postgres pg_dump -U postgres inf_sys_db > backup_$(date +%Y%m%d_%H%M%S).sql

# Восстановить из дампа
docker compose exec -T postgres psql -U postgres inf_sys_db < backup.sql
```

## Redis

```bash
# Подключиться к Redis CLI
docker compose exec redis redis-cli

# Проверить подключение
docker compose exec redis redis-cli ping

# Посмотреть все ключи
docker compose exec redis redis-cli KEYS "*"

# Очистить кэш
docker compose exec redis redis-cli FLUSHALL
```

## MinIO (S3 хранилище)

```bash
# Список бакетов
docker compose exec minio mc ls local

# Список файлов в бакете documents
docker compose exec minio mc ls local/documents

# Размер бакета
docker compose exec minio mc du local/documents
```

---

## Безопасность

### Fail2Ban

```bash
# Статус Fail2Ban
sudo fail2ban-client status

# Статус SSH jail (забаненные IP)
sudo fail2ban-client status sshd

# Разбанить IP
sudo fail2ban-client set sshd unbanip 1.2.3.4

# Забанить IP вручную
sudo fail2ban-client set sshd banip 1.2.3.4

# Перезапустить Fail2Ban
sudo systemctl restart fail2ban

# Логи Fail2Ban
sudo tail -100 /var/log/fail2ban.log
```

### UFW (Firewall)

```bash
# Статус файрвола
sudo ufw status
sudo ufw status verbose
sudo ufw status numbered

# Открыть порт
sudo ufw allow 8080/tcp

# Закрыть порт
sudo ufw delete allow 8080/tcp

# Заблокировать IP
sudo ufw deny from 1.2.3.4

# Разблокировать IP
sudo ufw delete deny from 1.2.3.4

# Перезагрузить правила
sudo ufw reload
```

### SSH

```bash
# Активные SSH сессии
who
w

# История входов
last -20

# Неудачные попытки входа
sudo grep "Failed password" /var/log/auth.log | tail -20

# Конфигурация SSH
sudo nano /etc/ssh/sshd_config

# Перезапустить SSH (осторожно!)
sudo systemctl restart ssh.socket
```

---

## Caddy (Reverse Proxy + SSL)

```bash
# Статус Caddy
sudo systemctl status caddy

# Перезапустить Caddy
sudo systemctl restart caddy

# Перезагрузить конфиг без даунтайма
sudo systemctl reload caddy

# Редактировать конфиг
sudo nano /etc/caddy/Caddyfile

# Проверить конфиг на ошибки
sudo caddy validate --config /etc/caddy/Caddyfile

# Форматировать конфиг
sudo caddy fmt --overwrite /etc/caddy/Caddyfile

# Логи Caddy
sudo journalctl -u caddy --since="1 hour ago"
sudo journalctl -u caddy -f  # в реальном времени
```

---

## Бэкапы

```bash
# Запустить бэкап вручную
docker compose --profile backup run --rm backup

# Проверить статус backup сервиса
docker compose --profile backup ps

# Посмотреть логи бэкапа
docker compose --profile backup logs backup

# Где хранятся бэкапы (volume)
docker volume inspect inf-sys-secretary-methodist_backup_data
```

---

## Мониторинг системы

```bash
# Использование диска
df -h

# Использование памяти
free -h

# Загрузка CPU и память по процессам
htop
# или
top

# Использование диска Docker
docker system df

# Очистка неиспользуемых Docker ресурсов
docker system prune -a  # ОСТОРОЖНО: удалит неиспользуемые образы!

# Размер директорий
du -sh /root/inf-sys-secretary-methodist
du -sh /var/lib/docker

# Сетевые соединения
ss -tlnp

# Проверка портов
sudo netstat -tlnp
```

---

## Git и деплой

```bash
cd /root/inf-sys-secretary-methodist

# Получить последние изменения
git pull origin main

# Пересобрать и перезапустить после изменений
docker compose up -d --build

# Посмотреть статус
git status

# Посмотреть последние коммиты
git log --oneline -10
```

---

## Миграции базы данных

```bash
# Применить все миграции
docker compose exec backend ./server migrate up

# Откатить последнюю миграцию
docker compose exec backend ./server migrate down 1

# Статус миграций
docker compose exec backend ./server migrate version
```

---

## Полезные алиасы

Добавь в `~/.bashrc` на сервере:

```bash
# Добавить алиасы
cat >> ~/.bashrc << 'EOF'

# Docker Compose aliases
alias dc="docker compose"
alias dps="docker compose ps"
alias dlogs="docker compose logs"
alias dlogsf="docker compose logs -f"
alias drestart="docker compose restart"
alias drebuild="docker compose up -d --build"

# Быстрый переход в проект
alias proj="cd /root/inf-sys-secretary-methodist"

# Fail2Ban
alias f2bstatus="sudo fail2ban-client status sshd"
alias f2bunban="sudo fail2ban-client set sshd unbanip"

# Логи
alias caddylogs="sudo journalctl -u caddy -f"
EOF

# Применить
source ~/.bashrc
```

---

## Troubleshooting

### Контейнер не запускается

```bash
# Посмотреть логи проблемного контейнера
docker compose logs backend --tail=200

# Проверить что контейнер упал
docker compose ps -a

# Попробовать запустить интерактивно для отладки
docker compose run --rm backend sh
```

### Нет места на диске

```bash
# Проверить использование
df -h

# Очистить Docker
docker system prune -a --volumes  # ОСТОРОЖНО!

# Очистить логи Docker
sudo truncate -s 0 /var/lib/docker/containers/*/*-json.log
```

### Сайт недоступен

```bash
# Проверить Caddy
sudo systemctl status caddy

# Проверить контейнеры
docker compose ps

# Проверить порты
ss -tlnp | grep -E '(3000|8080|9000)'

# Проверить файрвол
sudo ufw status
```

### Потерял доступ по SSH

1. Войти через **VNC-консоль** в панели Timeweb
2. Проверить Fail2Ban: `fail2ban-client status sshd`
3. Разбанить свой IP: `fail2ban-client set sshd unbanip ТВОЙ_IP`
4. Проверить SSH: `systemctl status ssh.socket`

---

## Важные файлы

| Файл | Описание |
|------|----------|
| `/root/inf-sys-secretary-methodist/.env` | Переменные окружения |
| `/root/inf-sys-secretary-methodist/compose.yml` | Docker Compose конфиг |
| `/etc/caddy/Caddyfile` | Конфиг Caddy (reverse proxy) |
| `/etc/ssh/sshd_config` | Конфиг SSH |
| `/etc/fail2ban/jail.local` | Конфиг Fail2Ban |
| `/etc/ssh/sshrc` | Скрипт Telegram уведомлений |

---

## Контакты и доступы

- **IP сервера:** 5.129.253.112
- **SSH порт:** 2222
- **Пользователь:** deploy (sudo)
- **Панель Timeweb:** https://timeweb.cloud

---

*Последнее обновление: Январь 2026*
