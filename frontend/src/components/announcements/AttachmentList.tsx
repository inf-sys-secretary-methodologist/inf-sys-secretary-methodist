'use client'

import * as React from 'react'
import { useId, useRef, useState } from 'react'
import { File, Image as ImageIcon, FileText, Trash2, Upload } from 'lucide-react'
import { useTranslations } from 'next-intl'

import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import type { AnnouncementAttachment } from '@/types/announcements'

interface AttachmentListProps {
  attachments: AnnouncementAttachment[]
  onUpload?: (file: File) => Promise<void> | void
  onRemove?: (attachmentId: number) => Promise<void> | void
  className?: string
}

// formatFileSize converts a byte count into a human-readable string.
// 1024 → "1.0 KB", 2_097_152 → "2.0 MB" — matches the assertions in tests.
function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
}

function iconForMime(mime: string) {
  if (mime.startsWith('image/')) return ImageIcon
  if (mime === 'application/pdf' || mime.startsWith('text/')) return FileText
  return File
}

export function AttachmentList({
  attachments,
  onUpload,
  onRemove,
  className,
}: AttachmentListProps) {
  const t = useTranslations('announcements')
  // Unique id so multiple AttachmentList instances on the same page don't
  // collide via the htmlFor → id link.
  const inputId = useId()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState(false)

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file || !onUpload) return
    setUploading(true)
    try {
      await onUpload(file)
    } finally {
      setUploading(false)
      // Reset so the same file can be re-selected later if needed.
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  return (
    <div className={cn('flex flex-col gap-2', className)}>
      {attachments.length === 0 ? (
        <p className="text-sm text-muted-foreground italic">{t('noAttachments')}</p>
      ) : (
        <ul className="flex flex-col gap-1.5">
          {attachments.map((att) => {
            const Icon = iconForMime(att.mime_type)
            return (
              <li
                key={att.id}
                className="flex items-center gap-3 rounded-md border border-border bg-card px-3 py-2"
              >
                <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate text-foreground">{att.file_name}</p>
                  <p className="text-xs text-muted-foreground">{formatFileSize(att.file_size)}</p>
                </div>
                {onRemove && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-muted-foreground hover:text-destructive"
                    onClick={() => onRemove(att.id)}
                    aria-label={t('attachments.remove')}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                )}
              </li>
            )
          })}
        </ul>
      )}

      {onUpload && (
        <div>
          <label
            htmlFor={inputId}
            className="inline-flex items-center gap-2 text-sm font-medium text-primary cursor-pointer hover:underline"
          >
            <Upload className="h-4 w-4" />
            {uploading ? t('attachments.uploading') : t('attachments.uploadFile')}
          </label>
          <input
            id={inputId}
            ref={fileInputRef}
            type="file"
            onChange={handleFileChange}
            disabled={uploading}
            className="sr-only"
          />
        </div>
      )}
    </div>
  )
}
