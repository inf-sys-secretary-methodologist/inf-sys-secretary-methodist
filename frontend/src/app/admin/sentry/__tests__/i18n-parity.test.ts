// i18n parity test — reads raw JSON message files (NOT through the
// useTranslations mock) so a missing key in any of the 4 locales
// fails the build. Mirror к feedback_i18n_json_load_parity_test:
// the global mock returns keys verbatim and hides namespace bugs.
import ru from '../../../../../messages/ru.json'
import en from '../../../../../messages/en.json'
import fr from '../../../../../messages/fr.json'
import ar from '../../../../../messages/ar.json'

type MessagesShape = {
  adminSentry?: {
    title?: string
    description?: string
    loadFailed?: string
    status?: {
      sectionLabel?: string
      configured?: string
      unconfigured?: string
    }
    fields?: {
      environment?: string
      release?: string
      tracesSampleRate?: string
      tracingEnabled?: string
      enabled?: string
      disabled?: string
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
const fieldsKeys = [
  'environment',
  'release',
  'tracesSampleRate',
  'tracingEnabled',
  'enabled',
  'disabled',
] as const

describe('adminSentry i18n parity × 4 locales', () => {
  it.each(locales)('%s has the top-level keys', (_name, msgs) => {
    expect(msgs.adminSentry).toBeDefined()
    expect(msgs.adminSentry?.title).toBeTruthy()
    expect(msgs.adminSentry?.description).toBeTruthy()
    expect(msgs.adminSentry?.loadFailed).toBeTruthy()
  })

  it.each(locales)('%s has all status sub-keys', (_name, msgs) => {
    const status = msgs.adminSentry?.status
    expect(status).toBeDefined()
    for (const k of statusKeys) {
      expect(status?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all fields sub-keys', (_name, msgs) => {
    const fields = msgs.adminSentry?.fields
    expect(fields).toBeDefined()
    for (const k of fieldsKeys) {
      expect(fields?.[k]).toBeTruthy()
    }
  })
})
