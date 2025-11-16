import type { Metadata } from 'next'
import Link from 'next/link'
import { LoginForm } from '@/components/auth/LoginForm'

export const metadata: Metadata = {
  title: 'Вход',
  description: 'Войдите в свою учетную запись',
}

export default function LoginPage() {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
          Добро пожаловать
        </h1>
        <p className="text-sm text-muted-foreground">
          Войдите в свою учетную запись для продолжения
        </p>
      </div>

      {/* Login Form */}
      <LoginForm redirectTo="/dashboard" />

      {/* Back to home link */}
      <div className="text-center text-sm">
        <Link
          href="/"
          className="font-medium text-muted-foreground hover:text-primary transition-colors"
        >
          ← Вернуться на главную
        </Link>
      </div>
    </div>
  )
}
