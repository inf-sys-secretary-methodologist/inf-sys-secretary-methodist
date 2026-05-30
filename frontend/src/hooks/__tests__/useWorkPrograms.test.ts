import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useWorkPrograms,
  useWorkProgram,
  createWorkProgram,
  submitWorkProgram,
  approveWorkProgram,
  rejectWorkProgram,
  discardWorkProgram,
  generateWorkProgram,
  createRevision,
  submitRevision,
  approveRevision,
  rejectRevision,
  pickWorkProgramErrorKey,
} from '../useWorkPrograms'
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

describe('useWorkPrograms hooks (queries)', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('useWorkPrograms', () => {
    it('returns the work-programs list from API', async () => {
      const mockResponse = {
        items: [
          {
            id: 1,
            discipline_id: 10,
            specialty_code: '09.03.01',
            applicable_from_year: 2026,
            title: 'Базы данных',
            status: 'approved',
            author_id: 5,
            version: 2,
          },
        ],
        total: 1,
      }
      mockedApiClient.get.mockResolvedValue({ data: mockResponse })

      const { result } = renderHook(() => useWorkPrograms(), { wrapper })

      await waitFor(() => {
        expect(result.current.items).toHaveLength(1)
      })
      expect(result.current.total).toBe(1)
      expect(result.current.items[0].title).toBe('Базы данных')
    })

    it('passes filter params as a query string', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { items: [], total: 0 } })

      renderHook(
        () =>
          useWorkPrograms({
            status: 'pending_approval',
            specialty_code: '09.03.01',
            applicable_from_year: 2026,
            limit: 20,
          }),
        { wrapper }
      )

      await waitFor(() => {
        const url = mockedApiClient.get.mock.calls[0]?.[0] as string
        expect(url).toContain('status=pending_approval')
        expect(url).toContain('specialty_code=09.03.01')
        expect(url).toContain('applicable_from_year=2026')
        expect(url).toContain('limit=20')
      })
    })

    it('returns empty array when API returns null data', () => {
      mockedApiClient.get.mockResolvedValue({ data: null })
      const { result } = renderHook(() => useWorkPrograms(), { wrapper })
      expect(result.current.items).toEqual([])
      expect(result.current.total).toBe(0)
    })

    it('omits undefined filter params from the query string', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { items: [], total: 0 } })

      renderHook(
        () =>
          useWorkPrograms({
            status: 'draft',
            discipline_id: undefined,
            author_id: undefined,
          }),
        { wrapper }
      )

      await waitFor(() => {
        const url = mockedApiClient.get.mock.calls[0]?.[0] as string
        expect(url).toContain('status=draft')
        expect(url).not.toContain('discipline_id=')
        expect(url).not.toContain('author_id=')
      })
    })

    it('does not fetch when enabled is false', () => {
      renderHook(() => useWorkPrograms(undefined, { enabled: false }), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('useWorkProgram', () => {
    it('fetches a single work program by id', async () => {
      const mockWP = {
        id: 7,
        discipline_id: 10,
        specialty_code: '09.03.01',
        applicable_from_year: 2026,
        title: 'ОС',
        annotation: '',
        status: 'draft',
        author_id: 5,
        version: 0,
        created_at: '2026-05-20T10:00:00Z',
        updated_at: '2026-05-20T10:00:00Z',
        goals: [],
        competences: [],
        topics: [],
        assessments: [],
        references: [],
        revisions: [],
      }
      mockedApiClient.get.mockResolvedValue({ data: mockWP })

      const { result } = renderHook(() => useWorkProgram(7), { wrapper })

      await waitFor(() => {
        expect(result.current.workProgram?.id).toBe(7)
      })
      expect(result.current.workProgram?.title).toBe('ОС')
    })

    it('does not fetch when id is null', () => {
      renderHook(() => useWorkProgram(null), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })

    it('hits the canonical detail URL', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { id: 99 } })
      renderHook(() => useWorkProgram(99), { wrapper })
      await waitFor(() => {
        const url = mockedApiClient.get.mock.calls[0]?.[0] as string
        expect(url).toBe('/api/v1/work-programs/99')
      })
    })

    it('does not fetch when enabled is false even with a valid id', () => {
      renderHook(() => useWorkProgram(7, { enabled: false }), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })
})

describe('useWorkPrograms hooks (mutations)', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('createWorkProgram POSTs to the collection URL with payload', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { id: 1, title: 'X' } })
    const result = await createWorkProgram({
      discipline_id: 10,
      specialty_code: '09.03.01',
      applicable_from_year: 2026,
      title: 'X',
    })
    expect(mockedApiClient.post).toHaveBeenCalledWith(
      '/api/v1/work-programs',
      expect.objectContaining({ discipline_id: 10, title: 'X' })
    )
    expect(result.id).toBe(1)
  })

  it('submitWorkProgram POSTs to /:id/submit with an empty body', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { id: 5, status: 'pending_approval' } })
    const result = await submitWorkProgram(5)
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/work-programs/5/submit', {})
    expect(result.status).toBe('pending_approval')
  })

  it('approveWorkProgram POSTs to /:id/approve with an empty body', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { id: 5, status: 'approved' } })
    const result = await approveWorkProgram(5)
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/work-programs/5/approve', {})
    expect(result.status).toBe('approved')
  })

  it('rejectWorkProgram POSTs to /:id/reject with the reason body', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { id: 5, status: 'draft' } })
    const result = await rejectWorkProgram(5, { reason: 'fix hours' })
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/work-programs/5/reject', {
      reason: 'fix hours',
    })
    expect(result.status).toBe('draft')
  })

  it('discardWorkProgram POSTs to /:id/discard with an empty body', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { id: 5, status: 'archived' } })
    const result = await discardWorkProgram(5)
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/work-programs/5/discard', {})
    expect(result.status).toBe('archived')
  })

  it('generateWorkProgram POSTs to /:id/generate with an empty body', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { id: 5, status: 'draft' } })
    const result = await generateWorkProgram(5)
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/work-programs/5/generate', {})
    expect(result.status).toBe('draft')
  })

  it('createRevision POSTs to /:id/revisions with the change body', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { id: 5, status: 'approved' } })
    const result = await createRevision(5, {
      change_type: 'hours',
      change_summary: 'Часы лекций 36 → 18',
    })
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/work-programs/5/revisions', {
      change_type: 'hours',
      change_summary: 'Часы лекций 36 → 18',
    })
    expect(result.id).toBe(5)
  })

  it('submitRevision POSTs to /:id/revisions/:rid/submit with an empty body', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { id: 5, status: 'approved' } })
    const result = await submitRevision(5, 9)
    expect(mockedApiClient.post).toHaveBeenCalledWith(
      '/api/v1/work-programs/5/revisions/9/submit',
      {}
    )
    expect(result.id).toBe(5)
  })

  it('approveRevision POSTs to /:id/revisions/:rid/approve with an empty body', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { id: 5, status: 'approved' } })
    const result = await approveRevision(5, 9)
    expect(mockedApiClient.post).toHaveBeenCalledWith(
      '/api/v1/work-programs/5/revisions/9/approve',
      {}
    )
    expect(result.id).toBe(5)
  })

  it('rejectRevision POSTs to /:id/revisions/:rid/reject with the reason body', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { id: 5, status: 'approved' } })
    const result = await rejectRevision(5, 9, { reason: 'нет обоснования' })
    expect(mockedApiClient.post).toHaveBeenCalledWith(
      '/api/v1/work-programs/5/revisions/9/reject',
      { reason: 'нет обоснования' }
    )
    expect(result.id).toBe(5)
  })
})

describe('pickWorkProgramErrorKey', () => {
  // Table-driven per CLAUDE.md ≥3-variant gate. Codes mirror
  // mapWorkProgramError in work_program_handler.go.
  type Row = { name: string; err: unknown; expected: string }

  const mkErr = (status?: number, code?: string) => ({
    response: {
      status,
      data: code ? { error: { code, message: 'x' } } : undefined,
    },
  })

  const cases: Row[] = [
    {
      name: 'IDENTITY_EXISTS → identityExists',
      err: mkErr(409, 'IDENTITY_EXISTS'),
      expected: 'identityExists',
    },
    {
      name: 'VERSION_CONFLICT → versionConflict',
      err: mkErr(409, 'VERSION_CONFLICT'),
      expected: 'versionConflict',
    },
    {
      name: 'INVALID_TRANSITION → invalidTransition',
      err: mkErr(422, 'INVALID_TRANSITION'),
      expected: 'invalidTransition',
    },
    {
      name: 'REJECT_REASON_REQUIRED → rejectReasonRequired',
      err: mkErr(422, 'REJECT_REASON_REQUIRED'),
      expected: 'rejectReasonRequired',
    },
    {
      name: 'INVALID_WORK_PROGRAM → invalidWorkProgram',
      err: mkErr(422, 'INVALID_WORK_PROGRAM'),
      expected: 'invalidWorkProgram',
    },
    {
      name: 'RATE_LIMITED → rateLimited',
      err: mkErr(429, 'RATE_LIMITED'),
      expected: 'rateLimited',
    },
    {
      name: 'DRAFT_NOT_EMPTY → draftNotEmpty',
      err: mkErr(409, 'DRAFT_NOT_EMPTY'),
      expected: 'draftNotEmpty',
    },
    {
      name: 'REVISION_NOT_PERMITTED → revisionNotPermitted',
      err: mkErr(422, 'REVISION_NOT_PERMITTED'),
      expected: 'revisionNotPermitted',
    },
    {
      // A sentinel code must win over a mismatched HTTP status — pins
      // that the code branch runs before the status fallbacks.
      name: 'code beats mismatched status (404 + VERSION_CONFLICT) → versionConflict',
      err: mkErr(404, 'VERSION_CONFLICT'),
      expected: 'versionConflict',
    },
    { name: 'plain 403 (no code) → forbidden', err: mkErr(403), expected: 'forbidden' },
    { name: 'plain 404 (no code) → notFound', err: mkErr(404), expected: 'notFound' },
    { name: 'unknown 500 → generic', err: mkErr(500), expected: 'generic' },
    { name: 'undefined error → generic (safe fallback)', err: undefined, expected: 'generic' },
  ]

  it.each(cases)('$name', ({ err, expected }) => {
    expect(pickWorkProgramErrorKey(err)).toBe(expected)
  })
})
