# Frontend Testing

Этот проект использует **Jest** и **React Testing Library** для юнит-тестирования React компонентов, и **Playwright** для E2E тестирования.

## Структура тестов

Согласно Next.js 15 best practices:

```
frontend/
├── tests/                      # E2E тесты (Playwright)
│   └── e2e/
│       └── home.spec.ts
├── src/
│   ├── components/
│   │   ├── __tests__/          # Unit тесты компонентов (colocated)
│   │   │   └── theme-toggle-button.test.tsx
│   │   └── ui/
│   │       ├── __tests__/      # Unit тесты UI компонентов (colocated)
│   │       │   └── button.test.tsx
│   │       └── button.tsx
│   └── test-utils/             # Утилиты для тестирования
│       └── index.tsx           # Кастомная render функция с провайдерами
└── docs/
    └── testing.md              # Эта документация
```

### Почему такая структура?

1. **E2E тесты в `tests/e2e/`** - отдельно от исходного кода, в корне проекта (рекомендация Next.js)
2. **Unit тесты colocated** - в `__tests__/` папках рядом с компонентами (Next.js best practice)
3. **Нет `src/__tests__/`** - не нужна, так как тесты находятся рядом с тестируемым кодом
4. **`test-utils/`** - переиспользуемые утилиты для тестов (custom render с провайдерами)

## Команды для запуска тестов

```bash
# Запустить все unit тесты (Jest)
npm test

# Запустить тесты в watch режиме
npm run test:watch

# Запустить тесты с покрытием кода
npm run test:coverage

# Запустить только юнит-тесты (алиас для npm test)
npm run test:unit

# Запустить E2E тесты (Playwright)
npm run test:e2e
```

## Конфигурация

### Jest Config (`jest.config.ts`)

- **testEnvironment**: `jsdom` - для тестирования React компонентов
- **setupFilesAfterEnv**: `jest.setup.ts` - файл с глобальными настройками
- **moduleNameMapper**: Алиасы для импорта (`@/` → `src/`)
- **testPathIgnorePatterns**: Игнорируются E2E тесты Playwright (`tests/e2e/`)
- **coverageThreshold**: Минимальное покрытие кода 70%

### Jest Setup (`jest.setup.ts`)

Глобальные моки для:
- `@testing-library/jest-dom` - дополнительные матчеры
- `next/navigation` - Next.js навигация
- `window.matchMedia` - media queries
- `IntersectionObserver` - lazy loading

### Test Utils (`src/test-utils/index.tsx`)

Кастомная функция `render` с провайдерами:
- `ThemeProvider` от `next-themes`

Использование:
```tsx
import { render, screen } from '@/test-utils'

test('example', () => {
  render(<MyComponent />)
  expect(screen.getByText('Hello')).toBeInTheDocument()
})
```

## Примеры тестов

### Простой тест компонента

```tsx
import { render, screen } from '@/test-utils'
import { MyComponent } from '../MyComponent'

describe('MyComponent', () => {
  it('renders correctly', () => {
    render(<MyComponent />)
    expect(screen.getByText('Hello')).toBeInTheDocument()
  })
})
```

### Тест с взаимодействием

```tsx
import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { Button } from '../Button'

describe('Button', () => {
  it('calls onClick when clicked', async () => {
    const user = userEvent.setup()
    const handleClick = jest.fn()

    render(<Button onClick={handleClick}>Click me</Button>)

    await user.click(screen.getByRole('button'))

    expect(handleClick).toHaveBeenCalledTimes(1)
  })
})
```

### Тест с async/await

```tsx
import { render, screen, waitFor } from '@/test-utils'
import { AsyncComponent } from '../AsyncComponent'

describe('AsyncComponent', () => {
  it('loads data', async () => {
    render(<AsyncComponent />)

    await waitFor(() => {
      expect(screen.getByText('Data loaded')).toBeInTheDocument()
    })
  })
})
```

## Best Practices

1. **Используйте test-utils** - всегда импортируйте из `@/test-utils`, а не напрямую из `@testing-library/react`
2. **Тестируйте поведение, а не реализацию** - проверяйте что компонент делает, а не как он это делает
3. **Используйте user-event** - для симуляции пользовательских действий вместо `fireEvent`
4. **Ждите асинхронные операции** - используйте `waitFor` для асинхронного кода
5. **Пишите понятные тесты** - название теста должно описывать что тестируется
6. **Избегайте дублирования** - используйте `beforeEach` для общих действий
7. **Тестируйте accessibility** - проверяйте `aria-label`, `role` и другие атрибуты
8. **Colocate тесты** - храните тесты рядом с компонентами в `__tests__/` папках

## Coverage

Цели покрытия кода (минимум 70%):
- Branches: 70%
- Functions: 70%
- Lines: 70%
- Statements: 70%

Для просмотра отчета о покрытии:
```bash
npm run test:coverage
```

Отчет будет сохранен в `coverage/lcov-report/index.html`

## E2E Тестирование (Playwright)

E2E тесты находятся в `tests/e2e/` и запускаются с помощью Playwright.

### Запуск E2E тестов

```bash
# Запустить все E2E тесты
npm run test:e2e

# Запустить в UI режиме
npx playwright test --ui

# Запустить в debug режиме
npx playwright test --debug
```

### Пример E2E теста

```typescript
import { test, expect } from '@playwright/test'

test('homepage has title', async ({ page }) => {
  await page.goto('/')

  await expect(page).toHaveTitle(/Информационная система/)
})
```

## Дополнительные ресурсы

- [Next.js Testing](https://nextjs.org/docs/app/building-your-application/testing)
- [Jest Documentation](https://jestjs.io/)
- [React Testing Library](https://testing-library.com/react)
- [Testing Library User Event](https://testing-library.com/docs/user-event/intro)
- [Playwright Documentation](https://playwright.dev/)
- [Testing Playground](https://testing-playground.com/) - use good testing practices to match elements.

---

**📅 Актуальность документа**  
**Последнее обновление**: 2025-01-15  
**Версия проекта**: 0.1.0  
**Статус**: Актуальный

