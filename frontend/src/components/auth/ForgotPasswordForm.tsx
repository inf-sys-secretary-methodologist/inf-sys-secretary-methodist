'use client'

import { useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { CheckCircle2, Loader2 } from 'lucide-react'
import { FloatingInput } from '@/components/ui/floating-input'
import { Button } from '@/components/ui/button'
import { authApi } from '@/lib/api/auth'
import {
  createPasswordRecoverySchema,
  type PasswordRecoveryFormData,
} from '@/lib/validations/auth'
import { cn } from '@/lib/utils'

interface ForgotPasswordFormProps {
  className?: string
}

export function ForgotPasswordForm({ className }: ForgotPasswordFormProps) {
  const t = useTranslations('forgotPasswordPage')
  const tAuth = useTranslations('auth')
  const tValidation = useTranslations('validation')
  const [submitted, setSubmitted] = useState(false)
  const [genericError, setGenericError] = useState<string | null>(null)

  const schema = useMemo(() => createPasswordRecoverySchema(tValidation), [tValidation])

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<PasswordRecoveryFormData>({
    resolver: zodResolver(schema),
    mode: 'onBlur',
  })

  const onSubmit = async (data: PasswordRecoveryFormData) => {
    setGenericError(null)
    try {
      await authApi.requestPasswordReset(data.email)
      // Backend honors anti-enumeration — we render the same success
      // copy regardless of whether the email exists.
      setSubmitted(true)
    } catch {
      setGenericError(t('genericError'))
    }
  }

  if (submitted) {
    return (
      <div className={cn('space-y-6 text-center', className)}>
        <CheckCircle2 className="w-12 h-12 mx-auto text-green-600" aria-hidden="true" />
        <div className="space-y-2">
          <h2 className="text-xl font-semibold">{t('successTitle')}</h2>
          <p className="text-sm text-muted-foreground">{t('successDescription')}</p>
        </div>
        <Link
          href="/login"
          className="inline-block text-sm font-medium text-primary hover:underline"
        >
          {t('backToLogin')}
        </Link>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className={cn('space-y-6', className)}>
      {genericError && (
        <div className="p-4 text-sm text-red-800 bg-red-50 border border-red-200 rounded-lg dark:bg-red-900/20 dark:text-red-400 dark:border-red-800">
          <p>{genericError}</p>
        </div>
      )}

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

      <div className="text-center text-sm">
        <Link href="/login" className="font-medium text-muted-foreground hover:text-primary transition-colors">
          {t('backToLogin')}
        </Link>
      </div>
    </form>
  )
}
