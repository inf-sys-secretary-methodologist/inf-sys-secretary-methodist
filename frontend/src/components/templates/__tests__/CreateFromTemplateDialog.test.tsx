import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { CreateFromTemplateDialog } from '../CreateFromTemplateDialog'
import { templatesApi } from '@/lib/api/templates'
import type { TemplateInfo } from '@/lib/api/templates'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: (namespace: string) => (key: string) => {
    const translations: Record<string, Record<string, string>> = {
      templates: {
        createFromTemplate: 'Create from Template',
        documentTitle: 'Document Title',
        enterTitle: 'Enter document title',
        fillVariables: 'Fill Variables',
        selectOption: 'Select an option',
        previewTitle: 'Preview',
        preview: 'Preview',
        previewError: 'Preview failed',
        createDocument: 'Create Document',
        createError: 'Failed to create document',
        createSuccess: 'Document created successfully',
        redirecting: 'Redirecting...',
      },
      common: {
        cancel: 'Cancel',
      },
    }
    return translations[namespace]?.[key] || key
  },
}))

// Mock next/navigation
const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}))

// Mock API
jest.mock('@/lib/api/templates', () => ({
  templatesApi: {
    preview: jest.fn(),
    createDocument: jest.fn(),
  },
}))

const mockTemplate: TemplateInfo = {
  id: 1,
  name: 'Test Template',
  code: 'test-template',
  description: 'A test template',
  has_template: true,
  template_variables: [
    { name: 'studentName', required: true, variable_type: 'text', default_value: 'John' },
    { name: 'date', required: false, variable_type: 'date' },
    { name: 'amount', required: true, variable_type: 'number' },
    {
      name: 'type',
      required: true,
      variable_type: 'select',
      options: ['Option A', 'Option B', 'Option C'],
    },
  ],
}

const mockApiClient = templatesApi as jest.Mocked<typeof templatesApi>

describe('CreateFromTemplateDialog', () => {
  const mockOnOpenChange = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
    jest.useFakeTimers()
  })

  afterEach(() => {
    jest.useRealTimers()
  })

  it('renders dialog with template info', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    expect(screen.getByText('Create from Template')).toBeInTheDocument()
    expect(screen.getByText(/Test Template.*A test template/)).toBeInTheDocument()
  })

  it('renders document title input', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    expect(screen.getByText('Document Title')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Enter document title')).toBeInTheDocument()
  })

  it('renders variable inputs', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    expect(screen.getByText('studentName')).toBeInTheDocument()
    expect(screen.getByText('date')).toBeInTheDocument()
    expect(screen.getByText('amount')).toBeInTheDocument()
    expect(screen.getByText('type')).toBeInTheDocument()
  })

  it('shows required indicator for required variables', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    // Required variables have asterisk
    const requiredIndicators = screen.getAllByText('*')
    expect(requiredIndicators.length).toBe(3) // studentName, amount, type
  })

  it('initializes variable inputs with default values', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    const input = screen.getByDisplayValue('John')
    expect(input).toBeInTheDocument()
  })

  it('renders different input types based on variable_type', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    // Text input
    expect(screen.getByDisplayValue('John')).toHaveAttribute('type', 'text')
    // Date input
    const dateInputs = document.querySelectorAll('input[type="date"]')
    expect(dateInputs.length).toBe(1)
    // Number input
    const numberInputs = document.querySelectorAll('input[type="number"]')
    expect(numberInputs.length).toBe(1)
    // Select
    expect(screen.getByText('Select an option')).toBeInTheDocument()
  })

  it('handles variable change', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    const input = screen.getByDisplayValue('John')
    fireEvent.change(input, { target: { value: 'Jane' } })
    expect(screen.getByDisplayValue('Jane')).toBeInTheDocument()
  })

  it('calls preview API when preview button clicked', async () => {
    mockApiClient.preview.mockResolvedValueOnce('<p>Preview content</p>')

    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )

    fireEvent.click(screen.getByText('Preview'))

    await waitFor(() => {
      expect(mockApiClient.preview).toHaveBeenCalledWith(1, expect.any(Object))
    })
  })

  it('shows preview content after successful preview', async () => {
    mockApiClient.preview.mockResolvedValueOnce('<p>Preview content</p>')

    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )

    fireEvent.click(screen.getByText('Preview'))

    await waitFor(() => {
      expect(screen.getByText('<p>Preview content</p>')).toBeInTheDocument()
    })
  })

  it('shows error on preview failure', async () => {
    mockApiClient.preview.mockRejectedValueOnce(new Error('API Error'))

    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )

    fireEvent.click(screen.getByText('Preview'))

    await waitFor(() => {
      expect(screen.getByText('Preview failed')).toBeInTheDocument()
    })
  })

  it('disables create button when title is empty', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    const createButton = screen.getByText('Create Document')
    expect(createButton).toBeDisabled()
  })

  it('enables create button when title is provided', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    const titleInput = screen.getByPlaceholderText('Enter document title')
    fireEvent.change(titleInput, { target: { value: 'My Document' } })

    const createButton = screen.getByText('Create Document')
    expect(createButton).not.toBeDisabled()
  })

  it('calls createDocument API when create button clicked', async () => {
    mockApiClient.createDocument.mockResolvedValueOnce({ id: 123 } as never)

    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )

    const titleInput = screen.getByPlaceholderText('Enter document title')
    fireEvent.change(titleInput, { target: { value: 'My Document' } })

    fireEvent.click(screen.getByText('Create Document'))

    await waitFor(() => {
      expect(mockApiClient.createDocument).toHaveBeenCalledWith(1, {
        title: 'My Document',
        variables: expect.any(Object),
      })
    })
  })

  it('shows success state after document creation', async () => {
    mockApiClient.createDocument.mockResolvedValueOnce({ id: 123 } as never)

    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )

    const titleInput = screen.getByPlaceholderText('Enter document title')
    fireEvent.change(titleInput, { target: { value: 'My Document' } })

    fireEvent.click(screen.getByText('Create Document'))

    await waitFor(() => {
      expect(screen.getByText('Document created successfully')).toBeInTheDocument()
      expect(screen.getByText('Redirecting...')).toBeInTheDocument()
    })
  })

  it('shows error on create failure', async () => {
    mockApiClient.createDocument.mockRejectedValueOnce(new Error('API Error'))

    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )

    const titleInput = screen.getByPlaceholderText('Enter document title')
    fireEvent.change(titleInput, { target: { value: 'My Document' } })

    fireEvent.click(screen.getByText('Create Document'))

    await waitFor(() => {
      expect(screen.getByText('Failed to create document')).toBeInTheDocument()
    })
  })

  it('clears preview when variables change', async () => {
    mockApiClient.preview.mockResolvedValueOnce('<p>Preview content</p>')

    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )

    fireEvent.click(screen.getByText('Preview'))

    await waitFor(() => {
      expect(screen.getByText('<p>Preview content</p>')).toBeInTheDocument()
    })

    const input = screen.getByDisplayValue('John')
    fireEvent.change(input, { target: { value: 'Jane' } })

    expect(screen.queryByText('<p>Preview content</p>')).not.toBeInTheDocument()
  })

  it('resets state when template changes', () => {
    const { rerender } = render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )

    const titleInput = screen.getByPlaceholderText('Enter document title')
    fireEvent.change(titleInput, { target: { value: 'My Document' } })

    const newTemplate = { ...mockTemplate, id: 2 }
    rerender(
      <CreateFromTemplateDialog
        template={newTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )

    expect(screen.getByPlaceholderText('Enter document title')).toHaveValue('')
  })

  it('redirects to documents page after successful creation', async () => {
    mockApiClient.createDocument.mockResolvedValueOnce({ id: 123 } as never)

    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )

    const titleInput = screen.getByPlaceholderText('Enter document title')
    fireEvent.change(titleInput, { target: { value: 'My Document' } })

    fireEvent.click(screen.getByText('Create Document'))

    await waitFor(() => {
      expect(screen.getByText('Document created successfully')).toBeInTheDocument()
    })

    // Advance timers to trigger the setTimeout callback
    jest.advanceTimersByTime(1500)

    expect(mockOnOpenChange).toHaveBeenCalledWith(false)
    expect(mockPush).toHaveBeenCalledWith('/documents')
  })

  it('does nothing when preview clicked with null template', async () => {
    render(<CreateFromTemplateDialog template={null} open={true} onOpenChange={mockOnOpenChange} />)

    fireEvent.click(screen.getByText('Preview'))

    // API should not be called
    await waitFor(() => {
      expect(mockApiClient.preview).not.toHaveBeenCalled()
    })
  })

  it('does nothing when create clicked with null template', async () => {
    render(<CreateFromTemplateDialog template={null} open={true} onOpenChange={mockOnOpenChange} />)

    // Try to enable the create button by adding a title
    const titleInput = screen.getByPlaceholderText('Enter document title')
    fireEvent.change(titleInput, { target: { value: 'My Document' } })

    fireEvent.click(screen.getByText('Create Document'))

    // API should not be called because template is null
    await waitFor(() => {
      expect(mockApiClient.createDocument).not.toHaveBeenCalled()
    })
  })

  it('handles date variable change', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    const dateInput = document.querySelector('input[type="date"]') as HTMLInputElement
    fireEvent.change(dateInput, { target: { value: '2024-12-25' } })
    expect(dateInput.value).toBe('2024-12-25')
  })

  it('handles number variable change', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    const numberInput = document.querySelector('input[type="number"]') as HTMLInputElement
    fireEvent.change(numberInput, { target: { value: '100' } })
    expect(numberInput.value).toBe('100')
  })

  it('handles select variable change', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    // Native select element
    const selectElement = document.querySelector('select') as HTMLSelectElement
    fireEvent.change(selectElement, { target: { value: 'Option A' } })

    expect(selectElement.value).toBe('Option A')
  })

  it('closes dialog when cancel button clicked', () => {
    render(
      <CreateFromTemplateDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
      />
    )
    // Mock uses 'cancel' key which returns 'cancel'
    const cancelButton = screen.getByRole('button', { name: /cancel/i })
    fireEvent.click(cancelButton)

    expect(mockOnOpenChange).toHaveBeenCalledWith(false)
  })
})
