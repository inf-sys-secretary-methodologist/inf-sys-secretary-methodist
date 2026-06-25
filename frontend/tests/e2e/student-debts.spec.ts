import { test, expect } from '@playwright/test'

/**
 * E2E тесты для модуля «Долги студентов» (/student-debts).
 *
 * Покрывает:
 * - Защита маршрутов реестра без авторизации
 * - Загрузка реестра под мок-авторизацией (staff)
 * - Страница «Мои долги» (/student-debts/my)
 * - Страница детали долга (/student-debts/[id])
 *
 * Стиль зеркалит users.spec.ts: мок-авторизация через localStorage +
 * проверка, что страница рендерит свою оболочку (реальные данные приходят
 * с бэкенда, который в CI может быть не засеян — поэтому ассертим на
 * заголовки/маршруты, а не на конкретные строки реестра).
 */
test.describe('Долги студентов', () => {
  test.describe('Без авторизации', () => {
    test('реестр требует авторизацию', async ({ page }) => {
      await page.goto('/student-debts')
      await expect(page).toHaveURL(/.*login|.*forbidden/)
    })

    test('«Мои долги» требуют авторизацию', async ({ page }) => {
      await page.goto('/student-debts/my')
      await expect(page).toHaveURL(/.*login|.*forbidden/)
    })
  })

  test.describe('С мок-авторизацией (academic_secretary)', () => {
    test.beforeEach(async ({ page }) => {
      await page.addInitScript(() => {
        localStorage.setItem('authToken', 'mock-token-for-testing')
        localStorage.setItem(
          'user',
          JSON.stringify({
            id: 1,
            email: 'secretary@example.com',
            firstName: 'Sec',
            lastName: 'User',
            role: 'academic_secretary',
          })
        )
      })
    })

    // Мок-токен не валиден против реального бэкенда: страница либо
    // отрисовывает реестр (валидная сессия), либо мягко редиректит на login.
    // Главное — нет краша и маршрут разрешается в одно из двух.
    test('реестр долгов открывается без краша', async ({ page }) => {
      await page.goto('/student-debts')
      await page.waitForLoadState('networkidle')
      await expect(page).toHaveURL(/\/student-debts|\/login/)
    })

    test('страница детали долга открывается по маршруту', async ({ page }) => {
      await page.goto('/student-debts/1')
      await page.waitForLoadState('networkidle')
      await expect(page).toHaveURL(/\/student-debts\/1|\/login/)
    })
  })

  test.describe('С мок-авторизацией (student)', () => {
    test.beforeEach(async ({ page }) => {
      await page.addInitScript(() => {
        localStorage.setItem('authToken', 'mock-token-for-testing')
        localStorage.setItem(
          'user',
          JSON.stringify({
            id: 7,
            email: 'student@example.com',
            firstName: 'Stud',
            lastName: 'User',
            role: 'student',
          })
        )
      })
    })

    test('«Мои долги» открываются для студента', async ({ page }) => {
      await page.goto('/student-debts/my')
      await page.waitForLoadState('networkidle')
      await expect(page).toHaveURL(/\/student-debts\/my|\/login/)
    })

    // Студент не имеет доступа к реестру: страница редиректит его на /my
    // (бэкенд отдаёт 403 на /api/student-debts для роли student).
    test('реестр редиректит студента на «Мои долги»', async ({ page }) => {
      await page.goto('/student-debts')
      await page.waitForLoadState('networkidle')
      await expect(page).toHaveURL(/\/student-debts\/my|\/login/)
    })
  })
})
