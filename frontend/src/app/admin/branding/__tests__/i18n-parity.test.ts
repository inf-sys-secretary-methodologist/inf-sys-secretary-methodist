// i18n parity — raw JSON load so a missing key in any locale
// fails the build. Mirror к feedback_i18n_json_load_parity_test.
import ru from '../../../../../messages/ru.json'
import en from '../../../../../messages/en.json'
import fr from '../../../../../messages/fr.json'
import ar from '../../../../../messages/ar.json'

type MessagesShape = {
  adminBranding?: {
    title?: string
    description?: string
    loadFailed?: string
    savedSuccess?: string
    saveFailed?: string
    save?: string
    saving?: string
    fields?: {
      appName?: string
      tagline?: string
      logoURL?: string
      faviconURL?: string
      primaryColor?: string
      secondaryColor?: string
    }
    errors?: {
      INVALID_APP_NAME?: string
      INVALID_TAGLINE?: string
      INVALID_COLOR?: string
      INVALID_URL?: string
    }
    placeholders?: {
      logoURL?: string
      faviconURL?: string
      primaryColor?: string
      secondaryColor?: string
    }
  }
  nav?: {
    branding?: string
  }
}

const locales: Array<readonly [string, MessagesShape]> = [
  ['ru', ru as MessagesShape],
  ['en', en as MessagesShape],
  ['fr', fr as MessagesShape],
  ['ar', ar as MessagesShape],
]

const fieldKeys = [
  'appName',
  'tagline',
  'logoURL',
  'faviconURL',
  'primaryColor',
  'secondaryColor',
] as const

const errorKeys = ['INVALID_APP_NAME', 'INVALID_TAGLINE', 'INVALID_COLOR', 'INVALID_URL'] as const

const placeholderKeys = ['logoURL', 'faviconURL', 'primaryColor', 'secondaryColor'] as const

describe('adminBranding i18n parity × 4 locales', () => {
  it.each(locales)('%s has the top-level keys', (_name, msgs) => {
    expect(msgs.adminBranding).toBeDefined()
    expect(msgs.adminBranding?.title).toBeTruthy()
    expect(msgs.adminBranding?.description).toBeTruthy()
    expect(msgs.adminBranding?.loadFailed).toBeTruthy()
    expect(msgs.adminBranding?.savedSuccess).toBeTruthy()
    expect(msgs.adminBranding?.save).toBeTruthy()
    expect(msgs.adminBranding?.saving).toBeTruthy()
  })

  it.each(locales)('%s has all field labels', (_name, msgs) => {
    const f = msgs.adminBranding?.fields
    expect(f).toBeDefined()
    for (const k of fieldKeys) {
      expect(f?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all error messages', (_name, msgs) => {
    const e = msgs.adminBranding?.errors
    expect(e).toBeDefined()
    for (const k of errorKeys) {
      expect(e?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all placeholders', (_name, msgs) => {
    const p = msgs.adminBranding?.placeholders
    expect(p).toBeDefined()
    for (const k of placeholderKeys) {
      expect(p?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has nav.branding', (_name, msgs) => {
    expect(msgs.nav?.branding).toBeTruthy()
  })
})
