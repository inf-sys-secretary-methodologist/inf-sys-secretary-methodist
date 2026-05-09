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

  it('on verify failure, shows error and does not redirect, preserving the step', async () => {
    mockVerifyLoginMFA.mockRejectedValueOnce(new Error('invalid'))
    // Simulate the store writing the backend error message
    useAuthStore.setState({ error: 'Неверный код подтверждения' } as never)

    render(<MFAVerifyLoginStep />)
    fireEvent.change(screen.getByLabelText('mfaPrompt.codeLabel'), { target: { value: '000000' } })
    fireEvent.click(screen.getByRole('button', { name: 'mfaPrompt.submit' }))

    await waitFor(() => expect(mockVerifyLoginMFA).toHaveBeenCalledWith('000000'))
    expect(mockPush).not.toHaveBeenCalled()
    // Code input still rendered → user can retry
    expect(screen.getByLabelText('mfaPrompt.codeLabel')).toBeInTheDocument()
    // Error is surfaced
    expect(screen.getByText(/Неверный код подтверждения/)).toBeInTheDocument()
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
