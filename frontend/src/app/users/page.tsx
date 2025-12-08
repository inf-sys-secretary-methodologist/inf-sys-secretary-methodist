'use client'

import { useEffect, useState, useCallback } from 'react'
import { withAuth } from '@/components/auth/withAuth'
import { UserRole } from '@/types/auth'
import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
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
  Mail,
  Calendar,
  MoreVertical,
  Loader2,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { usersApi, departmentsApi, UserWithOrg, Department } from '@/lib/api/users'
import { toast } from 'sonner'

const roleLabels: Record<string, string> = {
  system_admin: 'Системный администратор',
  methodist: 'Методист',
  academic_secretary: 'Учёный секретарь',
  teacher: 'Преподаватель',
  student: 'Студент',
}

const roleColors: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  system_admin: 'destructive',
  methodist: 'default',
  academic_secretary: 'secondary',
  teacher: 'outline',
  student: 'outline',
}

const statusLabels: Record<string, string> = {
  active: 'Активен',
  inactive: 'Неактивен',
  blocked: 'Заблокирован',
}

function UsersManagementPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [roleFilter, setRoleFilter] = useState<string>('all')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [departmentFilter, setDepartmentFilter] = useState<string>('all')
  const [users, setUsers] = useState<UserWithOrg[]>([])
  const [departments, setDepartments] = useState<Department[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)
  const limit = 10

  const fetchUsers = useCallback(async () => {
    setLoading(true)
    try {
      const response = await usersApi.list({
        page,
        limit,
        role: roleFilter !== 'all' ? roleFilter : undefined,
        status: statusFilter !== 'all' ? statusFilter : undefined,
        department_id: departmentFilter !== 'all' ? parseInt(departmentFilter) : undefined,
        search: searchQuery || undefined,
      })
      setUsers(response.data.users || [])
      setTotal(response.data.total)
      setTotalPages(response.data.total_pages)
    } catch (error) {
      console.error('Failed to fetch users:', error)
      toast.error('Не удалось загрузить список пользователей')
    } finally {
      setLoading(false)
    }
  }, [page, limit, roleFilter, statusFilter, departmentFilter, searchQuery])

  const fetchReferenceData = useCallback(async () => {
    try {
      const deptResponse = await departmentsApi.list(1, 100, true)
      setDepartments(deptResponse.data.departments || [])
    } catch (error) {
      console.error('Failed to fetch reference data:', error)
    }
  }, [])

  useEffect(() => {
    fetchReferenceData()
  }, [fetchReferenceData])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  const handleUpdateStatus = async (userId: number, newStatus: string) => {
    try {
      await usersApi.updateStatus(userId, newStatus)
      toast.success('Статус пользователя обновлен')
      fetchUsers()
    } catch (error) {
      console.error('Failed to update status:', error)
      toast.error('Не удалось обновить статус')
    }
  }

  const handleDeleteUser = async (userId: number) => {
    if (!confirm('Вы уверены, что хотите удалить этого пользователя?')) return

    try {
      await usersApi.delete(userId)
      toast.success('Пользователь удален')
      fetchUsers()
    } catch (error) {
      console.error('Failed to delete user:', error)
      toast.error('Не удалось удалить пользователя')
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  const resetFilters = () => {
    setSearchQuery('')
    setRoleFilter('all')
    setStatusFilter('all')
    setDepartmentFilter('all')
    setPage(1)
  }

  const roleStats = users.reduce(
    (acc, user) => {
      acc[user.role] = (acc[user.role] || 0) + 1
      return acc
    },
    {} as Record<string, number>
  )

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight flex items-center gap-2">
              <UsersIcon className="h-6 w-6 sm:h-8 sm:w-8" />
              Управление пользователями
            </h1>
            <p className="text-sm sm:text-base text-muted-foreground mt-1">
              Просмотр и управление пользователями системы
            </p>
          </div>
          <Button variant="outline" size="sm" onClick={fetchUsers} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Обновить
          </Button>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 sm:gap-4">
          <Card>
            <CardHeader className="pb-2 sm:pb-3">
              <CardDescription className="text-xs sm:text-sm">Всего</CardDescription>
              <CardTitle className="text-2xl sm:text-3xl">{total}</CardTitle>
            </CardHeader>
          </Card>

          {Object.entries(roleLabels).map(([role, label]) => {
            const count = roleStats[role] || 0

            return (
              <Card key={role}>
                <CardHeader className="pb-2 sm:pb-3">
                  <CardDescription className="text-xs sm:text-sm truncate">{label}</CardDescription>
                  <CardTitle className="text-2xl sm:text-3xl">{count}</CardTitle>
                </CardHeader>
              </Card>
            )
          })}
        </div>

        {/* Filters Card */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg sm:text-xl">Фильтры</CardTitle>
            <CardDescription>Поиск и фильтрация пользователей</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              {/* Search */}
              <div className="space-y-2">
                <Label htmlFor="search" className="text-sm">
                  Поиск
                </Label>
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    id="search"
                    placeholder="Имя или email..."
                    value={searchQuery}
                    onChange={(e) => {
                      setSearchQuery(e.target.value)
                      setPage(1)
                    }}
                    className="pl-9"
                  />
                </div>
              </div>

              {/* Role Filter */}
              <div className="space-y-2">
                <Label htmlFor="role-filter" className="text-sm">
                  Роль
                </Label>
                <div className="relative">
                  <Filter className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground z-10" />
                  <Select
                    value={roleFilter}
                    onValueChange={(value) => {
                      setRoleFilter(value)
                      setPage(1)
                    }}
                  >
                    <SelectTrigger id="role-filter" className="pl-9">
                      <SelectValue placeholder="Выберите роль" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">Все роли</SelectItem>
                      {Object.entries(roleLabels).map(([role, label]) => (
                        <SelectItem key={role} value={role}>
                          {label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              {/* Status Filter */}
              <div className="space-y-2">
                <Label htmlFor="status-filter" className="text-sm">
                  Статус
                </Label>
                <Select
                  value={statusFilter}
                  onValueChange={(value) => {
                    setStatusFilter(value)
                    setPage(1)
                  }}
                >
                  <SelectTrigger id="status-filter">
                    <SelectValue placeholder="Выберите статус" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Все статусы</SelectItem>
                    {Object.entries(statusLabels).map(([status, label]) => (
                      <SelectItem key={status} value={status}>
                        {label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Department Filter */}
              <div className="space-y-2">
                <Label htmlFor="department-filter" className="text-sm">
                  Подразделение
                </Label>
                <Select
                  value={departmentFilter}
                  onValueChange={(value) => {
                    setDepartmentFilter(value)
                    setPage(1)
                  }}
                >
                  <SelectTrigger id="department-filter">
                    <SelectValue placeholder="Выберите подразделение" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Все подразделения</SelectItem>
                    {departments.map((dept) => (
                      <SelectItem key={dept.id} value={dept.id.toString()}>
                        {dept.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            {(searchQuery ||
              roleFilter !== 'all' ||
              statusFilter !== 'all' ||
              departmentFilter !== 'all') && (
              <div className="mt-4 flex flex-col sm:flex-row items-start sm:items-center gap-2 text-sm text-muted-foreground">
                <span>Найдено пользователей: {total}</span>
                <Button variant="ghost" size="sm" onClick={resetFilters}>
                  Сбросить фильтры
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Users Table */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg sm:text-xl">Пользователи</CardTitle>
            <CardDescription>Список всех зарегистрированных пользователей</CardDescription>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="flex items-center justify-center py-12">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            ) : (
              <>
                {/* Mobile Cards View */}
                <div className="block sm:hidden space-y-4">
                  {users.length === 0 ? (
                    <p className="text-center text-muted-foreground py-8">
                      Пользователи не найдены
                    </p>
                  ) : (
                    users.map((user) => (
                      <div key={user.id} className="border rounded-lg p-4 space-y-3">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center">
                              <span className="text-sm font-semibold text-primary">
                                {user.name.charAt(0)}
                              </span>
                            </div>
                            <span className="font-medium">{user.name}</span>
                          </div>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="sm">
                                <MoreVertical className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem
                                onClick={() =>
                                  handleUpdateStatus(
                                    user.id,
                                    user.status === 'active' ? 'inactive' : 'active'
                                  )
                                }
                              >
                                {user.status === 'active' ? 'Деактивировать' : 'Активировать'}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() =>
                                  handleUpdateStatus(
                                    user.id,
                                    user.status === 'blocked' ? 'active' : 'blocked'
                                  )
                                }
                              >
                                {user.status === 'blocked' ? 'Разблокировать' : 'Заблокировать'}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="text-destructive"
                                onClick={() => handleDeleteUser(user.id)}
                              >
                                Удалить
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </div>
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                          <Mail className="h-4 w-4" />
                          <span className="break-all">{user.email}</span>
                        </div>
                        {user.department_name && (
                          <div className="text-sm text-muted-foreground">
                            {user.department_name}
                          </div>
                        )}
                        <div className="flex items-center justify-between">
                          <Badge variant={roleColors[user.role] || 'outline'}>
                            {roleLabels[user.role] || user.role}
                          </Badge>
                          <div className="flex items-center gap-1 text-xs text-muted-foreground">
                            <Calendar className="h-3 w-3" />
                            {formatDate(user.created_at)}
                          </div>
                        </div>
                      </div>
                    ))
                  )}
                </div>

                {/* Desktop Table View */}
                <div className="hidden sm:block rounded-md border overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Пользователь</TableHead>
                        <TableHead>Email</TableHead>
                        <TableHead>Роль</TableHead>
                        <TableHead>Статус</TableHead>
                        <TableHead className="hidden lg:table-cell">Подразделение</TableHead>
                        <TableHead className="hidden lg:table-cell">Дата создания</TableHead>
                        <TableHead className="text-right">Действия</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {users.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={7} className="h-24 text-center text-muted-foreground">
                            Пользователи не найдены
                          </TableCell>
                        </TableRow>
                      ) : (
                        users.map((user) => (
                          <TableRow key={user.id}>
                            <TableCell className="font-medium">
                              <div className="flex items-center gap-2">
                                <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center">
                                  <span className="text-sm font-semibold text-primary">
                                    {user.name.charAt(0)}
                                  </span>
                                </div>
                                <span className="truncate max-w-[150px]">{user.name}</span>
                              </div>
                            </TableCell>
                            <TableCell>
                              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                <Mail className="h-4 w-4 flex-shrink-0" />
                                <span className="truncate max-w-[150px]">{user.email}</span>
                              </div>
                            </TableCell>
                            <TableCell>
                              <Badge variant={roleColors[user.role] || 'outline'}>
                                {roleLabels[user.role] || user.role}
                              </Badge>
                            </TableCell>
                            <TableCell>
                              <Badge
                                variant={
                                  user.status === 'active'
                                    ? 'default'
                                    : user.status === 'blocked'
                                      ? 'destructive'
                                      : 'secondary'
                                }
                              >
                                {statusLabels[user.status] || user.status}
                              </Badge>
                            </TableCell>
                            <TableCell className="hidden lg:table-cell text-sm text-muted-foreground">
                              {user.department_name || '-'}
                            </TableCell>
                            <TableCell className="hidden lg:table-cell">
                              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                <Calendar className="h-4 w-4" />
                                {formatDate(user.created_at)}
                              </div>
                            </TableCell>
                            <TableCell className="text-right">
                              <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                  <Button variant="ghost" size="sm">
                                    <MoreVertical className="h-4 w-4" />
                                  </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent align="end">
                                  <DropdownMenuItem
                                    onClick={() =>
                                      handleUpdateStatus(
                                        user.id,
                                        user.status === 'active' ? 'inactive' : 'active'
                                      )
                                    }
                                  >
                                    {user.status === 'active' ? 'Деактивировать' : 'Активировать'}
                                  </DropdownMenuItem>
                                  <DropdownMenuItem
                                    onClick={() =>
                                      handleUpdateStatus(
                                        user.id,
                                        user.status === 'blocked' ? 'active' : 'blocked'
                                      )
                                    }
                                  >
                                    {user.status === 'blocked' ? 'Разблокировать' : 'Заблокировать'}
                                  </DropdownMenuItem>
                                  <DropdownMenuSeparator />
                                  <DropdownMenuItem
                                    className="text-destructive"
                                    onClick={() => handleDeleteUser(user.id)}
                                  >
                                    Удалить
                                  </DropdownMenuItem>
                                </DropdownMenuContent>
                              </DropdownMenu>
                            </TableCell>
                          </TableRow>
                        ))
                      )}
                    </TableBody>
                  </Table>
                </div>

                {/* Pagination */}
                {totalPages > 1 && (
                  <div className="flex items-center justify-between mt-4">
                    <p className="text-sm text-muted-foreground">
                      Страница {page} из {totalPages}
                    </p>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setPage((p) => Math.max(1, p - 1))}
                        disabled={page === 1}
                      >
                        <ChevronLeft className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                        disabled={page === totalPages}
                      >
                        <ChevronRight className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                )}
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  )
}

// Only system admins can access this page
export default withAuth(UsersManagementPage, {
  roles: [UserRole.SYSTEM_ADMIN],
})
