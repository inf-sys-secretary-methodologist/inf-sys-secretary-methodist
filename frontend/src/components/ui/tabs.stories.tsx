import type { Meta, StoryObj } from '@storybook/nextjs'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './tabs'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './card'
import { Input } from './input'
import { Label } from './label'

const meta: Meta<typeof Tabs> = {
  title: 'UI/Tabs',
  component: Tabs,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Компонент вкладок для организации контента в переключаемые панели.',
      },
    },
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <Tabs defaultValue="account" className="w-[400px]">
      <TabsList>
        <TabsTrigger value="account">Аккаунт</TabsTrigger>
        <TabsTrigger value="password">Пароль</TabsTrigger>
      </TabsList>
      <TabsContent value="account">
        <Card>
          <CardHeader>
            <CardTitle>Аккаунт</CardTitle>
            <CardDescription>
              Внесите изменения в ваш аккаунт. Нажмите сохранить когда закончите.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="space-y-1">
              <Label htmlFor="name">Имя</Label>
              <Input id="name" defaultValue="Иван Иванов" />
            </div>
            <div className="space-y-1">
              <Label htmlFor="username">Логин</Label>
              <Input id="username" defaultValue="@ivanov" />
            </div>
          </CardContent>
        </Card>
      </TabsContent>
      <TabsContent value="password">
        <Card>
          <CardHeader>
            <CardTitle>Пароль</CardTitle>
            <CardDescription>
              Измените пароль здесь. После сохранения вы будете разлогинены.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="space-y-1">
              <Label htmlFor="current">Текущий пароль</Label>
              <Input id="current" type="password" />
            </div>
            <div className="space-y-1">
              <Label htmlFor="new">Новый пароль</Label>
              <Input id="new" type="password" />
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  ),
}

export const Simple: Story = {
  render: () => (
    <Tabs defaultValue="tab1" className="w-[400px]">
      <TabsList className="grid w-full grid-cols-3">
        <TabsTrigger value="tab1">Вкладка 1</TabsTrigger>
        <TabsTrigger value="tab2">Вкладка 2</TabsTrigger>
        <TabsTrigger value="tab3">Вкладка 3</TabsTrigger>
      </TabsList>
      <TabsContent value="tab1" className="p-4">
        <p>Содержимое вкладки 1</p>
      </TabsContent>
      <TabsContent value="tab2" className="p-4">
        <p>Содержимое вкладки 2</p>
      </TabsContent>
      <TabsContent value="tab3" className="p-4">
        <p>Содержимое вкладки 3</p>
      </TabsContent>
    </Tabs>
  ),
}

export const WithDisabled: Story = {
  render: () => (
    <Tabs defaultValue="active" className="w-[400px]">
      <TabsList>
        <TabsTrigger value="active">Активная</TabsTrigger>
        <TabsTrigger value="disabled" disabled>
          Отключена
        </TabsTrigger>
        <TabsTrigger value="another">Другая</TabsTrigger>
      </TabsList>
      <TabsContent value="active" className="p-4">
        <p>Эта вкладка активна</p>
      </TabsContent>
      <TabsContent value="another" className="p-4">
        <p>Содержимое другой вкладки</p>
      </TabsContent>
    </Tabs>
  ),
}
