import { act, renderHook } from '@testing-library/react'
import { useAppearanceStore, useAppearanceHydrated } from '../appearanceStore'

describe('appearanceStore', () => {
  beforeEach(() => {
    // Reset store to defaults before each test
    const { result } = renderHook(() => useAppearanceStore())
    act(() => {
      result.current.resetToDefaults()
    })
  })

  describe('initial state', () => {
    it('has default background config', () => {
      const { result } = renderHook(() => useAppearanceStore())

      expect(result.current.background).toEqual({
        type: 'grain-gradient',
        enabled: true,
        speed: 1,
        intensity: 0.45,
      })
    })

    it('has reducedMotion set to false by default', () => {
      const { result } = renderHook(() => useAppearanceStore())

      expect(result.current.reducedMotion).toBe(false)
    })
  })

  describe('setBackgroundType', () => {
    it('updates background type', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setBackgroundType('warp')
      })

      expect(result.current.background.type).toBe('warp')
    })

    it('can set to none', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setBackgroundType('none')
      })

      expect(result.current.background.type).toBe('none')
    })

    it('can set to mesh-gradient', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setBackgroundType('mesh-gradient')
      })

      expect(result.current.background.type).toBe('mesh-gradient')
    })
  })

  describe('setBackgroundEnabled', () => {
    it('disables background', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setBackgroundEnabled(false)
      })

      expect(result.current.background.enabled).toBe(false)
    })

    it('enables background', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setBackgroundEnabled(false)
        result.current.setBackgroundEnabled(true)
      })

      expect(result.current.background.enabled).toBe(true)
    })
  })

  describe('setBackgroundSpeed', () => {
    it('sets speed within valid range', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setBackgroundSpeed(1.5)
      })

      expect(result.current.background.speed).toBe(1.5)
    })

    it('clamps speed to minimum 0.1', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setBackgroundSpeed(0)
      })

      expect(result.current.background.speed).toBe(0.1)
    })

    it('clamps speed to maximum 2', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setBackgroundSpeed(5)
      })

      expect(result.current.background.speed).toBe(2)
    })
  })

  describe('setBackgroundIntensity', () => {
    it('sets intensity within valid range', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setBackgroundIntensity(0.7)
      })

      expect(result.current.background.intensity).toBe(0.7)
    })

    it('clamps intensity to minimum 0.1', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setBackgroundIntensity(0)
      })

      expect(result.current.background.intensity).toBe(0.1)
    })

    it('clamps intensity to maximum 1', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setBackgroundIntensity(2)
      })

      expect(result.current.background.intensity).toBe(1)
    })
  })

  describe('setReducedMotion', () => {
    it('sets reduced motion to true', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setReducedMotion(true)
      })

      expect(result.current.reducedMotion).toBe(true)
    })

    it('sets reduced motion to false', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setReducedMotion(true)
        result.current.setReducedMotion(false)
      })

      expect(result.current.reducedMotion).toBe(false)
    })
  })

  describe('resetToDefaults', () => {
    it('resets all settings to defaults', () => {
      const { result } = renderHook(() => useAppearanceStore())

      // Change some settings
      act(() => {
        result.current.setBackgroundType('warp')
        result.current.setBackgroundEnabled(false)
        result.current.setBackgroundSpeed(2)
        result.current.setBackgroundIntensity(0.8)
        result.current.setReducedMotion(true)
      })

      // Reset
      act(() => {
        result.current.resetToDefaults()
      })

      expect(result.current.background).toEqual({
        type: 'grain-gradient',
        enabled: true,
        speed: 1,
        intensity: 0.45,
      })
      expect(result.current.reducedMotion).toBe(false)
    })
  })

  describe('setHasHydrated', () => {
    it('sets hydration state', () => {
      const { result } = renderHook(() => useAppearanceStore())

      act(() => {
        result.current.setHasHydrated(true)
      })

      expect(result.current._hasHydrated).toBe(true)
    })
  })
})

describe('useAppearanceHydrated', () => {
  it('returns hydration state', () => {
    const { result: storeResult } = renderHook(() => useAppearanceStore())
    const { result: _hydratedResult } = renderHook(() => useAppearanceHydrated())

    act(() => {
      storeResult.current.setHasHydrated(true)
    })

    // Re-render to get updated value
    const { result: updatedResult } = renderHook(() => useAppearanceHydrated())
    expect(updatedResult.current).toBe(true)
  })
})
