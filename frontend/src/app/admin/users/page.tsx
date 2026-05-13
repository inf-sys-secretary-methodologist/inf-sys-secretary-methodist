'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { ChevronLeft, ChevronRight, Loader2, Pencil, Trash2, Users } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useAuthCheck } from '@/hooks/useAuth'
import { useUsers } from '@/hooks/useUsers'
import { useDeleteUser, useUpdateUserRole, useUpdateUserStatus } from '@/hooks/useUserMutations'
import type { User, UserListFilter, UserRole, UserStatus } from '@/types/user'

const PAGE_SIZE = 20

const ROLE_VALUES: UserRole[] = [
  'system_admin',
  'methodist',
  'academic_secretary',
  'teacher',
  'student',
]

const STATUS_VALUES: UserStatus[] = ['active', 'inactive', 'blocked']

// AdminUsersPage — admin-only user management list view. Mirror к
// /admin/audit-logs filter/list/pagination shape but with role/status
// columns + per-row action affordances (dialogs land in Pair 5).
export default function AdminUsersPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('adminUsers')

  const [search, setSearch] = useState('')
  const [role, setRole] = useState<UserRole | ''>('')
  const [status, setStatus] = useState<UserStatus | ''>('')
  const [page, setPage] = useState(1)

  const filter = useMemo<UserListFilter>(
    () => ({
      search: search.trim() || undefined,
      role: role || undefined,
      status: status || undefined,
      page,
      limit: PAGE_SIZE,
    }),
    [search, role, status, page]
  )

  const enabled = !isLoading && isAuthenticated && user?.role === 'system_admin'
  const {
    users,
    page: currentPage,
    totalPages,
    isLoading: listLoading,
    error,
    mutate,
  } = useUsers(filter, { enabled })

  // Single dialog slot — only one dialog is open at a time. The
  // payload carries the row being edited so per-dialog state
  // (selected role / status) lives in the dialog component, not
  // here.
  type DialogState =
    | { kind: 'role'; user: User }
    | { kind: 'status'; user: User }
    | { kind: 'delete'; user: User }
    | null
  const [dialog, setDialog] = useState<DialogState>(null)
  const closeDialog = () => setDialog(null)
  const onMutationSuccess = () => {
    closeDialog()
    mutate()
  }

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  const handleReset = () => {
    setSearch('')
    setRole('')
    setStatus('')
    setPage(1)
  }

  return (
    <AppLayout>
      <div data-testid="admin-users-page" className="max-w-7xl mx-auto space-y-6">
        <header className="flex items-center gap-3">
          <Users className="h-7 w-7" />
          <div className="flex-1">
            <h1 className="text-2xl font-bold">{t('title')}</h1>
            <p className="text-sm text-muted-foreground">{t('description')}</p>
          </div>
        </header>

        <section
          aria-label={t('filters.search')}
          className="rounded-xl border border-border bg-card p-4 grid grid-cols-1 md:grid-cols-4 gap-3"
        >
          <div className="md:col-span-2">
            <Label htmlFor="users-search">{t('filters.search')}</Label>
            <Input
              id="users-search"
              data-testid="users-search"
              value={search}
              onChange={(e) => {
                setSearch(e.target.value)
                setPage(1)
              }}
            />
          </div>
          <div>
            <Label htmlFor="users-role-filter">{t('filters.role')}</Label>
            <select
              id="users-role-filter"
              data-testid="users-role-filter"
              value={role}
              onChange={(e) => {
                setRole(e.target.value as UserRole | '')
                setPage(1)
              }}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            >
              <option value="">{t('filters.allRoles')}</option>
              {ROLE_VALUES.map((r) => (
                <option key={r} value={r}>
                  {t(`roleOptions.${r}`)}
                </option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="users-status-filter">{t('filters.status')}</Label>
            <select
              id="users-status-filter"
              data-testid="users-status-filter"
              value={status}
              onChange={(e) => {
                setStatus(e.target.value as UserStatus | '')
                setPage(1)
              }}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            >
              <option value="">{t('filters.allStatuses')}</option>
              {STATUS_VALUES.map((s) => (
                <option key={s} value={s}>
                  {t(`statusOptions.${s}`)}
                </option>
              ))}
            </select>
          </div>
          <div className="md:col-span-4 flex justify-end">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              data-testid="users-reset"
              onClick={handleReset}
            >
              {t('filters.reset')}
            </Button>
          </div>
        </section>

        {listLoading ? (
          <div data-testid="users-loading" className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div
            data-testid="users-error"
            className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center"
          >
            <p className="font-medium text-destructive">{t('loadFailed')}</p>
          </div>
        ) : users.length === 0 ? (
          <div
            data-testid="users-empty"
            className="flex flex-col items-center justify-center py-16 text-center"
          >
            <Users className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-medium">{t('empty.title')}</h3>
            <p className="text-muted-foreground">{t('empty.description')}</p>
          </div>
        ) : (
          <>
            <div
              data-testid="users-table"
              className="rounded-xl border border-border bg-card overflow-x-auto"
            >
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('columns.user')}</TableHead>
                    <TableHead>{t('columns.role')}</TableHead>
                    <TableHead>{t('columns.status')}</TableHead>
                    <TableHead>{t('columns.department')}</TableHead>
                    <TableHead>{t('columns.position')}</TableHead>
                    <TableHead className="text-right">{t('columns.actions')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {users.map((u) => (
                    <UserRowView
                      key={u.id}
                      user={u}
                      onChangeRole={() => setDialog({ kind: 'role', user: u })}
                      onChangeStatus={() => setDialog({ kind: 'status', user: u })}
                      onDelete={() => setDialog({ kind: 'delete', user: u })}
                    />
                  ))}
                </TableBody>
              </Table>
            </div>

            {totalPages > 1 && (
              <div className="flex items-center justify-between">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  data-testid="users-pagination-prev"
                  disabled={currentPage <= 1}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                >
                  <ChevronLeft className="h-4 w-4 mr-1" />
                  {t('pagination.prev')}
                </Button>
                <span
                  data-testid="users-pagination-indicator"
                  className="text-sm text-muted-foreground"
                >
                  {t('pagination.pageOf', { page: currentPage, totalPages })}
                </span>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  data-testid="users-pagination-next"
                  disabled={currentPage >= totalPages}
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                >
                  {t('pagination.next')}
                  <ChevronRight className="h-4 w-4 ml-1" />
                </Button>
              </div>
            )}
          </>
        )}

        {dialog?.kind === 'role' && (
          <ChangeRoleDialog
            user={dialog.user}
            onCancel={closeDialog}
            onSuccess={onMutationSuccess}
          />
        )}
        {dialog?.kind === 'status' && (
          <ChangeStatusDialog
            user={dialog.user}
            onCancel={closeDialog}
            onSuccess={onMutationSuccess}
          />
        )}
        {dialog?.kind === 'delete' && (
          <DeleteUserDialog
            user={dialog.user}
            onCancel={closeDialog}
            onSuccess={onMutationSuccess}
          />
        )}
      </div>
    </AppLayout>
  )
}

function UserRowView({
  user,
  onChangeRole,
  onChangeStatus,
  onDelete,
}: {
  user: User
  onChangeRole: () => void
  onChangeStatus: () => void
  onDelete: () => void
}) {
  const t = useTranslations('adminUsers')
  return (
    <TableRow data-testid={`user-row-${user.id}`}>
      <TableCell>
        <div className="font-medium">{user.name}</div>
        <div className="text-xs text-muted-foreground">{user.email}</div>
      </TableCell>
      <TableCell>
        <span
          data-testid={`user-role-${user.id}`}
          className="inline-flex items-center rounded-full bg-muted px-2 py-0.5 text-xs font-medium"
        >
          {t(`roleOptions.${user.role}`)}
        </span>
      </TableCell>
      <TableCell>
        <span
          data-testid={`user-status-${user.id}`}
          className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
            user.status === 'active'
              ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
              : user.status === 'blocked'
                ? 'bg-destructive/15 text-destructive'
                : 'bg-muted text-muted-foreground'
          }`}
        >
          {t(`statusOptions.${user.status}`)}
        </span>
      </TableCell>
      <TableCell className="text-sm">{user.department_name ?? '—'}</TableCell>
      <TableCell className="text-sm">{user.position_name ?? '—'}</TableCell>
      <TableCell className="text-right">
        <div className="inline-flex items-center gap-1">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            data-testid={`change-role-button-${user.id}`}
            onClick={onChangeRole}
            title={t('actions.changeRole')}
          >
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            data-testid={`change-status-button-${user.id}`}
            onClick={onChangeStatus}
            title={t('actions.changeStatus')}
          >
            <span className="text-xs">{t('actions.changeStatus')}</span>
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            data-testid={`delete-button-${user.id}`}
            onClick={onDelete}
            title={t('actions.delete')}
          >
            <Trash2 className="h-4 w-4 text-destructive" />
          </Button>
        </div>
      </TableCell>
    </TableRow>
  )
}

const ROLE_VALUES_DIALOG: UserRole[] = [
  'system_admin',
  'methodist',
  'academic_secretary',
  'teacher',
  'student',
]

const STATUS_VALUES_DIALOG: UserStatus[] = ['active', 'inactive', 'blocked']

function ChangeRoleDialog({
  user,
  onCancel,
  onSuccess,
}: {
  user: User
  onCancel: () => void
  onSuccess: () => void
}) {
  const t = useTranslations('adminUsers')
  const [selected, setSelected] = useState<UserRole>(user.role)
  const { updateRole, isLoading } = useUpdateUserRole()

  const handleConfirm = async () => {
    try {
      await updateRole(user.id, selected)
      onSuccess()
    } catch {
      // Surface stays in hook.error; dialog stays open for retry.
    }
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onCancel()}>
      <DialogContent data-testid="change-role-dialog">
        <DialogHeader>
          <DialogTitle>{t('dialogs.changeRole.title')}</DialogTitle>
          <DialogDescription>
            {t('dialogs.changeRole.description', { name: user.name })}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-2">
          <Label htmlFor="change-role-select">{t('dialogs.changeRole.new')}</Label>
          <select
            id="change-role-select"
            data-testid="change-role-select"
            value={selected}
            onChange={(e) => setSelected(e.target.value as UserRole)}
            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
          >
            {ROLE_VALUES_DIALOG.map((r) => (
              <option key={r} value={r}>
                {t(`roleOptions.${r}`)}
              </option>
            ))}
          </select>
        </div>
        <DialogFooter>
          <Button
            type="button"
            variant="ghost"
            data-testid="change-role-cancel"
            onClick={onCancel}
            disabled={isLoading}
          >
            {t('dialogs.changeRole.cancel')}
          </Button>
          <Button
            type="button"
            data-testid="change-role-confirm"
            onClick={handleConfirm}
            disabled={isLoading}
          >
            {t('dialogs.changeRole.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function ChangeStatusDialog({
  user,
  onCancel,
  onSuccess,
}: {
  user: User
  onCancel: () => void
  onSuccess: () => void
}) {
  const t = useTranslations('adminUsers')
  const [selected, setSelected] = useState<UserStatus>(user.status)
  const { updateStatus, isLoading } = useUpdateUserStatus()

  const handleConfirm = async () => {
    try {
      await updateStatus(user.id, selected)
      onSuccess()
    } catch {
      // Stay open on error.
    }
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onCancel()}>
      <DialogContent data-testid="change-status-dialog">
        <DialogHeader>
          <DialogTitle>{t('dialogs.changeStatus.title')}</DialogTitle>
          <DialogDescription>
            {t('dialogs.changeStatus.description', { name: user.name })}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-2">
          <Label htmlFor="change-status-select">{t('dialogs.changeStatus.new')}</Label>
          <select
            id="change-status-select"
            data-testid="change-status-select"
            value={selected}
            onChange={(e) => setSelected(e.target.value as UserStatus)}
            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
          >
            {STATUS_VALUES_DIALOG.map((s) => (
              <option key={s} value={s}>
                {t(`statusOptions.${s}`)}
              </option>
            ))}
          </select>
        </div>
        <DialogFooter>
          <Button
            type="button"
            variant="ghost"
            data-testid="change-status-cancel"
            onClick={onCancel}
            disabled={isLoading}
          >
            {t('dialogs.changeStatus.cancel')}
          </Button>
          <Button
            type="button"
            data-testid="change-status-confirm"
            onClick={handleConfirm}
            disabled={isLoading}
          >
            {t('dialogs.changeStatus.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function DeleteUserDialog({
  user,
  onCancel,
  onSuccess,
}: {
  user: User
  onCancel: () => void
  onSuccess: () => void
}) {
  const t = useTranslations('adminUsers')
  const { deleteUser, isLoading } = useDeleteUser()

  const handleConfirm = async () => {
    try {
      await deleteUser(user.id)
      onSuccess()
    } catch {
      // Stay open on error.
    }
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onCancel()}>
      <DialogContent data-testid="delete-dialog">
        <DialogHeader>
          <DialogTitle>{t('dialogs.delete.title')}</DialogTitle>
          <DialogDescription>
            {t('dialogs.delete.description', { name: user.name })}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button
            type="button"
            variant="ghost"
            data-testid="delete-cancel"
            onClick={onCancel}
            disabled={isLoading}
          >
            {t('dialogs.delete.cancel')}
          </Button>
          <Button
            type="button"
            variant="destructive"
            data-testid="delete-confirm"
            onClick={handleConfirm}
            disabled={isLoading}
          >
            {t('dialogs.delete.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
