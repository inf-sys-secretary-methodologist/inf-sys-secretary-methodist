// i18n parity × 4 locales for the workProgram (РПД) namespace.
// Raw JSON load — a missing key in any locale fails the build (per
// feedback_i18n_json_load_parity_test). Mirrors the extracurricular
// parity test.

import ru from '../../../messages/ru.json'
import en from '../../../messages/en.json'
import fr from '../../../messages/fr.json'
import ar from '../../../messages/ar.json'

type EnumGroup = Record<string, string | undefined>

type MessagesShape = {
  workProgram?: {
    title?: string
    description?: string
    loading?: string
    loadFailed?: string
    createButton?: string
    countLabel?: string
    empty?: EnumGroup
    pagination?: EnumGroup
    filters?: {
      status?: string
      specialty?: string
      specialtyPlaceholder?: string
      year?: string
      yearPlaceholder?: string
      statusOptions?: EnumGroup
    }
    card?: {
      openAria?: string
      discipline?: string
      year?: string
      specialty?: string
      status?: EnumGroup
    }
    errors?: EnumGroup
  }
  nav?: {
    workPrograms?: string
  }
}

const locales: Array<readonly [string, MessagesShape]> = [
  ['ru', ru as MessagesShape],
  ['en', en as MessagesShape],
  ['fr', fr as MessagesShape],
  ['ar', ar as MessagesShape],
]

const statusOptionKeys = [
  'all',
  'draft',
  'pending',
  'approved',
  'needsRevision',
  'archived',
] as const

const cardStatusKeys = ['draft', 'pending', 'approved', 'needsRevision', 'archived'] as const

// 8 error keys mirror pickWorkProgramErrorKey + sentinel fallbacks.
const errorKeys = [
  'identityExists',
  'versionConflict',
  'invalidTransition',
  'rejectReasonRequired',
  'invalidWorkProgram',
  'forbidden',
  'notFound',
  'generic',
] as const

describe('workProgram i18n parity × 4 locales', () => {
  it.each(locales)('%s has the top-level workProgram strings', (_name, msgs) => {
    expect(msgs.workProgram).toBeDefined()
    expect(msgs.workProgram?.title).toBeTruthy()
    expect(msgs.workProgram?.description).toBeTruthy()
    expect(msgs.workProgram?.loading).toBeTruthy()
    expect(msgs.workProgram?.loadFailed).toBeTruthy()
    expect(msgs.workProgram?.createButton).toBeTruthy()
    expect(msgs.workProgram?.countLabel).toBeTruthy()
  })

  it.each(locales)('%s has empty-state strings', (_name, msgs) => {
    expect(msgs.workProgram?.empty?.title).toBeTruthy()
    expect(msgs.workProgram?.empty?.description).toBeTruthy()
  })

  it.each(locales)('%s has pagination strings', (_name, msgs) => {
    expect(msgs.workProgram?.pagination?.prev).toBeTruthy()
    expect(msgs.workProgram?.pagination?.next).toBeTruthy()
  })

  it.each(locales)('%s has filter labels', (_name, msgs) => {
    const f = msgs.workProgram?.filters
    expect(f?.status).toBeTruthy()
    expect(f?.specialty).toBeTruthy()
    expect(f?.specialtyPlaceholder).toBeTruthy()
    expect(f?.year).toBeTruthy()
    expect(f?.yearPlaceholder).toBeTruthy()
  })

  it.each(locales)('%s has all status filter options', (_name, msgs) => {
    const s = msgs.workProgram?.filters?.statusOptions
    expect(s).toBeDefined()
    for (const k of statusOptionKeys) {
      expect(s?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has card labels + all status pills', (_name, msgs) => {
    const c = msgs.workProgram?.card
    expect(c?.openAria).toBeTruthy()
    expect(c?.discipline).toBeTruthy()
    expect(c?.year).toBeTruthy()
    expect(c?.specialty).toBeTruthy()
    for (const k of cardStatusKeys) {
      expect(c?.status?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has all 8 error messages', (_name, msgs) => {
    const e = msgs.workProgram?.errors
    expect(e).toBeDefined()
    for (const k of errorKeys) {
      expect(e?.[k]).toBeTruthy()
    }
  })

  it.each(locales)('%s has nav.workPrograms', (_name, msgs) => {
    expect(msgs.nav?.workPrograms).toBeTruthy()
  })
})
