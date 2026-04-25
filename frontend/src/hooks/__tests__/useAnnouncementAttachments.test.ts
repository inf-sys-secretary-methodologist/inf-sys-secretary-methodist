import {
  uploadAnnouncementAttachment,
  deleteAnnouncementAttachment,
} from '../useAnnouncements'
import { apiClient } from '@/lib/api'

jest.mock('@/lib/api', () => ({
  apiClient: {
    post: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('announcement attachment mutations', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('uploadAnnouncementAttachment POSTs FormData to /:id/attachments', async () => {
    const mockResponse = {
      id: 11,
      announcement_id: 1,
      file_name: 'doc.pdf',
      file_size: 100,
      mime_type: 'application/pdf',
      created_at: '2026-04-25T10:00:00Z',
    }
    mockedApiClient.post.mockResolvedValue(mockResponse)

    const file = new File(['body'], 'doc.pdf', { type: 'application/pdf' })
    const result = await uploadAnnouncementAttachment(1, file)

    expect(mockedApiClient.post).toHaveBeenCalledWith(
      '/api/announcements/1/attachments',
      expect.any(FormData),
      expect.objectContaining({ headers: expect.objectContaining({ 'Content-Type': 'multipart/form-data' }) })
    )
    expect(result.file_name).toBe('doc.pdf')
  })

  it('deleteAnnouncementAttachment DELETEs /:id/attachments/:attachmentID', async () => {
    mockedApiClient.delete.mockResolvedValue(undefined)
    await deleteAnnouncementAttachment(7, 9)
    expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/announcements/7/attachments/9')
  })
})
