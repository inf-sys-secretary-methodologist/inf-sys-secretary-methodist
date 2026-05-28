import {
  navigationConfig,
  getAvailableNavEntries,
  getAvailableNavItems,
  isNavGroup,
  NavEntry,
  NavGroup,
} from '../navigation'
import { UserRole } from '@/types/auth'

describe('navigationConfig', () => {
  it('contains dashboard item', () => {
    const dashboardEntry = navigationConfig.find((entry) => entry.nameKey === 'dashboard')
    expect(dashboardEntry).toBeDefined()
    expect(isNavGroup(dashboardEntry!)).toBe(false)
    if (dashboardEntry && !isNavGroup(dashboardEntry)) {
      expect(dashboardEntry.url).toBe('/dashboard')
    }
  })

  it('contains all expected navigation entries', () => {
    const expectedKeys = [
      'dashboard',
      'documentsGroup',
      'analyticsGroup',
      'educationGroup',
      'communicationGroup',
      'adminGroup',
    ]
    const actualKeys = navigationConfig.map((entry) => entry.nameKey)
    expect(actualKeys).toEqual(expectedKeys)
  })

  it('all entries have required properties', () => {
    navigationConfig.forEach((entry) => {
      expect(entry.nameKey).toBeDefined()
      expect(entry.icon).toBeDefined()
      expect(typeof entry.nameKey).toBe('string')

      if (isNavGroup(entry)) {
        expect(entry.items).toBeDefined()
        expect(Array.isArray(entry.items)).toBe(true)
        expect(entry.items.length).toBeGreaterThan(0)
        entry.items.forEach((item) => {
          expect(item.nameKey).toBeDefined()
          expect(item.url).toBeDefined()
          expect(item.icon).toBeDefined()
        })
      } else {
        expect(entry.url).toBeDefined()
      }
    })
  })

  it('documentsGroup contains documents, files, and templates', () => {
    const docsGroup = navigationConfig.find((e) => e.nameKey === 'documentsGroup') as NavGroup
    expect(docsGroup).toBeDefined()
    expect(isNavGroup(docsGroup)).toBe(true)
    const itemKeys = docsGroup.items.map((i) => i.nameKey)
    expect(itemKeys).toContain('documents')
    expect(itemKeys).toContain('files')
    expect(itemKeys).toContain('templates')
  })

  it('educationGroup contains schedule, calendar, tasks, assignments, and curriculum', () => {
    const eduGroup = navigationConfig.find((e) => e.nameKey === 'educationGroup') as NavGroup
    expect(eduGroup).toBeDefined()
    expect(isNavGroup(eduGroup)).toBe(true)
    const itemKeys = eduGroup.items.map((i) => i.nameKey)
    expect(itemKeys).toContain('schedule')
    expect(itemKeys).toContain('calendar')
    expect(itemKeys).toContain('tasks')
    expect(itemKeys).toContain('assignments')
    expect(itemKeys).toContain('curriculum')
  })

  it('assignments entry is hidden from STUDENT and visible to all four non-student roles', () => {
    const eduGroup = navigationConfig.find((e) => e.nameKey === 'educationGroup') as NavGroup
    const assignments = eduGroup.items.find((i) => i.nameKey === 'assignments')
    expect(assignments).toBeDefined()
    expect(assignments!.url).toBe('/assignments')
    // Student must NOT appear in the roles whitelist — the page is the
    // grading view; the backend RequireNonStudent middleware is the
    // real gate, but the navigation already excludes student-side
    // visibility to avoid a dead-link round-trip.
    expect(assignments!.roles).not.toContain(UserRole.STUDENT)
    expect(assignments!.roles).toEqual(
      expect.arrayContaining([
        UserRole.SYSTEM_ADMIN,
        UserRole.METHODIST,
        UserRole.ACADEMIC_SECRETARY,
        UserRole.TEACHER,
      ])
    )
  })

  it('curriculum entry is hidden from STUDENT and visible to all four non-student roles', () => {
    const eduGroup = navigationConfig.find((e) => e.nameKey === 'educationGroup') as NavGroup
    const curriculum = eduGroup.items.find((i) => i.nameKey === 'curriculum')
    expect(curriculum).toBeDefined()
    expect(curriculum!.url).toBe('/curriculum')
    // Student is excluded by GET /api/curriculum RequireNonStudent
    // gate (v0.116.0); navigation mirrors that to avoid a dead link.
    expect(curriculum!.roles).not.toContain(UserRole.STUDENT)
    expect(curriculum!.roles).toEqual(
      expect.arrayContaining([
        UserRole.SYSTEM_ADMIN,
        UserRole.METHODIST,
        UserRole.ACADEMIC_SECRETARY,
        UserRole.TEACHER,
      ])
    )
  })

  it('educationGroup contains workPrograms (РПД) visible to all 5 roles incl student', () => {
    const eduGroup = navigationConfig.find((e) => e.nameKey === 'educationGroup') as NavGroup
    const wp = eduGroup.items.find((i) => i.nameKey === 'workPrograms')
    expect(wp).toBeDefined()
    expect(wp!.url).toBe('/work-programs')
    // 273-ФЗ ст. 29: students see approved РПД (open ЭИОС), so unlike
    // curriculum the entry is visible to all five roles; the backend
    // role-scopes the rows.
    expect(wp!.roles).toEqual(
      expect.arrayContaining([
        UserRole.SYSTEM_ADMIN,
        UserRole.METHODIST,
        UserRole.ACADEMIC_SECRETARY,
        UserRole.TEACHER,
        UserRole.STUDENT,
      ])
    )
  })

  it('workPrograms entry is visible to STUDENT (flat items)', () => {
    const flatKeys = getAvailableNavItems(UserRole.STUDENT).map((i) => i.nameKey)
    expect(flatKeys).toContain('workPrograms')
  })

  it('communicationGroup contains announcements, messages, and aiAssistant', () => {
    const commGroup = navigationConfig.find((e) => e.nameKey === 'communicationGroup') as NavGroup
    expect(commGroup).toBeDefined()
    expect(isNavGroup(commGroup)).toBe(true)
    const itemKeys = commGroup.items.map((i) => i.nameKey)
    expect(itemKeys).toContain('announcements')
    expect(itemKeys).toContain('messages')
    expect(itemKeys).toContain('aiAssistant')
  })

  it('adminGroup contains users, integration, settings', () => {
    const adminGroup = navigationConfig.find((e) => e.nameKey === 'adminGroup') as NavGroup
    expect(adminGroup).toBeDefined()
    expect(isNavGroup(adminGroup)).toBe(true)
    const itemKeys = adminGroup.items.map((i) => i.nameKey)
    expect(itemKeys).toContain('usersCatalog')
    expect(itemKeys).toContain('integration')
  })

  it('adminGroup contains curriculumApprove visible to methodist + SYSTEM_ADMIN', () => {
    const adminGroup = navigationConfig.find((e) => e.nameKey === 'adminGroup') as NavGroup
    const entry = adminGroup.items.find((i) => i.nameKey === 'curriculumApprove')
    expect(entry).toBeDefined()
    expect(entry!.url).toBe('/admin/curriculum/approve')
    // v0.158.0: methodist is the primary approver, system_admin retains
    // emergency override. Academic_secretary excluded (author cannot
    // approve own work).
    expect(entry!.roles).toEqual(
      expect.arrayContaining([UserRole.SYSTEM_ADMIN, UserRole.METHODIST])
    )
    expect(entry!.roles).not.toContain(UserRole.ACADEMIC_SECRETARY)
    expect(entry!.roles).not.toContain(UserRole.TEACHER)
    expect(entry!.roles).not.toContain(UserRole.STUDENT)
  })

  it('analyticsGroup contains annualReport visible only to methodist + system_admin', () => {
    const analyticsGroup = navigationConfig.find((e) => e.nameKey === 'analyticsGroup') as NavGroup
    expect(analyticsGroup).toBeDefined()
    const entry = analyticsGroup.items.find((i) => i.nameKey === 'annualReport')
    expect(entry).toBeDefined()
    expect(entry!.url).toBe('/reports/annual')
    // Mirror к backend ADR-6 — academic_secretary excluded (observer, not
    // decision-maker); teacher / student also out (no aggregate access).
    expect(entry!.roles).toEqual(
      expect.arrayContaining([UserRole.METHODIST, UserRole.SYSTEM_ADMIN])
    )
    expect(entry!.roles).not.toContain(UserRole.ACADEMIC_SECRETARY)
    expect(entry!.roles).not.toContain(UserRole.TEACHER)
    expect(entry!.roles).not.toContain(UserRole.STUDENT)
  })
})

describe('isNavGroup', () => {
  it('returns true for groups', () => {
    const group = {
      nameKey: 'test',
      icon: () => null,
      items: [{ nameKey: 'item', url: '/item', icon: () => null }],
    } as unknown as NavEntry
    expect(isNavGroup(group)).toBe(true)
  })

  it('returns false for items', () => {
    const item = {
      nameKey: 'test',
      url: '/test',
      icon: () => null,
    } as unknown as NavEntry
    expect(isNavGroup(item)).toBe(false)
  })
})

describe('getAvailableNavEntries', () => {
  it('returns empty array when no role provided', () => {
    expect(getAvailableNavEntries()).toEqual([])
    expect(getAvailableNavEntries(undefined)).toEqual([])
  })

  it('returns all entries for SYSTEM_ADMIN role', () => {
    const entries = getAvailableNavEntries(UserRole.SYSTEM_ADMIN)
    expect(entries.length).toBe(navigationConfig.length)
  })

  it('returns correct entries for STUDENT role', () => {
    const entries = getAvailableNavEntries(UserRole.STUDENT)
    const entryKeys = entries.map((e) => e.nameKey)

    expect(entryKeys).toContain('dashboard')
    expect(entryKeys).toContain('documentsGroup')
    expect(entryKeys).toContain('educationGroup')
    expect(entryKeys).toContain('communicationGroup')

    expect(entryKeys).not.toContain('analyticsGroup')
    expect(entryKeys).not.toContain('adminGroup')
  })

  it('returns correct entries for TEACHER role', () => {
    const entries = getAvailableNavEntries(UserRole.TEACHER)
    const entryKeys = entries.map((e) => e.nameKey)

    expect(entryKeys).toContain('dashboard')
    expect(entryKeys).toContain('educationGroup')
    expect(entryKeys).toContain('communicationGroup')

    // Teacher sees reports (limited) but not analytics — analyticsGroup has only 1 item
    // for teacher (reports), so it gets flattened to direct 'reports' link
    expect(entryKeys).toContain('reports')
    expect(entryKeys).not.toContain('analyticsGroup')
  })

  it('METHODIST does not see integration in admin group', () => {
    const entries = getAvailableNavEntries(UserRole.METHODIST)
    const adminGroup = entries.find((e) => e.nameKey === 'adminGroup') as NavGroup
    expect(adminGroup).toBeDefined()
    const adminItemKeys = adminGroup.items.map((i) => i.nameKey)
    expect(adminItemKeys).toContain('usersCatalog')
    expect(adminItemKeys).not.toContain('integration')
  })

  it('ACADEMIC_SECRETARY does not see integration', () => {
    const entries = getAvailableNavEntries(UserRole.ACADEMIC_SECRETARY)
    const adminGroup = entries.find((e) => e.nameKey === 'adminGroup') as NavGroup
    expect(adminGroup).toBeDefined()
    const adminItemKeys = adminGroup.items.map((i) => i.nameKey)
    expect(adminItemKeys).toContain('usersCatalog')
    expect(adminItemKeys).not.toContain('integration')
  })

  // v0.158.0: curriculumApprove visible to methodist + system_admin
  // (approvers). Hidden from academic_secretary (author) + teacher + student.
  it.each([UserRole.ACADEMIC_SECRETARY, UserRole.TEACHER, UserRole.STUDENT])(
    'curriculumApprove entry is hidden from %s',
    (role) => {
      const entries = getAvailableNavEntries(role)
      const adminGroup = entries.find((e) => e.nameKey === 'adminGroup')
      // Either adminGroup is filtered out entirely (student) или present but
      // without curriculumApprove (other non-approver roles).
      if (adminGroup && isNavGroup(adminGroup)) {
        const itemKeys = adminGroup.items.map((i) => i.nameKey)
        expect(itemKeys).not.toContain('curriculumApprove')
      }
      // Flat-items helper also confirms — single source of truth.
      const flatKeys = getAvailableNavItems(role).map((i) => i.nameKey)
      expect(flatKeys).not.toContain('curriculumApprove')
    }
  )

  it('curriculumApprove entry is visible to SYSTEM_ADMIN', () => {
    const flatKeys = getAvailableNavItems(UserRole.SYSTEM_ADMIN).map((i) => i.nameKey)
    expect(flatKeys).toContain('curriculumApprove')
  })

  it('accepts string role', () => {
    const entries = getAvailableNavEntries('student')
    expect(entries.length).toBeGreaterThan(0)
  })

  it('keeps documents group for student (documents + files, no templates)', () => {
    const entries = getAvailableNavEntries(UserRole.STUDENT)
    const docsGroup = entries.find((e) => e.nameKey === 'documentsGroup')
    expect(docsGroup).toBeDefined()
    expect(isNavGroup(docsGroup!)).toBe(true)
    const itemKeys = (docsGroup as NavGroup).items.map((i) => i.nameKey)
    expect(itemKeys).toContain('documents')
    expect(itemKeys).toContain('files')
    expect(itemKeys).not.toContain('templates')
  })
})

describe('getAvailableNavItems (legacy)', () => {
  it('returns empty array when no role provided', () => {
    expect(getAvailableNavItems()).toEqual([])
    expect(getAvailableNavItems(undefined)).toEqual([])
  })

  it('returns flat list of items for SYSTEM_ADMIN', () => {
    const items = getAvailableNavItems(UserRole.SYSTEM_ADMIN)
    const itemKeys = items.map((i) => i.nameKey)
    expect(itemKeys).toContain('dashboard')
    expect(itemKeys).toContain('documents')
    expect(itemKeys).toContain('templates')
    expect(itemKeys).toContain('reports')
    expect(itemKeys).toContain('analytics')
    expect(itemKeys).toContain('schedule')
    expect(itemKeys).toContain('calendar')
    expect(itemKeys).toContain('tasks')
    expect(itemKeys).toContain('announcements')
    expect(itemKeys).toContain('messages')
    expect(itemKeys).toContain('aiAssistant')
    expect(itemKeys).toContain('usersCatalog')
    expect(itemKeys).toContain('integration')
  })

  it('returns flat list of items for STUDENT', () => {
    const items = getAvailableNavItems(UserRole.STUDENT)
    const itemKeys = items.map((i) => i.nameKey)

    expect(itemKeys).toContain('dashboard')
    expect(itemKeys).toContain('documents')
    expect(itemKeys).toContain('files')
    expect(itemKeys).toContain('schedule')
    expect(itemKeys).toContain('calendar')
    expect(itemKeys).toContain('tasks')
    expect(itemKeys).toContain('announcements')
    expect(itemKeys).toContain('messages')
    expect(itemKeys).toContain('aiAssistant')

    expect(itemKeys).not.toContain('templates')
    expect(itemKeys).not.toContain('reports')
    expect(itemKeys).not.toContain('analytics')
    expect(itemKeys).not.toContain('usersCatalog')
    expect(itemKeys).not.toContain('integration')
  })

  it('all returned items have url property', () => {
    const items = getAvailableNavItems(UserRole.SYSTEM_ADMIN)
    items.forEach((item) => {
      expect(item.url).toBeDefined()
      expect(typeof item.url).toBe('string')
    })
  })
})
