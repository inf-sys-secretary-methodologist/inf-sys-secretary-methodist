import type { Meta, StoryObj } from '@storybook/nextjs'
import { Input } from './input'
import { Label } from './label'
import { Button } from './button'
import { Search, Eye, EyeOff } from 'lucide-react'
import { useState } from 'react'

const meta: Meta<typeof Input> = {
  title: 'UI/Input',
  component: Input,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Компонент поля ввода для текста.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    type: {
      control: 'select',
      options: ['text', 'password', 'email', 'number', 'search', 'tel', 'url'],
      description: 'Тип поля ввода',
    },
    placeholder: {
      control: 'text',
      description: 'Текст-подсказка',
    },
    disabled: {
      control: 'boolean',
      description: 'Отключено ли поле',
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    placeholder: 'Введите текст...',
  },
}

export const Email: Story = {
  args: {
    type: 'email',
    placeholder: 'Email',
  },
}

export const Password: Story = {
  args: {
    type: 'password',
    placeholder: 'Пароль',
  },
}

export const Disabled: Story = {
  args: {
    placeholder: 'Отключенное поле',
    disabled: true,
  },
}

export const WithLabel: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="email">Email</Label>
      <Input type="email" id="email" placeholder="Email" />
    </div>
  ),
}

export const WithButton: Story = {
  render: () => (
    <div className="flex w-full max-w-sm items-center space-x-2">
      <Input type="email" placeholder="Email" />
      <Button type="submit">Подписаться</Button>
    </div>
  ),
}

export const WithIcon: Story = {
  render: () => (
    <div className="relative w-full max-w-sm">
      <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      <Input className="pl-9" placeholder="Поиск..." />
    </div>
  ),
}

export const File: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="picture">Изображение</Label>
      <Input id="picture" type="file" />
    </div>
  ),
}

export const PasswordWithToggle: Story = {
  render: function PasswordToggle() {
    const [showPassword, setShowPassword] = useState(false)
    return (
      <div className="relative w-full max-w-sm">
        <Input type={showPassword ? 'text' : 'password'} placeholder="Пароль" className="pr-10" />
        <button
          type="button"
          onClick={() => setShowPassword(!showPassword)}
          className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
        >
          {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
        </button>
      </div>
    )
  },
}

export const AllStates: Story = {
  render: () => (
    <div className="flex flex-col gap-4 w-[300px]">
      <Input placeholder="Обычное" />
      <Input placeholder="Отключенное" disabled />
      <Input placeholder="Со значением" defaultValue="Привет мир" />
      <Input type="password" placeholder="Пароль" defaultValue="secret" />
    </div>
  ),
}
