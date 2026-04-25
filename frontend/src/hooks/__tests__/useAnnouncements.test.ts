import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useAnnouncements,
  useAnnouncement,
  createAnnouncement,
  updateAnnouncement,
  deleteAnnouncement,
  publishAnnouncement,
  unpublishAnnouncement,
  archiveAnnouncement,
} from '../useAnnouncements'
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

describe('useAnnouncements hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('useAnnouncements', () => {
    it('returns announcements list from API', async () => {
      const mockResponse = {
        announcements: [
          {
            id: 1,
            title: 'Major news',
            content: 'Content here',
            author_id: 1,
            status: 'published',
            priority: 'high',
            target_audience: 'all',
            is_pinned: true,
            view_count: 42,
            created_at: '2026-04-25T10:00:00Z',
            updated_at: '2026-04-25T10:00:00Z',
          },
        ],
        total: 1,
        limit: 20,
        offset: 0,
      }

      mockedApiClient.get.mockResolvedValue({ data: mockResponse })

      const { result } = renderHook(() => useAnnouncements(), { wrapper })

      await waitFor(() => {
        expect(result.current.announcements).toHaveLength(1)
      })
      expect(result.current.total).toBe(1)
    })

    it('passes filter params as query string', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { announcements: [], total: 0, limit: 20, offset: 0 },
      })

      renderHook(() => useAnnouncements({ status: 'draft', priority: 'urgent', is_pinned: true }), {
        wrapper,
      })

      await waitFor(() => {
        const url = mockedApiClient.get.mock.calls[0]?.[0] as string
        expect(url).toContain('status=draft')
        expect(url).toContain('priority=urgent')
        expect(url).toContain('is_pinned=true')
      })
    })

    it('returns empty array when API returns null data', async () => {
      mockedApiClient.get.mockResolvedValue({ data: null })

      const { result } = renderHook(() => useAnnouncements(), { wrapper })

      expect(result.current.announcements).toEqual([])
    })
  })

  describe('useAnnouncement', () => {
    it('fetches single announcement by id', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { id: 7, title: 'X', content: 'Y', status: 'draft' },
      })

      const { result } = renderHook(() => useAnnouncement(7), { wrapper })

      await waitFor(() => {
        expect(result.current.announcement?.id).toBe(7)
      })
    })

    it('does not fetch when id is null', () => {
      renderHook(() => useAnnouncement(null), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('CRUD mutations', () => {
    it('createAnnouncement POSTs to /api/announcements', async () => {
      mockedApiClient.post.mockResolvedValue({ id: 1, title: 'X' })
      const result = await createAnnouncement({
        title: 'X',
        content: 'Y',
        priority: 'normal',
        target_audience: 'all',
      })
      expect(mockedApiClient.post).toHaveBeenCalledWith(
        '/api/announcements',
        expect.objectContaining({ title: 'X' })
      )
      expect(result.id).toBe(1)
    })

    it('updateAnnouncement PUTs to /api/announcements/:id', async () => {
      mockedApiClient.put.mockResolvedValue({ id: 5, title: 'New' })
      await updateAnnouncement(5, { title: 'New' })
      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/announcements/5', { title: 'New' })
    })

    it('deleteAnnouncement DELETEs /api/announcements/:id', async () => {
      mockedApiClient.delete.mockResolvedValue(undefined)
      await deleteAnnouncement(9)
      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/announcements/9')
    })
  })

  describe('status actions', () => {
    it('publishAnnouncement POSTs /publish', async () => {
      mockedApiClient.post.mockResolvedValue(undefined)
      await publishAnnouncement(3)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/announcements/3/publish')
    })

    it('unpublishAnnouncement POSTs /unpublish', async () => {
      mockedApiClient.post.mockResolvedValue(undefined)
      await unpublishAnnouncement(3)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/announcements/3/unpublish')
    })

    it('archiveAnnouncement POSTs /archive', async () => {
      mockedApiClient.post.mockResolvedValue(undefined)
      await archiveAnnouncement(3)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/announcements/3/archive')
    })
  })
})
