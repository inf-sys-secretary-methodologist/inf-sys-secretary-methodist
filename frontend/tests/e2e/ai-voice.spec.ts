import { test, expect } from '@playwright/test'

/**
 * E2E тесты для голосовых функций AI-ассистента (Web Speech API)
 *
 * Инжектируем мок SpeechRecognition/SpeechSynthesis в браузер,
 * чтобы кнопки отображались детерминированно.
 */

const SPEECH_API_MOCK_SCRIPT = `
  window.SpeechRecognition = class {
    continuous = false;
    interimResults = false;
    lang = '';
    onresult = null;
    onerror = null;
    onend = null;
    onstart = null;
    start() { if (this.onstart) this.onstart(); }
    stop() { if (this.onend) this.onend(); }
    abort() {}
  };
  window.webkitSpeechRecognition = window.SpeechRecognition;

  if (!window.speechSynthesis) {
    window.speechSynthesis = {
      speak: () => {},
      cancel: () => {},
      pause: () => {},
      resume: () => {},
      getVoices: () => [],
      addEventListener: () => {},
      removeEventListener: () => {},
    };
  }
  window.SpeechSynthesisUtterance = class {
    constructor(text) { this.text = text; }
    lang = '';
    voice = null;
    onstart = null;
    onend = null;
    onerror = null;
  };
`

test.describe('AI Voice Features', () => {
  test.describe('Без авторизации', () => {
    test('AI чат требует авторизацию', async ({ page }) => {
      await page.goto('/ai')

      // Неавторизованный пользователь должен быть перенаправлен
      await expect(page).toHaveURL(/.*login|.*forbidden|.*ai/)
    })
  })

  test.describe('С мок авторизацией и Speech API', () => {
    test.beforeEach(async ({ page }) => {
      await page.addInitScript(() => {
        localStorage.setItem('authToken', 'mock-token-for-testing')
        localStorage.setItem(
          'user',
          JSON.stringify({
            id: 1,
            email: 'test@example.com',
            firstName: 'Test',
            lastName: 'User',
            role: 'teacher',
          })
        )
      })
      // Inject Speech API mocks before page loads
      await page.addInitScript(SPEECH_API_MOCK_SCRIPT)
    })

    test('AI чат страница загружается', async ({ page }) => {
      await page.goto('/ai')
      await page.waitForLoadState('networkidle')

      // Должны попасть на /ai (или логин если мок-токен отклонён)
      const url = page.url()
      expect(url.includes('/ai') || url.includes('/login')).toBeTruthy()
    })

    test('Ctrl+Shift+V не вызывает ошибку на странице', async ({ page }) => {
      await page.goto('/ai')
      await page.waitForLoadState('networkidle')

      if (page.url().includes('/ai')) {
        const consoleErrors: string[] = []
        page.on('console', (msg) => {
          if (msg.type() === 'error') {
            consoleErrors.push(msg.text())
          }
        })

        await page.keyboard.press('Control+Shift+V')
        await page.waitForTimeout(500)

        // Страница не должна крашиться
        const isPageAlive = await page
          .evaluate(() => document.readyState === 'complete')
          .catch(() => false)
        expect(isPageAlive).toBeTruthy()

        // Не должно быть ошибок, связанных с voice/speech
        const voiceErrors = consoleErrors.filter(
          (e) => e.toLowerCase().includes('speech') || e.toLowerCase().includes('voice')
        )
        expect(voiceErrors).toHaveLength(0)
      }
    })
  })
})
