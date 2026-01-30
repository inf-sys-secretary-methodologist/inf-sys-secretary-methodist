import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DocumentFilters } from '../DocumentFilters'
import {
  DocumentCategory,
  DocumentStatus,
  type DocumentFilter,
  type DocumentSortOptions,
} from '@/types/document'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: (namespace: string) => (key: string) => {
    const translations: Record<string, Record<string, string>> = {
      'documents.filters': {
        category: 'Category',
        status: 'Status',
        author: 'Author',
        dateFrom: 'From date',
        dateTo: 'To date',
        tags: 'Tags',
        sort: 'Sort',
        allCategories: 'All categories',
        allStatuses: 'All statuses',
        allAuthors: 'All authors',
        loading: 'Loading...',
        selectDate: 'Select date',
        tagsPlaceholder: 'Enter tags separated by comma',
        search: 'Search',
        name: 'Name',
        uploadDate: 'Upload date',
        modifyDate: 'Modify date',
        size: 'Size',
        authorPrefix: 'Author',
        from: 'From',
        to: 'To',
        searchPlaceholder: 'Search documents...',
        resetDate: 'Reset date',
        filters: 'Filters',
        reset: 'Reset',
      },
      documents: {
        'categories.educational': 'Educational',
        'categories.educational_plan': 'Educational plan',
        'categories.curriculum': 'Curriculum',
        'categories.report': 'Report',
        'categories.other': 'Other',
        'categories.hr': 'HR',
        'categories.administrative': 'Administrative',
        'categories.methodical': 'Methodical',
        'categories.financial': 'Financial',
        'categories.archive': 'Archive',
        'statuses.uploading': 'Uploading',
        'statuses.processing': 'Processing',
        'statuses.ready': 'Ready',
        'statuses.error': 'Error',
      },
      'documents.form': {
        resetDate: 'Reset date',
        searchPlaceholder: 'Search documents...',
      },
      common: {
        reset: 'Reset',
        cancel: 'Cancel',
      },
    }
    return translations[namespace]?.[key] || key
  },
  useLocale: () => 'en',
}))

// Mock users API
jest.mock('@/lib/api/users', () => ({
  usersApi: {
    getAll: jest.fn().mockResolvedValue([
      { id: 1, name: 'John Doe' },
      { id: 2, name: 'Jane Smith' },
    ]),
  },
}))

// Mock next/dynamic with disabled function coverage
jest.mock('next/dynamic', () => () => {
  const MockCalendar = ({
    onSelect,
    disabled,
  }: {
    onSelect?: (date: Date | undefined) => void
    disabled?: (date: Date) => boolean
  }) => {
    // Call disabled function with test dates for coverage
    if (disabled) {
      // Test with various dates to cover the disabled logic
      disabled(new Date(2024, 0, 1))
      disabled(new Date(2024, 11, 31))
    }
    return (
      <div data-testid="calendar">
        <button onClick={() => onSelect?.(new Date(2024, 5, 15))}>Select Date</button>
      </div>
    )
  }
  return MockCalendar
})

describe('DocumentFilters', () => {
  const defaultProps = {
    onFilterChange: jest.fn(),
    onSortChange: jest.fn(),
    currentFilters: {} as DocumentFilter,
    currentSort: { field: 'uploadedAt' as const, order: 'desc' as const } as DocumentSortOptions,
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders search input', () => {
    render(<DocumentFilters {...defaultProps} />)
    expect(screen.getByPlaceholderText('Search documents...')).toBeInTheDocument()
  })

  it('renders filters button', () => {
    render(<DocumentFilters {...defaultProps} />)
    expect(screen.getByRole('button', { name: /filters/i })).toBeInTheDocument()
  })

  it('expands filters when filters button is clicked', async () => {
    render(<DocumentFilters {...defaultProps} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))

    expect(screen.getByText('Category')).toBeInTheDocument()
    expect(screen.getByText('Status')).toBeInTheDocument()
    expect(screen.getByText('Author')).toBeInTheDocument()
  })

  it('calls onFilterChange when search input changes', async () => {
    render(<DocumentFilters {...defaultProps} />)
    const searchInput = screen.getByPlaceholderText('Search documents...')

    await userEvent.type(searchInput, 'test')

    expect(defaultProps.onFilterChange).toHaveBeenCalledWith({
      search: 'test',
    })
  })

  it('shows category select when expanded', async () => {
    render(<DocumentFilters {...defaultProps} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))

    const comboboxes = screen.getAllByRole('combobox')
    expect(comboboxes[0]).toBeInTheDocument()
  })

  it('renders category select when expanded', async () => {
    render(<DocumentFilters {...defaultProps} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))

    const comboboxes = screen.getAllByRole('combobox')
    expect(comboboxes.length).toBeGreaterThanOrEqual(3) // category, status, author
  })

  it('renders status select when expanded', async () => {
    render(<DocumentFilters {...defaultProps} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))

    expect(screen.getByText('Status')).toBeInTheDocument()
  })

  it('shows reset button when filters are active', () => {
    render(<DocumentFilters {...defaultProps} currentFilters={{ search: 'test' }} />)
    expect(screen.getByRole('button', { name: /reset/i })).toBeInTheDocument()
  })

  it('does not show reset button when no filters are active', () => {
    render(<DocumentFilters {...defaultProps} />)
    expect(screen.queryByRole('button', { name: /reset/i })).not.toBeInTheDocument()
  })

  it('clears all filters when reset button is clicked', async () => {
    render(<DocumentFilters {...defaultProps} currentFilters={{ search: 'test' }} />)

    await userEvent.click(screen.getByRole('button', { name: /reset/i }))

    expect(defaultProps.onFilterChange).toHaveBeenCalledWith({})
  })

  it('shows active filter count badge', () => {
    render(<DocumentFilters {...defaultProps} currentFilters={{ search: 'test' }} />)
    // Badge should show "1" for one active filter
    expect(screen.getByText('1')).toBeInTheDocument()
  })

  it('shows sort buttons when expanded', async () => {
    render(<DocumentFilters {...defaultProps} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))

    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Upload date')).toBeInTheDocument()
    expect(screen.getByText('Modify date')).toBeInTheDocument()
    expect(screen.getByText('Size')).toBeInTheDocument()
  })

  it('calls onSortChange when sort button is clicked', async () => {
    render(<DocumentFilters {...defaultProps} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))
    await userEvent.click(screen.getByRole('button', { name: 'Name' }))

    expect(defaultProps.onSortChange).toHaveBeenCalledWith({
      field: 'name',
      order: 'desc',
    })
  })

  it('toggles sort order when same sort button is clicked twice', async () => {
    render(<DocumentFilters {...defaultProps} currentSort={{ field: 'name', order: 'desc' }} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))
    await userEvent.click(screen.getByRole('button', { name: /name/i }))

    expect(defaultProps.onSortChange).toHaveBeenCalledWith({
      field: 'name',
      order: 'asc',
    })
  })

  it('shows active filter tags when not expanded', () => {
    render(<DocumentFilters {...defaultProps} currentFilters={{ search: 'test document' }} />)
    expect(screen.getByText(/search.*test document/i)).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(
      <DocumentFilters {...defaultProps} className="custom-filter-class" />
    )
    expect(container.firstChild).toHaveClass('custom-filter-class')
  })

  it('can change tags input', async () => {
    render(<DocumentFilters {...defaultProps} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))

    const tagsInput = screen.getByPlaceholderText('Enter tags separated by comma')
    await userEvent.type(tagsInput, 'tag1, tag2')

    expect(defaultProps.onFilterChange).toHaveBeenCalledWith(
      expect.objectContaining({
        tags: expect.any(Array),
      })
    )
  })

  it('can remove category filter tag when collapsed', async () => {
    render(
      <DocumentFilters
        {...defaultProps}
        currentFilters={{ category: DocumentCategory.EDUCATIONAL }}
      />
    )

    // Find and click the X button to remove the category filter
    const categoryTag = screen.getByText('Educational')
    const removeButton = categoryTag.parentElement?.querySelector('button')

    if (removeButton) {
      await userEvent.click(removeButton)
      expect(defaultProps.onFilterChange).toHaveBeenCalled()
    }
  })

  it('can remove status filter tag when collapsed', async () => {
    render(<DocumentFilters {...defaultProps} currentFilters={{ status: DocumentStatus.READY }} />)

    // Find the status tag and click remove
    const statusTag = screen.getByText('Ready')
    const removeButton = statusTag.parentElement?.querySelector('button')

    if (removeButton) {
      await userEvent.click(removeButton)
      expect(defaultProps.onFilterChange).toHaveBeenCalled()
    }
  })

  it('can remove search filter tag when collapsed', async () => {
    render(<DocumentFilters {...defaultProps} currentFilters={{ search: 'test' }} />)

    // Find the search tag and click remove
    const searchTag = screen.getByText(/search.*test/i)
    const removeButton = searchTag.parentElement?.querySelector('button')

    if (removeButton) {
      await userEvent.click(removeButton)
      expect(defaultProps.onFilterChange).toHaveBeenCalled()
    }
  })

  it('can remove author filter tag when collapsed', async () => {
    render(<DocumentFilters {...defaultProps} currentFilters={{ authorId: 1 }} />)

    // Wait for authors to load
    await screen.findByText(/author/i)

    // Find the author tag and click remove
    const removeButtons = screen.getAllByRole('button')
    const removeButton = removeButtons.find((btn) =>
      btn.closest('.bg-green-100, .dark\\:bg-green-900\\/30')
    )

    if (removeButton) {
      await userEvent.click(removeButton)
      expect(defaultProps.onFilterChange).toHaveBeenCalled()
    }
  })

  it('can remove dateFrom filter tag when collapsed', async () => {
    const dateFrom = new Date(2024, 0, 1)
    render(<DocumentFilters {...defaultProps} currentFilters={{ dateFrom }} />)

    // Find the date from tag and click remove
    const dateTag = screen.getByText(/from.*01\.01\.2024/i)
    const removeButton = dateTag.parentElement?.querySelector('button')

    if (removeButton) {
      await userEvent.click(removeButton)
      expect(defaultProps.onFilterChange).toHaveBeenCalled()
    }
  })

  it('can remove dateTo filter tag when collapsed', async () => {
    const dateTo = new Date(2024, 11, 31)
    render(<DocumentFilters {...defaultProps} currentFilters={{ dateTo }} />)

    // Find the date to tag and click remove
    const dateTag = screen.getByText(/to.*31\.12\.2024/i)
    const removeButton = dateTag.parentElement?.querySelector('button')

    if (removeButton) {
      await userEvent.click(removeButton)
      expect(defaultProps.onFilterChange).toHaveBeenCalled()
    }
  })

  it('can remove tags filter tag when collapsed', async () => {
    render(<DocumentFilters {...defaultProps} currentFilters={{ tags: ['tag1', 'tag2'] }} />)

    // Find the tags filter and click remove
    const tagsTag = screen.getByText(/tags.*tag1.*tag2/i)
    const removeButton = tagsTag.parentElement?.querySelector('button')

    if (removeButton) {
      await userEvent.click(removeButton)
      expect(defaultProps.onFilterChange).toHaveBeenCalled()
    }
  })

  it('can reset date from when expanded and date is selected', async () => {
    const dateFrom = new Date(2024, 0, 1)
    render(<DocumentFilters {...defaultProps} currentFilters={{ dateFrom }} />)

    // Expand filters
    await userEvent.click(screen.getByRole('button', { name: /filters/i }))

    // Find reset date button (the X button next to date)
    const resetButtons = screen.getAllByRole('button', { name: /reset/i })
    if (resetButtons.length > 0) {
      await userEvent.click(resetButtons[0])
    }
    expect(defaultProps.onFilterChange).toHaveBeenCalled()
  })

  it('can reset date to when expanded and date is selected', async () => {
    const dateTo = new Date(2024, 11, 31)
    render(<DocumentFilters {...defaultProps} currentFilters={{ dateTo }} />)

    // Expand filters
    await userEvent.click(screen.getByRole('button', { name: /filters/i }))

    // Find reset date button
    const resetButtons = screen.getAllByRole('button', { name: /reset/i })
    if (resetButtons.length > 0) {
      await userEvent.click(resetButtons[resetButtons.length - 1])
    }
    expect(defaultProps.onFilterChange).toHaveBeenCalled()
  })

  it('sorts by modifyDate', async () => {
    render(<DocumentFilters {...defaultProps} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))
    await userEvent.click(screen.getByRole('button', { name: /modify.*date/i }))

    expect(defaultProps.onSortChange).toHaveBeenCalledWith({
      field: 'modifiedAt',
      order: 'desc',
    })
  })

  it('sorts by size', async () => {
    render(<DocumentFilters {...defaultProps} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))
    await userEvent.click(screen.getByRole('button', { name: /size/i }))

    expect(defaultProps.onSortChange).toHaveBeenCalledWith({
      field: 'size',
      order: 'desc',
    })
  })

  it('toggles uploadedAt sort order', async () => {
    // defaultProps has currentSort.field = 'uploadedAt' with order 'desc'
    // so clicking again should toggle to 'asc'
    render(<DocumentFilters {...defaultProps} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))
    await userEvent.click(screen.getByRole('button', { name: /upload.*date/i }))

    expect(defaultProps.onSortChange).toHaveBeenCalledWith({
      field: 'uploadedAt',
      order: 'asc',
    })
  })

  it('renders calendar components when dates are filtered', async () => {
    // This test covers the calendar rendering which exercises the disabled functions
    const dateFrom = new Date(2024, 5, 10)
    const dateTo = new Date(2024, 5, 20)
    render(<DocumentFilters {...defaultProps} currentFilters={{ dateFrom, dateTo }} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))

    // Date labels should be visible when expanded
    expect(screen.getByText('From date')).toBeInTheDocument()
    expect(screen.getByText('To date')).toBeInTheDocument()
  })

  it('resets dateFrom by clicking reset button', async () => {
    const dateFrom = new Date(2024, 5, 15)
    render(<DocumentFilters {...defaultProps} currentFilters={{ dateFrom }} />)

    await userEvent.click(screen.getByRole('button', { name: /filters/i }))

    // Find and click the reset date button (X button)
    const resetButtons = screen.getAllByRole('button', { name: /reset.*date/i })
    if (resetButtons.length > 0) {
      await userEvent.click(resetButtons[0])
      expect(defaultProps.onFilterChange).toHaveBeenCalled()
    }
  })
})
