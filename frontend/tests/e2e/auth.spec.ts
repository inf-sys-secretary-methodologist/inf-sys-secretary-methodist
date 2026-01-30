import { test, expect } from '@playwright/test'

test.describe('Аутентификация', () => {
  test.beforeEach(async ({ page }) => {
    // Очистка cookies перед каждым тестом
    await page.context().clearCookies()
  })

  test('должна отображаться страница входа', async ({ page }) => {
    await page.goto('/login')

    // Проверяем заголовок
    await expect(page.getByRole('heading', { name: /добро пожаловать/i })).toBeVisible()

    // Проверяем наличие полей формы
    await expect(page.getByLabel(/электронная почта/i)).toBeVisible()
    await expect(page.getByLabel(/пароль/i)).toBeVisible()

    // Проверяем кнопку входа
    await expect(page.getByRole('button', { name: /войти/i })).toBeVisible()

    // Проверяем ссылку на регистрацию
    await expect(page.getByRole('link', { name: /зарегистрироваться/i })).toBeVisible()
  })

  test('должна показывать ошибку при пустых полях', async ({ page }) => {
    await page.goto('/login')

    // Кликаем на поле email и уходим (blur) для валидации
    const emailField = page.getByLabel(/электронная почта/i)
    await emailField.click()
    await emailField.blur()

    // Ожидаем появления ошибок валидации (mode: 'onBlur')
    await expect(page.getByText('Email обязателен')).toBeVisible({ timeout: 5000 })
  })

  test('форма входа принимает ввод данных', async ({ page }) => {
    await page.goto('/login')

    // Вводим данные в форму
    const emailField = page.getByLabel(/электронная почта/i)
    const passwordField = page.getByLabel(/пароль/i)
    const submitButton = page.getByRole('button', { name: /войти/i })

    await emailField.fill('test@example.com')
    await passwordField.fill('TestPassword123!')

    // Проверяем что данные введены
    await expect(emailField).toHaveValue('test@example.com')
    await expect(passwordField).toHaveValue('TestPassword123!')

    // Проверяем что кнопка отправки активна
    await expect(submitButton).toBeEnabled()
  })

  test('должна отображаться страница регистрации', async ({ page }) => {
    await page.goto('/register')

    // Проверяем заголовок
    await expect(page.getByRole('heading', { name: /регистрация|создать аккаунт/i })).toBeVisible()

    // Проверяем наличие полей формы (RegisterForm использует "Email" и "Пароль")
    await expect(page.getByLabel(/email/i)).toBeVisible()
    await expect(page.getByLabel(/пароль/i).first()).toBeVisible()

    // Проверяем ссылку на вход
    await expect(page.getByRole('link', { name: /войти/i })).toBeVisible()
  })

  test('навигация между login и register', async ({ page }) => {
    // Переходим на страницу входа
    await page.goto('/login')

    // Проверяем наличие ссылки на регистрацию
    const registerLink = page.getByRole('link', { name: /зарегистрироваться/i })
    await expect(registerLink).toBeVisible({ timeout: 5000 })

    // Переходим на страницу регистрации через прямую навигацию
    await page.goto('/register')

    // Проверяем наличие ссылки на вход
    const loginLink = page.getByRole('link', { name: /войти/i })
    await expect(loginLink).toBeVisible({ timeout: 5000 })
  })

  test('должна перенаправлять неавторизованного пользователя', async ({ page }) => {
    // Пытаемся перейти на защищённую страницу
    await page.goto('/dashboard')

    // Должны быть перенаправлены на login или forbidden
    await expect(page).toHaveURL(/.*login|.*forbidden/)
  })
})
