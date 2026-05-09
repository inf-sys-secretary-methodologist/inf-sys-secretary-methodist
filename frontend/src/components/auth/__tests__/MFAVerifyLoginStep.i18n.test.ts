/**
 * Guards the i18n contract MFAVerifyLoginStep + LoginForm depend on.
 * The Jest setup mocks `useTranslations` to return the key verbatim,
 * so component-level tests pass even when the namespace points at a
 * path that doesn't exist in the JSON. This test loads the actual
 * locale files and asserts every key the verify step consumes
 * resolves to a non-empty string and that all four locales expose
 * the same set of keys (mirror to MFASettingsCard.i18n.test.ts).
 */
import fs from 'fs'
import path from 'path'

const LOCALES = ['ru', 'en', 'fr', 'ar'] as const

// Keys MFAVerifyLoginStep reads via t('mfaPrompt.X') under namespace
// 'loginForm'. Keep this list in sync with the component.
const REQUIRED_KEYS = [
  'mfaPrompt.title',
  'mfaPrompt.subtitle',
  'mfaPrompt.codeLabel',
  'mfaPrompt.submit',
  'mfaPrompt.errorInvalidCode',
  'mfaPrompt.errorExpired',
  'mfaPrompt.errorIntermediateInvalid',
  'mfaPrompt.resendNote',
  'mfaPrompt.loginAgain',
] as const

function loadLocale(locale: string): Record<string, unknown> {
  const file = path.join(process.cwd(), 'messages', `${locale}.json`)
  return JSON.parse(fs.readFileSync(file, 'utf-8'))
}

function resolvePath(obj: Record<string, unknown>, dottedPath: string): unknown {
  return dottedPath.split('.').reduce<unknown>((acc, segment) => {
    if (acc && typeof acc === 'object' && segment in (acc as object)) {
      return (acc as Record<string, unknown>)[segment]
    }
    return undefined
  }, obj)
}

describe('MFAVerifyLoginStep i18n contract', () => {
  it.each(LOCALES)(
    '%s.json: every loginForm.mfaPrompt.* key resolves to a non-empty string',
    (locale) => {
      const data = loadLocale(locale)
      const root = resolvePath(data, 'loginForm.mfaPrompt')
      expect(root).toBeDefined()
      expect(typeof root).toBe('object')

      for (const k of REQUIRED_KEYS) {
        const value = resolvePath(data, `loginForm.${k}`)
        expect(typeof value).toBe('string')
        expect((value as string).trim().length).toBeGreaterThan(0)
      }
    }
  )

  it('all 4 locales expose the same set of keys under loginForm.mfaPrompt', () => {
    const seen = LOCALES.map((loc) => {
      const data = loadLocale(loc)
      const flat: string[] = []
      const walk = (prefix: string[], obj: unknown) => {
        if (obj && typeof obj === 'object') {
          for (const [k, v] of Object.entries(obj as Record<string, unknown>)) {
            walk([...prefix, k], v)
          }
        } else {
          flat.push(prefix.join('.'))
        }
      }
      walk([], resolvePath(data, 'loginForm.mfaPrompt'))
      return flat.sort()
    })
    // ru is the source of truth; the other three must match exactly.
    for (let i = 1; i < seen.length; i++) {
      expect(seen[i]).toEqual(seen[0])
    }
  })
})
