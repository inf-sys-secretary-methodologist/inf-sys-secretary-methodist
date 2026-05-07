import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useCurricula,
  useCurriculum,
  updateCurriculum,
  submitCurriculum,
  approveCurriculum,
  rejectCurriculum,
} from '../useCurricula'
import { apiClient } from '@/lib/api'
import type {
  Curriculum,
  CurriculumListResponse,
  UpdateCurriculumRequest,
  RejectCurriculumRequest,
} from '@/types/curriculum'

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

const sampleCurriculum: Curriculum = {
  id: 11,
  title: 'ИВТ-2026 / 4 года',
  code: '09.03.04-2026',
  specialty: 'Информатика и вычислительная техника',
  year: 2026,
  description: 'Учебный план направления подготовки',
  status: 'draft',
  created_by: 5,
  created_at: '2026-05-06T08:00:00Z',
  updated_at: '2026-05-06T08:00:00Z',
}

beforeEach(() => {
  jest.clearAllMocks()
})

describe('useCurricula', () => {
  it('returns items and total from /api/curriculum', async () => {
    const list: CurriculumListResponse = { items: [sampleCurriculum], total: 1 }
    mockedApiClient.get.mockResolvedValueOnce(apiOk(list))

    const { result } = renderHook(() => useCurricula(), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.items).toEqual([sampleCurriculum])
    expect(result.current.total).toBe(1)
    expect(result.current.error).toBeUndefined()
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/curriculum')
  })

  it('forwards filter as query string when provided', async () => {
    mockedApiClient.get.mockResolvedValueOnce(apiOk({ items: [], total: 0 }))

    renderHook(
      () =>
        useCurricula({
          status: 'pending_approval',
          year: 2026,
          specialty: 'Информатика',
          created_by: 5,
          limit: 100,
          offset: 50,
        }),
      { wrapper }
    )

    await waitFor(() => {
      expect(mockedApiClient.get).toHaveBeenCalled()
    })
    const call = mockedApiClient.get.mock.calls[0][0] as string
    expect(call).toContain('/api/curriculum?')
    expect(call).toContain('status=pending_approval')
    expect(call).toContain('year=2026')
    // Cyrillic specialty is URL-encoded.
    expect(call).toContain(
      'specialty=%D0%98%D0%BD%D1%84%D0%BE%D1%80%D0%BC%D0%B0%D1%82%D0%B8%D0%BA%D0%B0'
    )
    expect(call).toContain('created_by=5')
    expect(call).toContain('limit=100')
    expect(call).toContain('offset=50')
  })

  it('does NOT fetch when enabled is false (skip 401 round-trip)', async () => {
    renderHook(() => useCurricula(undefined, { enabled: false }), { wrapper })
    // Give SWR a microtask to settle so a fetch — if it were going to
    // fire — would have done so by now.
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })

  it('still does NOT fetch when enabled=false even with a filter', async () => {
    renderHook(() => useCurricula({ status: 'draft' }, { enabled: false }), { wrapper })
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })

  it('surfaces API error', async () => {
    mockedApiClient.get.mockRejectedValueOnce(new Error('boom'))

    const { result } = renderHook(() => useCurricula(), { wrapper })
    await waitFor(() => expect(result.current.error).toBeDefined())
    expect(result.current.items).toEqual([])
    expect(result.current.total).toBe(0)
  })
})

describe('useCurriculum', () => {
  it('fetches a single curriculum by id', async () => {
    mockedApiClient.get.mockResolvedValueOnce(apiOk(sampleCurriculum))

    const { result } = renderHook(() => useCurriculum(11), { wrapper })
    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(result.current.curriculum).toEqual(sampleCurriculum)
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/curriculum/11')
  })

  it('does not fetch when id is null', async () => {
    renderHook(() => useCurriculum(null), { wrapper })
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })

  it('does not fetch when enabled=false', async () => {
    renderHook(() => useCurriculum(11, { enabled: false }), { wrapper })
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })
})

describe('updateCurriculum', () => {
  const body: UpdateCurriculumRequest = {
    title: 'ИВТ-2026 / 4 года (v2)',
    code: '09.03.04-2026',
    specialty: 'Информатика и вычислительная техника',
    year: 2026,
    description: 'Обновлённое описание',
  }

  it('PUTs to /api/curriculum/:id and returns the unwrapped Curriculum', async () => {
    const updated: Curriculum = { ...sampleCurriculum, ...body, updated_at: '2026-05-06T09:00:00Z' }
    mockedApiClient.put.mockResolvedValueOnce(apiOk(updated))

    const out = await updateCurriculum(11, body)

    expect(mockedApiClient.put).toHaveBeenCalledWith('/api/curriculum/11', body)
    expect(out).toEqual(updated)
  })

  it('propagates axios errors so callers can branch on status code', async () => {
    mockedApiClient.put.mockRejectedValueOnce(new Error('409 code exists'))
    await expect(updateCurriculum(11, body)).rejects.toThrow('409 code exists')
  })
})

describe('submitCurriculum', () => {
  it('POSTs to /api/curriculum/:id/submit with empty body and returns Curriculum', async () => {
    const submitted: Curriculum = { ...sampleCurriculum, status: 'pending_approval' }
    mockedApiClient.post.mockResolvedValueOnce(apiOk(submitted))

    const out = await submitCurriculum(11)

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/curriculum/11/submit', {})
    expect(out).toEqual(submitted)
    expect(out.status).toBe('pending_approval')
  })

  it('propagates axios errors so callers can branch on status code', async () => {
    mockedApiClient.post.mockRejectedValueOnce(new Error('422 not draft'))
    await expect(submitCurriculum(11)).rejects.toThrow('422 not draft')
  })
})

describe('approveCurriculum', () => {
  it('POSTs to /api/curriculum/:id/approve with empty body and returns approved Curriculum', async () => {
    const approved: Curriculum = {
      ...sampleCurriculum,
      status: 'approved',
      approved_by: 1,
      approved_at: '2026-05-07T08:00:00Z',
    }
    mockedApiClient.post.mockResolvedValueOnce(apiOk(approved))

    const out = await approveCurriculum(11)

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/curriculum/11/approve', {})
    expect(out).toEqual(approved)
    expect(out.status).toBe('approved')
  })

  it('propagates axios errors so callers can branch on status code', async () => {
    mockedApiClient.post.mockRejectedValueOnce(new Error('422 not pending'))
    await expect(approveCurriculum(11)).rejects.toThrow('422 not pending')
  })
})

describe('rejectCurriculum', () => {
  const body: RejectCurriculumRequest = { reason: 'Не соответствует ФГОС' }

  it('POSTs to /api/curriculum/:id/reject with body and returns Curriculum (back to draft)', async () => {
    const rejected: Curriculum = { ...sampleCurriculum, status: 'draft' }
    mockedApiClient.post.mockResolvedValueOnce(apiOk(rejected))

    const out = await rejectCurriculum(11, body)

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/curriculum/11/reject', body)
    expect(out).toEqual(rejected)
    expect(out.status).toBe('draft')
  })

  it('propagates axios errors so callers can branch on status code', async () => {
    mockedApiClient.post.mockRejectedValueOnce(new Error('422 not pending'))
    await expect(rejectCurriculum(11, body)).rejects.toThrow('422 not pending')
  })
})
