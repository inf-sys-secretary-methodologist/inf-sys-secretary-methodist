import { mockDocuments, filterDocuments, sortDocuments } from '../mock-documents'
import { DocumentCategory, DocumentStatus, Document } from '@/types/document'

describe('mock-documents', () => {
  describe('mockDocuments', () => {
    it('exports an array of documents', () => {
      expect(Array.isArray(mockDocuments)).toBe(true)
      expect(mockDocuments.length).toBeGreaterThan(0)
    })

    it('each document has required fields', () => {
      mockDocuments.forEach((doc) => {
        expect(doc).toHaveProperty('id')
        expect(doc).toHaveProperty('name')
        expect(doc).toHaveProperty('category')
        expect(doc).toHaveProperty('status')
        expect(doc).toHaveProperty('metadata')
        expect(doc.metadata).toHaveProperty('size')
        expect(doc.metadata).toHaveProperty('mimeType')
        expect(doc.metadata).toHaveProperty('uploadedBy')
        expect(doc.metadata).toHaveProperty('uploadedAt')
      })
    })

    it('documents have valid categories', () => {
      const validCategories = Object.values(DocumentCategory)
      mockDocuments.forEach((doc) => {
        expect(validCategories).toContain(doc.category)
      })
    })

    it('documents have valid statuses', () => {
      const validStatuses = Object.values(DocumentStatus)
      mockDocuments.forEach((doc) => {
        expect(validStatuses).toContain(doc.status)
      })
    })
  })

  describe('filterDocuments', () => {
    const testDocuments: Document[] = [
      {
        id: '1',
        name: 'Test Document.pdf',
        category: DocumentCategory.EDUCATIONAL,
        status: DocumentStatus.READY,
        metadata: {
          size: 1000,
          mimeType: 'application/pdf',
          uploadedBy: 'User A',
          uploadedAt: new Date('2024-01-15'),
        },
        description: 'Test description',
        tags: ['test', 'example'],
      },
      {
        id: '2',
        name: 'Another File.docx',
        category: DocumentCategory.METHODICAL,
        status: DocumentStatus.PROCESSING,
        metadata: {
          size: 2000,
          mimeType: 'application/docx',
          uploadedBy: 'User B',
          uploadedAt: new Date('2024-02-20'),
        },
        description: 'Another description',
        tags: ['sample'],
        authorId: 1,
      },
      {
        id: '3',
        name: 'Third Doc.xlsx',
        category: DocumentCategory.EDUCATIONAL,
        status: DocumentStatus.READY,
        metadata: {
          size: 3000,
          mimeType: 'application/xlsx',
          uploadedBy: 'User C',
          uploadedAt: new Date('2024-03-10'),
        },
        tags: ['test', 'data'],
        authorId: 2,
      },
    ]

    it('returns all documents when no filters applied', () => {
      const result = filterDocuments(testDocuments, {})
      expect(result).toHaveLength(3)
    })

    it('filters by search term in name', () => {
      const result = filterDocuments(testDocuments, { search: 'Document' })
      expect(result).toHaveLength(1)
      expect(result[0].name).toBe('Test Document.pdf')
    })

    it('filters by search term in description', () => {
      const result = filterDocuments(testDocuments, { search: 'Another' })
      expect(result).toHaveLength(1)
      expect(result[0].id).toBe('2')
    })

    it('filters by search term in tags', () => {
      const result = filterDocuments(testDocuments, { search: 'sample' })
      expect(result).toHaveLength(1)
      expect(result[0].id).toBe('2')
    })

    it('search is case insensitive', () => {
      const result = filterDocuments(testDocuments, { search: 'DOCUMENT' })
      expect(result).toHaveLength(1)
      expect(result[0].name).toBe('Test Document.pdf')
    })

    it('filters by category', () => {
      const result = filterDocuments(testDocuments, {
        category: DocumentCategory.EDUCATIONAL,
      })
      expect(result).toHaveLength(2)
      expect(result.every((doc) => doc.category === DocumentCategory.EDUCATIONAL)).toBe(true)
    })

    it('filters by status', () => {
      const result = filterDocuments(testDocuments, {
        status: DocumentStatus.PROCESSING,
      })
      expect(result).toHaveLength(1)
      expect(result[0].status).toBe(DocumentStatus.PROCESSING)
    })

    it('filters by tags', () => {
      const result = filterDocuments(testDocuments, { tags: ['test'] })
      expect(result).toHaveLength(2)
    })

    it('filters by multiple tags', () => {
      const result = filterDocuments(testDocuments, { tags: ['example', 'sample'] })
      expect(result).toHaveLength(2) // Documents with either tag
    })

    it('tag filtering is case insensitive', () => {
      const result = filterDocuments(testDocuments, { tags: ['TEST'] })
      expect(result).toHaveLength(2)
    })

    it('filters by dateFrom', () => {
      const result = filterDocuments(testDocuments, {
        dateFrom: new Date('2024-02-01'),
      })
      expect(result).toHaveLength(2)
      expect(result.find((doc) => doc.id === '1')).toBeUndefined()
    })

    it('filters by dateTo', () => {
      const result = filterDocuments(testDocuments, {
        dateTo: new Date('2024-02-28'),
      })
      expect(result).toHaveLength(2)
      expect(result.find((doc) => doc.id === '3')).toBeUndefined()
    })

    it('filters by date range', () => {
      const result = filterDocuments(testDocuments, {
        dateFrom: new Date('2024-02-01'),
        dateTo: new Date('2024-02-28'),
      })
      expect(result).toHaveLength(1)
      expect(result[0].id).toBe('2')
    })

    it('filters by authorId', () => {
      const result = filterDocuments(testDocuments, { authorId: 1 })
      expect(result).toHaveLength(1)
      expect(result[0].authorId).toBe(1)
    })

    it('combines multiple filters', () => {
      const result = filterDocuments(testDocuments, {
        category: DocumentCategory.EDUCATIONAL,
        status: DocumentStatus.READY,
      })
      expect(result).toHaveLength(2)
    })

    it('returns empty array when no matches', () => {
      const result = filterDocuments(testDocuments, { search: 'nonexistent' })
      expect(result).toHaveLength(0)
    })

    it('handles documents without description', () => {
      const docsWithoutDesc: Document[] = [
        {
          id: '1',
          name: 'No Description.pdf',
          category: DocumentCategory.EDUCATIONAL,
          status: DocumentStatus.READY,
          metadata: {
            size: 1000,
            mimeType: 'application/pdf',
            uploadedBy: 'User',
            uploadedAt: new Date(),
          },
        },
      ]
      const result = filterDocuments(docsWithoutDesc, { search: 'test' })
      expect(result).toHaveLength(0)
    })

    it('handles documents without tags', () => {
      const docsWithoutTags: Document[] = [
        {
          id: '1',
          name: 'No Tags.pdf',
          category: DocumentCategory.EDUCATIONAL,
          status: DocumentStatus.READY,
          metadata: {
            size: 1000,
            mimeType: 'application/pdf',
            uploadedBy: 'User',
            uploadedAt: new Date(),
          },
        },
      ]
      const result = filterDocuments(docsWithoutTags, { tags: ['test'] })
      expect(result).toHaveLength(0)
    })

    it('handles empty tags filter', () => {
      const result = filterDocuments(testDocuments, { tags: [] })
      expect(result).toHaveLength(3)
    })
  })

  describe('sortDocuments', () => {
    const testDocuments: Document[] = [
      {
        id: '1',
        name: 'Beta Document.pdf',
        category: DocumentCategory.EDUCATIONAL,
        status: DocumentStatus.READY,
        metadata: {
          size: 2000,
          mimeType: 'application/pdf',
          uploadedBy: 'User',
          uploadedAt: new Date('2024-02-15'),
          modifiedAt: new Date('2024-03-01'),
        },
      },
      {
        id: '2',
        name: 'Alpha File.docx',
        category: DocumentCategory.METHODICAL,
        status: DocumentStatus.READY,
        metadata: {
          size: 1000,
          mimeType: 'application/docx',
          uploadedBy: 'User',
          uploadedAt: new Date('2024-01-10'),
        },
      },
      {
        id: '3',
        name: 'Gamma Report.xlsx',
        category: DocumentCategory.ADMINISTRATIVE,
        status: DocumentStatus.READY,
        metadata: {
          size: 3000,
          mimeType: 'application/xlsx',
          uploadedBy: 'User',
          uploadedAt: new Date('2024-03-20'),
          modifiedAt: new Date('2024-03-25'),
        },
      },
    ]

    it('sorts by name ascending', () => {
      const result = sortDocuments(testDocuments, 'name', 'asc')
      expect(result[0].name).toBe('Alpha File.docx')
      expect(result[1].name).toBe('Beta Document.pdf')
      expect(result[2].name).toBe('Gamma Report.xlsx')
    })

    it('sorts by name descending', () => {
      const result = sortDocuments(testDocuments, 'name', 'desc')
      expect(result[0].name).toBe('Gamma Report.xlsx')
      expect(result[1].name).toBe('Beta Document.pdf')
      expect(result[2].name).toBe('Alpha File.docx')
    })

    it('sorts by uploadedAt ascending', () => {
      const result = sortDocuments(testDocuments, 'uploadedAt', 'asc')
      expect(result[0].id).toBe('2') // Jan 10
      expect(result[1].id).toBe('1') // Feb 15
      expect(result[2].id).toBe('3') // Mar 20
    })

    it('sorts by uploadedAt descending', () => {
      const result = sortDocuments(testDocuments, 'uploadedAt', 'desc')
      expect(result[0].id).toBe('3') // Mar 20
      expect(result[1].id).toBe('1') // Feb 15
      expect(result[2].id).toBe('2') // Jan 10
    })

    it('sorts by modifiedAt ascending', () => {
      const result = sortDocuments(testDocuments, 'modifiedAt', 'asc')
      // Doc 2 has no modifiedAt, so uses uploadedAt (Jan 10)
      expect(result[0].id).toBe('2')
      expect(result[1].id).toBe('1') // Mar 1
      expect(result[2].id).toBe('3') // Mar 25
    })

    it('sorts by modifiedAt descending', () => {
      const result = sortDocuments(testDocuments, 'modifiedAt', 'desc')
      expect(result[0].id).toBe('3') // Mar 25
      expect(result[1].id).toBe('1') // Mar 1
      expect(result[2].id).toBe('2') // Jan 10 (no modifiedAt)
    })

    it('sorts by size ascending', () => {
      const result = sortDocuments(testDocuments, 'size', 'asc')
      expect(result[0].metadata.size).toBe(1000)
      expect(result[1].metadata.size).toBe(2000)
      expect(result[2].metadata.size).toBe(3000)
    })

    it('sorts by size descending', () => {
      const result = sortDocuments(testDocuments, 'size', 'desc')
      expect(result[0].metadata.size).toBe(3000)
      expect(result[1].metadata.size).toBe(2000)
      expect(result[2].metadata.size).toBe(1000)
    })

    it('does not modify original array', () => {
      const originalIds = testDocuments.map((d) => d.id)
      sortDocuments(testDocuments, 'name', 'asc')
      const afterIds = testDocuments.map((d) => d.id)
      expect(afterIds).toEqual(originalIds)
    })

    it('handles empty array', () => {
      const result = sortDocuments([], 'name', 'asc')
      expect(result).toEqual([])
    })

    it('handles single item array', () => {
      const singleDoc = [testDocuments[0]]
      const result = sortDocuments(singleDoc, 'name', 'asc')
      expect(result).toHaveLength(1)
      expect(result[0].id).toBe('1')
    })
  })
})
