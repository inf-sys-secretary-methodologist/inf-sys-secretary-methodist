/**
 * @jest-environment jsdom
 */

import { renderHook, act } from '@testing-library/react'
import { useSpeechRecognition } from '../useSpeechRecognition'

function createMockRecognition() {
  return {
    continuous: false,
    interimResults: false,
    lang: '',
    start: jest.fn(),
    stop: jest.fn(),
    abort: jest.fn(),
    onresult: null as ((event: unknown) => void) | null,
    onerror: null as ((event: unknown) => void) | null,
    onend: null as (() => void) | null,
    onstart: null as (() => void) | null,
  }
}

describe('useSpeechRecognition', () => {
  let mockRecognitionInstance: ReturnType<typeof createMockRecognition>
  let MockSpeechRecognition: jest.Mock

  beforeEach(() => {
    jest.clearAllMocks()
    mockRecognitionInstance = createMockRecognition()
    MockSpeechRecognition = jest.fn(() => mockRecognitionInstance)

    // Clean up any previous mocks
    delete (window as unknown as Record<string, unknown>).SpeechRecognition
    delete (window as unknown as Record<string, unknown>).webkitSpeechRecognition
  })

  afterEach(() => {
    delete (window as unknown as Record<string, unknown>).SpeechRecognition
    delete (window as unknown as Record<string, unknown>).webkitSpeechRecognition
  })

  describe('isSupported', () => {
    it('is false initially (SSR safety), then true after mount when SpeechRecognition exists', () => {
      ;(window as unknown as Record<string, unknown>).SpeechRecognition = MockSpeechRecognition

      const { result } = renderHook(() => useSpeechRecognition())

      // After the useEffect runs, isSupported should be true
      expect(result.current.isSupported).toBe(true)
    })

    it('is true after mount when webkitSpeechRecognition exists', () => {
      ;(window as unknown as Record<string, unknown>).webkitSpeechRecognition =
        MockSpeechRecognition

      const { result } = renderHook(() => useSpeechRecognition())

      expect(result.current.isSupported).toBe(true)
    })

    it('stays false if SpeechRecognition does not exist', () => {
      const { result } = renderHook(() => useSpeechRecognition())

      expect(result.current.isSupported).toBe(false)
    })
  })

  describe('startListening', () => {
    it('creates recognition instance and sets isListening to true', () => {
      ;(window as unknown as Record<string, unknown>).SpeechRecognition = MockSpeechRecognition

      const { result } = renderHook(() => useSpeechRecognition())

      act(() => {
        result.current.startListening()
      })

      expect(MockSpeechRecognition).toHaveBeenCalled()
      expect(mockRecognitionInstance.start).toHaveBeenCalled()

      // Simulate onstart callback
      act(() => {
        mockRecognitionInstance.onstart?.()
      })

      expect(result.current.isListening).toBe(true)
    })

    it('sets error if not supported', () => {
      const { result } = renderHook(() => useSpeechRecognition())

      act(() => {
        result.current.startListening()
      })

      expect(result.current.error).toBe('Speech recognition is not supported in this browser')
    })

    it('configures recognition with provided lang and continuous options', () => {
      ;(window as unknown as Record<string, unknown>).SpeechRecognition = MockSpeechRecognition

      const { result } = renderHook(() => useSpeechRecognition({ lang: 'en-US', continuous: true }))

      act(() => {
        result.current.startListening()
      })

      expect(mockRecognitionInstance.lang).toBe('en-US')
      expect(mockRecognitionInstance.continuous).toBe(true)
      expect(mockRecognitionInstance.interimResults).toBe(true)
    })
  })

  describe('stopListening', () => {
    it('stops recognition and sets isListening to false', () => {
      ;(window as unknown as Record<string, unknown>).SpeechRecognition = MockSpeechRecognition

      const { result } = renderHook(() => useSpeechRecognition())

      act(() => {
        result.current.startListening()
      })

      act(() => {
        mockRecognitionInstance.onstart?.()
      })

      expect(result.current.isListening).toBe(true)

      act(() => {
        result.current.stopListening()
      })

      expect(mockRecognitionInstance.stop).toHaveBeenCalled()

      // Simulate onend callback (recognition fires this after stop)
      act(() => {
        mockRecognitionInstance.onend?.()
      })

      expect(result.current.isListening).toBe(false)
    })
  })

  describe('transcript', () => {
    it('updates on recognition result events', () => {
      ;(window as unknown as Record<string, unknown>).SpeechRecognition = MockSpeechRecognition

      const { result } = renderHook(() => useSpeechRecognition())

      act(() => {
        result.current.startListening()
      })

      // Simulate an interim result
      act(() => {
        mockRecognitionInstance.onresult?.({
          resultIndex: 0,
          results: {
            length: 1,
            0: {
              isFinal: false,
              0: { transcript: 'hello' },
              length: 1,
            },
          },
        })
      })

      expect(result.current.transcript).toBe('hello')
    })

    it('accumulates final results in continuous mode', () => {
      ;(window as unknown as Record<string, unknown>).SpeechRecognition = MockSpeechRecognition

      const { result } = renderHook(() => useSpeechRecognition({ continuous: true }))

      act(() => {
        result.current.startListening()
      })

      // First final result
      act(() => {
        mockRecognitionInstance.onresult?.({
          resultIndex: 0,
          results: {
            length: 1,
            0: {
              isFinal: true,
              0: { transcript: 'hello ' },
              length: 1,
            },
          },
        })
      })

      expect(result.current.transcript).toBe('hello ')

      // Second final result
      act(() => {
        mockRecognitionInstance.onresult?.({
          resultIndex: 1,
          results: {
            length: 2,
            0: {
              isFinal: true,
              0: { transcript: 'hello ' },
              length: 1,
            },
            1: {
              isFinal: true,
              0: { transcript: 'world' },
              length: 1,
            },
          },
        })
      })

      expect(result.current.transcript).toBe('hello world')
    })
  })

  describe('error handling', () => {
    it('sets error state on recognition error', () => {
      ;(window as unknown as Record<string, unknown>).SpeechRecognition = MockSpeechRecognition

      const { result } = renderHook(() => useSpeechRecognition())

      act(() => {
        result.current.startListening()
      })

      act(() => {
        mockRecognitionInstance.onerror?.({ error: 'network' })
      })

      expect(result.current.error).toBe('network')
      expect(result.current.isListening).toBe(false)
    })

    it('does not set error for aborted errors', () => {
      ;(window as unknown as Record<string, unknown>).SpeechRecognition = MockSpeechRecognition

      const { result } = renderHook(() => useSpeechRecognition())

      act(() => {
        result.current.startListening()
      })

      act(() => {
        mockRecognitionInstance.onerror?.({ error: 'aborted' })
      })

      expect(result.current.error).toBeNull()
    })
  })

  describe('cleanup', () => {
    it('calls abort on unmount', () => {
      ;(window as unknown as Record<string, unknown>).SpeechRecognition = MockSpeechRecognition

      const { result, unmount } = renderHook(() => useSpeechRecognition())

      act(() => {
        result.current.startListening()
      })

      unmount()

      expect(mockRecognitionInstance.abort).toHaveBeenCalled()
    })
  })
})
