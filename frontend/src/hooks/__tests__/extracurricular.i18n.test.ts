// i18n parity × 4 locales for extracurricular namespace.
// Raw JSON load — missing key in any locale fails the build (per
// feedback_i18n_json_load_parity_test). Mirrors adminBranding pattern.

import ru from '../../../messages/ru.json'
import en from '../../../messages/en.json'
import fr from '../../../messages/fr.json'
import ar from '../../../messages/ar.json'

type EnumGroup = Record<string, string | undefined>

type MessagesShape = {
  extracurricular?: {
    title?: string
    description?: string
    loading?: string
    loadFailed?: string
    empty?: string
    create?: string
    edit?: string
    delete?: string
    register?: string
    unregister?: string
    registered?: string
    capacity?: string
    organizer?: string
    participants?: string
    location?: string
    startAt?: string
    endAt?: string
    backToList?: string
    calendar?: string
    upcoming?: string
    confirmDelete?: string
    confirmUnregister?: string
    status?: EnumGroup
    category?: EnumGroup
    audience?: EnumGroup
    form?: EnumGroup
    actions?: EnumGroup
    errors?: EnumGroup
  }
  nav?: {
    extracurricular?: string
  }
}

const locales: Array<readonly [string, MessagesShape]> = [
  ['ru', ru as MessagesShape],
  ['en', en as MessagesShape],
  ['fr', fr as MessagesShape],
  ['ar', ar as MessagesShape],
]

// Status / category / audience enum labels mirror domain VOs:
// internal/modules/extracurricular/domain/entities/value_objects.go
const statusKeys = ['label', 'all', 'draft', 'published', 'canceled', 'completed'] as const
const categoryKeys = [
  'label',
  'all',
  'academic',
  'cultural',
  'sports',
  'volunteer',
  'professional',
] as const
const audienceKeys = ['label', 'all', 'students', 'teachers', 'staff'] as const

const formKeys = [
  'createTitle',
  'editTitle',
  'titleLabel',
  'titlePlaceholder',
  'descriptionLabel',
  'descriptionPlaceholder',
  'categoryLabel',
  'audienceLabel',
  'locationLabel',
  'locationPlaceholder',
  'startAtLabel',
  'endAtLabel',
  'maxCapacityLabel',
  'maxCapacityPlaceholder',
  'save',
  'cancel',
] as const

const actionKeys = ['publish', 'cancel', 'complete'] as const

// 9 error keys mirror pickExtracurricularErrorKey + sentinel
// fallbacks (forbidden / notFound / generic).
const errorKeys = [
  'versionConflict',
  'invalidEvent',
  'alreadyRegistered',
  'eventFull',
  'registrationClosed',
  'cannotEdit',
  'forbidden',
  'notFound',
  'generic',
] as const

describe('extracurricular i18n parity × 4 locales', () => {
  it.each(locales)('%s has the top-level extracurricular namespace', (_name, msgs) => {
    expect(msgs.extracurricular).toBeDefined()
    expect(msgs.extracurricular?.title).toBeTruthy()
    expect(msgs.extracurricular?.description).toBeTruthy()
    expect(msgs.extracurricular?.loading).toBeTruthy()
    expect(msgs.extracurricular?.loadFailed).toBeTruthy()
    expect(msgs.extracurricular?.empty).toBeTruthy()
  })

  it.each(locales)('%s has CRUD verbs', (_name, msgs) => {
    expect(msgs.extracurricular?.create).toBeTruthy()
    expect(msgs.extracurricular?.edit).toBeTruthy()
    expect(msgs.extracurricular?.delete).toBeTruthy()
    expect(msgs.extracurricular?.register).toBeTruthy()
    expect(msgs.extracurricular?.unregister).toBeTruthy()
  })

  it.each(locales)('%s has detail-page strings', (_name, msgs) => {
    expect(msgs.extracurricular?.registered).toBeTruthy()
    expect(msgs.extracurricular?.capacity).toBeTruthy()
    expect(msgs.extracurricular?.organizer).toBeTruthy()
    expect(msgs.extracurricular?.participants).toBeTruthy()
    expect(msgs.extracurricular?.location).toBeTruthy()
    expect(msgs.extracurricular?.startAt).toBeTruthy()
    expect(msgs.extracurricular?.endAt).toBeTruthy()
    expect(msgs.extracurricular?.backToList).toBeTruthy()
    expect(msgs.extracurricular?.calendar).toBeTruthy()
    expect(msgs.extracurricular?.upcoming).toBeTruthy()
    expect(msgs.extracurricular?.confirmDelete).toBeTruthy()
    expect(msgs.extracurricular?.confirmUnregister).toBeTruthy()
  })

  it.each(locales)('%s has all status labels', (_name, msgs) => {
    const s = msgs.extracurricular?.status
    expect(s).toBeDefined()
    for (const k of statusKeys) {
      expect(s?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all category labels', (_name, msgs) => {
    const c = msgs.extracurricular?.category
    expect(c).toBeDefined()
    for (const k of categoryKeys) {
      expect(c?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all audience labels', (_name, msgs) => {
    const a = msgs.extracurricular?.audience
    expect(a).toBeDefined()
    for (const k of audienceKeys) {
      expect(a?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all form labels', (_name, msgs) => {
    const f = msgs.extracurricular?.form
    expect(f).toBeDefined()
    for (const k of formKeys) {
      expect(f?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all action labels', (_name, msgs) => {
    const a = msgs.extracurricular?.actions
    expect(a).toBeDefined()
    for (const k of actionKeys) {
      expect(a?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all 9 error messages', (_name, msgs) => {
    const e = msgs.extracurricular?.errors
    expect(e).toBeDefined()
    for (const k of errorKeys) {
      expect(e?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has nav.extracurricular', (_name, msgs) => {
    expect(msgs.nav?.extracurricular).toBeTruthy()
  })
})
