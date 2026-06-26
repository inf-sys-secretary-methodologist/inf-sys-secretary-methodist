import { render, screen } from '@/test-utils'
import { fireEvent, waitFor } from '@testing-library/react'

const mockUseSignatures = jest.fn()
const mockVerify = jest.fn()
jest.mock('@/hooks/useDocumentSignatures', () => ({
  useDocumentSignatures: (...args: unknown[]) => mockUseSignatures(...args),
  verifySignature: (...args: unknown[]) => mockVerify(...args),
}))

import { DocumentSignaturesPanel } from '../DocumentSignaturesPanel'

const signature = {
  id: 7,
  document_id: 42,
  document_version: 3,
  signer_id: 9,
  signer_name: 'Иванов И.И.',
  algorithm: 'ECDSA_P256_SHA256',
  digest_hex: 'ab12',
  signature_base64: 'MEQ=',
  certificate_pem: 'pem',
  signed_at: '2026-06-26T12:00:00Z',
}

beforeEach(() => {
  mockUseSignatures.mockReset()
  mockVerify.mockReset()
})

describe('DocumentSignaturesPanel', () => {
  it('shows the empty state when there are no signatures', () => {
    mockUseSignatures.mockReturnValue({ signatures: [], isLoading: false })
    render(<DocumentSignaturesPanel documentId={42} />)
    expect(screen.getByText('panel.empty')).toBeInTheDocument()
  })

  it('lists signatures with the signer name', () => {
    mockUseSignatures.mockReturnValue({ signatures: [signature], isLoading: false })
    render(<DocumentSignaturesPanel documentId={42} />)
    expect(screen.getByText('Иванов И.И.')).toBeInTheDocument()
  })

  it('verifies a signature and shows a valid badge', async () => {
    mockUseSignatures.mockReturnValue({ signatures: [signature], isLoading: false })
    mockVerify.mockResolvedValueOnce({
      signature_id: 7,
      valid: true,
      digest_match: true,
      crypto_valid: true,
      version_changed: false,
      status: 'valid',
    })
    render(<DocumentSignaturesPanel documentId={42} />)

    fireEvent.click(screen.getByRole('button', { name: 'actions.verify' }))

    await waitFor(() => expect(mockVerify).toHaveBeenCalledWith(42, 7))
    expect(screen.getByText('verdicts.valid')).toBeInTheDocument()
  })

  it('shows a document-modified badge when the digest no longer matches', async () => {
    mockUseSignatures.mockReturnValue({ signatures: [signature], isLoading: false })
    mockVerify.mockResolvedValueOnce({
      signature_id: 7,
      valid: false,
      digest_match: false,
      crypto_valid: false,
      version_changed: false,
      status: 'document_modified',
    })
    render(<DocumentSignaturesPanel documentId={42} />)

    fireEvent.click(screen.getByRole('button', { name: 'actions.verify' }))

    await waitFor(() => expect(screen.getByText('verdicts.document_modified')).toBeInTheDocument())
  })
})
