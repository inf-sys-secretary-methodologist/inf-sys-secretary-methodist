import type { Metadata } from 'next'
import Link from 'next/link'
import { RegisterForm } from '@/components/auth/RegisterForm'

export const metadata: Metadata = {
  title: 'Регистрация',
  description: 'Создайте новую учетную запись',
}

export default function RegisterPage() {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
          Создать аккаунт
        </h1>
        <p className="text-sm text-muted-foreground">
          Зарегистрируйтесь для доступа к системе
        </p>
      </div>

      {/* Register Form */}
      <RegisterForm redirectTo="/dashboard" />

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
