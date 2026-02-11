import { apiClient } from '../api'

// API Response wrapper type from backend
interface ApiResponse<T> {
  success: boolean
  data: T
  error?: {
    code: string
    message: string
  }
}

/**
 * SWR fetcher utility for making API requests
 * Handles ApiResponse wrapper format and error extraction
 *
 * @param url - The API endpoint to fetch
 * @returns The data from the API response
 * @throws Error with message from API response or 'Unknown error'
 *
 * @example
 * ```typescript
 * const { data, error } = useSWR('/api/users', swrFetcher<User[]>)
 * ```
 */
export async function swrFetcher<T>(url: string): Promise<T> {
  const response = await apiClient.get<ApiResponse<T>>(url)
  if (response && 'success' in response) {
    if (response.success) return response.data
    throw new Error(response.error?.message || 'Unknown error')
  }
  return response as T
}
