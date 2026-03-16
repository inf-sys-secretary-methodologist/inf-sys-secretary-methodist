'use client'

import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface VoiceState {
  // Settings
  autoSubmit: boolean
  autoRead: boolean
  voiceMode: boolean
  preferredVoiceURI: string

  // Actions
  setAutoSubmit: (value: boolean) => void
  setAutoRead: (value: boolean) => void
  setVoiceMode: (value: boolean) => void
  setPreferredVoiceURI: (uri: string) => void
}

export const useVoiceStore = create<VoiceState>()(
  persist(
    (set) => ({
      autoSubmit: true,
      autoRead: false,
      voiceMode: false,
      preferredVoiceURI: '',

      setAutoSubmit: (autoSubmit) => set({ autoSubmit }),
      setAutoRead: (autoRead) => set({ autoRead }),
      setVoiceMode: (voiceMode) => set({ voiceMode }),
      setPreferredVoiceURI: (preferredVoiceURI) => set({ preferredVoiceURI }),
    }),
    {
      name: 'voice-settings',
    }
  )
)
