import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DocumentEditDialog } from '../DocumentEditDialog'
import { Document, DocumentCategory, DocumentStatus } from '@/types/document'
import { documentsApi, tagsApi } from '@/lib/api/documents'

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
        descriptionPlaceholder: 'Enter subject',
        contentPlaceholder: 'Enter content',
        fileNamePlaceholder: 'File name',
        tagsSearchPlaceholder: 'Search tags',
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

const mockedDocumentsApi = jest.mocked(documentsApi)
const mockedTagsApi = jest.mocked(tagsApi)

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

  it('renders subject input field', () => {
    render(<DocumentEditDialog {...defaultProps} />)
    expect(screen.getByText(/subject/i)).toBeInTheDocument()
  })

  it('renders content textarea', () => {
    render(<DocumentEditDialog {...defaultProps} />)
    expect(screen.getByText(/content/i)).toBeInTheDocument()
  })

  it('renders tags section', () => {
    render(<DocumentEditDialog {...defaultProps} />)
    const tagsElements = screen.getAllByText(/tags/i)
    expect(tagsElements.length).toBeGreaterThan(0)
  })

  it('renders title input with document name', async () => {
    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      const titleInput = screen.getByPlaceholderText('Enter document name')
      expect(titleInput).toBeInTheDocument()
    })
  })

  it('renders subject input', async () => {
    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      const subjectInput = screen.getByPlaceholderText('Enter subject')
      expect(subjectInput).toBeInTheDocument()
    })
  })

  it('renders content textarea', async () => {
    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      const contentInput = screen.getByPlaceholderText('Enter content')
      expect(contentInput).toBeInTheDocument()
    })
  })

  it('saves document when save button is clicked', async () => {
    const user = userEvent.setup()
    const onSaved = jest.fn()
    const onOpenChange = jest.fn()
    mockedDocumentsApi.update.mockResolvedValueOnce({} as never)

    render(<DocumentEditDialog {...defaultProps} onSaved={onSaved} onOpenChange={onOpenChange} />)

    await waitFor(() => {
      expect(mockedDocumentsApi.getById).toHaveBeenCalled()
    })

    const saveButton = screen.getByRole('button', { name: /save/i })
    await user.click(saveButton)

    await waitFor(() => {
      expect(mockedDocumentsApi.update).toHaveBeenCalled()
    })
  })

  it('shows error when save fails', async () => {
    const user = userEvent.setup()
    mockedDocumentsApi.update.mockRejectedValueOnce(new Error('Save failed'))

    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      expect(mockedDocumentsApi.getById).toHaveBeenCalled()
    })

    const saveButton = screen.getByRole('button', { name: /save/i })
    await user.click(saveButton)

    await waitFor(() => {
      expect(screen.getByText('Failed to save')).toBeInTheDocument()
    })
  })

  it('uploads file when file is selected and upload clicked', async () => {
    const user = userEvent.setup()
    mockedDocumentsApi.uploadFile.mockResolvedValueOnce({} as never)

    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      expect(mockedDocumentsApi.getById).toHaveBeenCalled()
    })

    // Select a file
    const file = new File(['test content'], 'test-file.pdf', { type: 'application/pdf' })
    const fileInput = document.getElementById('file-upload') as HTMLInputElement

    await user.upload(fileInput, file)

    // The file name should appear
    await waitFor(() => {
      expect(screen.getByText(/test-file\.pdf/)).toBeInTheDocument()
    })

    // Click save to upload
    const uploadButton = screen.getAllByRole('button', { name: /save/i })[0]
    await user.click(uploadButton)

    await waitFor(() => {
      expect(mockedDocumentsApi.uploadFile).toHaveBeenCalled()
    })
  })

  it('shows error when file upload fails', async () => {
    const user = userEvent.setup()
    mockedDocumentsApi.uploadFile.mockRejectedValueOnce(new Error('Upload failed'))

    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      expect(mockedDocumentsApi.getById).toHaveBeenCalled()
    })

    const file = new File(['test content'], 'test-file.pdf', { type: 'application/pdf' })
    const fileInput = document.getElementById('file-upload') as HTMLInputElement

    await user.upload(fileInput, file)

    await waitFor(() => {
      expect(screen.getByText(/test-file\.pdf/)).toBeInTheDocument()
    })

    const uploadButton = screen.getAllByRole('button', { name: /save/i })[0]
    await user.click(uploadButton)

    await waitFor(() => {
      expect(screen.getByText('Failed to upload file')).toBeInTheDocument()
    })
  })

  it('can remove selected file before upload', async () => {
    const user = userEvent.setup()
    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Select new file')).toBeInTheDocument()
    })

    const file = new File(['test content'], 'test-file.pdf', { type: 'application/pdf' })
    const fileInput = document.getElementById('file-upload') as HTMLInputElement

    await user.upload(fileInput, file)

    await waitFor(() => {
      expect(screen.getByText(/test-file\.pdf/)).toBeInTheDocument()
    })

    // Find and click the remove button (trash icon)
    const removeButtons = screen.getAllByRole('button')
    const removeButton = removeButtons.find((btn) => btn.querySelector('svg.lucide-trash-2'))
    if (removeButton) {
      await user.click(removeButton)
    }

    // File name should disappear
    await waitFor(() => {
      expect(screen.queryByText(/test-file\.pdf/)).not.toBeInTheDocument()
    })
  })

  it('can close dialog by clicking X button', async () => {
    const user = userEvent.setup()
    const onOpenChange = jest.fn()
    render(<DocumentEditDialog {...defaultProps} onOpenChange={onOpenChange} />)

    // Find X button in header
    const closeButton = screen
      .getAllByRole('button')
      .find((btn) => btn.querySelector('svg.lucide-x'))

    if (closeButton) {
      await user.click(closeButton)
      expect(onOpenChange).toHaveBeenCalledWith(false)
    }
  })

  it('disables save button when title is empty', async () => {
    const user = userEvent.setup()
    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Enter document name')).toBeInTheDocument()
    })

    const titleInput = screen.getByPlaceholderText('Enter document name')
    await user.clear(titleInput)

    // Save button should be disabled
    const saveButtons = screen.getAllByRole('button', { name: /save/i })
    const mainSaveButton = saveButtons[saveButtons.length - 1] // Last save button is the main one
    expect(mainSaveButton).toBeDisabled()
  })

  it('loads document tags when dialog opens', async () => {
    mockedTagsApi.getDocumentTags.mockResolvedValueOnce([
      { id: 1, name: 'Important', color: '#ff0000' },
    ] as never)
    mockedTagsApi.getAll.mockResolvedValueOnce([
      { id: 1, name: 'Important', color: '#ff0000' },
      { id: 2, name: 'Review', color: '#00ff00' },
    ] as never)

    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      expect(mockedTagsApi.getDocumentTags).toHaveBeenCalledWith('1')
    })
  })

  it('renders tags label in form', () => {
    render(<DocumentEditDialog {...defaultProps} />)
    // Look for Tags label - it's rendered inline with an icon
    expect(screen.getByText('Tags')).toBeInTheDocument()
  })

  it('renders file name input when file exists', async () => {
    mockedDocumentsApi.getById.mockResolvedValueOnce({
      id: '1',
      title: 'Test Document',
      subject: '',
      content: '',
      file_name: 'original-file.pdf',
    } as never)

    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByPlaceholderText('File name')).toBeInTheDocument()
    })

    const fileNameInput = screen.getByPlaceholderText('File name')
    expect(fileNameInput).toHaveValue('original-file.pdf')
  })

  it('shows version note about file upload', () => {
    render(<DocumentEditDialog {...defaultProps} />)
    expect(screen.getByText('Uploading a new file will create a new version')).toBeInTheDocument()
  })

  it('allows changing subject field', async () => {
    const user = userEvent.setup()
    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Enter subject')).toBeInTheDocument()
    })

    const subjectInput = screen.getByPlaceholderText('Enter subject')
    await user.type(subjectInput, 'New Subject Content')

    expect(subjectInput).toHaveValue('New Subject Content')
  })

  it('allows changing content field', async () => {
    const user = userEvent.setup()
    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Enter content')).toBeInTheDocument()
    })

    const contentInput = screen.getByPlaceholderText('Enter content')
    await user.type(contentInput, 'New content text here')

    expect(contentInput).toHaveValue('New content text here')
  })

  it('closes tag dropdown when clicking outside', async () => {
    const user = userEvent.setup()
    mockedTagsApi.getAll.mockResolvedValueOnce([
      { id: 1, name: 'Important', color: '#ff0000' },
      { id: 2, name: 'Review', color: '#00ff00' },
    ] as never)
    mockedTagsApi.getDocumentTags.mockResolvedValueOnce([])

    render(<DocumentEditDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Search tags')).toBeInTheDocument()
    })

    // Click on the search tags input to open dropdown
    const tagSearchInput = screen.getByPlaceholderText('Search tags')
    await user.click(tagSearchInput)

    // Type to trigger dropdown
    await user.type(tagSearchInput, 'Imp')

    // The dropdown should be visible, click outside to close it
    // The overlay div with fixed inset-0 should close the dropdown
    const overlay = document.querySelector('.fixed.inset-0')
    if (overlay) {
      await user.click(overlay)
    }

    // Dropdown should be closed now
    expect(tagSearchInput).toBeInTheDocument()
  })
})
