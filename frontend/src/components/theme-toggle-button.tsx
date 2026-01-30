'use client'

import * as React from 'react'
import { Moon, Sun } from 'lucide-react'
import { useTheme } from 'next-themes'

export function ThemeToggleButton() {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  // Render a placeholder with the same dimensions to prevent layout shift
  if (!mounted) {
    return (
      <div
        className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border-2 border-gray-300 bg-white dark:border-gray-700 dark:bg-gray-900 shadow-md opacity-50"
        aria-hidden="true"
      >
        <Sun className="h-5 w-5 text-muted-foreground" />
      </div>
    )
  }

  return (
    <button
      onClick={() => {
        /* c8 ignore next */
        const newTheme = theme === 'dark' ? 'light' : 'dark'
        setTheme(newTheme)
      }}
      className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border-2 border-gray-300 bg-white text-gray-900 transition-all duration-200 hover:bg-gray-100 hover:scale-105 active:scale-95 dark:border-gray-700 dark:bg-gray-900 dark:text-white dark:hover:bg-gray-800 shadow-md hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`}
      type="button"
    >
      {theme === 'dark' ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
    </button>
  )
}
