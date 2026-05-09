import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { TemplateEditorDialog } from '../TemplateEditorDialog'
import { templatesApi } from '@/lib/api/templates'
import type { TemplateInfo } from '@/lib/api/templates'

jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => key,
}))

jest.mock('@/lib/api/templates', () => ({
  templatesApi: {
    update: jest.fn(),
  },
}))

const baseTemplate: TemplateInfo = {
  id: 42,
  name: 'Test Template',
  code: 'test-template',
  description: 'A template',
  template_content: 'Original {{name}}',
  template_variables: [],
  has_template: true,
  methodist_only: false,
}

describe('TemplateEditorDialog — methodist_only toggle (v0.126.3)', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders the methodist-only toggle reflecting initial template value', () => {
    render(
      <TemplateEditorDialog
        template={{ ...baseTemplate, methodist_only: true }}
        open
        onOpenChange={jest.fn()}
      />
    )
    const toggle = screen.getByLabelText('methodistOnlyLabel') as HTMLInputElement
    expect(toggle.checked).toBe(true)
  })

  it('renders unchecked toggle when methodist_only is false', () => {
    render(<TemplateEditorDialog template={baseTemplate} open onOpenChange={jest.fn()} />)
    const toggle = screen.getByLabelText('methodistOnlyLabel') as HTMLInputElement
    expect(toggle.checked).toBe(false)
  })

  it('flipping toggle marks the form dirty (Save enabled)', () => {
    render(<TemplateEditorDialog template={baseTemplate} open onOpenChange={jest.fn()} />)
    const saveBtn = screen.getByRole('button', { name: /save/i })
    expect(saveBtn).toBeDisabled()

    const toggle = screen.getByLabelText('methodistOnlyLabel') as HTMLInputElement
    fireEvent.click(toggle)

    expect(saveBtn).not.toBeDisabled()
  })

  it('save sends methodist_only when toggle changed', async () => {
    const updateMock = templatesApi.update as jest.MockedFunction<typeof templatesApi.update>
    updateMock.mockResolvedValueOnce(undefined)

    render(<TemplateEditorDialog template={baseTemplate} open onOpenChange={jest.fn()} />)
    fireEvent.click(screen.getByLabelText('methodistOnlyLabel'))
    fireEvent.click(screen.getByRole('button', { name: /save/i }))

    await waitFor(() => {
      expect(updateMock).toHaveBeenCalledWith(42, expect.objectContaining({ methodist_only: true }))
    })
  })

  it('save omits methodist_only when toggle untouched (preserve server value)', async () => {
    const updateMock = templatesApi.update as jest.MockedFunction<typeof templatesApi.update>
    updateMock.mockResolvedValueOnce(undefined)

    render(<TemplateEditorDialog template={baseTemplate} open onOpenChange={jest.fn()} />)
    // Touch only the content textarea so hasChanges flips and Save enables.
    const textarea = screen.getByPlaceholderText('contentPlaceholder')
    fireEvent.change(textarea, { target: { value: 'Modified' } })
    fireEvent.click(screen.getByRole('button', { name: /save/i }))

    await waitFor(() => {
      expect(updateMock).toHaveBeenCalledTimes(1)
    })
    const payload = updateMock.mock.calls[0][1]
    expect(payload).not.toHaveProperty('methodist_only')
  })
})
