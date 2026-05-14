'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { AlertCircle, CheckCircle2, Image as ImageIcon, Loader2 } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { useAuthCheck } from '@/hooks/useAuth'
import { useBranding, useUpdateBranding } from '@/hooks/useBranding'
import type { BrandSettings, UpdateBrandingRequest } from '@/types/branding'

// AdminBrandingPage — system_admin-only form для editing the
// brand_settings singleton. Surfaces all 6 fields с native
// color pickers + URL textboxes; PUT mutates and re-populates
// the form on 200. Domain validation errors map 422 → typed
// field error.
export default function AdminBrandingPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading: authLoading } = useAuthCheck()
  const t = useTranslations('adminBranding')

  const enabled = !authLoading && isAuthenticated && user?.role === 'system_admin'
  const { config, isLoading: configLoading, error, mutate } = useBranding({ enabled })
  const { updateBranding, isLoading: saving, errorCode } = useUpdateBranding()

  useEffect(() => {
    if (!authLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [authLoading, isAuthenticated, user, router])

  return (
    <AppLayout>
      <div data-testid="admin-branding-page" className="max-w-3xl mx-auto space-y-6">
        <header className="flex items-center gap-3">
          <ImageIcon className="h-7 w-7" />
          <div className="flex-1">
            <h1 className="text-2xl font-bold">{t('title')}</h1>
            <p className="text-sm text-muted-foreground">{t('description')}</p>
          </div>
        </header>

        {configLoading ? (
          <div data-testid="branding-loading" className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div
            data-testid="branding-error"
            className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center"
          >
            <p className="font-medium text-destructive">{t('loadFailed')}</p>
          </div>
        ) : config ? (
          <BrandingForm
            initial={config}
            saving={saving}
            errorCode={errorCode}
            onSubmit={async (req) => {
              const updated = await updateBranding(req)
              await mutate(updated, false)
            }}
          />
        ) : null}
      </div>
    </AppLayout>
  )
}

interface BrandingFormProps {
  initial: BrandSettings
  saving: boolean
  errorCode: string | null
  onSubmit: (req: UpdateBrandingRequest) => Promise<void>
}

function BrandingForm({ initial, saving, errorCode, onSubmit }: BrandingFormProps) {
  const t = useTranslations('adminBranding')
  const [appName, setAppName] = useState(initial.app_name)
  const [tagline, setTagline] = useState(initial.tagline)
  const [logoURL, setLogoURL] = useState(initial.logo_url)
  const [faviconURL, setFaviconURL] = useState(initial.favicon_url)
  const [primaryColor, setPrimaryColor] = useState(initial.primary_color)
  const [secondaryColor, setSecondaryColor] = useState(initial.secondary_color)
  const [successOpen, setSuccessOpen] = useState(false)

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setSuccessOpen(false)
    try {
      await onSubmit({
        app_name: appName,
        tagline,
        logo_url: logoURL,
        favicon_url: faviconURL,
        primary_color: primaryColor,
        secondary_color: secondaryColor,
      })
      setSuccessOpen(true)
    } catch {
      // Error state surfaced via errorCode prop — form stays
      // populated так admin сможет fix and retry.
    }
  }

  const errorMessage = errorCode ? t(`errors.${errorCode}`) : null

  return (
    <form data-testid="branding-form" onSubmit={handleSubmit} className="space-y-5">
      {successOpen ? (
        <div
          data-testid="branding-success"
          className="rounded-lg border border-green-200 bg-green-50 dark:bg-green-900/20 dark:border-green-800 p-3 flex items-center gap-2 text-green-800 dark:text-green-200"
        >
          <CheckCircle2 className="h-4 w-4" />
          <span className="text-sm font-medium">{t('savedSuccess')}</span>
        </div>
      ) : null}

      {errorMessage ? (
        <div
          data-testid="branding-save-error"
          className="rounded-lg border border-destructive/30 bg-destructive/5 p-3 flex items-center gap-2"
        >
          <AlertCircle className="h-4 w-4 text-destructive" />
          <span className="text-sm font-medium text-destructive">{errorMessage}</span>
        </div>
      ) : null}

      <Field label={t('fields.appName')} required>
        <input
          data-testid="branding-input-app-name"
          type="text"
          value={appName}
          maxLength={100}
          onChange={(e) => setAppName(e.target.value)}
          required
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
      </Field>

      <Field label={t('fields.tagline')}>
        <textarea
          data-testid="branding-input-tagline"
          value={tagline}
          maxLength={200}
          onChange={(e) => setTagline(e.target.value)}
          rows={2}
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
      </Field>

      <Field label={t('fields.logoURL')}>
        <input
          data-testid="branding-input-logo-url"
          type="url"
          value={logoURL}
          onChange={(e) => setLogoURL(e.target.value)}
          placeholder={t('placeholders.logoURL')}
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
      </Field>

      <Field label={t('fields.faviconURL')}>
        <input
          data-testid="branding-input-favicon-url"
          type="url"
          value={faviconURL}
          onChange={(e) => setFaviconURL(e.target.value)}
          placeholder={t('placeholders.faviconURL')}
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
      </Field>

      <Field label={t('fields.primaryColor')}>
        <div className="flex items-center gap-2">
          <input
            data-testid="branding-input-primary-color"
            type="text"
            value={primaryColor}
            onChange={(e) => setPrimaryColor(e.target.value)}
            placeholder={t('placeholders.primaryColor')}
            className="flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm font-mono focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
          <input
            data-testid="branding-picker-primary-color"
            type="color"
            value={primaryColor || '#000000'}
            onChange={(e) => setPrimaryColor(e.target.value)}
            className="h-10 w-10 cursor-pointer rounded border border-input bg-background"
            aria-label={t('fields.primaryColor')}
          />
        </div>
      </Field>

      <Field label={t('fields.secondaryColor')}>
        <div className="flex items-center gap-2">
          <input
            data-testid="branding-input-secondary-color"
            type="text"
            value={secondaryColor}
            onChange={(e) => setSecondaryColor(e.target.value)}
            placeholder={t('placeholders.secondaryColor')}
            className="flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm font-mono focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
          <input
            data-testid="branding-picker-secondary-color"
            type="color"
            value={secondaryColor || '#000000'}
            onChange={(e) => setSecondaryColor(e.target.value)}
            className="h-10 w-10 cursor-pointer rounded border border-input bg-background"
            aria-label={t('fields.secondaryColor')}
          />
        </div>
      </Field>

      <div className="flex items-center justify-end pt-2">
        <button
          data-testid="branding-submit"
          type="submit"
          disabled={saving}
          className="inline-flex items-center justify-center rounded-md bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed px-4 py-2 text-sm font-medium"
        >
          {saving ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              {t('saving')}
            </>
          ) : (
            t('save')
          )}
        </button>
      </div>
    </form>
  )
}

function Field({
  label,
  required,
  children,
}: {
  label: string
  required?: boolean
  children: React.ReactNode
}) {
  return (
    <label className="block space-y-1">
      <span className="text-sm font-medium">
        {label}
        {required ? <span className="text-destructive ml-0.5">*</span> : null}
      </span>
      {children}
    </label>
  )
}
