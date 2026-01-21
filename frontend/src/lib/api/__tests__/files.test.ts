import { filesApi } from '../files'
import { apiClient } from '../../api'

// Mock the API client
jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('filesApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('upload', () => {
    it('uploads file with entity info', async () => {
      const mockFileInfo = {
        id: '123',
        filename: 'test.pdf',
        original_filename: 'test.pdf',
        content_type: 'application/pdf',
        size: 1024,
      }
      mockedApiClient.post.mockResolvedValue({ data: mockFileInfo })

      const file = new File(['content'], 'test.pdf', { type: 'application/pdf' })
      const result = await filesApi.upload({
        file,
        entity_type: 'document',
        entity_id: 'doc-1',
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/files/upload', expect.any(FormData), {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      expect(result).toEqual(mockFileInfo)
    })

    it('uploads file for task entity', async () => {
      const mockFileInfo = { id: '456', filename: 'task-file.docx' }
      mockedApiClient.post.mockResolvedValue({ data: mockFileInfo })

      const file = new File(['content'], 'task-file.docx')
      await filesApi.upload({
        file,
        entity_type: 'task',
        entity_id: 'task-1',
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith(
        '/api/files/upload',
        expect.any(FormData),
        expect.any(Object)
      )
    })
  })

  describe('getByDocument', () => {
    it('fetches files by document ID', async () => {
      const mockFiles = [
        { id: '1', filename: 'file1.pdf' },
        { id: '2', filename: 'file2.pdf' },
      ]
      mockedApiClient.get.mockResolvedValue({ data: mockFiles })

      const result = await filesApi.getByDocument('doc-1')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/files/by-document/doc-1')
      expect(result).toEqual(mockFiles)
    })
  })

  describe('getByTask', () => {
    it('fetches files by task ID', async () => {
      const mockFiles = [{ id: '1', filename: 'task-file.pdf' }]
      mockedApiClient.get.mockResolvedValue({ data: mockFiles })

      const result = await filesApi.getByTask('task-1')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/files/by-task/task-1')
      expect(result).toEqual(mockFiles)
    })
  })

  describe('getByAnnouncement', () => {
    it('fetches files by announcement ID', async () => {
      const mockFiles = [{ id: '1', filename: 'announcement.pdf' }]
      mockedApiClient.get.mockResolvedValue({ data: mockFiles })

      const result = await filesApi.getByAnnouncement('ann-1')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/files/by-announcement/ann-1')
      expect(result).toEqual(mockFiles)
    })
  })

  describe('getById', () => {
    it('fetches single file by ID', async () => {
      const mockFile = { id: '123', filename: 'test.pdf' }
      mockedApiClient.get.mockResolvedValue({ data: mockFile })

      const result = await filesApi.getById('123')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/files/123')
      expect(result).toEqual(mockFile)
    })
  })

  describe('getDownloadUrl', () => {
    it('returns download URL', () => {
      const url = filesApi.getDownloadUrl('123')

      expect(url).toContain('/api/files/123/download')
    })

    it('uses API base URL', () => {
      const originalEnv = process.env.NEXT_PUBLIC_API_URL
      process.env.NEXT_PUBLIC_API_URL = 'https://api.example.com'

      const url = filesApi.getDownloadUrl('456')

      expect(url).toBe('https://api.example.com/api/files/456/download')

      process.env.NEXT_PUBLIC_API_URL = originalEnv
    })

    it('uses default URL when env not set', () => {
      const originalEnv = process.env.NEXT_PUBLIC_API_URL
      delete process.env.NEXT_PUBLIC_API_URL

      const url = filesApi.getDownloadUrl('789')

      expect(url).toBe('http://localhost:8080/api/files/789/download')

      process.env.NEXT_PUBLIC_API_URL = originalEnv
    })
  })

  describe('delete', () => {
    it('deletes file by ID', async () => {
      mockedApiClient.delete.mockResolvedValue(undefined)

      await filesApi.delete('123')

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/files/123')
    })
  })
})
