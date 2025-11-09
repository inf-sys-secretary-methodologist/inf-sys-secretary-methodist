import type { Metadata } from 'next'
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
      <RegisterForm redirectTo="/login" />
    </div>
  )
}
