export const locales = ['ru', 'en', 'fr', 'ar'] as const

export type Locale = (typeof locales)[number]

export const defaultLocale: Locale = 'ru'

export const localeNames: Record<Locale, string> = {
  ru: 'Русский',
  en: 'English',
  fr: 'Français',
  ar: 'العربية',
}

// RTL languages
export const rtlLocales: Locale[] = ['ar']

export function isRtlLocale(locale: Locale): boolean {
  return rtlLocales.includes(locale)
}
