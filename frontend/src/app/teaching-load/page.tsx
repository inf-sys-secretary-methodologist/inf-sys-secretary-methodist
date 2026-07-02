'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { Loader2, Plus, Pencil, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { AppLayout } from '@/components/layout'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { ConfirmDeleteDialog } from '@/components/files/ConfirmDeleteDialog'
import { TeachingLoadForm } from '@/components/schedule/TeachingLoadForm'
import { useAuthCheck } from '@/hooks/useAuth'
import { useAuthStore } from '@/stores/authStore'
import { useSemesters } from '@/hooks/useSchedule'
import { canManageTeachingLoad } from '@/lib/auth/permissions'
import {
  useTeachingLoads,
  createTeachingLoad,
  updateTeachingLoad,
  deleteTeachingLoad,
} from '@/hooks/useTeachingLoad'
import type { TeachingLoad, TeachingLoadInput } from '@/types/teachingLoad'

const ALL_SEMESTERS = '__all__'

export default function TeachingLoadPage() {
  const t = useTranslations('teachingLoad')
  useAuthCheck()
  const user = useAuthStore((s) => s.user)
  const canManage = canManageTeachingLoad(user?.role)

  const { semesters } = useSemesters()
  const [semesterFilter, setSemesterFilter] = useState<number | undefined>(undefined)
  const { items, isLoading, error, mutate } = useTeachingLoads(
    semesterFilter ? { semester_id: semesterFilter } : undefined,
    { enabled: !!user }
  )

  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<TeachingLoad | undefined>(undefined)
  const [deletingId, setDeletingId] = useState<number | null>(null)

  const openCreate = () => {
    setEditing(undefined)
    setFormOpen(true)
  }
  const openEdit = (load: TeachingLoad) => {
    setEditing(load)
    setFormOpen(true)
  }

  const handleSubmit = async (input: TeachingLoadInput) => {
    try {
      if (editing) {
        await updateTeachingLoad(editing.id, input)
        toast.success(t('updated'))
      } else {
        await createTeachingLoad(input)
        toast.success(t('created'))
      }
      setFormOpen(false)
      await mutate()
    } catch {
      toast.error(t('errors.saveFailed'))
    }
  }

  const handleDelete = async () => {
    if (deletingId === null) return
    try {
      await deleteTeachingLoad(deletingId)
      toast.success(t('deleted'))
      await mutate()
    } catch {
      toast.error(t('errors.deleteFailed'))
    } finally {
      setDeletingId(null)
    }
  }

  return (
    <AppLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">{t('title')}</h1>
            <p className="text-sm text-muted-foreground">{t('description')}</p>
          </div>
          {canManage && (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4 mr-2" />
              {t('create')}
            </Button>
          )}
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t('filters.title')}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="max-w-xs space-y-2">
              <Label>{t('filters.semester')}</Label>
              <Select
                value={semesterFilter ? String(semesterFilter) : ALL_SEMESTERS}
                onValueChange={(v) =>
                  setSemesterFilter(v === ALL_SEMESTERS ? undefined : Number(v))
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={ALL_SEMESTERS}>{t('filters.allSemesters')}</SelectItem>
                  {semesters.map((s) => (
                    <SelectItem key={s.id} value={String(s.id)}>
                      {s.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : error ? (
              <p className="text-sm text-destructive">{t('errors.loadFailed')}</p>
            ) : items.length === 0 ? (
              <p className="text-sm text-muted-foreground py-8 text-center">{t('empty')}</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm min-w-[48rem]">
                  <thead>
                    <tr className="border-b text-left text-muted-foreground">
                      <th className="py-2 pr-4">{t('columns.group')}</th>
                      <th className="py-2 pr-4">{t('columns.discipline')}</th>
                      <th className="py-2 pr-4">{t('columns.teacher')}</th>
                      <th className="py-2 pr-4">{t('columns.lessonType')}</th>
                      <th className="py-2 pr-4">{t('columns.pairsPerWeek')}</th>
                      <th className="py-2 pr-4">{t('columns.weekType')}</th>
                      {canManage && <th className="py-2" />}
                    </tr>
                  </thead>
                  <tbody>
                    {items.map((load) => (
                      <tr key={load.id} className="border-b last:border-0">
                        <td className="py-2 pr-4">{load.group_name}</td>
                        <td className="py-2 pr-4">{load.discipline_name}</td>
                        <td className="py-2 pr-4">{load.teacher_name}</td>
                        <td className="py-2 pr-4">{load.lesson_type_name}</td>
                        <td className="py-2 pr-4">{load.pairs_per_week}</td>
                        <td className="py-2 pr-4">{t(`weekType.${load.week_type}`)}</td>
                        {canManage && (
                          <td className="py-2">
                            <div className="flex justify-end gap-1">
                              <Button variant="ghost" size="sm" onClick={() => openEdit(load)}>
                                <Pencil className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => setDeletingId(load.id)}
                              >
                                <Trash2 className="h-4 w-4 text-destructive" />
                              </Button>
                            </div>
                          </td>
                        )}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Dialog open={formOpen} onOpenChange={setFormOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? t('form.editTitle') : t('form.createTitle')}</DialogTitle>
          </DialogHeader>
          <TeachingLoadForm
            entity={editing}
            onSubmit={handleSubmit}
            onCancel={() => setFormOpen(false)}
          />
        </DialogContent>
      </Dialog>

      <ConfirmDeleteDialog
        open={deletingId !== null}
        onConfirm={handleDelete}
        onCancel={() => setDeletingId(null)}
      />
    </AppLayout>
  )
}
