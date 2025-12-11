'use client'

import { useState, useMemo, useEffect, useCallback, useRef } from 'react'
import { useAuthCheck } from '@/hooks/useAuth'
import { UserRole } from '@/types/auth'
import { AppLayout } from '@/components/layout'
import { GlowingEffect } from '@/components/ui/glowing-effect'
import { Button } from '@/components/ui/button'
import { Upload, FileText, Users } from 'lucide-react'
import Link from 'next/link'
import { DocumentUploadComponent } from '@/components/documents/DocumentUpload'
import { DocumentList } from '@/components/documents/DocumentList'
import { DocumentFilters } from '@/components/documents/DocumentFilters'
import { DocumentPreview } from '@/components/documents/DocumentPreview'
import { ShareDocumentDialog } from '@/components/documents/ShareDocumentDialog'
import { DocumentEditDialog } from '@/components/documents/DocumentEditDialog'
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
import { documentsApi, DocumentInfo, SearchResultItem } from '@/lib/api/documents'

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
    authorId: doc.author_id,
  }
}

// Extended Document type with search highlighting
type DocumentWithHighlighting = Document & {
  highlighted?: { title: string; subject: string; content: string }
  rank?: number
}

// Helper to convert search result to frontend Document with highlighting
const mapSearchResultToDocument = (result: SearchResultItem): DocumentWithHighlighting => {
  const doc = result.document
  const fileUrl = doc.has_file ? documentsApi.getFileDownloadUrl(doc.id) : undefined
  const mimeType = doc.mime_type || 'application/octet-stream'
  const isImage = mimeType.startsWith('image/')
  const thumbnailUrl = isImage && fileUrl ? fileUrl : undefined

  return {
    id: String(doc.id),
    name: doc.title,
    category: DocumentCategory.OTHER,
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
    authorId: doc.author_id,
    highlighted: {
      title: result.highlighted_title,
      subject: result.highlighted_subject,
      content: result.highlighted_content,
    },
    rank: result.rank,
  }
}

export default function DocumentsPage() {
  const { user } = useAuthCheck()
  const userCanEdit = canEdit(user?.role)
  const [showUpload, setShowUpload] = useState(false)
  const [selectedDocument, setSelectedDocument] = useState<Document | null>(null)
  const [sharingDocument, setSharingDocument] = useState<Document | null>(null)
  const [editingDocument, setEditingDocument] = useState<Document | null>(null)
  const [documents, setDocuments] = useState<Document[]>([])
  const [isUploading, setIsUploading] = useState(false)
  const [_isLoading, setIsLoading] = useState(true)
  const [_error, setError] = useState<string | null>(null)

  // Search state
  const [isSearching, setIsSearching] = useState(false)
  const [searchResults, setSearchResults] = useState<DocumentWithHighlighting[]>([])
  const [searchTotal, setSearchTotal] = useState(0)
  const searchTimeoutRef = useRef<NodeJS.Timeout | null>(null)

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

  // Debounced search effect
  useEffect(() => {
    // Clear previous timeout
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current)
    }

    const searchQuery = filters.search?.trim()

    if (!searchQuery || searchQuery.length < 2) {
      // Clear search results and show regular documents
      setSearchResults([])
      setSearchTotal(0)
      setIsSearching(false)
      return
    }

    // Set searching state immediately for UI feedback
    setIsSearching(true)

    // Debounce the actual search call
    searchTimeoutRef.current = setTimeout(async () => {
      try {
        const result = await documentsApi.search({
          q: searchQuery,
          page: 1,
          page_size: 50,
        })
        const mappedResults = result.results.map(mapSearchResultToDocument)
        setSearchResults(mappedResults)
        setSearchTotal(result.total)
      } catch (err) {
        console.error('Search failed:', err)
        setSearchResults([])
        setSearchTotal(0)
      } finally {
        setIsSearching(false)
      }
    }, 300)

    // Cleanup on unmount
    return () => {
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current)
      }
    }
  }, [filters.search])

  // Determine which documents to show
  const isInSearchMode = Boolean(filters.search?.trim() && filters.search.trim().length >= 2)

  // Apply filters and sorting (for non-search mode)
  const filteredAndSortedDocuments = useMemo(() => {
    if (isInSearchMode) {
      // In search mode, apply only non-search filters to search results
      const { search: _search, ...otherFilters } = filters
      const filtered = filterDocuments(searchResults, otherFilters)
      // Search results are already sorted by rank, but we can apply user's sort preference
      return sortDocuments(filtered, sort.field, sort.order)
    }
    // Regular mode - apply all filters to all documents
    const filtered = filterDocuments(documents, filters)
    return sortDocuments(filtered, sort.field, sort.order)
  }, [documents, searchResults, filters, sort, isInSearchMode])

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

  const handleShare = (doc: Document) => {
    setSharingDocument(doc)
  }

  const handleEdit = (doc: Document) => {
    setEditingDocument(doc)
  }

  const handleEditSaved = async () => {
    await fetchDocuments()
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

        {/* Action Buttons */}
        <div className="flex justify-end gap-2">
          <Link href="/documents/shared">
            <Button variant="outline" className="flex items-center gap-2">
              <Users className="h-4 w-4" />
              <span className="hidden sm:inline">Общие документы</span>
              <span className="sm:hidden">Общие</span>
            </Button>
          </Link>
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

        {/* Upload Section */}
        {showUpload ? (
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

            {/* Search Results Info */}
            {isInSearchMode && (
              <div className="flex items-center justify-between px-4 py-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-200 dark:border-blue-800">
                <div className="flex items-center gap-2">
                  {isSearching ? (
                    <>
                      <div className="animate-spin h-4 w-4 border-2 border-blue-500 border-t-transparent rounded-full" />
                      <span className="text-blue-700 dark:text-blue-300 text-sm">Поиск...</span>
                    </>
                  ) : (
                    <>
                      <span className="text-blue-700 dark:text-blue-300 text-sm">
                        Найдено: <strong>{searchTotal}</strong> документов по запросу &quot;
                        {filters.search}&quot;
                      </span>
                    </>
                  )}
                </div>
                {searchTotal > 0 && !isSearching && (
                  <span className="text-xs text-blue-600 dark:text-blue-400">
                    Результаты отсортированы по релевантности
                  </span>
                )}
              </div>
            )}

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
                  onDelete={handleDelete}
                  onShare={userCanEdit ? handleShare : undefined}
                  onEdit={handleEdit}
                  canDelete={(doc) =>
                    user?.role === UserRole.SYSTEM_ADMIN || doc.authorId === user?.id
                  }
                  canShare={() => userCanEdit}
                  canEdit={(doc) =>
                    user?.role === UserRole.SYSTEM_ADMIN || doc.authorId === user?.id
                  }
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

      {/* Share Document Dialog */}
      <ShareDocumentDialog
        open={sharingDocument !== null}
        onOpenChange={(open) => !open && setSharingDocument(null)}
        documentId={sharingDocument?.id ? Number(sharingDocument.id) : 0}
        documentTitle={sharingDocument?.name || ''}
      />

      {/* Edit Document Dialog */}
      <DocumentEditDialog
        document={editingDocument}
        open={editingDocument !== null}
        onOpenChange={(open) => !open && setEditingDocument(null)}
        onSaved={handleEditSaved}
      />
    </AppLayout>
  )
}
