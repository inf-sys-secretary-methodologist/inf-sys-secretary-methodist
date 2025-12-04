import type { Meta, StoryObj } from '@storybook/nextjs'
import { NavBar } from './tubelight-navbar'
import { Home, FileText, Calendar, BarChart3, Settings, Users, Bell } from 'lucide-react'

const meta: Meta<typeof NavBar> = {
  title: 'Navigation/NavBar',
  component: NavBar,
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component: `
Навигационная панель с эффектом "tubelight" (свечение).
Вдохновлено дизайном от 21st.dev.

**Особенности:**
- Фиксированная позиция вверху страницы
- Эффект свечения при наведении и активном состоянии
- Градиент \`from-blue-500 to-purple-600\` для активного элемента
- Адаптивность: иконки всегда видны, текст только на lg+
- Полупрозрачный фон с backdrop-blur
        `,
      },
    },
  },
  tags: ['autodocs'],
  decorators: [
    (Story) => (
      <div className="h-[300px] bg-gray-100 dark:bg-gray-900">
        <Story />
        <div className="pt-24 px-8">
          <p className="text-gray-600 dark:text-gray-400">
            Наведите на элементы навигации чтобы увидеть эффект свечения.
          </p>
        </div>
      </div>
    ),
  ],
}

export default meta
type Story = StoryObj<typeof meta>

const defaultItems = [
  { name: 'Главная', url: '/', icon: Home },
  { name: 'Документы', url: '/documents', icon: FileText },
  { name: 'Календарь', url: '/calendar', icon: Calendar },
  { name: 'Отчеты', url: '/reports', icon: BarChart3 },
]

export const Default: Story = {
  args: {
    items: defaultItems,
  },
}

export const WithMoreItems: Story = {
  args: {
    items: [
      { name: 'Главная', url: '/', icon: Home },
      { name: 'Документы', url: '/documents', icon: FileText },
      { name: 'Студенты', url: '/students', icon: Users },
      { name: 'Календарь', url: '/calendar', icon: Calendar },
      { name: 'Отчеты', url: '/reports', icon: BarChart3 },
      { name: 'Настройки', url: '/settings', icon: Settings },
    ],
  },
}

export const MinimalNav: Story = {
  args: {
    items: [
      { name: 'Главная', url: '/', icon: Home },
      { name: 'Уведомления', url: '/notifications', icon: Bell },
      { name: 'Настройки', url: '/settings', icon: Settings },
    ],
  },
}

export const OnDarkBackground: Story = {
  decorators: [
    (Story) => (
      <div className="h-[300px] bg-gray-900">
        <Story />
        <div className="pt-24 px-8">
          <p className="text-gray-400">Навигация на темном фоне с полупрозрачным backdrop-blur.</p>
        </div>
      </div>
    ),
  ],
  args: {
    items: defaultItems,
  },
}
