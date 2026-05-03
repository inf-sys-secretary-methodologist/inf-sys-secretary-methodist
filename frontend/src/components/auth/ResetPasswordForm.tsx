'use client'

import { useEffect, useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useRouter, useSearchParams } from 'next/navigation'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { CheckCircle2, Eye, EyeOff, Loader2, AlertTriangle } from 'lucide-react'
import { FloatingInput } from '@/components/ui/floating-input'
import { Button } from '@/components/ui/button'
import { authApi } from '@/lib/api/auth'
import {
  createPasswordResetSchema,
  type PasswordResetFormData,
} from '@/lib/validations/auth'
import { cn } from '@/lib/utils'

type VerifyState = 'verifying' | 'valid' | 'invalid' | 'missing'

interface ResetPasswordFormProps {
  className?: string
}

export function ResetPasswordForm({ className }: ResetPasswordFormProps) {
  const t = useTranslations('resetPasswordPage')
  const tAuth = useTranslations('auth')
  const tValidation = useTranslations('validation')
  const router = useRouter()
  const searchParams = useSearchParams()
  const token = searchParams.get('token') ?? ''

  const [verifyState, setVerifyState] = useState<VerifyState>(token ? 'verifying' : 'missing')
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [submitted, setSubmitted] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)

  const schema = useMemo(() => createPasswordResetSchema(tValidation), [tValidation])

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<PasswordResetFormData>({
    resolver: zodResolver(schema),
    mode: 'onBlur',
  })

  useEffect(() => {
    if (!token) return
    let cancelled = false
    authApi
      .verifyPasswordResetToken(token)
      .then(() => {
        if (!cancelled) setVerifyState('valid')
      })
      .catch(() => {
        if (!cancelled) setVerifyState('invalid')
      })
    return () => {
      cancelled = true
    }
  }, [token])

  const onSubmit = async (data: PasswordResetFormData) => {
    setSubmitError(null)
    try {
      await authApi.confirmPasswordReset(token, data.password)
      setSubmitted(true)
      // Auto-redirect after a beat so the success message has time to read.
      setTimeout(() => router.push('/login'), 2500)
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status
      if (status === 410) {
        setVerifyState('invalid')
        return
      }
      if (status === 400) {
        setSubmitError(t('weakPasswordError'))
        return
      }
      setSubmitError(t('genericError'))
    }
  }

  // --- Token-state branches: render decisively before any form ---

  if (verifyState === 'missing') {
    return (
      <div className={cn('space-y-6 text-center', className)}>
        <AlertTriangle className="w-12 h-12 mx-auto text-amber-600" aria-hidden="true" />
        <p className="text-sm text-muted-foreground">{t('tokenMissing')}</p>
        <Link href="/forgot-password" className="inline-block text-sm font-medium text-primary hover:underline">
          {t('requestNewLink')}
        </Link>
      </div>
    )
  }

  if (verifyState === 'verifying') {
    return (
      <div className={cn('space-y-4 text-center py-8', className)}>
        <Loader2 className="w-8 h-8 mx-auto animate-spin text-primary" aria-hidden="true" />
        <p className="text-sm text-muted-foreground">{t('verifying')}</p>
      </div>
    )
  }

  if (verifyState === 'invalid') {
    return (
      <div className={cn('space-y-6 text-center', className)}>
        <AlertTriangle className="w-12 h-12 mx-auto text-amber-600" aria-hidden="true" />
        <div className="space-y-2">
          <h2 className="text-xl font-semibold">{t('linkExpiredTitle')}</h2>
          <p className="text-sm text-muted-foreground">{t('linkExpiredDescription')}</p>
        </div>
        <Link href="/forgot-password" className="inline-block text-sm font-medium text-primary hover:underline">
          {t('requestNewLink')}
        </Link>
      </div>
    )
  }

  if (submitted) {
    return (
      <div className={cn('space-y-6 text-center', className)}>
        <CheckCircle2 className="w-12 h-12 mx-auto text-green-600" aria-hidden="true" />
        <div className="space-y-2">
          <h2 className="text-xl font-semibold">{t('successTitle')}</h2>
          <p className="text-sm text-muted-foreground">{t('successDescription')}</p>
        </div>
        <Link href="/login" className="inline-block text-sm font-medium text-primary hover:underline">
          {t('goToLogin')}
        </Link>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className={cn('space-y-6', className)}>
      {submitError && (
        <div className="p-4 text-sm text-red-800 bg-red-50 border border-red-200 rounded-lg dark:bg-red-900/20 dark:text-red-400 dark:border-red-800">
          <p>{submitError}</p>
        </div>
      )}

      <div className="space-y-2">
        <div className="relative">
          <FloatingInput
            label={t('newPasswordLabel')}
            type={showPassword ? 'text' : 'password'}
            autoComplete="new-password"
            disabled={isSubmitting}
            {...register('password')}
            className={cn('pr-10', errors.password && 'border-red-500 focus-visible:ring-red-500')}
          />
          <button
            type="button"
            onClick={() => setShowPassword((s) => !s)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
            tabIndex={-1}
            aria-label={tAuth('password')}
          >
            {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
          </button>
        </div>
        {errors.password && (
          <p className="text-sm text-red-600 dark:text-red-400">{errors.password.message}</p>
        )}
      </div>

      <div className="space-y-2">
        <div className="relative">
          <FloatingInput
            label={t('confirmPasswordLabel')}
            type={showConfirmPassword ? 'text' : 'password'}
            autoComplete="new-password"
            disabled={isSubmitting}
            {...register('confirmPassword')}
            className={cn('pr-10', errors.confirmPassword && 'border-red-500 focus-visible:ring-red-500')}
          />
          <button
            type="button"
            onClick={() => setShowConfirmPassword((s) => !s)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
            tabIndex={-1}
            aria-label={tAuth('confirmPassword')}
          >
            {showConfirmPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
          </button>
        </div>
        {errors.confirmPassword && (
          <p className="text-sm text-red-600 dark:text-red-400">{errors.confirmPassword.message}</p>
        )}
      </div>

      <Button type="submit" disabled={isSubmitting} className="w-full" size="lg">
        {isSubmitting ? (
          <>
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            {t('submitting')}
          </>
        ) : (
          t('submit')
        )}
      </Button>
    </form>
  )
}
