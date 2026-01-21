import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ExportButton } from '../ExportButton'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      title: 'Export',
      pdf: 'PDF',
      excel: 'Excel',
    }
    return translations[key] || key
  },
}))

// Mock useDashboard hook
const mockExportDashboard = jest.fn()
jest.mock('@/hooks/useDashboard', () => ({
  exportDashboard: (...args: unknown[]) => mockExportDashboard(...args),
}))

describe('ExportButton', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockExportDashboard.mockResolvedValue({
      file_url: 'https://example.com/export.pdf',
      file_name: 'dashboard-export.pdf',
    })
  })

  it('renders export button', () => {
    render(<ExportButton />)
    expect(screen.getByRole('button', { name: /export/i })).toBeInTheDocument()
  })

  it('shows dropdown menu when clicked', async () => {
    render(<ExportButton />)

    await userEvent.click(screen.getByRole('button', { name: /export/i }))

    expect(screen.getByText('PDF')).toBeInTheDocument()
    expect(screen.getByText('Excel')).toBeInTheDocument()
  })

  it('hides dropdown when clicked again', async () => {
    render(<ExportButton />)

    await userEvent.click(screen.getByRole('button', { name: /export/i }))
    expect(screen.getByText('PDF')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: /export/i }))
    expect(screen.queryByText('PDF')).not.toBeInTheDocument()
  })

  it('calls exportDashboard with pdf format when PDF is clicked', async () => {
    render(<ExportButton />)

    await userEvent.click(screen.getByRole('button', { name: /export/i }))
    await userEvent.click(screen.getByText('PDF'))

    await waitFor(() => {
      expect(mockExportDashboard).toHaveBeenCalledWith({
        format: 'pdf',
        sections: ['stats', 'trends', 'activity'],
      })
    })
  })

  it('calls exportDashboard with xlsx format when Excel is clicked', async () => {
    render(<ExportButton />)

    await userEvent.click(screen.getByRole('button', { name: /export/i }))
    await userEvent.click(screen.getByText('Excel'))

    await waitFor(() => {
      expect(mockExportDashboard).toHaveBeenCalledWith({
        format: 'xlsx',
        sections: ['stats', 'trends', 'activity'],
      })
    })
  })

  it('closes dropdown when export option is clicked', async () => {
    render(<ExportButton />)

    await userEvent.click(screen.getByRole('button', { name: /export/i }))
    await userEvent.click(screen.getByText('PDF'))

    await waitFor(() => {
      expect(screen.queryByText('Excel')).not.toBeInTheDocument()
    })
  })

  it('applies custom className', () => {
    const { container } = render(<ExportButton className="custom-export" />)
    expect(container.firstChild).toHaveClass('custom-export')
  })

  it('handles export error gracefully', async () => {
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {})
    mockExportDashboard.mockRejectedValue(new Error('Export failed'))

    render(<ExportButton />)

    await userEvent.click(screen.getByRole('button', { name: /export/i }))
    await userEvent.click(screen.getByText('PDF'))

    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalledWith('Export failed:', expect.any(Error))
    })

    consoleSpy.mockRestore()
  })
})
