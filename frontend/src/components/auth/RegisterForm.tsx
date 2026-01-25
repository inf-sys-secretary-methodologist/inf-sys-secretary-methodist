'use client'

import { useState, useMemo } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { Eye, EyeOff, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { FloatingInput } from '@/components/ui/floating-input'
import { Button } from '@/components/ui/button'
import { useRegister } from '@/hooks/useAuth'
import { createRegisterSchema, type RegisterFormData } from '@/lib/validations/auth'
import { UserRole } from '@/types/auth'
import { cn } from '@/lib/utils'

interface RegisterFormProps {
  redirectTo?: string
  onSuccess?: () => void
  className?: string
}

export function RegisterForm({ redirectTo = '/login', onSuccess, className }: RegisterFormProps) {
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const { register: registerUser, error: authError, clearError } = useRegister()
  const t = useTranslations('registerForm')
  const tAuth = useTranslations('auth')
  const tRoles = useTranslations('roles')
  const tValidation = useTranslations('validation')

  const registerSchema = useMemo(() => createRegisterSchema(tValidation), [tValidation])

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    watch,
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    mode: 'onBlur',
    defaultValues: {
      role: UserRole.STUDENT,
    },
  })

  const password = watch('password')

  /* c8 ignore start - Form submit handler with error handling, tested in e2e */
  const onSubmit = async (data: RegisterFormData) => {
    try {
      clearError()
      await registerUser(
        {
          name: data.name,
          email: data.email,
          password: data.password,
          role: data.role,
        },
        redirectTo
      )

      toast.success(t('registerSuccess'), {
        description: t('redirectingToLogin'),
      })

      if (onSuccess) {
        onSuccess()
      }
    } catch (error: unknown) {
      const errorMessage =
        (error as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        authError ||
        t('registerError')
      toast.error(t('registerError'), {
        description: errorMessage,
      })
      console.error('Registration error:', error)
    }
  }
  /* c8 ignore stop */

  // Password strength indicator
  const getPasswordStrength = (pass: string): number => {
    if (!pass) return 0
    let strength = 0
    if (pass.length >= 8) strength += 25
    if (/[A-Z]/.test(pass)) strength += 25
    if (/[a-z]/.test(pass)) strength += 25
    if (/[0-9]/.test(pass)) strength += 12.5
    if (/[^a-zA-Z0-9]/.test(pass)) strength += 12.5
    return strength
  }

  const passwordStrength = getPasswordStrength(password || '')

  const getPasswordStrengthColor = (strength: number): string => {
    if (strength < 40) return 'bg-red-500'
    if (strength < 70) return 'bg-yellow-500'
    return 'bg-green-500'
  }

  const getPasswordStrengthText = (strength: number): string => {
    if (strength < 40) return t('passwordWeak')
    if (strength < 70) return t('passwordMedium')
    return t('passwordStrong')
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className={cn('space-y-6', className)}>
      {/* Global error message */}
      {authError && (
        <div className="p-4 text-sm text-red-800 bg-red-50 border border-red-200 rounded-lg dark:bg-red-900/20 dark:text-red-400 dark:border-red-800">
          <p>{authError}</p>
        </div>
      )}

      {/* Name field */}
      <div className="space-y-2">
        <FloatingInput
          label={t('name')}
          type="text"
          autoComplete="name"
          disabled={isSubmitting}
          {...register('name')}
          className={cn(errors.name && 'border-red-500 focus-visible:ring-red-500')}
        />
        {errors.name && (
          <p className="text-sm text-red-600 dark:text-red-400">{errors.name.message}</p>
        )}
      </div>

      {/* Email field */}
      <div className="space-y-2">
        <FloatingInput
          label={t('email')}
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
            label={t('password')}
            type={showPassword ? 'text' : 'password'}
            autoComplete="new-password"
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

        {/* Password strength indicator */}
        {password && (
          <div className="space-y-1">
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground">{t('passwordStrength')}</span>
              <span className={cn('font-medium', passwordStrength >= 70 && 'text-green-600')}>
                {getPasswordStrengthText(passwordStrength)}
              </span>
            </div>
            <div className="h-2 bg-gray-200 rounded-full overflow-hidden dark:bg-gray-700">
              <div
                className={cn(
                  'h-full transition-all duration-300',
                  getPasswordStrengthColor(passwordStrength)
                )}
                style={{ width: `${passwordStrength}%` }}
              />
            </div>
          </div>
        )}

        {errors.password && (
          <div className="space-y-1">
            <p className="text-sm font-medium text-red-600 dark:text-red-400">
              {t('passwordRequirements')}
            </p>
            <ul className="text-sm text-red-600 dark:text-red-400 list-disc list-inside space-y-0.5">
              {errors.password.message?.split('. ').map((msg, i) => (
                <li key={i}>{msg}</li>
              ))}
            </ul>
          </div>
        )}
      </div>

      {/* Confirm Password field */}
      <div className="space-y-2">
        <div className="relative">
          <FloatingInput
            label={t('confirmPassword')}
            type={showConfirmPassword ? 'text' : 'password'}
            autoComplete="new-password"
            disabled={isSubmitting}
            {...register('confirmPassword')}
            className={cn(
              'pr-10',
              errors.confirmPassword && 'border-red-500 focus-visible:ring-red-500'
            )}
          />
          <button
            type="button"
            onClick={() => setShowConfirmPassword(!showConfirmPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
            tabIndex={-1}
          >
            {showConfirmPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
          </button>
        </div>
        {errors.confirmPassword && (
          <p className="text-sm text-red-600 dark:text-red-400">{errors.confirmPassword.message}</p>
        )}
      </div>

      {/* Role selection */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-foreground">{t('role')}</label>
        <select
          {...register('role')}
          disabled={isSubmitting}
          className={cn(
            'w-full px-3 pr-10 py-2 text-sm rounded-lg border border-input bg-background',
            'focus:outline-none focus:ring-2 focus:ring-ring focus:border-ring',
            'disabled:cursor-not-allowed disabled:opacity-50',
            'appearance-none cursor-pointer',
            errors.role && 'border-red-500 focus:ring-red-500'
          )}
          style={{
            backgroundImage: `url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e")`,
            backgroundPosition: 'right 0.5rem center',
            backgroundRepeat: 'no-repeat',
            backgroundSize: '1.5em 1.5em',
          }}
        >
          <option value={UserRole.STUDENT}>{tRoles('student')}</option>
          <option value={UserRole.TEACHER}>{tRoles('teacher')}</option>
          <option value={UserRole.ACADEMIC_SECRETARY}>{tRoles('academic_secretary')}</option>
          <option value={UserRole.METHODIST}>{tRoles('methodist')}</option>
          <option value={UserRole.SYSTEM_ADMIN}>{tRoles('system_admin')}</option>
        </select>
        {errors.role && (
          <p className="text-sm text-red-600 dark:text-red-400">{errors.role.message}</p>
        )}
      </div>

      {/* Submit button */}
      <Button type="submit" disabled={isSubmitting} className="w-full" size="lg">
        {isSubmitting ? (
          <>
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            {t('registering')}
          </>
        ) : (
          tAuth('register')
        )}
      </Button>

      {/* Login link */}
      <div className="text-center text-sm">
        <span className="text-muted-foreground">{tAuth('hasAccount')} </span>
        <Link href="/login" className="font-medium text-primary hover:underline">
          {tAuth('login')}
        </Link>
      </div>
    </form>
  )
}
