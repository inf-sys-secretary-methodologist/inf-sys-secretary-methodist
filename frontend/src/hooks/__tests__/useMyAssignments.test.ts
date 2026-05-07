import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import { useMyAssignments, useMyAssignment, resubmitSubmission } from '../useMyAssignments'
import { apiClient } from '@/lib/api'
import type { StudentAssignmentView, MyAssignmentListResponse } from '@/types/assignments'

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

const apiOk = <T>(data: T) => ({ data })

const sampleView: StudentAssignmentView = {
  assignment_id: 10,
  title: 'Lab 1',
  description: 'Solve A',
  subject: 'Math',
  group_name: 'БСБО-01-22',
  max_score: 100,
  due_date: '2026-05-15T00:00:00Z',
  assignment_created_at: '2026-05-01T00:00:00Z',
  assignment_updated_at: '2026-05-01T00:00:00Z',
  submission_id: 1,
  student_id: 7,
  status: 'pending',
  feedback: '',
  return_reason: '',
  submission_created_at: '2026-05-01T00:00:00Z',
  submission_updated_at: '2026-05-01T00:00:00Z',
}

beforeEach(() => {
  jest.clearAllMocks()
})

describe('useMyAssignments', () => {
  it('returns items and total from /api/assignments/my', async () => {
    const list: MyAssignmentListResponse = { items: [sampleView], total: 1 }
    mockedApiClient.get.mockResolvedValueOnce(apiOk(list))

    const { result } = renderHook(() => useMyAssignments(), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.items).toEqual([sampleView])
    expect(result.current.total).toBe(1)
    expect(result.current.error).toBeUndefined()
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/assignments/my')
  })

  it('appends status query when provided', async () => {
    mockedApiClient.get.mockResolvedValueOnce(apiOk({ items: [], total: 0 }))

    renderHook(() => useMyAssignments('returned'), { wrapper })

    await waitFor(() => {
      expect(mockedApiClient.get).toHaveBeenCalled()
    })
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/assignments/my?status=returned')
  })

  it('surfaces API error and keeps items empty', async () => {
    mockedApiClient.get.mockRejectedValueOnce(new Error('boom'))

    const { result } = renderHook(() => useMyAssignments(), { wrapper })
    await waitFor(() => expect(result.current.error).toBeDefined())
    expect(result.current.items).toEqual([])
    expect(result.current.total).toBe(0)
  })
})

describe('useMyAssignment', () => {
  it('fetches a single my-assignment view by id', async () => {
    mockedApiClient.get.mockResolvedValueOnce(apiOk(sampleView))

    const { result } = renderHook(() => useMyAssignment(10), { wrapper })
    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(result.current.view).toEqual(sampleView)
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/assignments/10/my')
  })

  it('does not fetch when id is null (gates SWR until id is known)', () => {
    renderHook(() => useMyAssignment(null), { wrapper })
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })

  it('surfaces API error so the page can render an error state', async () => {
    mockedApiClient.get.mockRejectedValueOnce(new Error('404 not found'))

    const { result } = renderHook(() => useMyAssignment(10), { wrapper })
    await waitFor(() => expect(result.current.error).toBeDefined())
    expect(result.current.view).toBeUndefined()
  })
})

describe('resubmitSubmission', () => {
  it('POSTs to /api/assignments/:id/resubmit with empty body', async () => {
    mockedApiClient.post.mockResolvedValueOnce(apiOk({ assignment_id: 10, student_id: 7 }))

    const out = await resubmitSubmission(10)

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/assignments/10/resubmit', {})
    expect(out).toEqual({ assignment_id: 10, student_id: 7 })
  })

  it('propagates axios errors so callers can map status codes (409 not_returned, 403)', async () => {
    const err = new Error('409 not in returned state')
    mockedApiClient.post.mockRejectedValueOnce(err)

    await expect(resubmitSubmission(10)).rejects.toThrow('409 not in returned state')
  })
})
