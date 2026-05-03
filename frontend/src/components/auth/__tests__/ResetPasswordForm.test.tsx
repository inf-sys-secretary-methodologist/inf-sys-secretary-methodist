import { render, screen, waitFor } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { ResetPasswordForm } from '../ResetPasswordForm'
import { authApi } from '@/lib/api/auth'

jest.mock('@/lib/api/auth', () => ({
  authApi: {
    verifyPasswordResetToken: jest.fn(),
    confirmPasswordReset: jest.fn(),
  },
}))

const mockedApi = authApi as jest.Mocked<typeof authApi>

// Override the global next/navigation mock per-test so we can vary
// the ?token= query string and observe how the form branches.
jest.mock('next/navigation', () => {
  const actual = jest.requireActual('next/navigation')
  return {
    ...actual,
    useRouter: () => ({ push: jest.fn(), replace: jest.fn(), prefetch: jest.fn(), back: jest.fn() }),
    useSearchParams: jest.fn(),
    usePathname: () => '/reset-password',
  }
})

import { useSearchParams } from 'next/navigation'
const mockedUseSearchParams = useSearchParams as jest.Mock

function withToken(token: string | null) {
  const params = new URLSearchParams()
  if (token !== null) params.set('token', token)
  mockedUseSearchParams.mockReturnValue(params)
}

describe('ResetPasswordForm', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  // Pins the verify state machine: missing token -> dedicated panel,
  // valid token -> the new-password form, invalid token -> expired
  // panel. Without these, the verifyState branches can silently
  // regress and the user lands on a form that will always 410.
  it('renders "token missing" panel when ?token is absent', () => {
    withToken(null)
    render(<ResetPasswordForm />)
    expect(screen.getByText('tokenMissing')).toBeInTheDocument()
    expect(mockedApi.verifyPasswordResetToken).not.toHaveBeenCalled()
  })

  it('renders the password form once verifyPasswordResetToken resolves', async () => {
    withToken('good-tok')
    mockedApi.verifyPasswordResetToken.mockResolvedValueOnce(undefined)

    render(<ResetPasswordForm />)

    // Verifying state shows first
    expect(screen.getByText('verifying')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByLabelText('newPasswordLabel')).toBeInTheDocument()
    })
    expect(screen.getByLabelText('confirmPasswordLabel')).toBeInTheDocument()
    expect(mockedApi.verifyPasswordResetToken).toHaveBeenCalledWith('good-tok')
  })

  it('renders "link expired" panel when verify rejects', async () => {
    withToken('bad-tok')
    mockedApi.verifyPasswordResetToken.mockRejectedValueOnce(
      Object.assign(new Error('gone'), { response: { status: 410 } }),
    )

    render(<ResetPasswordForm />)

    await waitFor(() => {
      expect(screen.getByText('linkExpiredTitle')).toBeInTheDocument()
    })
    expect(screen.getByText('requestNewLink')).toBeInTheDocument()
  })

  // Pins the confirm-time error mapping. 410 on confirm means another
  // tab spent the token (or it expired between verify and submit) —
  // the form must drop back to the expired panel rather than show a
  // generic error, otherwise the user sits on a dead form forever.
  it('flips to "link expired" panel when confirm responds 410', async () => {
    withToken('good-tok')
    mockedApi.verifyPasswordResetToken.mockResolvedValueOnce(undefined)
    mockedApi.confirmPasswordReset.mockRejectedValueOnce(
      Object.assign(new Error('gone'), { response: { status: 410 } }),
    )

    const user = userEvent.setup()
    render(<ResetPasswordForm />)

    await waitFor(() => {
      expect(screen.getByLabelText('newPasswordLabel')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('newPasswordLabel'), 'NewStrongPass1!')
    await user.type(screen.getByLabelText('confirmPasswordLabel'), 'NewStrongPass1!')
    await user.click(screen.getByRole('button', { name: 'submit' }))

    await waitFor(() => {
      expect(screen.getByText('linkExpiredTitle')).toBeInTheDocument()
    })
  })
})
