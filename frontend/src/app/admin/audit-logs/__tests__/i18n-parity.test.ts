// i18n parity test — reads raw JSON message files (NOT through the
// useTranslations mock) so a missing key in any of the 4 locales
// fails the build. Mirror к feedback_i18n_json_load_parity_test:
// the global mock returns keys verbatim and hides namespace bugs.
import ru from '../../../../../messages/ru.json'
import en from '../../../../../messages/en.json'
import fr from '../../../../../messages/fr.json'
import ar from '../../../../../messages/ar.json'

type MessagesShape = {
  nav?: { auditLogs?: string }
  adminAuditLogs?: {
    title?: string
    description?: string
    loadFailed?: string
    empty?: { title?: string; description?: string }
    filters?: {
      action?: string
      resource?: string
      userId?: string
      from?: string
      to?: string
      reset?: string
    }
    columns?: {
      created_at?: string
      action?: string
      resource?: string
      actor?: string
      ip?: string
      correlation_id?: string
      fields?: string
    }
    pagination?: { prev?: string; next?: string; pageOf?: string }
    fields?: { showJson?: string; hide?: string }
  }
}

const locales: Array<readonly [string, MessagesShape]> = [
  ['ru', ru as MessagesShape],
  ['en', en as MessagesShape],
  ['fr', fr as MessagesShape],
  ['ar', ar as MessagesShape],
]

const filterKeys = ['action', 'resource', 'userId', 'from', 'to', 'reset'] as const
const columnKeys = [
  'created_at',
  'action',
  'resource',
  'actor',
  'ip',
  'correlation_id',
  'fields',
] as const
const paginationKeys = ['prev', 'next', 'pageOf'] as const

describe.each(locales)('messages/%s.json — audit logs parity', (name, msgs) => {
  it('has nav.auditLogs', () => {
    expect(msgs.nav?.auditLogs).toBeTruthy()
    expect(typeof msgs.nav?.auditLogs).toBe('string')
  })

  it('has adminAuditLogs.title', () => {
    expect(msgs.adminAuditLogs?.title).toBeTruthy()
    expect(msgs.adminAuditLogs?.title).not.toBe('title')
    expect(msgs.adminAuditLogs?.title).not.toBe('adminAuditLogs.title')
  })

  it('has adminAuditLogs.description', () => {
    expect(msgs.adminAuditLogs?.description).toBeTruthy()
  })

  it('has adminAuditLogs.loadFailed', () => {
    expect(msgs.adminAuditLogs?.loadFailed).toBeTruthy()
  })

  it('has adminAuditLogs.empty.{title, description}', () => {
    expect(msgs.adminAuditLogs?.empty?.title).toBeTruthy()
    expect(msgs.adminAuditLogs?.empty?.description).toBeTruthy()
  })

  it.each(filterKeys)('has adminAuditLogs.filters.%s', (key) => {
    const value = msgs.adminAuditLogs?.filters?.[key]
    expect(value).toBeTruthy()
    expect(typeof value).toBe('string')
  })

  it.each(columnKeys)('has adminAuditLogs.columns.%s', (key) => {
    const value = msgs.adminAuditLogs?.columns?.[key]
    expect(value).toBeTruthy()
    expect(typeof value).toBe('string')
  })

  it.each(paginationKeys)('has adminAuditLogs.pagination.%s', (key) => {
    const value = msgs.adminAuditLogs?.pagination?.[key]
    expect(value).toBeTruthy()
    expect(typeof value).toBe('string')
  })
})
