import { navigationConfig, getAvailableNavItems } from '../navigation'
import { UserRole } from '@/types/auth'

describe('navigationConfig', () => {
  it('contains dashboard item', () => {
    const dashboardItem = navigationConfig.find((item) => item.nameKey === 'dashboard')
    expect(dashboardItem).toBeDefined()
    expect(dashboardItem?.url).toBe('/dashboard')
  })

  it('contains all expected navigation items', () => {
    const expectedKeys = [
      'dashboard',
      'users',
      'documents',
      'reports',
      'calendar',
      'messages',
      'integration',
    ]
    const actualKeys = navigationConfig.map((item) => item.nameKey)
    expect(actualKeys).toEqual(expectedKeys)
  })

  it('all items have required properties', () => {
    navigationConfig.forEach((item) => {
      expect(item.nameKey).toBeDefined()
      expect(item.url).toBeDefined()
      expect(item.icon).toBeDefined()
      expect(typeof item.nameKey).toBe('string')
      expect(typeof item.url).toBe('string')
    })
  })
})

describe('getAvailableNavItems', () => {
  it('returns empty array when no role provided', () => {
    expect(getAvailableNavItems()).toEqual([])
    expect(getAvailableNavItems(undefined)).toEqual([])
  })

  it('returns items without role restrictions for any authenticated user', () => {
    const items = getAvailableNavItems(UserRole.STUDENT)

    // Dashboard and messages should be available to all
    const dashboardItem = items.find((item) => item.nameKey === 'dashboard')
    const messagesItem = items.find((item) => item.nameKey === 'messages')

    expect(dashboardItem).toBeDefined()
    expect(messagesItem).toBeDefined()
  })

  it('returns correct items for SYSTEM_ADMIN role', () => {
    const items = getAvailableNavItems(UserRole.SYSTEM_ADMIN)

    // Admin should have access to all items
    expect(items.length).toBe(navigationConfig.length)
  })

  it('returns correct items for STUDENT role', () => {
    const items = getAvailableNavItems(UserRole.STUDENT)
    const itemKeys = items.map((item) => item.nameKey)

    // Student should have access to dashboard, documents, calendar, messages
    expect(itemKeys).toContain('dashboard')
    expect(itemKeys).toContain('documents')
    expect(itemKeys).toContain('calendar')
    expect(itemKeys).toContain('messages')

    // Student should NOT have access to users, reports, integration
    expect(itemKeys).not.toContain('users')
    expect(itemKeys).not.toContain('reports')
    expect(itemKeys).not.toContain('integration')
  })

  it('returns correct items for TEACHER role', () => {
    const items = getAvailableNavItems(UserRole.TEACHER)
    const itemKeys = items.map((item) => item.nameKey)

    // Teacher should have access to dashboard, users, documents, calendar, messages
    expect(itemKeys).toContain('dashboard')
    expect(itemKeys).toContain('users')
    expect(itemKeys).toContain('documents')
    expect(itemKeys).toContain('calendar')
    expect(itemKeys).toContain('messages')

    // Teacher should NOT have access to reports, integration
    expect(itemKeys).not.toContain('reports')
    expect(itemKeys).not.toContain('integration')
  })

  it('returns correct items for METHODIST role', () => {
    const items = getAvailableNavItems(UserRole.METHODIST)
    const itemKeys = items.map((item) => item.nameKey)

    // Methodist should have access to all items
    expect(itemKeys).toContain('dashboard')
    expect(itemKeys).toContain('users')
    expect(itemKeys).toContain('documents')
    expect(itemKeys).toContain('reports')
    expect(itemKeys).toContain('calendar')
    expect(itemKeys).toContain('messages')
    expect(itemKeys).toContain('integration')
  })

  it('returns correct items for ACADEMIC_SECRETARY role', () => {
    const items = getAvailableNavItems(UserRole.ACADEMIC_SECRETARY)
    const itemKeys = items.map((item) => item.nameKey)

    // Academic Secretary should have access to most items
    expect(itemKeys).toContain('dashboard')
    expect(itemKeys).toContain('users')
    expect(itemKeys).toContain('documents')
    expect(itemKeys).toContain('reports')
    expect(itemKeys).toContain('calendar')
    expect(itemKeys).toContain('messages')

    // Academic Secretary should NOT have access to integration
    expect(itemKeys).not.toContain('integration')
  })

  it('accepts string role', () => {
    const items = getAvailableNavItems('student')
    expect(items.length).toBeGreaterThan(0)
  })
})
