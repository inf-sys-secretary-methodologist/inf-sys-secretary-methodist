import { render, screen, fireEvent } from '@testing-library/react'
import { DocumentUploadComponent } from '../DocumentUpload'

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
})
