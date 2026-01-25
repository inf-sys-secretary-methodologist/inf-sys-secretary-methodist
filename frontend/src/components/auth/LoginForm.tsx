'use client'

import { useState, useMemo } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { Eye, EyeOff, Loader2 } from 'lucide-react'
import { toast } from '@/components/providers/toaster-provider'
import { FloatingInput } from '@/components/ui/floating-input'
import { Button } from '@/components/ui/button'
import { useLogin } from '@/hooks/useAuth'
import { createLoginSchema, type LoginFormData } from '@/lib/validations/auth'
import { cn } from '@/lib/utils'

interface LoginFormProps {
  redirectTo?: string
  onSuccess?: () => void
  className?: string
}

export function LoginForm({ redirectTo = '/', onSuccess, className }: LoginFormProps) {
  const [showPassword, setShowPassword] = useState(false)
  const [localError, setLocalError] = useState<string | null>(null)
  const { login, error: authError, clearError } = useLogin()
  const t = useTranslations('loginForm')
  const tAuth = useTranslations('auth')
  const tValidation = useTranslations('validation')

  const loginSchema = useMemo(() => createLoginSchema(tValidation), [tValidation])

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    mode: 'onBlur',
  })

  const onSubmit = async (data: LoginFormData) => {
    try {
      clearError()
      setLocalError(null)
      await login(data, redirectTo)

      toast.success(t('loginSuccess'), {
        description: t('redirecting'),
      })

      if (onSuccess) {
        onSuccess()
      }
      /* c8 ignore start - Error handling with message extraction */
    } catch (error: unknown) {
      const rawMessage =
        (error as { response?: { data?: { error?: { message?: string }; message?: string } } })
          ?.response?.data?.error?.message ||
        (error as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        ''

      // Use translated error message
      const errorMessage =
        rawMessage === 'Unauthorized access' || rawMessage === ''
          ? t('invalidCredentials')
          : rawMessage

      // Set local error state for immediate feedback
      setLocalError(errorMessage)
      /* c8 ignore stop */

      // Show toast with unique ID and longer duration to prevent auto-dismissal
      toast.error(t('loginError'), {
        id: 'login-error',
        description: errorMessage,
        duration: 10000, // 10 seconds
      })
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className={cn('space-y-6', className)}>
      {/* Global error message */}
      {(localError || authError) && (
        <div className="p-4 text-sm text-red-800 bg-red-50 border border-red-200 rounded-lg dark:bg-red-900/20 dark:text-red-400 dark:border-red-800">
          <p>{localError || authError}</p>
        </div>
      )}

      {/* Email field */}
      <div className="space-y-2">
        <FloatingInput
          label={tAuth('email')}
          type="email"
          autoComplete="email"
          disabled={isSubmitting}
          {...register('email')}
          className={cn(errors.email && 'border-red-500 focus-visible:ring-red-500')}
        />
        {errors.email && (
          <p className="text-sm text-red-600 dark:text-red-400">{errors.email.message}</p>
        )}
      </div>

      {/* Password field */}
      <div className="space-y-2">
        <div className="relative">
          <FloatingInput
            label={tAuth('password')}
            type={showPassword ? 'text' : 'password'}
            autoComplete="current-password"
            disabled={isSubmitting}
            {...register('password')}
            className={cn('pr-10', errors.password && 'border-red-500 focus-visible:ring-red-500')}
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
            tabIndex={-1}
          >
            {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
          </button>
        </div>
        {errors.password && (
          <p className="text-sm text-red-600 dark:text-red-400">{errors.password.message}</p>
        )}
      </div>

      {/* Forgot password link */}
      <div className="flex items-center justify-between">
        <div className="text-sm">
          <Link href="/forgot-password" className="font-medium text-primary hover:underline">
            {tAuth('forgotPassword')}
          </Link>
        </div>
      </div>

      {/* Submit button */}
      <Button type="submit" disabled={isSubmitting} className="w-full" size="lg">
        {isSubmitting ? (
          <>
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            {t('loggingIn')}
          </>
        ) : (
          tAuth('login')
        )}
      </Button>

      {/* Register link */}
      <div className="text-center text-sm">
        <span className="text-muted-foreground">{tAuth('noAccount')} </span>
        <Link href="/register" className="font-medium text-primary hover:underline">
          {tAuth('register')}
        </Link>
      </div>
    </form>
  )
}
