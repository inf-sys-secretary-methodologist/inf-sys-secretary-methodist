# Frontend Architecture

## 📐 Архитектурные принципы

### 1. Абстракция зависимостей

**Проблема**: Прямая зависимость от внешних библиотек делает код трудно тестируемым и сложным для миграции.

**Решение**: Используем слой абстракции через хуки и утилиты.

#### Пример: Theme Management

❌ **Плохо** - прямая зависимость:
```tsx
import { useTheme } from "next-themes"

function MyComponent() {
  const { theme, setTheme } = useTheme()
  // ...
}
```

✅ **Хорошо** - через абстракцию:
```tsx
import { useTheme } from "@/hooks/use-theme"

function MyComponent() {
  const { isDark, toggleTheme } = useTheme()
  // ...
}
```

**Преимущества**:
- Легко заменить библиотеку темизации
- Дополнительные утилиты (`toggleTheme`, `isDark`, `isLight`)
- Единая точка изменений
- Упрощенное тестирование

### 2. Компонентная архитектура

#### UI Components (`/components/ui`)

Базовые переиспользуемые компоненты:

- **Button** - базовый компонент кнопки с вариантами (shadcn)
- **Input** - базовый input с современным стилем
- **FloatingInput** - input с анимацией floating label
- **TubelightNavbar** - навигационная панель с анимацией

Все UI компоненты:
- Полностью типизированы (TypeScript)
- Покрыты тестами
- Имеют варианты стилей через props
- Поддерживают темную/светлую тему

#### Feature Components (`/components`)

Компоненты бизнес-логики:

- **ThemeToggleButton** - переключатель темы
  - Использует `@/hooks/use-theme` (абстракция)
  - Анимированный переключатель с иконками
  - Keyboard accessible (Enter/Space)

### 3. Конфигурация приложения (`/config`)

Вся конфигурация вынесена в отдельные файлы:

#### `/config/navigation.ts`
```typescript
export interface NavItem {
  name: string
  url: string
  icon: LucideIcon
}

export const navigationItems: NavItem[] = [...]
```

**Преимущества**:
- Легко добавить/удалить пункт меню
- Типизация навигации
- Единая точка изменений
- Переиспользование в разных местах

### 4. Hooks (`/hooks`)

Кастомные хуки для переиспользуемой логики:

#### `/hooks/use-theme.ts`

```typescript
export function useTheme() {
  const { theme, setTheme, resolvedTheme, systemTheme } = useNextTheme()

  const toggleTheme = () => {
    setTheme(theme === 'dark' ? 'light' : 'dark')
  }

  return {
    theme,
    setTheme,
    resolvedTheme,
    systemTheme,
    toggleTheme,
    isDark: resolvedTheme === 'dark',
    isLight: resolvedTheme === 'light',
  }
}
```

**Преимущества**:
- Абстракция над `next-themes`
- Дополнительные утилиты
- Упрощенное тестирование
- Возможность подмены реализации

### 5. Тестирование

Все компоненты покрыты тестами:

```
frontend/src/components/ui/__tests__/
├── button.test.tsx
├── input.test.tsx
├── floating-input.test.tsx
└── tubelight-navbar.test.tsx

frontend/src/components/__tests__/
└── theme-toggle-button.test.tsx
```

**Подход к тестированию**:
- Unit тесты для каждого компонента
- Тестирование пользовательских сценариев
- Тестирование доступности (a11y)
- Mock внешних зависимостей

#### Пример мока хука:

```typescript
jest.mock('@/hooks/use-theme', () => ({
  useTheme: jest.fn(),
}))

const mockUseTheme = require('@/hooks/use-theme').useTheme

mockUseTheme.mockReturnValue({
  resolvedTheme: 'light',
  toggleTheme: jest.fn(),
  isDark: false,
  isLight: true,
})
```

### 6. Структура директорий

```
frontend/src/
├── app/                      # Next.js App Router
│   ├── layout.tsx           # Root layout
│   ├── page.tsx             # Home page
│   ├── loading.tsx          # Loading UI
│   ├── error.tsx            # Error UI
│   └── not-found.tsx        # 404 page
├── components/
│   ├── ui/                  # Базовые UI компоненты
│   │   ├── button.tsx
│   │   ├── input.tsx
│   │   ├── floating-input.tsx
│   │   ├── tubelight-navbar.tsx
│   │   └── __tests__/       # Тесты UI компонентов
│   ├── providers/           # React Context провайдеры
│   │   └── theme-provider.tsx
│   ├── theme-toggle-button.tsx
│   └── __tests__/           # Тесты компонентов
├── config/                  # Конфигурация приложения
│   └── navigation.ts        # Конфигурация навигации
├── hooks/                   # Кастомные React хуки
│   └── use-theme.ts         # Абстракция темизации
├── lib/                     # Утилиты и API клиенты
│   ├── utils.ts             # cn() и другие утилиты
│   └── api.ts               # API клиент
└── types/                   # TypeScript типы
    └── api.ts               # API типы
```

## 🎯 Принципы разработки

### 1. DRY (Don't Repeat Yourself)
- Переиспользуемые компоненты
- Абстракция общей логики в хуки
- Конфигурация в отдельных файлах

### 2. Single Responsibility
- Каждый компонент решает одну задачу
- Хуки инкапсулируют конкретную логику
- Утилиты делают одну вещь хорошо

### 3. Dependency Injection
- Абстракция внешних зависимостей
- Возможность подмены реализации
- Упрощенное тестирование

### 4. Composition over Inheritance
- Компоненты комбинируются, а не наследуются
- Props для кастомизации поведения
- Render props и children для гибкости

## 🔄 Будущие улучшения

### Планируется добавить:

1. **State Management**
   - Zustand для глобального состояния
   - React Query для server state

2. **Form Management**
   - React Hook Form
   - Zod для валидации

3. **Routing**
   - Next.js App Router (уже используется)
   - Protected routes

4. **API Layer**
   - API client с axios
   - Request/Response interceptors
   - Error handling

5. **Performance**
   - Code splitting
   - Lazy loading
   - Image optimization

## 📚 Связанные документы

- [Общая архитектура проекта](../docs/architecture/modular-architecture.md)
- [DDD Domain Modeling](../docs/architecture/ddd-domain-modeling.md)
- [TDD Guide](../docs/development/tdd-guide.md)
- [Tech Stack Rationale](../docs/architecture/tech-stack-rationale.md)

## 🎨 UI/UX Принципы

### Темизация
- Светлая и темная тема
- Адаптация всех компонентов
- Плавные переходы

### Доступность (a11y)
- ARIA атрибуты
- Keyboard navigation
- Focus management
- Screen reader support

### Адаптивность
- Mobile-first подход
- Responsive дизайн
- Touch-friendly интерфейсы

---

**Дата обновления**: 2025-11-09
**Версия**: 0.1.0
**Статус**: Актуальный
