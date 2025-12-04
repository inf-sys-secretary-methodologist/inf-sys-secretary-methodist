import type { Meta, StoryObj } from '@storybook/nextjs'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from './card'
import { Button } from './button'
import { Input } from './input'
import { Label } from './label'

const meta: Meta<typeof Card> = {
  title: 'UI/Card',
  component: Card,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Компонент карточки для отображения контента в контейнере.',
      },
    },
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <Card className="w-[350px]">
      <CardHeader>
        <CardTitle>Заголовок карточки</CardTitle>
        <CardDescription>Описание карточки.</CardDescription>
      </CardHeader>
      <CardContent>
        <p>Содержимое карточки.</p>
      </CardContent>
      <CardFooter>
        <Button>Действие</Button>
      </CardFooter>
    </Card>
  ),
}

export const SimpleCard: Story = {
  render: () => (
    <Card className="w-[350px] p-6">
      <p>Простая карточка с контентом и отступами.</p>
    </Card>
  ),
}

export const WithForm: Story = {
  render: () => (
    <Card className="w-[350px]">
      <CardHeader>
        <CardTitle>Создать проект</CardTitle>
        <CardDescription>Разверните новый проект одним кликом.</CardDescription>
      </CardHeader>
      <CardContent>
        <form>
          <div className="grid w-full items-center gap-4">
            <div className="flex flex-col space-y-1.5">
              <Label htmlFor="name">Название</Label>
              <Input id="name" placeholder="Название вашего проекта" />
            </div>
          </div>
        </form>
      </CardContent>
      <CardFooter className="flex justify-between">
        <Button variant="outline">Отмена</Button>
        <Button>Создать</Button>
      </CardFooter>
    </Card>
  ),
}

export const NotificationCard: Story = {
  render: () => (
    <Card className="w-[380px]">
      <CardHeader>
        <CardTitle>Уведомления</CardTitle>
        <CardDescription>У вас 3 непрочитанных сообщения.</CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4">
        <div className="flex items-center space-x-4 rounded-md border p-4">
          <div className="flex-1 space-y-1">
            <p className="text-sm font-medium leading-none">Push-уведомления</p>
            <p className="text-sm text-muted-foreground">Отправлять уведомления на устройство.</p>
          </div>
        </div>
        <div className="flex items-center space-x-4 rounded-md border p-4">
          <div className="flex-1 space-y-1">
            <p className="text-sm font-medium leading-none">Email-уведомления</p>
            <p className="text-sm text-muted-foreground">Получать письма об активности.</p>
          </div>
        </div>
      </CardContent>
      <CardFooter>
        <Button className="w-full">Отметить все как прочитанные</Button>
      </CardFooter>
    </Card>
  ),
}

export const StatsCard: Story = {
  render: () => (
    <Card className="w-[200px]">
      <CardHeader className="pb-2">
        <CardDescription>Общий доход</CardDescription>
        <CardTitle className="text-4xl">₽45,231.89</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-xs text-muted-foreground">+20.1% за последний месяц</div>
      </CardContent>
    </Card>
  ),
}
