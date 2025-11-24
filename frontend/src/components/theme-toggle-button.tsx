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
        className="inline-flex items-center justify-center rounded-2xl p-4 border-2 border-gray-300 bg-white dark:border-gray-700 dark:bg-gray-900 shadow-lg opacity-50"
        aria-hidden="true"
      >
        <Sun className="h-6 w-6 text-gray-400" />
      </div>
    )
  }

  return (
    <button
      onClick={() => {
        const newTheme = theme === 'dark' ? 'light' : 'dark'
        setTheme(newTheme)
      }}
      className="inline-flex items-center justify-center rounded-2xl p-4 transition-all duration-200 border-2 border-gray-300 bg-white text-gray-900 hover:bg-gray-100 hover:scale-105 active:scale-95 dark:border-gray-700 dark:bg-gray-900 dark:text-white dark:hover:bg-gray-800 shadow-lg hover:shadow-xl"
      aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`}
      type="button"
    >
      {theme === 'dark' ? (
        <Sun className="h-6 w-6 transition-transform duration-200" />
      ) : (
        <Moon className="h-6 w-6 transition-transform duration-200" />
      )}
    </button>
  )
}
