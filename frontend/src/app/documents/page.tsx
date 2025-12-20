'use client'

import { useState, useMemo, useEffect, useCallback, useRef } from 'react'
import { format } from 'date-fns'
import { useTranslations } from 'next-intl'
import { useAuthCheck } from '@/hooks/useAuth'
import { UserRole } from '@/types/auth'
import { AppLayout } from '@/components/layout'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
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
  DocumentCategoryToId,
  DocumentFilter,
  DocumentSortOptions,
  DocumentUpload,
  DocumentStatus,
} from '@/types/document'
import { filterDocuments, sortDocuments } from '@/lib/mock-documents'
import { canEdit } from '@/lib/auth/permissions'
import { documentsApi, DocumentInfo, SearchResultItem, tagsApi, TagInfo } from '@/lib/api/documents'

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

// Map backend category_id to frontend DocumentCategory
const mapCategoryIdToCategory = (categoryId?: number): DocumentCategory => {
  const categoryMap: Record<number, DocumentCategory> = {
    1: DocumentCategory.EDUCATIONAL,
    2: DocumentCategory.HR,
    3: DocumentCategory.ADMINISTRATIVE,
    4: DocumentCategory.METHODICAL,
    5: DocumentCategory.FINANCIAL,
    6: DocumentCategory.ARCHIVE,
  }
  return categoryMap[categoryId || 1] || DocumentCategory.EDUCATIONAL
}

// Helper to convert API DocumentInfo to frontend Document type
const mapDocumentInfoToDocument = (
  doc: DocumentInfo,
  tags?: TagInfo[],
  unknownAuthor: string = 'Unknown'
): Document => {
  const fileUrl = doc.has_file ? documentsApi.getFileDownloadUrl(doc.id) : undefined
  const mimeType = doc.mime_type || 'application/octet-stream'

  // For images, use the file URL as thumbnail
  const isImage = mimeType.startsWith('image/')
  const thumbnailUrl = isImage && fileUrl ? fileUrl : undefined

  return {
    id: String(doc.id),
    name: doc.title,
    category: mapCategoryIdToCategory(doc.category_id),
    status: mapBackendStatus(doc.status),
    description: doc.subject || undefined,
    tags: tags?.map((t) => t.name),
    url: fileUrl,
    thumbnailUrl,
    metadata: {
      size: doc.file_size || 0,
      mimeType,
      uploadedBy: doc.author_name || unknownAuthor,
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
const mapSearchResultToDocument = (
  result: SearchResultItem,
  tags?: TagInfo[],
  unknownAuthor: string = 'Unknown'
): DocumentWithHighlighting => {
  const doc = result.document
  const fileUrl = doc.has_file ? documentsApi.getFileDownloadUrl(doc.id) : undefined
  const mimeType = doc.mime_type || 'application/octet-stream'
  const isImage = mimeType.startsWith('image/')
  const thumbnailUrl = isImage && fileUrl ? fileUrl : undefined

  return {
    id: String(doc.id),
    name: doc.title,
    category: mapCategoryIdToCategory(doc.category_id),
    status: mapBackendStatus(doc.status),
    description: doc.subject || undefined,
    tags: tags?.map((t) => t.name),
    url: fileUrl,
    thumbnailUrl,
    metadata: {
      size: doc.file_size || 0,
      mimeType,
      uploadedBy: doc.author_name || unknownAuthor,
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
  const t = useTranslations('documents')
  const tCommon = useTranslations('common')
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

  // Fetch documents from API with filters
  const fetchDocuments = useCallback(
    async (currentFilters?: DocumentFilter) => {
      try {
        setIsLoading(true)
        setError(null)

        // Build API params from filters
        const params: Parameters<typeof documentsApi.list>[0] = {}

        if (currentFilters?.status) {
          params.status = currentFilters.status
        }
        if (currentFilters?.authorId) {
          params.author_id = currentFilters.authorId
        }
        if (currentFilters?.dateFrom) {
          params.from_date = format(currentFilters.dateFrom, 'yyyy-MM-dd')
        }
        if (currentFilters?.dateTo) {
          params.to_date = format(currentFilters.dateTo, 'yyyy-MM-dd')
        }

        const response = await documentsApi.list(params)

        // Fetch tags for all documents in parallel
        const tagsPromises = response.data.map((doc) =>
          tagsApi.getDocumentTags(doc.id).catch(() => [])
        )
        const allTags = await Promise.all(tagsPromises)

        // Map documents with their tags
        const unknownAuthor = tCommon('unknown')
        const mappedDocs = response.data.map((doc, index) =>
          mapDocumentInfoToDocument(doc, allTags[index], unknownAuthor)
        )
        setDocuments(mappedDocs)
      } catch (err) {
        console.error('Failed to fetch documents:', err)
        setError(t('loadFailed'))
      } finally {
        setIsLoading(false)
      }
    },
    [t, tCommon]
  )

  // Load documents on mount and when filters change (except search which is handled separately)
  useEffect(() => {
    // Skip if we're in search mode - search has its own effect
    const isSearchMode = filters.search && filters.search.trim().length >= 2
    if (!isSearchMode) {
      fetchDocuments(filters)
    }
  }, [fetchDocuments, filters.status, filters.authorId, filters.dateFrom, filters.dateTo])

  // Debounced search effect with filters
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
        const searchParams: Parameters<typeof documentsApi.search>[0] = {
          q: searchQuery,
          page: 1,
          page_size: 50,
        }

        // Add other filters to search
        if (filters.authorId) {
          searchParams.author_id = filters.authorId
        }
        if (filters.status) {
          searchParams.status = filters.status
        }
        if (filters.dateFrom) {
          searchParams.from_date = format(filters.dateFrom, 'yyyy-MM-dd')
        }
        if (filters.dateTo) {
          searchParams.to_date = format(filters.dateTo, 'yyyy-MM-dd')
        }

        const result = await documentsApi.search(searchParams)

        // Fetch tags for all search results in parallel
        const tagsPromises = result.results.map((r) =>
          tagsApi.getDocumentTags(r.document.id).catch(() => [])
        )
        const allTags = await Promise.all(tagsPromises)

        // Map results with their tags
        const unknownAuthor = tCommon('unknown')
        const mappedResults = result.results.map((r, index) =>
          mapSearchResultToDocument(r, allTags[index], unknownAuthor)
        )
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
  }, [filters.search, filters.authorId, filters.status, filters.dateFrom, filters.dateTo, tCommon])

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
        // 1. Create document record with proper category
        const categoryId = DocumentCategoryToId[upload.category] || 1
        const docInfo = await documentsApi.create({
          title: upload.file.name,
          document_type_id: 1, // Default document type (Входящий документ)
          category_id: categoryId,
          subject: upload.description,
          importance: 'normal',
          is_public: false,
        })

        // 2. Upload file to the document
        await documentsApi.uploadFile(docInfo.id, upload.file)

        // 3. Save tags if provided
        if (upload.tags && upload.tags.length > 0) {
          console.log('Upload tags:', upload.tags)
          // Get all available tags to match by name
          const availableTags = await tagsApi.getAll()
          console.log('Available tags:', availableTags)
          const tagIds: number[] = []

          for (const tagName of upload.tags) {
            const matchingTag = availableTags.find(
              (t) => t.name.toLowerCase() === tagName.toLowerCase()
            )
            console.log(`Matching tag "${tagName}":`, matchingTag)
            if (matchingTag) {
              tagIds.push(matchingTag.id)
            }
          }

          console.log('Tag IDs to save:', tagIds)
          // Add matched tags to document
          if (tagIds.length > 0) {
            console.log(`Saving tags to document ${docInfo.id}:`, tagIds)
            await tagsApi.setDocumentTags(docInfo.id, tagIds)
            console.log('Tags saved successfully')
          }
        } else {
          console.log('No tags provided in upload:', upload)
        }
      }

      // Refresh documents list
      await fetchDocuments()
      setShowUpload(false)
    } catch (err) {
      console.error('Failed to upload documents:', err)
      setError(t('uploadFailed'))
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
    if (confirm(t('confirmDelete', { name: doc.name }))) {
      try {
        await documentsApi.delete(doc.id)
        await fetchDocuments()
      } catch (err) {
        console.error('Failed to delete document:', err)
        setError(t('deleteFailed'))
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
            {t('title')}
          </h1>
          <p className="text-base sm:text-lg text-gray-600 dark:text-gray-300">{t('subtitle')}</p>
        </div>

        {/* Action Buttons */}
        <div className="flex justify-end gap-2">
          <Link href="/documents/shared">
            <Button variant="outline" className="flex items-center gap-2">
              <Users className="h-4 w-4" />
              <span className="hidden sm:inline">{t('sharedDocuments')}</span>
              <span className="sm:hidden">{t('sharedShort')}</span>
            </Button>
          </Link>
          <Button onClick={() => setShowUpload(!showUpload)} className="flex items-center gap-2">
            {showUpload ? (
              <>
                <FileText className="h-4 w-4" />
                <span className="hidden sm:inline">{t('showDocuments')}</span>
                <span className="sm:hidden">{t('documentsShort')}</span>
              </>
            ) : (
              <>
                <Upload className="h-4 w-4" />
                <span className="hidden sm:inline">{t('upload')}</span>
                <span className="sm:hidden">{t('uploadShort')}</span>
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
                      <span className="text-blue-700 dark:text-blue-300 text-sm">
                        {t('searching')}
                      </span>
                    </>
                  ) : (
                    <>
                      <span
                        className="text-blue-700 dark:text-blue-300 text-sm"
                        dangerouslySetInnerHTML={{
                          __html: t('searchResults', {
                            count: searchTotal,
                            query: filters.search || '',
                          }),
                        }}
                      />
                    </>
                  )}
                </div>
                {searchTotal > 0 && !isSearching && (
                  <span className="text-xs text-blue-600 dark:text-blue-400">
                    {t('sortedByRelevance')}
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
