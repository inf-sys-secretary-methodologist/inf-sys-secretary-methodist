import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { GenerateOrderRevisionsDialog } from '../GenerateOrderRevisionsDialog'

const mockGenerate = jest.fn()
jest.mock('@/hooks/useMinobrnaukiOrders', () => ({
  ...jest.requireActual('@/hooks/useMinobrnaukiOrders'),
  generateOrderRevisions: (...args: unknown[]) => mockGenerate(...args),
}))

const mockToastSuccess = jest.fn()
const mockToastError = jest.fn()
jest.mock('sonner', () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}))

const noop = () => {}

beforeEach(() => jest.clearAllMocks())

describe('GenerateOrderRevisionsDialog', () => {
  it('renders nothing when open=false', () => {
    render(<GenerateOrderRevisionsDialog orderId={7} open={false} onClose={noop} />)
    expect(screen.queryByText('generateDialog.title')).not.toBeInTheDocument()
  })

  it('shows the confirmation prompt when open', () => {
    render(<GenerateOrderRevisionsDialog orderId={7} open={true} onClose={noop} />)
    expect(screen.getByText('generateDialog.title')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'generateDialog.confirm' })).toBeInTheDocument()
  })

  it('triggers generation, then shows the run-summary counts + success toast + onGenerated', async () => {
    mockGenerate.mockResolvedValue({ generated: 3, skipped: 1, failures: 0 })
    const onGenerated = jest.fn()
    render(
      <GenerateOrderRevisionsDialog
        orderId={7}
        open={true}
        onClose={noop}
        onGenerated={onGenerated}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: 'generateDialog.confirm' }))

    await waitFor(() => expect(mockGenerate).toHaveBeenCalledWith(7))
    expect(mockToastSuccess).toHaveBeenCalled()
    expect(onGenerated).toHaveBeenCalled()
    // result panel appears with the counts
    expect(screen.getByText('generateDialog.result.title')).toBeInTheDocument()
    expect(screen.getByTestId('generate-result-generated')).toHaveTextContent('3')
    expect(screen.getByTestId('generate-result-skipped')).toHaveTextContent('1')
    expect(screen.getByTestId('generate-result-failures')).toHaveTextContent('0')
  })

  it('maps a backend error to a toast and keeps the dialog open (no result panel)', async () => {
    mockGenerate.mockRejectedValue({
      response: { status: 429, data: { error: { code: 'RATE_LIMITED' } } },
    })
    const onClose = jest.fn()
    render(<GenerateOrderRevisionsDialog orderId={7} open={true} onClose={onClose} />)

    fireEvent.click(screen.getByRole('button', { name: 'generateDialog.confirm' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(onClose).not.toHaveBeenCalled()
    expect(screen.queryByText('generateDialog.result.title')).not.toBeInTheDocument()
  })
})
