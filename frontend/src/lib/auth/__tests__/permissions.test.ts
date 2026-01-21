import {
  EDIT_ROLES,
  VIEW_ONLY_ROLES,
  canEdit,
  canCreate,
  canDelete,
  isViewOnly,
  isAdmin,
} from '../permissions'
import { UserRole } from '@/types/auth'

describe('permissions constants', () => {
  describe('EDIT_ROLES', () => {
    it('includes SYSTEM_ADMIN', () => {
      expect(EDIT_ROLES).toContain(UserRole.SYSTEM_ADMIN)
    })

    it('includes METHODIST', () => {
      expect(EDIT_ROLES).toContain(UserRole.METHODIST)
    })

    it('includes ACADEMIC_SECRETARY', () => {
      expect(EDIT_ROLES).toContain(UserRole.ACADEMIC_SECRETARY)
    })

    it('includes TEACHER', () => {
      expect(EDIT_ROLES).toContain(UserRole.TEACHER)
    })

    it('does not include STUDENT', () => {
      expect(EDIT_ROLES).not.toContain(UserRole.STUDENT)
    })
  })

  describe('VIEW_ONLY_ROLES', () => {
    it('includes STUDENT', () => {
      expect(VIEW_ONLY_ROLES).toContain(UserRole.STUDENT)
    })

    it('does not include SYSTEM_ADMIN', () => {
      expect(VIEW_ONLY_ROLES).not.toContain(UserRole.SYSTEM_ADMIN)
    })
  })
})

describe('canEdit', () => {
  it('returns true for SYSTEM_ADMIN', () => {
    expect(canEdit(UserRole.SYSTEM_ADMIN)).toBe(true)
  })

  it('returns true for METHODIST', () => {
    expect(canEdit(UserRole.METHODIST)).toBe(true)
  })

  it('returns true for ACADEMIC_SECRETARY', () => {
    expect(canEdit(UserRole.ACADEMIC_SECRETARY)).toBe(true)
  })

  it('returns true for TEACHER', () => {
    expect(canEdit(UserRole.TEACHER)).toBe(true)
  })

  it('returns false for STUDENT', () => {
    expect(canEdit(UserRole.STUDENT)).toBe(false)
  })

  it('returns false for undefined', () => {
    expect(canEdit(undefined)).toBe(false)
  })

  it('returns false for empty string', () => {
    expect(canEdit('')).toBe(false)
  })

  it('accepts string role values', () => {
    expect(canEdit('system_admin')).toBe(true)
    expect(canEdit('student')).toBe(false)
  })
})

describe('canCreate', () => {
  it('returns true for edit roles', () => {
    expect(canCreate(UserRole.SYSTEM_ADMIN)).toBe(true)
    expect(canCreate(UserRole.TEACHER)).toBe(true)
  })

  it('returns false for view-only roles', () => {
    expect(canCreate(UserRole.STUDENT)).toBe(false)
  })

  it('returns false for undefined', () => {
    expect(canCreate(undefined)).toBe(false)
  })
})

describe('canDelete', () => {
  it('returns true for edit roles', () => {
    expect(canDelete(UserRole.SYSTEM_ADMIN)).toBe(true)
    expect(canDelete(UserRole.METHODIST)).toBe(true)
  })

  it('returns false for view-only roles', () => {
    expect(canDelete(UserRole.STUDENT)).toBe(false)
  })

  it('returns false for undefined', () => {
    expect(canDelete(undefined)).toBe(false)
  })
})

describe('isViewOnly', () => {
  it('returns true for STUDENT', () => {
    expect(isViewOnly(UserRole.STUDENT)).toBe(true)
  })

  it('returns false for SYSTEM_ADMIN', () => {
    expect(isViewOnly(UserRole.SYSTEM_ADMIN)).toBe(false)
  })

  it('returns false for TEACHER', () => {
    expect(isViewOnly(UserRole.TEACHER)).toBe(false)
  })

  it('returns true for undefined', () => {
    expect(isViewOnly(undefined)).toBe(true)
  })

  it('returns true for empty string', () => {
    expect(isViewOnly('')).toBe(true)
  })
})

describe('isAdmin', () => {
  it('returns true for SYSTEM_ADMIN', () => {
    expect(isAdmin(UserRole.SYSTEM_ADMIN)).toBe(true)
  })

  it('returns false for other roles', () => {
    expect(isAdmin(UserRole.METHODIST)).toBe(false)
    expect(isAdmin(UserRole.ACADEMIC_SECRETARY)).toBe(false)
    expect(isAdmin(UserRole.TEACHER)).toBe(false)
    expect(isAdmin(UserRole.STUDENT)).toBe(false)
  })

  it('returns false for undefined', () => {
    expect(isAdmin(undefined)).toBe(false)
  })

  it('accepts string role value', () => {
    expect(isAdmin('system_admin')).toBe(true)
    expect(isAdmin('teacher')).toBe(false)
  })
})
