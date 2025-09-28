# 🧪 Стратегия тестирования

## 📋 Обзор стратегии

Комплексная стратегия тестирования для обеспечения качества и надежности микросервисной системы управления документооборотом.

## 🏗️ Пирамида тестирования

```
        ┌─────────────┐
        │     E2E     │ 10%
        │   Testing   │
        └─────────────┘
      ┌─────────────────┐
      │   Integration   │ 20%
      │    Testing      │
      └─────────────────┘
    ┌───────────────────────┐
    │    Unit Testing       │ 70%
    │   (Service Level)     │
    └───────────────────────┘
```

## 🔬 Уровни тестирования

### 1. **Unit Testing** (70% от всех тестов)

#### Цели:
- Тестирование отдельных функций и методов
- Быстрая обратная связь разработчикам
- Высокое покрытие кода (минимум 80%)

#### Go Backend:
```go
// Инструменты
- testing (встроенный пакет Go)
- testify - assertions и mocks
- gomock - генерация моков
- gofuzz - фаззинг тестирование

// Пример структуры теста
func TestDocumentService_CreateDocument(t *testing.T) {
    // Arrange
    mockRepo := &MockDocumentRepository{}
    service := NewDocumentService(mockRepo)

    // Act & Assert
    // ...
}
```

#### Frontend (Next.js):
```javascript
// Инструменты
- Jest - test runner
- React Testing Library - компонентное тестирование
- MSW (Mock Service Worker) - мокирование API

// Пример теста компонента
describe('DocumentForm', () => {
  test('should submit valid document data', async () => {
    // ...
  });
});
```

#### Покрытие по сервисам:
| Сервис | Минимальное покрытие | Целевое покрытие |
|--------|---------------------|------------------|
| auth-service | 85% | 95% |
| document-service | 80% | 90% |
| workflow-service | 80% | 90% |
| user-service | 85% | 95% |
| Остальные сервисы | 75% | 85% |

---

### 2. **Integration Testing** (20% от всех тестов)

#### Цели:
- Тестирование взаимодействия между сервисами
- Проверка API контрактов
- Тестирование баз данных

#### Типы интеграционных тестов:

**API Testing:**
```go
// Использование testcontainers для изолированного тестирования
func TestDocumentAPI_Integration(t *testing.T) {
    // Запуск контейнера PostgreSQL
    postgres := testcontainers.PostgreSQLContainer{}

    // Тестирование API endpoints
    // ...
}
```

**Database Testing:**
- Тестирование миграций
- Проверка индексов и производительности
- Тестирование транзакций

**Service-to-Service Testing:**
- Контрактное тестирование (Pact)
- Тестирование очередей сообщений
- Проверка межсервисных вызовов

#### Инструменты:
- **Testcontainers** - изолированное тестирование с БД
- **Pact** - contract testing
- **Newman** - автоматизированное тестирование API (Postman)

---

### 3. **End-to-End Testing** (10% от всех тестов)

#### Цели:
- Тестирование полных пользовательских сценариев
- Проверка интеграции всех компонентов
- Валидация бизнес-процессов

#### Критические сценарии для E2E:

**1. Создание и согласование учебного плана:**
```gherkin
Feature: Учебный план - полный жизненный цикл
  Scenario: Успешное создание и утверждение учебного плана
    Given методист авторизован в системе
    When создает новый учебный план
    And заполняет обязательные поля
    And отправляет на согласование
    Then план проходит все этапы workflow
    And становится доступен для просмотра преподавателям
```

**2. Workflow согласования документов:**
- Создание документа
- Прохождение этапов согласования
- Обработка отклонений и корректировок
- Финальное утверждение

**3. Управление пользователями:**
- Регистрация нового пользователя
- Назначение ролей и прав
- Смена пароля и профиля

#### Инструменты E2E:
- **Playwright** - основной инструмент для веб-тестирования
- **Docker Compose** - развертывание тестовой среды
- **Kubernetes** - тестирование в продакшн-подобной среде

---

### 4. **Performance Testing** 🚀

#### Цели:
- Проверка производительности под нагрузкой
- Выявление узких мест
- Планирование масштабирования

#### Типы нагрузочного тестирования:

**Load Testing:**
- **Нормальная нагрузка**: 500 одновременных пользователей
- **Пиковая нагрузка**: 1000 одновременных пользователей
- **Продолжительность**: 30 минут

**Stress Testing:**
- **Максимальная нагрузка**: 2000+ пользователей
- **Проверка деградации**: graceful degradation
- **Recovery testing**: восстановление после перегрузки

**Volume Testing:**
- **Большие объемы данных**: 1M+ документов
- **Долгосрочное тестирование**: 24 часа непрерывной работы

#### Инструменты:
```yaml
# k6 - основной инструмент нагрузочного тестирования
tools:
  - k6: "основной load testing"
  - Artillery: "альтернативный инструмент"
  - JMeter: "GUI для создания сценариев"
  - Gatling: "enterprise grade testing"
```

#### Метрики производительности:
| Метрика | Целевое значение | Критическое значение |
|---------|------------------|---------------------|
| Response Time (API) | < 200ms | < 500ms |
| Page Load Time | < 2s | < 5s |
| Throughput | > 1000 RPS | > 500 RPS |
| Error Rate | < 0.1% | < 1% |
| CPU Utilization | < 70% | < 90% |
| Memory Usage | < 80% | < 95% |

---

### 5. **Security Testing** 🔒

#### Области тестирования:
- **Аутентификация и авторизация**
- **Input validation и защита от инъекций**
- **Защита API endpoints**
- **Шифрование данных**

#### Типы security testing:

**SAST (Static Application Security Testing):**
```yaml
tools:
  go:
    - gosec: "статический анализ Go кода"
    - nancy: "проверка уязвимостей в зависимостях"
  javascript:
    - eslint-plugin-security: "статический анализ JS"
    - npm audit: "проверка npm пакетов"
```

**DAST (Dynamic Application Security Testing):**
- **OWASP ZAP** - автоматическое сканирование веб-приложений
- **Burp Suite** - ручное тестирование безопасности
- **SQLMap** - тестирование SQL инъекций

**Penetration Testing:**
- Ежеквартальное pen-testing
- Bug bounty программа (планируется)

#### Security тест-кейсы:
1. **SQL Injection защита**
2. **XSS защита**
3. **CSRF токены**
4. **Rate limiting**
5. **JWT токен безопасность**
6. **File upload безопасность**

---

## 🚀 CI/CD Pipeline для тестирования

### Stages в pipeline:

```yaml
stages:
  1. code_quality:
    - linting (golangci-lint, eslint)
    - code formatting check
    - dependency scanning

  2. unit_tests:
    - backend: "go test ./..."
    - frontend: "npm test"
    - coverage: минимум 80%

  3. integration_tests:
    - API tests с testcontainers
    - Database integration tests
    - Service contract tests

  4. security_tests:
    - SAST scanning
    - Dependency vulnerability check
    - Docker image scanning

  5. build_and_package:
    - Docker image build
    - Helm chart packaging

  6. deploy_staging:
    - Deployment в staging окружение
    - Smoke tests

  7. e2e_tests:
    - Playwright tests на staging
    - Performance baseline tests

  8. deploy_production:
    - Blue-green deployment
    - Health checks
    - Monitoring validation
```

### Параллелизация:
- Unit и integration тесты выполняются параллельно
- Security сканирование параллельно с тестированием
- E2E тесты только после успешного staging deployment

---

## 📊 Test Data Management

### Стратегия тестовых данных:

**Синтетические данные:**
- Генерация с помощью библиотек (Faker, Gofakeit)
- Соответствие реальным форматам
- GDPR compliance

**Анонимизированные данные:**
- Копии продакшн данных с удаленной PII
- Регулярное обновление датасетов
- Контроль доступа к тестовым данным

**Тестовые пользователи:**
```yaml
test_users:
  - role: "методист"
    username: "metodist.test"
    permissions: ["create_curriculum", "review_documents"]
  - role: "секретарь"
    username: "secretary.test"
    permissions: ["manage_schedule", "create_reports"]
```

---

## 📈 Метрики и отчетность

### KPI тестирования:
- **Test Coverage**: >80% для критических компонентов
- **Test Execution Time**: <15 минут для полного suite
- **Flaky Test Rate**: <2%
- **Bug Escape Rate**: <5% (багов, пропущенных в production)

### Дашборды:
- **Allure Report** - детальные отчеты по тестированию
- **SonarQube** - качество кода и покрытие
- **Grafana** - метрики производительности тестов

### Еженедельные отчеты:
- Статистика прохождения тестов
- Анализ failed тестов
- Performance регрессии
- Планы улучшения качества

---

## 🔄 Maintenance и улучшения

### Регулярное обслуживание:
- **Еженедельно**: Анализ flaky tests
- **Ежемесячно**: Обновление тестовых данных
- **Ежеквартально**: Ревизия тест-стратегии
- **Полугодично**: Performance baseline обновление

### Continuous Improvement:
- Автоматическое обнаружение медленных тестов
- A/B тестирование новых подходов
- Обучение команды новым техникам тестирования
- Внедрение лучших практик industry