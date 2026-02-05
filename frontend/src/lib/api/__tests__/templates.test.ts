import { templatesApi, TemplateInfo } from '../templates'
import { apiClient } from '../../api'

// Mock apiClient
jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
  },
}))

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>

describe('templatesApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('getAll', () => {
    it('returns templates list', async () => {
      const mockTemplates: TemplateInfo[] = [
        {
          id: 1,
          name: 'Test Template',
          code: 'test-template',
          description: 'A test template',
          has_template: true,
          template_variables: [{ name: 'name', required: true, variable_type: 'text' }],
        },
      ]

      mockApiClient.get.mockResolvedValueOnce({ templates: mockTemplates, total: 1 })

      const result = await templatesApi.getAll()

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/templates')
      expect(result).toEqual(mockTemplates)
    })

    it('returns empty array when no templates', async () => {
      mockApiClient.get.mockResolvedValueOnce({ templates: undefined, total: 0 })

      const result = await templatesApi.getAll()

      expect(result).toEqual([])
    })
  })

  describe('getById', () => {
    it('returns template by numeric id', async () => {
      const mockTemplate: TemplateInfo = {
        id: 1,
        name: 'Test Template',
        code: 'test-template',
        has_template: true,
      }

      mockApiClient.get.mockResolvedValueOnce(mockTemplate)

      const result = await templatesApi.getById(1)

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/templates/1')
      expect(result).toEqual(mockTemplate)
    })

    it('returns template by string id', async () => {
      const mockTemplate: TemplateInfo = {
        id: 1,
        name: 'Test Template',
        code: 'test-template',
        has_template: true,
      }

      mockApiClient.get.mockResolvedValueOnce(mockTemplate)

      const result = await templatesApi.getById('1')

      expect(mockApiClient.get).toHaveBeenCalledWith('/api/templates/1')
      expect(result).toEqual(mockTemplate)
    })
  })

  describe('preview', () => {
    it('returns preview content', async () => {
      const mockContent = '<html><body>Hello, John!</body></html>'
      mockApiClient.post.mockResolvedValueOnce({ content: mockContent })

      const result = await templatesApi.preview(1, { name: 'John' })

      expect(mockApiClient.post).toHaveBeenCalledWith('/api/templates/1/preview', {
        variables: { name: 'John' },
      })
      expect(result).toBe(mockContent)
    })

    it('handles empty variables', async () => {
      const mockContent = '<html><body>Default content</body></html>'
      mockApiClient.post.mockResolvedValueOnce({ content: mockContent })

      const result = await templatesApi.preview(1, {})

      expect(mockApiClient.post).toHaveBeenCalledWith('/api/templates/1/preview', {
        variables: {},
      })
      expect(result).toBe(mockContent)
    })
  })

  describe('createDocument', () => {
    it('creates document from template', async () => {
      const mockDocument = {
        id: 123,
        title: 'New Document',
        status: 'draft',
      }

      mockApiClient.post.mockResolvedValueOnce({ data: mockDocument })

      const result = await templatesApi.createDocument(1, {
        title: 'New Document',
        variables: { name: 'John' },
        category_id: 2,
      })

      expect(mockApiClient.post).toHaveBeenCalledWith('/api/templates/1/create', {
        title: 'New Document',
        variables: { name: 'John' },
        category_id: 2,
      })
      expect(result).toEqual(mockDocument)
    })

    it('creates document without category_id', async () => {
      const mockDocument = { id: 123, title: 'New Document', status: 'draft' }
      mockApiClient.post.mockResolvedValueOnce({ data: mockDocument })

      await templatesApi.createDocument(1, {
        title: 'New Document',
        variables: {},
      })

      expect(mockApiClient.post).toHaveBeenCalledWith('/api/templates/1/create', {
        title: 'New Document',
        variables: {},
      })
    })
  })

  describe('update', () => {
    it('updates template content', async () => {
      mockApiClient.put.mockResolvedValueOnce(undefined)

      await templatesApi.update(1, {
        template_content: 'New content',
      })

      expect(mockApiClient.put).toHaveBeenCalledWith('/api/templates/1', {
        template_content: 'New content',
      })
    })

    it('updates template variables', async () => {
      mockApiClient.put.mockResolvedValueOnce(undefined)

      await templatesApi.update(1, {
        template_variables: [{ name: 'date', required: true, variable_type: 'date' }],
      })

      expect(mockApiClient.put).toHaveBeenCalledWith('/api/templates/1', {
        template_variables: [{ name: 'date', required: true, variable_type: 'date' }],
      })
    })
  })
})
