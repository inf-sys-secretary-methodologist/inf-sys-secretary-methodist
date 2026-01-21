import { decodeJWT, isTokenExpired, getTokenExpiration, isTokenExpiringSoon } from '../jwt'
import { UserRole } from '@/types/auth'

// Helper to create a valid JWT-like token
function createMockToken(payload: object): string {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
  const body = btoa(JSON.stringify(payload))
  const signature = 'mock-signature'
  return `${header}.${body}.${signature}`
}

describe('decodeJWT', () => {
  it('decodes a valid JWT token', () => {
    const payload = {
      sub: '123',
      email: 'test@example.com',
      role: UserRole.STUDENT,
      exp: Math.floor(Date.now() / 1000) + 3600,
      iat: Math.floor(Date.now() / 1000),
    }
    const token = createMockToken(payload)

    const result = decodeJWT(token)

    expect(result).toEqual(payload)
  })

  it('returns null for invalid token format', () => {
    expect(decodeJWT('invalid-token')).toBeNull()
    expect(decodeJWT('only.two')).toBeNull()
    expect(decodeJWT('')).toBeNull()
  })

  it('returns null for token with invalid base64', () => {
    const token = 'header.!!!invalid-base64!!!.signature'
    expect(decodeJWT(token)).toBeNull()
  })

  it('returns null for token with invalid JSON', () => {
    const header = btoa(JSON.stringify({ alg: 'HS256' }))
    const invalidJson = btoa('not-json')
    const token = `${header}.${invalidJson}.signature`
    expect(decodeJWT(token)).toBeNull()
  })
})

describe('isTokenExpired', () => {
  it('returns false for non-expired token', () => {
    const payload = {
      sub: '123',
      exp: Math.floor(Date.now() / 1000) + 3600, // 1 hour from now
    }
    const token = createMockToken(payload)

    expect(isTokenExpired(token)).toBe(false)
  })

  it('returns true for expired token', () => {
    const payload = {
      sub: '123',
      exp: Math.floor(Date.now() / 1000) - 3600, // 1 hour ago
    }
    const token = createMockToken(payload)

    expect(isTokenExpired(token)).toBe(true)
  })

  it('returns true for invalid token', () => {
    expect(isTokenExpired('invalid-token')).toBe(true)
  })

  it('returns true for token without exp', () => {
    const payload = { sub: '123' }
    const token = createMockToken(payload)

    expect(isTokenExpired(token)).toBe(true)
  })
})

describe('getTokenExpiration', () => {
  it('returns expiration timestamp for valid token', () => {
    const exp = Math.floor(Date.now() / 1000) + 3600
    const payload = { sub: '123', exp }
    const token = createMockToken(payload)

    expect(getTokenExpiration(token)).toBe(exp)
  })

  it('returns null for invalid token', () => {
    expect(getTokenExpiration('invalid-token')).toBeNull()
  })

  it('returns null for token without exp', () => {
    const payload = { sub: '123' }
    const token = createMockToken(payload)

    expect(getTokenExpiration(token)).toBeNull()
  })
})

describe('isTokenExpiringSoon', () => {
  it('returns false for token expiring in more than 5 minutes', () => {
    const payload = {
      sub: '123',
      exp: Math.floor(Date.now() / 1000) + 600, // 10 minutes from now
    }
    const token = createMockToken(payload)

    expect(isTokenExpiringSoon(token)).toBe(false)
  })

  it('returns true for token expiring in less than 5 minutes', () => {
    const payload = {
      sub: '123',
      exp: Math.floor(Date.now() / 1000) + 120, // 2 minutes from now
    }
    const token = createMockToken(payload)

    expect(isTokenExpiringSoon(token)).toBe(true)
  })

  it('returns true for already expired token', () => {
    const payload = {
      sub: '123',
      exp: Math.floor(Date.now() / 1000) - 60, // 1 minute ago
    }
    const token = createMockToken(payload)

    expect(isTokenExpiringSoon(token)).toBe(true)
  })

  it('returns true for invalid token', () => {
    expect(isTokenExpiringSoon('invalid-token')).toBe(true)
  })

  it('returns true for token without exp', () => {
    const payload = { sub: '123' }
    const token = createMockToken(payload)

    expect(isTokenExpiringSoon(token)).toBe(true)
  })
})
