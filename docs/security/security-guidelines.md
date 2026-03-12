# 🔒 Руководство по безопасности

## 📋 Обзор безопасности

Comprehensive security framework для защиты информационной системы от современных угроз и обеспечения соответствия требованиям безопасности образовательных учреждений.

## 🎯 Уровень безопасности: СРЕДНИЙ

### Защита от основных угроз:
- ✅ SQL Injection
- ✅ XSS (Cross-Site Scripting)
- ✅ CSRF (Cross-Site Request Forgery)
- ✅ Broken Authentication
- ✅ Security Misconfiguration
- ✅ Insecure Direct Object References
- ✅ DDoS атаки базового уровня

### Нагрузочная устойчивость:
- **Целевая нагрузка**: 3000-5000 пользователей
- **Пиковая нагрузка**: до 1000 одновременных соединений
- **Rate limiting**: защита от перегрузки API

---

## 🔐 Аутентификация и авторизация

### OAuth 2.0 Implementation

#### Поддерживаемые провайдеры:
- **Google Workspace** - для внешних пользователей
- **Microsoft Azure AD** - корпоративная интеграция
- **Собственный IdP** - для специфических требований

#### JWT Токены:
```yaml
token_configuration:
  access_token:
    ttl: "15m"
    algorithm: "RS256"
    issuer: "inf-sys-auth"

  refresh_token:
    ttl: "7d"
    rotation: true
    max_lifetime: "30d"

  security:
    key_rotation: "90d"
    token_binding: true
    secure_storage: true
```

#### Session Management:
- **Redis** для хранения сессий
- **Sliding window** для продления активных сессий
- **Concurrent session limiting** по ролям
- **Device tracking** для безопасности

---

## 🛡️ API Security

### Authentication Layer:
```go
// Middleware структура
func AuthMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        // JWT validation
        // Role verification
        // Rate limiting
        // Request logging
    })
}
```

### Authorization Patterns:

#### RBAC (Role-Based Access Control):
```yaml
roles:
  методист:
    permissions:
      - "documents:create"
      - "documents:read"
      - "curriculum:manage"
      - "reports:generate"

  секретарь:
    permissions:
      - "schedule:manage"
      - "documents:read"
      - "users:view"
      - "reports:create"
```

#### ABAC (Attribute-Based Access Control) для документов:
```yaml
policies:
  document_access:
    - effect: "allow"
      condition: "user.role == 'методист' AND resource.type == 'curriculum'"
    - effect: "allow"
      condition: "user.department == resource.department AND action == 'read'"
```

### API Protection:

#### Rate Limiting:
```yaml
rate_limits:
  authenticated_users:
    requests_per_minute: 60
    burst: 10

  public_endpoints:
    requests_per_minute: 10
    burst: 5

  admin_operations:
    requests_per_minute: 30
    burst: 5
```

#### Input Validation:
- **Строгая валидация** всех входных данных
- **Sanitization** пользовательского ввода
- **File upload restrictions**: тип, размер, содержимое
- **SQL параметризация** для всех запросов

---

## 🗃️ Безопасность данных

### Encryption at Rest:
```yaml
database_encryption:
  postgresql:
    method: "AES-256-GCM"
    key_management: "Vault/KMS"
    sensitive_fields: ["password_hash", "personal_data", "documents"]

file_storage:
  encryption: "AES-256"
  key_rotation: "quarterly"
```

### Encryption in Transit:
- **TLS 1.3** для всех HTTP соединений
- **mTLS** между микросервисами
- **Certificate management** через Let's Encrypt + внутренний CA

### Data Classification:
| Уровень | Примеры данных | Защита |
|---------|---------------|--------|
| **Публичные** | Расписание занятий | Стандартная |
| **Внутренние** | Отчеты, планы | TLS + авторизация |
| **Конфиденциальные** | Персональные данные | Encryption + audit |
| **Строго конфиденциальные** | Пароли, ключи | Vault + HSM |

---

## 🚨 Защита от уязвимостей OWASP Top 10

### 1. **Injection Prevention**
```go
// Параметризованные запросы
func GetDocumentsByUser(userID int) ([]Document, error) {
    query := `SELECT * FROM documents WHERE user_id = $1`
    rows, err := db.Query(query, userID) // Безопасно!
    // НЕ ИСПОЛЬЗОВАТЬ: fmt.Sprintf("SELECT * FROM documents WHERE user_id = %d", userID)
}
```

### 2. **Broken Authentication Prevention**
- Multi-factor authentication для админов
- Account lockout после 5 неудачных попыток
- Password policy enforcement
- Session timeout для неактивных пользователей

### 3. **Sensitive Data Exposure Prevention**
- Логирование без sensitive данных
- Маскирование PII в логах
- Secure headers (HSTS, CSP, etc.)

### 4. **XSS Prevention**
```javascript
// React автоматически экранирует, но дополнительно:
import DOMPurify from 'dompurify';

function SafeHTML({ content }) {
  const cleanHTML = DOMPurify.sanitize(content);
  return <div dangerouslySetInnerHTML={{ __html: cleanHTML }} />;
}
```

### 5. **Broken Access Control Prevention**
- Принцип least privilege
- Regular access reviews
- Automated access provisioning/deprovisioning

---

## 🔍 Security Monitoring

### Logging & Auditing:
```yaml
security_events:
  authentication:
    - login_attempts
    - failed_logins
    - password_changes
    - token_refreshes

  authorization:
    - permission_denied
    - role_changes
    - access_pattern_anomalies

  data_access:
    - sensitive_data_access
    - bulk_data_exports
    - unauthorized_access_attempts
```

### SIEM Integration:
- **Centralized logging** через ELK Stack
- **Real-time alerting** на подозрительную активность
- **Compliance reporting** для аудитов

### Security Metrics:
- **Failed login rate** по пользователям и IP
- **Unusual access patterns** detection
- **API abuse** мониторинг
- **Performance anomalies** как индикатор атак

---

## 🛠️ DevSecOps Pipeline

### Security в CI/CD:

```yaml
pipeline_security:
  code_analysis:
    - gosec: "статический анализ Go"
    - semgrep: "универсальный SAST"
    - dependency_check: "проверка зависимостей"

  container_security:
    - trivy: "сканирование Docker images"
    - clair: "уязвимости в контейнерах"
    - admission_controllers: "Kubernetes security policies"

  runtime_security:
    - falco: "runtime threat detection"
    - network_policies: "ограничение сетевого трафика"
    - pod_security_standards: "secure pod configurations"
```

### Автоматические проверки:
- **Блокирование deployment** при критических уязвимостях
- **Автоматическое создание tickets** для security issues
- **Rollback** при обнаружении проблем в production

---

## 📋 Compliance и соответствие

### ФЗ-152 "О персональных данных":
- **Согласие на обработку** персональных данных
- **Минимизация** объема обрабатываемых данных
- **Уведомление** о нарушениях в течение 24 часов
- **Право на забвение** - удаление данных по запросу

### Внутренние политики:
- **Password policy**: минимум 8 символов, сложность
- **Data retention**: автоматическое удаление старых данных
- **Backup security**: шифрование резервных копий
- **Incident response**: план реагирования на инциденты

---

## 🚨 Incident Response Plan

### Классификация инцидентов:

**Критический (P0):**
- Компрометация административных аккаунтов
- Утечка персональных данных
- Полная недоступность системы

**Высокий (P1):**
- Несанкционированный доступ к документам
- Подозрительная активность в логах
- Частичная недоступность критических функций

**Средний (P2):**
- Превышение rate limits
- Неудачные попытки аутентификации
- Performance деградация

### Response procedure:
1. **Detection** (автоматический/ручной)
2. **Assessment** (оценка масштаба)
3. **Containment** (изоляция угрозы)
4. **Investigation** (анализ причин)
5. **Resolution** (устранение)
6. **Recovery** (восстановление)
7. **Post-incident review** (извлечение уроков)

---

## 🔧 Security Configuration

### Environment Security:

#### Development:
- Изолированные тестовые данные
- Ограниченный доступ к production данным
- Security scanning в local development

#### Staging:
- Production-like security configuration
- Limited access credentials
- Regular security testing

#### Production:
- **Zero trust network** architecture
- **Secret management** через Vault
- **Network segmentation** и firewalls
- **Regular security updates** и патчи

### Infrastructure Security:
```yaml
kubernetes_security:
  pod_security:
    - non_root_containers: true
    - read_only_root_filesystem: true
    - no_privilege_escalation: true

  network_policies:
    - deny_all_default: true
    - explicit_service_communication: true
    - ingress_whitelist: true

  rbac:
    - least_privilege_service_accounts: true
    - regular_permission_audits: true
```

---

## 📚 Security Training

### Для команды разработки:
- **Secure coding practices** - ежеквартально
- **OWASP Top 10** awareness
- **Threat modeling** для новых features
- **Security code review** процедуры

### Для пользователей:
- **Password security** обучение
- **Phishing awareness** training
- **Data handling** best practices
- **Incident reporting** процедуры

---

## 📊 Security Testing Schedule

### Continuous:
- **SAST** в каждом commit
- **Dependency scanning** daily
- **Container scanning** при build

### Weekly:
- **DAST** automated scanning
- **Security metrics** review
- **Log analysis** для аномалий

### Monthly:
- **Penetration testing** (automated)
- **Access review** audit
- **Security training** updates

### Quarterly:
- **Manual pen-testing** от внешних экспертов
- **Security architecture** review
- **Disaster recovery** testing
- **Compliance** audit
---

**📅 Актуальность документа**  
**Последнее обновление**: 2025-01-15  
**Версия проекта**: 0.2.0  
**Статус**: Актуальный
