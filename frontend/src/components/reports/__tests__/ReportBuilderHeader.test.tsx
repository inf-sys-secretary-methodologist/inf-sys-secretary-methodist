import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ReportBuilderHeader } from '../ReportBuilderHeader'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      save: 'Save',
      export: 'Export',
      exportPdf: 'Export as PDF',
      exportExcel: 'Export as Excel',
      exportCsv: 'Export as CSV',
    }
    return translations[key] || key
  },
}))

describe('ReportBuilderHeader', () => {
  const defaultProps = {
    reportName: 'Test Report',
    onNameChange: jest.fn(),
    onSave: jest.fn(),
    onExport: jest.fn(),
    onBack: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders report name', () => {
    render(<ReportBuilderHeader {...defaultProps} />)
    expect(screen.getByText('Test Report')).toBeInTheDocument()
  })

  it('renders back button', () => {
    render(<ReportBuilderHeader {...defaultProps} />)
    const buttons = screen.getAllByRole('button')
    // First button should be back button with ArrowLeft icon
    expect(buttons[0]).toBeInTheDocument()
  })

  it('renders save button', () => {
    render(<ReportBuilderHeader {...defaultProps} />)
    expect(screen.getByText('Save')).toBeInTheDocument()
  })

  it('renders export button', () => {
    render(<ReportBuilderHeader {...defaultProps} />)
    expect(screen.getByText('Export')).toBeInTheDocument()
  })

  it('calls onBack when back button is clicked', async () => {
    const user = userEvent.setup()
    const onBack = jest.fn()
    render(<ReportBuilderHeader {...defaultProps} onBack={onBack} />)

    const buttons = screen.getAllByRole('button')
    await user.click(buttons[0]) // Back button is first
    expect(onBack).toHaveBeenCalled()
  })

  it('calls onSave when save button is clicked', async () => {
    const user = userEvent.setup()
    const onSave = jest.fn()
    render(<ReportBuilderHeader {...defaultProps} onSave={onSave} />)

    await user.click(screen.getByText('Save'))
    expect(onSave).toHaveBeenCalled()
  })

  it('shows export dropdown menu on click', async () => {
    const user = userEvent.setup()
    render(<ReportBuilderHeader {...defaultProps} />)

    await user.click(screen.getByText('Export'))

    await waitFor(() => {
      expect(screen.getByText('Export as PDF')).toBeInTheDocument()
      expect(screen.getByText('Export as Excel')).toBeInTheDocument()
      expect(screen.getByText('Export as CSV')).toBeInTheDocument()
    })
  })

  it('calls onExport with "pdf" when PDF option is clicked', async () => {
    const user = userEvent.setup()
    const onExport = jest.fn()
    render(<ReportBuilderHeader {...defaultProps} onExport={onExport} />)

    await user.click(screen.getByText('Export'))
    await waitFor(() => {
      expect(screen.getByText('Export as PDF')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Export as PDF'))
    expect(onExport).toHaveBeenCalledWith('pdf')
  })

  it('calls onExport with "xlsx" when Excel option is clicked', async () => {
    const user = userEvent.setup()
    const onExport = jest.fn()
    render(<ReportBuilderHeader {...defaultProps} onExport={onExport} />)

    await user.click(screen.getByText('Export'))
    await waitFor(() => {
      expect(screen.getByText('Export as Excel')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Export as Excel'))
    expect(onExport).toHaveBeenCalledWith('xlsx')
  })

  it('calls onExport with "csv" when CSV option is clicked', async () => {
    const user = userEvent.setup()
    const onExport = jest.fn()
    render(<ReportBuilderHeader {...defaultProps} onExport={onExport} />)

    await user.click(screen.getByText('Export'))
    await waitFor(() => {
      expect(screen.getByText('Export as CSV')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Export as CSV'))
    expect(onExport).toHaveBeenCalledWith('csv')
  })

  it('allows editing report name on click', async () => {
    const user = userEvent.setup()
    const onNameChange = jest.fn()
    render(<ReportBuilderHeader {...defaultProps} onNameChange={onNameChange} />)

    // Click on report name to start editing
    await user.click(screen.getByText('Test Report'))

    // Should show input field
    const input = screen.getByRole('textbox')
    expect(input).toBeInTheDocument()
    expect(input).toHaveValue('Test Report')
  })

  it('calls onNameChange when editing report name', async () => {
    const user = userEvent.setup()
    const onNameChange = jest.fn()
    render(<ReportBuilderHeader {...defaultProps} onNameChange={onNameChange} />)

    // Click to edit
    await user.click(screen.getByText('Test Report'))

    // Type new name
    const input = screen.getByRole('textbox')
    await user.clear(input)
    await user.type(input, 'New Report Name')

    expect(onNameChange).toHaveBeenCalled()
  })

  it('exits edit mode on blur', async () => {
    const user = userEvent.setup()
    render(<ReportBuilderHeader {...defaultProps} />)

    // Click to edit
    await user.click(screen.getByText('Test Report'))

    // Should show input
    const input = screen.getByRole('textbox')
    expect(input).toBeInTheDocument()

    // Blur the input
    await user.tab()

    // Should show text again (not input)
    await waitFor(() => {
      expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    })
  })

  it('exits edit mode on Enter key', async () => {
    const user = userEvent.setup()
    render(<ReportBuilderHeader {...defaultProps} />)

    // Click to edit
    await user.click(screen.getByText('Test Report'))

    // Press Enter
    const input = screen.getByRole('textbox')
    await user.type(input, '{Enter}')

    // Should show text again
    await waitFor(() => {
      expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    })
  })
})
