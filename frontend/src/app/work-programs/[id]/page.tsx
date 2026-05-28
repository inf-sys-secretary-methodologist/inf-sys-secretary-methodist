'use client'

import { useMemo, type ReactNode } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { ArrowLeft, BookMarked, Calendar, GraduationCap, Loader2 } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { useWorkProgram } from '@/hooks/useWorkPrograms'
import { useAuthCheck } from '@/hooks/useAuth'
import { STATUS_STYLES, statusKey, revisionStatusKey } from '@/components/work-program/status'
import type { WorkProgram, WorkProgramStatus } from '@/types/workProgram'
import { cn } from '@/lib/utils'

// WorkProgramDetailPage — full РПД view with all six inner collections.
// Read-only in this slice (transition actions land in a later PR).
// Visible to all authenticated roles (no student redirect — 273-ФЗ
// ст. 29); the backend scopes what each role can fetch.
export default function WorkProgramDetailPage() {
  const params = useParams<{ id: string }>()
  const id = useMemo(() => {
    const raw = params?.id
    const parsed = typeof raw === 'string' ? Number(raw) : NaN
    return Number.isInteger(parsed) && parsed > 0 ? parsed : null
  }, [params])

  const { isAuthenticated, isLoading: authLoading } = useAuthCheck()
  const t = useTranslations('workProgram')

  const enabled = !authLoading && isAuthenticated && id !== null
  const { workProgram: wp, isLoading: detailLoading, error } = useWorkProgram(id, { enabled })

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
          href="/work-programs"
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
        ) : error || !wp ? (
          <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center">
            <p className="font-medium text-destructive">{t('detail.loadFailed')}</p>
          </div>
        ) : (
          <WorkProgramDetail wp={wp} t={t} />
        )}
      </div>
    </AppLayout>
  )
}

type T = ReturnType<typeof useTranslations>

function WorkProgramDetail({ wp, t }: { wp: WorkProgram; t: T }) {
  return (
    <>
      <header className="space-y-3">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <h1 className="text-2xl font-bold">{wp.title}</h1>
          <StatusPill status={wp.status} t={t} />
        </div>
        <dl className="flex flex-wrap gap-x-6 gap-y-1.5 text-sm text-muted-foreground">
          <div className="inline-flex items-center gap-1.5">
            <BookMarked className="h-3.5 w-3.5" />
            <span>{t('card.discipline', { id: wp.discipline_id })}</span>
          </div>
          <div className="inline-flex items-center gap-1.5">
            <GraduationCap className="h-3.5 w-3.5" />
            <span>{wp.specialty_code}</span>
          </div>
          <div className="inline-flex items-center gap-1.5">
            <Calendar className="h-3.5 w-3.5" />
            <span>{wp.applicable_from_year}</span>
          </div>
        </dl>
      </header>

      <section
        className={cn(
          'rounded-xl border p-4 text-sm',
          STATUS_STYLES[wp.status].bg,
          STATUS_STYLES[wp.status].text
        )}
      >
        {t(`detail.statusHint.${statusKey(wp.status)}`)}
      </section>

      {wp.reject_reason ? (
        <section className="rounded-xl border border-orange-300/40 bg-orange-50 dark:bg-orange-950/20 p-4 text-sm">
          <p className="font-medium">{t('detail.fields.rejectReason')}</p>
          <p className="mt-1 whitespace-pre-wrap">{wp.reject_reason}</p>
        </section>
      ) : null}

      {wp.annotation ? (
        <Section title={t('detail.sections.annotation')} count={1} t={t}>
          <p className="text-sm whitespace-pre-wrap">{wp.annotation}</p>
        </Section>
      ) : null}

      <Section title={t('detail.sections.goals')} count={wp.goals.length} t={t}>
        <ol className="list-decimal space-y-1.5 pl-5 text-sm">
          {wp.goals.map((g) => (
            <li key={g.id}>{g.text}</li>
          ))}
        </ol>
      </Section>

      <Section title={t('detail.sections.competences')} count={wp.competences.length} t={t}>
        <ul className="space-y-2 text-sm">
          {wp.competences.map((c) => (
            <li key={c.id} className="flex flex-wrap items-baseline gap-x-2">
              <span className="font-mono font-medium">{c.code}</span>
              <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                {t(`detail.competenceType.${c.type}`)}
              </span>
              <span>{c.description}</span>
            </li>
          ))}
        </ul>
      </Section>

      <Section title={t('detail.sections.topics')} count={wp.topics.length} t={t}>
        <ul className="space-y-2 text-sm">
          {wp.topics.map((tp) => (
            <li key={tp.id} className="flex flex-wrap items-baseline gap-x-2">
              <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                {t(`detail.topicKind.${tp.kind}`)}
              </span>
              <span className="font-medium">{tp.title}</span>
              <span className="text-muted-foreground">
                {t('detail.topicHours', { hours: tp.hours })}
              </span>
              {typeof tp.week_number === 'number' ? (
                <span className="text-muted-foreground">
                  {t('detail.topicWeek', { week: tp.week_number })}
                </span>
              ) : null}
            </li>
          ))}
        </ul>
      </Section>

      <Section title={t('detail.sections.assessments')} count={wp.assessments.length} t={t}>
        <ul className="space-y-2 text-sm">
          {wp.assessments.map((a) => (
            <li key={a.id} className="flex flex-wrap items-baseline gap-x-2">
              <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                {t(`detail.assessmentType.${a.type}`)}
              </span>
              <span>{a.description}</span>
              <span className="text-muted-foreground">
                {t('detail.maxScore', { score: a.max_score })}
              </span>
            </li>
          ))}
        </ul>
      </Section>

      <Section title={t('detail.sections.references')} count={wp.references.length} t={t}>
        <ul className="space-y-2 text-sm">
          {wp.references.map((r) => (
            <li key={r.id} className="flex flex-wrap items-baseline gap-x-2">
              <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                {t(`detail.referenceType.${r.kind}`)}
              </span>
              <span>{r.citation}</span>
              {typeof r.year === 'number' ? (
                <span className="text-muted-foreground">{r.year}</span>
              ) : null}
            </li>
          ))}
        </ul>
      </Section>

      <Section title={t('detail.sections.revisions')} count={wp.revisions.length} t={t}>
        <ul className="space-y-2 text-sm">
          {wp.revisions.map((rev) => (
            <li key={rev.id} className="flex flex-wrap items-baseline gap-x-2">
              <span className="font-medium">#{rev.revision_number}</span>
              <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                {t(`detail.revisionChangeType.${rev.change_type}`)}
              </span>
              <span>{rev.change_summary}</span>
              <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                {t(`detail.revisionStatus.${revisionStatusKey(rev.status)}`)}
              </span>
            </li>
          ))}
        </ul>
      </Section>
    </>
  )
}

// Section renders a titled block, falling back to the shared empty
// label when the collection is empty so the document structure stays
// visible even for a sparsely filled draft.
function Section({
  title,
  count,
  t,
  children,
}: {
  title: string
  count: number
  t: T
  children: ReactNode
}) {
  return (
    <section className="space-y-2 rounded-xl border border-border bg-card p-4">
      <h2 className="text-base font-semibold">{title}</h2>
      {count === 0 ? (
        <p className="text-sm text-muted-foreground">{t('detail.sections.empty')}</p>
      ) : (
        children
      )}
    </section>
  )
}

function StatusPill({ status, t }: { status: WorkProgramStatus; t: T }) {
  const styles = STATUS_STYLES[status]
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
