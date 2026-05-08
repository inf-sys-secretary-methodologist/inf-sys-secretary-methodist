'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { ShieldCheck, ShieldOff } from 'lucide-react'
import { toast } from 'sonner'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useMFA } from '@/hooks/useMFA'

interface MFASettingsCardProps {
  mfaEnabled: boolean
  onChange?: (enabled: boolean) => void
}

type Step = 'idle' | 'enrolling' | 'disabling'

const CODE_PATTERN = /^\d{6}$/

export function MFASettingsCard({ mfaEnabled, onChange }: MFASettingsCardProps) {
  // Namespace mirrors the JSON tree: messages/{locale}.json :
  //   adminSettings.security.mfa.{enable, disable, codeLabel, ...}
  // so each `t('mfa.x')` resolves to that path. Keep this in sync with
  // the JSON-key parity test in __tests__/MFASettingsCard.i18n.test.ts.
  const t = useTranslations('adminSettings.security')
  const { beginEnrollment, confirmEnrollment, disable } = useMFA()

  const [enabled, setEnabled] = useState(mfaEnabled)
  const [step, setStep] = useState<Step>('idle')
  const [secret, setSecret] = useState('')
  const [code, setCode] = useState('')
  const [busy, setBusy] = useState(false)

  const codeValid = CODE_PATTERN.test(code)

  const handleEnable = async () => {
    setBusy(true)
    try {
      const resp = await beginEnrollment()
      setSecret(resp.secret)
      setStep('enrolling')
      setCode('')
    } catch {
      toast.error(t('mfa.errorBegin'))
    } finally {
      setBusy(false)
    }
  }

  const handleConfirm = async () => {
    if (!codeValid) return
    setBusy(true)
    try {
      await confirmEnrollment(code)
      setEnabled(true)
      setStep('idle')
      setSecret('')
      setCode('')
      onChange?.(true)
      toast.success(t('mfa.successEnable'))
    } catch {
      toast.error(t('mfa.errorConfirm'))
    } finally {
      setBusy(false)
    }
  }

  const handleStartDisable = () => {
    setStep('disabling')
    setCode('')
  }

  const handleConfirmDisable = async () => {
    if (!codeValid) return
    setBusy(true)
    try {
      await disable(code)
      setEnabled(false)
      setStep('idle')
      setCode('')
      onChange?.(false)
      toast.success(t('mfa.successDisable'))
    } catch {
      toast.error(t('mfa.errorDisable'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {enabled ? (
            <ShieldCheck className="h-5 w-5 text-emerald-600" />
          ) : (
            <ShieldOff className="h-5 w-5 text-muted-foreground" />
          )}
          {t('mfa.title')}
        </CardTitle>
        <CardDescription>
          {enabled ? t('mfa.descriptionEnabled') : t('mfa.descriptionDisabled')}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {step === 'idle' && !enabled && (
          <Button onClick={handleEnable} disabled={busy}>
            {t('mfa.enable')}
          </Button>
        )}

        {step === 'idle' && enabled && (
          <Button variant="destructive" onClick={handleStartDisable} disabled={busy}>
            {t('mfa.disable')}
          </Button>
        )}

        {step === 'enrolling' && (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">{t('mfa.scanInstruction')}</p>
            <div className="rounded-md border bg-muted px-3 py-2 font-mono text-sm break-all">
              {secret}
            </div>
            <div className="space-y-2">
              <Label htmlFor="mfa-confirm-code">{t('mfa.codeLabel')}</Label>
              <Input
                id="mfa-confirm-code"
                value={code}
                inputMode="numeric"
                maxLength={6}
                onChange={(e) => setCode(e.target.value)}
                placeholder="123456"
              />
            </div>
            <Button onClick={handleConfirm} disabled={!codeValid || busy}>
              {t('mfa.confirm')}
            </Button>
          </div>
        )}

        {step === 'disabling' && (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">{t('mfa.disableInstruction')}</p>
            <div className="space-y-2">
              <Label htmlFor="mfa-disable-code">{t('mfa.codeLabel')}</Label>
              <Input
                id="mfa-disable-code"
                value={code}
                inputMode="numeric"
                maxLength={6}
                onChange={(e) => setCode(e.target.value)}
                placeholder="123456"
              />
            </div>
            <Button
              variant="destructive"
              onClick={handleConfirmDisable}
              disabled={!codeValid || busy}
            >
              {t('mfa.confirmDisable')}
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
