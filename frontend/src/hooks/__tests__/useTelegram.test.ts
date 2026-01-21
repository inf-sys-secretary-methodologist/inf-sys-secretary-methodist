import { renderHook, waitFor, act } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useTelegramStatus,
  useGenerateVerificationCode,
  useDisconnectTelegram,
} from '../useTelegram'
import { apiClient } from '@/lib/api'

// Mock the API client
jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

// Wrapper to reset SWR cache between tests
const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(
    SWRConfig,
    { value: { dedupingInterval: 0, provider: () => new Map() } },
    children
  )

describe('useTelegram hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('useTelegramStatus', () => {
    it('returns connected status', async () => {
      const mockStatus = {
        connected: true,
        username: 'testuser',
        first_name: 'Test',
        connected_at: '2024-01-01T00:00:00Z',
      }

      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: mockStatus,
      })

      const { result } = renderHook(() => useTelegramStatus(), { wrapper })

      await waitFor(() => {
        expect(result.current.isConnected).toBe(true)
      })

      expect(result.current.username).toBe('testuser')
      expect(result.current.firstName).toBe('Test')
    })

    it('returns disconnected status', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: { connected: false },
      })

      const { result } = renderHook(() => useTelegramStatus(), { wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.isConnected).toBe(false)
      expect(result.current.username).toBeUndefined()
    })

    it('handles non-wrapped response', async () => {
      mockedApiClient.get.mockResolvedValue({
        connected: true,
        username: 'direct',
      })

      const { result } = renderHook(() => useTelegramStatus(), { wrapper })

      await waitFor(() => {
        expect(result.current.isConnected).toBe(true)
      })

      expect(result.current.username).toBe('direct')
    })
  })

  describe('useGenerateVerificationCode', () => {
    it('generates verification code', async () => {
      const mockCode = {
        code: '123456',
        expires_at: '2024-01-01T01:00:00Z',
        bot_username: 'TestBot',
        bot_link: 'https://t.me/TestBot',
      }

      mockedApiClient.post.mockResolvedValue({
        success: true,
        data: mockCode,
      })

      const { result } = renderHook(() => useGenerateVerificationCode(), { wrapper })

      expect(result.current.isPending).toBe(false)
      expect(result.current.data).toBeNull()

      let generatedCode
      await act(async () => {
        generatedCode = await result.current.mutateAsync()
      })

      expect(generatedCode).toEqual(mockCode)
      expect(result.current.data).toEqual(mockCode)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/telegram/verification-code')
    })

    it('handles non-wrapped response', async () => {
      const mockCode = {
        code: '654321',
        expires_at: '2024-01-01T01:00:00Z',
        bot_username: 'DirectBot',
        bot_link: 'https://t.me/DirectBot',
      }

      mockedApiClient.post.mockResolvedValue(mockCode)

      const { result } = renderHook(() => useGenerateVerificationCode(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync()
      })

      expect(result.current.data).toEqual(mockCode)
    })

    it('handles error', async () => {
      mockedApiClient.post.mockResolvedValue({
        success: false,
        error: { code: 'ERROR', message: 'Failed to generate' },
      })

      const { result } = renderHook(() => useGenerateVerificationCode(), { wrapper })

      await act(async () => {
        try {
          await result.current.mutateAsync()
        } catch {
          // Expected error
        }
      })

      expect(result.current.error).not.toBeNull()
    })

    it('resets state', async () => {
      const mockCode = {
        code: '123456',
        expires_at: '2024-01-01T01:00:00Z',
        bot_username: 'TestBot',
        bot_link: 'https://t.me/TestBot',
      }

      mockedApiClient.post.mockResolvedValue({
        success: true,
        data: mockCode,
      })

      const { result } = renderHook(() => useGenerateVerificationCode(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync()
      })

      expect(result.current.data).not.toBeNull()

      act(() => {
        result.current.reset()
      })

      expect(result.current.data).toBeNull()
      expect(result.current.error).toBeNull()
    })
  })

  describe('useDisconnectTelegram', () => {
    it('disconnects telegram', async () => {
      mockedApiClient.post.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useDisconnectTelegram(), { wrapper })

      expect(result.current.isPending).toBe(false)

      await act(async () => {
        await result.current.mutateAsync()
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/telegram/disconnect')
    })
  })
})
