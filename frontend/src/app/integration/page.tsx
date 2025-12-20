'use client'

import { useState, useCallback } from 'react'
import { useTranslations } from 'next-intl'
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

const STATUS_COLORS: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  pending: 'secondary',
  in_progress: 'default',
  completed: 'outline',
  failed: 'destructive',
  cancelled: 'secondary',
}

function IntegrationPage() {
  const t = useTranslations('integration')
  const tCommon = useTranslations('common')
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
        toast.success(t('sync.started', { type: t(`entityTypes.${entityType}`) }), {
          description: `ID: ${result.sync_log_id}`,
        })
        mutateStats()
        mutateLogs()
      } catch (error) {
        console.error('Sync failed:', error)
        toast.error(t('sync.error'))
      } finally {
        setSyncingEntity(null)
      }
    },
    [mutateStats, mutateLogs, t]
  )

  const handleResolveConflict = useCallback(async () => {
    if (!selectedConflict) return

    try {
      await resolveConflict(selectedConflict.id, { resolution: selectedResolution })
      toast.success(t('conflicts.resolved'))
      mutateConflicts()
      setShowResolveDialog(false)
      setSelectedConflict(null)
    } catch (error) {
      console.error('Resolve failed:', error)
      toast.error(t('conflicts.resolveError'))
    }
  }, [selectedConflict, selectedResolution, mutateConflicts, t])

  const handleBulkResolve = useCallback(
    async (resolution: ConflictResolution) => {
      const ids = conflicts.map((c) => c.id)
      if (ids.length === 0) return

      try {
        const result = await bulkResolveConflicts({ ids, resolution })
        toast.success(t('conflicts.resolvedCount', { count: result.count }))
        mutateConflicts()
      } catch (error) {
        console.error('Bulk resolve failed:', error)
        toast.error(t('conflicts.bulkResolveError'))
      }
    },
    [conflicts, mutateConflicts, t]
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
            {t('title')}
          </h1>
          <p className="text-base sm:text-lg text-gray-600 dark:text-gray-300">{t('subtitle')}</p>
        </div>

        {/* Action Button */}
        <div className="flex justify-end">
          <Button variant="outline" size="sm" onClick={refreshAll}>
            <RefreshCw className="h-4 w-4 mr-2" />
            {t('refresh')}
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
                {t('stats.totalSyncs')}
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
                {t('stats.successful')}
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
                {t('stats.failed')}
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
                {t('stats.conflicts')}
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
              <h3 className="text-lg font-semibold">{t('sync.title')}</h3>
              <p className="text-sm text-muted-foreground">{t('sync.subtitle')}</p>
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
                {t('sync.employees')}
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
                {t('sync.students')}
              </Button>
            </div>
            {syncStats?.last_sync_at && (
              <p className="text-sm text-muted-foreground mt-4">
                {t('sync.lastSync')}: {formatDate(syncStats.last_sync_at)}
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
              <span className="hidden sm:inline">{t('tabs.logs')}</span>
            </TabsTrigger>
            <TabsTrigger value="conflicts" className="flex items-center gap-1">
              <AlertTriangle className="h-4 w-4" />
              <span className="hidden sm:inline">{t('tabs.conflicts')}</span>
              {conflictsTotal > 0 && (
                <Badge variant="destructive" className="ml-1">
                  {conflictsTotal}
                </Badge>
              )}
            </TabsTrigger>
            <TabsTrigger value="employees" className="flex items-center gap-1">
              <Users className="h-4 w-4" />
              <span className="hidden sm:inline">{t('tabs.employees')}</span>
            </TabsTrigger>
            <TabsTrigger value="students" className="flex items-center gap-1">
              <GraduationCap className="h-4 w-4" />
              <span className="hidden sm:inline">{t('tabs.students')}</span>
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
                  <h3 className="text-lg font-semibold">{t('logs.title')}</h3>
                  <p className="text-sm text-muted-foreground">{t('logs.subtitle')}</p>
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
                            <TableHead>{t('logs.columns.type')}</TableHead>
                            <TableHead>{t('logs.columns.status')}</TableHead>
                            <TableHead>{t('logs.columns.records')}</TableHead>
                            <TableHead>{t('logs.columns.success')}</TableHead>
                            <TableHead>{t('logs.columns.errors')}</TableHead>
                            <TableHead>{t('logs.columns.conflicts')}</TableHead>
                            <TableHead>{t('logs.columns.date')}</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {logs.length === 0 ? (
                            <TableRow>
                              <TableCell
                                colSpan={7}
                                className="text-center text-muted-foreground py-8"
                              >
                                {t('logs.noRecords')}
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
                                    {t(`entityTypes.${log.entity_type}`)}
                                  </div>
                                </TableCell>
                                <TableCell>
                                  <Badge variant={STATUS_COLORS[log.status]}>
                                    {t(`statuses.${log.status}`)}
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
                      <PaginationComponent
                        page={page}
                        totalPages={totalPages}
                        onPageChange={setPage}
                        t={t}
                      />
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
                    <h3 className="text-lg font-semibold">{t('conflicts.title')}</h3>
                    <p className="text-sm text-muted-foreground">{t('conflicts.subtitle')}</p>
                  </div>
                  {conflicts.length > 0 && (
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleBulkResolve('use_local')}
                      >
                        {t('conflicts.allLocal')}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleBulkResolve('use_external')}
                      >
                        {t('conflicts.allExternal')}
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
                    <p>{t('conflicts.noConflicts')}</p>
                  </div>
                ) : (
                  <>
                    <div className="rounded-md border overflow-x-auto">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>{t('conflicts.columns.type')}</TableHead>
                            <TableHead>{t('conflicts.columns.entityId')}</TableHead>
                            <TableHead>{t('conflicts.columns.conflictType')}</TableHead>
                            <TableHead>{t('conflicts.columns.fields')}</TableHead>
                            <TableHead>{t('conflicts.columns.date')}</TableHead>
                            <TableHead className="text-right">
                              {t('conflicts.columns.actions')}
                            </TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {conflicts.map((conflict) => (
                            <TableRow key={conflict.id}>
                              <TableCell>
                                <Badge variant="outline">
                                  {t(`entityTypes.${conflict.entity_type}`)}
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
                                  {t('conflicts.resolve')}
                                </Button>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                    {totalPages > 1 && (
                      <PaginationComponent
                        page={page}
                        totalPages={totalPages}
                        onPageChange={setPage}
                        t={t}
                      />
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
                  <h3 className="text-lg font-semibold">{t('employees.title')}</h3>
                  <p className="text-sm text-muted-foreground">{t('employees.subtitle')}</p>
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
                            <TableHead>{t('employees.columns.fullName')}</TableHead>
                            <TableHead>{t('employees.columns.email')}</TableHead>
                            <TableHead>{t('employees.columns.position')}</TableHead>
                            <TableHead>{t('employees.columns.department')}</TableHead>
                            <TableHead>{t('employees.columns.status')}</TableHead>
                            <TableHead>{t('employees.columns.synced')}</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {employees.length === 0 ? (
                            <TableRow>
                              <TableCell
                                colSpan={6}
                                className="text-center text-muted-foreground py-8"
                              >
                                {t('employees.noData')}
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
                                    {emp.is_active ? t('statuses.active') : t('statuses.inactive')}
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
                      <PaginationComponent
                        page={page}
                        totalPages={totalPages}
                        onPageChange={setPage}
                        t={t}
                      />
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
                  <h3 className="text-lg font-semibold">{t('students.title')}</h3>
                  <p className="text-sm text-muted-foreground">{t('students.subtitle')}</p>
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
                            <TableHead>{t('students.columns.fullName')}</TableHead>
                            <TableHead>{t('students.columns.group')}</TableHead>
                            <TableHead>{t('students.columns.faculty')}</TableHead>
                            <TableHead>{t('students.columns.course')}</TableHead>
                            <TableHead>{t('students.columns.status')}</TableHead>
                            <TableHead>{t('students.columns.synced')}</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {students.length === 0 ? (
                            <TableRow>
                              <TableCell
                                colSpan={6}
                                className="text-center text-muted-foreground py-8"
                              >
                                {t('students.noData')}
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
                                    {student.is_active
                                      ? t('statuses.active')
                                      : t('statuses.inactive')}
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
                      <PaginationComponent
                        page={page}
                        totalPages={totalPages}
                        onPageChange={setPage}
                        t={t}
                      />
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
            <AlertDialogTitle>{t('conflicts.dialog.title')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t('conflicts.dialog.subtitle', { id: selectedConflict?.entity_id || '' })}
            </AlertDialogDescription>
          </AlertDialogHeader>

          {selectedConflict && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <h4 className="font-medium text-sm">{t('conflicts.dialog.localData')}</h4>
                  <pre className="text-xs bg-muted p-2 rounded overflow-auto max-h-40">
                    {JSON.stringify(JSON.parse(selectedConflict.local_data || '{}'), null, 2)}
                  </pre>
                </div>
                <div className="space-y-2">
                  <h4 className="font-medium text-sm">{t('conflicts.dialog.externalData')}</h4>
                  <pre className="text-xs bg-muted p-2 rounded overflow-auto max-h-40">
                    {JSON.stringify(JSON.parse(selectedConflict.external_data || '{}'), null, 2)}
                  </pre>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">
                  {t('conflicts.dialog.selectResolution')}
                </label>
                <Select
                  value={selectedResolution}
                  onValueChange={(v) => setSelectedResolution(v as ConflictResolution)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="use_local">{t('conflicts.dialog.useLocal')}</SelectItem>
                    <SelectItem value="use_external">
                      {t('conflicts.dialog.useExternal')}
                    </SelectItem>
                    <SelectItem value="skip">{t('conflicts.dialog.skip')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}

          <AlertDialogFooter>
            <AlertDialogCancel>{tCommon('cancel')}</AlertDialogCancel>
            <AlertDialogAction onClick={handleResolveConflict}>
              {t('conflicts.dialog.apply')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </AppLayout>
  )
}

// Pagination component
function PaginationComponent({
  page,
  totalPages,
  onPageChange,
  t,
}: {
  page: number
  totalPages: number
  onPageChange: (page: number) => void
  t: ReturnType<typeof useTranslations>
}) {
  return (
    <div className="flex items-center justify-between mt-4">
      <p className="text-sm text-muted-foreground">
        {t('pagination.page', { current: page, total: totalPages })}
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
