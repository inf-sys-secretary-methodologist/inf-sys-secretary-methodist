'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useLocale } from 'next-intl'
import { Globe } from 'lucide-react'
import { locales, localeNames, type Locale } from '@/i18n/config'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

const COOKIE_NAME = 'NEXT_LOCALE'

function setLocaleCookie(locale: Locale) {
  // Set cookie with 1 year expiry
  const maxAge = 60 * 60 * 24 * 365
  document.cookie = `${COOKIE_NAME}=${locale}; path=/; max-age=${maxAge}; SameSite=Lax`
}

export function LanguageSwitcher() {
  const locale = useLocale() as Locale
  const router = useRouter()
  const [isPending, setIsPending] = useState(false)

  const handleLocaleChange = (newLocale: Locale) => {
    setIsPending(true)
    setLocaleCookie(newLocale)
    router.refresh()
    // Reset pending state after a short delay
    setTimeout(() => setIsPending(false), 500)
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          className="relative inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border-2 border-gray-300 bg-white text-gray-900 transition-all duration-200 hover:bg-gray-100 hover:scale-105 hover:shadow-lg active:scale-95 dark:border-gray-700 dark:bg-gray-900 dark:text-white dark:hover:bg-gray-800 shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
          disabled={isPending}
          aria-label="Change language"
          type="button"
        >
          <Globe className="h-5 w-5" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {locales.map((l) => (
          <DropdownMenuItem
            key={l}
            onClick={() => handleLocaleChange(l)}
            className={locale === l ? 'bg-accent' : ''}
          >
            <span className="mr-2">{getFlagEmoji(l)}</span>
            {localeNames[l]}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

function getFlagEmoji(locale: Locale): string {
  const flags: Record<Locale, string> = {
    ru: '🇷🇺',
    en: '🇬🇧',
    fr: '🇫🇷',
    ar: '🇸🇦',
  }
  /* c8 ignore next */
  return flags[locale] || '🌐'
}
