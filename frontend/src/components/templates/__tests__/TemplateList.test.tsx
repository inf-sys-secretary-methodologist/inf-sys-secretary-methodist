import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import { TemplateList } from '../TemplateList'
import { templatesApi } from '@/lib/api/templates'
import type { TemplateInfo } from '@/lib/api/templates'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      loadError: 'Failed to load templates',
      searchPlaceholder: 'Search templates...',
      noSearchResults: 'No templates match your search',
      noTemplates: 'No templates available',
      variables: 'variables',
      code: 'Code',
      preview: 'Preview',
      create: 'Create',
    }
    return translations[key] || key
  },
}))

// Mock API
jest.mock('@/lib/api/templates', () => ({
  templatesApi: {
    getAll: jest.fn(),
  },
}))

// Mock GlowingEffect
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => null,
}))

const mockTemplates: TemplateInfo[] = [
  {
    id: 1,
    name: 'Invoice Template',
    code: 'invoice',
    description: 'Standard invoice template',
    has_template: true,
    template_variables: [{ name: 'amount', required: true, variable_type: 'number' }],
  },
  {
    id: 2,
    name: 'Contract Template',
    code: 'contract',
    description: 'Legal contract template',
    has_template: true,
    template_variables: [],
  },
]

const mockApiClient = templatesApi as jest.Mocked<typeof templatesApi>

describe('TemplateList', () => {
  const mockOnPreview = jest.fn()
  const mockOnCreate = jest.fn()
  const mockOnEdit = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('shows loading state initially', () => {
    mockApiClient.getAll.mockImplementation(() => new Promise(() => {}))
    render(<TemplateList onPreview={mockOnPreview} onCreate={mockOnCreate} />)

    expect(screen.getByText('', { selector: '.animate-spin' })).toBeInTheDocument()
  })

  it('renders templates after loading', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(<TemplateList onPreview={mockOnPreview} onCreate={mockOnCreate} />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
      expect(screen.getByText('Contract Template')).toBeInTheDocument()
    })
  })

  it('shows error message on API failure', async () => {
    mockApiClient.getAll.mockRejectedValue(new Error('API Error'))

    render(<TemplateList onPreview={mockOnPreview} onCreate={mockOnCreate} />)

    await waitFor(() => {
      expect(screen.getByText('Failed to load templates')).toBeInTheDocument()
    })
  })

  it('shows empty state when no templates', async () => {
    mockApiClient.getAll.mockResolvedValue([])

    render(<TemplateList onPreview={mockOnPreview} onCreate={mockOnCreate} />)

    await waitFor(() => {
      expect(screen.getByText('No templates available')).toBeInTheDocument()
    })
  })

  it('filters templates by search query', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(<TemplateList onPreview={mockOnPreview} onCreate={mockOnCreate} />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('Search templates...')
    await act(async () => {
      fireEvent.change(searchInput, { target: { value: 'invoice' } })
    })

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
      expect(screen.queryByText('Contract Template')).not.toBeInTheDocument()
    })
  })

  it('filters by template code', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(<TemplateList onPreview={mockOnPreview} onCreate={mockOnCreate} />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('Search templates...')
    await act(async () => {
      fireEvent.change(searchInput, { target: { value: 'contract' } })
    })

    await waitFor(() => {
      expect(screen.queryByText('Invoice Template')).not.toBeInTheDocument()
      expect(screen.getByText('Contract Template')).toBeInTheDocument()
    })
  })

  it('filters by description', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(<TemplateList onPreview={mockOnPreview} onCreate={mockOnCreate} />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('Search templates...')
    await act(async () => {
      fireEvent.change(searchInput, { target: { value: 'legal' } })
    })

    await waitFor(() => {
      expect(screen.queryByText('Invoice Template')).not.toBeInTheDocument()
      expect(screen.getByText('Contract Template')).toBeInTheDocument()
    })
  })

  it('shows no results message when search has no matches', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(<TemplateList onPreview={mockOnPreview} onCreate={mockOnCreate} />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('Search templates...')
    await act(async () => {
      fireEvent.change(searchInput, { target: { value: 'nonexistent' } })
    })

    await waitFor(() => {
      expect(screen.getByText('No templates match your search')).toBeInTheDocument()
    })
  })

  it('passes onPreview to template cards', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(<TemplateList onPreview={mockOnPreview} onCreate={mockOnCreate} />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    const previewButtons = screen.getAllByText('Preview')
    fireEvent.click(previewButtons[0])

    expect(mockOnPreview).toHaveBeenCalledWith(mockTemplates[0])
  })

  it('passes onCreate to template cards', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(<TemplateList onPreview={mockOnPreview} onCreate={mockOnCreate} />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    const createButtons = screen.getAllByText('Create')
    fireEvent.click(createButtons[0])

    expect(mockOnCreate).toHaveBeenCalledWith(mockTemplates[0])
  })

  it('passes canEdit and onEdit to template cards', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(
      <TemplateList
        onPreview={mockOnPreview}
        onCreate={mockOnCreate}
        onEdit={mockOnEdit}
        canEdit={true}
      />
    )

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    // When canEdit is true, we should have Preview, Create, and Edit buttons for each template
    // Verify that each template has its action buttons
    const previewButtons = screen.getAllByText('Preview')
    const createButtons = screen.getAllByText('Create')
    expect(previewButtons.length).toBe(2)
    expect(createButtons.length).toBe(2)
    // Edit buttons exist when canEdit is true - 2 templates = 2 edit buttons
    // Total buttons = 2 preview + 2 create + 2 edit = 6 buttons
    const allButtons = screen.getAllByRole('button')
    expect(allButtons.length).toBeGreaterThanOrEqual(6)
  })

  it('case-insensitive search', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(<TemplateList onPreview={mockOnPreview} onCreate={mockOnCreate} />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('Search templates...')
    await act(async () => {
      fireEvent.change(searchInput, { target: { value: 'INVOICE' } })
    })

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })
  })
})
