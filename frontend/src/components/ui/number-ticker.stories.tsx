import type { Meta, StoryObj } from '@storybook/nextjs'
import { NumberTicker } from './number-ticker'

const meta: Meta<typeof NumberTicker> = {
  title: 'Magic UI/NumberTicker',
  component: NumberTicker,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `
Анимированный счетчик чисел от Magic UI. Использует spring анимацию
для плавного перехода от начального значения к конечному.

**Особенности:**
- Анимация запускается когда элемент появляется в viewport
- Настраиваемое направление (вверх/вниз)
- Поддержка десятичных знаков
- Задержка перед анимацией
        `,
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    value: {
      control: { type: 'number', min: 0, max: 100000 },
      description: 'Конечное значение',
    },
    startValue: {
      control: { type: 'number', min: 0, max: 100000 },
      description: 'Начальное значение',
    },
    direction: {
      control: 'select',
      options: ['up', 'down'],
      description: 'Направление анимации',
    },
    delay: {
      control: { type: 'range', min: 0, max: 2, step: 0.1 },
      description: 'Задержка в секундах',
    },
    decimalPlaces: {
      control: { type: 'range', min: 0, max: 4, step: 1 },
      description: 'Количество десятичных знаков',
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    value: 1234,
    startValue: 0,
    direction: 'up',
  },
}

export const CountDown: Story = {
  args: {
    value: 0,
    startValue: 1000,
    direction: 'down',
  },
}

export const WithDecimals: Story = {
  args: {
    value: 99.99,
    startValue: 0,
    decimalPlaces: 2,
  },
}

export const WithDelay: Story = {
  args: {
    value: 5000,
    delay: 0.5,
  },
}

export const LargeNumber: Story = {
  args: {
    value: 1234567,
    startValue: 0,
  },
}

export const InStatsCard: Story = {
  render: () => (
    <div className="space-y-4">
      <div className="p-6 rounded-2xl bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
        <h3 className="text-sm font-medium mb-2 text-gray-600 dark:text-gray-400">
          Всего документов
        </h3>
        <div className="text-4xl font-bold text-gray-900 dark:text-white">
          <NumberTicker value={12847} />
        </div>
      </div>
      <div className="p-6 rounded-2xl bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
        <h3 className="text-sm font-medium mb-2 text-gray-600 dark:text-gray-400">
          Активных пользователей
        </h3>
        <div className="text-4xl font-bold text-gray-900 dark:text-white">
          <NumberTicker value={342} delay={0.3} />
        </div>
      </div>
    </div>
  ),
}
