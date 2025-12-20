'use client'

import { useState, useEffect, useCallback } from 'react'
import dynamic from 'next/dynamic'
import { useTranslations } from 'next-intl'
import {
  History,
  RotateCcw,
  GitCompare,
  Download,
  Trash2,
  Plus,
  Clock,
  User,
  ChevronDown,
  ChevronUp,
  AlertCircle,
  Check,
  Loader2,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  documentsApi,
  DocumentVersionInfo,
  DocumentVersionListOutput,
  VersionDiffOutput,
} from '@/lib/api/documents'

// Lazy load TextDiff to reduce initial bundle (diff library ~15KB)
const TextDiff = dynamic(() => import('./TextDiff').then((mod) => mod.TextDiff), {
  loading: () => <div className="animate-pulse bg-gray-100 dark:bg-gray-800 rounded h-24" />,
  ssr: false,
})

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
  const tCommon = useTranslations('common')
  const [versionData, setVersionData] = useState<DocumentVersionListOutput | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedVersion, setExpandedVersion] = useState<number | null>(null)
  const [isCreatingVersion, setIsCreatingVersion] = useState(false)
  const [newVersionDescription, setNewVersionDescription] = useState('')
  const [showCreateForm, setShowCreateForm] = useState(false)
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
  }, [documentId])

  useEffect(() => {
    fetchVersions()
  }, [fetchVersions])

  const handleCreateVersion = async () => {
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
  }

  const handleRestoreVersion = async (version: number) => {
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
  }

  const handleDeleteVersion = async (version: number) => {
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
  }

  const handleVersionSelect = (version: number) => {
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
  }

  const handleCompare = async () => {
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
  }

  const handleDownloadVersionFile = async (version: number) => {
    try {
      const fileInfo = await documentsApi.getVersionFile(documentId, version)
      if (fileInfo.download_url) {
        window.open(fileInfo.download_url, '_blank')
      }
    } catch (err) {
      console.error('Failed to get version file:', err)
      setError(t('downloadError'))
    }
  }

  const formatDate = (dateString: string) => {
    return new Intl.DateTimeFormat('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(dateString))
  }

  const resetCompareMode = () => {
    setCompareMode(false)
    setSelectedVersions([])
    setComparisonResult(null)
  }

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

  const versions = versionData?.versions || []
  const latestVersion = versionData?.latest_version || currentVersion || 1

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <History className="h-5 w-5 text-gray-500" />
          <h3 className="font-semibold text-gray-900 dark:text-white">{t('history')}</h3>
          <span className="text-sm text-gray-500">
            ({versionData?.total || 0} {t('versionsCount')})
          </span>
        </div>

        <div className="flex items-center gap-2">
          {compareMode ? (
            <>
              <Button
                variant="outline"
                size="sm"
                onClick={handleCompare}
                disabled={selectedVersions.length !== 2 || isComparing}
              >
                {isComparing ? (
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                ) : (
                  <GitCompare className="h-4 w-4 mr-2" />
                )}
                {t('compareCount', { count: selectedVersions.length })}
              </Button>
              <Button variant="ghost" size="sm" onClick={resetCompareMode}>
                {tCommon('cancel')}
              </Button>
            </>
          ) : (
            <>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setCompareMode(true)}
                disabled={versions.length < 2}
              >
                <GitCompare className="h-4 w-4 mr-2" />
                {t('compare')}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowCreateForm(!showCreateForm)}
              >
                <Plus className="h-4 w-4 mr-2" />
                {t('createVersion')}
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Create Version Form */}
      {showCreateForm && (
        <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg space-y-3">
          <input
            type="text"
            placeholder={t('descriptionPlaceholder')}
            value={newVersionDescription}
            onChange={(e) => setNewVersionDescription(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md
                       bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          <div className="flex gap-2">
            <Button size="sm" onClick={handleCreateVersion} disabled={isCreatingVersion}>
              {isCreatingVersion ? (
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
              ) : (
                <Check className="h-4 w-4 mr-2" />
              )}
              {t('saveVersion')}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                setShowCreateForm(false)
                setNewVersionDescription('')
              }}
            >
              {tCommon('cancel')}
            </Button>
          </div>
        </div>
      )}

      {/* Comparison Result */}
      {comparisonResult && (
        <div className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg space-y-3">
          <div className="flex items-center justify-between">
            <h4 className="font-medium text-blue-900 dark:text-blue-100">
              {t('comparisonTitle', {
                from: comparisonResult.from_version,
                to: comparisonResult.to_version,
              })}
            </h4>
            <Button variant="ghost" size="sm" onClick={() => setComparisonResult(null)}>
              {tCommon('close')}
            </Button>
          </div>

          {comparisonResult.changed_fields.length === 0 ? (
            <p className="text-sm text-blue-700 dark:text-blue-300">{t('versionsIdentical')}</p>
          ) : (
            <div className="space-y-2">
              <p className="text-sm text-blue-700 dark:text-blue-300">
                {t('changedFields')}: {comparisonResult.changed_fields.join(', ')}
              </p>
              {comparisonResult.diff_data && (
                <div className="space-y-4 mt-3">
                  {Object.entries(comparisonResult.diff_data).map(([field, diff]) => {
                    const fieldLabels: Record<string, string> = {
                      title: t('fieldLabels.title'),
                      subject: t('fieldLabels.subject'),
                      content: t('fieldLabels.content'),
                      status: t('fieldLabels.status'),
                      file_name: t('fieldLabels.file_name'),
                    }
                    const isTextContent =
                      field === 'content' || field === 'subject' || field === 'title'
                    const oldValue = String(diff.from || '')
                    const newValue = String(diff.to || '')

                    return (
                      <div
                        key={field}
                        className="p-4 bg-white dark:bg-gray-800 rounded-lg border border-blue-200 dark:border-blue-700"
                      >
                        <div className="font-medium text-sm text-gray-700 dark:text-gray-300 mb-3">
                          {fieldLabels[field] || field}
                        </div>

                        {isTextContent && (oldValue.length > 20 || newValue.length > 20) ? (
                          <TextDiff
                            oldText={oldValue}
                            newText={newValue}
                            oldLabel={`${t('version')} ${comparisonResult.from_version}`}
                            newLabel={`${t('version')} ${comparisonResult.to_version}`}
                          />
                        ) : (
                          <div className="grid grid-cols-2 gap-4 text-sm">
                            <div>
                              <span className="text-red-600 dark:text-red-400 font-medium">
                                {t('before')}:
                              </span>
                              <p className="mt-1 text-gray-600 dark:text-gray-400 bg-red-50 dark:bg-red-900/20 p-2 rounded break-words">
                                {oldValue || t('emptyValue')}
                              </p>
                            </div>
                            <div>
                              <span className="text-green-600 dark:text-green-400 font-medium">
                                {t('after')}:
                              </span>
                              <p className="mt-1 text-gray-600 dark:text-gray-400 bg-green-50 dark:bg-green-900/20 p-2 rounded break-words">
                                {newValue || t('emptyValue')}
                              </p>
                            </div>
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Versions List */}
      {versions.length === 0 ? (
        <div className="text-center py-8 text-gray-500">
          <History className="h-12 w-12 mx-auto mb-3 opacity-50" />
          <p>{t('empty')}</p>
        </div>
      ) : (
        <div className="space-y-2">
          {versions.map((version) => (
            <VersionItem
              key={version.id}
              version={version}
              isLatest={version.version === latestVersion}
              isExpanded={expandedVersion === version.version}
              isSelected={selectedVersions.includes(version.version)}
              compareMode={compareMode}
              isRestoring={restoringVersion === version.version}
              isDeleting={deletingVersion === version.version}
              onToggleExpand={() =>
                setExpandedVersion(expandedVersion === version.version ? null : version.version)
              }
              onSelect={() => handleVersionSelect(version.version)}
              onRestore={() => handleRestoreVersion(version.version)}
              onDelete={() => handleDeleteVersion(version.version)}
              onDownloadFile={() => handleDownloadVersionFile(version.version)}
              formatDate={formatDate}
            />
          ))}
        </div>
      )}
    </div>
  )
}

interface VersionItemProps {
  version: DocumentVersionInfo
  isLatest: boolean
  isExpanded: boolean
  isSelected: boolean
  compareMode: boolean
  isRestoring: boolean
  isDeleting: boolean
  onToggleExpand: () => void
  onSelect: () => void
  onRestore: () => void
  onDelete: () => void
  onDownloadFile: () => void
  formatDate: (date: string) => string
}

function VersionItem({
  version,
  isLatest,
  isExpanded,
  isSelected,
  compareMode,
  isRestoring,
  isDeleting,
  onToggleExpand,
  onSelect,
  onRestore,
  onDelete,
  onDownloadFile,
  formatDate,
}: VersionItemProps) {
  const t = useTranslations('documents.versions')
  const tCommon = useTranslations('common')
  return (
    <div
      className={`
        border rounded-lg overflow-hidden transition-all
        ${isLatest ? 'border-blue-300 dark:border-blue-600' : 'border-gray-200 dark:border-gray-700'}
        ${isSelected ? 'ring-2 ring-blue-500' : ''}
        ${compareMode ? 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50' : ''}
      `}
      onClick={compareMode ? onSelect : undefined}
    >
      {/* Header */}
      <div
        className={`
          flex items-center justify-between p-3
          ${!compareMode ? 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50' : ''}
        `}
        onClick={!compareMode ? onToggleExpand : undefined}
      >
        <div className="flex items-center gap-3">
          {compareMode && (
            <div
              className={`
                w-5 h-5 rounded border-2 flex items-center justify-center
                ${isSelected ? 'bg-blue-500 border-blue-500' : 'border-gray-300 dark:border-gray-600'}
              `}
            >
              {isSelected && <Check className="h-3 w-3 text-white" />}
            </div>
          )}

          <div>
            <div className="flex items-center gap-2">
              <span className="font-medium text-gray-900 dark:text-white">
                {t('version')} {version.version}
              </span>
              {isLatest && (
                <span className="px-2 py-0.5 text-xs font-medium bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 rounded">
                  {t('current')}
                </span>
              )}
            </div>
            <div className="flex items-center gap-3 text-sm text-gray-500 mt-1">
              <span className="flex items-center gap-1">
                <Clock className="h-3 w-3" />
                {formatDate(version.created_at)}
              </span>
              {version.changed_by_name && (
                <span className="flex items-center gap-1">
                  <User className="h-3 w-3" />
                  {version.changed_by_name}
                </span>
              )}
            </div>
          </div>
        </div>

        {!compareMode && (
          <div className="flex items-center gap-2">
            {version.file_name && (
              <Button
                variant="ghost"
                size="sm"
                onClick={(e) => {
                  e.stopPropagation()
                  onDownloadFile()
                }}
              >
                <Download className="h-4 w-4" />
              </Button>
            )}

            {!isLatest && (
              <>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation()
                    onRestore()
                  }}
                  disabled={isRestoring}
                >
                  {isRestoring ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <RotateCcw className="h-4 w-4" />
                  )}
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation()
                    onDelete()
                  }}
                  disabled={isDeleting}
                  className="text-red-500 hover:text-red-700"
                >
                  {isDeleting ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Trash2 className="h-4 w-4" />
                  )}
                </Button>
              </>
            )}

            {isExpanded ? (
              <ChevronUp className="h-4 w-4 text-gray-400" />
            ) : (
              <ChevronDown className="h-4 w-4 text-gray-400" />
            )}
          </div>
        )}
      </div>

      {/* Expanded Details */}
      {isExpanded && !compareMode && (
        <div className="px-3 pb-3 pt-0 border-t border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/30">
          <div className="grid grid-cols-2 gap-4 text-sm pt-3">
            {version.title && (
              <div>
                <p className="text-gray-500 dark:text-gray-400">{t('title')}</p>
                <p className="text-gray-900 dark:text-white">{version.title}</p>
              </div>
            )}
            {version.status && (
              <div>
                <p className="text-gray-500 dark:text-gray-400">{t('status')}</p>
                <p className="text-gray-900 dark:text-white">{version.status}</p>
              </div>
            )}
            {version.change_description && (
              <div className="col-span-2">
                <p className="text-gray-500 dark:text-gray-400">{t('changes')}</p>
                <p className="text-gray-900 dark:text-white">{version.change_description}</p>
              </div>
            )}
            {version.file_name && (
              <div className="col-span-2">
                <p className="text-gray-500 dark:text-gray-400">{t('file')}</p>
                <p className="text-gray-900 dark:text-white">
                  {version.file_name}
                  {version.file_size && (
                    <span className="text-gray-500 ml-2">
                      (
                      {tCommon('fileSize.mb', {
                        size: (version.file_size / 1024 / 1024).toFixed(2),
                      })}
                      )
                    </span>
                  )}
                </p>
              </div>
            )}
            {version.subject && (
              <div className="col-span-2">
                <p className="text-gray-500 dark:text-gray-400">{t('topic')}</p>
                <p className="text-gray-900 dark:text-white">{version.subject}</p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
