import { renderHook, act } from '@testing-library/react'
import { useTheme } from '../use-theme'
import { useTheme as useNextTheme } from 'next-themes'

// Mock next-themes
jest.mock('next-themes')

const mockUseNextTheme = useNextTheme as jest.MockedFunction<typeof useNextTheme>

describe('useTheme', () => {
  const mockSetTheme = jest.fn()

  beforeEach(() => {
    mockUseNextTheme.mockReturnValue({
      theme: 'light',
      setTheme: mockSetTheme,
      resolvedTheme: 'light',
      systemTheme: 'light',
      themes: ['light', 'dark', 'system'],
      forcedTheme: undefined,
    })
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('returns theme values', () => {
    const { result } = renderHook(() => useTheme())

    expect(result.current.theme).toBe('light')
    expect(result.current.resolvedTheme).toBe('light')
    expect(result.current.systemTheme).toBe('light')
  })

  it('returns isDark as false when theme is light', () => {
    const { result } = renderHook(() => useTheme())

    expect(result.current.isDark).toBe(false)
    expect(result.current.isLight).toBe(true)
  })

  it('returns isDark as true when theme is dark', () => {
    mockUseNextTheme.mockReturnValue({
      theme: 'dark',
      setTheme: mockSetTheme,
      resolvedTheme: 'dark',
      systemTheme: 'light',
      themes: ['light', 'dark', 'system'],
      forcedTheme: undefined,
    })

    const { result } = renderHook(() => useTheme())

    expect(result.current.isDark).toBe(true)
    expect(result.current.isLight).toBe(false)
  })

  it('provides setTheme function', () => {
    const { result } = renderHook(() => useTheme())

    act(() => {
      result.current.setTheme('dark')
    })

    expect(mockSetTheme).toHaveBeenCalledWith('dark')
  })

  it('toggles theme from light to dark', () => {
    mockUseNextTheme.mockReturnValue({
      theme: 'light',
      setTheme: mockSetTheme,
      resolvedTheme: 'light',
      systemTheme: 'light',
      themes: ['light', 'dark', 'system'],
      forcedTheme: undefined,
    })

    const { result } = renderHook(() => useTheme())

    act(() => {
      result.current.toggleTheme()
    })

    expect(mockSetTheme).toHaveBeenCalledWith('dark')
  })

  it('toggles theme from dark to light', () => {
    mockUseNextTheme.mockReturnValue({
      theme: 'dark',
      setTheme: mockSetTheme,
      resolvedTheme: 'dark',
      systemTheme: 'light',
      themes: ['light', 'dark', 'system'],
      forcedTheme: undefined,
    })

    const { result } = renderHook(() => useTheme())

    act(() => {
      result.current.toggleTheme()
    })

    expect(mockSetTheme).toHaveBeenCalledWith('light')
  })
})
