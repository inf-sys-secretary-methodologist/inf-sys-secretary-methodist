'use client'

import { useTranslations } from 'next-intl'
import { Download, Trash2, History, FileText, FileImage, FileSpreadsheet, File } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { FileItem } from '@/types/files'

interface FileGridProps {
  files: FileItem[]
  onDownload?: (id: number) => void
  onDelete?: (id: number) => void
  onPreview?: (file: FileItem) => void
  onVersions?: (id: number) => void
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
}

function getMimeIcon(mimeType: string) {
  if (mimeType.startsWith('image/')) return FileImage
  if (mimeType.includes('spreadsheet') || mimeType.includes('excel')) return FileSpreadsheet
  if (mimeType.includes('pdf') || mimeType.includes('document') || mimeType.includes('text'))
    return FileText
  return File
}

export function FileGrid({ files, onDownload, onDelete, onPreview, onVersions }: FileGridProps) {
  const t = useTranslations('files')

  if (files.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <File className="mb-2 h-10 w-10" />
        <p>{t('noFiles')}</p>
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{t('grid.name')}</TableHead>
          <TableHead>{t('grid.size')}</TableHead>
          <TableHead>{t('grid.type')}</TableHead>
          <TableHead>{t('grid.uploaded')}</TableHead>
          <TableHead className="text-right">{t('grid.actions')}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {files.map((file) => {
          const Icon = getMimeIcon(file.mime_type)
          return (
            <TableRow key={file.id}>
              <TableCell>
                <button
                  type="button"
                  className="flex items-center gap-2 text-left hover:underline"
                  onClick={() => onPreview?.(file)}
                >
                  <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
                  <span className="truncate">{file.original_name}</span>
                </button>
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatFileSize(file.size)}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {file.mime_type.split('/').pop()}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {new Date(file.created_at).toLocaleDateString()}
              </TableCell>
              <TableCell className="text-right">
                <div className="flex items-center justify-end gap-1">
                  {onVersions && (
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={t('versions.title')}
                      onClick={() => onVersions(file.id)}
                    >
                      <History className="h-4 w-4" />
                    </Button>
                  )}
                  {onDownload && (
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={t('download')}
                      onClick={() => onDownload(file.id)}
                    >
                      <Download className="h-4 w-4" />
                    </Button>
                  )}
                  {onDelete && (
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={t('delete')}
                      onClick={() => onDelete(file.id)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              </TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
