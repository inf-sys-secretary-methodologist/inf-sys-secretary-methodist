import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DocumentPreview } from '../DocumentPreview'
import { Document, DocumentCategory, DocumentStatus } from '@/types/document'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: (namespace: string) => (key: string, params?: Record<string, unknown>) => {
    const translations: Record<string, Record<string, string>> = {
      common: {
        'fileSize.bytes': `${params?.size || 0} bytes`,
        'fileSize.kb': `${params?.size || 0} KB`,
        'fileSize.mb': `${params?.size || 0} MB`,
      },
      documents: {
        'filters.category': 'Category',
        'filters.status': 'Status',
        'filters.size': 'Size',
        'filters.author': 'Author',
        'statuses.ready': 'Ready',
        'statuses.uploading': 'Uploading',
        'statuses.processing': 'Processing',
        'statuses.error': 'Error',
        'categories.educational': 'Educational',
        'categories.administrative': 'Administrative',
      },
      documentPreview: {
        viewTab: 'Preview',
        historyTab: 'Versions',
        download: 'Download',
        open: 'Open in new tab',
        previewUnavailable: 'Preview not available',
        previewUnavailableHint: 'This file type cannot be previewed',
        downloadToView: 'Download to view',
        version: 'Version',
        modified: 'Modified',
        description: 'Description',
        tags: 'Tags',
      },
    }
    return translations[namespace]?.[key] || key
  },
  useLocale: () => 'en',
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
}

describe('DocumentPreview', () => {
  const defaultProps = {
    document: mockDocument,
    onClose: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders preview modal', () => {
    render(<DocumentPreview {...defaultProps} />)
    expect(screen.getByText('Test Document.pdf')).toBeInTheDocument()
  })

  it('displays document name', () => {
    render(<DocumentPreview {...defaultProps} />)
    expect(screen.getByText('Test Document.pdf')).toBeInTheDocument()
  })

  it('renders close button', () => {
    render(<DocumentPreview {...defaultProps} />)
    const closeButtons = screen.getAllByRole('button')
    expect(closeButtons.length).toBeGreaterThan(0)
  })

  it('calls onClose when escape key is pressed', () => {
    const onClose = jest.fn()
    render(<DocumentPreview {...defaultProps} onClose={onClose} />)

    fireEvent.keyDown(document, { key: 'Escape' })
    expect(onClose).toHaveBeenCalled()
  })

  it('renders download button when onDownload is provided', () => {
    const onDownload = jest.fn()
    render(<DocumentPreview {...defaultProps} onDownload={onDownload} />)

    expect(screen.getByText('Download')).toBeInTheDocument()
  })

  it('calls onDownload when download button is clicked', async () => {
    const user = userEvent.setup()
    const onDownload = jest.fn()
    render(<DocumentPreview {...defaultProps} onDownload={onDownload} />)

    await user.click(screen.getByText('Download'))
    expect(onDownload).toHaveBeenCalled()
  })

  it('applies custom className', () => {
    const { container } = render(<DocumentPreview {...defaultProps} className="custom-class" />)
    expect(container.querySelector('.custom-class')).toBeInTheDocument()
  })

  it('renders tabs for preview and versions', () => {
    render(<DocumentPreview {...defaultProps} />)
    expect(screen.getByText('Preview')).toBeInTheDocument()
    expect(screen.getByText('Versions')).toBeInTheDocument()
  })

  it('handles image document type', () => {
    const imageDoc: Document = {
      ...mockDocument,
      metadata: {
        ...mockDocument.metadata,
        mimeType: 'image/jpeg',
      },
    }
    render(<DocumentPreview {...defaultProps} document={imageDoc} />)
    expect(screen.getByText('Test Document.pdf')).toBeInTheDocument()
  })

  it('shows description when present', () => {
    render(<DocumentPreview {...defaultProps} />)
    expect(screen.getByText('Test description')).toBeInTheDocument()
  })
})
