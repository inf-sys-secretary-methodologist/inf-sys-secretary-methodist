'use client'

import { useState } from 'react'
import {
  FileText,
  Download,
  Trash2,
  Eye,
  Grid,
  List,
  FileImage,
  FileSpreadsheet,
  File,
  Share2,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Document,
  DocumentCategoryLabels,
  DocumentStatusLabels,
  DocumentStatus,
} from '@/types/document'

interface DocumentListProps {
  documents: Document[]
  onPreview?: (document: Document) => void
  onDownload?: (document: Document) => void
  onDelete?: (document: Document) => void
  onShare?: (document: Document) => void
  canDelete?: (document: Document) => boolean
  canShare?: (document: Document) => boolean
  isLoading?: boolean
  className?: string
}

type ViewMode = 'grid' | 'list'

// Helper to add auth token to URL for file access
const getAuthenticatedUrl = (url: string) => {
  const token = typeof window !== 'undefined' ? localStorage.getItem('authToken') : null
  return token ? `${url}?token=${token}&inline=true` : `${url}?inline=true`
}

// Get file type icon based on mime type
const getFileIcon = (mimeType: string, size: 'sm' | 'lg' = 'lg') => {
  const sizeClass = size === 'sm' ? 'h-8 w-8' : 'h-16 w-16'

  if (mimeType.startsWith('image/')) {
    return <FileImage className={`${sizeClass} text-blue-400`} />
  }
  if (mimeType === 'application/pdf') {
    return <FileText className={`${sizeClass} text-red-400`} />
  }
  if (
    mimeType.includes('spreadsheet') ||
    mimeType.includes('excel') ||
    mimeType === 'application/vnd.ms-excel'
  ) {
    return <FileSpreadsheet className={`${sizeClass} text-green-400`} />
  }
  if (mimeType.includes('word') || mimeType.includes('document')) {
    return <FileText className={`${sizeClass} text-blue-500`} />
  }
  return <File className={`${sizeClass} text-gray-400`} />
}

export function DocumentList({
  documents,
  onPreview,
  onDownload,
  onDelete,
  onShare,
  canDelete,
  canShare,
  isLoading = false,
  className = '',
}: DocumentListProps) {
  const [viewMode, setViewMode] = useState<ViewMode>('grid')

  const formatDate = (date: Date) => {
    return new Intl.DateTimeFormat('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(date))
  }

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} Б`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} КБ`
    return `${(bytes / 1024 / 1024).toFixed(2)} МБ`
  }

  const getStatusColor = (status: DocumentStatus) => {
    switch (status) {
      case DocumentStatus.READY:
        return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
      case DocumentStatus.UPLOADING:
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
      case DocumentStatus.PROCESSING:
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
      case DocumentStatus.ERROR:
        return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400'
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="text-muted-foreground">Загрузка документов...</p>
        </div>
      </div>
    )
  }

  if (documents.length === 0) {
    return (
      <div className="text-center py-12">
        <FileText className="h-16 w-16 mx-auto text-gray-400 mb-4" />
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">Нет документов</h3>
        <p className="text-gray-600 dark:text-gray-400">
          Загрузите документы, чтобы они появились здесь
        </p>
      </div>
    )
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {/* View Mode Toggle */}
      <div className="flex justify-end gap-2">
        <Button
          variant={viewMode === 'grid' ? 'default' : 'outline'}
          size="sm"
          onClick={() => setViewMode('grid')}
        >
          <Grid className="h-4 w-4" />
        </Button>
        <Button
          variant={viewMode === 'list' ? 'default' : 'outline'}
          size="sm"
          onClick={() => setViewMode('list')}
        >
          <List className="h-4 w-4" />
        </Button>
      </div>

      {/* Grid View */}
      {viewMode === 'grid' && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {documents.map((doc) => (
            <div
              key={doc.id}
              className="relative group border border-gray-200 dark:border-gray-700 rounded-lg p-4
                       bg-white dark:bg-gray-900 hover:shadow-lg transition-all"
            >
              {/* Document Icon/Thumbnail */}
              <div className="mb-4 flex items-center justify-center h-32 bg-gray-100 dark:bg-gray-800 rounded-lg overflow-hidden">
                {doc.thumbnailUrl ? (
                  <img
                    src={getAuthenticatedUrl(doc.thumbnailUrl)}
                    alt={doc.name}
                    className="max-h-full max-w-full object-contain"
                    onError={(e) => {
                      // Hide broken image and show icon instead
                      e.currentTarget.style.display = 'none'
                      e.currentTarget.nextElementSibling?.classList.remove('hidden')
                    }}
                  />
                ) : null}
                <div className={doc.thumbnailUrl ? 'hidden' : ''}>
                  {getFileIcon(doc.metadata.mimeType)}
                </div>
              </div>

              {/* Document Info */}
              <div className="space-y-2">
                <h4
                  className="font-semibold text-gray-900 dark:text-white truncate"
                  title={doc.name}
                >
                  {doc.name}
                </h4>

                <div className="flex items-center gap-2">
                  <span className={`text-xs px-2 py-1 rounded-full ${getStatusColor(doc.status)}`}>
                    {DocumentStatusLabels[doc.status]}
                  </span>
                  <span className="text-xs text-gray-500 dark:text-gray-400">
                    {DocumentCategoryLabels[doc.category]}
                  </span>
                </div>

                <div className="text-xs text-gray-500 dark:text-gray-400 space-y-1">
                  <p>Размер: {formatFileSize(doc.metadata.size)}</p>
                  <p>Загружено: {formatDate(doc.metadata.uploadedAt)}</p>
                </div>

                {doc.description && (
                  <p className="text-sm text-gray-600 dark:text-gray-300 line-clamp-2">
                    {doc.description}
                  </p>
                )}

                {doc.tags && doc.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {doc.tags.slice(0, 3).map((tag, idx) => (
                      <span
                        key={idx}
                        className="text-xs px-2 py-0.5 bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 rounded"
                      >
                        {tag}
                      </span>
                    ))}
                    {doc.tags.length > 3 && (
                      <span className="text-xs text-gray-500 dark:text-gray-400">
                        +{doc.tags.length - 3}
                      </span>
                    )}
                  </div>
                )}
              </div>

              {/* Actions */}
              <div className="mt-4 flex gap-2">
                {onPreview && doc.status === DocumentStatus.READY && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onPreview(doc)}
                    className="flex-1"
                  >
                    <Eye className="h-4 w-4 mr-1" />
                    Просмотр
                  </Button>
                )}
                {onShare && doc.status === DocumentStatus.READY && (!canShare || canShare(doc)) && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onShare(doc)}
                    title="Поделиться"
                  >
                    <Share2 className="h-4 w-4" />
                  </Button>
                )}
                {onDownload && doc.status === DocumentStatus.READY && (
                  <Button variant="outline" size="sm" onClick={() => onDownload(doc)}>
                    <Download className="h-4 w-4" />
                  </Button>
                )}
                {onDelete && (!canDelete || canDelete(doc)) && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onDelete(doc)}
                    className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950/20"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* List View */}
      {viewMode === 'list' && (
        <div className="space-y-2">
          {documents.map((doc) => (
            <div
              key={doc.id}
              className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700
                       rounded-lg bg-white dark:bg-gray-900 hover:shadow-md transition-all"
            >
              <div className="flex items-center gap-4 flex-1 min-w-0">
                {/* Icon */}
                <div className="flex-shrink-0 h-12 w-12 flex items-center justify-center bg-gray-100 dark:bg-gray-800 rounded overflow-hidden">
                  {doc.thumbnailUrl ? (
                    <img
                      src={getAuthenticatedUrl(doc.thumbnailUrl)}
                      alt={doc.name}
                      className="h-full w-full object-cover"
                      onError={(e) => {
                        e.currentTarget.style.display = 'none'
                        e.currentTarget.nextElementSibling?.classList.remove('hidden')
                      }}
                    />
                  ) : null}
                  <div
                    className={`flex items-center justify-center ${doc.thumbnailUrl ? 'hidden' : ''}`}
                  >
                    {getFileIcon(doc.metadata.mimeType, 'sm')}
                  </div>
                </div>

                {/* Info */}
                <div className="flex-1 min-w-0">
                  <h4 className="font-semibold text-gray-900 dark:text-white truncate">
                    {doc.name}
                  </h4>
                  <div className="flex items-center gap-3 mt-1">
                    <span
                      className={`text-xs px-2 py-0.5 rounded-full ${getStatusColor(doc.status)}`}
                    >
                      {DocumentStatusLabels[doc.status]}
                    </span>
                    <span className="text-xs text-gray-500 dark:text-gray-400">
                      {DocumentCategoryLabels[doc.category]}
                    </span>
                    <span className="text-xs text-gray-500 dark:text-gray-400">
                      {formatFileSize(doc.metadata.size)}
                    </span>
                    <span className="text-xs text-gray-500 dark:text-gray-400">
                      {formatDate(doc.metadata.uploadedAt)}
                    </span>
                  </div>
                  {doc.tags && doc.tags.length > 0 && (
                    <div className="flex gap-1 mt-2">
                      {doc.tags.slice(0, 5).map((tag, idx) => (
                        <span
                          key={idx}
                          className="text-xs px-2 py-0.5 bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 rounded"
                        >
                          {tag}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              {/* Actions */}
              <div className="flex items-center gap-2 flex-shrink-0">
                {onPreview && doc.status === DocumentStatus.READY && (
                  <Button variant="outline" size="sm" onClick={() => onPreview(doc)}>
                    <Eye className="h-4 w-4 mr-1" />
                    Просмотр
                  </Button>
                )}
                {onShare && doc.status === DocumentStatus.READY && (!canShare || canShare(doc)) && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onShare(doc)}
                    title="Поделиться"
                  >
                    <Share2 className="h-4 w-4" />
                  </Button>
                )}
                {onDownload && doc.status === DocumentStatus.READY && (
                  <Button variant="outline" size="sm" onClick={() => onDownload(doc)}>
                    <Download className="h-4 w-4" />
                  </Button>
                )}
                {onDelete && (!canDelete || canDelete(doc)) && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onDelete(doc)}
                    className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950/20"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
