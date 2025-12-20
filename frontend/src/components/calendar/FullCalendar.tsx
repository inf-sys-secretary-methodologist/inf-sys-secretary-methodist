'use client'

import * as React from 'react'
import { isSameDay } from 'date-fns'
import { useTranslations, useLocale } from 'next-intl'

import { cn } from '@/lib/utils'
import { CalendarHeader } from './CalendarHeader'
import { MonthView } from './MonthView'
import { WeekView } from './WeekView'
import { DayView } from './DayView'
import { EventModal } from './EventModal'
import { EventCard } from './EventCard'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useIsMobile } from '@/hooks/use-media-query'
import type { CalendarEvent, CalendarView, CreateEventInput } from '@/types/calendar'

interface FullCalendarProps {
  events: CalendarEvent[]
  isLoading?: boolean
  onCreateEvent?: (data: CreateEventInput) => Promise<void>
  onUpdateEvent?: (id: number, data: CreateEventInput) => Promise<void>
  onDeleteEvent?: (id: number) => Promise<void>
  onEventClick?: (event: CalendarEvent) => void
  className?: string
}

export function FullCalendar({
  events,
  isLoading,
  onCreateEvent,
  onUpdateEvent,
  onDeleteEvent,
  onEventClick,
  className,
}: FullCalendarProps) {
  const t = useTranslations('calendarView')
  const locale = useLocale()
  const isMobile = useIsMobile()
  const [currentDate, setCurrentDate] = React.useState(new Date())
  const [selectedDate, setSelectedDate] = React.useState<Date | undefined>()
  const [view, setView] = React.useState<CalendarView>('month')

  // Sync selectedDate with currentDate when navigating in day/week view
  const handleDateChange = React.useCallback(
    (date: Date) => {
      setCurrentDate(date)
      // In day view, also update selectedDate so the view stays in sync
      if (view === 'day') {
        setSelectedDate(date)
      }
    },
    [view]
  )

  // Handle view change with date sync
  const handleViewChange = React.useCallback(
    (newView: CalendarView) => {
      // When switching to day view, sync currentDate with selectedDate if it exists
      if (newView === 'day' && selectedDate) {
        setCurrentDate(selectedDate)
      }
      setView(newView)
    },
    [selectedDate]
  )
  const [modalOpen, setModalOpen] = React.useState(false)
  const [selectedEvent, setSelectedEvent] = React.useState<CalendarEvent | null>(null)
  const [initialDate, setInitialDate] = React.useState<Date | undefined>()
  const [dayEventsOpen, setDayEventsOpen] = React.useState(false)

  // Force day view on mobile
  React.useEffect(() => {
    if (isMobile && view === 'month') {
      // Keep month view but with compact layout
    }
  }, [isMobile, view])

  const handleAddEvent = () => {
    setSelectedEvent(null)
    setInitialDate(selectedDate || new Date())
    setModalOpen(true)
  }

  const handleEventClick = (event: CalendarEvent) => {
    if (onEventClick) {
      onEventClick(event)
    } else {
      setSelectedEvent(event)
      setModalOpen(true)
    }
  }

  const handleDateSelect = (date: Date) => {
    setSelectedDate(date)

    // On mobile, show day events modal
    if (isMobile) {
      const dayEvents = events.filter((e) => isSameDay(new Date(e.start_time), date))
      if (dayEvents.length > 0) {
        setDayEventsOpen(true)
      }
    }
  }

  const handleTimeSlotClick = onCreateEvent
    ? (date: Date) => {
        setSelectedEvent(null)
        setInitialDate(date)
        setModalOpen(true)
      }
    : undefined

  const handleSubmit = async (data: CreateEventInput) => {
    if (selectedEvent && onUpdateEvent) {
      await onUpdateEvent(selectedEvent.id, data)
    } else if (onCreateEvent) {
      await onCreateEvent(data)
    }
  }

  const handleDelete = async (id: number) => {
    if (onDeleteEvent) {
      await onDeleteEvent(id)
    }
  }

  const selectedDayEvents = selectedDate
    ? events.filter((e) => isSameDay(new Date(e.start_time), selectedDate))
    : []

  return (
    <div className={cn('flex h-full flex-col', className)}>
      {/* Header */}
      <CalendarHeader
        currentDate={currentDate}
        view={view}
        onDateChange={handleDateChange}
        onViewChange={handleViewChange}
        onAddEvent={onCreateEvent ? handleAddEvent : undefined}
      />

      {/* Mobile View Tabs */}
      {isMobile && (
        <div className="px-4 pb-2">
          <Tabs
            value={view}
            onValueChange={(v) => handleViewChange(v as CalendarView)}
            className="w-full"
          >
            <TabsList className="w-full">
              <TabsTrigger value="month" className="flex-1">
                {t('month')}
              </TabsTrigger>
              <TabsTrigger value="week" className="flex-1">
                {t('week')}
              </TabsTrigger>
              <TabsTrigger value="day" className="flex-1">
                {t('day')}
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>
      )}

      {/* Calendar Views */}
      <div className="flex-1 overflow-hidden border-t border-gray-200 dark:border-gray-700">
        {view === 'month' && (
          <MonthView
            currentDate={currentDate}
            selectedDate={selectedDate}
            events={events}
            onDateSelect={handleDateSelect}
            onEventClick={handleEventClick}
          />
        )}

        {view === 'week' && (
          <WeekView
            currentDate={currentDate}
            selectedDate={selectedDate}
            events={events}
            onDateSelect={handleDateSelect}
            onEventClick={handleEventClick}
            onTimeSlotClick={handleTimeSlotClick}
          />
        )}

        {view === 'day' && (
          <DayView
            currentDate={selectedDate || currentDate}
            events={events.filter((e) =>
              isSameDay(new Date(e.start_time), selectedDate || currentDate)
            )}
            onEventClick={handleEventClick}
            onTimeSlotClick={handleTimeSlotClick}
          />
        )}
      </div>

      {/* Event Modal */}
      <EventModal
        open={modalOpen}
        onOpenChange={setModalOpen}
        event={selectedEvent}
        initialDate={initialDate}
        onSubmit={onCreateEvent || onUpdateEvent ? handleSubmit : undefined}
        onDelete={selectedEvent && onDeleteEvent ? handleDelete : undefined}
        isLoading={isLoading}
      />

      {/* Day Events Modal (Mobile) */}
      <Dialog open={dayEventsOpen} onOpenChange={setDayEventsOpen}>
        <DialogContent className="max-h-[80vh] overflow-auto">
          <DialogHeader>
            <DialogTitle>
              {t('eventsOn')}{' '}
              {selectedDate
                ? new Intl.DateTimeFormat(locale, {
                    day: 'numeric',
                    month: 'long',
                  }).format(selectedDate)
                : ''}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            {selectedDayEvents.length === 0 ? (
              <p className="text-center text-muted-foreground py-4">{t('noEvents')}</p>
            ) : (
              selectedDayEvents.map((event) => (
                <EventCard
                  key={event.id}
                  event={event}
                  variant="full"
                  onClick={() => {
                    setDayEventsOpen(false)
                    handleEventClick(event)
                  }}
                />
              ))
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
