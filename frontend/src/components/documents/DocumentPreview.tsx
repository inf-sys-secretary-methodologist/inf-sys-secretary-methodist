'use client'

import { useEffect, useRef, useState } from 'react'
import { X, Download, ExternalLink, FileText, History } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Document, DocumentCategoryLabels, DocumentStatusLabels } from '@/types/document'
import { DocumentVersionHistory } from './DocumentVersionHistory'

type TabType = 'preview' | 'versions'

interface DocumentPreviewProps {
  document: Document
  onClose: () => void
  onDownload?: () => void
  onDocumentUpdated?: () => void
  className?: string
}

export function DocumentPreview({
  document: doc,
  onClose,
  onDownload,
  onDocumentUpdated,
  className = '',
}: DocumentPreviewProps) {
  const modalRef = useRef<HTMLDivElement>(null)
  const [activeTab, setActiveTab] = useState<TabType>('preview')

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose()
      }
    }

    const handleClickOutside = (e: MouseEvent) => {
      if (modalRef.current && !modalRef.current.contains(e.target as Node)) {
        onClose()
      }
    }

    window.document.addEventListener('keydown', handleEscape)
    window.document.addEventListener('mousedown', handleClickOutside)

    return () => {
      window.document.removeEventListener('keydown', handleEscape)
      window.document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [onClose])

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

  const isPDF = doc.metadata.mimeType === 'application/pdf'
  const isImage = doc.metadata.mimeType.startsWith('image/')

  // Helper to add auth token to URL for file access
  // inline=true tells backend to use Content-Disposition: inline for preview
  const getAuthenticatedUrl = (url: string, inline: boolean = false) => {
    const token = localStorage.getItem('authToken')
    const params = new URLSearchParams()
    if (token) params.set('token', token)
    if (inline) params.set('inline', 'true')
    const queryString = params.toString()
    return queryString ? `${url}?${queryString}` : url
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div
        ref={modalRef}
        className={`
          relative w-full max-w-6xl max-h-[90vh] m-4
          bg-white dark:bg-gray-900 rounded-lg shadow-2xl
          flex flex-col overflow-hidden
          ${className}
        `}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex-1 min-w-0 pr-4">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white truncate">
              {doc.name}
            </h2>
            <div className="flex items-center gap-3 mt-2 text-sm text-gray-600 dark:text-gray-400">
              <span>{DocumentCategoryLabels[doc.category]}</span>
              <span>•</span>
              <span>{formatFileSize(doc.metadata.size)}</span>
              <span>•</span>
              <span>{formatDate(doc.metadata.uploadedAt)}</span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {onDownload && (
              <Button variant="outline" size="sm" onClick={onDownload}>
                <Download className="h-4 w-4 mr-2" />
                Скачать
              </Button>
            )}
            {doc.url && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => window.open(getAuthenticatedUrl(doc.url!, true), '_blank')}
              >
                <ExternalLink className="h-4 w-4 mr-2" />
                Открыть
              </Button>
            )}
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
            >
              <X className="h-5 w-5 text-gray-500" />
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-gray-200 dark:border-gray-700 px-4">
          <button
            onClick={() => setActiveTab('preview')}
            className={`
              px-4 py-3 text-sm font-medium border-b-2 transition-colors
              ${
                activeTab === 'preview'
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'
              }
            `}
          >
            <FileText className="h-4 w-4 inline-block mr-2" />
            Просмотр
          </button>
          <button
            onClick={() => setActiveTab('versions')}
            className={`
              px-4 py-3 text-sm font-medium border-b-2 transition-colors
              ${
                activeTab === 'versions'
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'
              }
            `}
          >
            <History className="h-4 w-4 inline-block mr-2" />
            История версий
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto p-4">
          {activeTab === 'preview' ? (
            <>
              {isPDF && doc.url ? (
                <div className="h-[500px]">
                  <iframe
                    src={getAuthenticatedUrl(doc.url, true)}
                    className="w-full h-full border-0 rounded bg-white"
                    title={doc.name}
                  />
                </div>
              ) : isImage && doc.url ? (
                <div className="flex items-center justify-center">
                  <img
                    src={getAuthenticatedUrl(doc.url, true)}
                    alt={doc.name}
                    className="max-w-[600px] max-h-[400px] object-contain rounded shadow-lg"
                  />
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center h-full min-h-[400px] text-center">
                  <FileText className="h-24 w-24 text-gray-400 mb-4" />
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                    Предварительный просмотр недоступен
                  </h3>
                  <p className="text-gray-600 dark:text-gray-400 mb-4">
                    Этот тип файла не поддерживает предварительный просмотр.
                  </p>
                  {onDownload && (
                    <Button onClick={onDownload}>
                      <Download className="h-4 w-4 mr-2" />
                      Скачать для просмотра
                    </Button>
                  )}
                </div>
              )}
            </>
          ) : (
            <DocumentVersionHistory
              documentId={Number(doc.id)}
              currentVersion={doc.metadata.version}
              onVersionRestored={onDocumentUpdated}
            />
          )}
        </div>

        {/* Footer with metadata */}
        <div className="border-t border-gray-200 dark:border-gray-700 p-4 bg-gray-50 dark:bg-gray-800/50">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div>
              <p className="text-gray-500 dark:text-gray-400 mb-1">Статус</p>
              <p className="font-medium text-gray-900 dark:text-white">
                {DocumentStatusLabels[doc.status]}
              </p>
            </div>
            <div>
              <p className="text-gray-500 dark:text-gray-400 mb-1">Загрузил</p>
              <p className="font-medium text-gray-900 dark:text-white">{doc.metadata.uploadedBy}</p>
            </div>
            {doc.metadata.version && (
              <div>
                <p className="text-gray-500 dark:text-gray-400 mb-1">Версия</p>
                <p className="font-medium text-gray-900 dark:text-white">{doc.metadata.version}</p>
              </div>
            )}
            {doc.metadata.modifiedAt && (
              <div>
                <p className="text-gray-500 dark:text-gray-400 mb-1">Изменен</p>
                <p className="font-medium text-gray-900 dark:text-white">
                  {formatDate(doc.metadata.modifiedAt)}
                </p>
              </div>
            )}
          </div>

          {doc.description && (
            <div className="mt-4">
              <p className="text-gray-500 dark:text-gray-400 mb-1">Описание</p>
              <p className="text-gray-900 dark:text-white">{doc.description}</p>
            </div>
          )}

          {doc.tags && doc.tags.length > 0 && (
            <div className="mt-4">
              <p className="text-gray-500 dark:text-gray-400 mb-2">Теги</p>
              <div className="flex flex-wrap gap-2">
                {doc.tags.map((tag, idx) => (
                  <span
                    key={idx}
                    className="px-3 py-1 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-full text-sm"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
