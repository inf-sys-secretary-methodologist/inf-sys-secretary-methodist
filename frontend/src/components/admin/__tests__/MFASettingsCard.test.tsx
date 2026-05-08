import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { MFASettingsCard } from '../MFASettingsCard'

const mockBegin = jest.fn()
const mockConfirm = jest.fn()
const mockDisable = jest.fn()
jest.mock('@/hooks/useMFA', () => ({
  useMFA: () => ({
    beginEnrollment: (...args: unknown[]) => mockBegin(...args),
    confirmEnrollment: (...args: unknown[]) => mockConfirm(...args),
    disable: (...args: unknown[]) => mockDisable(...args),
  }),
}))

const mockToastSuccess = jest.fn()
const mockToastError = jest.fn()
jest.mock('sonner', () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}))

beforeEach(() => {
  jest.clearAllMocks()
})

describe('MFASettingsCard', () => {
  it('renders Enable button when MFA is disabled', () => {
    render(<MFASettingsCard mfaEnabled={false} />)
    expect(screen.getByRole('button', { name: 'mfa.enable' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'mfa.disable' })).not.toBeInTheDocument()
  })

  it('renders Disable button when MFA is enabled', () => {
    render(<MFASettingsCard mfaEnabled={true} />)
    expect(screen.getByRole('button', { name: 'mfa.disable' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'mfa.enable' })).not.toBeInTheDocument()
  })

  it('clicking Enable calls beginEnrollment and shows secret + verify input', async () => {
    mockBegin.mockResolvedValueOnce({
      otpauth_uri: 'otpauth://totp/test:admin?secret=ABC',
      secret: 'JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP',
    })
    render(<MFASettingsCard mfaEnabled={false} />)
    fireEvent.click(screen.getByRole('button', { name: 'mfa.enable' }))
    await waitFor(() => expect(mockBegin).toHaveBeenCalled())
    expect(screen.getByText(/JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP/)).toBeInTheDocument()
    expect(screen.getByLabelText('mfa.codeLabel')).toBeInTheDocument()
  })

  it('after Begin, submitting 6-digit code calls confirmEnrollment', async () => {
    mockBegin.mockResolvedValueOnce({
      otpauth_uri: 'otpauth://totp/test:admin?secret=ABC',
      secret: 'JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP',
    })
    mockConfirm.mockResolvedValueOnce(undefined)
    const onChange = jest.fn()
    render(<MFASettingsCard mfaEnabled={false} onChange={onChange} />)
    fireEvent.click(screen.getByRole('button', { name: 'mfa.enable' }))
    await waitFor(() => expect(mockBegin).toHaveBeenCalled())
    fireEvent.change(screen.getByLabelText('mfa.codeLabel'), { target: { value: '123456' } })
    fireEvent.click(screen.getByRole('button', { name: 'mfa.confirm' }))
    await waitFor(() => expect(mockConfirm).toHaveBeenCalledWith('123456'))
    await waitFor(() => expect(onChange).toHaveBeenCalledWith(true))
  })

  it('clicking Disable shows code input then calls disable on submit', async () => {
    mockDisable.mockResolvedValueOnce(undefined)
    const onChange = jest.fn()
    render(<MFASettingsCard mfaEnabled={true} onChange={onChange} />)
    fireEvent.click(screen.getByRole('button', { name: 'mfa.disable' }))
    expect(screen.getByLabelText('mfa.codeLabel')).toBeInTheDocument()
    fireEvent.change(screen.getByLabelText('mfa.codeLabel'), { target: { value: '654321' } })
    fireEvent.click(screen.getByRole('button', { name: 'mfa.confirmDisable' }))
    await waitFor(() => expect(mockDisable).toHaveBeenCalledWith('654321'))
    await waitFor(() => expect(onChange).toHaveBeenCalledWith(false))
  })

  it('keeps Confirm button disabled until 6-digit numeric code is entered', async () => {
    mockBegin.mockResolvedValueOnce({
      otpauth_uri: 'otpauth://totp/test:admin?secret=ABC',
      secret: 'JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP',
    })
    render(<MFASettingsCard mfaEnabled={false} />)
    fireEvent.click(screen.getByRole('button', { name: 'mfa.enable' }))
    await waitFor(() => expect(mockBegin).toHaveBeenCalled())
    const confirmBtn = screen.getByRole('button', { name: 'mfa.confirm' })
    expect(confirmBtn).toBeDisabled()
    fireEvent.change(screen.getByLabelText('mfa.codeLabel'), { target: { value: '12345' } })
    expect(confirmBtn).toBeDisabled() // 5 digits
    fireEvent.change(screen.getByLabelText('mfa.codeLabel'), { target: { value: 'abcdef' } })
    expect(confirmBtn).toBeDisabled() // non-numeric
    fireEvent.change(screen.getByLabelText('mfa.codeLabel'), { target: { value: '123456' } })
    expect(confirmBtn).not.toBeDisabled()
  })
})
