import type { Meta, StoryObj } from '@storybook/nextjs'
import {
  Table,
  TableHeader,
  TableBody,
  TableFooter,
  TableHead,
  TableRow,
  TableCell,
  TableCaption,
} from './table'
import { Badge } from './badge'

const meta: Meta<typeof Table> = {
  title: 'UI/Table',
  component: Table,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Компонент таблицы для отображения данных.',
      },
    },
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof meta>

const students = [
  { id: 1, name: 'Иванов Иван', group: 'ПМ-21', status: 'Активный' },
  { id: 2, name: 'Петрова Мария', group: 'ПМ-21', status: 'Активный' },
  { id: 3, name: 'Сидоров Петр', group: 'ИС-22', status: 'Отчислен' },
  { id: 4, name: 'Козлова Анна', group: 'ПМ-22', status: 'Активный' },
]

export const Default: Story = {
  render: () => (
    <Table>
      <TableCaption>Список студентов</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">ID</TableHead>
          <TableHead>Имя</TableHead>
          <TableHead>Группа</TableHead>
          <TableHead className="text-right">Статус</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {students.map((student) => (
          <TableRow key={student.id}>
            <TableCell className="font-medium">{student.id}</TableCell>
            <TableCell>{student.name}</TableCell>
            <TableCell>{student.group}</TableCell>
            <TableCell className="text-right">{student.status}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  ),
}

export const WithBadges: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Имя</TableHead>
          <TableHead>Группа</TableHead>
          <TableHead>Статус</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {students.map((student) => (
          <TableRow key={student.id}>
            <TableCell className="font-medium">{student.name}</TableCell>
            <TableCell>{student.group}</TableCell>
            <TableCell>
              <Badge variant={student.status === 'Активный' ? 'default' : 'destructive'}>
                {student.status}
              </Badge>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  ),
}

const invoices = [
  { invoice: 'INV001', status: 'Оплачено', method: 'Карта', amount: 250.0 },
  { invoice: 'INV002', status: 'Ожидает', method: 'Перевод', amount: 150.0 },
  { invoice: 'INV003', status: 'Оплачено', method: 'Карта', amount: 350.0 },
]

export const WithFooter: Story = {
  render: () => (
    <Table>
      <TableCaption>Счета за обучение</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">Счет</TableHead>
          <TableHead>Статус</TableHead>
          <TableHead>Метод</TableHead>
          <TableHead className="text-right">Сумма</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {invoices.map((invoice) => (
          <TableRow key={invoice.invoice}>
            <TableCell className="font-medium">{invoice.invoice}</TableCell>
            <TableCell>{invoice.status}</TableCell>
            <TableCell>{invoice.method}</TableCell>
            <TableCell className="text-right">{invoice.amount.toFixed(2)} $</TableCell>
          </TableRow>
        ))}
      </TableBody>
      <TableFooter>
        <TableRow>
          <TableCell colSpan={3}>Итого</TableCell>
          <TableCell className="text-right">
            {invoices.reduce((sum, inv) => sum + inv.amount, 0).toFixed(2)} $
          </TableCell>
        </TableRow>
      </TableFooter>
    </Table>
  ),
}

export const Documents: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Название</TableHead>
          <TableHead>Тип</TableHead>
          <TableHead>Дата</TableHead>
          <TableHead>Автор</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell className="font-medium">Учебный план 2024</TableCell>
          <TableCell>
            <Badge variant="secondary">План</Badge>
          </TableCell>
          <TableCell>15.01.2024</TableCell>
          <TableCell>Иванова А.П.</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">Расписание весна</TableCell>
          <TableCell>
            <Badge variant="secondary">Расписание</Badge>
          </TableCell>
          <TableCell>01.02.2024</TableCell>
          <TableCell>Петров С.В.</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">Отчет Q1</TableCell>
          <TableCell>
            <Badge variant="secondary">Отчет</Badge>
          </TableCell>
          <TableCell>31.03.2024</TableCell>
          <TableCell>Сидорова М.К.</TableCell>
        </TableRow>
      </TableBody>
    </Table>
  ),
}
