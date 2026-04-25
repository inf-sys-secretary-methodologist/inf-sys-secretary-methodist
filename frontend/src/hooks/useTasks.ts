'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type {
  Task,
  TaskListResponse,
  TaskFilterParams,
  CreateTaskInput,
  UpdateTaskInput,
} from '@/types/tasks'

const TASKS_BASE_URL = '/api/tasks'

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

// useTasks fetches a paginated list of tasks with optional filters.
export function useTasks(params?: TaskFilterParams) {
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
  const url = queryString ? `${TASKS_BASE_URL}?${queryString}` : TASKS_BASE_URL

  const { data, error, isLoading, mutate } = useSWR<TaskListResponse>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })

  return {
    tasks: data?.tasks || [],
    total: data?.total || 0,
    limit: data?.limit || 20,
    offset: data?.offset || 0,
    isLoading,
    error,
    mutate,
  }
}

// useTask fetches a single task by id. Pass null to skip the request.
export function useTask(id: number | null) {
  const { data, error, isLoading, mutate } = useSWR<Task>(
    id ? `${TASKS_BASE_URL}/${id}` : null,
    fetcher
  )

  return {
    task: data,
    isLoading,
    error,
    mutate,
  }
}

// Mutations

export async function createTask(input: CreateTaskInput): Promise<Task> {
  return apiClient.post<Task>(TASKS_BASE_URL, input)
}

export async function updateTask(id: number, input: UpdateTaskInput): Promise<Task> {
  return apiClient.put<Task>(`${TASKS_BASE_URL}/${id}`, input)
}

export async function deleteTask(id: number): Promise<void> {
  await apiClient.delete(`${TASKS_BASE_URL}/${id}`)
}
