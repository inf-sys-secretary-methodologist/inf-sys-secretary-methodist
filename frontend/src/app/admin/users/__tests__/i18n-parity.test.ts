// i18n parity test — reads raw JSON message files (NOT through the
// useTranslations mock) so a missing key in any of the 4 locales
// fails the build. Mirror к feedback_i18n_json_load_parity_test.
import ru from '../../../../../messages/ru.json'
import en from '../../../../../messages/en.json'
import fr from '../../../../../messages/fr.json'
import ar from '../../../../../messages/ar.json'

type DialogShape = {
  title?: string
  description?: string
  confirm?: string
  cancel?: string
  current?: string
  new?: string
}

type MessagesShape = {
  adminUsers?: {
    title?: string
    description?: string
    loadFailed?: string
    empty?: { title?: string; description?: string }
    filters?: {
      search?: string
      role?: string
      status?: string
      allRoles?: string
      allStatuses?: string
      reset?: string
    }
    columns?: {
      user?: string
      role?: string
      status?: string
      department?: string
      position?: string
      actions?: string
    }
    actions?: {
      changeRole?: string
      changeStatus?: string
      delete?: string
    }
    roleOptions?: {
      system_admin?: string
      methodist?: string
      academic_secretary?: string
      teacher?: string
      student?: string
    }
    statusOptions?: {
      active?: string
      inactive?: string
      blocked?: string
    }
    pagination?: {
      prev?: string
      next?: string
      pageOf?: string
    }
    dialogs?: {
      changeRole?: DialogShape
      changeStatus?: DialogShape
      delete?: DialogShape
    }
  }
}

const locales: Array<readonly [string, MessagesShape]> = [
  ['ru', ru as MessagesShape],
  ['en', en as MessagesShape],
  ['fr', fr as MessagesShape],
  ['ar', ar as MessagesShape],
]

const filtersKeys = ['search', 'role', 'status', 'allRoles', 'allStatuses', 'reset'] as const
const columnsKeys = ['user', 'role', 'status', 'department', 'position', 'actions'] as const
const actionsKeys = ['changeRole', 'changeStatus', 'delete'] as const
const roleOptionsKeys = [
  'system_admin',
  'methodist',
  'academic_secretary',
  'teacher',
  'student',
] as const
const statusOptionsKeys = ['active', 'inactive', 'blocked'] as const
const paginationKeys = ['prev', 'next', 'pageOf'] as const

describe('adminUsers i18n parity × 4 locales', () => {
  it.each(locales)('%s has the top-level keys', (_name, msgs) => {
    expect(msgs.adminUsers).toBeDefined()
    expect(msgs.adminUsers?.title).toBeTruthy()
    expect(msgs.adminUsers?.description).toBeTruthy()
    expect(msgs.adminUsers?.loadFailed).toBeTruthy()
    expect(msgs.adminUsers?.empty?.title).toBeTruthy()
    expect(msgs.adminUsers?.empty?.description).toBeTruthy()
  })

  it.each(locales)('%s has all filters sub-keys', (_name, msgs) => {
    const filters = msgs.adminUsers?.filters
    expect(filters).toBeDefined()
    for (const k of filtersKeys) {
      expect(filters?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all columns sub-keys', (_name, msgs) => {
    const columns = msgs.adminUsers?.columns
    expect(columns).toBeDefined()
    for (const k of columnsKeys) {
      expect(columns?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all actions sub-keys', (_name, msgs) => {
    const actions = msgs.adminUsers?.actions
    expect(actions).toBeDefined()
    for (const k of actionsKeys) {
      expect(actions?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all roleOptions sub-keys', (_name, msgs) => {
    const opts = msgs.adminUsers?.roleOptions
    expect(opts).toBeDefined()
    for (const k of roleOptionsKeys) {
      expect(opts?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all statusOptions sub-keys', (_name, msgs) => {
    const opts = msgs.adminUsers?.statusOptions
    expect(opts).toBeDefined()
    for (const k of statusOptionsKeys) {
      expect(opts?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all pagination sub-keys', (_name, msgs) => {
    const pagination = msgs.adminUsers?.pagination
    expect(pagination).toBeDefined()
    for (const k of paginationKeys) {
      expect(pagination?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all dialog keys (changeRole/changeStatus/delete)', (_name, msgs) => {
    const dialogs = msgs.adminUsers?.dialogs
    expect(dialogs).toBeDefined()
    expect(dialogs?.changeRole?.title).toBeTruthy()
    expect(dialogs?.changeRole?.confirm).toBeTruthy()
    expect(dialogs?.changeStatus?.title).toBeTruthy()
    expect(dialogs?.changeStatus?.confirm).toBeTruthy()
    expect(dialogs?.delete?.title).toBeTruthy()
    expect(dialogs?.delete?.confirm).toBeTruthy()
  })
})
