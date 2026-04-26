'use client'

import { useState } from 'react'
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
import { VersionHistory } from '@/components/files/VersionHistory'
import {
  useFiles,
  useFileVersions,
  uploadFile,
  deleteFile,
  downloadFile,
  createFileVersion,
} from '@/hooks/useFiles'
import type { FileItem, FileFilterParams } from '@/types/files'
import { useAuthCheck } from '@/hooks/useAuth'

export default function FilesPage() {
  const t = useTranslations('files')
  useAuthCheck()
  const [filters, setFilters] = useState<FileFilterParams>({ page: 1, limit: 20 })
  const [uploading, setUploading] = useState(false)
  const [previewFile, setPreviewFile] = useState<(FileItem & { downloadUrl?: string }) | null>(null)
  const [versionsFileId, setVersionsFileId] = useState<number | null>(null)
  const [versionUploadOpen, setVersionUploadOpen] = useState(false)

  const { files, totalPages, isLoading, error, mutate } = useFiles(filters)
  const { versions, mutate: mutateVersions } = useFileVersions(versionsFileId)

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

  const handleDelete = async (id: number) => {
    if (!confirm(t('confirm.delete'))) return
    try {
      await deleteFile(id)
      await mutate()
    } catch {
      toast.error(t('errors.deleteFailed'))
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

  const handleVersionDownload = async (_versionNumber: number) => {
    if (!versionsFileId) return
    try {
      const result = await downloadFile(versionsFileId)
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
    setFilters((prev) => ({ ...prev, page }))
  }

  return (
    <AppLayout>
      <div className="mx-auto max-w-6xl space-y-6 p-4 md:p-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <FolderOpen className="h-6 w-6" />
            <h1 className="text-2xl font-bold">{t('title')}</h1>
          </div>
        </div>

        {/* Upload zone */}
        <FileUploader onUpload={handleUpload} uploading={uploading} />

        {/* File grid */}
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
              files={files}
              onDownload={handleDownload}
              onDelete={handleDelete}
              onPreview={handlePreview}
              onVersions={(id) => setVersionsFileId(id)}
            />

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={filters.page === 1}
                  onClick={() => handlePageChange((filters.page || 1) - 1)}
                >
                  &laquo;
                </Button>
                <span className="text-sm text-muted-foreground">
                  {filters.page} / {totalPages}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={filters.page === totalPages}
                  onClick={() => handlePageChange((filters.page || 1) + 1)}
                >
                  &raquo;
                </Button>
              </div>
            )}
          </>
        )}

        {/* Preview dialog */}
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

        {/* Version history dialog */}
        <Dialog open={versionsFileId !== null} onOpenChange={(open) => !open && setVersionsFileId(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{t('versions.title')}</DialogTitle>
            </DialogHeader>
            <VersionHistory versions={versions} onDownload={handleVersionDownload} />
            <div className="mt-4">
              {versionUploadOpen ? (
                <FileUploader onUpload={handleVersionUpload} uploading={uploading} />
              ) : (
                <Button variant="outline" onClick={() => setVersionUploadOpen(true)}>
                  {t('versions.createVersion')}
                </Button>
              )}
            </div>
          </DialogContent>
        </Dialog>
      </div>
    </AppLayout>
  )
}
