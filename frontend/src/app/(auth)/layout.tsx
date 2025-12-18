import type { Metadata } from 'next'
import { ThemeSettingsPopover } from '@/components/theme-settings-popover'

export const metadata: Metadata = {
  title: {
    template: '%s | Аутентификация',
    default: 'Аутентификация',
  },
  description: 'Вход и регистрация в системе',
}

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4">
      {/* Theme Toggle */}
      <div className="fixed top-8 right-8 z-50">
        <ThemeSettingsPopover />
      </div>

      {/* Auth Card */}
      <div className="w-full max-w-md mt-16 sm:mt-0">
        <div className="bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl p-8">
          {children}
        </div>
      </div>

      {/* Footer */}
      <footer className="mt-8 text-center text-sm text-muted-foreground">
        <p>© 2025 Информационная система секретаря-методиста</p>
      </footer>
    </div>
  )
}
