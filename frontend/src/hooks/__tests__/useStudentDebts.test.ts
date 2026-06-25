import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useStudentDebts,
  useStudentDebt,
  useMyStudentDebts,
  useDebtStats,
  scheduleResit,
  recordResitResult,
  pickStudentDebtErrorKey,
} from '../useStudentDebts'
import { apiClient } from '@/lib/api'

jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(
    SWRConfig,
    { value: { dedupingInterval: 0, provider: () => new Map() } },
    children
  )

describe('useStudentDebts hooks (queries)', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('useStudentDebts', () => {
    it('returns the debts registry page from API', async () => {
      const mockResponse = {
        items: [
          {
            id: 1,
            student_full_name: 'Иванов Иван',
            group_name: 'ИВТ-21',
            discipline_name: 'Базы данных',
            semester: 3,
            control_form: 'exam',
            status: 'open',
            version: 0,
          },
        ],
        total: 1,
      }
      mockedApiClient.get.mockResolvedValue({ data: mockResponse })

      const { result } = renderHook(() => useStudentDebts(), { wrapper })

      await waitFor(() => {
        expect(result.current.items).toHaveLength(1)
      })
      expect(result.current.total).toBe(1)
      expect(result.current.items[0].student_full_name).toBe('Иванов Иван')
    })

    it('passes filter params as a query string', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { items: [], total: 0 } })

      renderHook(
        () => useStudentDebts({ group_name: 'ИВТ-21', status: 'open', semester: 3, limit: 50 }),
        { wrapper }
      )

      await waitFor(() => {
        const url = mockedApiClient.get.mock.calls[0]?.[0] as string
        expect(url).toContain('group_name=')
        expect(url).toContain('status=open')
        expect(url).toContain('semester=3')
        expect(url).toContain('limit=50')
      })
    })

    it('omits undefined/empty filter params from the query string', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { items: [], total: 0 } })

      renderHook(
        () => useStudentDebts({ status: 'commission', group_name: '', semester: undefined }),
        {
          wrapper,
        }
      )

      await waitFor(() => {
        const url = mockedApiClient.get.mock.calls[0]?.[0] as string
        expect(url).toContain('status=commission')
        expect(url).not.toContain('group_name=')
        expect(url).not.toContain('semester=')
      })
    })

    it('returns empty array when API returns null data', () => {
      mockedApiClient.get.mockResolvedValue({ data: null })
      const { result } = renderHook(() => useStudentDebts(), { wrapper })
      expect(result.current.items).toEqual([])
      expect(result.current.total).toBe(0)
    })

    it('does not fetch when enabled is false', () => {
      renderHook(() => useStudentDebts(undefined, { enabled: false }), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('useStudentDebt', () => {
    it('fetches a single debt by id with its attempts', async () => {
      const mockDebt = {
        id: 7,
        student_full_name: 'Петров',
        status: 'resit_scheduled',
        attempts: [],
      }
      mockedApiClient.get.mockResolvedValue({ data: mockDebt })

      const { result } = renderHook(() => useStudentDebt(7), { wrapper })

      await waitFor(() => {
        expect(result.current.debt?.id).toBe(7)
      })
      expect(result.current.debt?.status).toBe('resit_scheduled')
    })

    it('hits the canonical detail URL', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { id: 99 } })
      renderHook(() => useStudentDebt(99), { wrapper })
      await waitFor(() => {
        const url = mockedApiClient.get.mock.calls[0]?.[0] as string
        expect(url).toBe('/api/student-debts/99')
      })
    })

    it('does not fetch when id is null', () => {
      renderHook(() => useStudentDebt(null), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })

    it('does not fetch when enabled is false even with a valid id', () => {
      renderHook(() => useStudentDebt(7, { enabled: false }), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('useMyStudentDebts', () => {
    it('hits the /my endpoint and returns the items', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { items: [{ id: 3 }], total: 1 } })
      const { result } = renderHook(() => useMyStudentDebts(), { wrapper })
      await waitFor(() => {
        expect(result.current.items).toHaveLength(1)
      })
      const url = mockedApiClient.get.mock.calls[0]?.[0] as string
      expect(url).toContain('/api/student-debts/my')
    })

    it('does not fetch when enabled is false', () => {
      renderHook(() => useMyStudentDebts(undefined, { enabled: false }), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('useDebtStats', () => {
    it('hits the /stats endpoint and returns the aggregate', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          total: 10,
          open: 4,
          resit_scheduled: 2,
          commission: 1,
          closed_passed: 2,
          closed_failed: 1,
        },
      })
      const { result } = renderHook(() => useDebtStats(), { wrapper })
      await waitFor(() => {
        expect(result.current.stats?.total).toBe(10)
      })
      const url = mockedApiClient.get.mock.calls[0]?.[0] as string
      expect(url).toBe('/api/student-debts/stats')
    })
  })
})

describe('useStudentDebts hooks (mutations)', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('scheduleResit POSTs to /:id/resit with the body', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { id: 55, status: 'resit_scheduled' } })
    const result = await scheduleResit(55, {
      scheduled_date: '2026-07-01T09:00:00Z',
      examiner: 'Петров П.П.',
    })
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/student-debts/55/resit', {
      scheduled_date: '2026-07-01T09:00:00Z',
      examiner: 'Петров П.П.',
    })
    expect(result.status).toBe('resit_scheduled')
  })

  it('recordResitResult POSTs to /:id/attempts/:n/result with the body', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { id: 55, status: 'closed_passed' } })
    const result = await recordResitResult(55, 1, { result: 'passed', grade: 5 })
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/student-debts/55/attempts/1/result', {
      result: 'passed',
      grade: 5,
    })
    expect(result.status).toBe('closed_passed')
  })
})

describe('pickStudentDebtErrorKey', () => {
  type Row = { name: string; err: unknown; expected: string }
  const mkErr = (status?: number, code?: string) => ({
    response: { status, data: code ? { error: { code, message: 'x' } } : undefined },
  })

  const cases: Row[] = [
    { name: 'VERSION_CONFLICT', err: mkErr(409, 'VERSION_CONFLICT'), expected: 'versionConflict' },
    { name: 'IDENTITY_EXISTS', err: mkErr(409, 'IDENTITY_EXISTS'), expected: 'identityExists' },
    { name: 'DEBT_CLOSED', err: mkErr(409, 'DEBT_CLOSED'), expected: 'debtClosed' },
    {
      name: 'NO_SCHEDULED_RESIT',
      err: mkErr(409, 'NO_SCHEDULED_RESIT'),
      expected: 'noScheduledResit',
    },
    { name: 'ALREADY_RECORDED', err: mkErr(409, 'ALREADY_RECORDED'), expected: 'alreadyRecorded' },
    {
      name: 'INVALID_TRANSITION',
      err: mkErr(409, 'INVALID_TRANSITION'),
      expected: 'invalidTransition',
    },
    { name: 'VALIDATION_ERROR', err: mkErr(422, 'VALIDATION_ERROR'), expected: 'validationError' },
    {
      name: 'code beats mismatched status',
      err: mkErr(404, 'VERSION_CONFLICT'),
      expected: 'versionConflict',
    },
    { name: 'plain 403 → forbidden', err: mkErr(403), expected: 'forbidden' },
    { name: 'plain 404 → notFound', err: mkErr(404), expected: 'notFound' },
    { name: 'unknown 500 → generic', err: mkErr(500), expected: 'generic' },
    { name: 'undefined → generic', err: undefined, expected: 'generic' },
  ]

  it.each(cases)('$name', ({ err, expected }) => {
    expect(pickStudentDebtErrorKey(err)).toBe(expected)
  })
})
