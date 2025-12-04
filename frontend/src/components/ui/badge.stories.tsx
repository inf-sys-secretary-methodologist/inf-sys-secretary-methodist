import type { Meta, StoryObj } from '@storybook/nextjs'
import { Badge } from './badge'

const meta: Meta<typeof Badge> = {
  title: 'UI/Badge',
  component: Badge,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Компонент Badge для отображения меток, тегов или индикаторов статуса.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'secondary', 'destructive', 'outline'],
      description: 'Визуальный стиль бейджа',
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    children: 'Badge',
    variant: 'default',
  },
}

export const Secondary: Story = {
  args: {
    children: 'Secondary',
    variant: 'secondary',
  },
}

export const Destructive: Story = {
  args: {
    children: 'Destructive',
    variant: 'destructive',
  },
}

export const Outline: Story = {
  args: {
    children: 'Outline',
    variant: 'outline',
  },
}

export const AllVariants: Story = {
  render: () => (
    <div className="flex gap-2">
      <Badge variant="default">Default</Badge>
      <Badge variant="secondary">Secondary</Badge>
      <Badge variant="destructive">Destructive</Badge>
      <Badge variant="outline">Outline</Badge>
    </div>
  ),
}

export const StatusBadges: Story = {
  render: () => (
    <div className="flex gap-2">
      <Badge className="bg-green-500 hover:bg-green-600">Одобрено</Badge>
      <Badge className="bg-yellow-500 hover:bg-yellow-600">Ожидает</Badge>
      <Badge className="bg-red-500 hover:bg-red-600">Отклонено</Badge>
      <Badge className="bg-blue-500 hover:bg-blue-600">В процессе</Badge>
    </div>
  ),
}
