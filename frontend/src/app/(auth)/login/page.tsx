import type { Metadata } from 'next'
import Link from 'next/link'
import { getTranslations } from 'next-intl/server'
import { LoginForm } from '@/components/auth/LoginForm'
import { BrandedHeader } from '@/components/branding/BrandedHeader'

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
      {/* Branded header — client component reads /api/public/branding
          to render the configured app name, optional logo, and
          tagline. Falls back to the translated authPages.loginWelcome
          key while loading or if the fetch fails. */}
      <BrandedHeader titleFallback="authPages.loginWelcome" />
      <p className="text-center text-sm text-muted-foreground">{t('loginSubtitle')}</p>

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
