# 🔄 Sprint Management & Scrum Process Guide

## 📋 Обзор Scrum процесса

Agile-методология для команды из 5 разработчиков с адаптацией под специфику образовательного проекта и модульной архитектуры.

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
    focus: ["Go modular architecture", "API design", "database design"]
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

## 📋 Sprint Planning (Начало спринта)

### 1. Планирование спринта
**Где**: **Backlog** view
1. **Выберите задачи** для нового спринта (обычно 3-7 задач)
2. **Для каждой задачи**:
   - Откройте карточку
   - Sprint: `Backlog` → `Sprint 2`
   - Story Points: оцените сложность (1,2,3,5,8)
   - Priority: High/Medium/Low
   - Assignee: назначьте исполнителя

### 2. Проверка планирования
**Где**: **Sprint Planning** view
- Видите только задачи текущего спринта
- Все задачи в колонке **"Sprint Backlog"**

## 🔄 Ежедневная работа (Daily Scrum)

### 3. Ежедневный workflow
**Где**: **Sprint Planning** view

**Движение задач слева направо:**
```
Sprint Backlog → Ready → In Progress → Review → Done
```

### 4. Типичные действия команды:

**Разработчик берет задачу:**
- Перетаскивает из **Sprint Backlog** → **Ready** → **In Progress**
- ⚠️ **НЕ меняйте поле Sprint!** (остается Sprint 1)

**Задача готова к ревью:**
- Перетаскивает из **In Progress** → **Review**
- Добавляет комментарий с PR ссылкой

**После ревью:**
- **Review** → **Done** (если все ОК)
- **Review** → **In Progress** (если нужны правки)

## 📊 Отслеживание прогресса

### 5. Burndown и метрики
**Где**: **Burndown** view (Table)
- Видите все спринты группами
- Story Points по статусам
- Сколько задач Done vs In Progress

### 6. Daily Standup questions:
1. **"Что сделал вчера?"** - смотрите движение в Sprint Planning
2. **"Что буду делать сегодня?"** - задачи в In Progress
3. **"Какие блокеры?"** - задачи, застрявшие в Review

## 🎯 Sprint Review (Конец спринта)

### 7. Завершение спринта
**Где**: **Sprint Planning** view

**Для завершенных задач (Done):**
- Ничего не делаем - остаются как есть

**Для незавершенных задач:**
1. Откройте карточку задачи
2. Sprint: `Sprint 1` → `Backlog`
3. Status: любой → **"No Status"** (возвращаем в бэклог)

### 8. Подготовка следующего спринта
**Где**: **Backlog** view
- Выберите новые задачи
- Установите Sprint: `Sprint 2`
- Обновите фильтр в Sprint Planning: `sprint:"Sprint 2"`

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

## 🔧 Практические советы

### 9. Использование views:

**📋 Backlog** - планирование и общий обзор
**🎯 Sprint Planning** - ежедневная работа команды
**📊 Burndown** - отчетность и ретроспектива
**👥 Team Capacity** - распределение нагрузки

### 10. Лайфхаки:

**Комментарии в задачах:**
- Ссылки на PR: `Готов к ревью: #123`
- Блокеры: `Заблокировано: жду ответа от дизайнера`
- Обновления: `Прогресс: API готов, делаю тесты`

**Используйте @mentions:**
- `@username готово к ревью`
- `@team нужна помощь с этой задачей`

**Story Points для оценки:**
- 1 = очень просто (30 мин)
- 2 = просто (1-2 часа)
- 3 = средне (полдня)
- 5 = сложно (1-2 дня)
- 8 = очень сложно (3+ дня)

## ⚡ Быстрые действия

**Горячие клавиши:**
- `CMD/Ctrl + K` - быстрый поиск задач
- Drag & Drop между колонками
- Bulk edit для multiple задач
---

**📅 Актуальность документа**  
**Последнее обновление**: 2025-01-15  
**Версия проекта**: 0.1.0  
**Статус**: Актуальный

