'use client'

import { useTranslations } from 'next-intl'
import Image from 'next/image'

import { useBranding } from '@/hooks/useBranding'

interface BrandedHeaderProps {
  // titleFallback — i18n key fallback used while the public
  // branding fetch is loading or если it fails. Caller supplies
  // an appropriate key for the surrounding page context.
  titleFallback: string
}

// BrandedHeader is a client component rendering the runtime
// brand chrome (logo + app name + tagline) on the login page
// and other auth surfaces. Graceful degradation: shows the
// fallback title during loading and on fetch failure so the
// auth chrome is never blank.
export function BrandedHeader({ titleFallback }: BrandedHeaderProps) {
  const t = useTranslations()
  const { config, isLoading } = useBranding({ public: true })

  const appName = (!isLoading && config?.app_name) || t(titleFallback)
  const tagline = config?.tagline ?? ''
  const logoURL = config?.logo_url ?? ''
  const primaryColor = config?.primary_color ?? ''

  return (
    <div
      data-testid="branded-header"
      className="text-center space-y-3"
      style={
        primaryColor ? { borderTop: `3px solid ${primaryColor}`, paddingTop: '0.75rem' } : undefined
      }
    >
      {logoURL ? (
        // Logo is configured — render via next/image. The URL has
        // already passed the backend's http/https scheme whitelist
        // (defense-in-depth), но мы keep alt='' decorative since
        // the app name renders below as the page title.
        <div data-testid="branded-logo" className="flex justify-center">
          <Image src={logoURL} alt="" width={64} height={64} className="rounded-md" unoptimized />
        </div>
      ) : null}
      <h1
        data-testid="branded-title"
        className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white"
      >
        {appName}
      </h1>
      {tagline ? (
        <p data-testid="branded-tagline" className="text-sm text-muted-foreground">
          {tagline}
        </p>
      ) : null}
    </div>
  )
}
