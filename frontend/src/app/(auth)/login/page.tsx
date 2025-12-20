import type { Metadata } from 'next'
import Link from 'next/link'
import { getTranslations } from 'next-intl/server'
import { LoginForm } from '@/components/auth/LoginForm'

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('auth')
  return {
    title: t('loginTitle'),
    description: t('loginToSystem'),
  }
}

export default async function LoginPage() {
  const t = await getTranslations('authPages')

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
          {t('loginWelcome')}
        </h1>
        <p className="text-sm text-muted-foreground">{t('loginSubtitle')}</p>
      </div>

      {/* Login Form */}
      <LoginForm redirectTo="/dashboard" />

      {/* Back to home link */}
      <div className="text-center text-sm">
        <Link
          href="/"
          className="font-medium text-muted-foreground hover:text-primary transition-colors"
        >
          {t('backToHome')}
        </Link>
      </div>
    </div>
  )
}
