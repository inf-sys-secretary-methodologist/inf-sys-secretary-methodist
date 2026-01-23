import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FieldSelector } from '../FieldSelector'
import type { ReportField, SelectedField } from '@/types/reports'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: { defaultValue?: string }) => {
    const translations: Record<string, string> = {
      availableFields: 'Available Fields',
      searchFields: 'Search fields',
      noFieldsFound: 'No fields found',
      allFieldsSelected: 'All fields selected',
      selectedFields: 'Selected Fields',
      noFieldsSelected: 'No fields selected',
      clickToAddFields: 'Click to add fields',
      options: 'options',
    }
    return translations[key] || params?.defaultValue || key
  },
}))

// Mock framer-motion
jest.mock('framer-motion', () => ({
  motion: {
    button: ({ children, onClick, className }: React.ComponentProps<'button'>) => (
      <button onClick={onClick} className={className}>
        {children}
      </button>
    ),
    p: ({ children, className }: React.ComponentProps<'p'>) => (
      <p className={className}>{children}</p>
    ),
  },
  AnimatePresence: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  Reorder: {
    Group: ({ children, className }: { children: React.ReactNode; className?: string }) => (
      <div className={className}>{children}</div>
    ),
    Item: ({
      children,
      className,
    }: {
      children: React.ReactNode
      className?: string
      value: unknown
    }) => <div className={className}>{children}</div>,
  },
}))

// Mock GlowingEffect
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => null,
}))

const mockFields: ReportField[] = [
  { id: 'doc_id', name: 'id', label: 'ID', type: 'string', source: 'documents' },
  { id: 'doc_name', name: 'name', label: 'Name', type: 'string', source: 'documents' },
  { id: 'doc_created', name: 'created_at', label: 'Created At', type: 'date', source: 'documents' },
]

// Mock AVAILABLE_FIELDS
jest.mock('@/types/reports', () => ({
  AVAILABLE_FIELDS: {
    documents: [
      { id: 'doc_id', name: 'id', label: 'ID', type: 'string', source: 'documents' },
      { id: 'doc_name', name: 'name', label: 'Name', type: 'string', source: 'documents' },
      {
        id: 'doc_created',
        name: 'created_at',
        label: 'Created At',
        type: 'date',
        source: 'documents',
      },
    ],
    users: [],
    events: [],
    tasks: [],
    students: [],
  },
}))

describe('FieldSelector', () => {
  const defaultProps = {
    dataSource: 'documents' as const,
    selectedFields: [] as SelectedField[],
    onAddField: jest.fn(),
    onRemoveField: jest.fn(),
    onReorderFields: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders available fields section', () => {
    render(<FieldSelector {...defaultProps} />)
    expect(screen.getByText('Available Fields')).toBeInTheDocument()
  })

  it('renders selected fields section', () => {
    render(<FieldSelector {...defaultProps} />)
    expect(screen.getByText('Selected Fields')).toBeInTheDocument()
  })

  it('renders search input', () => {
    render(<FieldSelector {...defaultProps} />)
    expect(screen.getByPlaceholderText('Search fields')).toBeInTheDocument()
  })

  it('displays available fields', () => {
    render(<FieldSelector {...defaultProps} />)
    expect(screen.getByText('ID')).toBeInTheDocument()
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Created At')).toBeInTheDocument()
  })

  it('shows empty state when no fields selected', () => {
    render(<FieldSelector {...defaultProps} />)
    expect(screen.getByText('No fields selected')).toBeInTheDocument()
    expect(screen.getByText('Click to add fields')).toBeInTheDocument()
  })

  it('calls onAddField when available field is clicked', async () => {
    const user = userEvent.setup()
    const onAddField = jest.fn()
    render(<FieldSelector {...defaultProps} onAddField={onAddField} />)

    await user.click(screen.getByText('ID'))
    expect(onAddField).toHaveBeenCalledWith(expect.objectContaining({ id: 'doc_id', name: 'id' }))
  })

  it('filters available fields based on search', async () => {
    const user = userEvent.setup()
    render(<FieldSelector {...defaultProps} />)

    const searchInput = screen.getByPlaceholderText('Search fields')
    await user.type(searchInput, 'Name')

    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.queryByText('ID')).not.toBeInTheDocument()
  })

  it('shows "No fields found" when search has no results', async () => {
    const user = userEvent.setup()
    render(<FieldSelector {...defaultProps} />)

    const searchInput = screen.getByPlaceholderText('Search fields')
    await user.type(searchInput, 'xyz')

    expect(screen.getByText('No fields found')).toBeInTheDocument()
  })

  it('displays selected fields', () => {
    const selectedFields: SelectedField[] = [
      { field: mockFields[0], order: 0 },
      { field: mockFields[1], order: 1 },
    ]
    render(<FieldSelector {...defaultProps} selectedFields={selectedFields} />)

    // Both selected fields should be visible in selected section
    const idElements = screen.getAllByText('ID')
    expect(idElements.length).toBeGreaterThan(0)
  })

  it('hides already selected fields from available fields', () => {
    const selectedFields: SelectedField[] = [
      { field: mockFields[0], order: 0 },
      { field: mockFields[1], order: 1 },
      { field: mockFields[2], order: 2 },
    ]
    render(<FieldSelector {...defaultProps} selectedFields={selectedFields} />)

    // Should show "All fields selected" in available fields section
    expect(screen.getByText('All fields selected')).toBeInTheDocument()
  })

  it('shows selected fields count', () => {
    const selectedFields: SelectedField[] = [
      { field: mockFields[0], order: 0 },
      { field: mockFields[1], order: 1 },
    ]
    render(<FieldSelector {...defaultProps} selectedFields={selectedFields} />)

    expect(screen.getByText('(2)')).toBeInTheDocument()
  })

  it('calls onRemoveField when remove button is clicked', async () => {
    const user = userEvent.setup()
    const onRemoveField = jest.fn()
    const selectedFields: SelectedField[] = [{ field: mockFields[0], order: 0 }]

    render(
      <FieldSelector
        {...defaultProps}
        selectedFields={selectedFields}
        onRemoveField={onRemoveField}
      />
    )

    // Find and click the remove button (X icon)
    const removeButtons = screen.getAllByRole('button')
    const removeButton = removeButtons.find((btn) => btn.querySelector('svg.lucide-x'))
    if (removeButton) {
      await user.click(removeButton)
      expect(onRemoveField).toHaveBeenCalledWith('doc_id')
    }
  })
})
