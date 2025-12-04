import type { Meta, StoryObj } from '@storybook/nextjs'
import { Popover, PopoverTrigger, PopoverContent } from './popover'
import { Button } from './button'
import { Input } from './input'
import { Label } from './label'

const meta: Meta<typeof Popover> = {
  title: 'UI/Popover',
  component: Popover,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Всплывающее окно на базе Radix UI Popover.',
      },
    },
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline">Открыть popover</Button>
      </PopoverTrigger>
      <PopoverContent className="w-80">
        <div className="grid gap-4">
          <div className="space-y-2">
            <h4 className="font-medium leading-none">Размеры</h4>
            <p className="text-sm text-muted-foreground">Установите размеры для слоя.</p>
          </div>
          <div className="grid gap-2">
            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="width">Ширина</Label>
              <Input id="width" defaultValue="100%" className="col-span-2 h-8" />
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="height">Высота</Label>
              <Input id="height" defaultValue="25px" className="col-span-2 h-8" />
            </div>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  ),
}

export const Simple: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="secondary">Информация</Button>
      </PopoverTrigger>
      <PopoverContent>
        <p className="text-sm">Это простой popover с текстовой информацией.</p>
      </PopoverContent>
    </Popover>
  ),
}

export const WithActions: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button>Фильтры</Button>
      </PopoverTrigger>
      <PopoverContent className="w-80">
        <div className="grid gap-4">
          <h4 className="font-medium">Фильтры</h4>
          <div className="grid gap-2">
            <Label htmlFor="status">Статус</Label>
            <Input id="status" placeholder="Выберите статус" />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="date">Дата</Label>
            <Input id="date" type="date" />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" size="sm">
              Сбросить
            </Button>
            <Button size="sm">Применить</Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  ),
}
