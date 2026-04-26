import { can, Resource, Action } from '../permissions'
import { UserRole } from '@/types/auth'

describe('permission matrix — page-level scenarios', () => {
  describe('users page — admin-only CRUD', () => {
    it('only admin can create/delete users', () => {
      expect(can(UserRole.SYSTEM_ADMIN, Resource.USERS, Action.CREATE)).toBe(true)
      expect(can(UserRole.SYSTEM_ADMIN, Resource.USERS, Action.DELETE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.USERS, Action.CREATE)).toBe(false)
      expect(can(UserRole.TEACHER, Resource.USERS, Action.DELETE)).toBe(false)
    })

    it('non-admin roles can read users (limited)', () => {
      expect(can(UserRole.METHODIST, Resource.USERS, Action.READ)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.USERS, Action.READ)).toBe(true)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.USERS, Action.READ)).toBe(true)
    })

    it('student can only update own profile', () => {
      expect(can(UserRole.STUDENT, Resource.USERS, Action.UPDATE)).toBe(true)
      expect(can(UserRole.STUDENT, Resource.USERS, Action.CREATE)).toBe(false)
      expect(can(UserRole.STUDENT, Resource.USERS, Action.DELETE)).toBe(false)
    })
  })

  describe('schedule page', () => {
    it('secretary has full schedule control', () => {
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.SCHEDULE, Action.CREATE)).toBe(true)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.SCHEDULE, Action.DELETE)).toBe(true)
    })

    it('methodist can read and update schedule (limited)', () => {
      expect(can(UserRole.METHODIST, Resource.SCHEDULE, Action.READ)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.SCHEDULE, Action.UPDATE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.SCHEDULE, Action.CREATE)).toBe(false)
    })

    it('teacher and student can only read schedule', () => {
      expect(can(UserRole.TEACHER, Resource.SCHEDULE, Action.READ)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.SCHEDULE, Action.CREATE)).toBe(false)
      expect(can(UserRole.STUDENT, Resource.SCHEDULE, Action.READ)).toBe(true)
      expect(can(UserRole.STUDENT, Resource.SCHEDULE, Action.CREATE)).toBe(false)
    })
  })

  describe('reports page', () => {
    it('admin/methodist/secretary have full reports access', () => {
      expect(can(UserRole.SYSTEM_ADMIN, Resource.REPORTS, Action.CREATE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.REPORTS, Action.CREATE)).toBe(true)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.REPORTS, Action.CREATE)).toBe(true)
    })

    it('teacher has limited (read only)', () => {
      expect(can(UserRole.TEACHER, Resource.REPORTS, Action.READ)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.REPORTS, Action.CREATE)).toBe(false)
    })

    it('student denied', () => {
      expect(can(UserRole.STUDENT, Resource.REPORTS, Action.READ)).toBe(false)
    })
  })

  describe('integration page — admin only', () => {
    it.each([
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
      UserRole.STUDENT,
    ])('%s cannot access integration', (role) => {
      expect(can(role, Resource.INTEGRATION, Action.READ)).toBe(false)
    })

    it('admin has full integration access', () => {
      expect(can(UserRole.SYSTEM_ADMIN, Resource.INTEGRATION, Action.CREATE)).toBe(true)
      expect(can(UserRole.SYSTEM_ADMIN, Resource.INTEGRATION, Action.UPDATE)).toBe(true)
    })
  })

  describe('curriculum (documents/templates)', () => {
    it('methodist can create and edit curriculum', () => {
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.CREATE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.UPDATE)).toBe(true)
    })

    it('teacher can read and update curriculum (limited)', () => {
      expect(can(UserRole.TEACHER, Resource.CURRICULUM, Action.READ)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.CURRICULUM, Action.UPDATE)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.CURRICULUM, Action.CREATE)).toBe(false)
    })

    it('secretary can only read curriculum', () => {
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.CURRICULUM, Action.READ)).toBe(true)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.CURRICULUM, Action.CREATE)).toBe(false)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.CURRICULUM, Action.UPDATE)).toBe(false)
    })
  })

  describe('assignments (tasks)', () => {
    it('teacher can create assignments for own classes', () => {
      expect(can(UserRole.TEACHER, Resource.ASSIGNMENTS, Action.CREATE)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.ASSIGNMENTS, Action.READ)).toBe(true)
    })

    it('student can read and update own assignments', () => {
      expect(can(UserRole.STUDENT, Resource.ASSIGNMENTS, Action.READ)).toBe(true)
      expect(can(UserRole.STUDENT, Resource.ASSIGNMENTS, Action.UPDATE)).toBe(true)
      expect(can(UserRole.STUDENT, Resource.ASSIGNMENTS, Action.CREATE)).toBe(false)
    })
  })
})
