import {
  STUDENT_DEBTS_MANAGE_ROLES,
  STUDENT_DEBTS_REGISTRY_ROLES,
  canManageStudentDebts,
  canViewDebtRegistry,
} from '../permissions'
import { UserRole } from '@/types/auth'

// Student-debts role matrix mirrors the backend isDebtManager predicate
// (internal/modules/student_debts/application/usecases/role_predicates.go)
// and the read scope (read_scope.go / get_debt_usecase.go), design §5:
//   admin / methodist / secretary → manage (import/export, schedule, record)
//                                     + unrestricted read of the registry
//   teacher → read-only, scoped to owned disciplines (registry view, no write)
//   student → reads only own debts via /my (excluded from the registry)

describe('STUDENT_DEBTS_MANAGE_ROLES', () => {
  it('includes admin, methodist, academic_secretary (EDIT_ROLES for debts)', () => {
    expect(STUDENT_DEBTS_MANAGE_ROLES).toContain(UserRole.SYSTEM_ADMIN)
    expect(STUDENT_DEBTS_MANAGE_ROLES).toContain(UserRole.METHODIST)
    expect(STUDENT_DEBTS_MANAGE_ROLES).toContain(UserRole.ACADEMIC_SECRETARY)
  })

  it('excludes teacher (read-only) and student (own-only)', () => {
    expect(STUDENT_DEBTS_MANAGE_ROLES).not.toContain(UserRole.TEACHER)
    expect(STUDENT_DEBTS_MANAGE_ROLES).not.toContain(UserRole.STUDENT)
  })
})

describe('canManageStudentDebts', () => {
  it('is true for admin, methodist, academic_secretary', () => {
    expect(canManageStudentDebts(UserRole.SYSTEM_ADMIN)).toBe(true)
    expect(canManageStudentDebts(UserRole.METHODIST)).toBe(true)
    expect(canManageStudentDebts(UserRole.ACADEMIC_SECRETARY)).toBe(true)
  })

  it('is false for teacher and student', () => {
    expect(canManageStudentDebts(UserRole.TEACHER)).toBe(false)
    expect(canManageStudentDebts(UserRole.STUDENT)).toBe(false)
  })

  it('is false for undefined / empty role', () => {
    expect(canManageStudentDebts(undefined)).toBe(false)
    expect(canManageStudentDebts('')).toBe(false)
  })

  it('accepts string role values', () => {
    expect(canManageStudentDebts('methodist')).toBe(true)
    expect(canManageStudentDebts('teacher')).toBe(false)
  })
})

describe('STUDENT_DEBTS_REGISTRY_ROLES', () => {
  it('includes admin, methodist, secretary, teacher (server scopes teacher)', () => {
    expect(STUDENT_DEBTS_REGISTRY_ROLES).toContain(UserRole.SYSTEM_ADMIN)
    expect(STUDENT_DEBTS_REGISTRY_ROLES).toContain(UserRole.METHODIST)
    expect(STUDENT_DEBTS_REGISTRY_ROLES).toContain(UserRole.ACADEMIC_SECRETARY)
    expect(STUDENT_DEBTS_REGISTRY_ROLES).toContain(UserRole.TEACHER)
  })

  it('excludes student (registry endpoint denies them; they use /my)', () => {
    expect(STUDENT_DEBTS_REGISTRY_ROLES).not.toContain(UserRole.STUDENT)
  })
})

describe('canViewDebtRegistry', () => {
  it('is true for the four staff roles', () => {
    expect(canViewDebtRegistry(UserRole.SYSTEM_ADMIN)).toBe(true)
    expect(canViewDebtRegistry(UserRole.METHODIST)).toBe(true)
    expect(canViewDebtRegistry(UserRole.ACADEMIC_SECRETARY)).toBe(true)
    expect(canViewDebtRegistry(UserRole.TEACHER)).toBe(true)
  })

  it('is false for student and undefined', () => {
    expect(canViewDebtRegistry(UserRole.STUDENT)).toBe(false)
    expect(canViewDebtRegistry(undefined)).toBe(false)
  })
})
