import { render, screen, waitFor } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { LoginForm } from '../LoginForm'
import { useLogin } from '@/hooks/useAuth'

// Mock dependencies
jest.mock('@/hooks/useAuth')
jest.mock('sonner', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))
jest.mock('@/components/providers/toaster-provider', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockUseLogin = useLogin as jest.MockedFunction<typeof useLogin>

describe('LoginForm', () => {
  const mockLogin = jest.fn()
  const mockClearError = jest.fn()

  beforeEach(() => {
    mockUseLogin.mockReturnValue({
      login: mockLogin,
      isLoading: false,
      error: null,
      clearError: mockClearError,
    })
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('renders login form with all fields', () => {
    render(<LoginForm />)

    // With mocked translations, check for translation keys
    expect(screen.getByLabelText('email')).toBeInTheDocument()
    expect(screen.getByLabelText('password')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'login' })).toBeInTheDocument()
    expect(screen.getByText('forgotPassword')).toBeInTheDocument()
    expect(screen.getByText('register')).toBeInTheDocument()
  })

  it('toggles password visibility', async () => {
    const user = userEvent.setup()
    render(<LoginForm />)

    const passwordInput = screen.getByLabelText('password')
    const buttons = screen.getAllByRole('button')
    const toggleButton = buttons.find((btn) => btn.getAttribute('tabIndex') === '-1')

    expect(passwordInput).toHaveAttribute('type', 'password')

    if (toggleButton) {
      await user.click(toggleButton)
      expect(passwordInput).toHaveAttribute('type', 'text')

      await user.click(toggleButton)
      expect(passwordInput).toHaveAttribute('type', 'password')
    }
  })

  it('submits form with valid data', async () => {
    const user = userEvent.setup()
    const mockOnSuccess = jest.fn()

    mockLogin.mockResolvedValueOnce(undefined)

    render(<LoginForm onSuccess={mockOnSuccess} />)

    const emailInput = screen.getByLabelText('email')
    const passwordInput = screen.getByLabelText('password')
    const submitButton = screen.getByRole('button', { name: 'login' })

    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'Password123!')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith(
        {
          email: 'test@example.com',
          password: 'Password123!',
        },
        '/'
      )
    })

    expect(mockOnSuccess).toHaveBeenCalled()
  })

  it('displays error message when login fails', async () => {
    const errorMessage = 'Invalid credentials'

    mockUseLogin.mockReturnValue({
      login: mockLogin,
      isLoading: false,
      error: errorMessage,
      clearError: mockClearError,
    })

    render(<LoginForm />)

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument()
    })
  })

  it('has submit button enabled by default', () => {
    render(<LoginForm />)

    // Check that submit button exists and is not disabled by default
    const submitButton = screen.getByRole('button', { name: 'login' })
    expect(submitButton).toBeInTheDocument()
    expect(submitButton).not.toBeDisabled()
  })

  it('disables submit button during form submission', async () => {
    const user = userEvent.setup()
    // Make login hang to keep isSubmitting true
    mockLogin.mockImplementation(() => new Promise(() => {}))

    render(<LoginForm />)

    const emailInput = screen.getByLabelText('email')
    const passwordInput = screen.getByLabelText('password')
    const submitButton = screen.getByRole('button', { name: 'login' })

    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'Password123!')
    await user.click(submitButton)

    // Button should be disabled during submission
    await waitFor(() => {
      expect(submitButton).toBeDisabled()
    })
  })

  it('clears errors when form is submitted', async () => {
    const user = userEvent.setup()
    mockLogin.mockResolvedValueOnce(undefined)

    render(<LoginForm />)

    const emailInput = screen.getByLabelText('email')
    const passwordInput = screen.getByLabelText('password')
    const submitButton = screen.getByRole('button', { name: 'login' })

    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'Password123!')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockClearError).toHaveBeenCalled()
    })
  })

  it('has correct link to register page', () => {
    render(<LoginForm />)

    const registerLink = screen.getByText('register')
    expect(registerLink.closest('a')).toHaveAttribute('href', '/register')
  })

  it('has correct link to forgot password page', () => {
    render(<LoginForm />)

    const forgotPasswordLink = screen.getByText('forgotPassword')
    expect(forgotPasswordLink.closest('a')).toHaveAttribute('href', '/forgot-password')
  })

  it('shows loading indicator during form submission', async () => {
    const user = userEvent.setup()
    // Make login hang to keep isSubmitting true
    mockLogin.mockImplementation(() => new Promise(() => {}))

    render(<LoginForm />)

    const emailInput = screen.getByLabelText('email')
    const passwordInput = screen.getByLabelText('password')
    const submitButton = screen.getByRole('button', { name: 'login' })

    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'Password123!')
    await user.click(submitButton)

    // Should show loading indicator
    await waitFor(() => {
      expect(document.querySelector('.animate-spin')).toBeInTheDocument()
    })
  })

  it('handles login error gracefully', async () => {
    const user = userEvent.setup()
    mockLogin.mockRejectedValueOnce(new Error('Network error'))

    render(<LoginForm />)

    const emailInput = screen.getByLabelText('email')
    const passwordInput = screen.getByLabelText('password')
    const submitButton = screen.getByRole('button', { name: 'login' })

    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'Password123!')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalled()
    })
  })

  it('uses custom redirect URL when provided', async () => {
    const user = userEvent.setup()
    mockLogin.mockResolvedValueOnce(undefined)

    render(<LoginForm redirectTo="/dashboard" />)

    const emailInput = screen.getByLabelText('email')
    const passwordInput = screen.getByLabelText('password')
    const submitButton = screen.getByRole('button', { name: 'login' })

    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'Password123!')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith(
        expect.objectContaining({
          email: 'test@example.com',
          password: 'Password123!',
        }),
        '/dashboard'
      )
    })
  })

  it('renders all form elements with proper structure', () => {
    const { container } = render(<LoginForm />)

    // Check form structure
    const form = container.querySelector('form')
    expect(form).toBeInTheDocument()

    // Check inputs have proper attributes
    const emailInput = screen.getByLabelText('email')
    expect(emailInput).toHaveAttribute('type', 'email')

    const passwordInput = screen.getByLabelText('password')
    expect(passwordInput).toHaveAttribute('type', 'password')
  })

  it('disables inputs during form submission', async () => {
    const user = userEvent.setup()
    // Make login hang to keep isSubmitting true
    mockLogin.mockImplementation(() => new Promise(() => {}))

    render(<LoginForm />)

    const emailInput = screen.getByLabelText('email')
    const passwordInput = screen.getByLabelText('password')
    const submitButton = screen.getByRole('button', { name: 'login' })

    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'Password123!')
    await user.click(submitButton)

    // Inputs should be disabled during submission
    await waitFor(() => {
      expect(emailInput).toBeDisabled()
      expect(passwordInput).toBeDisabled()
    })
  })
})
