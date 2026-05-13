// i18n parity — raw JSON load so a missing key in any locale
// fails the build. Mirror к feedback_i18n_json_load_parity_test.
import ru from '../../../../../messages/ru.json'
import en from '../../../../../messages/en.json'
import fr from '../../../../../messages/fr.json'
import ar from '../../../../../messages/ar.json'

type MessagesShape = {
  adminComposio?: {
    title?: string
    description?: string
    loadFailed?: string
    status?: {
      sectionLabel?: string
      configured?: string
      unconfigured?: string
    }
    fields?: {
      apiKey?: string
      entityID?: string
      mcpConfigID?: string
      set?: string
      unset?: string
    }
  }
}

const locales: Array<readonly [string, MessagesShape]> = [
  ['ru', ru as MessagesShape],
  ['en', en as MessagesShape],
  ['fr', fr as MessagesShape],
  ['ar', ar as MessagesShape],
]

const statusKeys = ['sectionLabel', 'configured', 'unconfigured'] as const
const fieldKeys = ['apiKey', 'entityID', 'mcpConfigID', 'set', 'unset'] as const

describe('adminComposio i18n parity × 4 locales', () => {
  it.each(locales)('%s has the top-level keys', (_name, msgs) => {
    expect(msgs.adminComposio).toBeDefined()
    expect(msgs.adminComposio?.title).toBeTruthy()
    expect(msgs.adminComposio?.description).toBeTruthy()
    expect(msgs.adminComposio?.loadFailed).toBeTruthy()
  })

  it.each(locales)('%s has all status sub-keys', (_name, msgs) => {
    const s = msgs.adminComposio?.status
    expect(s).toBeDefined()
    for (const k of statusKeys) {
      expect(s?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all fields sub-keys', (_name, msgs) => {
    const f = msgs.adminComposio?.fields
    expect(f).toBeDefined()
    for (const k of fieldKeys) {
      expect(f?.[k]).toBeTruthy()
    }
  })
})
