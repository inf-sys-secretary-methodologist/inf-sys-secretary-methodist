import {
  WORK_PROGRAM_CREATE_ROLES,
  WORK_PROGRAM_APPROVE_ROLES,
  canCreateWorkProgram,
  canApproveWorkProgram,
} from '../permissions'
import { UserRole } from '@/types/auth'

// РПД role matrix per work-program ADR-5 (research-verified):
//   teacher  → author (create/submit own discipline)
//   methodist → reserve author + approver
//   system_admin → full override
//   academic_secretary → view-only (owns curriculum, not РПД)
//   student → view approved only (273-ФЗ ст. 29)

describe('WORK_PROGRAM_CREATE_ROLES', () => {
  it('includes teacher (primary author), methodist (reserve), system_admin', () => {
    expect(WORK_PROGRAM_CREATE_ROLES).toContain(UserRole.TEACHER)
    expect(WORK_PROGRAM_CREATE_ROLES).toContain(UserRole.METHODIST)
    expect(WORK_PROGRAM_CREATE_ROLES).toContain(UserRole.SYSTEM_ADMIN)
  })

  it('excludes academic_secretary (view-only) and student', () => {
    expect(WORK_PROGRAM_CREATE_ROLES).not.toContain(UserRole.ACADEMIC_SECRETARY)
    expect(WORK_PROGRAM_CREATE_ROLES).not.toContain(UserRole.STUDENT)
  })
})

describe('WORK_PROGRAM_APPROVE_ROLES', () => {
  it('includes methodist (approver) and system_admin (override)', () => {
    expect(WORK_PROGRAM_APPROVE_ROLES).toContain(UserRole.METHODIST)
    expect(WORK_PROGRAM_APPROVE_ROLES).toContain(UserRole.SYSTEM_ADMIN)
  })

  it('excludes teacher (author cannot approve), academic_secretary, student', () => {
    expect(WORK_PROGRAM_APPROVE_ROLES).not.toContain(UserRole.TEACHER)
    expect(WORK_PROGRAM_APPROVE_ROLES).not.toContain(UserRole.ACADEMIC_SECRETARY)
    expect(WORK_PROGRAM_APPROVE_ROLES).not.toContain(UserRole.STUDENT)
  })
})

describe('canCreateWorkProgram', () => {
  it('is true for teacher, methodist, system_admin', () => {
    expect(canCreateWorkProgram(UserRole.TEACHER)).toBe(true)
    expect(canCreateWorkProgram(UserRole.METHODIST)).toBe(true)
    expect(canCreateWorkProgram(UserRole.SYSTEM_ADMIN)).toBe(true)
  })

  it('is false for academic_secretary, student', () => {
    expect(canCreateWorkProgram(UserRole.ACADEMIC_SECRETARY)).toBe(false)
    expect(canCreateWorkProgram(UserRole.STUDENT)).toBe(false)
  })

  it('is false for undefined / empty role', () => {
    expect(canCreateWorkProgram(undefined)).toBe(false)
    expect(canCreateWorkProgram('')).toBe(false)
  })

  it('accepts string role values', () => {
    expect(canCreateWorkProgram('teacher')).toBe(true)
    expect(canCreateWorkProgram('academic_secretary')).toBe(false)
  })
})

describe('canApproveWorkProgram', () => {
  it('is true for methodist, system_admin', () => {
    expect(canApproveWorkProgram(UserRole.METHODIST)).toBe(true)
    expect(canApproveWorkProgram(UserRole.SYSTEM_ADMIN)).toBe(true)
  })

  it('is false for teacher, academic_secretary, student', () => {
    expect(canApproveWorkProgram(UserRole.TEACHER)).toBe(false)
    expect(canApproveWorkProgram(UserRole.ACADEMIC_SECRETARY)).toBe(false)
    expect(canApproveWorkProgram(UserRole.STUDENT)).toBe(false)
  })

  it('is false for undefined / empty role', () => {
    expect(canApproveWorkProgram(undefined)).toBe(false)
    expect(canApproveWorkProgram('')).toBe(false)
  })
})
