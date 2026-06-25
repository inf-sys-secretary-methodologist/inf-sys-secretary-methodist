import { statusKey, controlFormKey, resitResultKey, STATUS_STYLES } from '../status'
import { STUDENT_DEBT_STATUSES, CONTROL_FORMS, RESIT_RESULTS } from '@/types/studentDebts'

// The wire format stays backend-canonical (snake_case). These mappers
// collapse the multi-token wire values to the camelCase UI keys used under
// studentDebts.card.status.* / .controlForm.* / .detail.resitResult.* — no
// string munging in the type layer, explicit table here.

describe('statusKey', () => {
  it('maps each wire status to its camelCase UI key', () => {
    expect(statusKey('open')).toBe('open')
    expect(statusKey('resit_scheduled')).toBe('resitScheduled')
    expect(statusKey('commission')).toBe('commission')
    expect(statusKey('closed_passed')).toBe('closedPassed')
    expect(statusKey('closed_failed')).toBe('closedFailed')
  })
})

describe('controlFormKey', () => {
  it('maps each wire control form to its camelCase UI key', () => {
    expect(controlFormKey('zachet')).toBe('zachet')
    expect(controlFormKey('exam')).toBe('exam')
    expect(controlFormKey('course_project')).toBe('courseProject')
    expect(controlFormKey('differential_zachet')).toBe('differentialZachet')
  })
})

describe('resitResultKey', () => {
  it('maps each wire resit result to its camelCase UI key', () => {
    expect(resitResultKey('pending')).toBe('pending')
    expect(resitResultKey('passed')).toBe('passed')
    expect(resitResultKey('failed')).toBe('failed')
    expect(resitResultKey('no_show')).toBe('noShow')
  })
})

describe('STATUS_STYLES', () => {
  it('has an entry with bg/text/Icon for every wire status (no missing key)', () => {
    for (const s of STUDENT_DEBT_STATUSES) {
      const style = STATUS_STYLES[s]
      expect(style).toBeDefined()
      expect(typeof style.bg).toBe('string')
      expect(typeof style.text).toBe('string')
      expect(style.Icon).toBeTruthy()
    }
  })

  it('keeps closed_passed and closed_failed visually distinct', () => {
    expect(STATUS_STYLES.closed_passed.bg).not.toBe(STATUS_STYLES.closed_failed.bg)
  })
})

// Guard against drift: the mappers must cover every enum member, so the
// detail/card never interpolate a raw missing i18n key.
describe('enum coverage', () => {
  it('statusKey covers all STUDENT_DEBT_STATUSES', () => {
    for (const s of STUDENT_DEBT_STATUSES) {
      expect(statusKey(s)).toMatch(/^[a-zA-Z]+$/)
    }
  })
  it('controlFormKey covers all CONTROL_FORMS', () => {
    for (const c of CONTROL_FORMS) {
      expect(controlFormKey(c)).toMatch(/^[a-zA-Z]+$/)
    }
  })
  it('resitResultKey covers all RESIT_RESULTS', () => {
    for (const r of RESIT_RESULTS) {
      expect(resitResultKey(r)).toMatch(/^[a-zA-Z]+$/)
    }
  })
})
