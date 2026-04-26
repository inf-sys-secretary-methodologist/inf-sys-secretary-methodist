'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type {
  FileVersion,
  FileListResponse,
  FileFilterParams,
  UploadResponse,
  DownloadResponse,
} from '@/types/files'

const FILES_BASE_URL = '/api/files'

interface ApiResponse<T> {
  success: boolean
  data: T
  error?: { code: string; message: string }
  meta?: { request_id: string; timestamp: string; version: string }
}

const fetcher = async <T>(url: string): Promise<T> => {
  const response = await apiClient.get<ApiResponse<T>>(url)
  return response.data
}

export function useFiles(params?: FileFilterParams) {
  const searchParams = new URLSearchParams()
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value === undefined || value === null) return
      searchParams.append(key, String(value))
    })
  }

  const queryString = searchParams.toString()
  const url = queryString ? `${FILES_BASE_URL}?${queryString}` : FILES_BASE_URL

  const { data, error, isLoading, mutate } = useSWR<FileListResponse>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })

  return {
    files: data?.files || [],
    total: data?.total || 0,
    page: data?.page || 1,
    limit: data?.limit || 20,
    totalPages: data?.total_pages || 0,
    isLoading,
    error,
    mutate,
  }
}

export function useFileVersions(fileId: number | null) {
  const { data, error, isLoading, mutate } = useSWR<FileVersion[]>(
    fileId ? `${FILES_BASE_URL}/${fileId}/versions` : null,
    fetcher
  )

  return {
    versions: data || [],
    isLoading,
    error,
    mutate,
  }
}

export async function uploadFile(file: File): Promise<UploadResponse> {
  const formData = new FormData()
  formData.append('file', file)
  return apiClient.post<UploadResponse>(FILES_BASE_URL + '/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export async function deleteFile(id: number): Promise<void> {
  await apiClient.delete(`${FILES_BASE_URL}/${id}`)
}

export async function downloadFile(id: number): Promise<DownloadResponse> {
  return apiClient.get<DownloadResponse>(`${FILES_BASE_URL}/${id}/download`)
}

export async function createFileVersion(
  fileId: number,
  file: File,
  comment?: string
): Promise<FileVersion> {
  const formData = new FormData()
  formData.append('file', file)
  if (comment) {
    formData.append('comment', comment)
  }
  return apiClient.post<FileVersion>(`${FILES_BASE_URL}/${fileId}/versions`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
