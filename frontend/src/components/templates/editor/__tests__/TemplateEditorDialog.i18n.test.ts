/**
 * Guards the i18n contract TemplateEditorDialog depends on for the
 * v0.126.3 methodist-only toggle. The Jest setup mocks
 * `useTranslations` to return the key verbatim, so component-level
 * tests pass even when the namespace points at a path that doesn't
 * exist in the JSON. This test loads the actual locale files and
 * asserts that the new methodistOnly* keys resolve to non-empty
 * strings in all four locales (mirror to MFAVerifyLoginStep.i18n.test).
 */
import fs from 'fs'
import path from 'path'

const LOCALES = ['ru', 'en', 'fr', 'ar'] as const

const REQUIRED_KEYS = ['methodistOnlyLabel', 'methodistOnlyDescription'] as const

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

describe('TemplateEditorDialog methodist-only i18n contract', () => {
  it.each(LOCALES)(
    '%s.json: every templates.editor.methodistOnly* key resolves to a non-empty string',
    (locale) => {
      const data = loadLocale(locale)
      for (const k of REQUIRED_KEYS) {
        const value = resolvePath(data, `templates.editor.${k}`)
        expect(typeof value).toBe('string')
        expect((value as string).trim().length).toBeGreaterThan(0)
      }
    }
  )

  it('all 4 locales expose the same templates.editor key set', () => {
    const seen = LOCALES.map((loc) => {
      const data = loadLocale(loc)
      const editor = resolvePath(data, 'templates.editor') as Record<string, unknown>
      return Object.keys(editor).sort()
    })
    for (let i = 1; i < seen.length; i++) {
      expect(seen[i]).toEqual(seen[0])
    }
  })
})
