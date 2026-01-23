import axios from 'axios'

// Mock axios
jest.mock('axios', () => {
  // Import the storage from module scope
  const storage = {
    request: { onFulfilled: null, onRejected: null },
    response: { onFulfilled: null, onRejected: null },
  }

  const mockAxiosInstance = {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    patch: jest.fn(),
    delete: jest.fn(),
    interceptors: {
      request: {
        use: jest.fn((onFulfilled, onRejected) => {
          storage.request.onFulfilled = onFulfilled
          storage.request.onRejected = onRejected
        }),
      },
      response: {
        use: jest.fn((onFulfilled, onRejected) => {
          storage.response.onFulfilled = onFulfilled
          storage.response.onRejected = onRejected
        }),
      },
    },
  }
  return {
    create: jest.fn(() => mockAxiosInstance),
    __mockInstance: mockAxiosInstance,
    __interceptorStorage: storage,
  }
})

import { apiClient } from '../api'

const getMockAxiosInstance = () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return (axios as any).__mockInstance
}

const getInterceptorStorage = () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return (axios as any).__interceptorStorage
}

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

  describe('HTTP methods', () => {
    const mockAxios = getMockAxiosInstance()

    beforeEach(() => {
      mockAxios.get.mockReset()
      mockAxios.post.mockReset()
      mockAxios.put.mockReset()
      mockAxios.patch.mockReset()
      mockAxios.delete.mockReset()
    })

    describe('get', () => {
      it('makes GET request and returns data', async () => {
        const mockData = { id: 1, name: 'test' }
        mockAxios.get.mockResolvedValue({ data: mockData })

        const result = await apiClient.get('/test')
        expect(result).toEqual(mockData)
        expect(mockAxios.get).toHaveBeenCalledWith('/test', undefined)
      })

      it('passes config to axios', async () => {
        mockAxios.get.mockResolvedValue({ data: {} })
        const config = { headers: { 'X-Custom': 'value' } }

        await apiClient.get('/test', config)
        expect(mockAxios.get).toHaveBeenCalledWith('/test', config)
      })

      it('propagates errors', async () => {
        const error = new Error('Network error')
        mockAxios.get.mockRejectedValue(error)

        await expect(apiClient.get('/test')).rejects.toThrow('Network error')
      })
    })

    describe('post', () => {
      it('makes POST request with data and returns response', async () => {
        const mockData = { id: 1 }
        const postData = { name: 'new item' }
        mockAxios.post.mockResolvedValue({ data: mockData })

        const result = await apiClient.post('/test', postData)
        expect(result).toEqual(mockData)
        expect(mockAxios.post).toHaveBeenCalledWith('/test', postData, undefined)
      })

      it('can make POST request without data', async () => {
        mockAxios.post.mockResolvedValue({ data: {} })

        await apiClient.post('/test')
        expect(mockAxios.post).toHaveBeenCalledWith('/test', undefined, undefined)
      })

      it('passes config to axios', async () => {
        mockAxios.post.mockResolvedValue({ data: {} })
        const config = { timeout: 5000 }

        await apiClient.post('/test', { data: 1 }, config)
        expect(mockAxios.post).toHaveBeenCalledWith('/test', { data: 1 }, config)
      })
    })

    describe('put', () => {
      it('makes PUT request with data and returns response', async () => {
        const mockData = { id: 1, updated: true }
        const putData = { name: 'updated item' }
        mockAxios.put.mockResolvedValue({ data: mockData })

        const result = await apiClient.put('/test/1', putData)
        expect(result).toEqual(mockData)
        expect(mockAxios.put).toHaveBeenCalledWith('/test/1', putData, undefined)
      })

      it('can make PUT request without data', async () => {
        mockAxios.put.mockResolvedValue({ data: {} })

        await apiClient.put('/test/1')
        expect(mockAxios.put).toHaveBeenCalledWith('/test/1', undefined, undefined)
      })
    })

    describe('patch', () => {
      it('makes PATCH request with data and returns response', async () => {
        const mockData = { id: 1, patched: true }
        const patchData = { status: 'active' }
        mockAxios.patch.mockResolvedValue({ data: mockData })

        const result = await apiClient.patch('/test/1', patchData)
        expect(result).toEqual(mockData)
        expect(mockAxios.patch).toHaveBeenCalledWith('/test/1', patchData, undefined)
      })
    })

    describe('delete', () => {
      it('makes DELETE request and returns response', async () => {
        const mockData = { deleted: true }
        mockAxios.delete.mockResolvedValue({ data: mockData })

        const result = await apiClient.delete('/test/1')
        expect(result).toEqual(mockData)
        expect(mockAxios.delete).toHaveBeenCalledWith('/test/1', undefined)
      })

      it('passes config to axios', async () => {
        mockAxios.delete.mockResolvedValue({ data: {} })
        const config = { params: { force: true } }

        await apiClient.delete('/test/1', config)
        expect(mockAxios.delete).toHaveBeenCalledWith('/test/1', config)
      })
    })
  })

  describe('Request Interceptor', () => {
    it('adds Authorization header when token exists in localStorage', () => {
      mockLocalStorage.getItem.mockReturnValue('test-token')

      const config = { headers: {} as Record<string, string> }
      const result = getInterceptorStorage().request.onFulfilled?.(config)

      expect(result).toEqual({
        headers: { Authorization: 'Bearer test-token' },
      })
    })

    it('does not add Authorization header when no token', () => {
      mockLocalStorage.getItem.mockReturnValue(null)
      Object.defineProperty(document, 'cookie', {
        value: '',
        writable: true,
      })

      const config = { headers: {} as Record<string, string> }
      const result = getInterceptorStorage().request.onFulfilled?.(config)

      expect(result).toEqual({ headers: {} })
    })

    it('rejects on request error', async () => {
      const error = new Error('Request setup failed')

      await expect(getInterceptorStorage().request.onRejected?.(error)).rejects.toThrow(
        'Request setup failed'
      )
    })
  })

  describe('Response Interceptor', () => {
    beforeEach(() => {
      Object.defineProperty(window, 'location', {
        value: { pathname: '/dashboard', href: '' },
        writable: true,
      })
    })

    it('passes through successful responses', () => {
      const response = { data: { success: true }, status: 200 }
      const result = getInterceptorStorage().response.onFulfilled?.(response)

      expect(result).toEqual(response)
    })

    it('redirects to login on 401 for non-auth endpoints', async () => {
      const error = {
        response: { status: 401 },
        config: { url: '/api/users' },
      }

      await expect(getInterceptorStorage().response.onRejected?.(error)).rejects.toEqual(error)
      expect(window.location.href).toBe('/login')
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('authToken')
    })

    it('does not redirect on 401 for login endpoint', async () => {
      const error = {
        response: { status: 401 },
        config: { url: '/auth/login' },
      }

      await expect(getInterceptorStorage().response.onRejected?.(error)).rejects.toEqual(error)
      expect(window.location.href).not.toBe('/login')
    })

    it('does not redirect on 401 for register endpoint', async () => {
      const error = {
        response: { status: 401 },
        config: { url: '/auth/register' },
      }

      await expect(getInterceptorStorage().response.onRejected?.(error)).rejects.toEqual(error)
      expect(window.location.href).not.toBe('/login')
    })

    it('does not redirect when already on login page', async () => {
      Object.defineProperty(window, 'location', {
        value: { pathname: '/login', href: '' },
        writable: true,
      })

      const error = {
        response: { status: 401 },
        config: { url: '/api/users' },
      }

      await expect(getInterceptorStorage().response.onRejected?.(error)).rejects.toEqual(error)
      expect(window.location.href).not.toBe('/login')
    })

    it('does not redirect on non-401 errors', async () => {
      const error = {
        response: { status: 500 },
        config: { url: '/api/users' },
      }

      await expect(getInterceptorStorage().response.onRejected?.(error)).rejects.toEqual(error)
      expect(window.location.href).not.toBe('/login')
    })

    it('handles error without response', async () => {
      const error = {
        config: { url: '/api/users' },
      }

      await expect(getInterceptorStorage().response.onRejected?.(error)).rejects.toEqual(error)
    })
  })

  describe('getAuthToken (via interceptor)', () => {
    it('returns token from localStorage', () => {
      mockLocalStorage.getItem.mockReturnValue('local-token')

      const config = { headers: {} as Record<string, string> }
      getInterceptorStorage().request.onFulfilled?.(config)

      expect(config.headers.Authorization).toBe('Bearer local-token')
    })

    it('returns null when no token in localStorage', () => {
      mockLocalStorage.getItem.mockReturnValue(null)

      const config = { headers: {} as Record<string, string> }
      getInterceptorStorage().request.onFulfilled?.(config)

      // No Authorization header should be set
      expect(config.headers.Authorization).toBeUndefined()
    })
  })
})
