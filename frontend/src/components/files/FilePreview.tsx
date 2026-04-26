'use client'

import { useTranslations } from 'next-intl'
import { X, File } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface FilePreviewProps {
  fileName: string
  mimeType: string
  downloadUrl: string
  onClose: () => void
}

export function FilePreview({ fileName, mimeType, downloadUrl, onClose }: FilePreviewProps) {
  const t = useTranslations('files')

  const isImage = mimeType.startsWith('image/')
  const isPdf = mimeType === 'application/pdf'

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold truncate">{fileName}</h3>
        <Button variant="ghost" size="icon" aria-label={t('preview.close')} onClick={onClose}>
          <X className="h-4 w-4" />
        </Button>
      </div>

      <div className="flex items-center justify-center rounded-lg border bg-muted/30 min-h-[300px]">
        {isImage && (
          <img
            src={downloadUrl}
            alt={fileName}
            className="max-h-[500px] max-w-full object-contain rounded"
          />
        )}

        {isPdf && (
          <iframe
            src={downloadUrl}
            title={fileName}
            sandbox="allow-same-origin"
            className="h-[500px] w-full rounded"
          />
        )}

        {!isImage && !isPdf && (
          <div className="flex flex-col items-center gap-2 text-muted-foreground">
            <File className="h-12 w-12" />
            <p className="text-sm">{t('preview.noPreview')}</p>
          </div>
        )}
      </div>
    </div>
  )
}
