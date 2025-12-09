'use client'

import { useState, useMemo, useEffect, useCallback } from 'react'
import { useAuthCheck } from '@/hooks/useAuth'
import { AppLayout } from '@/components/layout'
import { GlowingEffect } from '@/components/ui/glowing-effect'
import { Button } from '@/components/ui/button'
import { Upload, FileText } from 'lucide-react'
import { DocumentUploadComponent } from '@/components/documents/DocumentUpload'
import { DocumentList } from '@/components/documents/DocumentList'
import { DocumentFilters } from '@/components/documents/DocumentFilters'
import { DocumentPreview } from '@/components/documents/DocumentPreview'
import {
  Document,
  DocumentCategory,
  DocumentFilter,
  DocumentSortOptions,
  DocumentUpload,
  DocumentStatus,
} from '@/types/document'
import { filterDocuments, sortDocuments } from '@/lib/mock-documents'
import { canEdit } from '@/lib/auth/permissions'
import { documentsApi, DocumentInfo } from '@/lib/api/documents'

// Map backend status to frontend status
const mapBackendStatus = (status: string): DocumentStatus => {
  const statusMap: Record<string, DocumentStatus> = {
    draft: DocumentStatus.PROCESSING,
    registered: DocumentStatus.READY,
    routing: DocumentStatus.PROCESSING,
    approval: DocumentStatus.PROCESSING,
    approved: DocumentStatus.READY,
    rejected: DocumentStatus.ERROR,
    execution: DocumentStatus.PROCESSING,
    executed: DocumentStatus.READY,
    archived: DocumentStatus.READY,
  }
  return statusMap[status] || DocumentStatus.PROCESSING
}

// Helper to convert API DocumentInfo to frontend Document type
const mapDocumentInfoToDocument = (doc: DocumentInfo): Document => {
  const fileUrl = doc.has_file ? documentsApi.getFileDownloadUrl(doc.id) : undefined
  const mimeType = doc.mime_type || 'application/octet-stream'

  // For images, use the file URL as thumbnail
  const isImage = mimeType.startsWith('image/')
  const thumbnailUrl = isImage && fileUrl ? fileUrl : undefined

  return {
    id: String(doc.id),
    name: doc.title,
    category: DocumentCategory.OTHER, // Backend uses category_id, we default to OTHER
    status: mapBackendStatus(doc.status),
    description: doc.subject || undefined,
    tags: undefined,
    url: fileUrl,
    thumbnailUrl,
    metadata: {
      size: doc.file_size || 0,
      mimeType,
      uploadedBy: doc.author_name || 'Неизвестно',
      uploadedAt: new Date(doc.created_at),
    },
  }
}

export default function DocumentsPage() {
  const { user } = useAuthCheck()
  const userCanEdit = canEdit(user?.role)
  const [showUpload, setShowUpload] = useState(false)
  const [selectedDocument, setSelectedDocument] = useState<Document | null>(null)
  const [documents, setDocuments] = useState<Document[]>([])
  const [isUploading, setIsUploading] = useState(false)
  const [_isLoading, setIsLoading] = useState(true)
  const [_error, setError] = useState<string | null>(null)

  // Filter and sort state
  const [filters, setFilters] = useState<DocumentFilter>({})
  const [sort, setSort] = useState<DocumentSortOptions>({
    field: 'uploadedAt',
    order: 'desc',
  })

  // Fetch documents from API
  const fetchDocuments = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const response = await documentsApi.list()
      const mappedDocs = response.data.map(mapDocumentInfoToDocument)
      setDocuments(mappedDocs)
    } catch (err) {
      console.error('Failed to fetch documents:', err)
      setError('Не удалось загрузить документы')
    } finally {
      setIsLoading(false)
    }
  }, [])

  // Load documents on mount
  useEffect(() => {
    fetchDocuments()
  }, [fetchDocuments])

  // Apply filters and sorting
  const filteredAndSortedDocuments = useMemo(() => {
    const filtered = filterDocuments(documents, filters)
    return sortDocuments(filtered, sort.field, sort.order)
  }, [documents, filters, sort])

  const handleUpload = async (uploads: DocumentUpload[]) => {
    setIsUploading(true)
    setError(null)

    try {
      for (const upload of uploads) {
        // 1. Create document record
        const docInfo = await documentsApi.create({
          title: upload.file.name,
          document_type_id: 1, // Default document type (Входящий документ)
          subject: upload.description,
          importance: 'normal',
          is_public: false,
        })

        // 2. Upload file to the document
        await documentsApi.uploadFile(docInfo.id, upload.file)
      }

      // Refresh documents list
      await fetchDocuments()
      setShowUpload(false)
    } catch (err) {
      console.error('Failed to upload documents:', err)
      setError('Не удалось загрузить документы')
    } finally {
      setIsUploading(false)
    }
  }

  const handlePreview = (doc: Document) => {
    setSelectedDocument(doc)
  }

  const handleDownload = (doc: Document) => {
    const downloadUrl = documentsApi.getFileDownloadUrl(doc.id)
    // Open in new tab with auth token
    const token = localStorage.getItem('authToken')
    if (token) {
      window.open(`${downloadUrl}?token=${token}`, '_blank')
    } else {
      window.open(downloadUrl, '_blank')
    }
  }

  const handleDelete = async (doc: Document) => {
    if (confirm(`Вы уверены, что хотите удалить документ "${doc.name}"?`)) {
      try {
        await documentsApi.delete(doc.id)
        await fetchDocuments()
      } catch (err) {
        console.error('Failed to delete document:', err)
        setError('Не удалось удалить документ')
      }
    }
  }

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto space-y-6 sm:space-y-8">
        {/* Page Header */}
        <div className="text-center space-y-2 sm:space-y-4">
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-gray-900 dark:text-white">
            Управление документами
          </h1>
          <p className="text-base sm:text-lg text-gray-600 dark:text-gray-300">
            Загрузка, поиск и управление документами
          </p>
        </div>

        {/* Upload Button - only for users with edit permissions */}
        {userCanEdit && (
          <div className="flex justify-end">
            <Button onClick={() => setShowUpload(!showUpload)} className="flex items-center gap-2">
              {showUpload ? (
                <>
                  <FileText className="h-4 w-4" />
                  <span className="hidden sm:inline">Показать документы</span>
                  <span className="sm:hidden">Документы</span>
                </>
              ) : (
                <>
                  <Upload className="h-4 w-4" />
                  <span className="hidden sm:inline">Загрузить документы</span>
                  <span className="sm:hidden">Загрузить</span>
                </>
              )}
            </Button>
          </div>
        )}

        {/* Upload Section - only for users with edit permissions */}
        {showUpload && userCanEdit ? (
          <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 lg:p-8 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
            <GlowingEffect
              spread={40}
              glow={true}
              disabled={false}
              proximity={64}
              inactiveZone={0.01}
              borderWidth={3}
            />
            <div className="relative z-10">
              <DocumentUploadComponent
                onUpload={handleUpload}
                onCancel={() => setShowUpload(false)}
                isUploading={isUploading}
              />
            </div>
          </div>
        ) : (
          <>
            {/* Filters */}
            <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
              <GlowingEffect
                spread={40}
                glow={true}
                disabled={false}
                proximity={64}
                inactiveZone={0.01}
                borderWidth={3}
              />
              <div className="relative z-10">
                <DocumentFilters
                  onFilterChange={setFilters}
                  onSortChange={setSort}
                  currentFilters={filters}
                  currentSort={sort}
                />
              </div>
            </div>

            {/* Document List */}
            <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
              <GlowingEffect
                spread={40}
                glow={true}
                disabled={false}
                proximity={64}
                inactiveZone={0.01}
                borderWidth={3}
              />
              <div className="relative z-10">
                <DocumentList
                  documents={filteredAndSortedDocuments}
                  onPreview={handlePreview}
                  onDownload={handleDownload}
                  onDelete={userCanEdit ? handleDelete : undefined}
                />
              </div>
            </div>
          </>
        )}
      </div>

      {/* Document Preview Modal */}
      {selectedDocument && (
        <DocumentPreview
          document={selectedDocument}
          onClose={() => setSelectedDocument(null)}
          onDownload={() => handleDownload(selectedDocument)}
        />
      )}
    </AppLayout>
  )
}
