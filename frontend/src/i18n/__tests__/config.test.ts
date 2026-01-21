import { locales, defaultLocale, localeNames, rtlLocales, isRtlLocale, Locale } from '../config'

describe('i18n config', () => {
  describe('locales', () => {
    it('contains expected locales', () => {
      expect(locales).toContain('ru')
      expect(locales).toContain('en')
      expect(locales).toContain('fr')
      expect(locales).toContain('ar')
    })

    it('has correct number of locales', () => {
      expect(locales.length).toBe(4)
    })
  })

  describe('defaultLocale', () => {
    it('is set to Russian', () => {
      expect(defaultLocale).toBe('ru')
    })

    it('is one of the available locales', () => {
      expect(locales).toContain(defaultLocale)
    })
  })

  describe('localeNames', () => {
    it('has a name for each locale', () => {
      locales.forEach((locale) => {
        expect(localeNames[locale]).toBeDefined()
        expect(typeof localeNames[locale]).toBe('string')
        expect(localeNames[locale].length).toBeGreaterThan(0)
      })
    })

    it('has correct locale names', () => {
      expect(localeNames.ru).toBe('Русский')
      expect(localeNames.en).toBe('English')
      expect(localeNames.fr).toBe('Français')
      expect(localeNames.ar).toBe('العربية')
    })
  })

  describe('rtlLocales', () => {
    it('contains Arabic', () => {
      expect(rtlLocales).toContain('ar')
    })

    it('does not contain LTR languages', () => {
      expect(rtlLocales).not.toContain('ru')
      expect(rtlLocales).not.toContain('en')
      expect(rtlLocales).not.toContain('fr')
    })
  })

  describe('isRtlLocale', () => {
    it('returns true for Arabic', () => {
      expect(isRtlLocale('ar')).toBe(true)
    })

    it('returns false for Russian', () => {
      expect(isRtlLocale('ru')).toBe(false)
    })

    it('returns false for English', () => {
      expect(isRtlLocale('en')).toBe(false)
    })

    it('returns false for French', () => {
      expect(isRtlLocale('fr')).toBe(false)
    })

    it('returns false for unknown locale', () => {
      // This tests the behavior with an invalid locale
      expect(isRtlLocale('de' as Locale)).toBe(false)
    })
  })
})
