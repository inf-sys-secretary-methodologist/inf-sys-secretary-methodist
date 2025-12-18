'use client'

import { useState, useCallback } from 'react'
import { withAuth } from '@/components/auth/withAuth'
import { UserRole } from '@/types/auth'
import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  RefreshCw,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Clock,
  Users,
  GraduationCap,
  Activity,
  Loader2,
  ChevronLeft,
  ChevronRight,
  Eye,
} from 'lucide-react'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { toast } from 'sonner'
import {
  useSyncStats,
  useSyncLogs,
  usePendingConflicts,
  useExternalEmployees,
  useExternalStudents,
  startSync,
  resolveConflict,
  bulkResolveConflicts,
  type SyncEntityType,
  type SyncConflict,
  type ConflictResolution,
} from '@/hooks/useIntegration'

const entityTypeLabels: Record<SyncEntityType, string> = {
  employee: 'Сотрудники',
  student: 'Студенты',
}

const statusLabels: Record<string, string> = {
  pending: 'Ожидание',
  in_progress: 'В процессе',
  completed: 'Завершено',
  failed: 'Ошибка',
  cancelled: 'Отменено',
}

const statusColors: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  pending: 'secondary',
  in_progress: 'default',
  completed: 'outline',
  failed: 'destructive',
  cancelled: 'secondary',
}

function IntegrationPage() {
  const [activeTab, setActiveTab] = useState('overview')
  const [syncingEntity, setSyncingEntity] = useState<SyncEntityType | null>(null)
  const [selectedConflict, setSelectedConflict] = useState<SyncConflict | null>(null)
  const [showResolveDialog, setShowResolveDialog] = useState(false)
  const [selectedResolution, setSelectedResolution] = useState<ConflictResolution>('use_local')
  const [page, setPage] = useState(1)
  const limit = 10

  // Data hooks
  const { stats: syncStats, isLoading: statsLoading, mutate: mutateStats } = useSyncStats()
  const {
    logs,
    total: logsTotal,
    isLoading: logsLoading,
    mutate: mutateLogs,
  } = useSyncLogs({ limit, offset: (page - 1) * limit })
  const {
    conflicts,
    total: conflictsTotal,
    isLoading: conflictsLoading,
    mutate: mutateConflicts,
  } = usePendingConflicts(limit, (page - 1) * limit)
  const {
    employees,
    total: employeesTotal,
    isLoading: employeesLoading,
  } = useExternalEmployees({ limit, offset: (page - 1) * limit })
  const {
    students,
    total: studentsTotal,
    isLoading: studentsLoading,
  } = useExternalStudents({ limit, offset: (page - 1) * limit })

  const totalPages = Math.ceil(
    (activeTab === 'logs'
      ? logsTotal
      : activeTab === 'conflicts'
        ? conflictsTotal
        : activeTab === 'employees'
          ? employeesTotal
          : activeTab === 'students'
            ? studentsTotal
            : 0) / limit
  )

  const handleStartSync = useCallback(
    async (entityType: SyncEntityType) => {
      setSyncingEntity(entityType)
      try {
        const result = await startSync({ entity_type: entityType, force: false })
        toast.success(`Синхронизация ${entityTypeLabels[entityType].toLowerCase()} запущена`, {
          description: `ID: ${result.sync_log_id}`,
        })
        mutateStats()
        mutateLogs()
      } catch (error) {
        console.error('Sync failed:', error)
        toast.error('Не удалось запустить синхронизацию')
      } finally {
        setSyncingEntity(null)
      }
    },
    [mutateStats, mutateLogs]
  )

  const handleResolveConflict = useCallback(async () => {
    if (!selectedConflict) return

    try {
      await resolveConflict(selectedConflict.id, { resolution: selectedResolution })
      toast.success('Конфликт разрешен')
      mutateConflicts()
      setShowResolveDialog(false)
      setSelectedConflict(null)
    } catch (error) {
      console.error('Resolve failed:', error)
      toast.error('Не удалось разрешить конфликт')
    }
  }, [selectedConflict, selectedResolution, mutateConflicts])

  const handleBulkResolve = useCallback(
    async (resolution: ConflictResolution) => {
      const ids = conflicts.map((c) => c.id)
      if (ids.length === 0) return

      try {
        const result = await bulkResolveConflicts({ ids, resolution })
        toast.success(`Разрешено конфликтов: ${result.count}`)
        mutateConflicts()
      } catch (error) {
        console.error('Bulk resolve failed:', error)
        toast.error('Не удалось разрешить конфликты')
      }
    },
    [conflicts, mutateConflicts]
  )

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('ru-RU', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const refreshAll = () => {
    mutateStats()
    mutateLogs()
    mutateConflicts()
  }

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Page Header */}
        <div className="text-center space-y-2 sm:space-y-4">
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-gray-900 dark:text-white">
            Интеграция с 1С
          </h1>
          <p className="text-base sm:text-lg text-gray-600 dark:text-gray-300">
            Синхронизация данных сотрудников и студентов
          </p>
        </div>

        {/* Action Button */}
        <div className="flex justify-end">
          <Button variant="outline" size="sm" onClick={refreshAll}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Обновить
          </Button>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4">
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
              <p className="text-xs sm:text-sm text-muted-foreground flex items-center gap-1">
                <Activity className="h-4 w-4" />
                Всего синхронизаций
              </p>
              <p className="text-2xl sm:text-3xl font-bold">
                {statsLoading ? (
                  <Loader2 className="h-6 w-6 animate-spin" />
                ) : (
                  syncStats?.total_syncs || 0
                )}
              </p>
            </div>
          </div>
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
              <p className="text-xs sm:text-sm text-muted-foreground flex items-center gap-1">
                <CheckCircle2 className="h-4 w-4 text-green-500" />
                Успешных
              </p>
              <p className="text-2xl sm:text-3xl font-bold text-green-600">
                {statsLoading ? (
                  <Loader2 className="h-6 w-6 animate-spin" />
                ) : (
                  syncStats?.successful_syncs || 0
                )}
              </p>
            </div>
          </div>
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
              <p className="text-xs sm:text-sm text-muted-foreground flex items-center gap-1">
                <XCircle className="h-4 w-4 text-red-500" />
                Ошибок
              </p>
              <p className="text-2xl sm:text-3xl font-bold text-red-600">
                {statsLoading ? (
                  <Loader2 className="h-6 w-6 animate-spin" />
                ) : (
                  syncStats?.failed_syncs || 0
                )}
              </p>
            </div>
          </div>
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
              <p className="text-xs sm:text-sm text-muted-foreground flex items-center gap-1">
                <AlertTriangle className="h-4 w-4 text-yellow-500" />
                Конфликтов
              </p>
              <p className="text-2xl sm:text-3xl font-bold text-yellow-600">
                {statsLoading ? (
                  <Loader2 className="h-6 w-6 animate-spin" />
                ) : (
                  syncStats?.total_conflicts || 0
                )}
              </p>
            </div>
          </div>
        </div>

        {/* Sync Actions */}
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
              <h3 className="text-lg font-semibold">Запуск синхронизации</h3>
              <p className="text-sm text-muted-foreground">
                Выберите тип данных для синхронизации с 1С
              </p>
            </div>
            <div className="flex flex-wrap gap-4">
              <Button
                onClick={() => handleStartSync('employee')}
                disabled={syncingEntity !== null}
                className="flex items-center gap-2"
              >
                {syncingEntity === 'employee' ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Users className="h-4 w-4" />
                )}
                Синхронизировать сотрудников
              </Button>
              <Button
                onClick={() => handleStartSync('student')}
                disabled={syncingEntity !== null}
                variant="secondary"
                className="flex items-center gap-2"
              >
                {syncingEntity === 'student' ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <GraduationCap className="h-4 w-4" />
                )}
                Синхронизировать студентов
              </Button>
            </div>
            {syncStats?.last_sync_at && (
              <p className="text-sm text-muted-foreground mt-4">
                Последняя синхронизация: {formatDate(syncStats.last_sync_at)}
              </p>
            )}
          </div>
        </div>

        {/* Tabs */}
        <Tabs
          value={activeTab}
          onValueChange={(v) => {
            setActiveTab(v)
            setPage(1)
          }}
        >
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="logs" className="flex items-center gap-1">
              <Clock className="h-4 w-4" />
              <span className="hidden sm:inline">История</span>
            </TabsTrigger>
            <TabsTrigger value="conflicts" className="flex items-center gap-1">
              <AlertTriangle className="h-4 w-4" />
              <span className="hidden sm:inline">Конфликты</span>
              {conflictsTotal > 0 && (
                <Badge variant="destructive" className="ml-1">
                  {conflictsTotal}
                </Badge>
              )}
            </TabsTrigger>
            <TabsTrigger value="employees" className="flex items-center gap-1">
              <Users className="h-4 w-4" />
              <span className="hidden sm:inline">Сотрудники</span>
            </TabsTrigger>
            <TabsTrigger value="students" className="flex items-center gap-1">
              <GraduationCap className="h-4 w-4" />
              <span className="hidden sm:inline">Студенты</span>
            </TabsTrigger>
          </TabsList>

          {/* Logs Tab */}
          <TabsContent value="logs">
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
                  <h3 className="text-lg font-semibold">История синхронизаций</h3>
                  <p className="text-sm text-muted-foreground">
                    Журнал всех операций синхронизации
                  </p>
                </div>
                {logsLoading ? (
                  <div className="flex justify-center py-8">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                  </div>
                ) : (
                  <>
                    <div className="rounded-md border overflow-x-auto">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Тип</TableHead>
                            <TableHead>Статус</TableHead>
                            <TableHead>Записей</TableHead>
                            <TableHead>Успешно</TableHead>
                            <TableHead>Ошибок</TableHead>
                            <TableHead>Конфликтов</TableHead>
                            <TableHead>Дата</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {logs.length === 0 ? (
                            <TableRow>
                              <TableCell
                                colSpan={7}
                                className="text-center text-muted-foreground py-8"
                              >
                                Нет записей
                              </TableCell>
                            </TableRow>
                          ) : (
                            logs.map((log) => (
                              <TableRow key={log.id}>
                                <TableCell>
                                  <div className="flex items-center gap-2">
                                    {log.entity_type === 'employee' ? (
                                      <Users className="h-4 w-4" />
                                    ) : (
                                      <GraduationCap className="h-4 w-4" />
                                    )}
                                    {entityTypeLabels[log.entity_type]}
                                  </div>
                                </TableCell>
                                <TableCell>
                                  <Badge variant={statusColors[log.status]}>
                                    {statusLabels[log.status]}
                                  </Badge>
                                </TableCell>
                                <TableCell>{log.total_records}</TableCell>
                                <TableCell className="text-green-600">
                                  {log.success_count}
                                </TableCell>
                                <TableCell className="text-red-600">{log.error_count}</TableCell>
                                <TableCell className="text-yellow-600">
                                  {log.conflict_count}
                                </TableCell>
                                <TableCell className="text-sm text-muted-foreground">
                                  {formatDate(log.started_at)}
                                </TableCell>
                              </TableRow>
                            ))
                          )}
                        </TableBody>
                      </Table>
                    </div>
                    {totalPages > 1 && (
                      <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
                    )}
                  </>
                )}
              </div>
            </div>
          </TabsContent>

          {/* Conflicts Tab */}
          <TabsContent value="conflicts">
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
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <h3 className="text-lg font-semibold">Конфликты синхронизации</h3>
                    <p className="text-sm text-muted-foreground">Требуют ручного разрешения</p>
                  </div>
                  {conflicts.length > 0 && (
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleBulkResolve('use_local')}
                      >
                        Все локальные
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleBulkResolve('use_external')}
                      >
                        Все внешние
                      </Button>
                    </div>
                  )}
                </div>
                {conflictsLoading ? (
                  <div className="flex justify-center py-8">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                  </div>
                ) : conflicts.length === 0 ? (
                  <div className="text-center py-8 text-muted-foreground">
                    <CheckCircle2 className="h-12 w-12 mx-auto mb-2 text-green-500" />
                    <p>Нет нерешенных конфликтов</p>
                  </div>
                ) : (
                  <>
                    <div className="rounded-md border overflow-x-auto">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Тип</TableHead>
                            <TableHead>ID сущности</TableHead>
                            <TableHead>Тип конфликта</TableHead>
                            <TableHead>Поля</TableHead>
                            <TableHead>Дата</TableHead>
                            <TableHead className="text-right">Действия</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {conflicts.map((conflict) => (
                            <TableRow key={conflict.id}>
                              <TableCell>
                                <Badge variant="outline">
                                  {entityTypeLabels[conflict.entity_type]}
                                </Badge>
                              </TableCell>
                              <TableCell className="font-mono text-sm">
                                {conflict.entity_id}
                              </TableCell>
                              <TableCell>{conflict.conflict_type}</TableCell>
                              <TableCell>
                                <div className="flex flex-wrap gap-1">
                                  {conflict.conflict_fields.slice(0, 3).map((field) => (
                                    <Badge key={field} variant="secondary" className="text-xs">
                                      {field}
                                    </Badge>
                                  ))}
                                  {conflict.conflict_fields.length > 3 && (
                                    <Badge variant="secondary" className="text-xs">
                                      +{conflict.conflict_fields.length - 3}
                                    </Badge>
                                  )}
                                </div>
                              </TableCell>
                              <TableCell className="text-sm text-muted-foreground">
                                {formatDate(conflict.created_at)}
                              </TableCell>
                              <TableCell className="text-right">
                                <Button
                                  size="sm"
                                  variant="outline"
                                  onClick={() => {
                                    setSelectedConflict(conflict)
                                    setShowResolveDialog(true)
                                  }}
                                >
                                  <Eye className="h-4 w-4 mr-1" />
                                  Разрешить
                                </Button>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                    {totalPages > 1 && (
                      <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
                    )}
                  </>
                )}
              </div>
            </div>
          </TabsContent>

          {/* Employees Tab */}
          <TabsContent value="employees">
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
                  <h3 className="text-lg font-semibold">Сотрудники из 1С</h3>
                  <p className="text-sm text-muted-foreground">
                    Синхронизированные данные сотрудников
                  </p>
                </div>
                {employeesLoading ? (
                  <div className="flex justify-center py-8">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                  </div>
                ) : (
                  <>
                    <div className="rounded-md border overflow-x-auto">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>ФИО</TableHead>
                            <TableHead>Email</TableHead>
                            <TableHead>Должность</TableHead>
                            <TableHead>Подразделение</TableHead>
                            <TableHead>Статус</TableHead>
                            <TableHead>Синхронизирован</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {employees.length === 0 ? (
                            <TableRow>
                              <TableCell
                                colSpan={6}
                                className="text-center text-muted-foreground py-8"
                              >
                                Нет данных
                              </TableCell>
                            </TableRow>
                          ) : (
                            employees.map((emp) => (
                              <TableRow key={emp.id}>
                                <TableCell className="font-medium">
                                  {emp.last_name} {emp.first_name} {emp.middle_name || ''}
                                </TableCell>
                                <TableCell className="text-sm">{emp.email || '-'}</TableCell>
                                <TableCell className="text-sm">{emp.position || '-'}</TableCell>
                                <TableCell className="text-sm">{emp.department || '-'}</TableCell>
                                <TableCell>
                                  <Badge variant={emp.is_active ? 'default' : 'secondary'}>
                                    {emp.is_active ? 'Активен' : 'Неактивен'}
                                  </Badge>
                                </TableCell>
                                <TableCell className="text-sm text-muted-foreground">
                                  {formatDate(emp.last_sync_at)}
                                </TableCell>
                              </TableRow>
                            ))
                          )}
                        </TableBody>
                      </Table>
                    </div>
                    {totalPages > 1 && (
                      <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
                    )}
                  </>
                )}
              </div>
            </div>
          </TabsContent>

          {/* Students Tab */}
          <TabsContent value="students">
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
                  <h3 className="text-lg font-semibold">Студенты из 1С</h3>
                  <p className="text-sm text-muted-foreground">
                    Синхронизированные данные студентов
                  </p>
                </div>
                {studentsLoading ? (
                  <div className="flex justify-center py-8">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                  </div>
                ) : (
                  <>
                    <div className="rounded-md border overflow-x-auto">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>ФИО</TableHead>
                            <TableHead>Группа</TableHead>
                            <TableHead>Факультет</TableHead>
                            <TableHead>Курс</TableHead>
                            <TableHead>Статус</TableHead>
                            <TableHead>Синхронизирован</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {students.length === 0 ? (
                            <TableRow>
                              <TableCell
                                colSpan={6}
                                className="text-center text-muted-foreground py-8"
                              >
                                Нет данных
                              </TableCell>
                            </TableRow>
                          ) : (
                            students.map((student) => (
                              <TableRow key={student.id}>
                                <TableCell className="font-medium">
                                  {student.last_name} {student.first_name}{' '}
                                  {student.middle_name || ''}
                                </TableCell>
                                <TableCell className="text-sm">
                                  {student.group_name || '-'}
                                </TableCell>
                                <TableCell className="text-sm">{student.faculty || '-'}</TableCell>
                                <TableCell className="text-sm">{student.course || '-'}</TableCell>
                                <TableCell>
                                  <Badge variant={student.is_active ? 'default' : 'secondary'}>
                                    {student.is_active ? 'Активен' : 'Неактивен'}
                                  </Badge>
                                </TableCell>
                                <TableCell className="text-sm text-muted-foreground">
                                  {formatDate(student.last_sync_at)}
                                </TableCell>
                              </TableRow>
                            ))
                          )}
                        </TableBody>
                      </Table>
                    </div>
                    {totalPages > 1 && (
                      <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
                    )}
                  </>
                )}
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </div>

      {/* Resolve Conflict Dialog */}
      <AlertDialog open={showResolveDialog} onOpenChange={setShowResolveDialog}>
        <AlertDialogContent className="max-w-2xl">
          <AlertDialogHeader>
            <AlertDialogTitle>Разрешение конфликта</AlertDialogTitle>
            <AlertDialogDescription>
              Выберите, какие данные использовать для сущности {selectedConflict?.entity_id}
            </AlertDialogDescription>
          </AlertDialogHeader>

          {selectedConflict && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <h4 className="font-medium text-sm">Локальные данные</h4>
                  <pre className="text-xs bg-muted p-2 rounded overflow-auto max-h-40">
                    {JSON.stringify(JSON.parse(selectedConflict.local_data || '{}'), null, 2)}
                  </pre>
                </div>
                <div className="space-y-2">
                  <h4 className="font-medium text-sm">Внешние данные (1С)</h4>
                  <pre className="text-xs bg-muted p-2 rounded overflow-auto max-h-40">
                    {JSON.stringify(JSON.parse(selectedConflict.external_data || '{}'), null, 2)}
                  </pre>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Выберите решение:</label>
                <Select
                  value={selectedResolution}
                  onValueChange={(v) => setSelectedResolution(v as ConflictResolution)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="use_local">Использовать локальные данные</SelectItem>
                    <SelectItem value="use_external">Использовать данные из 1С</SelectItem>
                    <SelectItem value="skip">Пропустить</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}

          <AlertDialogFooter>
            <AlertDialogCancel>Отмена</AlertDialogCancel>
            <AlertDialogAction onClick={handleResolveConflict}>Применить</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </AppLayout>
  )
}

// Pagination component
function Pagination({
  page,
  totalPages,
  onPageChange,
}: {
  page: number
  totalPages: number
  onPageChange: (page: number) => void
}) {
  return (
    <div className="flex items-center justify-between mt-4">
      <p className="text-sm text-muted-foreground">
        Страница {page} из {totalPages}
      </p>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(Math.max(1, page - 1))}
          disabled={page === 1}
        >
          <ChevronLeft className="h-4 w-4" />
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(Math.min(totalPages, page + 1))}
          disabled={page === totalPages}
        >
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}

// Protect page - only admin and methodist can access
export default withAuth(IntegrationPage, {
  roles: [UserRole.SYSTEM_ADMIN, UserRole.METHODIST],
})
