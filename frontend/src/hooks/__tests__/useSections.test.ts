import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import { useSections } from '../useSections'
import { apiClient } from '@/lib/api'
import type { Section, SectionListResponse } from '@/types/section'

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

const sampleSection: Section = {
  id: 101,
  curriculum_id: 11,
  title: 'Базовая часть',
  description: 'Дисциплины обязательной части',
  order_index: 0,
  version: 0,
  created_at: '2026-05-09T08:00:00Z',
  updated_at: '2026-05-09T08:00:00Z',
}

beforeEach(() => {
  jest.clearAllMocks()
})

describe('useSections', () => {
  it('returns items from /api/curricula/:id/sections', async () => {
    const list: SectionListResponse = { items: [sampleSection] }
    mockedApiClient.get.mockResolvedValueOnce(apiOk(list))

    const { result } = renderHook(() => useSections(11), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.items).toEqual([sampleSection])
    expect(result.current.error).toBeUndefined()
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/curricula/11/sections')
  })

  it('short-circuits when curriculumID is null', () => {
    renderHook(() => useSections(null), { wrapper })
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })

  it('short-circuits when opts.enabled=false', () => {
    renderHook(() => useSections(11, { enabled: false }), { wrapper })
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })

  it('returns empty items array when response data has empty list', async () => {
    mockedApiClient.get.mockResolvedValueOnce(apiOk({ items: [] } as SectionListResponse))

    const { result } = renderHook(() => useSections(11), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.items).toEqual([])
  })

  it('exposes mutate handle for SWR cache invalidation', async () => {
    mockedApiClient.get.mockResolvedValueOnce(apiOk({ items: [sampleSection] }))

    const { result } = renderHook(() => useSections(11), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(typeof result.current.mutate).toBe('function')
  })
})
