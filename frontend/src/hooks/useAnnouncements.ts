'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type {
  Announcement,
  AnnouncementAttachment,
  AnnouncementListResponse,
  AnnouncementFilterParams,
  CreateAnnouncementInput,
  UpdateAnnouncementInput,
} from '@/types/announcements'

const ANNOUNCEMENTS_BASE_URL = '/api/announcements'

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

// useAnnouncements fetches a paginated list of announcements with optional filters.
export function useAnnouncements(params?: AnnouncementFilterParams) {
  const searchParams = new URLSearchParams()
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value === undefined || value === null) return
      if (Array.isArray(value)) {
        value.forEach((v) => searchParams.append(key, String(v)))
      } else {
        searchParams.append(key, String(value))
      }
    })
  }

  const queryString = searchParams.toString()
  const url = queryString ? `${ANNOUNCEMENTS_BASE_URL}?${queryString}` : ANNOUNCEMENTS_BASE_URL

  const { data, error, isLoading, mutate } = useSWR<AnnouncementListResponse>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })

  return {
    announcements: data?.announcements || [],
    total: data?.total || 0,
    limit: data?.limit || 20,
    offset: data?.offset || 0,
    isLoading,
    error,
    mutate,
  }
}

// useAnnouncement fetches a single announcement by id. Pass null to skip.
export function useAnnouncement(id: number | null) {
  const { data, error, isLoading, mutate } = useSWR<Announcement>(
    id ? `${ANNOUNCEMENTS_BASE_URL}/${id}` : null,
    fetcher
  )

  return {
    announcement: data,
    isLoading,
    error,
    mutate,
  }
}

// Mutations

export async function createAnnouncement(
  input: CreateAnnouncementInput
): Promise<Announcement> {
  return apiClient.post<Announcement>(ANNOUNCEMENTS_BASE_URL, input)
}

export async function updateAnnouncement(
  id: number,
  input: UpdateAnnouncementInput
): Promise<Announcement> {
  return apiClient.put<Announcement>(`${ANNOUNCEMENTS_BASE_URL}/${id}`, input)
}

export async function deleteAnnouncement(id: number): Promise<void> {
  await apiClient.delete(`${ANNOUNCEMENTS_BASE_URL}/${id}`)
}

// Status actions

export async function publishAnnouncement(id: number): Promise<void> {
  await apiClient.post(`${ANNOUNCEMENTS_BASE_URL}/${id}/publish`)
}

export async function unpublishAnnouncement(id: number): Promise<void> {
  await apiClient.post(`${ANNOUNCEMENTS_BASE_URL}/${id}/unpublish`)
}

export async function archiveAnnouncement(id: number): Promise<void> {
  await apiClient.post(`${ANNOUNCEMENTS_BASE_URL}/${id}/archive`)
}

// Attachment mutations

export async function uploadAnnouncementAttachment(
  announcementId: number,
  file: File
): Promise<AnnouncementAttachment> {
  const formData = new FormData()
  formData.append('file', file)
  return apiClient.post<AnnouncementAttachment>(
    `${ANNOUNCEMENTS_BASE_URL}/${announcementId}/attachments`,
    formData,
    {
      headers: { 'Content-Type': 'multipart/form-data' },
    }
  )
}

export async function deleteAnnouncementAttachment(
  announcementId: number,
  attachmentId: number
): Promise<void> {
  await apiClient.delete(
    `${ANNOUNCEMENTS_BASE_URL}/${announcementId}/attachments/${attachmentId}`
  )
}
