import { UserRole } from '@/types/auth'

export interface JWTPayload {
  sub: string // user ID
  email: string
  role: UserRole
  exp: number // expiration timestamp
  iat: number // issued at timestamp
}

/**
 * Decode JWT token without verification (client-side)
 * Note: This is NOT secure verification, only for reading payload
 */
export function decodeJWT(token: string): JWTPayload | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) {
      return null
    }

    const payload = parts[1]
    const decoded = JSON.parse(atob(payload))

    return decoded as JWTPayload
  } catch (error) {
    console.error('Failed to decode JWT:', error)
    return null
  }
}

/**
 * Check if JWT token is expired
 */
export function isTokenExpired(token: string): boolean {
  const payload = decodeJWT(token)
  if (!payload || !payload.exp) {
    return true
  }

  // exp is in seconds, Date.now() is in milliseconds
  const now = Math.floor(Date.now() / 1000)
  return payload.exp < now
}

/**
 * Get token expiration time in seconds
 */
export function getTokenExpiration(token: string): number | null {
  const payload = decodeJWT(token)
  return payload?.exp || null
}

/**
 * Check if token is about to expire (within 5 minutes)
 */
export function isTokenExpiringSoon(token: string): boolean {
  const payload = decodeJWT(token)
  if (!payload || !payload.exp) {
    return true
  }

  const now = Math.floor(Date.now() / 1000)
  const fiveMinutes = 5 * 60
  return payload.exp - now < fiveMinutes
}
