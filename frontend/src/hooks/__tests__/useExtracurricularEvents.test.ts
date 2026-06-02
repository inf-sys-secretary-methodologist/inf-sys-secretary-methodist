import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useExtracurricularEvents,
  useExtracurricularEvent,
  createExtracurricularEvent,
  updateExtracurricularEvent,
  deleteExtracurricularEvent,
  registerForExtracurricularEvent,
  unregisterFromExtracurricularEvent,
  pickExtracurricularErrorKey,
} from '../useExtracurricularEvents'
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

describe('useExtracurricularEvents hooks (queries)', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('useExtracurricularEvents', () => {
    it('returns events list from API', async () => {
      const mockResponse = {
        items: [
          {
            id: 1,
            title: 'Spring concert',
            category: 'cultural',
            target_audience: 'all',
            status: 'published',
            location: 'Main hall',
            start_at: '2026-06-01T18:00:00Z',
            end_at: '2026-06-01T21:00:00Z',
            max_capacity: 200,
            organizer_id: 5,
            participant_count: 42,
            version: 3,
            created_at: '2026-05-20T10:00:00Z',
            updated_at: '2026-05-25T12:00:00Z',
          },
        ],
        total: 1,
      }
      mockedApiClient.get.mockResolvedValue({ data: mockResponse })

      const { result } = renderHook(() => useExtracurricularEvents(), { wrapper })

      await waitFor(() => {
        expect(result.current.events).toHaveLength(1)
      })
      expect(result.current.total).toBe(1)
      expect(result.current.events[0].title).toBe('Spring concert')
    })

    it('passes filter params as query string', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { items: [], total: 0 },
      })

      renderHook(
        () =>
          useExtracurricularEvents({
            status: 'published',
            category: 'sports',
            from: '2026-06-01',
            to: '2026-06-30',
            limit: 50,
          }),
        { wrapper }
      )

      await waitFor(() => {
        const url = mockedApiClient.get.mock.calls[0]?.[0] as string
        expect(url).toContain('status=published')
        expect(url).toContain('category=sports')
        expect(url).toContain('from=2026-06-01')
        expect(url).toContain('to=2026-06-30')
        expect(url).toContain('limit=50')
      })
    })

    it('returns empty array when API returns null data', async () => {
      mockedApiClient.get.mockResolvedValue({ data: null })

      const { result } = renderHook(() => useExtracurricularEvents(), { wrapper })

      expect(result.current.events).toEqual([])
      expect(result.current.total).toBe(0)
    })

    it('omits undefined/null filter params from query string', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { items: [], total: 0 },
      })

      renderHook(
        () =>
          useExtracurricularEvents({
            status: 'draft',
            category: undefined,
            organizer_id: undefined,
          }),
        { wrapper }
      )

      await waitFor(() => {
        const url = mockedApiClient.get.mock.calls[0]?.[0] as string
        expect(url).toContain('status=draft')
        expect(url).not.toContain('category=')
        expect(url).not.toContain('organizer_id=')
      })
    })
  })

  describe('useExtracurricularEvent', () => {
    it('fetches single event by id', async () => {
      const mockEvent = {
        id: 7,
        title: 'Hackathon',
        description: 'Annual programming hackathon',
        category: 'cultural',
        target_audience: 'students',
        status: 'published',
        location: 'Lab 3',
        start_at: '2026-07-01T09:00:00Z',
        end_at: '2026-07-02T18:00:00Z',
        max_capacity: 50,
        organizer_id: 12,
        participants: [],
        participant_count: 0,
        version: 1,
        created_at: '2026-05-20T10:00:00Z',
        updated_at: '2026-05-20T10:00:00Z',
      }
      mockedApiClient.get.mockResolvedValue({ data: mockEvent })

      const { result } = renderHook(() => useExtracurricularEvent(7), { wrapper })

      await waitFor(() => {
        expect(result.current.event?.id).toBe(7)
      })
      expect(result.current.event?.title).toBe('Hackathon')
    })

    it('does not fetch when id is null', () => {
      renderHook(() => useExtracurricularEvent(null), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })

    it('hits canonical event detail URL', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { id: 99, title: 'X', category: 'cultural', status: 'draft' },
      })

      renderHook(() => useExtracurricularEvent(99), { wrapper })

      await waitFor(() => {
        const url = mockedApiClient.get.mock.calls[0]?.[0] as string
        expect(url).toBe('/api/v1/extracurricular/events/99')
      })
    })
  })
})

describe('useExtracurricularEvents hooks (mutations)', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('createExtracurricularEvent', () => {
    it('POSTs to /api/v1/extracurricular/events with payload', async () => {
      mockedApiClient.post.mockResolvedValue({ data: { id: 1, title: 'X' } })
      const result = await createExtracurricularEvent({
        title: 'X',
        category: 'cultural',
        target_audience: 'students',
        start_at: '2026-06-01T10:00:00Z',
        end_at: '2026-06-01T12:00:00Z',
      })
      expect(mockedApiClient.post).toHaveBeenCalledWith(
        '/api/v1/extracurricular/events',
        expect.objectContaining({ title: 'X', category: 'cultural' })
      )
      expect(result.id).toBe(1)
    })
  })

  describe('updateExtracurricularEvent', () => {
    it('PUTs to /api/v1/extracurricular/events/:id with payload', async () => {
      mockedApiClient.put.mockResolvedValue({ data: { id: 5, title: 'New' } })
      const result = await updateExtracurricularEvent(5, {
        title: 'New',
        category: 'cultural',
        target_audience: 'all',
        start_at: '2026-06-01T10:00:00Z',
        end_at: '2026-06-01T12:00:00Z',
      })
      expect(mockedApiClient.put).toHaveBeenCalledWith(
        '/api/v1/extracurricular/events/5',
        expect.objectContaining({ title: 'New' })
      )
      expect(result.id).toBe(5)
    })
  })

  describe('deleteExtracurricularEvent', () => {
    it('DELETEs /api/v1/extracurricular/events/:id', async () => {
      mockedApiClient.delete.mockResolvedValue(undefined)
      await deleteExtracurricularEvent(9)
      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/v1/extracurricular/events/9')
    })
  })

  describe('registerForExtracurricularEvent', () => {
    it('POSTs to /:id/register', async () => {
      mockedApiClient.post.mockResolvedValue(undefined)
      await registerForExtracurricularEvent(3)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/extracurricular/events/3/register')
    })
  })

  describe('unregisterFromExtracurricularEvent', () => {
    it('DELETEs /:id/register', async () => {
      mockedApiClient.delete.mockResolvedValue(undefined)
      await unregisterFromExtracurricularEvent(3)
      expect(mockedApiClient.delete).toHaveBeenCalledWith(
        '/api/v1/extracurricular/events/3/register'
      )
    })
  })
})

describe('pickExtracurricularErrorKey', () => {
  // Table-driven per CLAUDE.md ≥3-variant gate.
  type Row = {
    name: string
    err: unknown
    expected: string
  }

  const mkErr = (status?: number, code?: string) => ({
    response: {
      status,
      data: code ? { error: { code, message: 'x' } } : undefined,
    },
  })

  const cases: Row[] = [
    {
      name: 'VERSION_CONFLICT code → versionConflict',
      err: mkErr(409, 'VERSION_CONFLICT'),
      expected: 'versionConflict',
    },
    {
      name: 'ALREADY_REGISTERED → alreadyRegistered',
      err: mkErr(409, 'ALREADY_REGISTERED'),
      expected: 'alreadyRegistered',
    },
    {
      name: 'EVENT_FULL → eventFull',
      err: mkErr(409, 'EVENT_FULL'),
      expected: 'eventFull',
    },
    {
      name: 'REGISTRATION_CLOSED → registrationClosed',
      err: mkErr(422, 'REGISTRATION_CLOSED'),
      expected: 'registrationClosed',
    },
    {
      name: 'CANNOT_EDIT → cannotEdit',
      err: mkErr(422, 'CANNOT_EDIT'),
      expected: 'cannotEdit',
    },
    {
      name: 'INVALID_EVENT → invalidEvent',
      err: mkErr(422, 'INVALID_EVENT'),
      expected: 'invalidEvent',
    },
    {
      name: 'plain 403 (no code) → forbidden',
      err: mkErr(403),
      expected: 'forbidden',
    },
    {
      name: 'plain 404 (no code) → notFound',
      err: mkErr(404),
      expected: 'notFound',
    },
    {
      name: 'unknown 500 → generic',
      err: mkErr(500),
      expected: 'generic',
    },
    {
      name: 'undefined error → generic (safe fallback)',
      err: undefined,
      expected: 'generic',
    },
  ]

  it.each(cases)('$name', ({ err, expected }) => {
    expect(pickExtracurricularErrorKey(err)).toBe(expected)
  })
})
