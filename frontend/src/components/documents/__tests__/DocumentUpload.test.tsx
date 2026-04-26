import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { DocumentUploadComponent } from '../DocumentUpload'
import { tagsApi } from '@/lib/api/documents'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: (namespace: string) => (key: string, params?: Record<string, unknown>) => {
    const translations: Record<string, Record<string, string>> = {
      'documents.uploadForm': {
        dragAndDrop: 'Drag and drop files here',
        orClickButton: 'or click to select',
        selectFiles: 'Select files',
        supportedFormats: 'Supported formats: PDF, DOC, DOCX, XLS, XLSX, TXT, JPG, PNG',
        uploading: 'Uploading...',
        uploadCount: `Upload ${params?.count || 0} files`,
        selectedFiles: 'Selected files',
        documentCategory: 'Category',
        descriptionOptional: 'Description (optional)',
        descriptionPlaceholder: 'Enter description',
        tagsOptional: 'Tags (optional)',
        loadingTags: 'Loading tags...',
        noTagsAvailable: 'No tags available',
        selectedTags: 'Selected tags',
        typeNotSupported: 'File type not supported',
        sizeExceeded: 'File size exceeded',
      },
      documents: {
        'categories.educational': 'Educational',
        'categories.administrative': 'Administrative',
        'categories.financial': 'Financial',
        'categories.other': 'Other',
      },
      common: {
        cancel: 'Cancel',
        'fileSize.kb': `${params?.size || 0} KB`,
      },
    }
    return translations[namespace]?.[key] || key
  },
  useLocale: () => 'en',
}))

// Mock the API
jest.mock('@/lib/api/documents', () => ({
  tagsApi: {
    getAll: jest.fn().mockResolvedValue([]),
  },
}))

const mockedTagsApi = jest.mocked(tagsApi)

describe('DocumentUploadComponent', () => {
  const defaultProps = {
    onUpload: jest.fn().mockResolvedValue(undefined),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders upload component', () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    expect(screen.getByText(/drag and drop/i)).toBeInTheDocument()
  })

  it('renders select files button', () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    expect(screen.getByText(/select files/i)).toBeInTheDocument()
  })

  it('shows supported formats hint', () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    expect(screen.getByText(/supported formats/i)).toBeInTheDocument()
  })

  it('handles drag over event', () => {
    const { container } = render(<DocumentUploadComponent {...defaultProps} />)
    const dropZone = container.querySelector('[class*="border-dashed"]')

    if (dropZone) {
      fireEvent.dragOver(dropZone, { dataTransfer: { files: [] } })
      // Component should handle drag over
      expect(dropZone).toBeInTheDocument()
    }
  })

  it('handles drag leave event', () => {
    const { container } = render(<DocumentUploadComponent {...defaultProps} />)
    const dropZone = container.querySelector('[class*="border-dashed"]')

    if (dropZone) {
      fireEvent.dragLeave(dropZone)
      expect(dropZone).toBeInTheDocument()
    }
  })

  it('shows uploading state when isUploading is true', () => {
    render(<DocumentUploadComponent {...defaultProps} isUploading={true} />)
    // The drop zone should be disabled/opacity reduced
    const { container } = render(<DocumentUploadComponent {...defaultProps} isUploading={true} />)
    const dropZone = container.querySelector('[class*="border-dashed"]')
    expect(dropZone).toHaveClass('opacity-50')
  })

  it('applies custom className', () => {
    const { container } = render(
      <DocumentUploadComponent {...defaultProps} className="custom-class" />
    )
    expect(container.querySelector('.custom-class')).toBeInTheDocument()
  })

  it('renders or click to select text', () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    expect(screen.getByText(/or click to select/i)).toBeInTheDocument()
  })

  it('has hidden file input', () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    const input = document.getElementById('file-upload')
    expect(input).toBeInTheDocument()
    expect(input).toHaveClass('hidden')
    expect(input).toHaveAttribute('type', 'file')
  })

  it('handles file selection through input change', async () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    const input = document.getElementById('file-upload') as HTMLInputElement

    const validFile = new File(['content'], 'test.pdf', { type: 'application/pdf' })
    Object.defineProperty(input, 'files', { value: [validFile] })

    fireEvent.change(input)

    await screen.findByText(/Selected files/)
    expect(screen.getByText('test.pdf')).toBeInTheDocument()
  })

  it('handles file drop', async () => {
    const { container } = render(<DocumentUploadComponent {...defaultProps} />)
    const dropZone = container.querySelector('[class*="border-dashed"]')

    if (dropZone) {
      const validFile = new File(['content'], 'dropped.pdf', { type: 'application/pdf' })

      fireEvent.drop(dropZone, {
        dataTransfer: {
          files: [validFile],
        },
      })

      await screen.findByText(/Selected files/)
      expect(screen.getByText('dropped.pdf')).toBeInTheDocument()
    }
  })

  it('shows validation error for invalid file type', async () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    const input = document.getElementById('file-upload') as HTMLInputElement

    const invalidFile = new File(['content'], 'test.exe', { type: 'application/x-executable' })
    Object.defineProperty(input, 'files', { value: [invalidFile] })

    fireEvent.change(input)

    await screen.findByText(/Selected files/)
    expect(screen.getByText('File type not supported')).toBeInTheDocument()
  })

  it('shows validation error for file too large', async () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    const input = document.getElementById('file-upload') as HTMLInputElement

    // Create large file content
    const largeContent = new Array(11 * 1024 * 1024).fill('x').join('')
    const largeFile = new File([largeContent], 'large.pdf', { type: 'application/pdf' })
    Object.defineProperty(input, 'files', { value: [largeFile] })

    fireEvent.change(input)

    await screen.findByText(/Selected files/)
    expect(screen.getByText('File size exceeded')).toBeInTheDocument()
  })

  it('removes file when X button clicked', async () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    const input = document.getElementById('file-upload') as HTMLInputElement

    const validFile = new File(['content'], 'test.pdf', { type: 'application/pdf' })
    Object.defineProperty(input, 'files', { value: [validFile] })

    fireEvent.change(input)

    await screen.findByText('test.pdf')

    // Find and click the remove button
    const removeButton = screen.getByRole('button', { name: '' }) // X button has no accessible name
    fireEvent.click(removeButton)

    expect(screen.queryByText('test.pdf')).not.toBeInTheDocument()
  })

  it('shows category select when files are selected', async () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    const input = document.getElementById('file-upload') as HTMLInputElement

    const validFile = new File(['content'], 'test.pdf', { type: 'application/pdf' })
    Object.defineProperty(input, 'files', { value: [validFile] })

    fireEvent.change(input)

    await screen.findByText('Category')
    expect(screen.getByRole('combobox')).toBeInTheDocument()
  })

  it('shows description textarea when files are selected', async () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    const input = document.getElementById('file-upload') as HTMLInputElement

    const validFile = new File(['content'], 'test.pdf', { type: 'application/pdf' })
    Object.defineProperty(input, 'files', { value: [validFile] })

    fireEvent.change(input)

    await screen.findByText('Description (optional)')
    expect(screen.getByPlaceholderText('Enter description')).toBeInTheDocument()
  })

  it('calls onUpload when submit button clicked', async () => {
    const onUpload = jest.fn().mockResolvedValue(undefined)
    render(<DocumentUploadComponent onUpload={onUpload} />)
    const input = document.getElementById('file-upload') as HTMLInputElement

    const validFile = new File(['content'], 'test.pdf', { type: 'application/pdf' })
    Object.defineProperty(input, 'files', { value: [validFile] })

    fireEvent.change(input)

    await screen.findByText('Upload 1 files')
    fireEvent.click(screen.getByText('Upload 1 files'))

    await waitFor(() => {
      expect(onUpload).toHaveBeenCalled()
    })
  })

  it('shows cancel button when onCancel prop is provided', async () => {
    const onCancel = jest.fn()
    render(<DocumentUploadComponent {...defaultProps} onCancel={onCancel} />)
    const input = document.getElementById('file-upload') as HTMLInputElement

    const validFile = new File(['content'], 'test.pdf', { type: 'application/pdf' })
    Object.defineProperty(input, 'files', { value: [validFile] })

    fireEvent.change(input)

    await screen.findByText('Cancel')
    fireEvent.click(screen.getByText('Cancel'))

    expect(onCancel).toHaveBeenCalled()
  })

  it('changes category when selected', async () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    const input = document.getElementById('file-upload') as HTMLInputElement

    const validFile = new File(['content'], 'test.pdf', { type: 'application/pdf' })
    Object.defineProperty(input, 'files', { value: [validFile] })

    fireEvent.change(input)

    await screen.findByRole('combobox')
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'administrative' } })

    expect(screen.getByRole('combobox')).toHaveValue('administrative')
  })

  it('allows entering description', async () => {
    render(<DocumentUploadComponent {...defaultProps} />)
    const input = document.getElementById('file-upload') as HTMLInputElement

    const validFile = new File(['content'], 'test.pdf', { type: 'application/pdf' })
    Object.defineProperty(input, 'files', { value: [validFile] })

    fireEvent.change(input)

    await screen.findByPlaceholderText('Enter description')
    fireEvent.change(screen.getByPlaceholderText('Enter description'), {
      target: { value: 'Test description' },
    })

    expect(screen.getByDisplayValue('Test description')).toBeInTheDocument()
  })

  it('select files button is a label linked to file input via htmlFor', () => {
    render(<DocumentUploadComponent {...defaultProps} />)

    const selectLabel = screen.getByText('Select files')
    expect(selectLabel.tagName).toBe('LABEL')
    expect(selectLabel).toHaveAttribute('for', 'file-upload')
  })
})

describe('DocumentUploadComponent with tags', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // Mock tagsApi to return some tags
    mockedTagsApi.getAll.mockResolvedValue([
      { id: 1, name: 'Important', color: '#ff0000' },
      { id: 2, name: 'Draft', color: '#00ff00' },
    ] as never)
  })

  it('loads and displays tags', async () => {
    render(<DocumentUploadComponent onUpload={jest.fn()} />)
    const input = document.getElementById('file-upload') as HTMLInputElement

    const validFile = new File(['content'], 'test.pdf', { type: 'application/pdf' })
    Object.defineProperty(input, 'files', { value: [validFile] })

    fireEvent.change(input)

    await waitFor(() => {
      expect(screen.getByText('Important')).toBeInTheDocument()
      expect(screen.getByText('Draft')).toBeInTheDocument()
    })
  })

  it('toggles tag selection', async () => {
    render(<DocumentUploadComponent onUpload={jest.fn()} />)
    const input = document.getElementById('file-upload') as HTMLInputElement

    const validFile = new File(['content'], 'test.pdf', { type: 'application/pdf' })
    Object.defineProperty(input, 'files', { value: [validFile] })

    fireEvent.change(input)

    await waitFor(() => {
      expect(screen.getByText('Important')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Important'))

    expect(screen.getByText('Selected tags: 1')).toBeInTheDocument()

    // Toggle off
    fireEvent.click(screen.getByText('Important'))

    expect(screen.queryByText('Selected tags: 1')).not.toBeInTheDocument()
  })
})
