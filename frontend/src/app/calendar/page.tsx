'use client'

import * as React from 'react'
import dynamic from 'next/dynamic'
import { startOfMonth, endOfMonth, startOfWeek, endOfWeek, addMonths, subMonths } from 'date-fns'
import { toast } from 'sonner'
import { useTranslations } from 'next-intl'
import { Loader2 } from 'lucide-react'

import { useAuthCheck } from '@/hooks/useAuth'
import {
  useEventsByDateRange,
  createEvent,
  updateEvent,
  deleteEvent,
} from '@/hooks/useCalendarEvents'
import { AppLayout } from '@/components/layout'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import type { CreateEventInput } from '@/types/calendar'
import { canEdit } from '@/lib/auth/permissions'

// Динамический импорт FullCalendar - тяжелый компонент
const FullCalendar = dynamic(
  () => import('@/components/calendar').then((mod) => ({ default: mod.FullCalendar })),
  {
    loading: () => (
      <div className="flex items-center justify-center h-96">
        <Loader2 className="h-8 w-8 animate-spin text-gray-500" />
      </div>
    ),
    ssr: false,
  }
)

export default function CalendarPage() {
  const { user } = useAuthCheck()
  const t = useTranslations('calendar')
  const userCanEdit = canEdit(user?.role)
  const [currentMonth] = React.useState(new Date())
  const [isSubmitting, setIsSubmitting] = React.useState(false)

  // Calculate date range for fetching events (current month + buffer)
  const rangeStart = startOfWeek(startOfMonth(subMonths(currentMonth, 1)))
  const rangeEnd = endOfWeek(endOfMonth(addMonths(currentMonth, 1)))

  const { events, isLoading: eventsLoading, mutate } = useEventsByDateRange(rangeStart, rangeEnd)

  const handleCreateEvent = async (data: CreateEventInput) => {
    setIsSubmitting(true)
    try {
      await createEvent(data)
      await mutate()
      toast.success(t('eventCreated'))
    } catch {
      toast.error(t('createError'))
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleUpdateEvent = async (id: number, data: CreateEventInput) => {
    setIsSubmitting(true)
    try {
      await updateEvent(id, data)
      await mutate()
      toast.success(t('eventUpdated'))
    } catch {
      toast.error(t('updateError'))
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDeleteEvent = async (id: number) => {
    setIsSubmitting(true)
    try {
      await deleteEvent(id)
      await mutate()
      toast.success(t('eventDeleted'))
    } catch {
      toast.error(t('deleteError'))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <AppLayout>
      <div className="mx-auto max-w-[1600px] space-y-6 sm:space-y-8">
        {/* Page Header */}
        <div className="text-center space-y-2 sm:space-y-4">
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-gray-900 dark:text-white">
            {t('title')}
          </h1>
          <p className="text-base sm:text-lg text-gray-600 dark:text-gray-300">{t('subtitle')}</p>
        </div>

        <div className="relative h-[calc(100vh-16rem)] sm:h-[calc(100vh-18rem)] rounded-xl sm:rounded-2xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-black/95 overflow-hidden">
          <GlowingEffect
            spread={40}
            glow={true}
            disabled={false}
            proximity={64}
            inactiveZone={0.01}
            borderWidth={3}
          />
          <div className="relative z-10 h-full">
            <FullCalendar
              events={events}
              isLoading={eventsLoading || isSubmitting}
              onCreateEvent={userCanEdit ? handleCreateEvent : undefined}
              onUpdateEvent={userCanEdit ? handleUpdateEvent : undefined}
              onDeleteEvent={userCanEdit ? handleDeleteEvent : undefined}
            />
          </div>
        </div>
      </div>
    </AppLayout>
  )
}
