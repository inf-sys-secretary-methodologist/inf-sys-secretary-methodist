'use client'

import { useState, useEffect, useCallback, useMemo } from 'react'
import { useTranslations } from 'next-intl'
import { History, AlertCircle, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { documentsApi, DocumentVersionListOutput, VersionDiffOutput } from '@/lib/api/documents'
import { VersionListHeader } from './version-history/VersionListHeader'
import { CreateVersionForm } from './version-history/CreateVersionForm'
import { VersionComparisonView } from './version-history/VersionComparisonView'
import { VersionListItem } from './version-history/VersionListItem'

interface DocumentVersionHistoryProps {
  documentId: number | string
  currentVersion?: number
  onVersionRestored?: () => void
  className?: string
}

export function DocumentVersionHistory({
  documentId,
  currentVersion,
  onVersionRestored,
  className = '',
}: DocumentVersionHistoryProps) {
  const t = useTranslations('documents.versions')

  // Data state
  const [versionData, setVersionData] = useState<DocumentVersionListOutput | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // UI state
  const [expandedVersion, setExpandedVersion] = useState<number | null>(null)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newVersionDescription, setNewVersionDescription] = useState('')
  const [isCreatingVersion, setIsCreatingVersion] = useState(false)
  const [restoringVersion, setRestoringVersion] = useState<number | null>(null)
  const [deletingVersion, setDeletingVersion] = useState<number | null>(null)

  // Comparison state
  const [compareMode, setCompareMode] = useState(false)
  const [selectedVersions, setSelectedVersions] = useState<number[]>([])
  const [comparisonResult, setComparisonResult] = useState<VersionDiffOutput | null>(null)
  const [isComparing, setIsComparing] = useState(false)

  const fetchVersions = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const data = await documentsApi.getVersions(documentId)
      setVersionData(data)
    } catch (err) {
      console.error('Failed to fetch versions:', err)
      setError(t('loadError'))
    } finally {
      setIsLoading(false)
    }
  }, [documentId, t])

  useEffect(() => {
    fetchVersions()
  }, [fetchVersions])

  /* c8 ignore start -- Event handlers with browser dialogs and async API calls, tested via E2E */
  const handleCreateVersion = useCallback(async () => {
    try {
      setIsCreatingVersion(true)
      await documentsApi.createVersion(documentId, {
        change_description: newVersionDescription || undefined,
      })
      setNewVersionDescription('')
      setShowCreateForm(false)
      await fetchVersions()
    } catch (err) {
      console.error('Failed to create version:', err)
      setError(t('createError'))
    } finally {
      setIsCreatingVersion(false)
    }
  }, [documentId, newVersionDescription, fetchVersions, t])

  const handleCancelCreate = useCallback(() => {
    setShowCreateForm(false)
    setNewVersionDescription('')
  }, [])

  const handleRestoreVersion = useCallback(
    async (version: number) => {
      if (!confirm(t('confirmRestore', { version: version.toString() }))) {
        return
      }
      try {
        setRestoringVersion(version)
        await documentsApi.restoreVersion(documentId, version)
        await fetchVersions()
        onVersionRestored?.()
      } catch (err) {
        console.error('Failed to restore version:', err)
        setError(t('restoreError'))
      } finally {
        setRestoringVersion(null)
      }
    },
    [documentId, fetchVersions, onVersionRestored, t]
  )

  const handleDeleteVersion = useCallback(
    async (version: number) => {
      if (!confirm(t('confirmDelete', { version: version.toString() }))) {
        return
      }
      try {
        setDeletingVersion(version)
        await documentsApi.deleteVersion(documentId, version)
        await fetchVersions()
      } catch (err) {
        console.error('Failed to delete version:', err)
        setError(t('deleteError'))
      } finally {
        setDeletingVersion(null)
      }
    },
    [documentId, fetchVersions, t]
  )

  const handleVersionSelect = useCallback(
    (version: number) => {
      if (!compareMode) return

      setSelectedVersions((prev) => {
        if (prev.includes(version)) {
          return prev.filter((v) => v !== version)
        }
        if (prev.length >= 2) {
          return [prev[1], version]
        }
        return [...prev, version]
      })
    },
    [compareMode]
  )

  const handleCompare = useCallback(async () => {
    if (selectedVersions.length !== 2) return

    try {
      setIsComparing(true)
      const [from, to] = selectedVersions.sort((a, b) => a - b)
      const result = await documentsApi.compareVersions(documentId, from, to)
      setComparisonResult(result)
    } catch (err) {
      console.error('Failed to compare versions:', err)
      setError(t('compareError'))
    } finally {
      setIsComparing(false)
    }
  }, [documentId, selectedVersions, t])

  const handleDownloadVersionFile = useCallback(
    async (version: number) => {
      try {
        const fileInfo = await documentsApi.getVersionFile(documentId, version)
        if (fileInfo.download_url) {
          window.open(fileInfo.download_url, '_blank')
        }
      } catch (err) {
        console.error('Failed to get version file:', err)
        setError(t('downloadError'))
      }
    },
    [documentId, t]
  )
  /* c8 ignore stop */

  const resetCompareMode = useCallback(() => {
    setCompareMode(false)
    setSelectedVersions([])
    setComparisonResult(null)
  }, [])

  const formatDate = useCallback((dateString: string) => {
    return new Intl.DateTimeFormat('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(dateString))
  }, [])

  /* c8 ignore next 2 - Data fallbacks */
  const versions = versionData?.versions || []
  const latestVersion = versionData?.latest_version || currentVersion || 1

  const canCompare = useMemo(() => versions.length >= 2, [versions.length])

  if (isLoading) {
    return (
      <div className={`flex items-center justify-center p-8 ${className}`}>
        <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
        <span className="ml-2 text-gray-500">{t('loadingHistory')}</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className={`p-4 bg-red-50 dark:bg-red-900/20 rounded-lg ${className}`}>
        <div className="flex items-center gap-2 text-red-600 dark:text-red-400">
          <AlertCircle className="h-5 w-5" />
          <span>{error}</span>
        </div>
        <Button variant="outline" size="sm" onClick={fetchVersions} className="mt-2">
          {t('retry')}
        </Button>
      </div>
    )
  }

  return (
    <div className={`space-y-4 ${className}`}>
      <VersionListHeader
        totalVersions={versionData?.total || 0}
        compareMode={compareMode}
        selectedVersionsCount={selectedVersions.length}
        isComparing={isComparing}
        canCompare={canCompare}
        onCompare={handleCompare}
        onCancelCompare={resetCompareMode}
        onToggleCompareMode={() => setCompareMode(true)}
        onShowCreateForm={() => setShowCreateForm(!showCreateForm)}
      />

      {showCreateForm && (
        <CreateVersionForm
          description={newVersionDescription}
          isCreating={isCreatingVersion}
          onDescriptionChange={setNewVersionDescription}
          onCreate={handleCreateVersion}
          onCancel={handleCancelCreate}
        />
      )}

      {comparisonResult && (
        <VersionComparisonView
          comparisonResult={comparisonResult}
          onClose={() => setComparisonResult(null)}
        />
      )}

      {versions.length === 0 ? (
        <div className="text-center py-8 text-gray-500">
          <History className="h-12 w-12 mx-auto mb-3 opacity-50" />
          <p>{t('empty')}</p>
        </div>
      ) : (
        <div className="space-y-2">
          {versions.map((version) => (
            <VersionListItem
              key={version.id}
              version={version}
              isLatest={version.version === latestVersion}
              isExpanded={expandedVersion === version.version}
              isSelected={selectedVersions.includes(version.version)}
              compareMode={compareMode}
              isRestoring={restoringVersion === version.version}
              isDeleting={deletingVersion === version.version}
              /* c8 ignore start - Version item callbacks */
              onToggleExpand={() =>
                setExpandedVersion(expandedVersion === version.version ? null : version.version)
              }
              onSelect={() => handleVersionSelect(version.version)}
              onRestore={() => handleRestoreVersion(version.version)}
              onDelete={() => handleDeleteVersion(version.version)}
              onDownloadFile={() => handleDownloadVersionFile(version.version)}
              /* c8 ignore stop */
              formatDate={formatDate}
            />
          ))}
        </div>
      )}
    </div>
  )
}
