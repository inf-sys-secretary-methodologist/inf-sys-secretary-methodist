import { renderHook } from '@testing-library/react'
import { useRouteAnnouncer } from '../useRouteAnnouncer'

// Mock next/navigation
const mockUsePathname = jest.fn()
jest.mock('next/navigation', () => ({
  usePathname: () => mockUsePathname(),
}))

// Mock next-intl
const mockTranslations = jest.fn()
jest.mock('next-intl', () => ({
  useTranslations: () => mockTranslations,
}))

// Mock the announcer
const mockAnnounce = jest.fn()
jest.mock('@/components/ui/screen-reader-announcer', () => ({
  useAnnouncer: () => ({ announce: mockAnnounce }),
}))

describe('useRouteAnnouncer', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockUsePathname.mockReturnValue('/')
    mockTranslations.mockImplementation((key: string, params?: { page: string }) => {
      if (key === 'navigatedTo') return `Navigated to ${params?.page}`
      if (key.startsWith('routes.')) return key.replace('routes.', '').toUpperCase()
      return key
    })
  })

  it('does not announce on initial mount', () => {
    renderHook(() => useRouteAnnouncer())

    expect(mockAnnounce).not.toHaveBeenCalled()
  })

  it('announces navigation when pathname changes', () => {
    mockUsePathname.mockReturnValue('/')

    const { rerender } = renderHook(() => useRouteAnnouncer())

    // Change pathname
    mockUsePathname.mockReturnValue('/dashboard')
    rerender()

    expect(mockAnnounce).toHaveBeenCalledWith('Navigated to DASHBOARD')
  })

  it('does not announce when pathname stays the same', () => {
    mockUsePathname.mockReturnValue('/dashboard')

    const { rerender } = renderHook(() => useRouteAnnouncer())

    // Rerender with same path
    rerender()

    expect(mockAnnounce).not.toHaveBeenCalled()
  })

  it('uses route keys for known routes', () => {
    mockUsePathname.mockReturnValue('/')

    const { rerender } = renderHook(() => useRouteAnnouncer())

    // Navigate to documents
    mockUsePathname.mockReturnValue('/documents')
    rerender()

    expect(mockTranslations).toHaveBeenCalledWith('routes.documents')
  })

  it('generates route name for unknown routes', () => {
    mockUsePathname.mockReturnValue('/')

    const { rerender } = renderHook(() => useRouteAnnouncer())

    // Navigate to unknown route
    mockUsePathname.mockReturnValue('/some-unknown-route')
    rerender()

    // Should generate name from pathname
    expect(mockAnnounce).toHaveBeenCalled()
  })

  it('handles dynamic routes with [id] pattern', () => {
    mockUsePathname.mockReturnValue('/')

    const { rerender } = renderHook(() => useRouteAnnouncer())

    // Navigate to dynamic route
    mockUsePathname.mockReturnValue('/users/[id]')
    rerender()

    expect(mockAnnounce).toHaveBeenCalledWith('Navigated to Details')
  })

  it('handles nested routes', () => {
    mockUsePathname.mockReturnValue('/')

    const { rerender } = renderHook(() => useRouteAnnouncer())

    // Navigate to nested route
    mockUsePathname.mockReturnValue('/settings/appearance')
    rerender()

    expect(mockTranslations).toHaveBeenCalledWith('routes.appearance')
  })

  it('handles documents/shared route', () => {
    mockUsePathname.mockReturnValue('/')

    const { rerender } = renderHook(() => useRouteAnnouncer())

    mockUsePathname.mockReturnValue('/documents/shared')
    rerender()

    expect(mockTranslations).toHaveBeenCalledWith('routes.sharedDocuments')
  })

  it('converts kebab-case route names to title case', () => {
    mockUsePathname.mockReturnValue('/')

    const { rerender } = renderHook(() => useRouteAnnouncer())

    // Navigate to kebab-case route not in routeKeys
    mockUsePathname.mockReturnValue('/my-custom-route')
    rerender()

    expect(mockAnnounce).toHaveBeenCalledWith('Navigated to My Custom Route')
  })

  it('announces multiple navigation changes correctly', () => {
    mockUsePathname.mockReturnValue('/')

    const { rerender } = renderHook(() => useRouteAnnouncer())

    // First navigation
    mockUsePathname.mockReturnValue('/dashboard')
    rerender()
    expect(mockAnnounce).toHaveBeenCalledTimes(1)

    // Second navigation
    mockUsePathname.mockReturnValue('/calendar')
    rerender()
    expect(mockAnnounce).toHaveBeenCalledTimes(2)

    // Third navigation
    mockUsePathname.mockReturnValue('/notifications')
    rerender()
    expect(mockAnnounce).toHaveBeenCalledTimes(3)
  })
})
