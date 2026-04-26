export interface FileItem {
  id: number
  original_name: string
  size: number
  mime_type: string
  checksum: string
  uploaded_by: number
  document_id?: number
  task_id?: number
  announcement_id?: number
  is_temporary: boolean
  created_at: string
  updated_at: string
  download_url?: string
}

export interface FileVersion {
  id: number
  version_number: number
  size: number
  checksum: string
  comment?: string
  created_by: number
  created_at: string
  download_url?: string
}

export interface FileListResponse {
  files: FileItem[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface FileFilterParams {
  page?: number
  limit?: number
  mime_type?: string
  uploaded_by?: number
}

export interface UploadResponse {
  file_id: number
  original_name: string
  size: number
  mime_type: string
  checksum: string
}

export interface DownloadResponse {
  presigned_url: string
  file_name: string
  mime_type: string
  size: number
}
