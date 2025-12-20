import type { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { ThemeSettingsPopover } from '@/components/theme-settings-popover'
import { LanguageSwitcher } from '@/components/language-switcher'

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('authPages')
  return {
    title: {
      template: `%s | ${t('metaTitle')}`,
      default: t('metaTitle'),
    },
    description: t('metaDescription'),
  }
}

export default async function AuthLayout({ children }: { children: React.ReactNode }) {
  const t = await getTranslations('authPages')

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4">
      {/* Theme & Language Toggle */}
      <div className="fixed top-8 right-8 z-50 flex items-center gap-2">
        <LanguageSwitcher />
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
        <p>{t('footer')}</p>
      </footer>
    </div>
  )
}
