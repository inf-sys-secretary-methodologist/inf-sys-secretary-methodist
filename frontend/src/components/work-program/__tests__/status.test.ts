import { STATUS_STYLES, statusKey, revisionStatusKey } from '../status'
import { WORK_PROGRAM_STATUSES } from '@/types/workProgram'

describe('work-program statusKey', () => {
  it('collapses pending_approval to the short UI key "pending"', () => {
    expect(statusKey('pending_approval')).toBe('pending')
  })

  it('collapses needs_revision to the camelCase UI key "needsRevision"', () => {
    expect(statusKey('needs_revision')).toBe('needsRevision')
  })

  it('passes single-token statuses through unchanged', () => {
    expect(statusKey('draft')).toBe('draft')
    expect(statusKey('approved')).toBe('approved')
    expect(statusKey('archived')).toBe('archived')
  })
})

describe('work-program revisionStatusKey', () => {
  it('collapses revision pending_approval to the short UI key "pending"', () => {
    expect(revisionStatusKey('pending_approval')).toBe('pending')
  })

  it('passes draft / approved / rejected through unchanged', () => {
    expect(revisionStatusKey('draft')).toBe('draft')
    expect(revisionStatusKey('approved')).toBe('approved')
    expect(revisionStatusKey('rejected')).toBe('rejected')
  })
})

describe('work-program STATUS_STYLES', () => {
  it('has a style entry for every lifecycle status', () => {
    for (const s of WORK_PROGRAM_STATUSES) {
      expect(STATUS_STYLES[s]).toBeDefined()
      expect(STATUS_STYLES[s].bg).toBeTruthy()
      expect(STATUS_STYLES[s].text).toBeTruthy()
      expect(STATUS_STYLES[s].Icon).toBeTruthy()
    }
  })
})
