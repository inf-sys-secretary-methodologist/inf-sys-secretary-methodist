import type { Meta, StoryObj } from '@storybook/nextjs'
import { Calendar } from './calendar'
import { useState } from 'react'
import type { DateRange } from 'react-day-picker'

const meta = {
  title: 'UI/Calendar',
  component: Calendar,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `
Компонент календаря на базе react-day-picker.

**Особенности:**
- Выбор одной даты или диапазона
- Поддержка локализации
- Кастомизация стилей через classNames
        `,
      },
    },
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Calendar>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: function DefaultCalendar() {
    const [date, setDate] = useState<Date | undefined>(new Date())
    return (
      <Calendar mode="single" selected={date} onSelect={setDate} className="rounded-md border" />
    )
  },
}

export const WithRange: Story = {
  render: function RangeCalendar() {
    const [range, setRange] = useState<DateRange | undefined>({
      from: new Date(2024, 0, 20),
      to: new Date(2024, 0, 25),
    })
    return (
      <Calendar
        mode="range"
        selected={range}
        onSelect={setRange}
        className="rounded-md border"
        numberOfMonths={2}
      />
    )
  },
}

export const WithDisabledDates: Story = {
  render: function DisabledCalendar() {
    const [date, setDate] = useState<Date | undefined>(new Date())
    return (
      <Calendar
        mode="single"
        selected={date}
        onSelect={setDate}
        disabled={(date) => date < new Date()}
        className="rounded-md border"
      />
    )
  },
}

export const MultipleMonths: Story = {
  render: function MultiMonthCalendar() {
    const [date, setDate] = useState<Date | undefined>(new Date())
    return (
      <Calendar
        mode="single"
        selected={date}
        onSelect={setDate}
        numberOfMonths={2}
        className="rounded-md border"
      />
    )
  },
}
