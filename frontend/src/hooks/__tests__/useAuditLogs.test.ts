import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import { useAuditLogs } from '../useAuditLogs'
import { apiClient } from '@/lib/api'
import type { AuditLog, AuditLogPagination } from '@/types/audit'

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

const sampleLog: AuditLog = {
  id: 11,
  created_at: '2026-05-10T12:30:00Z',
  action: 'curriculum.approved',
  resource: 'curriculum',
  actor_user_id: 42,
  actor_ip: '10.0.0.5',
  correlation_id: 'req-7c4f',
  fields: { curriculum_id: 7 },
}

const samplePagination: AuditLogPagination = {
  page: 1,
  per_page: 50,
  total: 1,
  total_pages: 1,
}

// apiOk mirrors the curriculum/assignments hook test convention —
// apiClient.get is typed Promise<T>, tests mock the resolved value
// as the wire envelope shape so the hook's fetcher reads it through
// the same path production code does.
const apiOk = <T>(envelope: T) => envelope

beforeEach(() => {
  jest.clearAllMocks()
})

describe('useAuditLogs', () => {
  it('fetches /api/admin/audit-logs with no query string when filter omitted', async () => {
    mockedApiClient.get.mockResolvedValueOnce(
      apiOk({ success: true, data: [sampleLog], meta: { pagination: samplePagination } })
    )

    const { result } = renderHook(() => useAuditLogs(), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.items).toEqual([sampleLog])
    expect(result.current.total).toBe(1)
    expect(result.current.pagination).toEqual(samplePagination)
    expect(result.current.error).toBeUndefined()
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/admin/audit-logs')
  })

  it('forwards every filter dimension as the matching query param', async () => {
    mockedApiClient.get.mockResolvedValueOnce(
      apiOk({
        success: true,
        data: [],
        meta: { pagination: { page: 1, per_page: 25, total: 0, total_pages: 0 } },
      })
    )

    renderHook(
      () =>
        useAuditLogs({
          action: 'curriculum.approved',
          resource: 'curriculum',
          user_id: 42,
          from: '2026-05-01T00:00:00Z',
          to: '2026-05-31T00:00:00Z',
          limit: 25,
          offset: 50,
        }),
      { wrapper }
    )

    await waitFor(() => {
      expect(mockedApiClient.get).toHaveBeenCalled()
    })

    const call = mockedApiClient.get.mock.calls[0][0]
    expect(call).toContain('/api/admin/audit-logs?')
    expect(call).toContain('action=curriculum.approved')
    expect(call).toContain('resource=curriculum')
    expect(call).toContain('user_id=42')
    // RFC3339 colon encodes to %3A — but URLSearchParams keeps `:`
    // because it is a valid sub-delim — assert on the unencoded form.
    expect(call).toContain('from=2026-05-01T00%3A00%3A00Z')
    expect(call).toContain('to=2026-05-31T00%3A00%3A00Z')
    expect(call).toContain('limit=25')
    expect(call).toContain('offset=50')
  })

  it('omits filter params that are undefined', async () => {
    mockedApiClient.get.mockResolvedValueOnce(
      apiOk({ success: true, data: [], meta: { pagination: samplePagination } })
    )

    renderHook(() => useAuditLogs({ action: 'auth.login' }), { wrapper })

    await waitFor(() => {
      expect(mockedApiClient.get).toHaveBeenCalled()
    })
    const call = mockedApiClient.get.mock.calls[0][0]
    expect(call).toBe('/api/admin/audit-logs?action=auth.login')
    expect(call).not.toContain('resource=')
    expect(call).not.toContain('user_id=')
  })

  it('skips fetch entirely when opts.enabled is false', () => {
    renderHook(() => useAuditLogs({ action: 'x' }, { enabled: false }), { wrapper })
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })

  it('surfaces API error and yields empty items', async () => {
    mockedApiClient.get.mockRejectedValueOnce(new Error('403 forbidden'))

    const { result } = renderHook(() => useAuditLogs(), { wrapper })

    await waitFor(() => expect(result.current.error).toBeDefined())
    expect(result.current.items).toEqual([])
    expect(result.current.total).toBe(0)
  })

  it('returns total from meta.pagination, not from items.length', async () => {
    mockedApiClient.get.mockResolvedValueOnce(
      apiOk({
        success: true,
        data: [sampleLog],
        meta: { pagination: { page: 1, per_page: 50, total: 999, total_pages: 20 } },
      })
    )

    const { result } = renderHook(() => useAuditLogs(), { wrapper })
    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(result.current.items.length).toBe(1)
    expect(result.current.total).toBe(999)
    expect(result.current.pagination?.total_pages).toBe(20)
  })

  it('encodes filter values with special characters', async () => {
    mockedApiClient.get.mockResolvedValueOnce(
      apiOk({ success: true, data: [], meta: { pagination: samplePagination } })
    )

    renderHook(() => useAuditLogs({ resource: 'document space' }), { wrapper })

    await waitFor(() => {
      expect(mockedApiClient.get).toHaveBeenCalled()
    })
    const call = mockedApiClient.get.mock.calls[0][0]
    // URLSearchParams uses `+` for space in form-encoded query strings.
    expect(call).toContain('resource=document+space')
  })
})
