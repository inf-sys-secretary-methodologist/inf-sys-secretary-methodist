'use client'

import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type BackgroundType = 'none' | 'grain-gradient' | 'warp' | 'mesh-gradient'

export interface BackgroundConfig {
  type: BackgroundType
  enabled: boolean
  speed: number // 0.1 - 2.0
  intensity: number // 0.1 - 1.0
}

interface AppearanceState {
  // State
  background: BackgroundConfig
  reducedMotion: boolean

  // Actions
  setBackgroundType: (type: BackgroundType) => void
  setBackgroundEnabled: (enabled: boolean) => void
  setBackgroundSpeed: (speed: number) => void
  setBackgroundIntensity: (intensity: number) => void
  setReducedMotion: (reduced: boolean) => void
  resetToDefaults: () => void
}

const defaultBackground: BackgroundConfig = {
  type: 'grain-gradient',
  enabled: true,
  speed: 1,
  intensity: 0.45,
}

export const useAppearanceStore = create<AppearanceState>()(
  persist(
    (set) => ({
      // Initial state
      background: defaultBackground,
      reducedMotion: false,

      // Actions
      setBackgroundType: (type) =>
        set((state) => ({
          background: { ...state.background, type },
        })),

      setBackgroundEnabled: (enabled) =>
        set((state) => ({
          background: { ...state.background, enabled },
        })),

      setBackgroundSpeed: (speed) =>
        set((state) => ({
          background: { ...state.background, speed: Math.max(0.1, Math.min(2, speed)) },
        })),

      setBackgroundIntensity: (intensity) =>
        set((state) => ({
          background: { ...state.background, intensity: Math.max(0.1, Math.min(1, intensity)) },
        })),

      setReducedMotion: (reducedMotion) => set({ reducedMotion }),

      resetToDefaults: () =>
        set({
          background: defaultBackground,
          reducedMotion: false,
        }),
    }),
    {
      name: 'appearance-settings',
      partialize: (state) => ({
        background: state.background,
        reducedMotion: state.reducedMotion,
      }),
    }
  )
)
