import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { RecordMinobrnaukiOrderDialog } from '../RecordMinobrnaukiOrderDialog'

const mockRecord = jest.fn()
jest.mock('@/hooks/useMinobrnaukiOrders', () => ({
  ...jest.requireActual('@/hooks/useMinobrnaukiOrders'),
  recordMinobrnaukiOrder: (...args: unknown[]) => mockRecord(...args),
}))

const mockDocCreate = jest.fn()
const mockDocUpload = jest.fn()
jest.mock('@/lib/api/documents', () => ({
  documentsApi: {
    create: (...args: unknown[]) => mockDocCreate(...args),
    uploadFile: (...args: unknown[]) => mockDocUpload(...args),
  },
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

function fillRequired() {
  fireEvent.change(screen.getByLabelText('recordDialog.labels.orderNumber'), {
    target: { value: '№ 777' },
  })
  fireEvent.change(screen.getByLabelText('recordDialog.labels.title'), {
    target: { value: 'Новый приказ' },
  })
  fireEvent.change(screen.getByLabelText('recordDialog.labels.publishedAt'), {
    target: { value: '2026-04-01' },
  })
}

describe('RecordMinobrnaukiOrderDialog', () => {
  it('renders no fields when open=false', () => {
    render(<RecordMinobrnaukiOrderDialog open={false} onClose={noop} />)
    expect(screen.queryByLabelText('recordDialog.labels.title')).not.toBeInTheDocument()
  })

  it('starts with empty required inputs', () => {
    render(<RecordMinobrnaukiOrderDialog open={true} onClose={noop} />)
    expect(screen.getByLabelText('recordDialog.labels.orderNumber')).toHaveValue('')
    expect(screen.getByLabelText('recordDialog.labels.title')).toHaveValue('')
  })

  it('submits the parsed input and calls onCreated + success toast', async () => {
    mockRecord.mockResolvedValue({ id: 9 })
    const onClose = jest.fn()
    const onCreated = jest.fn()
    render(<RecordMinobrnaukiOrderDialog open={true} onClose={onClose} onCreated={onCreated} />)

    fillRequired()
    fireEvent.change(screen.getByLabelText('recordDialog.labels.affected'), {
      target: { value: '10, 11' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'recordDialog.record' }))

    await waitFor(() => expect(mockRecord).toHaveBeenCalledTimes(1))
    expect(mockRecord).toHaveBeenCalledWith({
      order_number: '№ 777',
      title: 'Новый приказ',
      published_at: '2026-04-01',
      change_scope: 'minor',
      summary: '',
      affected_work_program_ids: [10, 11],
    })
    expect(mockToastSuccess).toHaveBeenCalled()
    expect(onCreated).toHaveBeenCalled()
  })

  it('keeps the Record button disabled until required fields are filled', () => {
    render(<RecordMinobrnaukiOrderDialog open={true} onClose={noop} />)
    expect(screen.getByRole('button', { name: 'recordDialog.record' })).toBeDisabled()
    fillRequired()
    expect(screen.getByRole('button', { name: 'recordDialog.record' })).not.toBeDisabled()
  })

  it('uploads an attached file first, then records the order with its document id', async () => {
    mockDocCreate.mockResolvedValue({ id: 42 })
    mockDocUpload.mockResolvedValue({ id: 42 })
    mockRecord.mockResolvedValue({ id: 9 })
    const onCreated = jest.fn()
    render(<RecordMinobrnaukiOrderDialog open={true} onClose={noop} onCreated={onCreated} />)

    fillRequired()
    const file = new File(['%PDF-1.7 ...'], 'prikaz-1234.pdf', { type: 'application/pdf' })
    fireEvent.change(screen.getByLabelText('recordDialog.labels.document'), {
      target: { files: [file] },
    })
    fireEvent.click(screen.getByRole('button', { name: 'recordDialog.record' }))

    // Document is created + the file uploaded before the order is recorded.
    await waitFor(() => expect(mockDocCreate).toHaveBeenCalledTimes(1))
    expect(mockDocUpload).toHaveBeenCalledWith(42, file)
    await waitFor(() => expect(mockRecord).toHaveBeenCalledTimes(1))
    expect(mockRecord).toHaveBeenCalledWith(expect.objectContaining({ document_id: 42 }))
    expect(onCreated).toHaveBeenCalled()
  })

  it('records without a document_id when no file is attached', async () => {
    mockRecord.mockResolvedValue({ id: 9 })
    render(<RecordMinobrnaukiOrderDialog open={true} onClose={noop} />)

    fillRequired()
    fireEvent.click(screen.getByRole('button', { name: 'recordDialog.record' }))

    await waitFor(() => expect(mockRecord).toHaveBeenCalledTimes(1))
    expect(mockDocCreate).not.toHaveBeenCalled()
    expect(mockRecord.mock.calls[0][0].document_id).toBeUndefined()
  })

  it('shows an error toast and does not record when the file upload fails', async () => {
    mockDocCreate.mockRejectedValue(new Error('upload failed'))
    const onClose = jest.fn()
    render(<RecordMinobrnaukiOrderDialog open={true} onClose={onClose} />)

    fillRequired()
    const file = new File(['x'], 'order.docx', {
      type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    })
    fireEvent.change(screen.getByLabelText('recordDialog.labels.document'), {
      target: { files: [file] },
    })
    fireEvent.click(screen.getByRole('button', { name: 'recordDialog.record' }))

    // The toast names the upload as the failure, not a generic "record failed".
    await waitFor(() =>
      expect(mockToastError).toHaveBeenCalledWith('recordDialog.errors.uploadFailed')
    )
    expect(mockRecord).not.toHaveBeenCalled()
    expect(onClose).not.toHaveBeenCalled()
  })

  it('maps a backend error to a toast and keeps the dialog open', async () => {
    mockRecord.mockRejectedValue({
      response: { data: { error: { code: 'INVALID_MINOBRNAUKI_ORDER' } } },
    })
    const onClose = jest.fn()
    render(<RecordMinobrnaukiOrderDialog open={true} onClose={onClose} />)

    fillRequired()
    fireEvent.click(screen.getByRole('button', { name: 'recordDialog.record' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(onClose).not.toHaveBeenCalled()
  })
})
