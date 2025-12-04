import type { Meta, StoryObj } from '@storybook/nextjs'
import { TrendChart } from './TrendChart'

const meta: Meta<typeof TrendChart> = {
  title: 'Dashboard/TrendChart',
  component: TrendChart,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `
График трендов для дашборда с использованием Recharts.

**Используемые компоненты:**
- \`GlowingEffect\` - эффект свечения границы при наведении
- \`recharts\` - AreaChart с градиентной заливкой

**Особенности:**
- Поддержка нескольких datasets с разными цветами
- Градиентная заливка под графиком
- Адаптивный размер (ResponsiveContainer)
- Русский формат дат (DD.MM)

**Стиль карточки:**
- \`bg-white dark:bg-black/95\`
- \`border border-gray-200 dark:border-gray-700\`
- \`rounded-2xl\`
        `,
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    title: {
      control: 'text',
      description: 'Заголовок графика',
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

// Generate sample data
const generateTrendData = (days: number, baseValue: number, variance: number) => {
  const data = []
  const today = new Date()
  for (let i = days - 1; i >= 0; i--) {
    const date = new Date(today)
    date.setDate(date.getDate() - i)
    data.push({
      date: date.toISOString(),
      value: Math.floor(baseValue + Math.random() * variance - variance / 2),
    })
  }
  return data
}

export const Default: Story = {
  args: {
    title: 'Динамика документов',
    datasets: [
      {
        name: 'Документы',
        data: generateTrendData(14, 50, 30),
        color: '#3B82F6',
      },
    ],
  },
  decorators: [
    (Story) => (
      <div className="w-[600px]">
        <Story />
      </div>
    ),
  ],
}

export const MultipleDatasets: Story = {
  args: {
    title: 'Сравнение активности',
    datasets: [
      {
        name: 'Документы',
        data: generateTrendData(14, 50, 20),
        color: '#3B82F6',
      },
      {
        name: 'Отчеты',
        data: generateTrendData(14, 30, 15),
        color: '#10B981',
      },
    ],
  },
  decorators: [
    (Story) => (
      <div className="w-[600px]">
        <Story />
      </div>
    ),
  ],
}

export const ThreeDatasets: Story = {
  args: {
    title: 'Обзор активности',
    datasets: [
      {
        name: 'Документы',
        data: generateTrendData(14, 60, 25),
        color: '#3B82F6',
      },
      {
        name: 'Мероприятия',
        data: generateTrendData(14, 40, 20),
        color: '#8B5CF6',
      },
      {
        name: 'Задачи',
        data: generateTrendData(14, 80, 30),
        color: '#F59E0B',
      },
    ],
  },
  decorators: [
    (Story) => (
      <div className="w-[700px]">
        <Story />
      </div>
    ),
  ],
}

export const WeeklyData: Story = {
  args: {
    title: 'Еженедельная статистика',
    datasets: [
      {
        name: 'Активность',
        data: generateTrendData(7, 100, 50),
        color: '#EC4899',
      },
    ],
  },
  decorators: [
    (Story) => (
      <div className="w-[500px]">
        <Story />
      </div>
    ),
  ],
}

export const MonthlyData: Story = {
  args: {
    title: 'Ежемесячный отчет',
    datasets: [
      {
        name: 'Студенты',
        data: generateTrendData(30, 200, 80),
        color: '#06B6D4',
      },
      {
        name: 'Документы',
        data: generateTrendData(30, 150, 60),
        color: '#F97316',
      },
    ],
  },
  decorators: [
    (Story) => (
      <div className="w-[800px]">
        <Story />
      </div>
    ),
  ],
}

export const DashboardExample: Story = {
  render: () => (
    <div className="space-y-6 p-6 bg-gray-100 dark:bg-gray-900 rounded-xl">
      <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Аналитика</h2>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <TrendChart
          title="Документы за 2 недели"
          datasets={[
            {
              name: 'Создано',
              data: generateTrendData(14, 45, 20),
              color: '#3B82F6',
            },
            {
              name: 'Обработано',
              data: generateTrendData(14, 40, 18),
              color: '#10B981',
            },
          ]}
        />
        <TrendChart
          title="Мероприятия"
          datasets={[
            {
              name: 'Запланировано',
              data: generateTrendData(14, 12, 8),
              color: '#8B5CF6',
            },
          ]}
        />
      </div>
    </div>
  ),
}
