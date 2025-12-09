import { apiClient } from '../api'

export interface FileInfo {
  id: string
  filename: string
  original_filename: string
  content_type: string
  size: number
  entity_type: string
  entity_id: string
  uploaded_by: string
  created_at: string
  updated_at: string
}

export interface UploadFileParams {
  file: File
  entity_type: 'document' | 'task' | 'announcement'
  entity_id: string
}

export interface UploadFileResponse {
  data: FileInfo
}

export interface FilesListResponse {
  data: FileInfo[]
}

export const filesApi = {
  /**
   * Upload a file and attach it to an entity (document, task, announcement)
   */
  async upload(params: UploadFileParams): Promise<FileInfo> {
    const formData = new FormData()
    formData.append('file', params.file)
    formData.append('entity_type', params.entity_type)
    formData.append('entity_id', params.entity_id)

    const response = await apiClient.post<UploadFileResponse>('/api/files/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  },

  /**
   * Get files by document ID
   */
  async getByDocument(documentId: string): Promise<FileInfo[]> {
    const response = await apiClient.get<FilesListResponse>(`/api/files/by-document/${documentId}`)
    return response.data
  },

  /**
   * Get files by task ID
   */
  async getByTask(taskId: string): Promise<FileInfo[]> {
    const response = await apiClient.get<FilesListResponse>(`/api/files/by-task/${taskId}`)
    return response.data
  },

  /**
   * Get files by announcement ID
   */
  async getByAnnouncement(announcementId: string): Promise<FileInfo[]> {
    const response = await apiClient.get<FilesListResponse>(
      `/api/files/by-announcement/${announcementId}`
    )
    return response.data
  },

  /**
   * Get file info by ID
   */
  async getById(fileId: string): Promise<FileInfo> {
    const response = await apiClient.get<{ data: FileInfo }>(`/api/files/${fileId}`)
    return response.data
  },

  /**
   * Download file - returns blob URL
   */
  getDownloadUrl(fileId: string): string {
    const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    return `${baseUrl}/api/files/${fileId}/download`
  },

  /**
   * Delete file by ID
   */
  async delete(fileId: string): Promise<void> {
    await apiClient.delete(`/api/files/${fileId}`)
  },
}
