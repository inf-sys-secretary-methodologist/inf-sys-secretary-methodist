import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import { useDisciplineItems, fetchDisciplineItem } from '../useDisciplineItems'
import { apiClient } from '@/lib/api'
import type { DisciplineItem, DisciplineItemListResponse } from '@/types/disciplineItem'

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

const sampleItem: DisciplineItem = {
  id: 202,
  section_id: 101,
  title: 'Математический анализ',
  hours_lectures: 36,
  hours_practice: 36,
  hours_lab: 0,
  hours_self: 72,
  control_form: 'exam',
  credits: 4,
  semester: 1,
  order_index: 0,
  version: 5,
  created_at: '2026-05-09T08:00:00Z',
  updated_at: '2026-05-09T08:00:00Z',
}

beforeEach(() => {
  jest.clearAllMocks()
})

describe('useDisciplineItems', () => {
  it('fetches /api/sections/:id/items and returns items', async () => {
    const list: DisciplineItemListResponse = { items: [sampleItem] }
    mockedApiClient.get.mockResolvedValueOnce(apiOk(list))

    const { result } = renderHook(() => useDisciplineItems(101), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.items).toEqual([sampleItem])
    expect(result.current.error).toBeUndefined()
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/sections/101/items')
  })

  it('short-circuits when sectionID is null', () => {
    renderHook(() => useDisciplineItems(null), { wrapper })
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })

  it('short-circuits when opts.enabled=false', () => {
    renderHook(() => useDisciplineItems(101, { enabled: false }), { wrapper })
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })

  it('returns empty items array when response data has empty list', async () => {
    mockedApiClient.get.mockResolvedValueOnce(apiOk({ items: [] } as DisciplineItemListResponse))

    const { result } = renderHook(() => useDisciplineItems(101), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.items).toEqual([])
  })

  it('exposes mutate handle for SWR cache invalidation', async () => {
    mockedApiClient.get.mockResolvedValueOnce(apiOk({ items: [sampleItem] }))

    const { result } = renderHook(() => useDisciplineItems(101), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(typeof result.current.mutate).toBe('function')
  })
})

describe('fetchDisciplineItem', () => {
  it('GETs /api/items/:id and returns the unwrapped DisciplineItem', async () => {
    mockedApiClient.get.mockResolvedValueOnce(apiOk(sampleItem))

    const result = await fetchDisciplineItem(202)

    expect(result).toEqual(sampleItem)
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/items/202')
  })

  it('propagates axios errors to caller', async () => {
    mockedApiClient.get.mockRejectedValueOnce(new Error('boom'))
    await expect(fetchDisciplineItem(202)).rejects.toThrow('boom')
  })

  it('returns the post-conflict refreshed item even when version differs', async () => {
    // 409 VERSION_CONFLICT recovery flow per plan ADR-12: client
    // refetches outside the failed tx; backend returns the current
    // server-truth version (which differs from the version client tried
    // to update with). The hook does not interpret version drift —
    // caller (BulkEditTable conflict resolver) compares.
    const refreshed: DisciplineItem = { ...sampleItem, version: 7, title: 'Renamed by other user' }
    mockedApiClient.get.mockResolvedValueOnce(apiOk(refreshed))

    const result = await fetchDisciplineItem(202)

    expect(result.version).toBe(7)
    expect(result.title).toBe('Renamed by other user')
  })
})
