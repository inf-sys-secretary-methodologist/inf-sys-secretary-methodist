'use client'

import { useEffect, useRef, useState } from 'react'
import { useTranslations } from 'next-intl'
import Image from 'next/image'
import {
  X,
  Download,
  ExternalLink,
  FileText,
  History,
  Send,
  CheckCircle2,
  XCircle,
  FileSignature,
  Route,
  Stamp,
  UserCheck,
  CheckCheck,
  Archive,
  RotateCcw,
} from 'lucide-react'
import { getStoredToken } from '@/lib/auth/token'
import { Button } from '@/components/ui/button'
import { Document, DocumentStatus } from '@/types/document'
import { UserRole } from '@/types/auth'
import { useAuthCheck } from '@/hooks/useAuth'
import { DocumentVersionHistory } from './DocumentVersionHistory'
import { SubmitDocumentDialog } from './SubmitDocumentDialog'
import { ApproveDocumentDialog } from './ApproveDocumentDialog'
import { RejectDocumentDialog } from './RejectDocumentDialog'
import { RegisterDocumentDialog } from './RegisterDocumentDialog'
import { StartRoutingDialog } from './StartRoutingDialog'
import { SignVisaDialog } from './SignVisaDialog'
import { AssignExecutorDialog } from './AssignExecutorDialog'
import { MarkExecutedDialog } from './MarkExecutedDialog'
import { ArchiveDocumentDialog } from './ArchiveDocumentDialog'
import { ResubmitDocumentDialog } from './ResubmitDocumentDialog'

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
  const t = useTranslations('common')
  const tDocs = useTranslations('documents')
  const tPreview = useTranslations('documentPreview')
  const tWorkflow = useTranslations('documentsWorkflow')
  const { user } = useAuthCheck()
  const modalRef = useRef<HTMLDivElement>(null)
  const [activeTab, setActiveTab] = useState<TabType>('preview')
  const [submitOpen, setSubmitOpen] = useState(false)
  const [approveOpen, setApproveOpen] = useState(false)
  const [rejectOpen, setRejectOpen] = useState(false)
  const [registerOpen, setRegisterOpen] = useState(false)
  const [routingOpen, setRoutingOpen] = useState(false)
  const [signVisaOpen, setSignVisaOpen] = useState(false)
  const [assignExecutorOpen, setAssignExecutorOpen] = useState(false)
  const [markExecutedOpen, setMarkExecutedOpen] = useState(false)
  const [archiveOpen, setArchiveOpen] = useState(false)
  const [resubmitOpen, setResubmitOpen] = useState(false)

  // Workflow gates (v0.148.0 #227). Submit visible to author OR
  // any edit-role on a draft. Approve/Reject visible only к
  // academic_secretary + system_admin AND only when in approval queue.
  const role = user?.role
  const canSubmit =
    doc.status === DocumentStatus.DRAFT &&
    (role === UserRole.METHODIST ||
      role === UserRole.ACADEMIC_SECRETARY ||
      role === UserRole.SYSTEM_ADMIN ||
      role === UserRole.TEACHER)
  const canAdminApprove =
    doc.status === DocumentStatus.APPROVAL &&
    (role === UserRole.ACADEMIC_SECRETARY || role === UserRole.SYSTEM_ADMIN)
  // v0.149.0 Phase 2 — Register transition (#230). approved →
  // registered. Admin-only role gate.
  const canRegister =
    doc.status === DocumentStatus.APPROVED &&
    (role === UserRole.ACADEMIC_SECRETARY || role === UserRole.SYSTEM_ADMIN)
  // v0.150.0 Phase 3 — Routing transitions (#231). Single-step visa
  // per ADR-1: registered → routing → execution.
  const canStartRouting =
    doc.status === DocumentStatus.REGISTERED &&
    (role === UserRole.ACADEMIC_SECRETARY || role === UserRole.SYSTEM_ADMIN)
  const canSignVisa =
    doc.status === DocumentStatus.ROUTING &&
    (role === UserRole.ACADEMIC_SECRETARY || role === UserRole.SYSTEM_ADMIN)
  // v0.151.0 Phase 4 — Execution transitions (#232). AssignExecutor +
  // MarkExecuted both gated на execution + admin role.
  const canAssignExecutor =
    doc.status === DocumentStatus.EXECUTION &&
    (role === UserRole.ACADEMIC_SECRETARY || role === UserRole.SYSTEM_ADMIN)
  const canMarkExecuted =
    doc.status === DocumentStatus.EXECUTION &&
    (role === UserRole.ACADEMIC_SECRETARY || role === UserRole.SYSTEM_ADMIN)
  // v0.152.0 Phase 5 — Archive terminal transition (#233). executed →
  // archived gated на admin role only (ADR-1).
  const canArchive =
    doc.status === DocumentStatus.EXECUTED &&
    (role === UserRole.ACADEMIC_SECRETARY || role === UserRole.SYSTEM_ADMIN)
  // v0.152.0 Phase 5 — Resubmit rework cycle (#233). rejected → draft
  // gated по any edit-role role (mirror к canSubmit pattern); backend
  // usecase enforces author-or-edit-role authorization, mirror к Submit.
  const canResubmit =
    doc.status === DocumentStatus.REJECTED &&
    (role === UserRole.METHODIST ||
      role === UserRole.ACADEMIC_SECRETARY ||
      role === UserRole.SYSTEM_ADMIN ||
      role === UserRole.TEACHER)

  /* c8 ignore start - Keyboard and click handlers, tested in e2e */
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
  /* c8 ignore stop */

  /* c8 ignore start - Format and helper functions, tested in e2e */
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
    if (bytes < 1024) return t('fileSize.bytes', { size: bytes.toString() })
    if (bytes < 1024 * 1024) return t('fileSize.kb', { size: (bytes / 1024).toFixed(2) })
    return t('fileSize.mb', { size: (bytes / 1024 / 1024).toFixed(2) })
  }

  const isPDF = doc.metadata.mimeType === 'application/pdf'
  const isImage = doc.metadata.mimeType.startsWith('image/')

  // Helper to add auth token to URL for file access
  // inline=true tells backend to use Content-Disposition: inline for preview
  const getAuthenticatedUrl = (url: string, inline: boolean = false) => {
    const token = getStoredToken()
    const params = new URLSearchParams()
    if (token) params.set('token', token)
    if (inline) params.set('inline', 'true')
    const queryString = params.toString()
    return queryString ? `${url}?${queryString}` : url
  }
  /* c8 ignore stop */

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
              <span>{tDocs(`categories.${doc.category}`)}</span>
              <span>•</span>
              <span>{formatFileSize(doc.metadata.size)}</span>
              <span>•</span>
              <span>{formatDate(doc.metadata.uploadedAt)}</span>
            </div>
          </div>

          <div className="flex items-center gap-2 flex-wrap">
            {canSubmit && (
              <Button variant="default" size="sm" onClick={() => setSubmitOpen(true)}>
                <Send className="h-4 w-4 mr-2" />
                {tWorkflow('actions.submitButton')}
              </Button>
            )}
            {canAdminApprove && (
              <>
                <Button variant="default" size="sm" onClick={() => setApproveOpen(true)}>
                  <CheckCircle2 className="h-4 w-4 mr-2" />
                  {tWorkflow('actions.approveButton')}
                </Button>
                <Button variant="destructive" size="sm" onClick={() => setRejectOpen(true)}>
                  <XCircle className="h-4 w-4 mr-2" />
                  {tWorkflow('actions.rejectButton')}
                </Button>
              </>
            )}
            {canRegister && (
              <Button variant="default" size="sm" onClick={() => setRegisterOpen(true)}>
                <FileSignature className="h-4 w-4 mr-2" />
                {tWorkflow('actions.registerButton')}
              </Button>
            )}
            {canStartRouting && (
              <Button variant="default" size="sm" onClick={() => setRoutingOpen(true)}>
                <Route className="h-4 w-4 mr-2" />
                {tWorkflow('actions.routeButton')}
              </Button>
            )}
            {canSignVisa && (
              <Button variant="default" size="sm" onClick={() => setSignVisaOpen(true)}>
                <Stamp className="h-4 w-4 mr-2" />
                {tWorkflow('actions.signVisaButton')}
              </Button>
            )}
            {canAssignExecutor && (
              <Button variant="outline" size="sm" onClick={() => setAssignExecutorOpen(true)}>
                <UserCheck className="h-4 w-4 mr-2" />
                {tWorkflow('actions.assignExecutorButton')}
              </Button>
            )}
            {canMarkExecuted && (
              <Button variant="default" size="sm" onClick={() => setMarkExecutedOpen(true)}>
                <CheckCheck className="h-4 w-4 mr-2" />
                {tWorkflow('actions.markExecutedButton')}
              </Button>
            )}
            {canArchive && (
              <Button variant="destructive" size="sm" onClick={() => setArchiveOpen(true)}>
                <Archive className="h-4 w-4 mr-2" />
                {tWorkflow('actions.archiveButton')}
              </Button>
            )}
            {canResubmit && (
              <Button variant="default" size="sm" onClick={() => setResubmitOpen(true)}>
                <RotateCcw className="h-4 w-4 mr-2" />
                {tWorkflow('actions.resubmitButton')}
              </Button>
            )}
            {onDownload && (
              <Button variant="outline" size="sm" onClick={onDownload}>
                <Download className="h-4 w-4 mr-2" />
                {tPreview('download')}
              </Button>
            )}
            {doc.url && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => window.open(getAuthenticatedUrl(doc.url!, true), '_blank')}
              >
                <ExternalLink className="h-4 w-4 mr-2" />
                {tPreview('open')}
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

        {/* c8 ignore start - Tabs conditional styling */}
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
            {tPreview('viewTab')}
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
            {tPreview('historyTab')}
          </button>
        </div>
        {/* c8 ignore stop */}

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
                <div className="flex items-center justify-center relative w-full min-h-[400px]">
                  <Image
                    src={getAuthenticatedUrl(doc.url, true)}
                    alt={doc.name}
                    width={600}
                    height={400}
                    className="object-contain rounded shadow-lg"
                  />
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center h-full min-h-[400px] text-center">
                  <FileText className="h-24 w-24 text-gray-400 mb-4" />
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                    {tPreview('previewUnavailable')}
                  </h3>
                  <p className="text-gray-600 dark:text-gray-400 mb-4">
                    {tPreview('previewUnavailableHint')}
                  </p>
                  {onDownload && (
                    <Button onClick={onDownload}>
                      <Download className="h-4 w-4 mr-2" />
                      {tPreview('downloadToView')}
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
              <p className="text-gray-500 dark:text-gray-400 mb-1">{tDocs('filters.status')}</p>
              <p className="font-medium text-gray-900 dark:text-white">
                {tDocs(`statuses.${doc.status}`)}
              </p>
            </div>
            <div>
              <p className="text-gray-500 dark:text-gray-400 mb-1">{tDocs('filters.author')}</p>
              <p className="font-medium text-gray-900 dark:text-white">{doc.metadata.uploadedBy}</p>
            </div>
            {doc.metadata.version && (
              <div>
                <p className="text-gray-500 dark:text-gray-400 mb-1">{tPreview('version')}</p>
                <p className="font-medium text-gray-900 dark:text-white">{doc.metadata.version}</p>
              </div>
            )}
            {doc.metadata.modifiedAt && (
              <div>
                <p className="text-gray-500 dark:text-gray-400 mb-1">{tPreview('modified')}</p>
                <p className="font-medium text-gray-900 dark:text-white">
                  {formatDate(doc.metadata.modifiedAt)}
                </p>
              </div>
            )}
          </div>

          {doc.description && (
            <div className="mt-4">
              <p className="text-gray-500 dark:text-gray-400 mb-1">{tPreview('description')}</p>
              <p className="text-gray-900 dark:text-white">{doc.description}</p>
            </div>
          )}

          {/* c8 ignore start - Tags conditional rendering */}
          {doc.tags && doc.tags.length > 0 && (
            <div className="mt-4">
              <p className="text-gray-500 dark:text-gray-400 mb-2">{tPreview('tags')}</p>
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
          {/* c8 ignore stop */}
        </div>
      </div>
      <SubmitDocumentDialog
        documentId={Number(doc.id)}
        open={submitOpen}
        onClose={() => setSubmitOpen(false)}
        onSubmitted={onDocumentUpdated}
      />
      <ApproveDocumentDialog
        documentId={Number(doc.id)}
        open={approveOpen}
        onClose={() => setApproveOpen(false)}
        onApproved={onDocumentUpdated}
      />
      <RejectDocumentDialog
        documentId={Number(doc.id)}
        open={rejectOpen}
        onClose={() => setRejectOpen(false)}
        onRejected={onDocumentUpdated}
      />
      <RegisterDocumentDialog
        documentId={Number(doc.id)}
        open={registerOpen}
        onClose={() => setRegisterOpen(false)}
        onRegistered={onDocumentUpdated}
      />
      <StartRoutingDialog
        documentId={Number(doc.id)}
        open={routingOpen}
        onClose={() => setRoutingOpen(false)}
        onRouted={onDocumentUpdated}
      />
      <SignVisaDialog
        documentId={Number(doc.id)}
        open={signVisaOpen}
        onClose={() => setSignVisaOpen(false)}
        onSigned={onDocumentUpdated}
      />
      <AssignExecutorDialog
        documentId={Number(doc.id)}
        open={assignExecutorOpen}
        onClose={() => setAssignExecutorOpen(false)}
        onAssigned={onDocumentUpdated}
      />
      <MarkExecutedDialog
        documentId={Number(doc.id)}
        open={markExecutedOpen}
        onClose={() => setMarkExecutedOpen(false)}
        onMarked={onDocumentUpdated}
      />
      <ArchiveDocumentDialog
        documentId={Number(doc.id)}
        open={archiveOpen}
        onClose={() => setArchiveOpen(false)}
        onArchived={onDocumentUpdated}
      />
      <ResubmitDocumentDialog
        documentId={Number(doc.id)}
        open={resubmitOpen}
        onClose={() => setResubmitOpen(false)}
        onResubmitted={onDocumentUpdated}
      />
    </div>
  )
}
