import { act, renderHook } from '@testing-library/react'
import { useVoiceStore } from '../voiceStore'

describe('voiceStore', () => {
  beforeEach(() => {
    // Reset store to defaults before each test
    const { result } = renderHook(() => useVoiceStore())
    act(() => {
      result.current.setAutoSubmit(true)
      result.current.setAutoRead(false)
      result.current.setVoiceMode(false)
      result.current.setPreferredVoiceURI('')
    })
  })

  describe('initial state', () => {
    it('has default values', () => {
      const { result } = renderHook(() => useVoiceStore())

      expect(result.current.autoSubmit).toBe(true)
      expect(result.current.autoRead).toBe(false)
      expect(result.current.voiceMode).toBe(false)
      expect(result.current.preferredVoiceURI).toBe('')
    })
  })

  describe('setAutoSubmit', () => {
    it('toggles autoSubmit to false', () => {
      const { result } = renderHook(() => useVoiceStore())

      act(() => {
        result.current.setAutoSubmit(false)
      })

      expect(result.current.autoSubmit).toBe(false)
    })

    it('toggles autoSubmit to true', () => {
      const { result } = renderHook(() => useVoiceStore())

      act(() => {
        result.current.setAutoSubmit(false)
        result.current.setAutoSubmit(true)
      })

      expect(result.current.autoSubmit).toBe(true)
    })
  })

  describe('setAutoRead', () => {
    it('toggles autoRead to true', () => {
      const { result } = renderHook(() => useVoiceStore())

      act(() => {
        result.current.setAutoRead(true)
      })

      expect(result.current.autoRead).toBe(true)
    })

    it('toggles autoRead to false', () => {
      const { result } = renderHook(() => useVoiceStore())

      act(() => {
        result.current.setAutoRead(true)
        result.current.setAutoRead(false)
      })

      expect(result.current.autoRead).toBe(false)
    })
  })

  describe('setVoiceMode', () => {
    it('toggles voiceMode to true', () => {
      const { result } = renderHook(() => useVoiceStore())

      act(() => {
        result.current.setVoiceMode(true)
      })

      expect(result.current.voiceMode).toBe(true)
    })

    it('toggles voiceMode to false', () => {
      const { result } = renderHook(() => useVoiceStore())

      act(() => {
        result.current.setVoiceMode(true)
        result.current.setVoiceMode(false)
      })

      expect(result.current.voiceMode).toBe(false)
    })
  })

  describe('setPreferredVoiceURI', () => {
    it('updates preferredVoiceURI', () => {
      const { result } = renderHook(() => useVoiceStore())

      act(() => {
        result.current.setPreferredVoiceURI('Google US English')
      })

      expect(result.current.preferredVoiceURI).toBe('Google US English')
    })

    it('can reset preferredVoiceURI to empty string', () => {
      const { result } = renderHook(() => useVoiceStore())

      act(() => {
        result.current.setPreferredVoiceURI('some-voice')
        result.current.setPreferredVoiceURI('')
      })

      expect(result.current.preferredVoiceURI).toBe('')
    })
  })
})
