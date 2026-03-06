import Cookies from 'js-cookie'

const AUTH_TOKEN_KEY = 'authToken'
const AUTH_COOKIE_NAME = 'auth-storage'

/**
 * Retrieves the auth token from available browser storage.
 *
 * Lookup order:
 *  1. localStorage (`authToken` key)
 *  2. Cookie fallback (`auth-storage` cookie, Zustand-persisted state)
 *
 * When the token is found in a cookie but not in localStorage, it is
 * automatically promoted to localStorage for faster future lookups.
 *
 * Returns `null` when running server-side or when no token is available.
 */
export function getStoredToken(): string | null {
  if (typeof window === 'undefined') return null

  // 1. Try localStorage (most reliable, set during login)
  const localToken = localStorage.getItem(AUTH_TOKEN_KEY)
  if (localToken) {
    return localToken
  }

  // 2. Fallback to cookie (might not have token if too large)
  /* c8 ignore start - Cookie fallback logic, browser-specific */
  try {
    const cookieValue = Cookies.get(AUTH_COOKIE_NAME)
    if (cookieValue) {
      const decoded = decodeURIComponent(cookieValue)
      const parsed = JSON.parse(decoded)
      const token: string | undefined = parsed.state?.token
      if (token) {
        // Promote to localStorage for future requests
        localStorage.setItem(AUTH_TOKEN_KEY, token)
        return token
      }
    }
  } catch {
    // Cookie parsing failed — ignore and return null
  }
  /* c8 ignore stop */

  return null
}
