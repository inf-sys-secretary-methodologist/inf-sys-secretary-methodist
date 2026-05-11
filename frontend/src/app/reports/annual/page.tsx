'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { Download, Loader2 } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { useAuthCheck } from '@/hooks/useAuth'
import { annualReportApi } from '@/lib/api/annualReport'

const allowedRoles = ['methodist', 'system_admin'] as const

// AnnualReportPage — read-only download UI for the methodist annual
// report. Year selector covers the last 10 calendar years (matches the
// backend [2000, 2100] guard with reasonable UX bound). Role gate
// redirects to /forbidden mirroring the AdminCurriculumApprovePage
// pattern (single-role allowlist applied here as multi-role: methodist
// + system_admin per ADR-6).
export default function AnnualReportPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('reports.annual')

  const currentYear = new Date().getFullYear()
  const years = useMemo(() => Array.from({ length: 10 }, (_, i) => currentYear - i), [currentYear])
  const [year, setYear] = useState(currentYear)
  const [downloading, setDownloading] = useState(false)
  const [errorKey, setErrorKey] = useState<string | null>(null)

  useEffect(() => {
    if (isLoading || !isAuthenticated) return
    if (!user || !allowedRoles.includes(user.role as (typeof allowedRoles)[number])) {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  const handleDownload = async () => {
    setDownloading(true)
    setErrorKey(null)
    try {
      const blob = await annualReportApi.download(year)
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `annual_report_${year}.docx`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
    } catch {
      setErrorKey('errorServer')
    } finally {
      setDownloading(false)
    }
  }

  if (isLoading) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </AppLayout>
    )
  }

  return (
    <AppLayout>
      <div className="max-w-2xl mx-auto space-y-6 p-6">
        <header>
          <h1 className="text-2xl font-bold">{t('title')}</h1>
        </header>

        <div className="space-y-3">
          <label htmlFor="annual-report-year" className="block text-sm font-medium">
            {t('yearLabel')}
          </label>
          <select
            id="annual-report-year"
            value={year}
            onChange={(e) => setYear(Number(e.target.value))}
            className="w-full rounded-md border border-input bg-background px-3 py-2"
            disabled={downloading}
          >
            {years.map((y) => (
              <option key={y} value={y}>
                {y}
              </option>
            ))}
          </select>
        </div>

        <Button onClick={handleDownload} disabled={downloading}>
          {downloading ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <Download className="mr-2 h-4 w-4" />
          )}
          {t('downloadButton')}
        </Button>

        {errorKey && (
          <p role="alert" className="text-sm text-destructive">
            {t(errorKey)}
          </p>
        )}
      </div>
    </AppLayout>
  )
}
