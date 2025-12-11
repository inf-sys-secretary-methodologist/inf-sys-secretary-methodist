/**
 * Document Types and Interfaces
 *
 * Defines the structure for documents in the system including metadata,
 * categories, and upload states.
 */

export enum DocumentCategory {
  SYLLABUS = 'syllabus',
  ATTENDANCE = 'attendance',
  GRADES = 'grades',
  REPORT = 'report',
  ASSIGNMENT = 'assignment',
  EXAM = 'exam',
  OTHER = 'other',
}

export enum DocumentStatus {
  UPLOADING = 'uploading',
  PROCESSING = 'processing',
  READY = 'ready',
  ERROR = 'error',
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
}

export interface DocumentSortOptions {
  field: 'name' | 'uploadedAt' | 'modifiedAt' | 'size'
  order: 'asc' | 'desc'
}

// Mapping for display purposes
export const DocumentCategoryLabels: Record<DocumentCategory, string> = {
  [DocumentCategory.SYLLABUS]: 'Учебный план',
  [DocumentCategory.ATTENDANCE]: 'Посещаемость',
  [DocumentCategory.GRADES]: 'Оценки',
  [DocumentCategory.REPORT]: 'Отчет',
  [DocumentCategory.ASSIGNMENT]: 'Задание',
  [DocumentCategory.EXAM]: 'Экзамен',
  [DocumentCategory.OTHER]: 'Другое',
}

export const DocumentStatusLabels: Record<DocumentStatus, string> = {
  [DocumentStatus.UPLOADING]: 'Загрузка',
  [DocumentStatus.PROCESSING]: 'Обработка',
  [DocumentStatus.READY]: 'Готов',
  [DocumentStatus.ERROR]: 'Ошибка',
}

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
