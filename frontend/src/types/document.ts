/**
 * Document Types and Interfaces
 *
 * Defines the structure for documents in the system including metadata,
 * categories, and upload states.
 */

export enum DocumentCategory {
  EDUCATIONAL = 'educational', // Учебная деятельность - ID: 1
  HR = 'hr', // Кадровые вопросы - ID: 2
  ADMINISTRATIVE = 'administrative', // Административные - ID: 3
  METHODICAL = 'methodical', // Методическая работа - ID: 4
  FINANCIAL = 'financial', // Финансовые - ID: 5
  ARCHIVE = 'archive', // Архив - ID: 6
}

// Mapping frontend category to backend category_id
export const DocumentCategoryToId: Record<DocumentCategory, number> = {
  [DocumentCategory.EDUCATIONAL]: 1,
  [DocumentCategory.HR]: 2,
  [DocumentCategory.ADMINISTRATIVE]: 3,
  [DocumentCategory.METHODICAL]: 4,
  [DocumentCategory.FINANCIAL]: 5,
  [DocumentCategory.ARCHIVE]: 6,
}

export enum DocumentStatus {
  // Upload statuses
  UPLOADING = 'uploading',
  PROCESSING = 'processing',
  READY = 'ready',
  ERROR = 'error',
  // Workflow statuses (from backend)
  DRAFT = 'draft',
  REGISTERED = 'registered',
  ROUTING = 'routing',
  APPROVAL = 'approval',
  APPROVED = 'approved',
  REJECTED = 'rejected',
  EXECUTION = 'execution',
  EXECUTED = 'executed',
  ARCHIVED = 'archived',
}

export interface DocumentMetadata {
  size: number
  mimeType: string
  uploadedBy: string
  uploadedAt: Date
  modifiedAt?: Date
  version?: number
}

export interface Document {
  id: string
  name: string
  category: DocumentCategory
  status: DocumentStatus
  metadata: DocumentMetadata
  url?: string
  thumbnailUrl?: string
  description?: string
  tags?: string[]
  authorId?: number
}

export interface DocumentUpload {
  file: File
  category: DocumentCategory
  description?: string
  tags?: string[]
}

export interface DocumentFilter {
  category?: DocumentCategory
  status?: DocumentStatus
  search?: string
  tags?: string[]
  dateFrom?: Date
  dateTo?: Date
  authorId?: number
}

export interface DocumentSortOptions {
  field: 'name' | 'uploadedAt' | 'modifiedAt' | 'size'
  order: 'asc' | 'desc'
}

// Labels are now provided via i18n (messages/*.json)
// Use t('documents.categories.educational'), t('documents.statuses.uploading'), etc.

// Allowed file types for upload
export const ALLOWED_FILE_TYPES = [
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/vnd.ms-excel',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  'text/plain',
  'image/jpeg',
  'image/png',
]

export const ALLOWED_FILE_EXTENSIONS = [
  '.pdf',
  '.doc',
  '.docx',
  '.xls',
  '.xlsx',
  '.txt',
  '.jpg',
  '.jpeg',
  '.png',
]

// Maximum file size in bytes (10MB)
export const MAX_FILE_SIZE = 10 * 1024 * 1024
