'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuthStore } from '@/stores/authStore'

interface MFAVerifyLoginStepProps {
  redirectTo?: string
}

// Backend rejects non-numeric and non-6-digit bodies at the binding
// layer, but we mirror the same constraint here so the submit button
// stays disabled until the input is presentable.
const CODE_PATTERN = /^\d{6}$/

export function MFAVerifyLoginStep({ redirectTo = '/' }: MFAVerifyLoginStepProps) {
  const t = useTranslations('loginForm')
  const router = useRouter()

  const intermediateToken = useAuthStore((s) => s.mfaIntermediateToken)
  const verifyLoginMFA = useAuthStore((s) => s.verifyLoginMFA)
  const clearMFAChallenge = useAuthStore((s) => s.clearMFAChallenge)
  const authError = useAuthStore((s) => s.error)

  const [code, setCode] = useState('')
  const [busy, setBusy] = useState(false)

  // Defence-in-depth: the parent (LoginForm) only mounts this step
  // when an intermediate is held, but if state changes underneath
  // (logout, clearMFAChallenge from another tab) we render nothing
  // rather than POST a stale payload.
  if (!intermediateToken) return null

  const codeValid = CODE_PATTERN.test(code)

  const handleSubmit = async () => {
    if (!codeValid || busy) return
    setBusy(true)
    try {
      await verifyLoginMFA(code)
      // Match the LoginForm cookie-write delay so downstream guards
      // see the new auth cookie before the route change.
      await new Promise((resolve) => setTimeout(resolve, 100))
      router.push(redirectTo)
    } catch {
      // The store has already written the backend message to
      // state.error. Stay on the step so the user can retry.
    } finally {
      setBusy(false)
    }
  }

  const handleLoginAgain = () => {
    setCode('')
    clearMFAChallenge()
  }

  return (
    <div className="space-y-6" data-testid="mfa-verify-login-step">
      <div className="space-y-1">
        <h2 className="text-lg font-semibold">{t('mfaPrompt.title')}</h2>
        <p className="text-sm text-muted-foreground">{t('mfaPrompt.subtitle')}</p>
      </div>

      {authError && (
        <div className="p-4 text-sm text-red-800 bg-red-50 border border-red-200 rounded-lg dark:bg-red-900/20 dark:text-red-400 dark:border-red-800">
          <p>{authError}</p>
        </div>
      )}

      <div className="space-y-2">
        <Label htmlFor="mfa-login-code">{t('mfaPrompt.codeLabel')}</Label>
        <Input
          id="mfa-login-code"
          value={code}
          inputMode="numeric"
          autoComplete="one-time-code"
          maxLength={6}
          onChange={(e) => setCode(e.target.value)}
          placeholder="123456"
          disabled={busy}
        />
      </div>

      <p className="text-xs text-muted-foreground">{t('mfaPrompt.resendNote')}</p>

      <div className="space-y-3">
        <Button
          type="button"
          onClick={handleSubmit}
          disabled={!codeValid || busy}
          className="w-full"
          size="lg"
        >
          {busy ? (
            <>
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              {t('mfaPrompt.submit')}
            </>
          ) : (
            t('mfaPrompt.submit')
          )}
        </Button>
        <Button
          type="button"
          variant="ghost"
          onClick={handleLoginAgain}
          disabled={busy}
          className="w-full"
        >
          {t('mfaPrompt.loginAgain')}
        </Button>
      </div>
    </div>
  )
}
