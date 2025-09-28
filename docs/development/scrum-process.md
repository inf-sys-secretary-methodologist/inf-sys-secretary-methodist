# 🔄 Scrum Process Guide

## 📋 Обзор Scrum процесса

Agile-методология для команды из 5 разработчиков с адаптацией под специфику образовательного проекта и микросервисной архитектуры.

## 👥 Scrum Team

### Роли в команде:

#### Product Owner (1 человек)
- **Ответственность**: Видение продукта, приоритизация задач
- **Обязанности**:
  - Управление Product Backlog
  - Написание User Stories
  - Взаимодействие с заказчиками (методисты, секретари)
  - Acceptance критерии для задач

#### Scrum Master (1 человек)
- **Ответственность**: Процесс разработки, устранение блокеров
- **Обязанности**:
  - Фасилитация Scrum событий
  - Устранение препятствий для команды
  - Коучинг команды по Agile практикам
  - Взаимодействие с заинтересованными сторонами

#### Development Team (3 человека)
```yaml
team_composition:
  backend_developer:
    count: 2
    focus: ["Go microservices", "API design", "database design"]
    secondary: ["DevOps", "integration testing"]

  frontend_developer:
    count: 1
    focus: ["Next.js", "TypeScript", "MUI components"]
    secondary: ["UX/UI", "E2E testing"]

cross_functional_skills:
  - все участвуют в code review
  - все участвуют в архитектурных решениях
  - ротация по DevOps задачам
```

---

## 📅 Sprint Structure

### Sprint Duration: **2 недели**

#### Обоснование:
- Быстрая обратная связь от пользователей
- Регулярные демо для заказчиков
- Возможность адаптации к изменениям требований
- Подходит для размера команды

### Sprint Calendar:
```
Week 1:
  Понедельник: Sprint Planning (4 часа)
  Среда: Daily Standup (15 мин) + Mid-sprint check
  Пятница: Daily Standup + Refinement (1 час)

Week 2:
  Понедельник: Daily Standup
  Среда: Daily Standup
  Пятница: Sprint Review (2 часа) + Retrospective (1.5 часа)
```

---

## 📊 Sprint Events

### 1. **Sprint Planning** (4 часа)

#### Part 1: What (2 часа)
- **Цель**: Определить Sprint Goal и выбрать задачи
- **Участники**: Вся команда
- **Результат**: Sprint Backlog

**Процесс:**
1. Обзор Product Backlog (30 мин)
2. Определение Sprint Goal (30 мин)
3. Выбор User Stories (60 мин)

#### Part 2: How (2 часа)
- **Цель**: Техническое планирование реализации
- **Результат**: Детализированные задачи с оценками

**Процесс:**
1. Разбивка User Stories на задачи (90 мин)
2. Оценка сложности (Planning Poker) (30 мин)

### 2. **Daily Standup** (15 минут)

#### Формат (каждый участник):
1. **Что сделал вчера?**
2. **Что планирую сегодня?**
3. **Есть ли блокеры?**

#### Правила:
- Строго 15 минут
- Обсуждение решений - после standup
- Фокус на Sprint Goal
- Виртуальный формат с возможностью очного участия

### 3. **Sprint Review** (2 часа)

#### Agenda:
1. **Демонстрация** выполненных задач (90 мин)
2. **Feedback** от заказчиков (20 мин)
3. **Планирование** следующего спринта (10 мин)

#### Участники:
- Scrum Team
- Заказчики (методисты, секретари)
- Stakeholders

### 4. **Sprint Retrospective** (1.5 часа)

#### Формат "Start-Stop-Continue":
```yaml
retrospective_format:
  start: "Что начать делать?"
  stop: "Что перестать делать?"
  continue: "Что продолжить делать?"

action_items:
  maximum: 3
  owner: "assigned to team member"
  deadline: "next sprint"
```

---

## 📝 Backlog Management

### Product Backlog Structure:

#### Epic Level:
```yaml
examples:
  - name: "Система управления документами"
    description: "Полный lifecycle документов от создания до архива"
    business_value: "High"
    estimated_sprints: 4-6

  - name: "Интеграция с 1С"
    description: "Двухсторонний обмен данными с 1С"
    business_value: "Medium"
    estimated_sprints: 2-3
```

#### User Story Level:
```yaml
template: |
  Как [роль]
  Я хочу [функциональность]
  Чтобы [бизнес-ценность]

example:
  title: "Создание учебного плана"
  story: |
    Как методист
    Я хочу создать новый учебный план через веб-интерфейс
    Чтобы автоматизировать процесс планирования учебного процесса

  acceptance_criteria:
    - Форма создания учебного плана с валидацией
    - Сохранение в базе данных
    - Автоматическое создание workflow согласования
    - Email уведомление участникам процесса

  story_points: 8
  priority: "High"
  dependencies: ["Система аутентификации", "Workflow engine"]
```

### Definition of Ready (DoR):
- ✅ User Story написана и понятна
- ✅ Acceptance criteria определены
- ✅ Story points оценены
- ✅ Dependencies выявлены
- ✅ UI/UX макеты готовы (при необходимости)
- ✅ Technical approach согласован

### Definition of Done (DoD):
- ✅ Код написан и соответствует стандартам
- ✅ Unit тесты покрывают >80% кода
- ✅ Integration тесты пройдены
- ✅ Code review выполнен
- ✅ Документация API обновлена
- ✅ Acceptance criteria выполнены
- ✅ Deployed в staging и протестировано
- ✅ Product Owner acceptance получен

---

## 📏 Estimation и Velocity

### Story Points Scale (Fibonacci):
```yaml
story_points:
  1: "Trivial task (30 min - 2 hours)"
  2: "Simple feature (half day)"
  3: "Small feature (1 day)"
  5: "Medium feature (2-3 days)"
  8: "Large feature (1 week)"
  13: "Very large feature (1+ week)" # Требует разбивки
  21: "Epic - must be broken down"
```

### Planning Poker Process:
1. **PO презентует** User Story
2. **Команда задает** уточняющие вопросы
3. **Закрытое голосование** по story points
4. **Обсуждение** различий в оценках
5. **Финальная оценка** на основе консенсуса

### Velocity Tracking:
```yaml
velocity_metrics:
  measurement: "story_points_per_sprint"
  team_velocity_target: "25-35 points per sprint"
  individual_capacity: "6-8 points per developer"

tracking:
  sprint_1: 0 # baseline
  sprint_2: 18 # learning curve
  sprint_3: 26 # stabilization
  target_velocity: 30 # after sprint 4-5
```

---

## 🔄 Workflow Integration

### GitHub Integration:

#### Branch Strategy:
```yaml
git_workflow:
  main_branch: "main"
  development_branch: "develop"
  feature_branches: "feature/PROJ-123-description"
  hotfix_branches: "hotfix/PROJ-456-critical-fix"

  branch_protection:
    - require_pr_review: 2_approvals
    - require_status_checks: true
    - require_up_to_date: true
    - restrict_push: true
```

#### GitHub Projects Integration:
```yaml
project_automation:
  columns:
    - "📋 Product Backlog"
    - "🚀 Sprint Backlog"
    - "⚡ In Progress"
    - "👀 In Review"
    - "✅ Done"

  automation_rules:
    - pr_opened: move_to_in_review
    - pr_merged: move_to_done
    - issue_assigned: move_to_in_progress
```

### Task Breakdown:
```yaml
task_hierarchy:
  epic: "Система документооборота"
    user_story: "Создание учебного плана"
      task: "Backend API для создания плана"
      task: "Frontend форма создания"
      task: "Валидация данных"
      task: "Unit тесты"
      task: "Integration тесты"
```

---

## 📊 Metrics и KPIs

### Sprint Metrics:

#### Burndown Chart:
- **Ежедневное обновление** оставшихся story points
- **Ideal burndown line** vs actual progress
- **Scope changes** tracking

#### Velocity Chart:
- **Story points** выполненные за спринт
- **Trend analysis** для планирования capacity
- **Predictability** команды

### Quality Metrics:
```yaml
quality_kpis:
  code_quality:
    - test_coverage: ">80%"
    - code_review_coverage: "100%"
    - technical_debt_ratio: "<5%"

  delivery_quality:
    - escaped_defects: "<2 per sprint"
    - story_completion_rate: ">90%"
    - sprint_goal_achievement: ">80%"

  process_quality:
    - sprint_commitment_accuracy: ">85%"
    - retrospective_action_completion: ">80%"
    - team_satisfaction: ">4/5"
```

---

## 🛠️ Tools и Infrastructure

### Scrum Tools:
```yaml
primary_tools:
  project_management: "GitHub Projects"
  communication: "Slack + Daily meetings"
  documentation: "GitHub Wiki + Markdown"
  estimation: "Planning Poker app"

secondary_tools:
  retrospectives: "Miro/FunRetro"
  burndown_charts: "GitHub insights"
  time_tracking: "Toggl (optional)"
```

### Development Tools Integration:
```yaml
integration_points:
  github:
    - automatic_issue_linking: "commit messages"
    - pr_template: "includes_acceptance_criteria_check"
    - milestone_tracking: "sprint_goals"

  ci_cd:
    - automated_testing: "on_pr_creation"
    - deployment_gates: "requires_story_completion"
    - release_notes: "auto_generated_from_commits"
```

---

## 🎯 Sprint Planning Techniques

### Capacity Planning:
```yaml
capacity_calculation:
  total_sprint_days: 10
  holidays_pto: 1
  meetings_overhead: 1.5
  support_maintenance: 0.5
  effective_development_days: 7

  per_developer:
    story_points_per_day: 1.2
    total_capacity: 8.4_points
    team_capacity: 25_points # 3 developers
```

### Risk Management:
```yaml
risk_mitigation:
  technical_risks:
    - proof_of_concept: "для новых технологий"
    - spike_stories: "для исследования неопределенности"
    - buffer_time: "15% от sprint capacity"

  dependency_risks:
    - early_identification: "во время planning"
    - communication_plan: "с владельцами dependencies"
    - fallback_options: "alternative solutions"
```

---

## 📈 Continuous Improvement

### Retrospective Action Items:

#### Types of Improvements:
```yaml
improvement_categories:
  process:
    - "Улучшение daily standup формата"
    - "Оптимизация code review процесса"
    - "Автоматизация рутинных задач"

  technical:
    - "Рефакторинг legacy кода"
    - "Улучшение архитектуры"
    - "Performance оптимизация"

  team:
    - "Обучение новым технологиям"
    - "Улучшение коммуникации"
    - "Knowledge sharing sessions"
```

### Learning & Development:
```yaml
team_development:
  tech_talks:
    frequency: "bi_weekly"
    duration: "30-45 minutes"
    topics: ["new_technologies", "best_practices", "lessons_learned"]

  code_reviews:
    mandatory: true
    min_reviewers: 2
    focus: ["security", "performance", "maintainability"]

  pair_programming:
    frequency: "as_needed"
    scenarios: ["complex_features", "knowledge_transfer", "debugging"]
```

---

## 🚨 Risk Management в Sprint

### Common Risks:

#### Technical Risks:
```yaml
technical_risks:
  integration_complexity:
    probability: "Medium"
    impact: "High"
    mitigation: "Early prototyping, dedicated integration sprint"

  performance_issues:
    probability: "Medium"
    impact: "Medium"
    mitigation: "Performance testing in each sprint"

  security_vulnerabilities:
    probability: "Low"
    impact: "High"
    mitigation: "Security reviews, automated scanning"
```

#### Process Risks:
```yaml
process_risks:
  scope_creep:
    probability: "High"
    impact: "Medium"
    mitigation: "Strong PO, clear acceptance criteria"

  team_unavailability:
    probability: "Medium"
    impact: "Medium"
    mitigation: "Cross-training, documentation"

  changing_requirements:
    probability: "High"
    impact: "Low"
    mitigation: "Agile mindset, regular stakeholder communication"
```

---

## 📋 Sprint Artifacts

### Sprint Backlog:
```yaml
sprint_backlog_structure:
  sprint_goal: "Clear, measurable objective"
  user_stories:
    - id: "PROJ-123"
      title: "Создание системы аутентификации"
      story_points: 8
      priority: "High"
      assignee: "backend_dev_1"
      status: "In Progress"

  tasks:
    - id: "PROJ-123-1"
      title: "OAuth provider integration"
      estimate: "4 hours"
      assignee: "backend_dev_1"
      status: "To Do"
```

### Sprint Burndown:
- **Ежедневное обновление** остающихся story points
- **Visualization** в GitHub Projects
- **Trend analysis** для future planning

### Definition of Done Checklist:
```yaml
dod_checklist:
  development:
    - code_implemented: true
    - unit_tests_written: true
    - integration_tests_passed: true
    - code_reviewed: true

  quality:
    - security_scan_passed: true
    - performance_criteria_met: true
    - accessibility_checked: true
    - documentation_updated: true

  deployment:
    - deployed_to_staging: true
    - acceptance_testing_passed: true
    - product_owner_approval: true
```

---

## 🔄 Release Management

### Release Strategy:

#### Release Types:
```yaml
release_types:
  hotfix:
    frequency: "as_needed"
    scope: "critical_bug_fixes"
    approval: "scrum_master + po"

  minor_release:
    frequency: "every_sprint"
    scope: "new_features + bug_fixes"
    approval: "product_owner"

  major_release:
    frequency: "every_3-4_sprints"
    scope: "significant_features + architecture_changes"
    approval: "stakeholders + product_owner"
```

#### Release Process:
1. **Code Freeze** (за 2 дня до релиза)
2. **Final Testing** в staging environment
3. **Release Notes** generation
4. **Deployment** в production
5. **Post-release Monitoring** (24 часа)

### Version Management:
```yaml
versioning_scheme: "semantic_versioning"
format: "MAJOR.MINOR.PATCH"
examples:
  - "1.0.0" # Первый production release
  - "1.1.0" # Новая функциональность
  - "1.1.1" # Bug fix
```

---

## 📊 Reporting и Communication

### Sprint Reports:

#### Для Stakeholders:
```yaml
weekly_report:
  audience: "management + customers"
  format: "executive_summary"
  content:
    - sprint_progress: "completed vs planned"
    - key_achievements: "major_milestones"
    - risks_and_issues: "blockers_and_mitigation"
    - next_sprint_goals: "upcoming_deliverables"
```

#### Для команды:
```yaml
internal_metrics:
  daily: "burndown_chart_update"
  weekly: "velocity_tracking"
  sprint_end: "retrospective_notes + action_items"
  monthly: "team_satisfaction_survey"
```

### Communication Plan:
```yaml
communication_schedule:
  stakeholders:
    frequency: "weekly"
    method: "email_report + optional_demo"
    duration: "30_minutes"

  management:
    frequency: "bi_weekly"
    method: "progress_presentation"
    duration: "1_hour"

  end_users:
    frequency: "each_release"
    method: "release_notes + training_session"
    duration: "varies"
```

---

## 🛠️ Technical Practices

### Engineering Practices:

#### Code Quality:
```yaml
code_standards:
  code_review:
    required_reviewers: 2
    automated_checks: ["linting", "testing", "security"]
    review_checklist: true

  testing_pyramid:
    unit_tests: "70% of all tests"
    integration_tests: "20% of all tests"
    e2e_tests: "10% of all tests"

  continuous_integration:
    build_on_commit: true
    automated_testing: true
    deployment_gates: true
```

#### Technical Debt Management:
```yaml
technical_debt:
  measurement: "sonarqube_debt_ratio"
  target: "<5%"
  dedicated_time: "20% of each sprint"

  debt_tracking:
    - regular_code_quality_reviews
    - refactoring_stories_in_backlog
    - architecture_evolution_planning
```

---

## 📅 Calendar и Schedule

### Recurring Meetings:
```yaml
scrum_calendar:
  sprint_planning:
    frequency: "every_2_weeks"
    duration: "4_hours"
    day: "monday_9am"

  daily_standup:
    frequency: "daily"
    duration: "15_minutes"
    time: "9:30am"
    format: "hybrid_optional_physical"

  sprint_review:
    frequency: "every_2_weeks"
    duration: "2_hours"
    day: "friday_2pm"

  retrospective:
    frequency: "every_2_weeks"
    duration: "1.5_hours"
    day: "friday_4pm"

  backlog_refinement:
    frequency: "weekly"
    duration: "1_hour"
    day: "wednesday_3pm"
```

### Non-Scrum Meetings:
```yaml
additional_meetings:
  architecture_review:
    frequency: "monthly"
    participants: ["senior_developers", "architects"]
    duration: "2_hours"

  stakeholder_sync:
    frequency: "monthly"
    participants: ["po", "key_stakeholders"]
    duration: "1_hour"

  team_building:
    frequency: "quarterly"
    format: "informal"
    duration: "2-4_hours"
```

---

## 🎯 Success Criteria

### Sprint Success Metrics:
```yaml
success_kpis:
  delivery:
    - sprint_goal_achievement: ">80%"
    - story_completion_rate: ">90%"
    - planned_vs_actual_velocity: "±15%"

  quality:
    - escaped_defects: "<2_per_sprint"
    - test_coverage: ">80%"
    - code_review_feedback: "positive"

  team:
    - team_satisfaction: ">4/5"
    - retrospective_action_completion: ">80%"
    - knowledge_sharing_sessions: ">1_per_sprint"
```

### Long-term Success:
- **Product quality** improvement over time
- **Team velocity** stabilization and growth
- **Stakeholder satisfaction** with delivery
- **Technical debt** reduction
- **Team skill development** and retention