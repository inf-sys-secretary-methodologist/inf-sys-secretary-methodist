import { render, screen, fireEvent } from '@testing-library/react'
import { TemplatePreviewDialog } from '../TemplatePreviewDialog'
import type { TemplateInfo } from '@/lib/api/templates'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      info: 'Info',
      variables: 'Variables',
      content: 'Content',
      code: 'Code',
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
    }
    return translations[key] || key
  },
}))

const mockTemplate: TemplateInfo = {
  id: 1,
  name: 'Test Template',
  code: 'test-template',
  description: 'A test template description',
  has_template: true,
  template_content: '<html>Template content here</html>',
  template_variables: [
    {
      name: 'studentName',
      required: true,
      variable_type: 'text',
      description: 'Name of the student',
      default_value: 'John',
    },
    {
      name: 'date',
      required: false,
      variable_type: 'date',
    },
  ],
}

describe('TemplatePreviewDialog', () => {
  const mockOnOpenChange = jest.fn()
  const mockOnCreate = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('returns null when template is null', () => {
    const { container } = render(
      <TemplatePreviewDialog
        template={null}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders template name in title', () => {
    render(
      <TemplatePreviewDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )
    expect(screen.getByText('Test Template')).toBeInTheDocument()
  })

  it('renders template description', () => {
    render(
      <TemplatePreviewDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )
    expect(screen.getByText('A test template description')).toBeInTheDocument()
  })

  it('renders tabs', () => {
    render(
      <TemplatePreviewDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )
    expect(screen.getByText('Info')).toBeInTheDocument()
    expect(screen.getByText('Variables')).toBeInTheDocument()
    expect(screen.getByText('Content')).toBeInTheDocument()
  })

  it('shows info tab by default with template code', () => {
    render(
      <TemplatePreviewDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )
    expect(screen.getByText('test-template')).toBeInTheDocument()
  })

  it('shows variables count in info tab', () => {
    render(
      <TemplatePreviewDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )
    expect(screen.getByText('2')).toBeInTheDocument()
  })

  it('shows Yes when has_template is true', () => {
    render(
      <TemplatePreviewDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )
    expect(screen.getByText('Yes')).toBeInTheDocument()
  })

  it('shows No when has_template is false', () => {
    const templateWithoutTemplate = { ...mockTemplate, has_template: false }
    render(
      <TemplatePreviewDialog
        template={templateWithoutTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )
    expect(screen.getByText('No')).toBeInTheDocument()
  })

  it('renders close button', () => {
    render(
      <TemplatePreviewDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )
    // Find the Close button in the footer (not the dialog X button)
    const closeButtons = screen.getAllByText('Close')
    const footerCloseButton = closeButtons.find((el) =>
      el.closest('button')?.className.includes('border')
    )
    expect(footerCloseButton).toBeInTheDocument()
  })

  it('renders create document button', () => {
    render(
      <TemplatePreviewDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )
    expect(screen.getByText('Create Document')).toBeInTheDocument()
  })

  it('calls onCreate and closes dialog when create button clicked', () => {
    render(
      <TemplatePreviewDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )

    fireEvent.click(screen.getByText('Create Document'))

    expect(mockOnOpenChange).toHaveBeenCalledWith(false)
    expect(mockOnCreate).toHaveBeenCalledWith(mockTemplate)
  })

  it('calls onOpenChange with false when close button clicked', () => {
    render(
      <TemplatePreviewDialog
        template={mockTemplate}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )

    // Find the Close button in the footer (variant="outline" button)
    const closeButtons = screen.getAllByText('Close')
    const footerCloseButton = closeButtons.find((el) =>
      el.closest('button')?.className.includes('border')
    )
    fireEvent.click(footerCloseButton!.closest('button')!)

    expect(mockOnOpenChange).toHaveBeenCalledWith(false)
  })

  it('shows 0 variables count when no variables', () => {
    const templateWithoutVars = { ...mockTemplate, template_variables: undefined }
    render(
      <TemplatePreviewDialog
        template={templateWithoutVars}
        open={true}
        onOpenChange={mockOnOpenChange}
        onCreate={mockOnCreate}
      />
    )
    expect(screen.getByText('0')).toBeInTheDocument()
  })
})
