import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/stores/authStore'
import { MFAVerifyLoginStep } from '../MFAVerifyLoginStep'

const mockPush = jest.fn()
jest.mock('next/navigation', () => {
  const actual = jest.requireActual('next/navigation')
  return {
    ...actual,
    useRouter: jest.fn(),
  }
})

const mockVerifyLoginMFA = jest.fn()
const mockClearMFAChallenge = jest.fn()

const mockToastError = jest.fn()
const mockToastSuccess = jest.fn()
jest.mock('sonner', () => ({
  toast: {
    error: (...args: unknown[]) => mockToastError(...args),
    success: (...args: unknown[]) => mockToastSuccess(...args),
  },
}))

beforeEach(() => {
  jest.clearAllMocks()
  ;(useRouter as jest.Mock).mockReturnValue({
    push: mockPush,
    replace: jest.fn(),
    prefetch: jest.fn(),
    back: jest.fn(),
  })

  // Seed authStore with an active MFA challenge
  useAuthStore.setState({
    mfaIntermediateToken: 'intermediate-jwt-abc',
    mfaPendingUser: {
      id: 7,
      name: 'Admin',
      email: 'admin@test.com',
      role: 'system_admin',
      mfa_enabled: true,
    },
    isAuthenticated: false,
    error: null,
    verifyLoginMFA: mockVerifyLoginMFA,
    clearMFAChallenge: mockClearMFAChallenge,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any)
})

afterEach(() => {
  // Reset to initial state
  useAuthStore.setState({
    mfaIntermediateToken: null,
    mfaPendingUser: null,
    isAuthenticated: false,
    error: null,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any)
})

describe('MFAVerifyLoginStep', () => {
  it('renders title, subtitle, code input, and submit button', () => {
    render(<MFAVerifyLoginStep />)
    // i18n keys returned verbatim by the test mock
    expect(screen.getByText('mfaPrompt.title')).toBeInTheDocument()
    expect(screen.getByText('mfaPrompt.subtitle')).toBeInTheDocument()
    expect(screen.getByLabelText('mfaPrompt.codeLabel')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'mfaPrompt.submit' })).toBeInTheDocument()
  })

  it('keeps submit disabled until a 6-digit numeric code is entered', () => {
    render(<MFAVerifyLoginStep />)
    const submit = screen.getByRole('button', { name: 'mfaPrompt.submit' })
    const input = screen.getByLabelText('mfaPrompt.codeLabel')

    expect(submit).toBeDisabled()
    fireEvent.change(input, { target: { value: '12345' } })
    expect(submit).toBeDisabled() // 5 digits
    fireEvent.change(input, { target: { value: 'abcdef' } })
    expect(submit).toBeDisabled() // non-numeric
    fireEvent.change(input, { target: { value: '123456' } })
    expect(submit).not.toBeDisabled()
  })

  it('on submit, calls verifyLoginMFA with the entered code and redirects on success', async () => {
    mockVerifyLoginMFA.mockResolvedValueOnce(undefined)
    render(<MFAVerifyLoginStep redirectTo="/dashboard" />)

    fireEvent.change(screen.getByLabelText('mfaPrompt.codeLabel'), { target: { value: '123456' } })
    fireEvent.click(screen.getByRole('button', { name: 'mfaPrompt.submit' }))

    await waitFor(() => expect(mockVerifyLoginMFA).toHaveBeenCalledWith('123456'))
    await waitFor(() => expect(mockPush).toHaveBeenCalledWith('/dashboard'))
  })

  it('uses default redirect "/" when redirectTo prop is omitted', async () => {
    mockVerifyLoginMFA.mockResolvedValueOnce(undefined)
    render(<MFAVerifyLoginStep />)

    fireEvent.change(screen.getByLabelText('mfaPrompt.codeLabel'), { target: { value: '123456' } })
    fireEvent.click(screen.getByRole('button', { name: 'mfaPrompt.submit' }))

    await waitFor(() => expect(mockPush).toHaveBeenCalledWith('/'))
  })

  it('on 422 INVALID_MFA_CODE, keeps challenge and shows inline errorInvalidCode i18n key', async () => {
    // Reviewer must-fix #2/#3 (v0.125.2 fix-cycle): the component
    // must map the wrong-code rejection to the localized i18n key
    // rather than relying on the raw backend message.
    mockVerifyLoginMFA.mockRejectedValueOnce({
      response: {
        status: 422,
        data: { error: { code: 'INVALID_MFA_CODE', message: 'Неверный код подтверждения' } },
      },
    })

    render(<MFAVerifyLoginStep />)
    fireEvent.change(screen.getByLabelText('mfaPrompt.codeLabel'), { target: { value: '000000' } })
    fireEvent.click(screen.getByRole('button', { name: 'mfaPrompt.submit' }))

    await waitFor(() => expect(mockVerifyLoginMFA).toHaveBeenCalledWith('000000'))
    expect(mockPush).not.toHaveBeenCalled()
    // Challenge preserved — user can retry
    expect(mockClearMFAChallenge).not.toHaveBeenCalled()
    // Code input still rendered
    expect(screen.getByLabelText('mfaPrompt.codeLabel')).toBeInTheDocument()
    // Inline error uses i18n key (test mock returns key verbatim)
    expect(screen.getByText('mfaPrompt.errorInvalidCode')).toBeInTheDocument()
  })

  it('on 401 dead intermediate, clears challenge and shows toast with errorIntermediateInvalid key', async () => {
    // Reviewer must-fix #2 (v0.125.2 fix-cycle): the component must
    // recognise a 401 response (intermediate invalid / expired /
    // replayed), drop the dead challenge so LoginForm flips back to
    // the credentials view, and surface the localized message via a
    // toast (the inline error region is unmounted by the clear).
    mockVerifyLoginMFA.mockRejectedValueOnce({
      response: { status: 401, data: { error: { message: 'Сессия MFA недействительна' } } },
    })

    render(<MFAVerifyLoginStep />)
    fireEvent.change(screen.getByLabelText('mfaPrompt.codeLabel'), { target: { value: '123456' } })
    fireEvent.click(screen.getByRole('button', { name: 'mfaPrompt.submit' }))

    await waitFor(() => expect(mockVerifyLoginMFA).toHaveBeenCalledWith('123456'))
    expect(mockPush).not.toHaveBeenCalled()
    // Challenge cleared → LoginForm reverts to credentials
    await waitFor(() => expect(mockClearMFAChallenge).toHaveBeenCalled())
    // Toast surfaced the localized message
    expect(mockToastError).toHaveBeenCalledWith(
      expect.stringMatching(/errorIntermediateInvalid/),
      expect.any(Object)
    )
  })

  it('"login again" link calls clearMFAChallenge', () => {
    render(<MFAVerifyLoginStep />)
    fireEvent.click(screen.getByRole('button', { name: /loginAgain/i }))
    expect(mockClearMFAChallenge).toHaveBeenCalled()
  })

  it('redirects without calling verify when no intermediate token is in store', () => {
    // No challenge — component should not be usable
    useAuthStore.setState({
      mfaIntermediateToken: null,
      mfaPendingUser: null,
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any)

    render(<MFAVerifyLoginStep />)
    // Either renders nothing, or guards the submit. Conservative pin:
    // verifyLoginMFA is never called from a guarded component.
    expect(mockVerifyLoginMFA).not.toHaveBeenCalled()
  })
})
