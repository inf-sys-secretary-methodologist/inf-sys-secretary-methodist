import { getUserLocale } from '../locale'
import { cookies } from 'next/headers'

// Mock next/headers
jest.mock('next/headers', () => ({
  cookies: jest.fn(),
}))

const mockedCookies = jest.mocked(cookies)

describe('getUserLocale', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('returns locale from cookie when set', async () => {
    mockedCookies.mockResolvedValue({
      get: jest.fn().mockReturnValue({ value: 'en' }),
    } as unknown as ReturnType<typeof cookies>)

    const locale = await getUserLocale()

    expect(locale).toBe('en')
  })

  it('returns French locale from cookie', async () => {
    mockedCookies.mockResolvedValue({
      get: jest.fn().mockReturnValue({ value: 'fr' }),
    } as unknown as ReturnType<typeof cookies>)

    const locale = await getUserLocale()

    expect(locale).toBe('fr')
  })

  it('returns Arabic locale from cookie', async () => {
    mockedCookies.mockResolvedValue({
      get: jest.fn().mockReturnValue({ value: 'ar' }),
    } as unknown as ReturnType<typeof cookies>)

    const locale = await getUserLocale()

    expect(locale).toBe('ar')
  })

  it('returns default locale when cookie is not set', async () => {
    mockedCookies.mockResolvedValue({
      get: jest.fn().mockReturnValue(undefined),
    } as unknown as ReturnType<typeof cookies>)

    const locale = await getUserLocale()

    expect(locale).toBe('ru') // default locale
  })

  it('returns default locale when cookie value is null', async () => {
    mockedCookies.mockResolvedValue({
      get: jest.fn().mockReturnValue({ value: null }),
    } as unknown as ReturnType<typeof cookies>)

    const locale = await getUserLocale()

    expect(locale).toBe('ru') // default locale
  })

  it('calls cookies with correct cookie name', async () => {
    const getMock = jest.fn().mockReturnValue({ value: 'en' })
    mockedCookies.mockResolvedValue({
      get: getMock,
    } as unknown as ReturnType<typeof cookies>)

    await getUserLocale()

    expect(getMock).toHaveBeenCalledWith('NEXT_LOCALE')
  })
})
