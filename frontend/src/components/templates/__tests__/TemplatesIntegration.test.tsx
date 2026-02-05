import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import { TemplateList } from '../TemplateList'
import { TemplatePreviewDialog } from '../TemplatePreviewDialog'
import { CreateFromTemplateDialog } from '../CreateFromTemplateDialog'
import { templatesApi } from '@/lib/api/templates'
import type { TemplateInfo } from '@/lib/api/templates'
import { useState } from 'react'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: (namespace: string) => (key: string) => {
    const translations: Record<string, Record<string, string>> = {
      templates: {
        loadError: 'Failed to load templates',
        searchPlaceholder: 'Search templates...',
        noSearchResults: 'No templates match your search',
        noTemplates: 'No templates available',
        variables: 'variables',
        code: 'Code',
        preview: 'Preview',
        create: 'Create',
        info: 'Info',
        content: 'Content',
        variablesCount: 'Variables Count',
        hasTemplate: 'Has Template',
        yes: 'Yes',
        no: 'No',
        required: 'Required',
        defaultValue: 'Default',
        noVariables: 'No variables defined',
        noContent: 'No content available',
        close: 'Close',
        createFromThis: 'Create Document',
        createFromTemplate: 'Create from Template',
        documentTitle: 'Document Title',
        enterTitle: 'Enter document title',
        fillVariables: 'Fill Variables',
        selectOption: 'Select an option',
        previewTitle: 'Preview',
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
    getAll: jest.fn(),
    preview: jest.fn(),
    createDocument: jest.fn(),
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
    template_content: '<html>Invoice content</html>',
    template_variables: [
      { name: 'amount', required: true, variable_type: 'number' },
      { name: 'date', required: false, variable_type: 'date' },
    ],
  },
  {
    id: 2,
    name: 'Contract Template',
    code: 'contract',
    description: 'Legal contract template',
    has_template: true,
    template_content: '<html>Contract content</html>',
    template_variables: [
      { name: 'partyName', required: true, variable_type: 'text', default_value: 'John Doe' },
    ],
  },
]

const mockApiClient = templatesApi as jest.Mocked<typeof templatesApi>

// Integration test component that combines all template components
function TemplatesIntegrationTest() {
  const [previewTemplate, setPreviewTemplate] = useState<TemplateInfo | null>(null)
  const [createTemplate, setCreateTemplate] = useState<TemplateInfo | null>(null)
  const [previewOpen, setPreviewOpen] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)

  const handlePreview = (template: TemplateInfo) => {
    setPreviewTemplate(template)
    setPreviewOpen(true)
  }

  const handleCreate = (template: TemplateInfo) => {
    setCreateTemplate(template)
    setCreateOpen(true)
  }

  const handleCreateFromPreview = (template: TemplateInfo) => {
    setPreviewOpen(false)
    setCreateTemplate(template)
    setCreateOpen(true)
  }

  return (
    <>
      <TemplateList onPreview={handlePreview} onCreate={handleCreate} />
      <TemplatePreviewDialog
        template={previewTemplate}
        open={previewOpen}
        onOpenChange={setPreviewOpen}
        onCreate={handleCreateFromPreview}
      />
      <CreateFromTemplateDialog
        template={createTemplate}
        open={createOpen}
        onOpenChange={setCreateOpen}
      />
    </>
  )
}

describe('Templates Integration Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('loads and displays templates from API', async () => {
    mockApiClient.getAll.mockResolvedValueOnce(mockTemplates)

    render(<TemplatesIntegrationTest />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
      expect(screen.getByText('Contract Template')).toBeInTheDocument()
    })
  })

  it('opens preview dialog when clicking preview button', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(<TemplatesIntegrationTest />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    const previewButtons = screen.getAllByText('Preview')
    fireEvent.click(previewButtons[0])

    await waitFor(() => {
      expect(screen.getByText('Info')).toBeInTheDocument()
      expect(screen.getByText('variables')).toBeInTheDocument()
      expect(screen.getByText('Content')).toBeInTheDocument()
    })
  })

  it('opens create dialog when clicking create button', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(<TemplatesIntegrationTest />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    const createButtons = screen.getAllByText('Create')
    fireEvent.click(createButtons[0])

    await waitFor(() => {
      expect(screen.getByText('Create from Template')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Enter document title')).toBeInTheDocument()
    })
  })

  it('transitions from preview to create dialog', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(<TemplatesIntegrationTest />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    // Open preview
    const previewButtons = screen.getAllByText('Preview')
    fireEvent.click(previewButtons[0])

    await waitFor(() => {
      expect(screen.getByText('Info')).toBeInTheDocument()
    })

    // Click create from preview
    const createFromPreviewButtons = screen.getAllByText('Create Document')
    fireEvent.click(createFromPreviewButtons[0])

    await waitFor(() => {
      expect(screen.getByText('Create from Template')).toBeInTheDocument()
    })
  })

  it('filters templates and opens preview for filtered result', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)

    render(<TemplatesIntegrationTest />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    // Filter to only show contract
    const searchInput = screen.getByPlaceholderText('Search templates...')
    await act(async () => {
      fireEvent.change(searchInput, { target: { value: 'contract' } })
    })

    await waitFor(() => {
      expect(screen.queryByText('Invoice Template')).not.toBeInTheDocument()
      expect(screen.getByText('Contract Template')).toBeInTheDocument()
    })

    // Open preview for filtered template
    const previewButton = screen.getByText('Preview')
    fireEvent.click(previewButton)

    await waitFor(() => {
      expect(screen.getByText('Contract Template')).toBeInTheDocument()
    })
  })

  it('complete flow: search, preview, create document', async () => {
    mockApiClient.getAll.mockResolvedValue(mockTemplates)
    mockApiClient.createDocument.mockResolvedValueOnce({ id: 123 } as never)

    render(<TemplatesIntegrationTest />)

    // Wait for templates to load
    await waitFor(() => {
      expect(screen.getByText('Contract Template')).toBeInTheDocument()
    })

    // Search for contract
    const searchInput = screen.getByPlaceholderText('Search templates...')
    await act(async () => {
      fireEvent.change(searchInput, { target: { value: 'contract' } })
    })

    await waitFor(() => {
      expect(screen.queryByText('Invoice Template')).not.toBeInTheDocument()
    })

    // Click create on contract template
    const createButton = screen.getByText('Create')
    fireEvent.click(createButton)

    await waitFor(() => {
      expect(screen.getByText('Create from Template')).toBeInTheDocument()
    })

    // Fill in title
    const titleInput = screen.getByPlaceholderText('Enter document title')
    fireEvent.change(titleInput, { target: { value: 'My Contract Document' } })

    // Click create
    const createDocButton = screen.getByRole('button', { name: 'Create Document' })
    fireEvent.click(createDocButton)

    await waitFor(() => {
      expect(mockApiClient.createDocument).toHaveBeenCalledWith(2, {
        title: 'My Contract Document',
        variables: { partyName: 'John Doe' },
      })
    })
  })

  it('handles API error gracefully', async () => {
    mockApiClient.getAll.mockRejectedValueOnce(new Error('Network error'))

    render(<TemplatesIntegrationTest />)

    await waitFor(() => {
      expect(screen.getByText('Failed to load templates')).toBeInTheDocument()
    })
  })

  it('shows empty state when no templates match search', async () => {
    mockApiClient.getAll.mockResolvedValueOnce(mockTemplates)

    render(<TemplatesIntegrationTest />)

    await waitFor(() => {
      expect(screen.getByText('Invoice Template')).toBeInTheDocument()
    })

    // Search for non-existent template
    const searchInput = screen.getByPlaceholderText('Search templates...')
    await act(async () => {
      fireEvent.change(searchInput, { target: { value: 'nonexistent' } })
    })

    await waitFor(() => {
      expect(screen.getByText('No templates match your search')).toBeInTheDocument()
    })
  })
})
