import axios, { AxiosInstance, AxiosRequestConfig } from 'axios'
import { getStoredToken, setStoredToken, clearStoredToken } from '@/lib/auth/token'

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
        const token = getStoredToken()
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

  public clearAuthToken(): void {
    clearStoredToken()
  }

  public setAuthToken(token: string): void {
    setStoredToken(token)
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
