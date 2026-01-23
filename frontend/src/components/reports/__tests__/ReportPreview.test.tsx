import { render, screen, fireEvent, act, waitFor } from '@testing-library/react'
import { ReportPreview } from '../ReportPreview'
import type { SelectedField, ReportFilter, ReportField } from '@/types/reports'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: { defaultValue?: string; count?: number }) => {
    const translations: Record<string, string> = {
      preview: 'Preview',
      refreshPreview: 'Refresh Preview',
      selectFieldsFirst: 'Select fields first',
      selectFieldsHint: 'Choose fields to include in your report',
      clickToPreview: 'Click to preview data',
      generatePreview: 'Generate Preview',
      loading: 'Loading...',
      previewNote: 'This is a preview with sample data',
    }
    if (key === 'showingResults') return `Showing ${params?.count || 0} results`
    if (key === 'filtersApplied') return `${params?.count || 0} filters applied`
    if (key.startsWith('fields.')) return params?.defaultValue || key
    return translations[key] || key
  },
}))

// Mock GlowingEffect
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => null,
}))

describe('ReportPreview', () => {
  const mockStringField: ReportField = {
    id: 'name',
    name: 'name',
    label: 'Name',
    type: 'string',
    source: 'users',
  }

  const mockNumberField: ReportField = {
    id: 'age',
    name: 'age',
    label: 'Age',
    type: 'number',
    source: 'users',
  }

  const mockDateField: ReportField = {
    id: 'created_at',
    name: 'created_at',
    label: 'Created At',
    type: 'date',
    source: 'users',
  }

  const mockBooleanField: ReportField = {
    id: 'active',
    name: 'active',
    label: 'Active',
    type: 'boolean',
    source: 'users',
  }

  const mockEnumField: ReportField = {
    id: 'status',
    name: 'status',
    label: 'Status',
    type: 'enum',
    source: 'users',
    enumValues: ['active', 'inactive', 'pending'],
  }

  const createSelectedField = (field: ReportField): SelectedField => ({
    field,
    order: 1,
  })

  beforeEach(() => {
    jest.useFakeTimers()
  })

  afterEach(() => {
    jest.useRealTimers()
  })

  it('renders preview title', () => {
    render(<ReportPreview dataSource="users" selectedFields={[]} filters={[]} />)
    expect(screen.getByText('Preview')).toBeInTheDocument()
  })

  it('shows empty state when no fields selected', () => {
    render(<ReportPreview dataSource="users" selectedFields={[]} filters={[]} />)
    expect(screen.getByText('Select fields first')).toBeInTheDocument()
    expect(screen.getByText('Choose fields to include in your report')).toBeInTheDocument()
  })

  it('disables refresh button when no fields selected', () => {
    render(<ReportPreview dataSource="users" selectedFields={[]} filters={[]} />)
    const refreshButton = screen.getByRole('button', { name: /Refresh Preview/i })
    expect(refreshButton).toBeDisabled()
  })

  it('shows click to preview state when fields are selected but no data', () => {
    const selectedFields: SelectedField[] = [createSelectedField(mockStringField)]
    render(<ReportPreview dataSource="users" selectedFields={selectedFields} filters={[]} />)
    expect(screen.getByText('Click to preview data')).toBeInTheDocument()
  })

  it('enables refresh button when fields are selected', () => {
    const selectedFields: SelectedField[] = [createSelectedField(mockStringField)]
    render(<ReportPreview dataSource="users" selectedFields={selectedFields} filters={[]} />)
    const refreshButton = screen.getByRole('button', { name: /Refresh Preview/i })
    expect(refreshButton).not.toBeDisabled()
  })

  it('shows loading state when refreshing', async () => {
    const selectedFields: SelectedField[] = [createSelectedField(mockStringField)]
    render(<ReportPreview dataSource="users" selectedFields={selectedFields} filters={[]} />)

    const refreshButton = screen.getByRole('button', { name: /Refresh Preview/i })
    fireEvent.click(refreshButton)

    // Check loading state (button becomes disabled while loading)
    expect(refreshButton).toBeDisabled()
  })

  it('generates preview data after clicking refresh', async () => {
    const selectedFields: SelectedField[] = [createSelectedField(mockStringField)]
    render(<ReportPreview dataSource="users" selectedFields={selectedFields} filters={[]} />)

    const refreshButton = screen.getByRole('button', { name: /Refresh Preview/i })
    fireEvent.click(refreshButton)

    // Fast-forward timers
    act(() => {
      jest.advanceTimersByTime(500)
    })

    await waitFor(() => {
      expect(screen.getByText(/Showing \d+ results/)).toBeInTheDocument()
    })
  })

  it('shows table with preview data', async () => {
    const selectedFields: SelectedField[] = [createSelectedField(mockStringField)]
    render(<ReportPreview dataSource="users" selectedFields={selectedFields} filters={[]} />)

    const generateButton = screen.getByRole('button', { name: /Generate Preview/i })
    fireEvent.click(generateButton)

    act(() => {
      jest.advanceTimersByTime(500)
    })

    await waitFor(() => {
      // Table should be rendered
      expect(screen.getByRole('table')).toBeInTheDocument()
    })
  })

  it('shows filter count when filters are applied', async () => {
    const selectedFields: SelectedField[] = [createSelectedField(mockStringField)]
    const filters: ReportFilter[] = [
      { id: 'filter-1', field: mockStringField, operator: 'equals', value: 'test' },
      { id: 'filter-2', field: mockNumberField, operator: 'equals', value: 10 },
    ]

    render(<ReportPreview dataSource="users" selectedFields={selectedFields} filters={filters} />)

    const refreshButton = screen.getByRole('button', { name: /Refresh Preview/i })
    fireEvent.click(refreshButton)

    act(() => {
      jest.advanceTimersByTime(500)
    })

    await waitFor(() => {
      expect(screen.getByText('2 filters applied')).toBeInTheDocument()
    })
  })

  it('renders preview note after generating data', async () => {
    const selectedFields: SelectedField[] = [createSelectedField(mockStringField)]
    render(<ReportPreview dataSource="users" selectedFields={selectedFields} filters={[]} />)

    const refreshButton = screen.getByRole('button', { name: /Refresh Preview/i })
    fireEvent.click(refreshButton)

    act(() => {
      jest.advanceTimersByTime(500)
    })

    await waitFor(() => {
      expect(screen.getByText('This is a preview with sample data')).toBeInTheDocument()
    })
  })

  it('handles multiple field types', async () => {
    const selectedFields: SelectedField[] = [
      createSelectedField(mockStringField),
      createSelectedField(mockNumberField),
      createSelectedField(mockDateField),
      createSelectedField(mockBooleanField),
      createSelectedField(mockEnumField),
    ]

    render(<ReportPreview dataSource="users" selectedFields={selectedFields} filters={[]} />)

    const refreshButton = screen.getByRole('button', { name: /Refresh Preview/i })
    fireEvent.click(refreshButton)

    act(() => {
      jest.advanceTimersByTime(500)
    })

    await waitFor(() => {
      expect(screen.getByRole('table')).toBeInTheDocument()
      // Check all column headers are present
      const headers = screen.getAllByRole('columnheader')
      expect(headers.length).toBe(5)
    })
  })
})
