'use client'

import { useEffect } from 'react'
import { AlertCircle, Home, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useTranslations } from 'next-intl'
import * as Sentry from '@sentry/nextjs'

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  const t = useTranslations('errorPages.error')

  useEffect(() => {
    // Log error to console (can be extended to send to error tracking service)
    console.error('Application error:', {
      message: error.message,
      digest: error.digest,
      stack: error.stack,
      timestamp: new Date().toISOString(),
    })

    // Send to Sentry error tracking service
    Sentry.captureException(error, {
      level: 'error',
      tags: {
        errorBoundary: 'app-error',
      },
      contexts: {
        errorInfo: {
          digest: error.digest,
          timestamp: new Date().toISOString(),
        },
      },
    })
  }, [error])

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <div className="max-w-md w-full space-y-6 text-center">
        {/* Error Icon */}
        <div className="flex justify-center">
          <div className="rounded-full bg-destructive/10 p-6">
            <AlertCircle className="h-16 w-16 text-destructive" />
          </div>
        </div>

        {/* Error Title */}
        <div className="space-y-2">
          <h1 className="text-4xl font-bold tracking-tight">{t('title')}</h1>
          <p className="text-muted-foreground">{t('description')}</p>
        </div>

        {/* Error Message (Development mode or non-sensitive errors) */}
        {process.env.NODE_ENV === 'development' && error.message && (
          <div className="p-4 rounded-lg bg-destructive/10 border border-destructive/20 text-left">
            <p className="text-sm font-semibold text-destructive mb-2">{t('details')}</p>
            <p className="text-xs text-destructive/80 font-mono break-words">{error.message}</p>
            {error.digest && (
              <p className="text-xs text-muted-foreground mt-2">Error ID: {error.digest}</p>
            )}
          </div>
        )}

        {/* Production error digest */}
        {process.env.NODE_ENV === 'production' && error.digest && (
          <div className="p-4 rounded-lg bg-muted border text-left">
            <p className="text-sm text-muted-foreground">{t('supportCode')}</p>
            <p className="text-xs font-mono text-foreground mt-1">{error.digest}</p>
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row gap-3 justify-center">
          <Button onClick={reset} variant="default" className="gap-2">
            <RefreshCw className="h-4 w-4" />
            {t('retry')}
          </Button>
          <Button
            onClick={() => (window.location.href = '/dashboard')}
            variant="outline"
            className="gap-2"
          >
            <Home className="h-4 w-4" />
            {t('backHome')}
          </Button>
        </div>
      </div>
    </div>
  )
}
