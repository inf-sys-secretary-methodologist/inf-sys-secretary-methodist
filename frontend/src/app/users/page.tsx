'use client'

import { useState } from 'react'
import { withAuth } from '@/components/auth/withAuth'
import { UserRole } from '@/types/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import {
  Search,
  Filter,
  Users as UsersIcon,
  Shield,
  Mail,
  Calendar,
  MoreVertical,
} from 'lucide-react'

// Mock data - TODO: Replace with API calls
const mockUsers = [
  {
    id: '1',
    email: 'admin@example.com',
    name: 'Администратор Системы',
    role: UserRole.SYSTEM_ADMIN,
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-11-16T14:30:00Z',
  },
  {
    id: '2',
    email: 'methodist@example.com',
    name: 'Петров Иван Сергеевич',
    role: UserRole.METHODIST,
    createdAt: '2024-02-20T09:00:00Z',
    updatedAt: '2024-11-10T11:20:00Z',
  },
  {
    id: '3',
    email: 'secretary@example.com',
    name: 'Сидорова Мария Ивановна',
    role: UserRole.ACADEMIC_SECRETARY,
    createdAt: '2024-02-21T09:00:00Z',
    updatedAt: '2024-11-12T16:45:00Z',
  },
  {
    id: '4',
    email: 'teacher@example.com',
    name: 'Смирнов Алексей Петрович',
    role: UserRole.TEACHER,
    createdAt: '2024-03-01T10:00:00Z',
    updatedAt: '2024-11-14T09:15:00Z',
  },
  {
    id: '5',
    email: 'student@example.com',
    name: 'Иванов Дмитрий Александрович',
    role: UserRole.STUDENT,
    createdAt: '2024-09-01T08:00:00Z',
    updatedAt: '2024-11-15T13:00:00Z',
  },
]

const roleLabels: Record<UserRole, string> = {
  [UserRole.SYSTEM_ADMIN]: 'Системный администратор',
  [UserRole.METHODIST]: 'Методист',
  [UserRole.ACADEMIC_SECRETARY]: 'Учёный секретарь',
  [UserRole.TEACHER]: 'Преподаватель',
  [UserRole.STUDENT]: 'Студент',
}

const roleColors: Record<UserRole, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  [UserRole.SYSTEM_ADMIN]: 'destructive',
  [UserRole.METHODIST]: 'default',
  [UserRole.ACADEMIC_SECRETARY]: 'secondary',
  [UserRole.TEACHER]: 'outline',
  [UserRole.STUDENT]: 'outline',
}

function UsersManagementPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [roleFilter, setRoleFilter] = useState<string>('all')
  const [users] = useState(mockUsers) // TODO: Fetch from API

  const filteredUsers = users.filter((user) => {
    const matchesSearch =
      user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.email.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesRole = roleFilter === 'all' || user.role === roleFilter

    return matchesSearch && matchesRole
  })

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  const handleRoleChange = (userId: string, newRole: UserRole) => {
    // TODO: Implement API call to update user role
    console.log('Updating user role:', { userId, newRole })

    // TODO: Update users state after successful API call
  }

  return (
    <div className="container mx-auto p-6">
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <UsersIcon className="h-8 w-8" />
            Управление пользователями
          </h1>
          <p className="text-muted-foreground mt-1">
            Просмотр и управление пользователями системы
          </p>
        </div>

        {/* Stats Cards */}
        <div className="grid gap-4 md:grid-cols-5">
          <Card>
            <CardHeader className="pb-3">
              <CardDescription>Всего пользователей</CardDescription>
              <CardTitle className="text-3xl">{users.length}</CardTitle>
            </CardHeader>
          </Card>

          {Object.entries(UserRole).map(([key, value]) => {
            const count = users.filter((u) => u.role === value).length
            if (count === 0) return null

            return (
              <Card key={key}>
                <CardHeader className="pb-3">
                  <CardDescription>{roleLabels[value]}</CardDescription>
                  <CardTitle className="text-3xl">{count}</CardTitle>
                </CardHeader>
              </Card>
            )
          })}
        </div>

        {/* Filters Card */}
        <Card>
          <CardHeader>
            <CardTitle>Фильтры</CardTitle>
            <CardDescription>
              Поиск и фильтрация пользователей
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2">
              {/* Search */}
              <div className="space-y-2">
                <Label htmlFor="search">Поиск</Label>
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    id="search"
                    placeholder="Имя или email..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-9"
                  />
                </div>
              </div>

              {/* Role Filter */}
              <div className="space-y-2">
                <Label htmlFor="role-filter">Роль</Label>
                <div className="relative">
                  <Filter className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground z-10" />
                  <Select value={roleFilter} onValueChange={setRoleFilter}>
                    <SelectTrigger id="role-filter" className="pl-9">
                      <SelectValue placeholder="Выберите роль" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">Все роли</SelectItem>
                      {Object.entries(UserRole).map(([key, value]) => (
                        <SelectItem key={key} value={value}>
                          {roleLabels[value]}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </div>

            {searchQuery || roleFilter !== 'all' ? (
              <div className="mt-4 flex items-center gap-2 text-sm text-muted-foreground">
                <span>Найдено пользователей: {filteredUsers.length}</span>
                {(searchQuery || roleFilter !== 'all') && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => {
                      setSearchQuery('')
                      setRoleFilter('all')
                    }}
                  >
                    Сбросить фильтры
                  </Button>
                )}
              </div>
            ) : null}
          </CardContent>
        </Card>

        {/* Users Table */}
        <Card>
          <CardHeader>
            <CardTitle>Пользователи</CardTitle>
            <CardDescription>
              Список всех зарегистрированных пользователей
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Пользователь</TableHead>
                    <TableHead>Email</TableHead>
                    <TableHead>Роль</TableHead>
                    <TableHead>Дата создания</TableHead>
                    <TableHead>Посл. обновление</TableHead>
                    <TableHead className="text-right">Действия</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredUsers.length === 0 ? (
                    <TableRow>
                      <TableCell
                        colSpan={6}
                        className="h-24 text-center text-muted-foreground"
                      >
                        Пользователи не найдены
                      </TableCell>
                    </TableRow>
                  ) : (
                    filteredUsers.map((user) => (
                      <TableRow key={user.id}>
                        <TableCell className="font-medium">
                          <div className="flex items-center gap-2">
                            <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center">
                              <span className="text-sm font-semibold text-primary">
                                {user.name.charAt(0)}
                              </span>
                            </div>
                            <span>{user.name}</span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2 text-sm text-muted-foreground">
                            <Mail className="h-4 w-4" />
                            {user.email}
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant={roleColors[user.role]}>
                            {roleLabels[user.role]}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2 text-sm text-muted-foreground">
                            <Calendar className="h-4 w-4" />
                            {formatDate(user.createdAt)}
                          </div>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatDate(user.updatedAt)}
                        </TableCell>
                        <TableCell className="text-right">
                          <Button variant="ghost" size="sm">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

// Only system admins can access this page
export default withAuth(UsersManagementPage, {
  roles: [UserRole.SYSTEM_ADMIN],
})
