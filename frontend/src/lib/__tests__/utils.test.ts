import { cn, getValidAvatarUrl } from '../utils'

describe('cn (className utility)', () => {
  it('merges class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('handles conditional class names', () => {
    expect(cn('foo', true && 'bar', false && 'baz')).toBe('foo bar')
  })

  it('merges tailwind classes correctly', () => {
    expect(cn('px-2 py-1', 'px-4')).toBe('py-1 px-4')
  })

  it('handles arrays', () => {
    expect(cn(['foo', 'bar'])).toBe('foo bar')
  })

  it('handles undefined and null', () => {
    expect(cn('foo', undefined, null, 'bar')).toBe('foo bar')
  })

  it('handles empty strings', () => {
    expect(cn('foo', '', 'bar')).toBe('foo bar')
  })

  it('handles objects', () => {
    expect(cn('foo', { bar: true, baz: false })).toBe('foo bar')
  })
})

describe('getValidAvatarUrl', () => {
  it('returns undefined for null input', () => {
    expect(getValidAvatarUrl(null)).toBeUndefined()
  })

  it('returns undefined for undefined input', () => {
    expect(getValidAvatarUrl(undefined)).toBeUndefined()
  })

  it('returns undefined for empty string', () => {
    expect(getValidAvatarUrl('')).toBeUndefined()
  })

  it('returns http URL as-is', () => {
    const url = 'http://example.com/avatar.jpg'
    expect(getValidAvatarUrl(url)).toBe(url)
  })

  it('returns https URL as-is', () => {
    const url = 'https://example.com/avatar.jpg'
    expect(getValidAvatarUrl(url)).toBe(url)
  })

  it('returns data URI as-is', () => {
    const dataUri =
      'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=='
    expect(getValidAvatarUrl(dataUri)).toBe(dataUri)
  })

  it('returns undefined for relative paths', () => {
    expect(getValidAvatarUrl('/avatars/user123.jpg')).toBeUndefined()
  })

  it('returns undefined for paths without protocol', () => {
    expect(getValidAvatarUrl('avatars/user123.jpg')).toBeUndefined()
  })
})
