import { renderHook, act } from '@testing-library/react'
import { useMediaQuery, useIsMobile, useIsTablet, useIsDesktop } from '../use-media-query'

describe('useMediaQuery', () => {
  let listeners: Map<string, (event: MediaQueryListEvent) => void>
  let matchesMap: Map<string, boolean>

  beforeEach(() => {
    listeners = new Map()
    matchesMap = new Map()

    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation((query: string) => ({
        matches: matchesMap.get(query) || false,
        media: query,
        onchange: null,
        addListener: jest.fn(),
        removeListener: jest.fn(),
        addEventListener: jest.fn((_, callback) => {
          listeners.set(query, callback)
        }),
        removeEventListener: jest.fn((_, callback) => {
          if (listeners.get(query) === callback) {
            listeners.delete(query)
          }
        }),
        dispatchEvent: jest.fn(),
      })),
    })
  })

  afterEach(() => {
    listeners.clear()
    matchesMap.clear()
  })

  it('returns false initially when query does not match', () => {
    const { result } = renderHook(() => useMediaQuery('(min-width: 768px)'))
    expect(result.current).toBe(false)
  })

  it('returns true when query matches', () => {
    matchesMap.set('(min-width: 768px)', true)
    const { result } = renderHook(() => useMediaQuery('(min-width: 768px)'))
    expect(result.current).toBe(true)
  })

  it('updates value when media query changes', () => {
    const { result } = renderHook(() => useMediaQuery('(min-width: 768px)'))
    expect(result.current).toBe(false)

    act(() => {
      const listener = listeners.get('(min-width: 768px)')
      if (listener) {
        listener({ matches: true } as MediaQueryListEvent)
      }
    })

    expect(result.current).toBe(true)
  })

  it('removes listener on unmount', () => {
    const { unmount } = renderHook(() => useMediaQuery('(min-width: 768px)'))
    expect(listeners.has('(min-width: 768px)')).toBe(true)

    unmount()
    expect(listeners.has('(min-width: 768px)')).toBe(false)
  })
})

describe('useIsMobile', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation((query: string) => ({
        matches: query === '(max-width: 768px)',
        media: query,
        onchange: null,
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
      })),
    })
  })

  it('returns true for mobile viewport', () => {
    const { result } = renderHook(() => useIsMobile())
    expect(result.current).toBe(true)
  })
})

describe('useIsTablet', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation((query: string) => ({
        matches: query === '(min-width: 769px) and (max-width: 1024px)',
        media: query,
        onchange: null,
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
      })),
    })
  })

  it('returns true for tablet viewport', () => {
    const { result } = renderHook(() => useIsTablet())
    expect(result.current).toBe(true)
  })
})

describe('useIsDesktop', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation((query: string) => ({
        matches: query === '(min-width: 1025px)',
        media: query,
        onchange: null,
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
      })),
    })
  })

  it('returns true for desktop viewport', () => {
    const { result } = renderHook(() => useIsDesktop())
    expect(result.current).toBe(true)
  })
})
