import type { Metadata } from 'next'
import Link from 'next/link'
import { getTranslations } from 'next-intl/server'
import { RegisterForm } from '@/components/auth/RegisterForm'

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('auth')
  return {
    title: t('registerTitle'),
    description: t('register'),
  }
}

export default async function RegisterPage() {
  const t = await getTranslations('authPages')

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
          {t('registerTitle')}
        </h1>
        <p className="text-sm text-muted-foreground">{t('registerSubtitle')}</p>
      </div>

      {/* Register Form */}
      <RegisterForm redirectTo="/dashboard" />

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
