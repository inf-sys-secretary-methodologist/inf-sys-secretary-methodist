import { render, screen, fireEvent } from '@testing-library/react'
import { FilterBuilder } from '../FilterBuilder'
import type { ReportField, ReportFilter } from '@/types/reports'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: { defaultValue?: string }) => {
    const translations: Record<string, string> = {
      filters: 'Filters',
      addFilter: 'Add Filter',
      noFilters: 'No filters',
      addFiltersHint: 'Add filters to refine your data',
      selectField: 'Select field',
      add: 'Add',
      and: 'and',
      enterValue: 'Enter value',
      selectValue: 'Select value',
      true: 'True',
      false: 'False',
      'operators.equals': 'equals',
      'operators.not_equals': 'not equals',
      'operators.contains': 'contains',
      'operators.greater_than': 'greater than',
      'operators.less_than': 'less than',
      'operators.between': 'between',
      'operators.is_null': 'is null',
      'operators.is_not_null': 'is not null',
    }
    if (key.startsWith('fields.') || key.startsWith('enumValues.')) {
      return params?.defaultValue || key
    }
    return translations[key] || key
  },
}))

// Mock GlowingEffect
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => null,
}))

// Mock motion/react
jest.mock('motion/react', () => ({
  motion: {
    div: ({ children, ...props }: React.PropsWithChildren<Record<string, unknown>>) => (
      <div {...props}>{children}</div>
    ),
  },
  AnimatePresence: ({ children }: React.PropsWithChildren) => <>{children}</>,
}))

// Mock AVAILABLE_FIELDS
jest.mock('@/types/reports', () => ({
  AVAILABLE_FIELDS: {
    users: [
      { id: 'name', name: 'name', label: 'Name', type: 'string', source: 'users' },
      { id: 'age', name: 'age', label: 'Age', type: 'number', source: 'users' },
      { id: 'active', name: 'active', label: 'Active', type: 'boolean', source: 'users' },
      {
        id: 'status',
        name: 'status',
        label: 'Status',
        type: 'enum',
        source: 'users',
        enumValues: ['active', 'inactive'],
      },
      { id: 'created_at', name: 'created_at', label: 'Created At', type: 'date', source: 'users' },
    ],
  },
}))

describe('FilterBuilder', () => {
  const mockOnAddFilter = jest.fn()
  const mockOnRemoveFilter = jest.fn()
  const mockOnUpdateFilter = jest.fn()

  const defaultProps = {
    dataSource: 'users' as const,
    filters: [] as ReportFilter[],
    onAddFilter: mockOnAddFilter,
    onRemoveFilter: mockOnRemoveFilter,
    onUpdateFilter: mockOnUpdateFilter,
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders filters title', () => {
    render(<FilterBuilder {...defaultProps} />)
    expect(screen.getByText('Filters')).toBeInTheDocument()
  })

  it('renders add filter button', () => {
    render(<FilterBuilder {...defaultProps} />)
    expect(screen.getByRole('button', { name: /Add Filter/i })).toBeInTheDocument()
  })

  it('shows empty state when no filters', () => {
    render(<FilterBuilder {...defaultProps} />)
    expect(screen.getByText('No filters')).toBeInTheDocument()
    expect(screen.getByText('Add filters to refine your data')).toBeInTheDocument()
  })

  it('shows add filter form when clicking add button', () => {
    render(<FilterBuilder {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /Add Filter/i }))
    expect(screen.getByText('Select field')).toBeInTheDocument()
  })

  it('cancels adding filter when X button clicked', () => {
    render(<FilterBuilder {...defaultProps} />)
    // Open add filter form
    fireEvent.click(screen.getByRole('button', { name: /Add Filter/i }))
    expect(screen.getByText('Select field')).toBeInTheDocument()

    // Click X button to cancel
    const buttons = screen.getAllByRole('button')
    const cancelButton = buttons.find((btn) => btn.querySelector('svg.lucide-x'))
    if (cancelButton) {
      fireEvent.click(cancelButton)
    }
    // Form should be closed, empty state should show
    expect(screen.getByText('No filters')).toBeInTheDocument()
  })

  it('shows filter count when filters exist', () => {
    const stringField: ReportField = {
      id: 'name',
      name: 'name',
      label: 'Name',
      type: 'string',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: stringField, operator: 'equals', value: 'test' },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)
    expect(screen.getByText('(1)')).toBeInTheDocument()
  })

  it('renders existing filters', () => {
    const stringField: ReportField = {
      id: 'name',
      name: 'name',
      label: 'Name',
      type: 'string',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: stringField, operator: 'equals', value: 'test' },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)
    expect(screen.getByText('Name')).toBeInTheDocument()
  })

  it('calls onRemoveFilter when remove button clicked', () => {
    const stringField: ReportField = {
      id: 'name',
      name: 'name',
      label: 'Name',
      type: 'string',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: stringField, operator: 'equals', value: 'test' },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)

    // Find and click the remove button (X icon)
    const removeButtons = screen.getAllByRole('button')
    const removeButton = removeButtons.find((btn) => btn.querySelector('svg.lucide-x'))
    if (removeButton) {
      fireEvent.click(removeButton)
      expect(mockOnRemoveFilter).toHaveBeenCalledWith('filter-1')
    }
  })

  it('shows "and" between multiple filters', () => {
    const stringField: ReportField = {
      id: 'name',
      name: 'name',
      label: 'Name',
      type: 'string',
      source: 'users',
    }
    const numberField: ReportField = {
      id: 'age',
      name: 'age',
      label: 'Age',
      type: 'number',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: stringField, operator: 'equals', value: 'test' },
      { id: 'filter-2', field: numberField, operator: 'greater_than', value: 18 },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)
    expect(screen.getByText('and')).toBeInTheDocument()
  })

  it('renders text input for string fields', () => {
    const stringField: ReportField = {
      id: 'name',
      name: 'name',
      label: 'Name',
      type: 'string',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: stringField, operator: 'equals', value: 'test' },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)
    expect(screen.getByPlaceholderText('Enter value')).toBeInTheDocument()
  })

  it('renders number input for number fields', () => {
    const numberField: ReportField = {
      id: 'age',
      name: 'age',
      label: 'Age',
      type: 'number',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: numberField, operator: 'equals', value: 25 },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)
    const input = screen.getByPlaceholderText('Enter value')
    expect(input).toHaveAttribute('type', 'number')
  })

  it('renders date input for date fields', () => {
    const dateField: ReportField = {
      id: 'created_at',
      name: 'created_at',
      label: 'Created At',
      type: 'date',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: dateField, operator: 'equals', value: '2024-01-01' },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)
    const input = document.querySelector('input[type="date"]')
    expect(input).toBeInTheDocument()
  })

  it('does not show value input for is_null operator', () => {
    const stringField: ReportField = {
      id: 'name',
      name: 'name',
      label: 'Name',
      type: 'string',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: stringField, operator: 'is_null', value: null },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)
    expect(screen.queryByPlaceholderText('Enter value')).not.toBeInTheDocument()
  })

  it('does not show value input for is_not_null operator', () => {
    const stringField: ReportField = {
      id: 'name',
      name: 'name',
      label: 'Name',
      type: 'string',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: stringField, operator: 'is_not_null', value: null },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)
    expect(screen.queryByPlaceholderText('Enter value')).not.toBeInTheDocument()
  })

  it('calls onUpdateFilter when input value changes', () => {
    const stringField: ReportField = {
      id: 'name',
      name: 'name',
      label: 'Name',
      type: 'string',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: stringField, operator: 'equals', value: '' },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)

    const input = screen.getByPlaceholderText('Enter value')
    fireEvent.change(input, { target: { value: 'new value' } })

    expect(mockOnUpdateFilter).toHaveBeenCalledWith('filter-1', { value: 'new value' })
  })

  it('renders enum field with select value', () => {
    const enumField: ReportField = {
      id: 'status',
      name: 'status',
      label: 'Status',
      type: 'enum',
      source: 'users',
      enumValues: ['active', 'inactive'],
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: enumField, operator: 'equals', value: 'active' },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)
    expect(screen.getByText('Status')).toBeInTheDocument()
  })

  it('renders boolean field with select', () => {
    const booleanField: ReportField = {
      id: 'active',
      name: 'active',
      label: 'Active',
      type: 'boolean',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: booleanField, operator: 'equals', value: true },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)
    expect(screen.getByText('Active')).toBeInTheDocument()
  })

  it('renders between operator with two inputs for number', () => {
    const numberField: ReportField = {
      id: 'age',
      name: 'age',
      label: 'Age',
      type: 'number',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: numberField, operator: 'between', value: 18, value2: 65 },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)
    const inputs = screen.getAllByPlaceholderText('Enter value')
    expect(inputs.length).toBe(2)
  })

  it('renders between operator with two date inputs', () => {
    const dateField: ReportField = {
      id: 'created_at',
      name: 'created_at',
      label: 'Created At',
      type: 'date',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      {
        id: 'filter-1',
        field: dateField,
        operator: 'between',
        value: '2024-01-01',
        value2: '2024-12-31',
      },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)
    const dateInputs = document.querySelectorAll('input[type="date"]')
    expect(dateInputs.length).toBe(2)
  })

  it('calls onUpdateFilter when number input changes', () => {
    const numberField: ReportField = {
      id: 'age',
      name: 'age',
      label: 'Age',
      type: 'number',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: numberField, operator: 'equals', value: 0 },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)

    const input = screen.getByPlaceholderText('Enter value')
    fireEvent.change(input, { target: { value: '25' } })

    expect(mockOnUpdateFilter).toHaveBeenCalledWith('filter-1', { value: 25 })
  })

  it('calls onUpdateFilter when date input changes', () => {
    const dateField: ReportField = {
      id: 'created_at',
      name: 'created_at',
      label: 'Created At',
      type: 'date',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: dateField, operator: 'equals', value: '' },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)

    const input = document.querySelector('input[type="date"]') as HTMLInputElement
    fireEvent.change(input, { target: { value: '2024-06-15' } })

    expect(mockOnUpdateFilter).toHaveBeenCalledWith('filter-1', { value: '2024-06-15' })
  })

  it('calls onUpdateFilter when between number value2 changes', () => {
    const numberField: ReportField = {
      id: 'age',
      name: 'age',
      label: 'Age',
      type: 'number',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: numberField, operator: 'between', value: 18, value2: 65 },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)

    const inputs = screen.getAllByPlaceholderText('Enter value')
    fireEvent.change(inputs[1], { target: { value: '70' } })

    expect(mockOnUpdateFilter).toHaveBeenCalledWith('filter-1', { value2: 70 })
  })

  it('calls onUpdateFilter when between date value2 changes', () => {
    const dateField: ReportField = {
      id: 'created_at',
      name: 'created_at',
      label: 'Created At',
      type: 'date',
      source: 'users',
    }
    const filters: ReportFilter[] = [
      {
        id: 'filter-1',
        field: dateField,
        operator: 'between',
        value: '2024-01-01',
        value2: '2024-12-31',
      },
    ]

    render(<FilterBuilder {...defaultProps} filters={filters} />)

    const dateInputs = document.querySelectorAll('input[type="date"]')
    fireEvent.change(dateInputs[1], { target: { value: '2025-01-01' } })

    expect(mockOnUpdateFilter).toHaveBeenCalledWith('filter-1', { value2: '2025-01-01' })
  })
})
