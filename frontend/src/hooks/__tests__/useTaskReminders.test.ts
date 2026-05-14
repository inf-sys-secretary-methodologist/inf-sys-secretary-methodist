import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import { useTaskReminders, createTaskReminder, deleteTaskReminder } from '../useTaskReminders'
import { apiClient } from '@/lib/api'
import type { TaskReminder } from '@/types/taskReminders'

jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
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

const sampleReminder: TaskReminder = {
  id: 1,
  task_id: 42,
  user_id: 7,
  reminder_type: 'telegram',
  minutes_before: 60,
  is_sent: false,
  sent_at: null,
  created_at: '2026-05-14T08:00:00Z',
}

beforeEach(() => {
  jest.clearAllMocks()
})

describe('useTaskReminders', () => {
  it('fetches /api/tasks/:id/reminders and returns reminders', async () => {
    mockedApiClient.get.mockResolvedValueOnce([sampleReminder])

    const { result } = renderHook(() => useTaskReminders(42), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.reminders).toEqual([sampleReminder])
    expect(result.current.error).toBeUndefined()
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/tasks/42/reminders')
  })

  it('short-circuits when taskID is null', () => {
    renderHook(() => useTaskReminders(null), { wrapper })
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })

  it('returns empty reminders array when response is empty', async () => {
    mockedApiClient.get.mockResolvedValueOnce([] as TaskReminder[])

    const { result } = renderHook(() => useTaskReminders(42), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.reminders).toEqual([])
  })

  it('exposes mutate handle for SWR cache invalidation', async () => {
    mockedApiClient.get.mockResolvedValueOnce([sampleReminder])

    const { result } = renderHook(() => useTaskReminders(42), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(typeof result.current.mutate).toBe('function')
  })
})

describe('createTaskReminder', () => {
  it('POSTs body to /api/tasks/:id/reminders and returns the created reminder', async () => {
    mockedApiClient.post.mockResolvedValueOnce(sampleReminder)

    const result = await createTaskReminder(42, {
      reminder_type: 'telegram',
      minutes_before: 60,
    })

    expect(result).toEqual(sampleReminder)
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/tasks/42/reminders', {
      reminder_type: 'telegram',
      minutes_before: 60,
    })
  })

  it('propagates axios errors to the caller', async () => {
    mockedApiClient.post.mockRejectedValueOnce(new Error('boom'))
    await expect(
      createTaskReminder(42, { reminder_type: 'telegram', minutes_before: 60 })
    ).rejects.toThrow('boom')
  })
})

describe('deleteTaskReminder', () => {
  it('DELETEs /api/tasks/:id/reminders/:reminderID', async () => {
    mockedApiClient.delete.mockResolvedValueOnce(undefined as never)

    await deleteTaskReminder(42, 1)

    expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/tasks/42/reminders/1')
  })

  it('propagates axios errors to the caller', async () => {
    mockedApiClient.delete.mockRejectedValueOnce(new Error('forbidden'))
    await expect(deleteTaskReminder(42, 1)).rejects.toThrow('forbidden')
  })
})
