import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import { useExtracurricularEvents, useExtracurricularEvent } from '../useExtracurricularEvents'
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
        category: 'academic',
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
