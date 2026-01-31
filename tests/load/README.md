# Load Testing with k6

Нагрузочное тестирование API с использованием [k6](https://k6.io/).

## Установка k6

### macOS
```bash
brew install k6
```

### Linux
```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

### Docker
```bash
docker run --rm -i grafana/k6 run - <tests/load/health-test.js
```

## Быстрый старт

```bash
# Smoke test (1 user, 30s)
k6 run tests/load/health-test.js

# Load test (50 users)
k6 run -e SCENARIO=load tests/load/api-test.js

# Stress test (до 200 users)
k6 run -e SCENARIO=stress tests/load/api-test.js

# Против production
k6 run -e BASE_URL=https://api.example.com tests/load/health-test.js
```

## Тесты

### health-test.js
Тестирует публичные endpoints без авторизации:
- `/health` - полный health check
- `/live` - liveness probe
- `/ready` - readiness probe
- `/swagger/index.html` - документация

### api-test.js
Тестирует основные API endpoints с авторизацией:
- `POST /api/auth/login` - аутентификация
- `GET /api/documents` - список документов
- `GET /api/events` - события/расписание
- `GET /api/notifications` - уведомления
- `GET /api/conversations` - чаты

## Сценарии

| Сценарий | VUs | Длительность | Назначение |
|----------|-----|--------------|------------|
| `smoke` | 1 | 30s | Базовая проверка работоспособности |
| `load` | 50 | 5m | Нормальная нагрузка |
| `stress` | 200 | 16m | Поиск пределов системы |
| `spike` | 100 | 1.5m | Внезапный всплеск трафика |
| `soak` | 50 | 30m | Длительная стабильность |

## Пороговые значения (Thresholds)

По умолчанию тесты проверяют:
- **p95 < 500ms** - 95% запросов быстрее 500ms
- **p99 < 1000ms** - 99% запросов быстрее 1s
- **error rate < 1%** - менее 1% ошибок
- **throughput > 100 req/s** - более 100 запросов в секунду

## Подготовка тестового пользователя

Для api-test.js нужен тестовый пользователь в БД:

```bash
# Подключиться к БД
docker compose exec postgres psql -U postgres -d inf_sys_db

# Создать тестового пользователя (пароль: LoadTest123!)
INSERT INTO users (email, password_hash, first_name, last_name, role, is_active)
VALUES (
    'loadtest@example.com',
    '$2a$14$...', -- bcrypt hash of LoadTest123!
    'Load',
    'Tester',
    'user',
    true
);
```

Или используйте свои credentials:
```bash
k6 run -e TEST_USER_EMAIL=your@email.com -e TEST_USER_PASSWORD=password tests/load/api-test.js
```

## Вывод результатов

### Консоль
```bash
k6 run tests/load/health-test.js
```

### JSON отчёт
```bash
k6 run --out json=results.json tests/load/health-test.js
```

### InfluxDB + Grafana
```bash
k6 run --out influxdb=http://localhost:8086/k6 tests/load/api-test.js
```

### Grafana Cloud k6
```bash
K6_CLOUD_TOKEN=<token> k6 cloud tests/load/api-test.js
```

## Интеграция с CI/CD

### GitHub Actions
```yaml
- name: Run load tests
  run: |
    curl -sS https://dl.k6.io/k6-latest-linux-amd64.tar.gz | tar xvzf -
    ./k6 run --out json=k6-results.json tests/load/health-test.js

- name: Upload results
  uses: actions/upload-artifact@v4
  with:
    name: k6-results
    path: k6-results.json
```

## Метрики

### Стандартные k6 метрики
- `http_req_duration` - время ответа
- `http_req_failed` - процент ошибок
- `http_reqs` - количество запросов
- `iterations` - завершённые итерации
- `vus` - виртуальные пользователи

### Кастомные метрики
- `auth_duration` - время авторизации
- `documents_duration` - время запросов документов
- `events_duration` - время запросов событий
- `errors` - общий rate ошибок

## Ссылки

- [k6 Documentation](https://k6.io/docs/)
- [k6 Examples](https://github.com/grafana/k6/tree/master/examples)
- [k6 Cloud](https://k6.io/cloud/)
