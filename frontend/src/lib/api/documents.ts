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

  // ============== Sharing API ==============

  /**
   * Share document with a user or role
   */
  async shareDocument(
    documentId: number | string,
    params: ShareDocumentParams
  ): Promise<PermissionInfo> {
    const response = await apiClient.post<PermissionResponse>(
      `/api/documents/${documentId}/share`,
      params
    )
    return response.data
  },

  /**
   * Get document permissions
   */
  async getPermissions(documentId: number | string): Promise<PermissionInfo[]> {
    const response = await apiClient.get<PermissionsListResponse>(
      `/api/documents/${documentId}/permissions`
    )
    return response.data
  },

  /**
   * Revoke permission
   */
  async revokePermission(documentId: number | string, permissionId: number): Promise<void> {
    await apiClient.delete(`/api/documents/${documentId}/permissions/${permissionId}`)
  },

  /**
   * Create public link
   */
  async createPublicLink(
    documentId: number | string,
    params: CreatePublicLinkParams
  ): Promise<PublicLinkInfo> {
    const response = await apiClient.post<PublicLinkResponse>(
      `/api/documents/${documentId}/public-links`,
      params
    )
    return response.data
  },

  /**
   * Get document public links
   */
  async getPublicLinks(documentId: number | string): Promise<PublicLinkInfo[]> {
    const response = await apiClient.get<PublicLinksListResponse>(
      `/api/documents/${documentId}/public-links`
    )
    return response.data
  },

  /**
   * Deactivate public link
   */
  async deactivatePublicLink(documentId: number | string, linkId: number): Promise<void> {
    await apiClient.post(`/api/documents/${documentId}/public-links/${linkId}/deactivate`)
  },

  /**
   * Delete public link
   */
  async deletePublicLink(documentId: number | string, linkId: number): Promise<void> {
    await apiClient.delete(`/api/documents/${documentId}/public-links/${linkId}`)
  },

  /**
   * Get documents shared with current user
   */
  async getSharedDocuments(params?: {
    permission?: string
    limit?: number
    offset?: number
  }): Promise<DocumentInfo[]> {
    const response = await apiClient.get<DocumentsListResponse>('/api/documents/shared', { params })
    return response.data || []
  },

  /**
   * Get documents that current user has shared with others
   */
  async getMySharedDocuments(params?: {
    limit?: number
    offset?: number
  }): Promise<MySharedDocumentOutput[]> {
    const response = await apiClient.get<MySharedDocumentsResponse>('/api/documents/my-shared', {
      params,
    })
    return response.data || []
  },

  /**
   * Access document via public link (no auth required)
   */
  async accessPublicDocument(token: string, password?: string): Promise<PublicDocumentAccess> {
    const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    const response = await fetch(`${baseUrl}/api/public/documents/${token}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password }),
    })
    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.message || 'Failed to access document')
    }
    const result = await response.json()
    return result.data
  },
}

// ============== Sharing Types ==============

export type PermissionLevel = 'read' | 'write' | 'delete' | 'admin'
export type UserRole = 'admin' | 'secretary' | 'methodist' | 'teacher' | 'student'

export interface ShareDocumentParams {
  user_id?: number
  role?: UserRole
  permission: PermissionLevel
  expires_at?: string
}

export interface PermissionInfo {
  id: number
  document_id: number
  user_id?: number
  user_name?: string
  user_email?: string
  role?: string
  permission: PermissionLevel
  granted_by?: number
  granted_by_name?: string
  expires_at?: string
  created_at: string
}

export interface PermissionResponse {
  success: boolean
  data: PermissionInfo
}

export interface PermissionsListResponse {
  success: boolean
  data: PermissionInfo[]
}

export interface CreatePublicLinkParams {
  permission: 'read' | 'download'
  expires_at?: string
  max_uses?: number
  password?: string
}

export interface PublicLinkInfo {
  id: number
  document_id: number
  document_title?: string
  token: string
  url: string
  permission: 'read' | 'download'
  created_by: number
  created_by_name?: string
  expires_at?: string
  max_uses?: number
  use_count: number
  has_password: boolean
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface PublicLinkResponse {
  success: boolean
  data: PublicLinkInfo
}

export interface PublicLinksListResponse {
  success: boolean
  data: PublicLinkInfo[]
}

export interface PublicDocumentAccess {
  id: number
  title: string
  subject?: string
  content?: string
  author_name?: string
  registration_number?: string
  registration_date?: string
  file_name?: string
  file_size?: number
  mime_type?: string
  can_download: boolean
  created_at: string
}

// Types for "My Shared Documents" feature
export interface SharedWithInfo {
  permission_id: number
  user_id?: number
  user_name?: string
  user_email?: string
  role?: string
  permission: PermissionLevel
  granted_at: string
  expires_at?: string
}

export interface MySharedDocumentOutput {
  document_id: number
  document_title: string
  shared_with: SharedWithInfo[]
}

export interface MySharedDocumentsResponse {
  success: boolean
  data: MySharedDocumentOutput[]
}
