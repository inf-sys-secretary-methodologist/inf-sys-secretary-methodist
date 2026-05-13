// i18n parity test — reads raw JSON message files (NOT through the
// useTranslations mock) so a missing key in any of the 4 locales
// fails the build. Mirror к feedback_i18n_json_load_parity_test:
// the global mock returns keys verbatim and hides namespace bugs.
import ru from '../../../../../messages/ru.json'
import en from '../../../../../messages/en.json'
import fr from '../../../../../messages/fr.json'
import ar from '../../../../../messages/ar.json'

type MessagesShape = {
  nav?: { backups?: string }
  adminBackups?: {
    title?: string
    description?: string
    loadFailed?: string
    empty?: { title?: string; description?: string }
    metrics?: {
      sectionLabel?: string
      postgres?: string
      minio?: string
      lastRun?: string
      lastSuccess?: string
      age?: string
      duration?: string
      sizeBytes?: string
      totalCount?: string
      successCount?: string
      failureCount?: string
      never?: string
      ok?: string
      failed?: string
    }
    remoteSync?: { title?: string; ok?: string; failed?: string }
    columns?: {
      name?: string
      type?: string
      size?: string
      modifiedAt?: string
      encryption?: string
      actions?: string
    }
    actions?: { download?: string }
    encryption?: { none?: string; age?: string; gpg?: string; tooltip?: string }
    types?: { postgres?: string; minio?: string }
  }
}

const locales: Array<readonly [string, MessagesShape]> = [
  ['ru', ru as MessagesShape],
  ['en', en as MessagesShape],
  ['fr', fr as MessagesShape],
  ['ar', ar as MessagesShape],
]

const metricsKeys = [
  'sectionLabel',
  'postgres',
  'minio',
  'lastRun',
  'lastSuccess',
  'age',
  'duration',
  'sizeBytes',
  'totalCount',
  'successCount',
  'failureCount',
  'never',
  'ok',
  'failed',
] as const

const columnKeys = ['name', 'type', 'size', 'modifiedAt', 'encryption', 'actions'] as const
const encryptionKeys = ['none', 'age', 'gpg', 'tooltip'] as const
const typeKeys = ['postgres', 'minio'] as const
const remoteSyncKeys = ['title', 'ok', 'failed'] as const

describe.each(locales)('messages/%s.json — backups parity', (_name, msgs) => {
  it('has nav.backups', () => {
    expect(msgs.nav?.backups).toBeTruthy()
    expect(typeof msgs.nav?.backups).toBe('string')
  })

  it('has adminBackups.title (non-key)', () => {
    expect(msgs.adminBackups?.title).toBeTruthy()
    expect(msgs.adminBackups?.title).not.toBe('title')
    expect(msgs.adminBackups?.title).not.toBe('adminBackups.title')
  })

  it('has adminBackups.description', () => {
    expect(msgs.adminBackups?.description).toBeTruthy()
  })

  it('has adminBackups.loadFailed', () => {
    expect(msgs.adminBackups?.loadFailed).toBeTruthy()
  })

  it('has adminBackups.empty.{title, description}', () => {
    expect(msgs.adminBackups?.empty?.title).toBeTruthy()
    expect(msgs.adminBackups?.empty?.description).toBeTruthy()
  })

  it.each(metricsKeys)('has adminBackups.metrics.%s', (key) => {
    expect(msgs.adminBackups?.metrics?.[key]).toBeTruthy()
  })

  it.each(remoteSyncKeys)('has adminBackups.remoteSync.%s', (key) => {
    expect(msgs.adminBackups?.remoteSync?.[key]).toBeTruthy()
  })

  it.each(columnKeys)('has adminBackups.columns.%s', (key) => {
    expect(msgs.adminBackups?.columns?.[key]).toBeTruthy()
  })

  it('has adminBackups.actions.download', () => {
    expect(msgs.adminBackups?.actions?.download).toBeTruthy()
  })

  it.each(encryptionKeys)('has adminBackups.encryption.%s', (key) => {
    expect(msgs.adminBackups?.encryption?.[key]).toBeTruthy()
  })

  it.each(typeKeys)('has adminBackups.types.%s', (key) => {
    expect(msgs.adminBackups?.types?.[key]).toBeTruthy()
  })
})
