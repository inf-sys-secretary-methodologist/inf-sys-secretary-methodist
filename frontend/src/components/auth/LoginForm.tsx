'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Eye, EyeOff, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { FloatingInput } from '@/components/ui/floating-input'
import { Button } from '@/components/ui/button'
import { useLogin } from '@/hooks/useAuth'
import { loginSchema, type LoginFormData } from '@/lib/validations/auth'
import { cn } from '@/lib/utils'

interface LoginFormProps {
  redirectTo?: string
  onSuccess?: () => void
  className?: string
}

export function LoginForm({ redirectTo = '/', onSuccess, className }: LoginFormProps) {
  const router = useRouter()
  const [showPassword, setShowPassword] = useState(false)
  const { login, isLoading, error: authError, clearError } = useLogin()

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    mode: 'onBlur',
  })

  const onSubmit = async (data: LoginFormData) => {
    try {
      clearError()
      await login(data, redirectTo)

      toast.success('Вход выполнен успешно!', {
        description: 'Перенаправление...',
      })

      if (onSuccess) {
        onSuccess()
      }
    } catch (error: unknown) {
      const errorMessage = (error as { response?: { data?: { message?: string } } })?.response?.data?.message || authError || 'Ошибка входа'
      toast.error('Ошибка входа', {
        description: errorMessage,
      })
      console.error('Login error:', error)
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className={cn('space-y-6', className)}>
      {/* Global error message */}
      {authError && (
        <div className="p-4 text-sm text-red-800 bg-red-50 border border-red-200 rounded-lg dark:bg-red-900/20 dark:text-red-400 dark:border-red-800">
          <p>{authError}</p>
        </div>
      )}

      {/* Email field */}
      <div className="space-y-2">
        <FloatingInput
          label="Email"
          type="email"
          autoComplete="email"
          disabled={isSubmitting}
          {...register('email')}
          className={cn(
            errors.email && 'border-red-500 focus-visible:ring-red-500'
          )}
        />
        {errors.email && (
          <p className="text-sm text-red-600 dark:text-red-400">
            {errors.email.message}
          </p>
        )}
      </div>

      {/* Password field */}
      <div className="space-y-2">
        <div className="relative">
          <FloatingInput
            label="Пароль"
            type={showPassword ? 'text' : 'password'}
            autoComplete="current-password"
            disabled={isSubmitting}
            {...register('password')}
            className={cn(
              'pr-10',
              errors.password && 'border-red-500 focus-visible:ring-red-500'
            )}
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
            tabIndex={-1}
          >
            {showPassword ? (
              <EyeOff className="w-5 h-5" />
            ) : (
              <Eye className="w-5 h-5" />
            )}
          </button>
        </div>
        {errors.password && (
          <p className="text-sm text-red-600 dark:text-red-400">
            {errors.password.message}
          </p>
        )}
      </div>

      {/* Forgot password link */}
      <div className="flex items-center justify-between">
        <div className="text-sm">
          <Link
            href="/forgot-password"
            className="font-medium text-primary hover:underline"
          >
            Забыли пароль?
          </Link>
        </div>
      </div>

      {/* Submit button */}
      <Button
        type="submit"
        disabled={isSubmitting || isLoading}
        className="w-full"
        size="lg"
      >
        {isSubmitting || isLoading ? (
          <>
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            Вход...
          </>
        ) : (
          'Войти'
        )}
      </Button>

      {/* Register link */}
      <div className="text-center text-sm">
        <span className="text-muted-foreground">Нет аккаунта? </span>
        <Link
          href="/register"
          className="font-medium text-primary hover:underline"
        >
          Зарегистрироваться
        </Link>
      </div>
    </form>
  )
}
