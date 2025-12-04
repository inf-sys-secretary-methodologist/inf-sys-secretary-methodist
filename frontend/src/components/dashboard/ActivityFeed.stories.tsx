import type { Meta, StoryObj } from '@storybook/nextjs'
import { ActivityFeed } from './ActivityFeed'
import type { ActivityItem } from '@/types/dashboard'

const meta: Meta<typeof ActivityFeed> = {
  title: 'Dashboard/ActivityFeed',
  component: ActivityFeed,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `
Лента последних действий для дашборда.

**Используемые компоненты:**
- \`GlowingEffect\` - эффект свечения границы при наведении

**Типы активности:**
- \`document\` - Документы (FileText)
- \`report\` - Отчеты (ClipboardList)
- \`task\` - Задачи (ClipboardList)
- \`event\` - Мероприятия (Calendar)
- \`announcement\` - Объявления (Megaphone)

**Действия:**
- \`created\` - создан(о)
- \`updated\` - обновлен(о)
- \`deleted\` - удален(о)

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
      description: 'Заголовок ленты',
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

const sampleActivities: ActivityItem[] = [
  {
    id: 1,
    type: 'document',
    action: 'created',
    title: 'Рабочая программа по математике',
    description: 'Новая версия для 2024-2025 учебного года',
    user_id: 1,
    user_name: 'Иванова А.П.',
    created_at: new Date(Date.now() - 5 * 60000).toISOString(), // 5 минут назад
  },
  {
    id: 2,
    type: 'report',
    action: 'updated',
    title: 'Квартальный отчет Q3',
    description: 'Обновлены данные по успеваемости',
    user_id: 2,
    user_name: 'Петров С.В.',
    created_at: new Date(Date.now() - 2 * 3600000).toISOString(), // 2 часа назад
  },
  {
    id: 3,
    type: 'event',
    action: 'created',
    title: 'Педагогический совет',
    user_id: 3,
    user_name: 'Сидорова М.К.',
    created_at: new Date(Date.now() - 5 * 3600000).toISOString(), // 5 часов назад
  },
  {
    id: 4,
    type: 'announcement',
    action: 'created',
    title: 'Изменение расписания занятий',
    description: 'С 15 ноября вступает в силу новое расписание',
    user_id: 1,
    user_name: 'Иванова А.П.',
    created_at: new Date(Date.now() - 24 * 3600000).toISOString(), // вчера
  },
  {
    id: 5,
    type: 'task',
    action: 'updated',
    title: 'Подготовка к аккредитации',
    description: 'Завершен сбор документов',
    user_id: 4,
    user_name: 'Козлов Д.А.',
    created_at: new Date(Date.now() - 48 * 3600000).toISOString(), // 2 дня назад
  },
]

export const Default: Story = {
  args: {
    activities: sampleActivities,
  },
  decorators: [
    (Story) => (
      <div className="w-[450px]">
        <Story />
      </div>
    ),
  ],
}

export const CustomTitle: Story = {
  args: {
    activities: sampleActivities.slice(0, 3),
    title: 'Недавние изменения',
  },
  decorators: [
    (Story) => (
      <div className="w-[450px]">
        <Story />
      </div>
    ),
  ],
}

export const Empty: Story = {
  args: {
    activities: [],
  },
  decorators: [
    (Story) => (
      <div className="w-[450px]">
        <Story />
      </div>
    ),
  ],
}

export const DocumentsOnly: Story = {
  args: {
    activities: [
      {
        id: 1,
        type: 'document',
        action: 'created',
        title: 'Учебный план 2024',
        user_id: 1,
        user_name: 'Иванова А.П.',
        created_at: new Date().toISOString(),
      },
      {
        id: 2,
        type: 'document',
        action: 'updated',
        title: 'Методические рекомендации',
        description: 'Добавлен новый раздел',
        user_id: 2,
        user_name: 'Петров С.В.',
        created_at: new Date(Date.now() - 30 * 60000).toISOString(),
      },
      {
        id: 3,
        type: 'document',
        action: 'deleted',
        title: 'Устаревший шаблон',
        user_id: 1,
        user_name: 'Иванова А.П.',
        created_at: new Date(Date.now() - 3600000).toISOString(),
      },
    ],
    title: 'Документы',
  },
  decorators: [
    (Story) => (
      <div className="w-[450px]">
        <Story />
      </div>
    ),
  ],
}

export const ManyActivities: Story = {
  args: {
    activities: [
      ...sampleActivities,
      {
        id: 6,
        type: 'document',
        action: 'created',
        title: 'Справка об обучении',
        user_id: 5,
        user_name: 'Николаев И.С.',
        created_at: new Date(Date.now() - 72 * 3600000).toISOString(),
      },
      {
        id: 7,
        type: 'report',
        action: 'created',
        title: 'Отчет по практике',
        description: 'Группа ПМ-21',
        user_id: 3,
        user_name: 'Сидорова М.К.',
        created_at: new Date(Date.now() - 96 * 3600000).toISOString(),
      },
      {
        id: 8,
        type: 'event',
        action: 'updated',
        title: 'День открытых дверей',
        user_id: 2,
        user_name: 'Петров С.В.',
        created_at: new Date(Date.now() - 120 * 3600000).toISOString(),
      },
    ],
    title: 'Вся активность',
  },
  decorators: [
    (Story) => (
      <div className="w-[450px]">
        <Story />
      </div>
    ),
  ],
}

export const DashboardLayout: Story = {
  render: () => (
    <div className="p-6 bg-gray-100 dark:bg-gray-900 rounded-xl">
      <h2 className="text-2xl font-bold mb-6 text-gray-900 dark:text-white">Дашборд</h2>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ActivityFeed activities={sampleActivities.slice(0, 4)} title="Последние действия" />
        <ActivityFeed
          activities={sampleActivities.filter((a) => a.type === 'document')}
          title="Документы"
        />
      </div>
    </div>
  ),
}
