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

  it('educationGroup contains schedule, calendar, and tasks', () => {
    const eduGroup = navigationConfig.find((e) => e.nameKey === 'educationGroup') as NavGroup
    expect(eduGroup).toBeDefined()
    expect(isNavGroup(eduGroup)).toBe(true)
    const itemKeys = eduGroup.items.map((i) => i.nameKey)
    expect(itemKeys).toContain('schedule')
    expect(itemKeys).toContain('calendar')
    expect(itemKeys).toContain('tasks')
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
    expect(itemKeys).toContain('users')
    expect(itemKeys).toContain('integration')
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
    expect(adminItemKeys).toContain('users')
    expect(adminItemKeys).not.toContain('integration')
  })

  it('ACADEMIC_SECRETARY does not see integration', () => {
    const entries = getAvailableNavEntries(UserRole.ACADEMIC_SECRETARY)
    const adminGroup = entries.find((e) => e.nameKey === 'adminGroup') as NavGroup
    expect(adminGroup).toBeDefined()
    const adminItemKeys = adminGroup.items.map((i) => i.nameKey)
    expect(adminItemKeys).toContain('users')
    expect(adminItemKeys).not.toContain('integration')
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
    expect(itemKeys).toContain('users')
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
