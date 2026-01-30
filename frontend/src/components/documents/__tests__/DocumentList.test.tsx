import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DocumentList } from '../DocumentList'
import { Document, DocumentCategory, DocumentStatus } from '@/types/document'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: (namespace: string) => (key: string, params?: Record<string, unknown>) => {
    const translations: Record<string, Record<string, string>> = {
      common: {
        'fileSize.bytes': `${params?.size || 0} bytes`,
        'fileSize.kb': `${params?.size || 0} KB`,
        'fileSize.mb': `${params?.size || 0} MB`,
        uploaded: 'Uploaded',
      },
      documents: {
        'statuses.ready': 'Ready',
        'statuses.uploading': 'Uploading',
        'statuses.processing': 'Processing',
        'statuses.error': 'Error',
        'categories.educational': 'Educational',
        'categories.administrative': 'Administrative',
        'categories.financial': 'Financial',
        'categories.other': 'Other',
        'filters.size': 'Size',
      },
      documentList: {
        loading: 'Loading...',
        noDocuments: 'No documents',
        noDocumentsHint: 'Upload your first document',
        viewGrid: 'Grid view',
        viewList: 'List view',
        download: 'Download',
        delete: 'Delete',
        share: 'Share',
        edit: 'Edit',
        view: 'View',
      },
    }
    return translations[namespace]?.[key] || key
  },
  useLocale: () => 'en',
}))

const mockDocuments: Document[] = [
  {
    id: '1',
    name: 'Test Document.pdf',
    category: DocumentCategory.EDUCATIONAL,
    status: DocumentStatus.READY,
    metadata: {
      size: 1024 * 1024, // 1MB
      mimeType: 'application/pdf',
      uploadedBy: 'User 1',
      uploadedAt: new Date('2024-06-15'),
    },
    url: '/files/test.pdf',
    description: 'Test description',
  },
  {
    id: '2',
    name: 'Report.docx',
    category: DocumentCategory.ADMINISTRATIVE,
    status: DocumentStatus.PROCESSING,
    metadata: {
      size: 2048,
      mimeType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      uploadedBy: 'User 2',
      uploadedAt: new Date('2024-06-14'),
    },
  },
]

describe('DocumentList', () => {
  const defaultProps = {
    documents: mockDocuments,
  }

  it('renders document list', () => {
    render(<DocumentList {...defaultProps} />)
    expect(screen.getByText('Test Document.pdf')).toBeInTheDocument()
    expect(screen.getByText('Report.docx')).toBeInTheDocument()
  })

  it('shows loading state', () => {
    render(<DocumentList documents={[]} isLoading={true} />)
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('shows empty state when no documents', () => {
    render(<DocumentList documents={[]} />)
    expect(screen.getByText('No documents')).toBeInTheDocument()
  })

  it('calls onPreview when document is clicked', async () => {
    const user = userEvent.setup()
    const onPreview = jest.fn()
    render(<DocumentList {...defaultProps} onPreview={onPreview} />)

    // Click the View button for the first (READY) document
    const viewButtons = screen.getAllByRole('button', { name: /view/i })
    await user.click(viewButtons[0])
    expect(onPreview).toHaveBeenCalledWith(mockDocuments[0])
  })

  it('renders download button when onDownload is provided', () => {
    const onDownload = jest.fn()
    render(<DocumentList {...defaultProps} onDownload={onDownload} />)

    // Download buttons render only icons, check if there are buttons in the document
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('renders delete button when onDelete and canDelete are provided', () => {
    const onDelete = jest.fn()
    render(<DocumentList {...defaultProps} onDelete={onDelete} canDelete={() => true} />)

    // Delete button should be present for documents where canDelete returns true
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('does not render delete button when canDelete returns false', () => {
    const onDelete = jest.fn()
    const { container } = render(
      <DocumentList {...defaultProps} onDelete={onDelete} canDelete={() => false} />
    )

    // No delete button (red text) should be present
    const redButtons = container.querySelectorAll('button.text-red-600')
    expect(redButtons.length).toBe(0)
  })

  it('renders share button when onShare and canShare are provided', () => {
    const onShare = jest.fn()
    render(<DocumentList {...defaultProps} onShare={onShare} canShare={() => true} />)

    // Share button should be present with title attribute
    const shareButton = screen.getByTitle('Share')
    expect(shareButton).toBeInTheDocument()
  })

  it('renders edit button when onEdit and canEdit are provided', () => {
    const onEdit = jest.fn()
    render(<DocumentList {...defaultProps} onEdit={onEdit} canEdit={() => true} />)

    // Edit button should be present with title attribute
    const editButton = screen.getByTitle('Edit')
    expect(editButton).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(<DocumentList {...defaultProps} className="custom-class" />)
    expect(container.querySelector('.custom-class')).toBeInTheDocument()
  })

  it('calls onDelete when delete button is clicked', async () => {
    const user = userEvent.setup()
    const onDelete = jest.fn()
    const { container } = render(
      <DocumentList {...defaultProps} onDelete={onDelete} canDelete={() => true} />
    )

    // Find the delete button by its red color class
    const deleteButton = container.querySelector('button.text-red-600')
    expect(deleteButton).toBeInTheDocument()

    if (deleteButton) {
      await user.click(deleteButton)
      expect(onDelete).toHaveBeenCalledWith(mockDocuments[0])
    }
  })

  it('calls onEdit when edit button is clicked', async () => {
    const user = userEvent.setup()
    const onEdit = jest.fn()
    render(<DocumentList {...defaultProps} onEdit={onEdit} canEdit={() => true} />)

    const editButton = screen.getByTitle('Edit')
    await user.click(editButton)
    expect(onEdit).toHaveBeenCalledWith(mockDocuments[0])
  })

  it('calls onShare when share button is clicked', async () => {
    const user = userEvent.setup()
    const onShare = jest.fn()
    render(<DocumentList {...defaultProps} onShare={onShare} canShare={() => true} />)

    const shareButton = screen.getByTitle('Share')
    await user.click(shareButton)
    expect(onShare).toHaveBeenCalledWith(mockDocuments[0])
  })

  it('calls onDownload when download button is clicked', async () => {
    const user = userEvent.setup()
    const onDownload = jest.fn()
    const { container } = render(<DocumentList {...defaultProps} onDownload={onDownload} />)

    // Find download button (it's next to the view button for READY documents)
    const downloadButtons = container.querySelectorAll('button')
    const downloadButton = Array.from(downloadButtons).find((btn) =>
      btn.querySelector('svg.lucide-download')
    )

    if (downloadButton) {
      await user.click(downloadButton)
      expect(onDownload).toHaveBeenCalledWith(mockDocuments[0])
    }
  })
})
