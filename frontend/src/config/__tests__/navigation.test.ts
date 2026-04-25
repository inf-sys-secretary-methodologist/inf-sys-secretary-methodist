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
      'calendar',
      'tasks',
      'messages',
      'aiAssistant',
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

  it('documentsGroup contains documents and templates', () => {
    const docsGroup = navigationConfig.find((e) => e.nameKey === 'documentsGroup') as NavGroup
    expect(docsGroup).toBeDefined()
    expect(isNavGroup(docsGroup)).toBe(true)

    const itemKeys = docsGroup.items.map((i) => i.nameKey)
    expect(itemKeys).toContain('documents')
    expect(itemKeys).toContain('templates')
  })

  it('analyticsGroup contains reports and analytics', () => {
    const analyticsGroup = navigationConfig.find((e) => e.nameKey === 'analyticsGroup') as NavGroup
    expect(analyticsGroup).toBeDefined()
    expect(isNavGroup(analyticsGroup)).toBe(true)

    const itemKeys = analyticsGroup.items.map((i) => i.nameKey)
    expect(itemKeys).toContain('reports')
    expect(itemKeys).toContain('analytics')
  })

  it('adminGroup contains users and integration', () => {
    const adminGroup = navigationConfig.find((e) => e.nameKey === 'adminGroup') as NavGroup
    expect(adminGroup).toBeDefined()
    expect(isNavGroup(adminGroup)).toBe(true)

    const itemKeys = adminGroup.items.map((i) => i.nameKey)
    expect(itemKeys).toContain('users')
    expect(itemKeys).toContain('integration')
  })

  it('contains announcements entry pointing to /announcements for all authenticated users', () => {
    const annEntry = navigationConfig.find((e) => e.nameKey === 'announcements')
    expect(annEntry).toBeDefined()
    expect(isNavGroup(annEntry!)).toBe(false)
    if (annEntry && !isNavGroup(annEntry)) {
      expect(annEntry.url).toBe('/announcements')
      // Visible to all roles (including STUDENT in view-only mode).
      expect(annEntry.roles).toEqual(
        expect.arrayContaining([
          UserRole.SYSTEM_ADMIN,
          UserRole.METHODIST,
          UserRole.ACADEMIC_SECRETARY,
          UserRole.TEACHER,
          UserRole.STUDENT,
        ])
      )
    }
  })

  it('contains tasks entry pointing to /tasks for privileged roles', () => {
    const tasksEntry = navigationConfig.find((e) => e.nameKey === 'tasks')
    expect(tasksEntry).toBeDefined()
    expect(isNavGroup(tasksEntry!)).toBe(false)
    if (tasksEntry && !isNavGroup(tasksEntry)) {
      expect(tasksEntry.url).toBe('/tasks')
      expect(tasksEntry.roles).toEqual(
        expect.arrayContaining([
          UserRole.SYSTEM_ADMIN,
          UserRole.METHODIST,
          UserRole.ACADEMIC_SECRETARY,
        ])
      )
      // Students should not see tasks (matches route-config.ts)
      expect(tasksEntry.roles).not.toContain(UserRole.STUDENT)
    }
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

  it('returns entries without role restrictions for any authenticated user', () => {
    const entries = getAvailableNavEntries(UserRole.STUDENT)

    // Dashboard, messages, and aiAssistant should be available to all
    const dashboardEntry = entries.find((e) => e.nameKey === 'dashboard')
    const messagesEntry = entries.find((e) => e.nameKey === 'messages')
    const aiAssistantEntry = entries.find((e) => e.nameKey === 'aiAssistant')

    expect(dashboardEntry).toBeDefined()
    expect(messagesEntry).toBeDefined()
    expect(aiAssistantEntry).toBeDefined()
  })

  it('returns all entries for SYSTEM_ADMIN role', () => {
    const entries = getAvailableNavEntries(UserRole.SYSTEM_ADMIN)

    // Admin should have access to all top-level entries
    expect(entries.length).toBe(navigationConfig.length)
  })

  it('returns correct entries for STUDENT role', () => {
    const entries = getAvailableNavEntries(UserRole.STUDENT)
    const entryKeys = entries.map((e) => e.nameKey)

    // Student should have dashboard, documents group (only documents item), calendar, messages, aiAssistant
    expect(entryKeys).toContain('dashboard')
    expect(entryKeys).toContain('calendar')
    expect(entryKeys).toContain('messages')
    expect(entryKeys).toContain('aiAssistant')

    // Documents group should be flattened to single item since student only has access to documents
    const docsEntry = entries.find((e) => e.nameKey === 'documents')
    expect(docsEntry).toBeDefined()

    // Student should NOT have access to analytics group or admin group
    expect(entryKeys).not.toContain('analyticsGroup')
    expect(entryKeys).not.toContain('adminGroup')
    expect(entryKeys).not.toContain('users')
    expect(entryKeys).not.toContain('integration')
  })

  it('returns correct entries for TEACHER role', () => {
    const entries = getAvailableNavEntries(UserRole.TEACHER)
    const entryKeys = entries.map((e) => e.nameKey)

    // Teacher should have dashboard, documents group, calendar, messages, aiAssistant, and users from admin
    expect(entryKeys).toContain('dashboard')
    expect(entryKeys).toContain('calendar')
    expect(entryKeys).toContain('messages')
    expect(entryKeys).toContain('aiAssistant')

    // Teacher should have access to users (inside admin group) but not integration.
    // Admin group also contains settingsPage (no role restriction → visible to all),
    // so the group has 2+ visible items and is NOT flattened. Look inside adminGroup.
    const adminGroup = entries.find((e) => e.nameKey === 'adminGroup') as NavGroup
    expect(adminGroup).toBeDefined()
    expect(isNavGroup(adminGroup)).toBe(true)
    const adminItemKeys = adminGroup.items.map((i) => i.nameKey)
    expect(adminItemKeys).toContain('users')
    expect(adminItemKeys).not.toContain('integration')

    // Teacher should NOT have access to analytics group
    expect(entryKeys).not.toContain('analyticsGroup')
    expect(entryKeys).not.toContain('reports')
    expect(entryKeys).not.toContain('analytics')
  })

  it('returns correct entries for METHODIST role', () => {
    const entries = getAvailableNavEntries(UserRole.METHODIST)
    const entryKeys = entries.map((e) => e.nameKey)

    // Methodist should have access to all entries
    expect(entryKeys).toContain('dashboard')
    expect(entryKeys).toContain('documentsGroup')
    expect(entryKeys).toContain('analyticsGroup')
    expect(entryKeys).toContain('calendar')
    expect(entryKeys).toContain('messages')
    expect(entryKeys).toContain('aiAssistant')
    expect(entryKeys).toContain('adminGroup')
  })

  it('returns correct entries for ACADEMIC_SECRETARY role', () => {
    const entries = getAvailableNavEntries(UserRole.ACADEMIC_SECRETARY)
    const entryKeys = entries.map((e) => e.nameKey)

    // Academic Secretary should have most entries
    expect(entryKeys).toContain('dashboard')
    expect(entryKeys).toContain('documentsGroup')
    expect(entryKeys).toContain('analyticsGroup')
    expect(entryKeys).toContain('calendar')
    expect(entryKeys).toContain('messages')
    expect(entryKeys).toContain('aiAssistant')

    // Academic Secretary should have users but not integration.
    // Admin group also contains settingsPage (no role restriction), so the group
    // has 2+ visible items and is NOT flattened. Look inside adminGroup.
    const adminGroup = entries.find((e) => e.nameKey === 'adminGroup') as NavGroup
    expect(adminGroup).toBeDefined()
    expect(isNavGroup(adminGroup)).toBe(true)
    const adminItemKeys = adminGroup.items.map((i) => i.nameKey)
    expect(adminItemKeys).toContain('users')
    expect(adminItemKeys).not.toContain('integration')
  })

  it('accepts string role', () => {
    const entries = getAvailableNavEntries('student')
    expect(entries.length).toBeGreaterThan(0)
  })

  it('flattens groups with single item', () => {
    // For student, documents group has only documents (not templates)
    // So it should return the item directly, not the group
    const entries = getAvailableNavEntries(UserRole.STUDENT)

    const docsEntry = entries.find((e) => e.nameKey === 'documents')
    expect(docsEntry).toBeDefined()
    expect(isNavGroup(docsEntry!)).toBe(false)
  })
})

describe('getAvailableNavItems (legacy)', () => {
  it('returns empty array when no role provided', () => {
    expect(getAvailableNavItems()).toEqual([])
    expect(getAvailableNavItems(undefined)).toEqual([])
  })

  it('returns flat list of items for SYSTEM_ADMIN', () => {
    const items = getAvailableNavItems(UserRole.SYSTEM_ADMIN)

    // Should include all items from all groups
    const itemKeys = items.map((i) => i.nameKey)
    expect(itemKeys).toContain('dashboard')
    expect(itemKeys).toContain('documents')
    expect(itemKeys).toContain('templates')
    expect(itemKeys).toContain('reports')
    expect(itemKeys).toContain('analytics')
    expect(itemKeys).toContain('calendar')
    expect(itemKeys).toContain('messages')
    expect(itemKeys).toContain('aiAssistant')
    expect(itemKeys).toContain('users')
    expect(itemKeys).toContain('integration')
  })

  it('returns flat list of items for STUDENT', () => {
    const items = getAvailableNavItems(UserRole.STUDENT)
    const itemKeys = items.map((i) => i.nameKey)

    expect(itemKeys).toContain('dashboard')
    expect(itemKeys).toContain('documents')
    expect(itemKeys).toContain('calendar')
    expect(itemKeys).toContain('messages')
    expect(itemKeys).toContain('aiAssistant')

    expect(itemKeys).not.toContain('templates')
    expect(itemKeys).not.toContain('reports')
    expect(itemKeys).not.toContain('analytics')
    expect(itemKeys).not.toContain('users')
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
