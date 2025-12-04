import type { Meta, StoryObj } from '@storybook/nextjs'
import { Separator } from './separator'

const meta: Meta<typeof Separator> = {
  title: 'UI/Separator',
  component: Separator,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Визуальный разделитель контента.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    orientation: {
      control: 'select',
      options: ['horizontal', 'vertical'],
      description: 'Ориентация разделителя',
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Horizontal: Story = {
  render: () => (
    <div className="w-[300px]">
      <div className="space-y-1">
        <h4 className="text-sm font-medium leading-none">Radix Primitives</h4>
        <p className="text-sm text-muted-foreground">Открытая библиотека UI компонентов.</p>
      </div>
      <Separator className="my-4" />
      <div className="flex h-5 items-center space-x-4 text-sm">
        <div>Документы</div>
        <Separator orientation="vertical" />
        <div>Студенты</div>
        <Separator orientation="vertical" />
        <div>Отчеты</div>
      </div>
    </div>
  ),
}

export const Vertical: Story = {
  render: () => (
    <div className="flex h-5 items-center space-x-4 text-sm">
      <div>Главная</div>
      <Separator orientation="vertical" />
      <div>Документы</div>
      <Separator orientation="vertical" />
      <div>Настройки</div>
    </div>
  ),
}

export const InCard: Story = {
  render: () => (
    <div className="w-[350px] rounded-lg border p-4">
      <h3 className="font-semibold">Информация о студенте</h3>
      <Separator className="my-3" />
      <div className="space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-muted-foreground">Имя:</span>
          <span>Иванов Иван</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Группа:</span>
          <span>ПМ-21</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Курс:</span>
          <span>3</span>
        </div>
      </div>
    </div>
  ),
}
