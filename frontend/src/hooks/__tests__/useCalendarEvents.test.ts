import { renderHook, waitFor, act } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useEvents,
  useEventsByDateRange,
  useEvent,
  useUpcomingEvents,
  useCalendarOperations,
  createEvent,
  updateEvent,
  deleteEvent,
  cancelEvent,
  rescheduleEvent,
} from '../useCalendarEvents'
import { apiClient } from '@/lib/api'

// Mock the API client
jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

// Wrapper to reset SWR cache between tests
const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(
    SWRConfig,
    { value: { dedupingInterval: 0, provider: () => new Map() } },
    children
  )

describe('useCalendarEvents hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('useEvents', () => {
    it('returns events list', async () => {
      const mockResponse = {
        events: [
          { id: 1, title: 'Event 1', start_time: '2024-01-01T10:00:00Z' },
          { id: 2, title: 'Event 2', start_time: '2024-01-02T10:00:00Z' },
        ],
        total: 2,
        page: 1,
        page_size: 20,
        total_pages: 1,
      }

      mockedApiClient.get.mockResolvedValue({
        data: mockResponse,
      })

      const { result } = renderHook(() => useEvents(), { wrapper })

      await waitFor(() => {
        expect(result.current.events).toHaveLength(2)
      })

      expect(result.current.total).toBe(2)
      expect(result.current.page).toBe(1)
    })

    it('passes filter parameters', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { events: [], total: 0, page: 1, page_size: 20, total_pages: 0 },
      })

      renderHook(() => useEvents({ event_type: 'meeting', page: 2, page_size: 10 }), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith(
          expect.stringContaining('event_type=meeting')
        )
      })
    })

    it('returns empty array when no data', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: null,
      })

      const { result } = renderHook(() => useEvents(), { wrapper })

      expect(result.current.events).toEqual([])
      expect(result.current.total).toBe(0)
    })
  })

  describe('useEventsByDateRange', () => {
    it('fetches events by date range', async () => {
      const mockEvents = [{ id: 1, title: 'Event in range', start_time: '2024-01-15T10:00:00Z' }]

      mockedApiClient.get.mockResolvedValue({
        data: mockEvents,
      })

      const start = new Date('2024-01-01')
      const end = new Date('2024-01-31')

      const { result } = renderHook(() => useEventsByDateRange(start, end), { wrapper })

      await waitFor(() => {
        expect(result.current.events).toHaveLength(1)
      })

      expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('/api/events/range'))
    })
  })

  describe('useEvent', () => {
    it('fetches single event', async () => {
      const mockEvent = { id: 1, title: 'Single Event', start_time: '2024-01-01T10:00:00Z' }

      mockedApiClient.get.mockResolvedValue({
        data: mockEvent,
      })

      const { result } = renderHook(() => useEvent(1), { wrapper })

      await waitFor(() => {
        expect(result.current.event).toEqual(mockEvent)
      })
    })

    it('does not fetch when id is null', () => {
      renderHook(() => useEvent(null), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('useUpcomingEvents', () => {
    it('fetches upcoming events with default limit', async () => {
      const mockEvents = [
        { id: 1, title: 'Upcoming 1' },
        { id: 2, title: 'Upcoming 2' },
      ]

      mockedApiClient.get.mockResolvedValue({
        data: mockEvents,
      })

      const { result } = renderHook(() => useUpcomingEvents(), { wrapper })

      await waitFor(() => {
        expect(result.current.events).toHaveLength(2)
      })

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/events/upcoming?limit=10')
    })

    it('uses custom limit parameter', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: [],
      })

      renderHook(() => useUpcomingEvents(5), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith('/api/events/upcoming?limit=5')
      })
    })
  })

  describe('createEvent', () => {
    it('creates event', async () => {
      const newEvent = { id: 1, title: 'New Event' }
      mockedApiClient.post.mockResolvedValue(newEvent)

      const result = await createEvent({
        title: 'New Event',
        start_time: '2024-01-01T10:00:00Z',
        end_time: '2024-01-01T11:00:00Z',
        event_type: 'meeting',
        all_day: false,
        is_recurring: false,
      })

      expect(result).toEqual(newEvent)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/events', {
        title: 'New Event',
        start_time: '2024-01-01T10:00:00Z',
        end_time: '2024-01-01T11:00:00Z',
        event_type: 'meeting',
        all_day: false,
        is_recurring: false,
      })
    })
  })

  describe('updateEvent', () => {
    it('updates event', async () => {
      const updatedEvent = { id: 1, title: 'Updated Event' }
      mockedApiClient.put.mockResolvedValue(updatedEvent)

      const result = await updateEvent(1, { title: 'Updated Event' })

      expect(result).toEqual(updatedEvent)
      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/events/1', {
        title: 'Updated Event',
      })
    })
  })

  describe('deleteEvent', () => {
    it('deletes event', async () => {
      mockedApiClient.delete.mockResolvedValue(undefined)

      await deleteEvent(1)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/events/1')
    })
  })

  describe('cancelEvent', () => {
    it('cancels event', async () => {
      const cancelledEvent = { id: 1, status: 'cancelled' }
      mockedApiClient.post.mockResolvedValue(cancelledEvent)

      const result = await cancelEvent(1)

      expect(result).toEqual(cancelledEvent)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/events/1/cancel')
    })
  })

  describe('rescheduleEvent', () => {
    it('reschedules event', async () => {
      const rescheduledEvent = { id: 1, start_time: '2024-02-01T10:00:00Z' }
      mockedApiClient.post.mockResolvedValue(rescheduledEvent)

      const result = await rescheduleEvent(1, '2024-02-01T10:00:00Z', '2024-02-01T11:00:00Z')

      expect(result).toEqual(rescheduledEvent)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/events/1/reschedule', {
        start_time: '2024-02-01T10:00:00Z',
        end_time: '2024-02-01T11:00:00Z',
      })
    })
  })

  describe('useCalendarOperations', () => {
    it('provides calendar operations', async () => {
      const { result } = renderHook(() => useCalendarOperations(), { wrapper })

      expect(result.current.create).toBeDefined()
      expect(result.current.update).toBeDefined()
      expect(result.current.remove).toBeDefined()
      expect(result.current.cancel).toBeDefined()
      expect(result.current.reschedule).toBeDefined()
    })

    it('create operation works', async () => {
      mockedApiClient.post.mockResolvedValue({ id: 1 })

      const { result } = renderHook(() => useCalendarOperations(), { wrapper })

      await act(async () => {
        await result.current.create({
          title: 'Test',
          start_time: '2024-01-01T10:00:00Z',
          end_time: '2024-01-01T11:00:00Z',
          event_type: 'meeting',
          all_day: false,
          is_recurring: false,
        })
      })

      expect(mockedApiClient.post).toHaveBeenCalled()
    })
  })
})
