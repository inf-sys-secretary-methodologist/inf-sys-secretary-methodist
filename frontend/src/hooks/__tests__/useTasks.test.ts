import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import { useTasks, useTask, createTask, updateTask, deleteTask } from '../useTasks'
import { apiClient } from '@/lib/api'

jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(
    SWRConfig,
    { value: { dedupingInterval: 0, provider: () => new Map() } },
    children
  )

describe('useTasks hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('useTasks', () => {
    it('returns tasks list from API', async () => {
      const mockResponse = {
        tasks: [
          {
            id: 1,
            title: 'Task A',
            status: 'new',
            priority: 'normal',
            author_id: 1,
            progress: 0,
            is_overdue: false,
            created_at: '2026-04-25T10:00:00Z',
            updated_at: '2026-04-25T10:00:00Z',
          },
          {
            id: 2,
            title: 'Task B',
            status: 'in_progress',
            priority: 'high',
            author_id: 1,
            progress: 50,
            is_overdue: false,
            created_at: '2026-04-25T10:00:00Z',
            updated_at: '2026-04-25T10:00:00Z',
          },
        ],
        total: 2,
        limit: 20,
        offset: 0,
      }

      mockedApiClient.get.mockResolvedValue({ data: mockResponse })

      const { result } = renderHook(() => useTasks(), { wrapper })

      await waitFor(() => {
        expect(result.current.tasks).toHaveLength(2)
      })

      expect(result.current.total).toBe(2)
      expect(result.current.tasks[0].title).toBe('Task A')
    })

    it('passes filter parameters as query string', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { tasks: [], total: 0, limit: 20, offset: 0 },
      })

      renderHook(() => useTasks({ status: 'in_progress', priority: 'high', limit: 10 }), {
        wrapper,
      })

      await waitFor(() => {
        const calledUrl = mockedApiClient.get.mock.calls[0]?.[0] as string
        expect(calledUrl).toContain('status=in_progress')
        expect(calledUrl).toContain('priority=high')
        expect(calledUrl).toContain('limit=10')
      })
    })

    it('returns empty array when API returns null data', async () => {
      mockedApiClient.get.mockResolvedValue({ data: null })

      const { result } = renderHook(() => useTasks(), { wrapper })

      expect(result.current.tasks).toEqual([])
      expect(result.current.total).toBe(0)
    })
  })

  describe('useTask', () => {
    it('fetches a single task by id', async () => {
      const mockTask = {
        id: 42,
        title: 'Detailed task',
        status: 'assigned',
        priority: 'high',
        author_id: 1,
        progress: 25,
        is_overdue: false,
        created_at: '2026-04-25T10:00:00Z',
        updated_at: '2026-04-25T10:00:00Z',
      }

      mockedApiClient.get.mockResolvedValue({ data: mockTask })

      const { result } = renderHook(() => useTask(42), { wrapper })

      await waitFor(() => {
        expect(result.current.task?.id).toBe(42)
      })

      expect(result.current.task?.title).toBe('Detailed task')
    })

    it('does not fetch when id is null', () => {
      renderHook(() => useTask(null), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('mutations', () => {
    it('createTask sends POST with input', async () => {
      const created = { id: 99, title: 'New', status: 'new' }
      mockedApiClient.post.mockResolvedValue(created)

      const result = await createTask({ title: 'New', priority: 'normal' })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/tasks', {
        title: 'New',
        priority: 'normal',
      })
      expect(result).toEqual(created)
    })

    it('updateTask sends PUT to /api/tasks/:id', async () => {
      const updated = { id: 5, title: 'Updated', status: 'in_progress' }
      mockedApiClient.put.mockResolvedValue(updated)

      const result = await updateTask(5, { title: 'Updated' })

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/tasks/5', { title: 'Updated' })
      expect(result).toEqual(updated)
    })

    it('deleteTask sends DELETE to /api/tasks/:id', async () => {
      mockedApiClient.delete.mockResolvedValue(undefined)

      await deleteTask(7)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/tasks/7')
    })
  })
})
