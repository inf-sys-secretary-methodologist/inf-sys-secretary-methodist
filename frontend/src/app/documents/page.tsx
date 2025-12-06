'use client'

import { useState, useMemo } from 'react'
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
  DocumentFilter,
  DocumentSortOptions,
  DocumentUpload,
  DocumentStatus,
} from '@/types/document'
import { mockDocuments, filterDocuments, sortDocuments } from '@/lib/mock-documents'

export default function DocumentsPage() {
  const { user } = useAuthCheck()
  const [showUpload, setShowUpload] = useState(false)
  const [selectedDocument, setSelectedDocument] = useState<Document | null>(null)
  const [documents, setDocuments] = useState<Document[]>(mockDocuments)
  const [isUploading, setIsUploading] = useState(false)

  // Filter and sort state
  const [filters, setFilters] = useState<DocumentFilter>({})
  const [sort, setSort] = useState<DocumentSortOptions>({
    field: 'uploadedAt',
    order: 'desc',
  })

  // Apply filters and sorting
  const filteredAndSortedDocuments = useMemo(() => {
    const filtered = filterDocuments(documents, filters)
    return sortDocuments(filtered, sort.field, sort.order)
  }, [documents, filters, sort])

  const handleUpload = async (uploads: DocumentUpload[]) => {
    setIsUploading(true)

    // Simulate upload process
    await new Promise((resolve) => setTimeout(resolve, 2000))

    // Create mock uploaded documents
    const newDocuments: Document[] = uploads.map((upload, index) => ({
      id: `new-${Date.now()}-${index}`,
      name: upload.file.name,
      category: upload.category,
      status: DocumentStatus.READY,
      metadata: {
        size: upload.file.size,
        mimeType: upload.file.type,
        uploadedBy: user?.name || 'Текущий пользователь',
        uploadedAt: new Date(),
      },
      description: upload.description,
      tags: upload.tags,
    }))

    setDocuments((prev) => [...newDocuments, ...prev])
    setIsUploading(false)
    setShowUpload(false)
  }

  const handlePreview = (doc: Document) => {
    setSelectedDocument(doc)
  }

  const handleDownload = (doc: Document) => {
    console.log('Downloading document:', doc.name)
    if (doc.url) {
      window.open(doc.url, '_blank')
    }
  }

  const handleDelete = async (doc: Document) => {
    if (confirm(`Вы уверены, что хотите удалить документ "${doc.name}"?`)) {
      setDocuments((prev) => prev.filter((d) => d.id !== doc.id))
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

        {/* Upload Button */}
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
