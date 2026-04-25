'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { Plus, Megaphone, Loader2 } from 'lucide-react'
import { toast } from 'sonner'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { AnnouncementCard } from '@/components/announcements/AnnouncementCard'
import { AnnouncementFilters } from '@/components/announcements/AnnouncementFilters'
import { AnnouncementForm } from '@/components/announcements/AnnouncementForm'
import { AttachmentList } from '@/components/announcements/AttachmentList'
import {
  useAnnouncements,
  useAnnouncement,
  createAnnouncement,
  updateAnnouncement,
  deleteAnnouncement,
  publishAnnouncement,
  unpublishAnnouncement,
  archiveAnnouncement,
  uploadAnnouncementAttachment,
  deleteAnnouncementAttachment,
} from '@/hooks/useAnnouncements'
import type {
  Announcement,
  AnnouncementFilterParams,
  AnnouncementStatus,
  CreateAnnouncementInput,
} from '@/types/announcements'
import { useAuthCheck } from '@/hooks/useAuth'
import { canEdit } from '@/lib/auth/permissions'
import { useAuthStore } from '@/stores/authStore'

type StatusTab = 'all' | AnnouncementStatus

// Backend supports up to 100 (validate min=1,max=100 in ListAnnouncementsRequest).
// Pagination UI is tracked separately; until then this acts as a single-page
// fetch ceiling.
const ANNOUNCEMENTS_PAGE_SIZE = 100

export default function AnnouncementsPage() {
  const t = useTranslations('announcements')
  useAuthCheck()
  const user = useAuthStore((s) => s.user)
  const userCanEdit = canEdit(user?.role)

  const [tab, setTab] = useState<StatusTab>('all')
  const [filters, setFilters] = useState<AnnouncementFilterParams>({})
  const [editingId, setEditingId] = useState<number | null>(null)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [creating, setCreating] = useState(false)

  const effectiveFilters: AnnouncementFilterParams = {
    ...filters,
    status: tab === 'all' ? filters.status : tab,
    limit: ANNOUNCEMENTS_PAGE_SIZE,
  }
  const { announcements, total, isLoading, error, mutate } = useAnnouncements(effectiveFilters)
  const { announcement: editingAnnouncement, mutate: mutateOne } = useAnnouncement(editingId)

  const openCreate = () => {
    setEditingId(null)
    setCreating(true)
    setIsFormOpen(true)
  }

  const openEdit = (a: Announcement) => {
    setCreating(false)
    setEditingId(a.id)
    setIsFormOpen(true)
  }

  const closeForm = () => {
    setIsFormOpen(false)
    setEditingId(null)
    setCreating(false)
  }

  const handleSubmit = async (input: CreateAnnouncementInput) => {
    try {
      if (editingId) {
        await updateAnnouncement(editingId, input)
      } else {
        await createAnnouncement(input)
      }
      closeForm()
      await mutate()
    } catch {
      toast.error(t(editingId ? 'errors.updateFailed' : 'errors.createFailed'))
    }
  }

  const handleDelete = async (a: Announcement) => {
    try {
      await deleteAnnouncement(a.id)
      await mutate()
    } catch {
      toast.error(t('errors.deleteFailed'))
    }
  }

  const handlePublish = async (a: Announcement) => {
    try {
      await publishAnnouncement(a.id)
      await mutate()
    } catch {
      toast.error(t('errors.publishFailed'))
    }
  }

  const handleUnpublish = async (a: Announcement) => {
    try {
      await unpublishAnnouncement(a.id)
      await mutate()
    } catch {
      toast.error(t('errors.unpublishFailed'))
    }
  }

  const handleArchive = async (a: Announcement) => {
    try {
      await archiveAnnouncement(a.id)
      await mutate()
    } catch {
      toast.error(t('errors.archiveFailed'))
    }
  }

  const handleUploadAttachment = async (file: File) => {
    if (!editingId) return
    try {
      await uploadAnnouncementAttachment(editingId, file)
      await mutateOne()
    } catch {
      toast.error(t('errors.uploadFailed'))
    }
  }

  const handleRemoveAttachment = async (attachmentId: number) => {
    if (!editingId) return
    try {
      await deleteAnnouncementAttachment(editingId, attachmentId)
      await mutateOne()
    } catch {
      toast.error(t('errors.removeAttachmentFailed'))
    }
  }

  const tabValues: StatusTab[] = ['all', 'draft', 'published', 'archived']

  return (
    <AppLayout>
      <div className="max-w-6xl mx-auto space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">{t('title')}</h1>
            <p className="text-muted-foreground">{t('description')}</p>
          </div>
          {userCanEdit && (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4 mr-2" />
              {t('create')}
            </Button>
          )}
        </div>

        {userCanEdit && (
          <Tabs value={tab} onValueChange={(v) => setTab(v as StatusTab)}>
            <TabsList>
              {tabValues.map((v) => (
                <TabsTrigger key={v} value={v}>
                  {t(`tabs.${v}`)}
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
        )}

        <AnnouncementFilters value={filters} onChange={setFilters} />

        {isLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="rounded-xl bg-card border border-border p-8 text-center">
            <p className="text-destructive font-medium">{t('loadFailed')}</p>
          </div>
        ) : announcements.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <Megaphone className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-medium">{t('noAnnouncements')}</h3>
          </div>
        ) : (
          <div className="grid gap-3">
            {announcements.map((a) => (
              <AnnouncementCard
                key={a.id}
                announcement={a}
                onClick={() => openEdit(a)}
                onEdit={userCanEdit ? () => openEdit(a) : undefined}
                onDelete={userCanEdit ? () => handleDelete(a) : undefined}
                onPublish={userCanEdit ? () => handlePublish(a) : undefined}
                onUnpublish={userCanEdit ? () => handleUnpublish(a) : undefined}
                onArchive={userCanEdit ? () => handleArchive(a) : undefined}
              />
            ))}
          </div>
        )}

        {announcements.length > 0 && (
          <p className="text-sm text-muted-foreground text-right">
            {announcements.length} / {total}
          </p>
        )}

        <Dialog open={isFormOpen} onOpenChange={(open) => !open && closeForm()}>
          <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>
                {creating ? t('form.createTitle') : t('form.editTitle')}
              </DialogTitle>
            </DialogHeader>
            <AnnouncementForm
              announcement={editingAnnouncement}
              onSubmit={handleSubmit}
              onCancel={closeForm}
            />
            {!creating && editingId && editingAnnouncement && (
              <div className="mt-6 border-t border-border pt-4">
                <h3 className="text-sm font-semibold text-foreground mb-2">
                  {t('attachments.title')}
                </h3>
                <AttachmentList
                  attachments={editingAnnouncement.attachments ?? []}
                  onUpload={userCanEdit ? handleUploadAttachment : undefined}
                  onRemove={userCanEdit ? handleRemoveAttachment : undefined}
                />
              </div>
            )}
          </DialogContent>
        </Dialog>
      </div>
    </AppLayout>
  )
}
