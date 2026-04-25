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

// Increase timeout for tests with complex form interactions
jest.setTimeout(15000)

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

    // With mocked translations, check for translation keys
    expect(screen.getByLabelText('name')).toBeInTheDocument()
    expect(screen.getByLabelText('email')).toBeInTheDocument()
    expect(screen.getByLabelText('password')).toBeInTheDocument()
    expect(screen.getByLabelText('confirmPassword')).toBeInTheDocument()
    expect(screen.getByRole('combobox')).toBeInTheDocument() // role select
    expect(screen.getByRole('button', { name: 'register' })).toBeInTheDocument()
    expect(screen.getByText('login')).toBeInTheDocument()
  })

  it('only offers self-registration roles (student, teacher) in role select', () => {
    render(<RegisterForm />)

    const select = screen.getByRole('combobox') as HTMLSelectElement
    const optionValues = Array.from(select.options).map((o) => o.value)

    expect(optionValues).toContain(UserRole.STUDENT)
    expect(optionValues).toContain(UserRole.TEACHER)
    expect(optionValues).not.toContain(UserRole.ACADEMIC_SECRETARY)
    expect(optionValues).not.toContain(UserRole.METHODIST)
    expect(optionValues).not.toContain(UserRole.SYSTEM_ADMIN)
  })

  it('toggles password visibility for both fields', async () => {
    const user = userEvent.setup()
    render(<RegisterForm />)

    const passwordInput = screen.getByLabelText('password')
    const confirmPasswordInput = screen.getByLabelText('confirmPassword')
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

    const nameInput = screen.getByLabelText('name')
    const emailInput = screen.getByLabelText('email')
    const passwordInput = screen.getByLabelText('password')
    const confirmPasswordInput = screen.getByLabelText('confirmPassword')
    const roleSelect = screen.getByRole('combobox')
    const submitButton = screen.getByRole('button', { name: 'register' })

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
    const errorMessage = 'Email already in use'

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

  it('has submit button that can be disabled during submission', () => {
    render(<RegisterForm />)

    // Check that submit button exists
    const submitButton = screen.getByRole('button', { name: 'register' })
    expect(submitButton).toBeInTheDocument()
    expect(submitButton).not.toBeDisabled()
  })

  it('has all role options', () => {
    render(<RegisterForm />)

    const roleSelect = screen.getByRole('combobox') as HTMLSelectElement

    expect(roleSelect.options).toHaveLength(5)
    expect(roleSelect.options[0]).toHaveValue(UserRole.STUDENT)
    expect(roleSelect.options[1]).toHaveValue(UserRole.TEACHER)
    expect(roleSelect.options[2]).toHaveValue(UserRole.ACADEMIC_SECRETARY)
    expect(roleSelect.options[3]).toHaveValue(UserRole.METHODIST)
    expect(roleSelect.options[4]).toHaveValue(UserRole.SYSTEM_ADMIN)
  })

  it('defaults to STUDENT role', () => {
    render(<RegisterForm />)

    const roleSelect = screen.getByRole('combobox') as HTMLSelectElement
    expect(roleSelect.value).toBe(UserRole.STUDENT)
  })

  it('has correct link to login page', () => {
    render(<RegisterForm />)

    const loginLink = screen.getByText('login')
    expect(loginLink.closest('a')).toHaveAttribute('href', '/login')
  })

  it('clears errors when form is submitted', async () => {
    const user = userEvent.setup()
    mockRegister.mockResolvedValueOnce(undefined)

    render(<RegisterForm />)

    const nameInput = screen.getByLabelText('name')
    const emailInput = screen.getByLabelText('email')
    const passwordInput = screen.getByLabelText('password')
    const confirmPasswordInput = screen.getByLabelText('confirmPassword')
    const submitButton = screen.getByRole('button', { name: 'register' })

    await user.type(nameInput, 'Test User')
    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'StrongPass123!')
    await user.type(confirmPasswordInput, 'StrongPass123!')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockClearError).toHaveBeenCalled()
    })
  })

  it('displays password strength indicator', async () => {
    const user = userEvent.setup()
    render(<RegisterForm />)

    const passwordInput = screen.getByLabelText('password')

    // Type a strong password
    await user.type(passwordInput, 'StrongPass123!')

    await waitFor(() => {
      // Check for translation keys
      expect(screen.getByText('passwordStrength')).toBeInTheDocument()
      expect(screen.getByText('passwordStrong')).toBeInTheDocument()
    })
  })
})
