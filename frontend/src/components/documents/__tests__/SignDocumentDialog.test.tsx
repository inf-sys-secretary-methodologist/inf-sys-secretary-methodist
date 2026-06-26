import { render, screen } from '@/test-utils'
import { fireEvent, waitFor } from '@testing-library/react'

jest.mock('sonner', () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}))

const mockSign = jest.fn()
jest.mock('@/hooks/useDocumentSignatures', () => ({
  signDocument: (...args: unknown[]) => mockSign(...args),
  pickSignatureErrorKey: () => 'generic',
}))

import { toast } from 'sonner'
import { SignDocumentDialog } from '../SignDocumentDialog'

beforeEach(() => {
  mockSign.mockReset()
  ;(toast.success as jest.Mock).mockClear()
  ;(toast.error as jest.Mock).mockClear()
})

describe('SignDocumentDialog', () => {
  it('renders the title when open', () => {
    render(<SignDocumentDialog documentId={1} open onClose={jest.fn()} />)
    expect(screen.getByText('signDialog.title')).toBeInTheDocument()
  })

  it('signs the document and fires onSigned + success toast', async () => {
    mockSign.mockResolvedValueOnce({ id: 7 })
    const onSigned = jest.fn()
    const onClose = jest.fn()
    render(<SignDocumentDialog documentId={42} open onClose={onClose} onSigned={onSigned} />)

    fireEvent.click(screen.getByRole('button', { name: 'signDialog.confirm' }))

    await waitFor(() => expect(mockSign).toHaveBeenCalledWith(42))
    expect(toast.success).toHaveBeenCalledWith('signDialog.successToast')
    expect(onSigned).toHaveBeenCalled()
    expect(onClose).toHaveBeenCalled()
  })

  it('shows an error toast when signing fails', async () => {
    mockSign.mockRejectedValueOnce(new Error('boom'))
    render(<SignDocumentDialog documentId={42} open onClose={jest.fn()} />)

    fireEvent.click(screen.getByRole('button', { name: 'signDialog.confirm' }))

    await waitFor(() => expect(toast.error).toHaveBeenCalledWith('errors.generic'))
  })
})
