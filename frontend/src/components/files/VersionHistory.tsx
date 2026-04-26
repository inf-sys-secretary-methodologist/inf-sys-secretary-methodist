'use client'

import { useTranslations } from 'next-intl'
import { Download, History } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { FileVersion } from '@/types/files'

interface VersionHistoryProps {
  versions: FileVersion[]
  onDownload?: (versionNumber: number) => void
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function VersionHistory({ versions, onDownload }: VersionHistoryProps) {
  const t = useTranslations('files')

  return (
    <div className="flex flex-col gap-3">
      <h3 className="text-lg font-semibold flex items-center gap-2">
        <History className="h-5 w-5" />
        {t('versions.title')}
      </h3>

      {versions.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t('versions.noVersions')}</p>
      ) : (
        <div className="space-y-2">
          {versions.map((version, idx) => (
            <div
              key={version.id}
              className="flex items-center justify-between rounded-lg border p-3"
            >
              <div className="flex flex-col gap-0.5">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">
                    {t('versions.version', { number: version.version_number })}
                  </span>
                  {idx === 0 && (
                    <span className="rounded bg-primary/10 px-1.5 py-0.5 text-xs text-primary">
                      {t('versions.current')}
                    </span>
                  )}
                </div>
                <span className="text-xs text-muted-foreground">
                  {formatFileSize(version.size)} &middot;{' '}
                  {new Date(version.created_at).toLocaleDateString()}
                </span>
                {version.comment && (
                  <span className="text-xs text-muted-foreground italic">
                    {version.comment}
                  </span>
                )}
              </div>
              {onDownload && (
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={t('versions.download')}
                  onClick={() => onDownload(version.version_number)}
                >
                  <Download className="h-4 w-4" />
                </Button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
