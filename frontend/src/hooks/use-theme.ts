'use client'

import { useTheme as useNextTheme } from 'next-themes'

/**
 * Хук для работы с темой приложения
 *
 * @example
 * ```tsx
 * const { theme, setTheme } = useTheme()
 *
 * // Переключить тему
 * setTheme('dark')
 * setTheme('light')
 * setTheme('system')
 * ```
 */
export function useTheme() {
  const { theme, setTheme, resolvedTheme, systemTheme } = useNextTheme()

  const toggleTheme = () => {
    setTheme(theme === 'dark' ? 'light' : 'dark')
  }

  return {
    theme,
    setTheme,
    resolvedTheme,
    systemTheme,
    toggleTheme,
    isDark: resolvedTheme === 'dark',
    isLight: resolvedTheme === 'light',
  }
}
