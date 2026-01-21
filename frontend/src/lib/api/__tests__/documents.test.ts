import { documentsApi, tagsApi } from '../documents'
import { apiClient } from '../../api'

// Mock the API client
jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('documentsApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('create', () => {
    it('creates new document', async () => {
      const newDoc = { title: 'New Doc', document_type_id: 1 }
      const mockResponse = { success: true, data: { id: 1, ...newDoc } }
      mockedApiClient.post.mockResolvedValue(mockResponse)

      const result = await documentsApi.create(newDoc)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/documents', newDoc)
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('list', () => {
    it('fetches documents with default parameters', async () => {
      const mockResponse = { success: true, data: [], meta: { pagination: { total: 0 } } }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      await documentsApi.list()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents', { params: undefined })
    })

    it('fetches documents with filters', async () => {
      const mockResponse = { success: true, data: [], meta: { pagination: { total: 0 } } }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      await documentsApi.list({
        page: 2,
        page_size: 20,
        status: 'draft',
        importance: 'high',
      })

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents', {
        params: { page: 2, page_size: 20, status: 'draft', importance: 'high' },
      })
    })
  })

  describe('getById', () => {
    it('fetches single document', async () => {
      const mockDoc = { id: 1, title: 'Test Document' }
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockDoc })

      const result = await documentsApi.getById(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/1')
      expect(result).toEqual(mockDoc)
    })
  })

  describe('update', () => {
    it('updates document', async () => {
      const updateData = { title: 'Updated Title' }
      mockedApiClient.put.mockResolvedValue({ success: true, data: { id: 1, ...updateData } })

      const result = await documentsApi.update(1, updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/documents/1', updateData)
      expect(result).toEqual({ id: 1, ...updateData })
    })
  })

  describe('delete', () => {
    it('deletes document', async () => {
      mockedApiClient.delete.mockResolvedValue({ message: 'Deleted' })

      await documentsApi.delete(1)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/documents/1')
    })
  })

  describe('uploadFile', () => {
    it('uploads file to document', async () => {
      mockedApiClient.post.mockResolvedValue({
        success: true,
        data: { id: 1, file_name: 'test.pdf' },
      })

      const file = new File(['content'], 'test.pdf', { type: 'application/pdf' })
      const result = await documentsApi.uploadFile(1, file)

      expect(mockedApiClient.post).toHaveBeenCalledWith(
        '/api/documents/1/file',
        expect.any(FormData),
        { headers: { 'Content-Type': 'multipart/form-data' } }
      )
      expect(result).toEqual({ id: 1, file_name: 'test.pdf' })
    })
  })

  describe('getFileDownloadUrl', () => {
    it('returns download URL', () => {
      const url = documentsApi.getFileDownloadUrl(1)
      expect(url).toContain('/api/documents/1/file')
    })

    it('uses API base URL from env', () => {
      const originalEnv = process.env.NEXT_PUBLIC_API_URL
      process.env.NEXT_PUBLIC_API_URL = 'https://api.example.com'

      const url = documentsApi.getFileDownloadUrl(1)

      expect(url).toBe('https://api.example.com/api/documents/1/file')

      process.env.NEXT_PUBLIC_API_URL = originalEnv
    })
  })

  describe('deleteFile', () => {
    it('deletes file from document', async () => {
      mockedApiClient.delete.mockResolvedValue({ message: 'Deleted' })

      await documentsApi.deleteFile(1)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/documents/1/file')
    })
  })

  describe('search', () => {
    it('searches documents', async () => {
      const mockResults = {
        results: [],
        query: 'test',
        total: 0,
        page: 1,
        page_size: 10,
        total_pages: 0,
      }
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockResults })

      const result = await documentsApi.search({ q: 'test query' })

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/search', {
        params: { q: 'test query' },
      })
      expect(result).toEqual(mockResults)
    })
  })

  // Sharing API tests
  describe('shareDocument', () => {
    it('shares document with user', async () => {
      const mockPermission = { id: 1, document_id: 1, user_id: 2, permission: 'read' }
      mockedApiClient.post.mockResolvedValue({ success: true, data: mockPermission })

      const result = await documentsApi.shareDocument(1, { user_id: 2, permission: 'read' })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/documents/1/share', {
        user_id: 2,
        permission: 'read',
      })
      expect(result).toEqual(mockPermission)
    })
  })

  describe('getPermissions', () => {
    it('fetches document permissions', async () => {
      const mockPermissions = [{ id: 1, permission: 'read' }]
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockPermissions })

      const result = await documentsApi.getPermissions(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/1/permissions')
      expect(result).toEqual(mockPermissions)
    })
  })

  describe('revokePermission', () => {
    it('revokes permission', async () => {
      mockedApiClient.delete.mockResolvedValue({ message: 'Revoked' })

      await documentsApi.revokePermission(1, 2)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/documents/1/permissions/2')
    })
  })

  describe('createPublicLink', () => {
    it('creates public link', async () => {
      const mockLink = { id: 1, token: 'abc123', url: 'http://example.com/abc123' }
      mockedApiClient.post.mockResolvedValue({ success: true, data: mockLink })

      const result = await documentsApi.createPublicLink(1, { permission: 'read' })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/documents/1/public-links', {
        permission: 'read',
      })
      expect(result).toEqual(mockLink)
    })
  })

  describe('getPublicLinks', () => {
    it('fetches public links', async () => {
      const mockLinks = [{ id: 1, token: 'abc123' }]
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockLinks })

      const result = await documentsApi.getPublicLinks(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/1/public-links')
      expect(result).toEqual(mockLinks)
    })
  })

  describe('deactivatePublicLink', () => {
    it('deactivates public link', async () => {
      mockedApiClient.post.mockResolvedValue({ message: 'Deactivated' })

      await documentsApi.deactivatePublicLink(1, 2)

      expect(mockedApiClient.post).toHaveBeenCalledWith(
        '/api/documents/1/public-links/2/deactivate'
      )
    })
  })

  describe('deletePublicLink', () => {
    it('deletes public link', async () => {
      mockedApiClient.delete.mockResolvedValue({ message: 'Deleted' })

      await documentsApi.deletePublicLink(1, 2)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/documents/1/public-links/2')
    })
  })

  describe('getSharedDocuments', () => {
    it('fetches documents shared with current user', async () => {
      const mockDocs = [{ id: 1, title: 'Shared Doc' }]
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockDocs })

      const result = await documentsApi.getSharedDocuments()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/shared', {
        params: undefined,
      })
      expect(result).toEqual(mockDocs)
    })
  })

  describe('getMySharedDocuments', () => {
    it('fetches documents shared by current user', async () => {
      const mockDocs = [{ document_id: 1, document_title: 'My Shared Doc', shared_with: [] }]
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockDocs })

      const result = await documentsApi.getMySharedDocuments()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/my-shared', {
        params: undefined,
      })
      expect(result).toEqual(mockDocs)
    })
  })

  describe('accessPublicDocument', () => {
    it('accesses document via public link', async () => {
      const mockDoc = { id: 1, title: 'Public Doc' }
      global.fetch = jest.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ data: mockDoc }),
      })

      const result = await documentsApi.accessPublicDocument('abc123')

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/public/documents/abc123'),
        expect.objectContaining({ method: 'POST' })
      )
      expect(result).toEqual(mockDoc)
    })

    it('throws error on failed access', async () => {
      global.fetch = jest.fn().mockResolvedValue({
        ok: false,
        json: () => Promise.resolve({ message: 'Invalid token' }),
      })

      await expect(documentsApi.accessPublicDocument('invalid')).rejects.toThrow('Invalid token')
    })
  })

  // Versioning API tests
  describe('getVersions', () => {
    it('fetches document versions', async () => {
      const mockVersions = {
        versions: [{ version: 1 }],
        total: 1,
        document_id: 1,
        latest_version: 1,
      }
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockVersions })

      const result = await documentsApi.getVersions(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/1/versions')
      expect(result).toEqual(mockVersions)
    })
  })

  describe('getVersion', () => {
    it('fetches specific version', async () => {
      const mockVersion = { id: 1, document_id: 1, version: 2 }
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockVersion })

      const result = await documentsApi.getVersion(1, 2)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/1/versions/2')
      expect(result).toEqual(mockVersion)
    })
  })

  describe('createVersion', () => {
    it('creates manual version snapshot', async () => {
      const mockVersion = { id: 2, document_id: 1, version: 2 }
      mockedApiClient.post.mockResolvedValue({ success: true, data: mockVersion })

      const result = await documentsApi.createVersion(1, { change_description: 'Manual backup' })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/documents/1/versions', {
        change_description: 'Manual backup',
      })
      expect(result).toEqual(mockVersion)
    })
  })

  describe('restoreVersion', () => {
    it('restores document version', async () => {
      const mockDoc = { id: 1, version: 2 }
      mockedApiClient.post.mockResolvedValue({ success: true, data: mockDoc })

      const result = await documentsApi.restoreVersion(1, 2)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/documents/1/versions/2/restore')
      expect(result).toEqual(mockDoc)
    })
  })

  describe('compareVersions', () => {
    it('compares two versions', async () => {
      const mockDiff = { document_id: 1, from_version: 1, to_version: 2, changed_fields: ['title'] }
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockDiff })

      const result = await documentsApi.compareVersions(1, 1, 2)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/1/versions/compare', {
        params: { from: 1, to: 2 },
      })
      expect(result).toEqual(mockDiff)
    })
  })

  describe('deleteVersion', () => {
    it('deletes specific version', async () => {
      mockedApiClient.delete.mockResolvedValue({ message: 'Deleted' })

      await documentsApi.deleteVersion(1, 2)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/documents/1/versions/2')
    })
  })

  describe('getVersionFile', () => {
    it('gets file from specific version', async () => {
      const mockFile = { file_name: 'doc.pdf', file_size: 1024 }
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockFile })

      const result = await documentsApi.getVersionFile(1, 2)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/1/versions/2/file')
      expect(result).toEqual(mockFile)
    })
  })
})

describe('tagsApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('getAll', () => {
    it('fetches all tags', async () => {
      const mockTags = [{ id: 1, name: 'Important' }]
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockTags })

      const result = await tagsApi.getAll()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/tags')
      expect(result).toEqual(mockTags)
    })
  })

  describe('search', () => {
    it('searches tags by name', async () => {
      const mockTags = [{ id: 1, name: 'Important' }]
      mockedApiClient.get.mockResolvedValue({ success: true, data: mockTags })

      const result = await tagsApi.search('imp', 5)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/tags/search', {
        params: { q: 'imp', limit: 5 },
      })
      expect(result).toEqual(mockTags)
    })
  })

  describe('getDocumentTags', () => {
    it('fetches tags for document', async () => {
      const mockTags = [{ id: 1, name: 'Important' }]
      mockedApiClient.get.mockResolvedValue({ success: true, data: { tags: mockTags } })

      const result = await tagsApi.getDocumentTags(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/1/tags')
      expect(result).toEqual(mockTags)
    })
  })

  describe('setDocumentTags', () => {
    it('sets tags for document', async () => {
      mockedApiClient.put.mockResolvedValue({ message: 'Updated' })

      await tagsApi.setDocumentTags(1, [1, 2, 3])

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/documents/1/tags', {
        tag_ids: [1, 2, 3],
      })
    })
  })

  describe('addTagToDocument', () => {
    it('adds tag to document', async () => {
      mockedApiClient.post.mockResolvedValue({ message: 'Added' })

      await tagsApi.addTagToDocument(1, 2)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/documents/1/tags/2')
    })
  })

  describe('removeTagFromDocument', () => {
    it('removes tag from document', async () => {
      mockedApiClient.delete.mockResolvedValue({ message: 'Removed' })

      await tagsApi.removeTagFromDocument(1, 2)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/documents/1/tags/2')
    })
  })
})
