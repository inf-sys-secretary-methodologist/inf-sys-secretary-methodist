/**
 * @jest-environment jsdom
 */

import { renderHook, act } from '@testing-library/react'
import { useSpeechSynthesis } from '../useSpeechSynthesis'

function createMockUtterance() {
  return {
    lang: '',
    voice: null as SpeechSynthesisVoice | null,
    text: '',
    onstart: null as (() => void) | null,
    onend: null as (() => void) | null,
    onerror: null as (() => void) | null,
    onpause: null as (() => void) | null,
    onresume: null as (() => void) | null,
  }
}

describe('useSpeechSynthesis', () => {
  let mockUtterance: ReturnType<typeof createMockUtterance>
  let mockSpeechSynthesis: {
    speak: jest.Mock
    cancel: jest.Mock
    pause: jest.Mock
    resume: jest.Mock
    getVoices: jest.Mock
    addEventListener: jest.Mock
    removeEventListener: jest.Mock
  }

  const mockVoiceRu: SpeechSynthesisVoice = {
    voiceURI: 'ru-voice',
    name: 'Russian Voice',
    lang: 'ru-RU',
    localService: true,
    default: false,
  }

  const mockVoiceEn: SpeechSynthesisVoice = {
    voiceURI: 'en-voice',
    name: 'English Voice',
    lang: 'en-US',
    localService: true,
    default: false,
  }

  beforeEach(() => {
    jest.clearAllMocks()

    mockUtterance = createMockUtterance()

    // Mock SpeechSynthesisUtterance constructor
    ;(global as Record<string, unknown>).SpeechSynthesisUtterance = jest.fn((text: string) => {
      mockUtterance.text = text
      return mockUtterance
    })

    mockSpeechSynthesis = {
      speak: jest.fn(),
      cancel: jest.fn(),
      pause: jest.fn(),
      resume: jest.fn(),
      getVoices: jest.fn(() => [mockVoiceRu, mockVoiceEn]),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
    }

    Object.defineProperty(window, 'speechSynthesis', {
      value: mockSpeechSynthesis,
      writable: true,
      configurable: true,
    })
  })

  afterEach(() => {
    delete (global as Record<string, unknown>).SpeechSynthesisUtterance
  })

  describe('isSupported', () => {
    it('is false initially, true after mount when speechSynthesis exists', () => {
      const { result } = renderHook(() => useSpeechSynthesis())

      expect(result.current.isSupported).toBe(true)
    })

    it('stays false if speechSynthesis does not exist', () => {
      // Must delete the property so 'speechSynthesis' in window returns false
      const saved = Object.getOwnPropertyDescriptor(window, 'speechSynthesis')
      delete (window as unknown as Record<string, unknown>).speechSynthesis

      const { result } = renderHook(() => useSpeechSynthesis())

      expect(result.current.isSupported).toBe(false)

      // Restore for other tests
      if (saved) {
        Object.defineProperty(window, 'speechSynthesis', saved)
      }
    })
  })

  describe('voices', () => {
    it('loads voices from speechSynthesis.getVoices() filtered by lang', () => {
      const { result } = renderHook(() => useSpeechSynthesis({ lang: 'ru-RU' }))

      // Only the Russian voice should match
      expect(result.current.voices).toEqual([mockVoiceRu])
    })

    it('returns all voices if no voices match the lang filter', () => {
      mockSpeechSynthesis.getVoices.mockReturnValue([mockVoiceEn])

      const { result } = renderHook(() => useSpeechSynthesis({ lang: 'ru-RU' }))

      // No Russian voices, so all voices returned
      expect(result.current.voices).toEqual([mockVoiceEn])
    })
  })

  describe('speak', () => {
    it('calls speechSynthesis.speak with an utterance', () => {
      const { result } = renderHook(() => useSpeechSynthesis())

      act(() => {
        result.current.speak('Hello world')
      })

      expect(mockSpeechSynthesis.cancel).toHaveBeenCalled()
      expect(global.SpeechSynthesisUtterance).toHaveBeenCalledWith('Hello world')
      expect(mockSpeechSynthesis.speak).toHaveBeenCalled()
    })

    it('strips markdown from text before speaking', () => {
      const { result } = renderHook(() => useSpeechSynthesis())

      act(() => {
        result.current.speak('**bold** and *italic* and `code`')
      })

      expect(global.SpeechSynthesisUtterance).toHaveBeenCalledWith('bold and italic and code')
    })

    it('strips code blocks from text', () => {
      const { result } = renderHook(() => useSpeechSynthesis())

      act(() => {
        result.current.speak('Before ```js\nconst x = 1\n``` After')
      })

      expect(global.SpeechSynthesisUtterance).toHaveBeenCalledWith('Before  After')
    })

    it('applies preferredVoiceURI when matching voice exists', () => {
      const { result } = renderHook(() =>
        useSpeechSynthesis({ lang: 'ru-RU', preferredVoiceURI: 'ru-voice' })
      )

      act(() => {
        result.current.speak('test')
      })

      expect(mockUtterance.voice).toBe(mockVoiceRu)
    })
  })

  describe('cancel', () => {
    it('calls speechSynthesis.cancel and sets isSpeaking to false', () => {
      const { result } = renderHook(() => useSpeechSynthesis())

      act(() => {
        result.current.cancel()
      })

      expect(mockSpeechSynthesis.cancel).toHaveBeenCalled()
      expect(result.current.isSpeaking).toBe(false)
    })
  })

  describe('isSpeaking lifecycle', () => {
    it('tracks utterance lifecycle via onstart and onend', () => {
      const { result } = renderHook(() => useSpeechSynthesis())

      act(() => {
        result.current.speak('Hello')
      })

      // Simulate onstart
      act(() => {
        mockUtterance.onstart?.()
      })

      expect(result.current.isSpeaking).toBe(true)

      // Simulate onend
      act(() => {
        mockUtterance.onend?.()
      })

      expect(result.current.isSpeaking).toBe(false)
    })
  })

  describe('cleanup', () => {
    it('calls cancel on unmount', () => {
      const { unmount } = renderHook(() => useSpeechSynthesis())

      unmount()

      expect(mockSpeechSynthesis.cancel).toHaveBeenCalled()
    })
  })
})
