import type { Meta, StoryObj } from '@storybook/nextjs'
import { StatsCard } from './StatsCard'
import { FileText, Users, Calendar, ClipboardList, TrendingUp, Activity } from 'lucide-react'

const meta: Meta<typeof StatsCard> = {
  title: 'Dashboard/StatsCard',
  component: StatsCard,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `
Карточка статистики для дашборда с анимированными числами и эффектом свечения.

**Используемые компоненты:**
- \`GlowingEffect\` - эффект свечения границы при наведении
- \`NumberTicker\` - анимированное отображение чисел

**Цветовая схема изменения:**
- Положительное: \`bg-green-500/20 text-green-600 border-green-500/50\`
- Отрицательное: \`bg-red-500/20 text-red-600 border-red-500/50\`

**Стиль карточки:**
- \`bg-white dark:bg-black/95\`
- \`border border-gray-200 dark:border-gray-700\`
- \`rounded-2xl\`
- \`hover:scale-105\` при наведении
        `,
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    value: {
      control: { type: 'number', min: 0, max: 100000 },
      description: 'Числовое значение',
    },
    change: {
      control: { type: 'number', min: -100, max: 100, step: 0.1 },
      description: 'Изменение в процентах',
    },
    title: {
      control: 'text',
      description: 'Заголовок карточки',
    },
    period: {
      control: 'text',
      description: 'Период (например: "месяц", "неделю")',
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    icon: FileText,
    title: 'Документы',
    value: 1234,
    change: 12.5,
    period: 'месяц',
  },
}

export const NegativeChange: Story = {
  args: {
    icon: Users,
    title: 'Активные студенты',
    value: 856,
    change: -5.3,
    period: 'неделю',
  },
}

export const LargeNumber: Story = {
  args: {
    icon: Activity,
    title: 'Всего записей',
    value: 125847,
    change: 8.2,
    period: 'год',
  },
}

export const ZeroChange: Story = {
  args: {
    icon: Calendar,
    title: 'Мероприятия',
    value: 42,
    change: 0,
    period: 'месяц',
  },
}

export const StatsGrid: Story = {
  render: () => (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-2xl">
      <StatsCard icon={FileText} title="Документы" value={1234} change={12.5} period="месяц" />
      <StatsCard icon={Users} title="Студенты" value={856} change={-2.3} period="месяц" />
      <StatsCard icon={Calendar} title="Мероприятия" value={42} change={25.0} period="месяц" />
      <StatsCard icon={ClipboardList} title="Отчеты" value={18} change={5.8} period="месяц" />
    </div>
  ),
}

export const DashboardLayout: Story = {
  render: () => (
    <div className="p-6 bg-gray-100 dark:bg-gray-900 rounded-xl">
      <h2 className="text-2xl font-bold mb-6 text-gray-900 dark:text-white">Обзор статистики</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard icon={FileText} title="Документы" value={1234} change={12.5} period="месяц" />
        <StatsCard icon={Users} title="Студенты" value={856} change={3.2} period="месяц" />
        <StatsCard icon={Calendar} title="Мероприятия" value={42} change={-1.5} period="месяц" />
        <StatsCard icon={TrendingUp} title="Отчеты" value={18} change={8.7} period="месяц" />
      </div>
    </div>
  ),
}
