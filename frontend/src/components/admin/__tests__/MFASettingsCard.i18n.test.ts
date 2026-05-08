/**
 * Guards the i18n contract MFASettingsCard depends on. The Jest setup
 * mocks `useTranslations` to return the key verbatim, which means the
 * component-level tests pass even when the namespace points at a path
 * that doesn't exist in the JSON. This test loads the actual locale
 * files and asserts every key the card consumes resolves to a string.
 */
import fs from 'fs'
import path from 'path'

const LOCALES = ['ru', 'en', 'fr', 'ar'] as const

// Keys MFASettingsCard reads via t('mfa.X') under namespace
// 'adminSettings.security'. Keep this list in sync with the component.
const REQUIRED_KEYS = [
  'mfa.title',
  'mfa.descriptionDisabled',
  'mfa.descriptionEnabled',
  'mfa.enable',
  'mfa.disable',
  'mfa.scanInstruction',
  'mfa.codeLabel',
  'mfa.confirm',
  'mfa.confirmDisable',
  'mfa.disableInstruction',
  'mfa.successEnable',
  'mfa.successDisable',
  'mfa.errorBegin',
  'mfa.errorConfirm',
  'mfa.errorDisable',
  'mfa.secretLabel',
  'mfa.otpauthLabel',
] as const

// Page-level keys consumed under 'adminSettings.security' namespace.
const PAGE_KEYS = ['title', 'subtitle'] as const

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

describe('MFASettingsCard i18n contract', () => {
  it.each(LOCALES)(
    '%s.json: every adminSettings.security.{mfa.*, title, subtitle} key resolves to a non-empty string',
    (locale) => {
      const data = loadLocale(locale)
      const root = resolvePath(data, 'adminSettings.security')
      expect(root).toBeDefined()
      expect(typeof root).toBe('object')

      for (const k of [...PAGE_KEYS, ...REQUIRED_KEYS]) {
        const value = resolvePath(data, `adminSettings.security.${k}`)
        expect(typeof value).toBe('string')
        expect((value as string).trim().length).toBeGreaterThan(0)
      }
    }
  )

  it('all 4 locales expose the same set of keys under adminSettings.security', () => {
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
      walk([], resolvePath(data, 'adminSettings.security'))
      return flat.sort()
    })
    // ru is the source of truth; the other three must match exactly.
    for (let i = 1; i < seen.length; i++) {
      expect(seen[i]).toEqual(seen[0])
    }
  })
})
