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
import { UserMenu } from '@/components/UserMenu'
import { ThemeToggleButton } from '@/components/theme-toggle-button'
import { NavBar } from '@/components/ui/tubelight-navbar'
import { getAvailableNavItems } from '@/config/navigation'
import type { CreateEventInput } from '@/types/calendar'

export default function CalendarPage() {
  const { user, isLoading: authLoading } = useAuthCheck()
  const [currentMonth] = React.useState(new Date())
  const [isSubmitting, setIsSubmitting] = React.useState(false)

  // Calculate date range for fetching events (current month + buffer)
  const rangeStart = startOfWeek(startOfMonth(subMonths(currentMonth, 1)))
  const rangeEnd = endOfWeek(endOfMonth(addMonths(currentMonth, 1)))

  const { events, isLoading: eventsLoading, mutate } = useEventsByDateRange(rangeStart, rangeEnd)

  const navItems = getAvailableNavItems(user?.role)

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

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="text-muted-foreground">Загрузка...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col bg-background">
      {/* Navigation Bar */}
      <NavBar items={navItems} />

      {/* Top Navigation */}
      <div
        className="fixed top-8 right-8 z-50 pointer-events-auto flex items-center gap-3"
        style={{ isolation: 'isolate' }}
      >
        <UserMenu />
        <ThemeToggleButton />
      </div>

      {/* Main Content */}
      <main className="flex-1 pt-24 pb-8">
        <div className="mx-auto max-w-[1600px] px-4 sm:px-6 lg:px-8 h-[calc(100vh-8rem)]">
          <div className="relative h-full rounded-2xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-black/95 overflow-hidden">
            <div className="relative z-10 h-full">
              <FullCalendar
                events={events}
                isLoading={eventsLoading || isSubmitting}
                onCreateEvent={handleCreateEvent}
                onUpdateEvent={handleUpdateEvent}
                onDeleteEvent={handleDeleteEvent}
              />
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
