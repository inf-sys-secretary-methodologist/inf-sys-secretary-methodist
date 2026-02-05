import { render, screen, fireEvent } from '@testing-library/react'
import { TemplateCard } from '../TemplateCard'
import type { TemplateInfo } from '@/lib/api/templates'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      variables: 'variables',
      code: 'Code',
      preview: 'Preview',
      create: 'Create',
    }
    return translations[key] || key
  },
}))

// Mock GlowingEffect
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => null,
}))

const mockTemplate: TemplateInfo = {
  id: 1,
  name: 'Test Template',
  code: 'test-template',
  description: 'A test template description',
  has_template: true,
  template_variables: [
    { name: 'name', required: true, variable_type: 'text' },
    { name: 'date', required: false, variable_type: 'date' },
  ],
}

describe('TemplateCard', () => {
  const mockOnPreview = jest.fn()
  const mockOnCreate = jest.fn()
  const mockOnEdit = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders template name', () => {
    render(
      <TemplateCard template={mockTemplate} onPreview={mockOnPreview} onCreate={mockOnCreate} />
    )
    expect(screen.getByText('Test Template')).toBeInTheDocument()
  })

  it('renders template description', () => {
    render(
      <TemplateCard template={mockTemplate} onPreview={mockOnPreview} onCreate={mockOnCreate} />
    )
    expect(screen.getByText('A test template description')).toBeInTheDocument()
  })

  it('renders template code', () => {
    render(
      <TemplateCard template={mockTemplate} onPreview={mockOnPreview} onCreate={mockOnCreate} />
    )
    expect(screen.getByText(/Code:.*test-template/)).toBeInTheDocument()
  })

  it('renders variables count badge', () => {
    render(
      <TemplateCard template={mockTemplate} onPreview={mockOnPreview} onCreate={mockOnCreate} />
    )
    expect(screen.getByText('2 variables')).toBeInTheDocument()
  })

  it('does not render variables badge when no variables', () => {
    const templateWithoutVars = { ...mockTemplate, template_variables: [] }
    render(
      <TemplateCard
        template={templateWithoutVars}
        onPreview={mockOnPreview}
        onCreate={mockOnCreate}
      />
    )
    expect(screen.queryByText(/variables/)).not.toBeInTheDocument()
  })

  it('does not render description when not provided', () => {
    const templateWithoutDesc = { ...mockTemplate, description: undefined }
    render(
      <TemplateCard
        template={templateWithoutDesc}
        onPreview={mockOnPreview}
        onCreate={mockOnCreate}
      />
    )
    expect(screen.queryByText('A test template description')).not.toBeInTheDocument()
  })

  it('calls onPreview when preview button clicked', () => {
    render(
      <TemplateCard template={mockTemplate} onPreview={mockOnPreview} onCreate={mockOnCreate} />
    )

    fireEvent.click(screen.getByText('Preview'))
    expect(mockOnPreview).toHaveBeenCalledWith(mockTemplate)
  })

  it('calls onCreate when create button clicked', () => {
    render(
      <TemplateCard template={mockTemplate} onPreview={mockOnPreview} onCreate={mockOnCreate} />
    )

    fireEvent.click(screen.getByText('Create'))
    expect(mockOnCreate).toHaveBeenCalledWith(mockTemplate)
  })

  it('renders edit button when canEdit is true', () => {
    render(
      <TemplateCard
        template={mockTemplate}
        onPreview={mockOnPreview}
        onCreate={mockOnCreate}
        onEdit={mockOnEdit}
        canEdit={true}
      />
    )

    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBe(3) // Preview, Create, Edit
  })

  it('does not render edit button when canEdit is false', () => {
    render(
      <TemplateCard
        template={mockTemplate}
        onPreview={mockOnPreview}
        onCreate={mockOnCreate}
        onEdit={mockOnEdit}
        canEdit={false}
      />
    )

    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBe(2) // Preview, Create only
  })

  it('calls onEdit when edit button clicked', () => {
    render(
      <TemplateCard
        template={mockTemplate}
        onPreview={mockOnPreview}
        onCreate={mockOnCreate}
        onEdit={mockOnEdit}
        canEdit={true}
      />
    )

    const buttons = screen.getAllByRole('button')
    fireEvent.click(buttons[2]) // Edit is the third button
    expect(mockOnEdit).toHaveBeenCalledWith(mockTemplate)
  })

  it('applies custom className', () => {
    const { container } = render(
      <TemplateCard
        template={mockTemplate}
        onPreview={mockOnPreview}
        onCreate={mockOnCreate}
        className="custom-class"
      />
    )
    expect(container.querySelector('.custom-class')).toBeInTheDocument()
  })

  it('handles undefined template_variables', () => {
    const templateWithUndefinedVars = { ...mockTemplate, template_variables: undefined }
    render(
      <TemplateCard
        template={templateWithUndefinedVars}
        onPreview={mockOnPreview}
        onCreate={mockOnCreate}
      />
    )
    expect(screen.queryByText(/variables/)).not.toBeInTheDocument()
  })
})
