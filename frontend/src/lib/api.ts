import axios, { AxiosInstance, AxiosRequestConfig } from 'axios'

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 10000,
    })

    this.setupInterceptors()
  }

  private setupInterceptors() {
    this.client.interceptors.request.use(
      (config) => {
        const token = this.getAuthToken()
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    /* c8 ignore start - Response interceptor for 401 handling */
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        // Only redirect on 401 if NOT on login/register endpoints
        // Login 401 means wrong credentials, not session expiration
        const requestUrl = error.config?.url || ''
        const isAuthEndpoint =
          requestUrl.includes('/auth/login') || requestUrl.includes('/auth/register')

        if (error.response?.status === 401 && !isAuthEndpoint) {
          // Prevent redirect loop - don't redirect if already on login page
          const currentPath = typeof window !== 'undefined' ? window.location.pathname : ''
          if (!currentPath.startsWith('/login')) {
            this.clearAuthToken()
            window.location.href = '/login'
          }
        }
        return Promise.reject(error)
      }
    )
    /* c8 ignore stop */
  }

  private getAuthToken(): string | null {
    if (typeof window !== 'undefined') {
      // Try localStorage first
      const localToken = localStorage.getItem('authToken')
      if (localToken) {
        console.log('🔑 Token from localStorage:', localToken.substring(0, 20) + '...')
        return localToken
      }

      /* c8 ignore start - Cookie fallback logic, browser-specific */
      // Fallback to cookie (for cases when localStorage wasn't set yet)
      try {
        const cookieValue = this.getCookieValue('auth-storage')
        console.log('🍪 Cookie value exists:', !!cookieValue)
        if (cookieValue) {
          const decoded = decodeURIComponent(cookieValue)
          const parsed = JSON.parse(decoded)
          const token = parsed.state?.token
          console.log('🔑 Token from cookie:', token ? token.substring(0, 20) + '...' : 'null')
          if (token) {
            // Also save to localStorage for future requests
            localStorage.setItem('authToken', token)
            return token
          }
        }
      } catch (e) {
        console.error('❌ Cookie parsing failed:', e)
      }
      /* c8 ignore stop */
    }
    return null
  }

  /* c8 ignore start - Cookie parsing helper, browser-specific */
  private getCookieValue(name: string): string | null {
    const nameEQ = name + '='
    const ca = document.cookie.split(';')
    for (let i = 0; i < ca.length; i++) {
      let c = ca[i]
      while (c.charAt(0) === ' ') c = c.substring(1, c.length)
      if (c.indexOf(nameEQ) === 0) {
        return c.substring(nameEQ.length, c.length)
      }
    }
    return null
  }
  /* c8 ignore stop */

  public clearAuthToken(): void {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('authToken')
    }
  }

  public setAuthToken(token: string): void {
    if (typeof window !== 'undefined') {
      localStorage.setItem('authToken', token)
    }
  }

  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get<T>(url, config)
    return response.data
  }

  async post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.post<T>(url, data, config)
    return response.data
  }

  async put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.put<T>(url, data, config)
    return response.data
  }

  async patch<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.patch<T>(url, data, config)
    return response.data
  }

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete<T>(url, config)
    return response.data
  }
}

export const apiClient = new ApiClient()
