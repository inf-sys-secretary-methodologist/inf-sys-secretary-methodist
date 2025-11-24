'use client'

import { useState, useMemo } from 'react'
import { useAuthCheck } from '@/hooks/useAuth'
import { UserMenu } from '@/components/UserMenu'
import { ThemeToggleButton } from '@/components/theme-toggle-button'
import { GlowingEffect } from '@/components/ui/glowing-effect'
import { NavBar } from '@/components/ui/tubelight-navbar'
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
  DocumentStatus
} from '@/types/document'
import { mockDocuments, filterDocuments, sortDocuments } from '@/lib/mock-documents'
import { getAvailableNavItems } from '@/config/navigation'

export default function DocumentsPage() {
  const { user, isLoading } = useAuthCheck()
  const [showUpload, setShowUpload] = useState(false)
  const [selectedDocument, setSelectedDocument] = useState<Document | null>(null)
  const [documents, setDocuments] = useState<Document[]>(mockDocuments)
  const [isUploading, setIsUploading] = useState(false)

  // Filter and sort state
  const [filters, setFilters] = useState<DocumentFilter>({})
  const [sort, setSort] = useState<DocumentSortOptions>({
    field: 'uploadedAt',
    order: 'desc'
  })

  // Get navigation items filtered by user role
  const navItems = getAvailableNavItems(user?.role)

  // Apply filters and sorting
  const filteredAndSortedDocuments = useMemo(() => {
    const filtered = filterDocuments(documents, filters)
    return sortDocuments(filtered, sort.field, sort.order)
  }, [documents, filters, sort])

  const handleUpload = async (uploads: DocumentUpload[]) => {
    setIsUploading(true)

    // Simulate upload process
    // In production, this would make actual API calls
    await new Promise(resolve => setTimeout(resolve, 2000))

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
        uploadedAt: new Date()
      },
      description: upload.description,
      tags: upload.tags
    }))

    setDocuments(prev => [...newDocuments, ...prev])
    setIsUploading(false)
    setShowUpload(false)
  }

  const handlePreview = (doc: Document) => {
    setSelectedDocument(doc)
  }

  const handleDownload = (doc: Document) => {
    // In production, this would trigger actual download
    console.log('Downloading document:', doc.name)
    if (doc.url) {
      window.open(doc.url, '_blank')
    }
  }

  const handleDelete = async (doc: Document) => {
    if (confirm(`Вы уверены, что хотите удалить документ "${doc.name}"?`)) {
      // In production, this would make API call
      setDocuments(prev => prev.filter(d => d.id !== doc.id))
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="text-muted-foreground">Загрузка...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background p-8">
      {/* Navigation Bar */}
      <NavBar items={navItems} />

      {/* Top Navigation */}
      <div className="fixed top-8 right-8 z-50 pointer-events-auto flex items-center gap-3" style={{ isolation: 'isolate' }}>
        <UserMenu />
        <ThemeToggleButton />
      </div>

      <div className="max-w-7xl mx-auto space-y-8">
        {/* Page Header */}
        <div className="text-center space-y-4 pt-24">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white">
            Управление документами
          </h1>
          <p className="text-lg text-gray-600 dark:text-gray-300">
            Загрузка, поиск и управление документами
          </p>
        </div>

        {/* Upload Button */}
        <div className="flex justify-end">
          <Button
            onClick={() => setShowUpload(!showUpload)}
            className="flex items-center gap-2"
          >
            {showUpload ? (
              <>
                <FileText className="h-4 w-4" />
                Показать документы
              </>
            ) : (
              <>
                <Upload className="h-4 w-4" />
                Загрузить документы
              </>
            )}
          </Button>
        </div>

        {/* Upload Section */}
        {showUpload ? (
          <div className="relative overflow-hidden rounded-2xl p-8 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
            <GlowingEffect spread={40} glow={true} disabled={false} proximity={64} inactiveZone={0.01} borderWidth={3} />
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
            <div className="relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
              <GlowingEffect spread={40} glow={true} disabled={false} proximity={64} inactiveZone={0.01} borderWidth={3} />
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
            <div className="relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
              <GlowingEffect spread={40} glow={true} disabled={false} proximity={64} inactiveZone={0.01} borderWidth={3} />
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
    </div>
  )
}
