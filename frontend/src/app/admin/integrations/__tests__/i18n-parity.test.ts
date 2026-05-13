// i18n parity — raw JSON load so a missing key in any locale
// fails the build. Mirror к feedback_i18n_json_load_parity_test.
import ru from '../../../../../messages/ru.json'
import en from '../../../../../messages/en.json'
import fr from '../../../../../messages/fr.json'
import ar from '../../../../../messages/ar.json'

type MessagesShape = {
  adminIntegrations?: {
    title?: string
    description?: string
    loadFailed?: string
    vapid?: {
      sectionLabel?: string
      configured?: string
      unconfigured?: string
      publicKey?: string
      subject?: string
    }
    n8n?: {
      sectionLabel?: string
      enabled?: string
      disabled?: string
      webhookUrl?: string
    }
  }
}

const locales: Array<readonly [string, MessagesShape]> = [
  ['ru', ru as MessagesShape],
  ['en', en as MessagesShape],
  ['fr', fr as MessagesShape],
  ['ar', ar as MessagesShape],
]

const vapidKeys = ['sectionLabel', 'configured', 'unconfigured', 'publicKey', 'subject'] as const
const n8nKeys = ['sectionLabel', 'enabled', 'disabled', 'webhookUrl'] as const

describe('adminIntegrations i18n parity × 4 locales', () => {
  it.each(locales)('%s has the top-level keys', (_name, msgs) => {
    expect(msgs.adminIntegrations).toBeDefined()
    expect(msgs.adminIntegrations?.title).toBeTruthy()
    expect(msgs.adminIntegrations?.description).toBeTruthy()
    expect(msgs.adminIntegrations?.loadFailed).toBeTruthy()
  })

  it.each(locales)('%s has all vapid sub-keys', (_name, msgs) => {
    const v = msgs.adminIntegrations?.vapid
    expect(v).toBeDefined()
    for (const k of vapidKeys) {
      expect(v?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all n8n sub-keys', (_name, msgs) => {
    const n = msgs.adminIntegrations?.n8n
    expect(n).toBeDefined()
    for (const k of n8nKeys) {
      expect(n?.[k]).toBeTruthy()
    }
  })
})
