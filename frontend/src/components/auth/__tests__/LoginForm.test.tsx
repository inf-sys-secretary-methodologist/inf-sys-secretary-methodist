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

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/пароль/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /войти/i })).toBeInTheDocument()
    expect(screen.getByText(/забыли пароль/i)).toBeInTheDocument()
    expect(screen.getByText(/зарегистрироваться/i)).toBeInTheDocument()
  })

  it('validates email field', async () => {
    const user = userEvent.setup()
    render(<LoginForm />)

    const emailInput = screen.getByLabelText(/email/i)

    // Blur without entering email
    await user.click(emailInput)
    await user.tab()

    await waitFor(() => {
      expect(screen.getByText(/email обязателен/i)).toBeInTheDocument()
    })

    // Enter invalid email
    await user.clear(emailInput)
    await user.type(emailInput, 'invalid-email')
    await user.tab()

    await waitFor(() => {
      expect(screen.getByText(/неверный формат email/i)).toBeInTheDocument()
    })
  })

  it('validates password field', async () => {
    const user = userEvent.setup()
    render(<LoginForm />)

    const passwordInput = screen.getByLabelText(/пароль/i)

    // Blur without entering password
    await user.click(passwordInput)
    await user.tab()

    await waitFor(() => {
      expect(screen.getByText(/пароль обязателен/i)).toBeInTheDocument()
    })
  })

  it('toggles password visibility', async () => {
    const user = userEvent.setup()
    render(<LoginForm />)

    const passwordInput = screen.getByLabelText(/пароль/i)
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

    const emailInput = screen.getByLabelText(/email/i)
    const passwordInput = screen.getByLabelText(/пароль/i)
    const submitButton = screen.getByRole('button', { name: /войти/i })

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
    const errorMessage = 'Неверный email или пароль'

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

  it('disables form during submission', async () => {
    mockUseLogin.mockReturnValue({
      login: mockLogin,
      isLoading: true,
      error: null,
      clearError: mockClearError,
    })

    render(<LoginForm />)

    await waitFor(() => {
      const submitButton = screen.getByRole('button', { name: /вход\.\.\./i })
      expect(submitButton).toBeDisabled()
    })
  })

  it('clears errors when form is submitted', async () => {
    const user = userEvent.setup()
    mockLogin.mockResolvedValueOnce(undefined)

    render(<LoginForm />)

    const emailInput = screen.getByLabelText(/email/i)
    const passwordInput = screen.getByLabelText(/пароль/i)
    const submitButton = screen.getByRole('button', { name: /войти/i })

    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'Password123!')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockClearError).toHaveBeenCalled()
    })
  })

  it('has correct link to register page', () => {
    render(<LoginForm />)

    const registerLink = screen.getByText(/зарегистрироваться/i)
    expect(registerLink.closest('a')).toHaveAttribute('href', '/register')
  })

  it('has correct link to forgot password page', () => {
    render(<LoginForm />)

    const forgotPasswordLink = screen.getByText(/забыли пароль/i)
    expect(forgotPasswordLink.closest('a')).toHaveAttribute('href', '/forgot-password')
  })
})
