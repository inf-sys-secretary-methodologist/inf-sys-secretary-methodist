import { cookies } from 'next/headers'
import { Locale, defaultLocale } from '@/i18n/config'

const COOKIE_NAME = 'NEXT_LOCALE'

export async function getUserLocale(): Promise<Locale> {
  const cookieStore = await cookies()
  return (cookieStore.get(COOKIE_NAME)?.value as Locale) || defaultLocale
}
