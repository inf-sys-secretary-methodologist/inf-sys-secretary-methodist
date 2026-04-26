import {
  EDIT_ROLES,
  VIEW_ONLY_ROLES,
  canEdit,
  canCreate,
  canDelete,
  isViewOnly,
  isAdmin,
  Resource,
  Action,
  AccessLevel,
  can,
  getAccessLevel,
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

describe('Resource enum', () => {
  it('has all 8 resources', () => {
    expect(Object.values(Resource)).toHaveLength(8)
    expect(Resource.USERS).toBe('users')
    expect(Resource.CURRICULUM).toBe('curriculum')
    expect(Resource.SCHEDULE).toBe('schedule')
    expect(Resource.ASSIGNMENTS).toBe('assignments')
    expect(Resource.REPORTS).toBe('reports')
    expect(Resource.INTEGRATION).toBe('integration')
    expect(Resource.SYSTEM_SETTINGS).toBe('system_settings')
    expect(Resource.PERSONAL_SETTINGS).toBe('personal_settings')
  })
})

describe('Action enum', () => {
  it('has all 5 actions', () => {
    expect(Object.values(Action)).toHaveLength(5)
    expect(Action.READ).toBe('read')
    expect(Action.CREATE).toBe('create')
    expect(Action.UPDATE).toBe('update')
    expect(Action.DELETE).toBe('delete')
    expect(Action.APPROVE).toBe('approve')
  })
})

describe('AccessLevel enum', () => {
  it('has 4 levels in correct order', () => {
    expect(AccessLevel.DENIED).toBe(0)
    expect(AccessLevel.LIMITED).toBe(1)
    expect(AccessLevel.OWN).toBe(2)
    expect(AccessLevel.FULL).toBe(3)
  })
})

describe('getAccessLevel', () => {
  it('returns full for system_admin on any resource', () => {
    expect(getAccessLevel(UserRole.SYSTEM_ADMIN, Resource.USERS)).toBe(AccessLevel.FULL)
    expect(getAccessLevel(UserRole.SYSTEM_ADMIN, Resource.INTEGRATION)).toBe(AccessLevel.FULL)
    expect(getAccessLevel(UserRole.SYSTEM_ADMIN, Resource.SYSTEM_SETTINGS)).toBe(AccessLevel.FULL)
  })

  it('returns own for personal_settings for all roles', () => {
    expect(getAccessLevel(UserRole.SYSTEM_ADMIN, Resource.PERSONAL_SETTINGS)).toBe(AccessLevel.OWN)
    expect(getAccessLevel(UserRole.STUDENT, Resource.PERSONAL_SETTINGS)).toBe(AccessLevel.OWN)
    expect(getAccessLevel(UserRole.TEACHER, Resource.PERSONAL_SETTINGS)).toBe(AccessLevel.OWN)
  })

  it('returns denied for student on reports', () => {
    expect(getAccessLevel(UserRole.STUDENT, Resource.REPORTS)).toBe(AccessLevel.DENIED)
  })

  it('returns denied for non-admin on integration', () => {
    expect(getAccessLevel(UserRole.METHODIST, Resource.INTEGRATION)).toBe(AccessLevel.DENIED)
    expect(getAccessLevel(UserRole.TEACHER, Resource.INTEGRATION)).toBe(AccessLevel.DENIED)
    expect(getAccessLevel(UserRole.STUDENT, Resource.INTEGRATION)).toBe(AccessLevel.DENIED)
  })

  it('returns denied for non-admin on system_settings', () => {
    expect(getAccessLevel(UserRole.METHODIST, Resource.SYSTEM_SETTINGS)).toBe(AccessLevel.DENIED)
    expect(getAccessLevel(UserRole.ACADEMIC_SECRETARY, Resource.SYSTEM_SETTINGS)).toBe(AccessLevel.DENIED)
  })

  it('returns denied for undefined role', () => {
    expect(getAccessLevel(undefined, Resource.USERS)).toBe(AccessLevel.DENIED)
  })
})

describe('can', () => {
  describe('system_admin has full access to everything', () => {
    it.each([
      [Resource.USERS, Action.CREATE],
      [Resource.USERS, Action.DELETE],
      [Resource.CURRICULUM, Action.APPROVE],
      [Resource.INTEGRATION, Action.UPDATE],
      [Resource.SYSTEM_SETTINGS, Action.UPDATE],
      [Resource.REPORTS, Action.CREATE],
      [Resource.SCHEDULE, Action.CREATE],
    ])('can %s.%s', (resource, action) => {
      expect(can(UserRole.SYSTEM_ADMIN, resource, action)).toBe(true)
    })
  })

  describe('student restrictions', () => {
    it.each([
      [Resource.REPORTS, Action.READ],
      [Resource.REPORTS, Action.CREATE],
      [Resource.INTEGRATION, Action.READ],
      [Resource.SYSTEM_SETTINGS, Action.READ],
      [Resource.USERS, Action.CREATE],
      [Resource.USERS, Action.DELETE],
      [Resource.CURRICULUM, Action.CREATE],
      [Resource.SCHEDULE, Action.CREATE],
    ])('cannot %s.%s', (resource, action) => {
      expect(can(UserRole.STUDENT, resource, action)).toBe(false)
    })

    it('can read+update own personal_settings', () => {
      expect(can(UserRole.STUDENT, Resource.PERSONAL_SETTINGS, Action.READ)).toBe(true)
      expect(can(UserRole.STUDENT, Resource.PERSONAL_SETTINGS, Action.UPDATE)).toBe(true)
    })

    it('can read schedule', () => {
      expect(can(UserRole.STUDENT, Resource.SCHEDULE, Action.READ)).toBe(true)
    })

    it('can read assignments', () => {
      expect(can(UserRole.STUDENT, Resource.ASSIGNMENTS, Action.READ)).toBe(true)
    })
  })

  describe('methodist permissions', () => {
    it('has full access to curriculum (except approve)', () => {
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.CREATE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.UPDATE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.READ)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.APPROVE)).toBe(false)
    })

    it('has full access to reports', () => {
      expect(can(UserRole.METHODIST, Resource.REPORTS, Action.READ)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.REPORTS, Action.CREATE)).toBe(true)
    })

    it('denied integration', () => {
      expect(can(UserRole.METHODIST, Resource.INTEGRATION, Action.READ)).toBe(false)
    })

    it('limited schedule access (read + limited update, no create)', () => {
      expect(can(UserRole.METHODIST, Resource.SCHEDULE, Action.READ)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.SCHEDULE, Action.UPDATE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.SCHEDULE, Action.CREATE)).toBe(false)
    })
  })

  describe('academic_secretary permissions', () => {
    it('has full access to schedule', () => {
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.SCHEDULE, Action.CREATE)).toBe(true)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.SCHEDULE, Action.UPDATE)).toBe(true)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.SCHEDULE, Action.DELETE)).toBe(true)
    })

    it('has full access to reports', () => {
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.REPORTS, Action.CREATE)).toBe(true)
    })

    it('can only read curriculum', () => {
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.CURRICULUM, Action.READ)).toBe(true)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.CURRICULUM, Action.CREATE)).toBe(false)
    })
  })

  describe('teacher permissions', () => {
    it('limited reports access (read only)', () => {
      expect(can(UserRole.TEACHER, Resource.REPORTS, Action.READ)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.REPORTS, Action.CREATE)).toBe(false)
    })

    it('can read schedule but not create', () => {
      expect(can(UserRole.TEACHER, Resource.SCHEDULE, Action.READ)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.SCHEDULE, Action.CREATE)).toBe(false)
    })

    it('has full assignments access (create own)', () => {
      expect(can(UserRole.TEACHER, Resource.ASSIGNMENTS, Action.CREATE)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.ASSIGNMENTS, Action.READ)).toBe(true)
    })

    it('can read+update curriculum (limited)', () => {
      expect(can(UserRole.TEACHER, Resource.CURRICULUM, Action.READ)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.CURRICULUM, Action.UPDATE)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.CURRICULUM, Action.CREATE)).toBe(false)
    })

    it('denied integration and system_settings', () => {
      expect(can(UserRole.TEACHER, Resource.INTEGRATION, Action.READ)).toBe(false)
      expect(can(UserRole.TEACHER, Resource.SYSTEM_SETTINGS, Action.READ)).toBe(false)
    })
  })

  describe('approve action', () => {
    it('only system_admin can approve curriculum', () => {
      expect(can(UserRole.SYSTEM_ADMIN, Resource.CURRICULUM, Action.APPROVE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.APPROVE)).toBe(false)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.CURRICULUM, Action.APPROVE)).toBe(false)
      expect(can(UserRole.TEACHER, Resource.CURRICULUM, Action.APPROVE)).toBe(false)
      expect(can(UserRole.STUDENT, Resource.CURRICULUM, Action.APPROVE)).toBe(false)
    })
  })

  describe('edge cases', () => {
    it('returns false for undefined role', () => {
      expect(can(undefined, Resource.USERS, Action.READ)).toBe(false)
    })

    it('returns false for empty string role', () => {
      expect(can('', Resource.USERS, Action.READ)).toBe(false)
    })

    it('accepts string role values', () => {
      expect(can('system_admin', Resource.USERS, Action.CREATE)).toBe(true)
      expect(can('student', Resource.REPORTS, Action.READ)).toBe(false)
    })
  })
})
