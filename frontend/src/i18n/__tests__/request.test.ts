// Mock next-intl/server
const mockGetRequestConfig = jest.fn((configFn) => configFn)

jest.mock('next-intl/server', () => ({
  getRequestConfig: mockGetRequestConfig,
}))

// Mock getUserLocale
const mockGetUserLocale = jest.fn()
jest.mock('@/services/locale', () => ({
  getUserLocale: () => mockGetUserLocale(),
}))

describe('i18n request config', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // Reset module cache to re-import fresh
    jest.resetModules()
  })

  it('exports a function from getRequestConfig', async () => {
    mockGetUserLocale.mockResolvedValue('en')

    const requestConfig = await import('../request')
    expect(requestConfig.default).toBeDefined()
  })

  it('calls getRequestConfig with async function', async () => {
    mockGetUserLocale.mockResolvedValue('en')

    await import('../request')

    expect(mockGetRequestConfig).toHaveBeenCalledWith(expect.any(Function))
  })

  it('returns locale from getUserLocale', async () => {
    mockGetUserLocale.mockResolvedValue('ru')

    const requestConfig = await import('../request')
    const configFn = requestConfig.default

    const result = await configFn()

    expect(result.locale).toBe('ru')
  })

  it('loads messages for the locale', async () => {
    mockGetUserLocale.mockResolvedValue('en')

    const requestConfig = await import('../request')
    const configFn = requestConfig.default

    const result = await configFn()

    expect(result.messages).toBeDefined()
  })

  it('calls getUserLocale to get current locale', async () => {
    mockGetUserLocale.mockResolvedValue('en')

    const requestConfig = await import('../request')
    const configFn = requestConfig.default

    await configFn()

    expect(mockGetUserLocale).toHaveBeenCalled()
  })

  it('handles different locales', async () => {
    // Test with Russian
    mockGetUserLocale.mockResolvedValue('ru')

    jest.resetModules()
    const requestConfigRu = await import('../request')
    const configFnRu = requestConfigRu.default
    const resultRu = await configFnRu()

    expect(resultRu.locale).toBe('ru')

    // Test with English
    mockGetUserLocale.mockResolvedValue('en')

    jest.resetModules()
    const requestConfigEn = await import('../request')
    const configFnEn = requestConfigEn.default
    const resultEn = await configFnEn()

    expect(resultEn.locale).toBe('en')
  })
})
