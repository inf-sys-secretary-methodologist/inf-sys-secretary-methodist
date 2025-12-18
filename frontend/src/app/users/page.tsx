'use client'

import { useEffect, useState, useCallback } from 'react'
import { withAuth } from '@/components/auth/withAuth'
import { useAuthStore } from '@/stores/authStore'
import { UserRole } from '@/types/auth'
import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
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
  X,
  Mail,
  Calendar,
  MoreVertical,
  Loader2,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { usersApi, departmentsApi, type UserWithOrg, type Department } from '@/lib/api/users'
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
  const { isAuthenticated, isLoading: authLoading } = useAuthStore()
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
  const [isFiltersExpanded, setIsFiltersExpanded] = useState(false)
  const limit = 10

  const hasActiveFilters =
    searchQuery || roleFilter !== 'all' || statusFilter !== 'all' || departmentFilter !== 'all'

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
    // Only fetch data when authenticated
    if (!authLoading && isAuthenticated) {
      fetchReferenceData()
    }
  }, [fetchReferenceData, authLoading, isAuthenticated])

  useEffect(() => {
    // Only fetch data when authenticated
    if (!authLoading && isAuthenticated) {
      fetchUsers()
    }
  }, [fetchUsers, authLoading, isAuthenticated])

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
      <div className="max-w-7xl mx-auto space-y-6 sm:space-y-8">
        {/* Page Header */}
        <div className="text-center space-y-2 sm:space-y-4">
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-gray-900 dark:text-white">
            Управление пользователями
          </h1>
          <p className="text-base sm:text-lg text-gray-600 dark:text-gray-300">
            Просмотр и управление пользователями системы
          </p>
        </div>

        {/* Action Button */}
        <div className="flex justify-end">
          <Button variant="outline" size="sm" onClick={fetchUsers} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Обновить
          </Button>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 sm:gap-4">
          <div className="relative overflow-hidden rounded-xl p-4 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
            <GlowingEffect
              spread={40}
              glow={true}
              disabled={false}
              proximity={64}
              inactiveZone={0.01}
              borderWidth={3}
            />
            <div className="relative z-10">
              <p className="text-xs sm:text-sm text-muted-foreground">Всего</p>
              <p className="text-2xl sm:text-3xl font-bold">{total}</p>
            </div>
          </div>

          {Object.entries(roleLabels).map(([role, label]) => {
            const count = roleStats[role] || 0

            return (
              <div
                key={role}
                className="relative overflow-hidden rounded-xl p-4 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700"
              >
                <GlowingEffect
                  spread={40}
                  glow={true}
                  disabled={false}
                  proximity={64}
                  inactiveZone={0.01}
                  borderWidth={3}
                />
                <div className="relative z-10">
                  <p className="text-xs sm:text-sm text-muted-foreground truncate">{label}</p>
                  <p className="text-2xl sm:text-3xl font-bold">{count}</p>
                </div>
              </div>
            )
          })}
        </div>

        {/* Filters Section */}
        <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
          <GlowingEffect
            spread={40}
            glow={true}
            disabled={false}
            proximity={64}
            inactiveZone={0.01}
            borderWidth={3}
          />
          <div className="relative z-10 space-y-4">
            {/* Search Bar */}
            <div className="flex gap-3">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => {
                    setSearchQuery(e.target.value)
                    setPage(1)
                  }}
                  placeholder="Поиск по имени или email..."
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                           bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                           focus:ring-2 focus:ring-blue-500 focus:border-transparent
                           placeholder:text-gray-400 dark:placeholder:text-gray-500"
                />
              </div>

              <Button
                variant={isFiltersExpanded ? 'default' : 'outline'}
                onClick={() => setIsFiltersExpanded(!isFiltersExpanded)}
                className="flex-shrink-0"
              >
                <Filter className="h-4 w-4 mr-2" />
                Фильтры
                {hasActiveFilters && !isFiltersExpanded && (
                  <span className="ml-2 px-2 py-0.5 bg-blue-500 text-white text-xs rounded-full">
                    {
                      [
                        roleFilter !== 'all',
                        statusFilter !== 'all',
                        departmentFilter !== 'all',
                      ].filter(Boolean).length
                    }
                  </span>
                )}
              </Button>

              {hasActiveFilters && (
                <Button variant="outline" onClick={resetFilters} className="flex-shrink-0">
                  <X className="h-4 w-4 mr-2" />
                  Сбросить
                </Button>
              )}
            </div>

            {/* Expanded Filters Panel */}
            {isFiltersExpanded && (
              <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg bg-gray-50 dark:bg-gray-800/50 space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  {/* Role Filter */}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Роль
                    </label>
                    <select
                      value={roleFilter}
                      onChange={(e) => {
                        setRoleFilter(e.target.value)
                        setPage(1)
                      }}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                               bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                               focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    >
                      <option value="all">Все роли</option>
                      {Object.entries(roleLabels).map(([role, label]) => (
                        <option key={role} value={role}>
                          {label}
                        </option>
                      ))}
                    </select>
                  </div>

                  {/* Status Filter */}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Статус
                    </label>
                    <select
                      value={statusFilter}
                      onChange={(e) => {
                        setStatusFilter(e.target.value)
                        setPage(1)
                      }}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                               bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                               focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    >
                      <option value="all">Все статусы</option>
                      {Object.entries(statusLabels).map(([status, label]) => (
                        <option key={status} value={status}>
                          {label}
                        </option>
                      ))}
                    </select>
                  </div>

                  {/* Department Filter */}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Подразделение
                    </label>
                    <select
                      value={departmentFilter}
                      onChange={(e) => {
                        setDepartmentFilter(e.target.value)
                        setPage(1)
                      }}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                               bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                               focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    >
                      <option value="all">Все подразделения</option>
                      {departments.map((dept) => (
                        <option key={dept.id} value={dept.id.toString()}>
                          {dept.name}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
              </div>
            )}

            {/* Active Filters Summary */}
            {hasActiveFilters && !isFiltersExpanded && (
              <div className="flex flex-wrap gap-2">
                {searchQuery && (
                  <span className="px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400 rounded-full text-sm flex items-center gap-2">
                    Поиск: {searchQuery}
                    <button
                      onClick={() => {
                        setSearchQuery('')
                        setPage(1)
                      }}
                      className="hover:bg-blue-200 dark:hover:bg-blue-800/50 rounded-full p-0.5"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </span>
                )}
                {roleFilter !== 'all' && (
                  <span className="px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400 rounded-full text-sm flex items-center gap-2">
                    {roleLabels[roleFilter]}
                    <button
                      onClick={() => {
                        setRoleFilter('all')
                        setPage(1)
                      }}
                      className="hover:bg-blue-200 dark:hover:bg-blue-800/50 rounded-full p-0.5"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </span>
                )}
                {statusFilter !== 'all' && (
                  <span className="px-3 py-1 bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400 rounded-full text-sm flex items-center gap-2">
                    {statusLabels[statusFilter]}
                    <button
                      onClick={() => {
                        setStatusFilter('all')
                        setPage(1)
                      }}
                      className="hover:bg-green-200 dark:hover:bg-green-800/50 rounded-full p-0.5"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </span>
                )}
                {departmentFilter !== 'all' && (
                  <span className="px-3 py-1 bg-purple-100 dark:bg-purple-900/30 text-purple-800 dark:text-purple-400 rounded-full text-sm flex items-center gap-2">
                    {departments.find((d) => d.id.toString() === departmentFilter)?.name ||
                      'Подразделение'}
                    <button
                      onClick={() => {
                        setDepartmentFilter('all')
                        setPage(1)
                      }}
                      className="hover:bg-purple-200 dark:hover:bg-purple-800/50 rounded-full p-0.5"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </span>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Users Table */}
        <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
          <GlowingEffect
            spread={40}
            glow={true}
            disabled={false}
            proximity={64}
            inactiveZone={0.01}
            borderWidth={3}
          />
          <div className="relative z-10">
            <div className="mb-4">
              <h3 className="text-lg sm:text-xl font-semibold">Пользователи</h3>
              <p className="text-sm text-muted-foreground">
                Список всех зарегистрированных пользователей
              </p>
            </div>
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
                          <Badge
                            variant={roleColors[user.role] || 'outline'}
                            className="whitespace-nowrap"
                          >
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
                        <TableHead className="min-w-[180px]">Роль</TableHead>
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
                              <Badge
                                variant={roleColors[user.role] || 'outline'}
                                className="whitespace-nowrap"
                              >
                                {roleLabels[user.role] || user.role}
                              </Badge>
                            </TableCell>
                            <TableCell>
                              <Badge
                                variant={
                                  user.status === 'active'
                                    ? 'outline'
                                    : user.status === 'blocked'
                                      ? 'destructive'
                                      : 'secondary'
                                }
                                className={
                                  user.status === 'active'
                                    ? 'border-green-500 text-green-500 bg-green-500/10'
                                    : ''
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
          </div>
        </div>
      </div>
    </AppLayout>
  )
}

// Admins, methodists, academic secretaries and teachers can access this page
export default withAuth(UsersManagementPage, {
  roles: [UserRole.SYSTEM_ADMIN, UserRole.METHODIST, UserRole.ACADEMIC_SECRETARY, UserRole.TEACHER],
})
