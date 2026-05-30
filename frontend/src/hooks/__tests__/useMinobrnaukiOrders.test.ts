import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useMinobrnaukiOrders,
  useMinobrnaukiOrder,
  recordMinobrnaukiOrder,
  generateOrderRevisions,
  pickMinobrnaukiOrderErrorKey,
} from '../useMinobrnaukiOrders'
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

describe('useMinobrnaukiOrders hooks (queries)', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('useMinobrnaukiOrders (list)', () => {
    it('returns the orders page from API', async () => {
      const mockResponse = {
        items: [
          {
            id: 1,
            order_number: '№ 1234',
            title: 'О внесении изменений в ФГОС',
            published_at: '2026-03-01',
            change_scope: 'major',
            uploaded_by: 5,
            created_at: '2026-03-02T10:00:00Z',
          },
        ],
        total: 1,
      }
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockResponse })

      const { result } = renderHook(() => useMinobrnaukiOrders(), { wrapper })

      await waitFor(() => expect(result.current.isLoading).toBe(false))
      expect(result.current.items).toHaveLength(1)
      expect(result.current.total).toBe(1)
      expect(result.current.items[0].order_number).toBe('№ 1234')
      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v1/minobrnauki-orders')
    })

    it('appends only the set filter params to the query string', async () => {
      mockedApiClient.get.mockResolvedValue({ success: true, data: { items: [], total: 0 } })

      renderHook(
        () =>
          useMinobrnaukiOrders({ change_scope: 'minor', uploaded_by: 42, limit: 10, offset: 5 }),
        { wrapper }
      )

      await waitFor(() =>
        expect(mockedApiClient.get).toHaveBeenCalledWith(
          '/api/v1/minobrnauki-orders?change_scope=minor&uploaded_by=42&limit=10&offset=5'
        )
      )
    })

    it('does not fetch when enabled=false', async () => {
      renderHook(() => useMinobrnaukiOrders(undefined, { enabled: false }), { wrapper })
      // give SWR a tick — the key must stay null so no request fires
      await new Promise((r) => setTimeout(r, 10))
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('useMinobrnaukiOrder (detail)', () => {
    it('fetches one order with its affected work-program ids', async () => {
      const mockOrder = {
        id: 7,
        order_number: '№ 99',
        title: 'Приказ',
        published_at: '2026-01-15',
        change_scope: 'minor',
        uploaded_by: 3,
        created_at: '2026-01-16T08:00:00Z',
        affected_work_program_ids: [10, 11, 12],
      }
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockOrder })

      const { result } = renderHook(() => useMinobrnaukiOrder(7), { wrapper })

      await waitFor(() => expect(result.current.isLoading).toBe(false))
      expect(result.current.order?.affected_work_program_ids).toEqual([10, 11, 12])
      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v1/minobrnauki-orders/7')
    })

    it('does not fetch when id is null', async () => {
      renderHook(() => useMinobrnaukiOrder(null), { wrapper })
      await new Promise((r) => setTimeout(r, 10))
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })
})

describe('recordMinobrnaukiOrder (mutation)', () => {
  beforeEach(() => jest.clearAllMocks())

  it('POSTs the input and returns the created order', async () => {
    const created = {
      id: 9,
      order_number: '№ 555',
      title: 'Новый приказ',
      published_at: '2026-04-01',
      change_scope: 'major' as const,
      uploaded_by: 5,
      created_at: '2026-04-02T09:00:00Z',
      affected_work_program_ids: [1, 2],
    }
    mockedApiClient.post.mockResolvedValue({ success: true, data: created })

    const input = {
      order_number: '№ 555',
      title: 'Новый приказ',
      published_at: '2026-04-01',
      change_scope: 'major' as const,
      affected_work_program_ids: [1, 2],
    }
    const result = await recordMinobrnaukiOrder(input)

    expect(result.id).toBe(9)
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/minobrnauki-orders', input)
  })
})

describe('generateOrderRevisions (mutation)', () => {
  beforeEach(() => jest.clearAllMocks())

  it('POSTs to the generate-revisions endpoint and returns the run summary', async () => {
    const summary = { generated: 3, skipped: 1, failures: 0 }
    mockedApiClient.post.mockResolvedValue({ success: true, data: summary })

    const result = await generateOrderRevisions(7)

    expect(result).toEqual(summary)
    expect(mockedApiClient.post).toHaveBeenCalledWith(
      '/api/v1/minobrnauki-orders/7/generate-revisions'
    )
  })
})

describe('pickMinobrnaukiOrderErrorKey', () => {
  const codeErr = (code: string) => ({ response: { data: { error: { code } } } })
  const statusErr = (status: number) => ({ response: { status } })

  it.each([
    [codeErr('INVALID_MINOBRNAUKI_ORDER'), 'invalidOrder'],
    [codeErr('RATE_LIMITED'), 'rateLimited'],
    [statusErr(403), 'forbidden'],
    [statusErr(404), 'notFound'],
    [statusErr(400), 'invalidInput'],
    [undefined, 'generic'],
    [{}, 'generic'],
  ])('maps %o → %s', (err, expected) => {
    expect(pickMinobrnaukiOrderErrorKey(err)).toBe(expected)
  })
})
