import { render, screen, waitFor } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { RegisterForm } from '../RegisterForm'
import { useRegister } from '@/hooks/useAuth'
import { UserRole } from '@/types/auth'

// Mock dependencies
jest.mock('@/hooks/useAuth')
jest.mock('sonner', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockUseRegister = useRegister as jest.MockedFunction<typeof useRegister>

describe('RegisterForm', () => {
  const mockRegister = jest.fn()
  const mockClearError = jest.fn()

  beforeEach(() => {
    mockUseRegister.mockReturnValue({
      register: mockRegister,
      isLoading: false,
      error: null,
      clearError: mockClearError,
    })
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('renders register form with all fields', () => {
    render(<RegisterForm />)

    expect(screen.getByLabelText(/имя/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    const passwordInputs = screen.getAllByPlaceholderText(/пароль/i)
    expect(passwordInputs.length).toBe(2) // password and confirmPassword
    expect(screen.getByRole('combobox')).toBeInTheDocument() // role select
    expect(screen.getByRole('button', { name: /зарегистрироваться/i })).toBeInTheDocument()
    expect(screen.getByText(/войти/i)).toBeInTheDocument()
  })

  it('validates name field', async () => {
    const user = userEvent.setup()
    render(<RegisterForm />)

    const nameInput = screen.getByLabelText(/имя/i)

    // Blur without entering name
    await user.click(nameInput)
    await user.tab()

    await waitFor(() => {
      expect(screen.getByText(/имя должно содержать минимум 2 символа/i)).toBeInTheDocument()
    })
  })

  it('validates email field', async () => {
    const user = userEvent.setup()
    render(<RegisterForm />)

    const emailInput = screen.getByLabelText(/email/i)

    // Enter invalid email
    await user.type(emailInput, 'invalid-email')
    await user.tab()

    await waitFor(() => {
      expect(screen.getByText(/неверный формат email/i)).toBeInTheDocument()
    })
  })

  it('validates password requirements', async () => {
    const user = userEvent.setup()
    render(<RegisterForm />)

    const passwordInput = screen.getAllByLabelText(/пароль/i)[0]

    // Weak password
    await user.type(passwordInput, 'weak')
    await user.tab()

    await waitFor(() => {
      expect(screen.getByText(/пароль должен:/i)).toBeInTheDocument()
    })
  })

  it('displays password strength indicator', async () => {
    const user = userEvent.setup()
    render(<RegisterForm />)

    const passwordInput = screen.getAllByLabelText(/пароль/i)[0]

    // Type a strong password
    await user.type(passwordInput, 'StrongPass123!')

    await waitFor(() => {
      expect(screen.getByText(/сложность пароля:/i)).toBeInTheDocument()
      expect(screen.getByText(/сильный/i)).toBeInTheDocument()
    })
  })

  it('validates password confirmation', async () => {
    const user = userEvent.setup()
    render(<RegisterForm />)

    const passwordInput = screen.getAllByLabelText(/пароль/i)[0]
    const confirmPasswordInput = screen.getByLabelText(/подтвердите пароль/i)

    await user.type(passwordInput, 'StrongPass123!')
    await user.type(confirmPasswordInput, 'DifferentPass123!')
    await user.tab()

    await waitFor(() => {
      expect(screen.getByText(/пароли не совпадают/i)).toBeInTheDocument()
    })
  })

  it('toggles password visibility for both fields', async () => {
    const user = userEvent.setup()
    render(<RegisterForm />)

    const passwordInputs = screen.getAllByPlaceholderText(/пароль/i)
    const passwordInput = passwordInputs[0]
    const confirmPasswordInput = screen.getByPlaceholderText(/подтвердите пароль/i)
    const buttons = screen.getAllByRole('button')
    const toggleButtons = buttons.filter((btn) => btn.getAttribute('tabIndex') === '-1')

    // Initially hidden
    expect(passwordInput).toHaveAttribute('type', 'password')
    expect(confirmPasswordInput).toHaveAttribute('type', 'password')

    if (toggleButtons.length >= 2) {
      // Toggle password field
      await user.click(toggleButtons[0])
      expect(passwordInput).toHaveAttribute('type', 'text')

      // Toggle confirm password field
      await user.click(toggleButtons[1])
      expect(confirmPasswordInput).toHaveAttribute('type', 'text')
    }
  })

  it('submits form with valid data', async () => {
    const user = userEvent.setup()
    const mockOnSuccess = jest.fn()

    mockRegister.mockResolvedValueOnce(undefined)

    render(<RegisterForm onSuccess={mockOnSuccess} />)

    const nameInput = screen.getByLabelText(/имя/i)
    const emailInput = screen.getByLabelText(/email/i)
    const passwordInputs = screen.getAllByPlaceholderText(/пароль/i)
    const passwordInput = passwordInputs[0]
    const confirmPasswordInput = screen.getByPlaceholderText(/подтвердите пароль/i)
    const roleSelect = screen.getByRole('combobox')
    const submitButton = screen.getByRole('button', { name: /зарегистрироваться/i })

    await user.type(nameInput, 'Test User')
    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'StrongPass123!')
    await user.type(confirmPasswordInput, 'StrongPass123!')
    await user.selectOptions(roleSelect, UserRole.TEACHER)
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith(
        {
          name: 'Test User',
          email: 'test@example.com',
          password: 'StrongPass123!',
          role: UserRole.TEACHER,
        },
        '/login'
      )
    })

    expect(mockOnSuccess).toHaveBeenCalled()
  })

  it('displays error message when registration fails', async () => {
    const errorMessage = 'Email уже используется'

    mockUseRegister.mockReturnValue({
      register: mockRegister,
      isLoading: false,
      error: errorMessage,
      clearError: mockClearError,
    })

    render(<RegisterForm />)

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument()
    })
  })

  it('disables form during submission', async () => {
    mockUseRegister.mockReturnValue({
      register: mockRegister,
      isLoading: true,
      error: null,
      clearError: mockClearError,
    })

    render(<RegisterForm />)

    await waitFor(() => {
      const submitButton = screen.getByRole('button', { name: /регистрация\.\.\./i })
      expect(submitButton).toBeDisabled()
    })
  })

  it('has all role options', () => {
    render(<RegisterForm />)

    const roleSelect = screen.getByRole('combobox') as HTMLSelectElement

    expect(roleSelect.options).toHaveLength(5)
    expect(roleSelect.options[0]).toHaveValue(UserRole.STUDENT)
    expect(roleSelect.options[1]).toHaveValue(UserRole.TEACHER)
    expect(roleSelect.options[2]).toHaveValue(UserRole.SECRETARY)
    expect(roleSelect.options[3]).toHaveValue(UserRole.METHODIST)
    expect(roleSelect.options[4]).toHaveValue(UserRole.ADMIN)
  })

  it('defaults to STUDENT role', () => {
    render(<RegisterForm />)

    const roleSelect = screen.getByRole('combobox') as HTMLSelectElement
    expect(roleSelect.value).toBe(UserRole.STUDENT)
  })

  it('has correct link to login page', () => {
    render(<RegisterForm />)

    const loginLink = screen.getByText(/войти/i)
    expect(loginLink.closest('a')).toHaveAttribute('href', '/login')
  })

  it('clears errors when form is submitted', async () => {
    const user = userEvent.setup()
    mockRegister.mockResolvedValueOnce(undefined)

    render(<RegisterForm />)

    const nameInput = screen.getByLabelText(/имя/i)
    const emailInput = screen.getByLabelText(/email/i)
    const passwordInputs = screen.getAllByPlaceholderText(/пароль/i)
    const passwordInput = passwordInputs[0]
    const confirmPasswordInput = screen.getByPlaceholderText(/подтвердите пароль/i)
    const submitButton = screen.getByRole('button', { name: /зарегистрироваться/i })

    await user.type(nameInput, 'Test User')
    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'StrongPass123!')
    await user.type(confirmPasswordInput, 'StrongPass123!')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockClearError).toHaveBeenCalled()
    })
  })
})
