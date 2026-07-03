import { SCHEDULE_GENERATE_ROLES, canGenerateSchedule } from '../permissions'
import { UserRole } from '@/types/auth'

// The schedule-generation role set mirrors the backend write gate in
// internal/modules/schedule/interfaces/http/handlers/generate_schedule_handler.go
// (requireGenerateWrite): academic planning belongs to methodist + academic
// secretary, admin retains override. Teacher and student only read the
// resulting timetable, so both are excluded from preview/apply.

describe('SCHEDULE_GENERATE_ROLES', () => {
  it('includes admin, methodist, academic_secretary', () => {
    expect(SCHEDULE_GENERATE_ROLES).toContain(UserRole.SYSTEM_ADMIN)
    expect(SCHEDULE_GENERATE_ROLES).toContain(UserRole.METHODIST)
    expect(SCHEDULE_GENERATE_ROLES).toContain(UserRole.ACADEMIC_SECRETARY)
  })

  it('excludes teacher and student (read-only on the timetable)', () => {
    expect(SCHEDULE_GENERATE_ROLES).not.toContain(UserRole.TEACHER)
    expect(SCHEDULE_GENERATE_ROLES).not.toContain(UserRole.STUDENT)
  })
})

describe('canGenerateSchedule', () => {
  it('is true for admin, methodist, academic_secretary', () => {
    expect(canGenerateSchedule(UserRole.SYSTEM_ADMIN)).toBe(true)
    expect(canGenerateSchedule(UserRole.METHODIST)).toBe(true)
    expect(canGenerateSchedule(UserRole.ACADEMIC_SECRETARY)).toBe(true)
  })

  it('is false for teacher and student', () => {
    expect(canGenerateSchedule(UserRole.TEACHER)).toBe(false)
    expect(canGenerateSchedule(UserRole.STUDENT)).toBe(false)
  })

  it('is false for undefined / empty role', () => {
    expect(canGenerateSchedule(undefined)).toBe(false)
    expect(canGenerateSchedule('')).toBe(false)
  })

  it('accepts string role values', () => {
    expect(canGenerateSchedule('methodist')).toBe(true)
    expect(canGenerateSchedule('teacher')).toBe(false)
  })
})
