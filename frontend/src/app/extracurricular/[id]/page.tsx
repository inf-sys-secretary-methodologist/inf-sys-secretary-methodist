'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { ArrowLeft, MapPin, Calendar as CalendarIcon, Users, Loader2 } from 'lucide-react'
import { toast } from 'sonner'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import {
  useExtracurricularEvent,
  registerForExtracurricularEvent,
  unregisterFromExtracurricularEvent,
  pickExtracurricularErrorKey,
} from '@/hooks/useExtracurricularEvents'
import type { EventCategory, EventStatus, EventTargetAudience } from '@/types/extracurricular'
import { useAuthCheck } from '@/hooks/useAuth'
import { useAuthStore } from '@/stores/authStore'

const STATUS_COLORS: Record<EventStatus, string> = {
  draft: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300',
  published: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
  canceled: 'bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300',
  completed: 'bg-gray-100 text-gray-500 dark:bg-gray-900/40 dark:text-gray-400',
}

const CATEGORY_COLORS: Record<EventCategory, string> = {
  academic: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  cultural: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300',
  sports: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
  volunteer: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  professional: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/40 dark:text-cyan-300',
}

const AUDIENCE_COLORS: Record<EventTargetAudience, string> = {
  all: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300',
  students: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  teachers: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  staff: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/40 dark:text-cyan-300',
}

function canSeeParticipants(role: string | undefined): boolean {
  return role !== undefined && role !== 'student'
}

export default function ExtracurricularEventDetailPage() {
  const t = useTranslations('extracurricular')
  const params = useParams()
  const router = useRouter()
  useAuthCheck()
  const user = useAuthStore((s) => s.user)

  const idRaw = (params?.id as string | undefined) ?? null
  const id = idRaw ? Number(idRaw) : null
  const { event, isLoading, error, mutate } = useExtracurricularEvent(id)
  const [busy, setBusy] = useState(false)

  if (isLoading) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center py-24">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </AppLayout>
    )
  }

  if (error || !event) {
    return (
      <AppLayout>
        <div className="max-w-3xl mx-auto p-4 md:p-6">
          <div className="rounded-xl bg-card border border-border p-8 text-center">
            <p className="text-destructive font-medium">{t('loadFailed')}</p>
            <Button
              variant="outline"
              className="mt-4"
              onClick={() => router.push('/extracurricular')}
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              {t('backToList')}
            </Button>
          </div>
        </div>
      </AppLayout>
    )
  }

  const isRegistered = Boolean(user && event.participants?.some((p) => p.user_id === user.id))
  const eventID = event.id

  const handleRegister = async () => {
    if (busy) return
    setBusy(true)
    try {
      await registerForExtracurricularEvent(eventID)
      await mutate()
    } catch (err) {
      toast.error(t(`errors.${pickExtracurricularErrorKey(err)}`))
    } finally {
      setBusy(false)
    }
  }

  const handleUnregister = async () => {
    if (busy) return
    setBusy(true)
    try {
      await unregisterFromExtracurricularEvent(eventID)
      await mutate()
    } catch (err) {
      toast.error(t(`errors.${pickExtracurricularErrorKey(err)}`))
    } finally {
      setBusy(false)
    }
  }

  const showParticipants = canSeeParticipants(user?.role)

  return (
    <AppLayout>
      <div className="max-w-3xl mx-auto p-4 md:p-6 space-y-6">
        <Button variant="ghost" size="sm" onClick={() => router.push('/extracurricular')}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          {t('backToList')}
        </Button>

        <div className="flex flex-wrap items-center gap-2">
          <span
            className={cn(
              'text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full',
              STATUS_COLORS[event.status]
            )}
          >
            {t(`status.${event.status}`)}
          </span>
          <span
            className={cn(
              'text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full',
              CATEGORY_COLORS[event.category]
            )}
          >
            {t(`category.${event.category}`)}
          </span>
          <span
            className={cn(
              'text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full',
              AUDIENCE_COLORS[event.target_audience]
            )}
          >
            {t(`audience.${event.target_audience}`)}
          </span>
        </div>

        <h1 className="text-2xl md:text-3xl font-bold">{event.title}</h1>

        {event.description && (
          <p className="text-base text-muted-foreground whitespace-pre-line">{event.description}</p>
        )}

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
          {event.location && (
            <div className="flex items-center gap-2">
              <MapPin className="h-4 w-4 text-muted-foreground" />
              <span>{event.location}</span>
            </div>
          )}
          <div className="flex items-center gap-2">
            <CalendarIcon className="h-4 w-4 text-muted-foreground" />
            <span>
              {t('startAt')}: {new Date(event.start_at).toLocaleString()}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <CalendarIcon className="h-4 w-4 text-muted-foreground" />
            <span>
              {t('endAt')}: {new Date(event.end_at).toLocaleString()}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <Users className="h-4 w-4 text-muted-foreground" />
            <span>
              {t('participants')}:{' '}
              {event.max_capacity != null
                ? `${event.participant_count} / ${event.max_capacity}`
                : event.participant_count}
            </span>
          </div>
        </div>

        {event.status === 'published' && (
          <div className="flex justify-end gap-2">
            {isRegistered ? (
              <Button variant="outline" disabled={busy} onClick={handleUnregister}>
                {t('unregister')}
              </Button>
            ) : (
              <Button disabled={busy} onClick={handleRegister}>
                {t('register')}
              </Button>
            )}
          </div>
        )}

        {showParticipants && event.participants && event.participants.length > 0 && (
          <div className="space-y-2">
            <h2 className="text-lg font-semibold">{t('participants')}</h2>
            <ul className="space-y-1">
              {event.participants.map((p) => (
                <li key={p.user_id} className="text-sm flex items-center gap-2">
                  <Users className="h-3.5 w-3.5 text-muted-foreground" />
                  <span>user_id {p.user_id}</span>
                  <span className="text-xs text-muted-foreground">
                    {new Date(p.registered_at).toLocaleDateString()}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </AppLayout>
  )
}
