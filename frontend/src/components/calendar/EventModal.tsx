'use client'

import * as React from 'react'
import dynamic from 'next/dynamic'
import { format } from 'date-fns'
import { ru } from 'date-fns/locale'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod/v4'
import { CalendarIcon, Clock, MapPin, Loader2 } from 'lucide-react'

import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'

// Lazy load Calendar to reduce initial bundle (react-day-picker ~100KB)
const Calendar = dynamic(() => import('@/components/ui/calendar').then((mod) => mod.Calendar), {
  loading: () => (
    <div className="flex items-center justify-center p-4">
      <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
    </div>
  ),
  ssr: false,
})
import type { CalendarEvent, CreateEventInput, EventType } from '@/types/calendar'
import { EVENT_TYPE_LABELS } from '@/types/calendar'

const eventSchema = z.object({
  title: z.string().min(1, 'Название обязательно').max(500),
  description: z.string().max(5000).optional(),
  event_type: z.enum(['meeting', 'deadline', 'task', 'reminder', 'holiday', 'personal']),
  start_date: z.date(),
  start_time: z.string().regex(/^\d{2}:\d{2}$/, 'Формат HH:MM'),
  end_date: z.date().optional(),
  end_time: z
    .string()
    .refine((val) => val === '' || /^\d{2}:\d{2}$/.test(val), 'Формат HH:MM')
    .optional(),
  all_day: z.boolean(),
  location: z.string().max(500).optional(),
  color: z.string().optional(),
})

type EventFormData = z.infer<typeof eventSchema>

interface EventModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  event?: CalendarEvent | null
  initialDate?: Date
  onSubmit?: (data: CreateEventInput) => Promise<void>
  onDelete?: (id: number) => Promise<void>
  isLoading?: boolean
}

const EVENT_TYPES: EventType[] = ['meeting', 'deadline', 'task', 'reminder', 'holiday', 'personal']

const COLOR_OPTIONS = [
  { value: 'default', label: 'По умолчанию' },
  { value: '#3b82f6', label: 'Синий' },
  { value: '#ef4444', label: 'Красный' },
  { value: '#22c55e', label: 'Зелёный' },
  { value: '#eab308', label: 'Жёлтый' },
  { value: '#a855f7', label: 'Фиолетовый' },
  { value: '#6b7280', label: 'Серый' },
]

export function EventModal({
  open,
  onOpenChange,
  event,
  initialDate,
  onSubmit,
  onDelete,
  isLoading,
}: EventModalProps) {
  const isEditing = !!event

  const defaultValues: Partial<EventFormData> = React.useMemo(() => {
    if (event) {
      const startDate = new Date(event.start_time)
      const endDate = event.end_time ? new Date(event.end_time) : undefined
      return {
        title: event.title,
        description: event.description || '',
        event_type: event.event_type,
        start_date: startDate,
        start_time: format(startDate, 'HH:mm'),
        end_date: endDate,
        end_time: endDate ? format(endDate, 'HH:mm') : '',
        all_day: event.all_day,
        location: event.location || '',
        color: event.color || '',
      }
    }
    const date = initialDate || new Date()
    return {
      title: '',
      description: '',
      event_type: 'meeting' as EventType,
      start_date: date,
      start_time: format(date, 'HH:mm'),
      all_day: false,
      location: '',
      color: '',
    }
  }, [event, initialDate])

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    reset,
    formState: { errors },
  } = useForm<EventFormData>({
    resolver: zodResolver(eventSchema),
    defaultValues,
  })

  React.useEffect(() => {
    reset(defaultValues)
  }, [defaultValues, reset])

  const allDay = watch('all_day')
  const startDate = watch('start_date')
  const endDate = watch('end_date')

  const handleFormSubmit = async (data: EventFormData) => {
    if (!onSubmit) return

    const startDateTime = new Date(data.start_date)
    if (!data.all_day) {
      const [hours, minutes] = data.start_time.split(':').map(Number)
      startDateTime.setHours(hours, minutes, 0, 0)
    }

    let endDateTime: Date | undefined
    if (data.end_date) {
      endDateTime = new Date(data.end_date)
      if (!data.all_day && data.end_time) {
        const [hours, minutes] = data.end_time.split(':').map(Number)
        endDateTime.setHours(hours, minutes, 0, 0)
      } else {
        // For all-day events, set end of day
        endDateTime.setHours(23, 59, 59, 999)
      }
    }

    // Создаём объект только с заполненными полями
    const input: CreateEventInput = {
      title: data.title,
      event_type: data.event_type,
      start_time: startDateTime.toISOString(),
      all_day: data.all_day,
      is_recurring: false,
    }

    // Добавляем опциональные поля только если они заполнены
    if (data.description) input.description = data.description
    if (endDateTime) input.end_time = endDateTime.toISOString()
    if (data.location) input.location = data.location
    if (data.color) input.color = data.color

    await onSubmit(input)
    onOpenChange(false)
  }

  // View-only mode when onSubmit is not provided
  const isViewOnly = !onSubmit

  const handleDelete = async () => {
    if (event && onDelete) {
      await onDelete(event.id)
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>
            {isViewOnly
              ? 'Просмотр события'
              : isEditing
                ? 'Редактировать событие'
                : 'Новое событие'}
          </DialogTitle>
          <DialogDescription>
            {isViewOnly
              ? 'Детали события'
              : isEditing
                ? 'Измените детали события'
                : 'Заполните информацию о новом событии'}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-4">
          {/* Title */}
          <div className="space-y-2">
            <Label htmlFor="title">Название {!isViewOnly && '*'}</Label>
            <Input
              id="title"
              placeholder="Введите название события"
              disabled={isViewOnly}
              {...register('title')}
            />
            {errors.title && <p className="text-sm text-destructive">{errors.title.message}</p>}
          </div>

          {/* Event Type */}
          <div className="space-y-2">
            <Label>Тип события</Label>
            <Select
              value={watch('event_type')}
              onValueChange={(v) => setValue('event_type', v as EventType)}
              disabled={isViewOnly}
            >
              <SelectTrigger>
                <SelectValue placeholder="Выберите тип" />
              </SelectTrigger>
              <SelectContent>
                {EVENT_TYPES.map((type) => (
                  <SelectItem key={type} value={type}>
                    {EVENT_TYPE_LABELS[type]}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* All Day Toggle */}
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="all_day"
              disabled={isViewOnly}
              {...register('all_day')}
              className="h-4 w-4 rounded border-gray-300"
            />
            <Label htmlFor="all_day">Весь день</Label>
          </div>

          {/* Start Date & Time */}
          <div className="grid gap-4 sm:grid-cols-2">
            {/* Start Date */}
            <div className="space-y-2">
              <Label>Дата начала {!isViewOnly && '*'}</Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    disabled={isViewOnly}
                    className={cn(
                      'w-full justify-start text-left font-normal',
                      !startDate && 'text-muted-foreground'
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {startDate ? format(startDate, 'd MMM yyyy', { locale: ru }) : 'Выберите дату'}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={startDate}
                    onSelect={(date) => date && setValue('start_date', date)}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            </div>

            {/* Start Time */}
            {!allDay && (
              <div className="space-y-2">
                <Label htmlFor="start_time">Время начала {!isViewOnly && '*'}</Label>
                <div className="relative">
                  <Clock className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    id="start_time"
                    type="time"
                    className="pl-9"
                    disabled={isViewOnly}
                    {...register('start_time')}
                  />
                </div>
              </div>
            )}
          </div>

          {/* End Date & Time */}
          <div className="grid gap-4 sm:grid-cols-2">
            {/* End Date */}
            <div className="space-y-2">
              <Label>Дата окончания</Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    disabled={isViewOnly}
                    className={cn(
                      'w-full justify-start text-left font-normal',
                      !endDate && 'text-muted-foreground'
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {endDate ? format(endDate, 'd MMM yyyy', { locale: ru }) : 'Выберите дату'}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={endDate}
                    onSelect={(date) => setValue('end_date', date)}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            </div>

            {/* End Time */}
            {!allDay && (
              <div className="space-y-2">
                <Label htmlFor="end_time">Время окончания</Label>
                <div className="relative">
                  <Clock className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    id="end_time"
                    type="time"
                    className="pl-9"
                    disabled={isViewOnly}
                    {...register('end_time')}
                  />
                </div>
              </div>
            )}
          </div>

          {/* Location */}
          <div className="space-y-2">
            <Label htmlFor="location">Место</Label>
            <div className="relative">
              <MapPin className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                id="location"
                placeholder="Введите место проведения"
                className="pl-9"
                disabled={isViewOnly}
                {...register('location')}
              />
            </div>
          </div>

          {/* Color */}
          <div className="space-y-2">
            <Label>Цвет</Label>
            <Select
              value={watch('color') || 'default'}
              onValueChange={(v) => setValue('color', v === 'default' ? '' : v)}
              disabled={isViewOnly}
            >
              <SelectTrigger>
                <SelectValue placeholder="По умолчанию" />
              </SelectTrigger>
              <SelectContent>
                {COLOR_OPTIONS.map((color) => (
                  <SelectItem key={color.value} value={color.value}>
                    <div className="flex items-center gap-2">
                      {color.value !== 'default' && (
                        <div
                          className="h-4 w-4 rounded-full"
                          style={{ backgroundColor: color.value }}
                        />
                      )}
                      {color.label}
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="description">Описание</Label>
            <textarea
              id="description"
              placeholder="Добавьте описание события"
              disabled={isViewOnly}
              className="min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              {...register('description')}
            />
          </div>

          <DialogFooter className="gap-2 sm:gap-0">
            {!isViewOnly && isEditing && onDelete && (
              <Button
                type="button"
                variant="destructive"
                onClick={handleDelete}
                disabled={isLoading}
              >
                Удалить
              </Button>
            )}
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {isViewOnly ? 'Закрыть' : 'Отмена'}
            </Button>
            {!isViewOnly && (
              <Button type="submit" disabled={isLoading}>
                {isLoading ? 'Сохранение...' : isEditing ? 'Сохранить' : 'Создать'}
              </Button>
            )}
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
