import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useDocumentSignatures,
  signDocument,
  verifySignature,
  pickSignatureErrorKey,
} from '../useDocumentSignatures'
import { canSignDocuments } from '@/lib/auth/permissions'
import { UserRole } from '@/types/auth'
import { apiClient } from '@/lib/api'

jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(
    SWRConfig,
    { value: { dedupingInterval: 0, provider: () => new Map() } },
    children
  )

const signatureFixture = {
  id: 7,
  document_id: 42,
  document_version: 3,
  signer_id: 9,
  signer_name: 'Иванов И.И.',
  algorithm: 'ECDSA_P256_SHA256',
  digest_hex: 'ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12',
  signature_base64: 'MEQCIQ==',
  certificate_pem: '-----BEGIN CERTIFICATE-----\nx\n-----END CERTIFICATE-----\n',
  signed_at: '2026-06-26T12:00:00Z',
}

describe('useDocumentSignatures', () => {
  beforeEach(() => jest.clearAllMocks())

  it('fetches a document signatures list', async () => {
    mockedApiClient.get.mockResolvedValueOnce({ success: true, data: [signatureFixture] })
    const { result } = renderHook(() => useDocumentSignatures(42), { wrapper })
    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.signatures).toHaveLength(1)
    expect(result.current.signatures[0].signer_name).toBe('Иванов И.И.')
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/42/signatures')
  })

  it('skips the fetch when disabled', async () => {
    renderHook(() => useDocumentSignatures(42, { enabled: false }), { wrapper })
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })

  it('skips the fetch when id is null', async () => {
    renderHook(() => useDocumentSignatures(null), { wrapper })
    expect(mockedApiClient.get).not.toHaveBeenCalled()
  })
})

describe('signDocument', () => {
  beforeEach(() => jest.clearAllMocks())

  it('POSTs to the sign endpoint with no body and returns the signature', async () => {
    mockedApiClient.post.mockResolvedValueOnce({ success: true, data: signatureFixture })
    const sig = await signDocument(42)
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/documents/42/sign')
    expect(sig.id).toBe(7)
  })
})

describe('verifySignature', () => {
  beforeEach(() => jest.clearAllMocks())

  it('GETs the verify endpoint and returns the verdict', async () => {
    const verdict = {
      signature_id: 7,
      valid: true,
      digest_match: true,
      crypto_valid: true,
      version_changed: false,
      status: 'valid',
    }
    mockedApiClient.get.mockResolvedValueOnce({ success: true, data: verdict })
    const v = await verifySignature(42, 7)
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/documents/42/signatures/7/verify')
    expect(v.valid).toBe(true)
    expect(v.status).toBe('valid')
  })
})

describe('pickSignatureErrorKey', () => {
  it('maps HTTP statuses to message keys', () => {
    expect(pickSignatureErrorKey({ response: { status: 403 } })).toBe('forbidden')
    expect(pickSignatureErrorKey({ response: { status: 404 } })).toBe('notFound')
    expect(pickSignatureErrorKey({ response: { status: 422 } })).toBe('invalid')
    expect(pickSignatureErrorKey(new Error('boom'))).toBe('generic')
    expect(pickSignatureErrorKey(null)).toBe('generic')
  })
})

describe('canSignDocuments', () => {
  it('allows non-student staff and denies students/anon', () => {
    expect(canSignDocuments(UserRole.METHODIST)).toBe(true)
    expect(canSignDocuments(UserRole.ACADEMIC_SECRETARY)).toBe(true)
    expect(canSignDocuments(UserRole.TEACHER)).toBe(true)
    expect(canSignDocuments(UserRole.SYSTEM_ADMIN)).toBe(true)
    expect(canSignDocuments(UserRole.STUDENT)).toBe(false)
    expect(canSignDocuments(undefined)).toBe(false)
  })
})
