import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useAssignments,
  useAssignment,
  useSubmissions,
  saveGrade,
} from '../useAssignments'
import { apiClient } from '@/lib/api'
import type {
  Assignment,
  AssignmentListResponse,
  SubmissionListResponse,
} from '@/types/assignments'

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

// Mirrors the existing useTasks test convention: apiClient.get/post
// is typed Promise<T>, and tests mock the resolved value as
// { data: <payload> } because the fetcher reads response.data and the
// hook keeps the unwrapped business object as its data.
const apiOk = <T,>(data: T) => ({ data })

const sampleAssignment: Assignment = {
  id: 10,
  title: 'Lab 1',
  description: 'doubly-linked list',
  teacher_id: 42,
  group_name: 'ИС-21',
  subject: 'Algorithms',
  max_score: 100,
  due_date: '2026-05-15T00:00:00Z',
  created_at: '2026-05-01T00:00:00Z',
  updated_at: '2026-05-01T00:00:00Z',
}

beforeEach(() => {
  jest.clearAllMocks()
})

describe('useAssignments', () => {
  it('returns items and total from /api/assignments', async () => {
    const list: AssignmentListResponse = { items: [sampleAssignment], total: 1 }
    mockedApiClient.get.mockResolvedValueOnce(apiOk(list))

    const { result } = renderHook(() => useAssignments(), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.items).toEqual([sampleAssignment])
    expect(result.current.total).toBe(1)
    expect(result.current.error).toBeUndefined()
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/assignments')
  })

  it('forwards filter as query string when provided', async () => {
    mockedApiClient.get.mockResolvedValueOnce(apiOk({ items: [], total: 0 }))

    renderHook(
      () => useAssignments({ subject: 'Algo', group_name: 'ИС-21', page_size: 25, offset: 50 }),
      { wrapper }
    )

    await waitFor(() => {
      expect(mockedApiClient.get).toHaveBeenCalled()
    })
    const call = mockedApiClient.get.mock.calls[0][0]
    expect(call).toContain('/api/assignments?')
    expect(call).toContain('subject=Algo')
    expect(call).toContain('group_name=%D0%98%D0%A1-21')
    expect(call).toContain('page_size=25')
    expect(call).toContain('offset=50')
  })

  it('surfaces API error', async () => {
    mockedApiClient.get.mockRejectedValueOnce(new Error('boom'))

    const { result } = renderHook(() => useAssignments(), { wrapper })
    await waitFor(() => expect(result.current.error).toBeDefined())
    expect(result.current.items).toEqual([])
  })
})

describe('useAssignment', () => {
  it('fetches a single assignment by id', async () => {
    mockedApiClient.get.mockResolvedValueOnce(apiOk(sampleAssignment))

    const { result } = renderHook(() => useAssignment(10), { wrapper })
    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(result.current.assignment).toEqual(sampleAssignment)
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/assignments/10')
  })

  it('does not fetch when id is null', () => {
    renderHook(() => useAssignment(null), { wrapper })
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })
})

describe('useSubmissions', () => {
  it('fetches submissions for an assignment', async () => {
    const list: SubmissionListResponse = {
      items: [
        {
          id: 1,
          assignment_id: 10,
          student_id: 7,
          student_name: 'Иван Петров',
          status: 'pending',
          created_at: '2026-05-01T00:00:00Z',
          updated_at: '2026-05-01T00:00:00Z',
        },
      ],
    }
    mockedApiClient.get.mockResolvedValueOnce(apiOk(list))

    const { result } = renderHook(() => useSubmissions(10), { wrapper })
    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.items).toEqual(list.items)
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/assignments/10/submissions')
  })

  it('appends status query when provided', async () => {
    mockedApiClient.get.mockResolvedValueOnce(apiOk({ items: [] }))
    renderHook(() => useSubmissions(10, 'graded'), { wrapper })
    await waitFor(() => {
      expect(mockedApiClient.get).toHaveBeenCalled()
    })
    expect(mockedApiClient.get).toHaveBeenCalledWith(
      '/api/assignments/10/submissions?status=graded'
    )
  })

  it('does not fetch when assignmentId is null', () => {
    renderHook(() => useSubmissions(null), { wrapper })
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })
})

describe('saveGrade', () => {
  it('POSTs to /api/assignments/:id/grades and returns response data', async () => {
    mockedApiClient.post.mockResolvedValueOnce(
      apiOk({ assignment_id: 10, student_id: 7, value: 85 })
    )

    const out = await saveGrade(10, { student_id: 7, value: 85, feedback: 'good' })
    expect(out).toEqual({ assignment_id: 10, student_id: 7, value: 85 })
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/assignments/10/grades', {
      student_id: 7,
      value: 85,
      feedback: 'good',
    })
  })

  it('throws on API error', async () => {
    mockedApiClient.post.mockRejectedValueOnce(new Error('409 already graded'))
    await expect(
      saveGrade(10, { student_id: 7, value: 85 })
    ).rejects.toThrow('409 already graded')
  })
})
