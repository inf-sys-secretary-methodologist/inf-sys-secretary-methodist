import { apiClient } from '../api'

describe('apiClient', () => {
  // Mock localStorage
  const mockLocalStorage = (() => {
    let store: Record<string, string> = {}
    return {
      getItem: jest.fn((key: string) => store[key] || null),
      setItem: jest.fn((key: string, value: string) => {
        store[key] = value
      }),
      removeItem: jest.fn((key: string) => {
        delete store[key]
      }),
      clear: () => {
        store = {}
      },
    }
  })()

  beforeEach(() => {
    Object.defineProperty(window, 'localStorage', {
      value: mockLocalStorage,
      writable: true,
    })
    mockLocalStorage.clear()
    jest.clearAllMocks()
  })

  describe('setAuthToken', () => {
    it('stores token in localStorage', () => {
      apiClient.setAuthToken('test-token-123')
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('authToken', 'test-token-123')
    })

    it('can store different tokens', () => {
      apiClient.setAuthToken('token-1')
      apiClient.setAuthToken('token-2')
      expect(mockLocalStorage.setItem).toHaveBeenCalledTimes(2)
      expect(mockLocalStorage.setItem).toHaveBeenLastCalledWith('authToken', 'token-2')
    })
  })

  describe('clearAuthToken', () => {
    it('removes token from localStorage', () => {
      apiClient.clearAuthToken()
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('authToken')
    })

    it('can be called multiple times', () => {
      apiClient.clearAuthToken()
      apiClient.clearAuthToken()
      expect(mockLocalStorage.removeItem).toHaveBeenCalledTimes(2)
    })
  })

  describe('token management flow', () => {
    it('set then clear removes token', () => {
      apiClient.setAuthToken('my-token')
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('authToken', 'my-token')

      apiClient.clearAuthToken()
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('authToken')
    })
  })
})
