import type { Meta, StoryObj } from '@storybook/nextjs'
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
} from './select'
import { Label } from './label'

const meta: Meta<typeof Select> = {
  title: 'UI/Select',
  component: Select,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Компонент выбора на базе Radix UI Select.',
      },
    },
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <Select>
      <SelectTrigger className="w-[180px]">
        <SelectValue placeholder="Выберите..." />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="option1">Опция 1</SelectItem>
        <SelectItem value="option2">Опция 2</SelectItem>
        <SelectItem value="option3">Опция 3</SelectItem>
      </SelectContent>
    </Select>
  ),
}

export const WithGroups: Story = {
  render: () => (
    <Select>
      <SelectTrigger className="w-[280px]">
        <SelectValue placeholder="Выберите тип документа" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectLabel>Учебные документы</SelectLabel>
          <SelectItem value="syllabus">Учебный план</SelectItem>
          <SelectItem value="program">Рабочая программа</SelectItem>
          <SelectItem value="schedule">Расписание</SelectItem>
        </SelectGroup>
        <SelectSeparator />
        <SelectGroup>
          <SelectLabel>Отчеты</SelectLabel>
          <SelectItem value="report-q">Квартальный отчет</SelectItem>
          <SelectItem value="report-y">Годовой отчет</SelectItem>
        </SelectGroup>
        <SelectSeparator />
        <SelectGroup>
          <SelectLabel>Справки</SelectLabel>
          <SelectItem value="cert-study">Справка об обучении</SelectItem>
          <SelectItem value="cert-grade">Справка об успеваемости</SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  ),
}

export const WithLabel: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="status">Статус</Label>
      <Select>
        <SelectTrigger id="status">
          <SelectValue placeholder="Выберите статус" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="active">Активный</SelectItem>
          <SelectItem value="inactive">Неактивный</SelectItem>
          <SelectItem value="pending">Ожидает</SelectItem>
        </SelectContent>
      </Select>
    </div>
  ),
}

export const Disabled: Story = {
  render: () => (
    <Select disabled>
      <SelectTrigger className="w-[180px]">
        <SelectValue placeholder="Отключено" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="option1">Опция 1</SelectItem>
      </SelectContent>
    </Select>
  ),
}

export const FormExample: Story = {
  render: () => (
    <div className="w-[350px] space-y-4">
      <div className="grid gap-1.5">
        <Label>Курс</Label>
        <Select defaultValue="3">
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1">1 курс</SelectItem>
            <SelectItem value="2">2 курс</SelectItem>
            <SelectItem value="3">3 курс</SelectItem>
            <SelectItem value="4">4 курс</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div className="grid gap-1.5">
        <Label>Группа</Label>
        <Select>
          <SelectTrigger>
            <SelectValue placeholder="Выберите группу" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="pm-21">ПМ-21</SelectItem>
            <SelectItem value="pm-22">ПМ-22</SelectItem>
            <SelectItem value="is-21">ИС-21</SelectItem>
            <SelectItem value="is-22">ИС-22</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  ),
}
