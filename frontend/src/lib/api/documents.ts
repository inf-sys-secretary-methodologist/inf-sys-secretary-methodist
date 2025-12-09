import { apiClient } from '../api'

// Backend API response structure
export interface DocumentInfo {
  id: number
  document_type_id: number
  document_type_name?: string
  category_id?: number
  category_name?: string
  registration_number?: string
  registration_date?: string
  title: string
  subject?: string
  content?: string
  author_id: number
  author_name?: string
  recipient_id?: number
  recipient_name?: string
  status: string
  file_name?: string
  file_size?: number
  mime_type?: string
  has_file: boolean
  version: number
  deadline?: string
  execution_date?: string
  importance: string
  is_public: boolean
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface CreateDocumentParams {
  title: string
  document_type_id: number
  category_id?: number
  subject?: string
  content?: string
  recipient_id?: number
  deadline?: string
  importance?: 'low' | 'normal' | 'high' | 'urgent'
  is_public?: boolean
}

export interface UpdateDocumentParams {
  title?: string
  subject?: string
  content?: string
  category_id?: number
  recipient_id?: number
  deadline?: string
  importance?: 'low' | 'normal' | 'high' | 'urgent'
  is_public?: boolean
}

export interface DocumentsListResponse {
  success: boolean
  data: DocumentInfo[]
  meta: {
    timestamp: string
    pagination: {
      page: number
      per_page: number
      total: number
      total_pages: number
    }
  }
}

// Search types
export interface SearchParams {
  q: string
  document_type_id?: number
  category_id?: number
  author_id?: number
  status?: string
  importance?: string
  from_date?: string
  to_date?: string
  page?: number
  page_size?: number
}

export interface SearchResultItem {
  document: DocumentInfo
  rank: number
  highlighted_title: string
  highlighted_subject: string
  highlighted_content: string
}

export interface SearchOutput {
  results: SearchResultItem[]
  query: string
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface SearchResponse {
  success: boolean
  data: SearchOutput
  meta?: {
    timestamp: string
  }
}

export interface DocumentResponse {
  success: boolean
  data: DocumentInfo
  meta?: {
    timestamp: string
  }
}

export const documentsApi = {
  /**
   * Create a new document
   */
  async create(params: CreateDocumentParams): Promise<DocumentInfo> {
    const response = await apiClient.post<DocumentResponse>('/api/documents', params)
    return response.data
  },

  /**
   * Get list of documents
   */
  async list(params?: {
    page?: number
    page_size?: number
    category_id?: number
    status?: string
    search?: string
  }): Promise<DocumentsListResponse> {
    const response = await apiClient.get<DocumentsListResponse>('/api/documents', { params })
    return response
  },

  /**
   * Get document by ID
   */
  async getById(id: number | string): Promise<DocumentInfo> {
    const response = await apiClient.get<DocumentResponse>(`/api/documents/${id}`)
    return response.data
  },

  /**
   * Update document
   */
  async update(id: number | string, params: UpdateDocumentParams): Promise<DocumentInfo> {
    const response = await apiClient.put<DocumentResponse>(`/api/documents/${id}`, params)
    return response.data
  },

  /**
   * Delete document
   */
  async delete(id: number | string): Promise<void> {
    await apiClient.delete(`/api/documents/${id}`)
  },

  /**
   * Upload file to document
   */
  async uploadFile(documentId: number | string, file: File): Promise<DocumentInfo> {
    const formData = new FormData()
    formData.append('file', file)

    const response = await apiClient.post<DocumentResponse>(
      `/api/documents/${documentId}/file`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    )
    return response.data
  },

  /**
   * Get download URL for document file
   */
  getFileDownloadUrl(documentId: number | string): string {
    const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    return `${baseUrl}/api/documents/${documentId}/file`
  },

  /**
   * Delete file from document
   */
  async deleteFile(documentId: number | string): Promise<void> {
    await apiClient.delete(`/api/documents/${documentId}/file`)
  },

  /**
   * Full-text search documents
   */
  async search(params: SearchParams): Promise<SearchOutput> {
    const response = await apiClient.get<SearchResponse>('/api/documents/search', { params })
    return response.data
  },
}
