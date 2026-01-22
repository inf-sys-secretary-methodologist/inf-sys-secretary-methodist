import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DocumentEditDialog } from '../DocumentEditDialog'
import { Document, DocumentCategory, DocumentStatus } from '@/types/document'

// Mock ResizeObserver
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}))

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: (namespace: string) => (key: string, params?: Record<string, unknown>) => {
    const translations: Record<string, Record<string, string>> = {
      'documents.edit': {
        title: 'Edit Document',
        titleLabel: 'Title',
        subjectLabel: 'Subject',
        contentLabel: 'Content',
        attachedFile: 'Attached file',
        selectNewFile: 'Select new file',
        versionNote: 'Uploading a new file will create a new version',
        saveError: 'Failed to save',
        tagAddError: 'Failed to add tag',
        tagRemoveError: 'Failed to remove tag',
        fileUploadedSuccess: `File ${params?.name || ''} uploaded`,
        fileUploadError: 'Failed to upload file',
        tagsLabel: 'Tags',
        searchTags: 'Search tags',
      },
      'documents.form': {
        documentNamePlaceholder: 'Enter document name',
        subjectPlaceholder: 'Enter subject',
        contentPlaceholder: 'Enter content',
        fileNamePlaceholder: 'File name',
      },
      common: {
        save: 'Save',
        cancel: 'Cancel',
        saving: 'Saving...',
        'fileSize.bytes': `${params?.size || 0} bytes`,
        'fileSize.kb': `${params?.size || 0} KB`,
        'fileSize.mb': `${params?.size || 0} MB`,
      },
    }
    return translations[namespace]?.[key] || key
  },
  useLocale: () => 'en',
}))

// Mock the API
jest.mock('@/lib/api/documents', () => ({
  documentsApi: {
    getById: jest.fn().mockResolvedValue({
      id: '1',
      title: 'Test Document',
      subject: '',
      content: '',
      file_name: 'test.pdf',
    }),
    update: jest.fn(),
    uploadFile: jest.fn(),
  },
  tagsApi: {
    getAll: jest.fn().mockResolvedValue([]),
    getDocumentTags: jest.fn().mockResolvedValue([]),
    addTagToDocument: jest.fn(),
    removeTagFromDocument: jest.fn(),
  },
}))

const mockDocument: Document = {
  id: '1',
  name: 'Test Document.pdf',
  category: DocumentCategory.EDUCATIONAL,
  status: DocumentStatus.READY,
  metadata: {
    size: 1024 * 1024,
    mimeType: 'application/pdf',
    uploadedBy: 'Test User',
    uploadedAt: new Date('2024-06-15'),
  },
  url: '/files/test.pdf',
  description: 'Test description',
  tags: ['test', 'document'],
}

describe('DocumentEditDialog', () => {
  const defaultProps = {
    document: mockDocument,
    open: true,
    onOpenChange: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders edit dialog when open', () => {
    render(<DocumentEditDialog {...defaultProps} />)
    expect(screen.getByText('Edit Document')).toBeInTheDocument()
  })

  it('does not render when open is false', () => {
    render(<DocumentEditDialog {...defaultProps} open={false} />)
    expect(screen.queryByText('Edit Document')).not.toBeInTheDocument()
  })

  it('does not render when document is null', () => {
    render(<DocumentEditDialog {...defaultProps} document={null} />)
    expect(screen.queryByText('Edit Document')).not.toBeInTheDocument()
  })

  it('renders title label', () => {
    render(<DocumentEditDialog {...defaultProps} />)
    expect(screen.getByText(/title/i)).toBeInTheDocument()
  })

  it('renders save button', () => {
    render(<DocumentEditDialog {...defaultProps} />)
    const saveButton = screen.getByRole('button', { name: /save/i })
    expect(saveButton).toBeInTheDocument()
  })

  it('renders cancel button', () => {
    render(<DocumentEditDialog {...defaultProps} />)
    const cancelButton = screen.getByRole('button', { name: /cancel/i })
    expect(cancelButton).toBeInTheDocument()
  })

  it('calls onOpenChange when closed', async () => {
    const user = userEvent.setup()
    const onOpenChange = jest.fn()
    render(<DocumentEditDialog {...defaultProps} onOpenChange={onOpenChange} />)

    // Click the Cancel button to close
    const cancelButton = screen.getByRole('button', { name: /cancel/i })
    await user.click(cancelButton)
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })

  it('loads document details when opened', async () => {
    render(<DocumentEditDialog {...defaultProps} />)

    const { documentsApi } = await import('@/lib/api/documents')
    await waitFor(() => {
      expect(documentsApi.getById).toHaveBeenCalledWith('1')
    })
  })

  it('renders file upload section', () => {
    render(<DocumentEditDialog {...defaultProps} />)
    expect(screen.getByText('Attached file')).toBeInTheDocument()
  })

  it('renders select new file button', () => {
    render(<DocumentEditDialog {...defaultProps} />)
    expect(screen.getByText('Select new file')).toBeInTheDocument()
  })
})
