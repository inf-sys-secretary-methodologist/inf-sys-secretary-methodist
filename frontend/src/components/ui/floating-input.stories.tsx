import type { Meta, StoryObj } from '@storybook/nextjs'
import { FloatingInput } from './floating-input'
import { Button } from './button'

const meta: Meta<typeof FloatingInput> = {
  title: 'UI/FloatingInput',
  component: FloatingInput,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `
Input с плавающим label. При фокусе или наличии значения label
анимированно перемещается вверх и уменьшается.

**Особенности:**
- Плавная анимация label (200ms)
- Автоматический фон под label
- Поддержка всех стандартных input типов
- Доступность (связь label и input через htmlFor)
        `,
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    label: {
      control: 'text',
      description: 'Текст label',
    },
    type: {
      control: 'select',
      options: ['text', 'email', 'password', 'number', 'tel'],
      description: 'Тип input',
    },
    disabled: {
      control: 'boolean',
      description: 'Отключен',
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    label: 'Имя пользователя',
  },
}

export const Email: Story = {
  args: {
    label: 'Email',
    type: 'email',
  },
}

export const Password: Story = {
  args: {
    label: 'Пароль',
    type: 'password',
  },
}

export const WithValue: Story = {
  args: {
    label: 'Имя',
    defaultValue: 'Иван Иванов',
  },
}

export const Disabled: Story = {
  args: {
    label: 'Отключено',
    disabled: true,
    defaultValue: 'Нельзя редактировать',
  },
}

export const LoginForm: Story = {
  render: () => (
    <div className="w-[350px] p-6 rounded-2xl bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 space-y-4">
      <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-6">Вход в систему</h2>
      <FloatingInput label="Email" type="email" />
      <FloatingInput label="Пароль" type="password" />
      <Button className="w-full">Войти</Button>
    </div>
  ),
}

export const RegistrationForm: Story = {
  render: () => (
    <div className="w-[350px] p-6 rounded-2xl bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 space-y-4">
      <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-6">Регистрация</h2>
      <FloatingInput label="Имя" type="text" />
      <FloatingInput label="Фамилия" type="text" />
      <FloatingInput label="Email" type="email" />
      <FloatingInput label="Пароль" type="password" />
      <FloatingInput label="Подтверждение пароля" type="password" />
      <Button className="w-full">Зарегистрироваться</Button>
    </div>
  ),
}
