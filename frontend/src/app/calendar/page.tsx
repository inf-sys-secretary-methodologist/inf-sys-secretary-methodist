'use client'

import * as React from 'react'
import { startOfMonth, endOfMonth, startOfWeek, endOfWeek, addMonths, subMonths } from 'date-fns'
import { toast } from 'sonner'

import { useAuthCheck } from '@/hooks/useAuth'
import {
  useEventsByDateRange,
  createEvent,
  updateEvent,
  deleteEvent,
} from '@/hooks/useCalendarEvents'
import { FullCalendar } from '@/components/calendar'
import { AppLayout } from '@/components/layout'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import type { CreateEventInput } from '@/types/calendar'
import { canEdit } from '@/lib/auth/permissions'

export default function CalendarPage() {
  const { user } = useAuthCheck()
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
      toast.success('Событие создано')
    } catch {
      toast.error('Ошибка при создании события')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleUpdateEvent = async (id: number, data: CreateEventInput) => {
    setIsSubmitting(true)
    try {
      await updateEvent(id, data)
      await mutate()
      toast.success('Событие обновлено')
    } catch {
      toast.error('Ошибка при обновлении события')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDeleteEvent = async (id: number) => {
    setIsSubmitting(true)
    try {
      await deleteEvent(id)
      await mutate()
      toast.success('Событие удалено')
    } catch {
      toast.error('Ошибка при удалении события')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <AppLayout>
      <div className="mx-auto max-w-[1600px] h-[calc(100vh-8rem)] sm:h-[calc(100vh-10rem)]">
        <div className="relative h-full rounded-xl sm:rounded-2xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-black/95 overflow-hidden">
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
