'use client'

import { useEffect, useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import axios from 'axios'
import { toast } from 'sonner'
import {
  ArrowLeft,
  Loader2,
  PenLine,
  Send,
  Clock,
  CheckCircle2,
  Archive,
  BookMarked,
} from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { useCurriculum, submitCurriculum } from '@/hooks/useCurricula'
import { useAuthCheck } from '@/hooks/useAuth'
import { EditCurriculumDialog } from '@/components/curriculum/EditCurriculumDialog'
import type { CurriculumStatus } from '@/types/curriculum'
import { cn } from '@/lib/utils'

// statusKey collapses the wire-format 'pending_approval' to the
// shorter UI key 'pending' (matches CurriculumCard convention).
function statusKey(status: CurriculumStatus): string {
  return status === 'pending_approval' ? 'pending' : status
}

const STATUS_PILL: Record<CurriculumStatus, { bg: string; text: string; Icon: typeof Clock }> = {
  draft: {
    bg: 'bg-slate-100 dark:bg-slate-800/40',
    text: 'text-slate-700 dark:text-slate-300',
    Icon: PenLine,
  },
  pending_approval: {
    bg: 'bg-amber-50 dark:bg-amber-950/30',
    text: 'text-amber-700 dark:text-amber-300',
    Icon: Clock,
  },
  approved: {
    bg: 'bg-emerald-50 dark:bg-emerald-950/30',
    text: 'text-emerald-700 dark:text-emerald-300',
    Icon: CheckCircle2,
  },
  archived: {
    bg: 'bg-zinc-100 dark:bg-zinc-800/40',
    text: 'text-zinc-600 dark:text-zinc-400',
    Icon: Archive,
  },
}

// CurriculumDetailPage — single-curriculum view. Status='draft' enables
// Edit dialog + Submit button; other statuses are read-only with a
// status hint explaining why. Mirrors /assignments/[id] page-shell
// guard order (auth → fetch → render) plus the FetchOpts.enabled SEC
// pattern from v0.114.0.
export default function CurriculumDetailPage() {
  const router = useRouter()
  const params = useParams<{ id: string }>()
  const id = useMemo(() => {
    const raw = params?.id
    const parsed = typeof raw === 'string' ? Number(raw) : NaN
    return Number.isInteger(parsed) && parsed > 0 ? parsed : null
  }, [params])

  const { user, isAuthenticated, isLoading: authLoading } = useAuthCheck()
  const t = useTranslations('curriculum')
  const [editOpen, setEditOpen] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const enabled = !authLoading && isAuthenticated && user?.role !== 'student' && id !== null
  const {
    curriculum,
    isLoading: detailLoading,
    error,
    mutate,
  } = useCurriculum(id, {
    enabled,
  })

  useEffect(() => {
    if (!authLoading && isAuthenticated && user?.role === 'student') {
      router.replace('/forbidden')
    }
  }, [authLoading, isAuthenticated, user, router])

  const handleSubmit = async () => {
    if (id == null || submitting) return
    setSubmitting(true)
    try {
      await submitCurriculum(id)
      toast.success(t('submitToast.success'))
      mutate()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 422:
          key = 'submitToast.errors.notDraft'
          break
        case 403:
          key = 'submitToast.errors.forbidden'
          break
        default:
          key = 'submitToast.errors.generic'
      }
      toast.error(t(key))
    } finally {
      setSubmitting(false)
    }
  }

  if (authLoading || !isAuthenticated) {
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
      <div className="max-w-4xl mx-auto space-y-6">
        <Link
          href="/curriculum"
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          {t('detail.backToList')}
        </Link>

        {id == null ? (
          <div className="rounded-xl border border-border bg-card p-6 text-center">
            <p className="font-medium">{t('detail.notFound')}</p>
          </div>
        ) : detailLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error || !curriculum ? (
          <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center">
            <p className="font-medium text-destructive">{t('detail.loadFailed')}</p>
          </div>
        ) : (
          <>
            <header className="space-y-3">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <h1 className="text-2xl font-bold">{curriculum.title}</h1>
                <StatusPill status={curriculum.status} t={t} />
              </div>
              <dl className="flex flex-wrap gap-x-6 gap-y-1.5 text-sm text-muted-foreground">
                <div className="inline-flex items-center gap-1.5">
                  <BookMarked className="h-3.5 w-3.5" />
                  <span>{curriculum.code}</span>
                </div>
                <div>
                  <dt className="sr-only">{t('detail.fields.specialty')}</dt>
                  <dd>{curriculum.specialty}</dd>
                </div>
                <div>
                  <dt className="sr-only">{t('detail.fields.year')}</dt>
                  <dd>{curriculum.year}</dd>
                </div>
              </dl>
              {curriculum.description && (
                <p className="text-sm whitespace-pre-wrap">{curriculum.description}</p>
              )}
            </header>

            {curriculum.status === 'draft' ? (
              <section className="flex flex-wrap gap-2">
                <Button onClick={() => setEditOpen(true)} variant="outline">
                  <PenLine className="h-4 w-4 mr-2" />
                  {t('detail.actions.edit')}
                </Button>
                <Button onClick={handleSubmit} disabled={submitting}>
                  {submitting ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <Send className="h-4 w-4 mr-2" />
                  )}
                  {submitting ? t('detail.actions.submitting') : t('detail.actions.submit')}
                </Button>
              </section>
            ) : (
              <section
                className={cn(
                  'rounded-xl border p-4 text-sm',
                  STATUS_PILL[curriculum.status].bg,
                  STATUS_PILL[curriculum.status].text
                )}
              >
                {t(`detail.statusHint.${statusKey(curriculum.status)}`)}
              </section>
            )}

            <EditCurriculumDialog
              curriculum={curriculum}
              open={editOpen}
              onClose={() => setEditOpen(false)}
              onSaved={() => mutate()}
            />
          </>
        )}
      </div>
    </AppLayout>
  )
}

function StatusPill({
  status,
  t,
}: {
  status: CurriculumStatus
  t: ReturnType<typeof useTranslations>
}) {
  const styles = STATUS_PILL[status]
  const Icon = styles.Icon
  return (
    <div
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium',
        styles.bg,
        styles.text
      )}
    >
      <Icon className="h-3.5 w-3.5" />
      {t(`card.status.${statusKey(status)}`)}
    </div>
  )
}
