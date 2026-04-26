'use client'

import { useMemo, useState } from 'react'
import { useTranslations } from 'next-intl'
import { FolderOpen, Loader2 } from 'lucide-react'
import { toast } from 'sonner'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { FileUploader } from '@/components/files/FileUploader'
import { FileGrid } from '@/components/files/FileGrid'
import { FilePreview } from '@/components/files/FilePreview'
import { FileFilters, type FileFilterValues } from '@/components/files/FileFilters'
import { ConfirmDeleteDialog } from '@/components/files/ConfirmDeleteDialog'
import { VersionHistory } from '@/components/files/VersionHistory'
import {
  useFiles,
  useFileVersions,
  uploadFile,
  deleteFile,
  downloadFile,
  downloadFileVersion,
  createFileVersion,
} from '@/hooks/useFiles'
import type { FileItem, FileFilterParams } from '@/types/files'
import { useAuthCheck } from '@/hooks/useAuth'
import { useAuthStore } from '@/stores/authStore'
import { canEdit } from '@/lib/auth/permissions'

const MIME_TYPE_GROUPS: Record<string, (mime: string) => boolean> = {
  image: (m) => m.startsWith('image/'),
  documents: (m) => m.includes('pdf') || m.includes('document') || m.includes('text/'),
  spreadsheets: (m) => m.includes('spreadsheet') || m.includes('excel') || m.includes('csv'),
  presentations: (m) => m.includes('presentation') || m.includes('powerpoint'),
  archives: (m) => m.includes('zip') || m.includes('tar') || m.includes('rar') || m.includes('7z'),
  other: () => true,
}

function matchesMimeGroup(mimeType: string, group: string): boolean {
  if (group === 'other') {
    return !['image', 'documents', 'spreadsheets', 'presentations', 'archives'].some(
      (g) => MIME_TYPE_GROUPS[g](mimeType)
    )
  }
  return MIME_TYPE_GROUPS[group]?.(mimeType) ?? false
}

export default function FilesPage() {
  const t = useTranslations('files')
  useAuthCheck()
  const user = useAuthStore((s) => s.user)
  const userCanEdit = canEdit(user?.role)

  const [paginationParams, setPaginationParams] = useState<FileFilterParams>({ page: 1, limit: 100 })
  const [clientFilters, setClientFilters] = useState<FileFilterValues>({})
  const [uploading, setUploading] = useState(false)
  const [previewFile, setPreviewFile] = useState<(FileItem & { downloadUrl?: string }) | null>(null)
  const [versionsFileId, setVersionsFileId] = useState<number | null>(null)
  const [versionUploadOpen, setVersionUploadOpen] = useState(false)
  const [deleteFileId, setDeleteFileId] = useState<number | null>(null)

  const { files, totalPages, isLoading, error, mutate } = useFiles(paginationParams)
  const { versions, mutate: mutateVersions } = useFileVersions(versionsFileId)

  const filteredFiles = useMemo(() => {
    let result = files
    if (clientFilters.search) {
      const q = clientFilters.search.toLowerCase()
      result = result.filter((f) => f.original_name.toLowerCase().includes(q))
    }
    if (clientFilters.fileType) {
      result = result.filter((f) => matchesMimeGroup(f.mime_type, clientFilters.fileType!))
    }
    return result
  }, [files, clientFilters])

  const handleUpload = async (file: File) => {
    setUploading(true)
    try {
      await uploadFile(file)
      toast.success(t('dropzone.success'))
      await mutate()
    } catch {
      toast.error(t('errors.uploadFailed'))
    } finally {
      setUploading(false)
    }
  }

  const handleDownload = async (id: number) => {
    try {
      const result = await downloadFile(id)
      window.open(result.presigned_url, '_blank')
    } catch {
      toast.error(t('errors.downloadFailed'))
    }
  }

  const handleDeleteRequest = (id: number) => {
    setDeleteFileId(id)
  }

  const handleDeleteConfirm = async () => {
    if (deleteFileId === null) return
    try {
      await deleteFile(deleteFileId)
      await mutate()
    } catch {
      toast.error(t('errors.deleteFailed'))
    } finally {
      setDeleteFileId(null)
    }
  }

  const handlePreview = async (file: FileItem) => {
    try {
      const result = await downloadFile(file.id)
      setPreviewFile({ ...file, downloadUrl: result.presigned_url })
    } catch {
      toast.error(t('errors.downloadFailed'))
    }
  }

  const handleVersionDownload = async (versionNumber: number) => {
    if (!versionsFileId) return
    try {
      const result = await downloadFileVersion(versionsFileId, versionNumber)
      window.open(result.presigned_url, '_blank')
    } catch {
      toast.error(t('errors.downloadFailed'))
    }
  }

  const handleVersionUpload = async (file: File) => {
    if (!versionsFileId) return
    setUploading(true)
    try {
      await createFileVersion(versionsFileId, file)
      toast.success(t('dropzone.success'))
      await mutateVersions()
    } catch {
      toast.error(t('errors.versionFailed'))
    } finally {
      setUploading(false)
      setVersionUploadOpen(false)
    }
  }

  const handlePageChange = (page: number) => {
    setPaginationParams((prev) => ({ ...prev, page }))
  }

  return (
    <AppLayout>
      <div className="mx-auto max-w-6xl space-y-6 p-4 md:p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <FolderOpen className="h-6 w-6" />
            <h1 className="text-2xl font-bold">{t('title')}</h1>
          </div>
        </div>

        {userCanEdit && <FileUploader onUpload={handleUpload} uploading={uploading} />}

        <FileFilters value={clientFilters} onChange={setClientFilters} />

        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            <span className="ml-2 text-muted-foreground">{t('loading')}</span>
          </div>
        ) : error ? (
          <div className="text-center py-12 text-destructive">{t('loadFailed')}</div>
        ) : (
          <>
            <FileGrid
              files={filteredFiles}
              onDownload={handleDownload}
              onDelete={userCanEdit ? handleDeleteRequest : undefined}
              onPreview={handlePreview}
              onVersions={(id) => setVersionsFileId(id)}
            />

            {totalPages > 1 && (
              <div className="flex items-center justify-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={paginationParams.page === 1}
                  onClick={() => handlePageChange((paginationParams.page || 1) - 1)}
                >
                  &laquo;
                </Button>
                <span className="text-sm text-muted-foreground">
                  {paginationParams.page} / {totalPages}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={paginationParams.page === totalPages}
                  onClick={() => handlePageChange((paginationParams.page || 1) + 1)}
                >
                  &raquo;
                </Button>
              </div>
            )}
          </>
        )}

        <Dialog open={!!previewFile} onOpenChange={(open) => !open && setPreviewFile(null)}>
          <DialogContent className="max-w-3xl">
            {previewFile && previewFile.downloadUrl && (
              <FilePreview
                fileName={previewFile.original_name}
                mimeType={previewFile.mime_type}
                downloadUrl={previewFile.downloadUrl}
                onClose={() => setPreviewFile(null)}
              />
            )}
          </DialogContent>
        </Dialog>

        <Dialog open={versionsFileId !== null} onOpenChange={(open) => {
          if (!open) {
            setVersionsFileId(null)
            setVersionUploadOpen(false)
          }
        }}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{t('versions.title')}</DialogTitle>
            </DialogHeader>
            <VersionHistory versions={versions} onDownload={handleVersionDownload} />
            {userCanEdit && (
              <div className="mt-4">
                {versionUploadOpen ? (
                  <FileUploader onUpload={handleVersionUpload} uploading={uploading} />
                ) : (
                  <Button variant="outline" onClick={() => setVersionUploadOpen(true)}>
                    {t('versions.createVersion')}
                  </Button>
                )}
              </div>
            )}
          </DialogContent>
        </Dialog>

        <ConfirmDeleteDialog
          open={deleteFileId !== null}
          onConfirm={handleDeleteConfirm}
          onCancel={() => setDeleteFileId(null)}
        />
      </div>
    </AppLayout>
  )
}
