import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DocumentFilters } from '../DocumentFilters'
import type { DocumentFilter, DocumentSortOptions } from '@/types/document'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
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
      'categories.educational_plan': 'Educational plan',
      'categories.curriculum': 'Curriculum',
      'categories.report': 'Report',
      'categories.other': 'Other',
      'statuses.draft': 'Draft',
      'statuses.published': 'Published',
      'statuses.archived': 'Archived',
    }
    return translations[key] || key
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

// Mock next/dynamic
jest.mock('next/dynamic', () => () => {
  const MockCalendar = () => <div data-testid="calendar">Calendar</div>
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
})
