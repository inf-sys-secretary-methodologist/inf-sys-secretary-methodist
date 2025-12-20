'use client'

import { useState, useEffect, useCallback } from 'react'
import { useTranslations } from 'next-intl'
import { useAuthCheck, useAuth } from '@/hooks/useAuth'
import { UserRole } from '@/types/auth'
import { AppLayout } from '@/components/layout'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ArrowLeft, Share2, FileText, Users, Send } from 'lucide-react'
import Link from 'next/link'
import { DocumentList } from '@/components/documents/DocumentList'
import { DocumentPreview } from '@/components/documents/DocumentPreview'
import { Document, DocumentCategory, DocumentStatus } from '@/types/document'

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
import {
  documentsApi,
  DocumentInfo,
  MySharedDocumentOutput,
  SharedWithInfo,
  PermissionLevel,
} from '@/lib/api/documents'

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
const mapDocumentInfoToDocument = (doc: DocumentInfo, unknownLabel: string): Document => {
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
    tags: undefined,
    url: fileUrl,
    thumbnailUrl,
    metadata: {
      size: doc.file_size || 0,
      mimeType,
      uploadedBy: doc.author_name || unknownLabel,
      uploadedAt: new Date(doc.created_at),
    },
  }
}

// Permission keys for translation
const PERMISSION_KEYS: Record<PermissionLevel, string> = {
  read: 'permissionRead',
  write: 'permissionWrite',
  delete: 'permissionDelete',
  admin: 'permissionAdmin',
}

// Role keys for translation
const ROLE_KEYS: Record<string, string> = {
  admin: 'roleAdmin',
  secretary: 'roleSecretary',
  methodist: 'roleMethodist',
  teacher: 'roleTeacher',
  student: 'roleStudent',
}

export default function SharedDocumentsPage() {
  useAuthCheck()
  const { user } = useAuth()
  const canShare = user?.role !== UserRole.STUDENT
  const t = useTranslations('documents.sharedPage')
  const tShare = useTranslations('documents.share')

  const [sharedWithMe, setSharedWithMe] = useState<Document[]>([])
  const [mySharedDocs, setMySharedDocs] = useState<MySharedDocumentOutput[]>([])
  const [selectedDocument, setSelectedDocument] = useState<Document | null>(null)
  const [isLoadingSharedWithMe, setIsLoadingSharedWithMe] = useState(true)
  const [isLoadingMyShared, setIsLoadingMyShared] = useState(true)
  const [errorSharedWithMe, setErrorSharedWithMe] = useState<string | null>(null)
  const [errorMyShared, setErrorMyShared] = useState<string | null>(null)

  // Fetch shared documents (shared with me)
  const fetchSharedDocuments = useCallback(async () => {
    try {
      setIsLoadingSharedWithMe(true)
      setErrorSharedWithMe(null)
      const response = await documentsApi.getSharedDocuments()
      const mappedDocs = response.map((doc) => mapDocumentInfoToDocument(doc, t('unknown')))
      setSharedWithMe(mappedDocs)
    } catch (err) {
      console.error('Failed to fetch shared documents:', err)
      setErrorSharedWithMe(t('loadError'))
    } finally {
      setIsLoadingSharedWithMe(false)
    }
  }, [t])

  // Fetch my shared documents (documents I shared with others)
  const fetchMySharedDocuments = useCallback(async () => {
    try {
      setIsLoadingMyShared(true)
      setErrorMyShared(null)
      const response = await documentsApi.getMySharedDocuments()
      setMySharedDocs(response)
    } catch (err) {
      console.error('Failed to fetch my shared documents:', err)
      setErrorMyShared(t('loadMySharedError'))
    } finally {
      setIsLoadingMyShared(false)
    }
  }, [t])

  useEffect(() => {
    fetchSharedDocuments()
    if (canShare) {
      fetchMySharedDocuments()
    } else {
      setIsLoadingMyShared(false)
    }
  }, [fetchSharedDocuments, fetchMySharedDocuments, canShare])

  const handlePreview = (doc: Document) => {
    setSelectedDocument(doc)
  }

  const handleDownload = (doc: Document) => {
    const downloadUrl = documentsApi.getFileDownloadUrl(doc.id)
    const token = localStorage.getItem('authToken')
    if (token) {
      window.open(`${downloadUrl}?token=${token}`, '_blank')
    } else {
      window.open(downloadUrl, '_blank')
    }
  }

  // Render loading state
  const renderLoading = () => (
    <div className="flex items-center justify-center py-12">
      <div className="text-center space-y-4">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
        <p className="text-muted-foreground">{t('loading')}</p>
      </div>
    </div>
  )

  // Render error state
  const renderError = (error: string) => (
    <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
      <p className="text-red-700 dark:text-red-300">{error}</p>
    </div>
  )

  // Render empty state for "shared with me"
  const renderEmptySharedWithMe = () => (
    <div className="text-center py-12">
      <Share2 className="h-16 w-16 mx-auto text-gray-400 mb-4" />
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
        {t('noSharedWithMe')}
      </h3>
      <p className="text-gray-600 dark:text-gray-400">{t('noSharedWithMeDesc')}</p>
      <Link href="/documents" className="mt-4 inline-block">
        <Button variant="outline">
          <FileText className="h-4 w-4 mr-2" />
          {t('goToDocuments')}
        </Button>
      </Link>
    </div>
  )

  // Render empty state for "my shared"
  const renderEmptyMyShared = () => (
    <div className="text-center py-12">
      <Send className="h-16 w-16 mx-auto text-gray-400 mb-4" />
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
        {t('noMyShared')}
      </h3>
      <p className="text-gray-600 dark:text-gray-400">{t('noMySharedDesc')}</p>
      <Link href="/documents" className="mt-4 inline-block">
        <Button variant="outline">
          <FileText className="h-4 w-4 mr-2" />
          {t('goToDocuments')}
        </Button>
      </Link>
    </div>
  )

  // Render shared with info
  const renderSharedWithInfo = (sharedWith: SharedWithInfo) => {
    const roleKey = sharedWith.role ? ROLE_KEYS[sharedWith.role] : null
    const displayName =
      sharedWith.user_name || (roleKey ? `${t('role')}: ${tShare(roleKey)}` : t('unknown'))
    const expiresInfo = sharedWith.expires_at
      ? `→ ${new Date(sharedWith.expires_at).toLocaleDateString()}`
      : t('noExpiration')

    const permissionKey = PERMISSION_KEYS[sharedWith.permission]

    return (
      <div
        key={sharedWith.permission_id}
        className="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-800/50 rounded-lg text-sm"
      >
        <div className="flex items-center gap-2">
          <Users className="h-4 w-4 text-gray-500" />
          <span className="font-medium">{displayName}</span>
          {sharedWith.user_email && (
            <span className="text-gray-500 dark:text-gray-400">({sharedWith.user_email})</span>
          )}
        </div>
        <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400">
          <span className="px-2 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded text-xs">
            {tShare(permissionKey)}
          </span>
          <span className="text-xs">{expiresInfo}</span>
        </div>
      </div>
    )
  }

  // Helper to format sharing summary
  const formatSharingSummary = (sharedWith: SharedWithInfo[]): string => {
    const userShares = sharedWith.filter((s) => s.user_id != null)
    const roleShares = sharedWith.filter((s) => s.role != null && s.user_id == null)

    const parts: string[] = []

    if (userShares.length > 0) {
      const userWord = userShares.length === 1 ? t('user') : t('users')
      parts.push(`${userShares.length} ${userWord}`)
    }

    if (roleShares.length > 0) {
      const roleWord = roleShares.length === 1 ? t('roleSingular') : t('roles')
      parts.push(`${roleShares.length} ${roleWord}`)
    }

    return parts.length > 0 ? `${t('sharedWith')} ${parts.join(` ${t('and')} `)}` : t('noAccess')
  }

  // Render my shared documents list
  const renderMySharedList = () => (
    <div className="space-y-4">
      {mySharedDocs.map((doc) => (
        <div
          key={doc.document_id}
          className="p-4 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg space-y-3"
        >
          <div className="flex items-center justify-between">
            <Link
              href={`/documents/${doc.document_id}`}
              className="text-lg font-semibold text-gray-900 dark:text-white hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
            >
              {doc.document_title}
            </Link>
            <span className="text-sm text-gray-500 dark:text-gray-400">
              {formatSharingSummary(doc.shared_with)}
            </span>
          </div>
          <div className="space-y-2">{doc.shared_with.map(renderSharedWithInfo)}</div>
        </div>
      ))}
    </div>
  )

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

        {/* Back Button */}
        <div className="flex justify-end">
          <Link href="/documents">
            <Button variant="outline" className="flex items-center gap-2">
              <ArrowLeft className="h-4 w-4" />
              <span className="hidden sm:inline">{t('backToDocuments')}</span>
              <span className="sm:hidden">{t('back')}</span>
            </Button>
          </Link>
        </div>

        {/* Tabs */}
        <Tabs defaultValue="shared-with-me" className="w-full">
          {canShare ? (
            <TabsList className="grid w-full grid-cols-2 mb-6">
              <TabsTrigger value="shared-with-me" className="flex items-center gap-2">
                <Share2 className="h-4 w-4" />
                <span className="hidden sm:inline">{t('sharedWithMe')}</span>
                <span className="sm:hidden">{t('incoming')}</span>
              </TabsTrigger>
              <TabsTrigger value="my-shared" className="flex items-center gap-2">
                <Send className="h-4 w-4" />
                <span className="hidden sm:inline">{t('myShared')}</span>
                <span className="sm:hidden">{t('outgoing')}</span>
              </TabsTrigger>
            </TabsList>
          ) : null}

          {/* Shared With Me Tab */}
          <TabsContent value="shared-with-me">
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
                {errorSharedWithMe && renderError(errorSharedWithMe)}
                {isLoadingSharedWithMe ? (
                  renderLoading()
                ) : sharedWithMe.length === 0 ? (
                  renderEmptySharedWithMe()
                ) : (
                  <DocumentList
                    documents={sharedWithMe}
                    onPreview={handlePreview}
                    onDownload={handleDownload}
                  />
                )}
              </div>
            </div>
          </TabsContent>

          {/* My Shared Documents Tab - only for non-students */}
          {canShare && (
            <TabsContent value="my-shared">
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
                  {errorMyShared && renderError(errorMyShared)}
                  {isLoadingMyShared
                    ? renderLoading()
                    : mySharedDocs.length === 0
                      ? renderEmptyMyShared()
                      : renderMySharedList()}
                </div>
              </div>
            </TabsContent>
          )}
        </Tabs>
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
