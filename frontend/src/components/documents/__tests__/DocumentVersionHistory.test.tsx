import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DocumentVersionHistory } from '../DocumentVersionHistory'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: (namespace: string) => (key: string, params?: Record<string, unknown>) => {
    const translations: Record<string, Record<string, string>> = {
      'documents.versions': {
        history: 'Version History',
        versionsCount: 'versions',
        loadingHistory: 'Loading...',
        loadError: 'Error loading versions',
        createError: 'Error creating version',
        restoreError: 'Error restoring version',
        deleteError: 'Error deleting version',
        compareError: 'Error comparing versions',
        downloadError: 'Error downloading version',
        retry: 'Retry',
        compare: 'Compare',
        compareCount: `Compare (${params?.count || 0}/2)`,
        createVersion: 'Create version',
        descriptionPlaceholder: 'Version description',
        saveVersion: 'Save',
        version: 'Version',
        current: 'Current',
        empty: 'No versions',
        confirmRestore: 'Restore version?',
        confirmDelete: 'Delete version?',
        comparisonTitle: `Comparing v${params?.from} to v${params?.to}`,
        versionsIdentical: 'Versions are identical',
        changedFields: 'Changed fields',
        before: 'Before',
        after: 'After',
        emptyValue: '(empty)',
        title: 'Title',
        status: 'Status',
        changes: 'Changes',
        file: 'File',
        topic: 'Topic',
        'fieldLabels.title': 'Title',
        'fieldLabels.subject': 'Subject',
        'fieldLabels.content': 'Content',
        'fieldLabels.status': 'Status',
        'fieldLabels.file_name': 'File name',
      },
      common: {
        cancel: 'Cancel',
        close: 'Close',
        'fileSize.mb': `${params?.size || 0} MB`,
      },
    }
    return translations[namespace]?.[key] || key
  },
  useLocale: () => 'en',
}))

// Mock next/dynamic
jest.mock('next/dynamic', () => () => {
  const MockTextDiff = () => <div data-testid="text-diff">Text Diff</div>
  return MockTextDiff
})

const mockVersionData = {
  versions: [
    {
      id: 1,
      document_id: 1,
      version: 2,
      title: 'Version 2',
      created_at: '2024-06-15T10:00:00',
      changed_by_name: 'User 1',
    },
    {
      id: 2,
      document_id: 1,
      version: 1,
      title: 'Version 1',
      created_at: '2024-06-14T10:00:00',
      changed_by_name: 'User 2',
    },
  ],
  total: 2,
  document_id: 1,
  latest_version: 2,
}

// Mock the API - define the mock implementation inside
jest.mock('@/lib/api/documents', () => ({
  documentsApi: {
    getVersions: jest.fn(),
    createVersion: jest.fn(),
    restoreVersion: jest.fn(),
    deleteVersion: jest.fn(),
    compareVersions: jest.fn(),
    getVersionFile: jest.fn(),
  },
}))

describe('DocumentVersionHistory', () => {
  const defaultProps = {
    documentId: 1,
  }

  beforeEach(() => {
    jest.clearAllMocks()
    // Set up the mock implementation for each test
    const { documentsApi } = await import('@/lib/api/documents')
    documentsApi.getVersions.mockResolvedValue(mockVersionData)
  })

  it('renders version history component', async () => {
    render(<DocumentVersionHistory {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Version History')).toBeInTheDocument()
    })
  })

  it('shows loading state initially', () => {
    render(<DocumentVersionHistory {...defaultProps} />)
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('displays versions after loading', async () => {
    render(<DocumentVersionHistory {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText(/Version 2/)).toBeInTheDocument()
    })
  })

  it('applies custom className', async () => {
    const { container } = render(
      <DocumentVersionHistory {...defaultProps} className="custom-class" />
    )

    await waitFor(() => {
      expect(screen.getByText('Version History')).toBeInTheDocument()
    })

    expect(container.querySelector('.custom-class')).toBeInTheDocument()
  })

  it('renders compare button', async () => {
    render(<DocumentVersionHistory {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Compare')).toBeInTheDocument()
    })
  })

  it('can enter compare mode', async () => {
    const user = userEvent.setup()
    render(<DocumentVersionHistory {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Compare')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Compare'))
    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })

  it('shows error state when API fails', async () => {
    const { documentsApi } = await import('@/lib/api/documents')
    documentsApi.getVersions.mockRejectedValueOnce(new Error('API Error'))

    render(<DocumentVersionHistory {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Error loading versions')).toBeInTheDocument()
    })

    expect(screen.getByText('Retry')).toBeInTheDocument()
  })

  it('shows create version button', async () => {
    render(<DocumentVersionHistory {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Create version')).toBeInTheDocument()
    })
  })

  it('displays version count', async () => {
    render(<DocumentVersionHistory {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('(2 versions)')).toBeInTheDocument()
    })
  })
})
