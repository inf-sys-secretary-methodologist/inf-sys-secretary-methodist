import { render, screen, waitFor } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { UserMenu } from '../UserMenu'
import { useAuth, useLogout } from '@/hooks/useAuth'
import { mockUser } from '@/test-utils'
import { UserRole } from '@/types/auth'

// Mock dependencies
jest.mock('@/hooks/useAuth')
jest.mock('sonner', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockUseAuth = useAuth as jest.MockedFunction<typeof useAuth>
const mockUseLogout = useLogout as jest.MockedFunction<typeof useLogout>

describe('UserMenu', () => {
  const mockLogout = jest.fn()
  const mockLogin = jest.fn()
  const mockRegister = jest.fn()
  const mockCheckAuth = jest.fn()
  const mockClearError = jest.fn()
  const mockLogoutAuth = jest.fn()

  beforeEach(() => {
    mockUseAuth.mockReturnValue({
      user: mockUser,
      isAuthenticated: true,
      isLoading: false,
      error: null,
      login: mockLogin,
      register: mockRegister,
      logout: mockLogoutAuth,
      checkAuth: mockCheckAuth,
      clearError: mockClearError,
    })

    mockUseLogout.mockReturnValue({
      logout: mockLogout,
      isLoading: false,
    })
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('does not render when user is not authenticated', () => {
    mockUseAuth.mockReturnValue({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
      login: mockLogin,
      register: mockRegister,
      logout: mockLogoutAuth,
      checkAuth: mockCheckAuth,
      clearError: mockClearError,
    })

    render(<UserMenu />)
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('renders user avatar with initials', () => {
    render(<UserMenu />)

    const avatar = screen.getByText('TU') // "Test User" -> "TU"
    expect(avatar).toBeInTheDocument()
  })

  it('displays user name and role in trigger', () => {
    render(<UserMenu />)

    expect(screen.getByText(mockUser.name)).toBeInTheDocument()
    expect(screen.getByText('Студент')).toBeInTheDocument() // STUDENT role
  })

  it('opens dropdown menu on click', async () => {
    const user = userEvent.setup()
    render(<UserMenu />)

    const trigger = screen.getByRole('button')
    await user.click(trigger)

    await waitFor(() => {
      expect(screen.getByText(/профиль/i)).toBeInTheDocument()
      expect(screen.getByText(/настройки/i)).toBeInTheDocument()
      expect(screen.getByText(/выйти/i)).toBeInTheDocument()
    })
  })

  it('displays full user info in dropdown', async () => {
    const user = userEvent.setup()
    render(<UserMenu />)

    const trigger = screen.getByRole('button')
    await user.click(trigger)

    await waitFor(() => {
      const allNameElements = screen.getAllByText(mockUser.name)
      expect(allNameElements.length).toBeGreaterThan(0)
      expect(screen.getByText(mockUser.email)).toBeInTheDocument()
    })
  })

  it('has profile link with correct href', async () => {
    const user = userEvent.setup()
    render(<UserMenu />)

    const trigger = screen.getByRole('button')
    await user.click(trigger)

    await waitFor(() => {
      const profileLink = screen.getByText(/профиль/i).closest('a')
      expect(profileLink).toHaveAttribute('href', '/profile')
    })
  })

  it('has settings link with correct href', async () => {
    const user = userEvent.setup()
    render(<UserMenu />)

    const trigger = screen.getByRole('button')
    await user.click(trigger)

    await waitFor(() => {
      const settingsLink = screen.getByText(/настройки/i).closest('a')
      expect(settingsLink).toHaveAttribute('href', '/settings')
    })
  })

  it('calls logout function when logout is clicked', async () => {
    const user = userEvent.setup()
    mockLogout.mockResolvedValueOnce(undefined)

    render(<UserMenu />)

    const trigger = screen.getByRole('button')
    await user.click(trigger)

    await waitFor(() => {
      expect(screen.getByText(/выйти/i)).toBeInTheDocument()
    })

    const logoutButton = screen.getByText(/выйти/i)
    await user.click(logoutButton)

    await waitFor(() => {
      expect(mockLogout).toHaveBeenCalledWith('/login')
    })
  })

  it('disables logout button when loading', async () => {
    const user = userEvent.setup()

    mockUseLogout.mockReturnValue({
      logout: mockLogout,
      isLoading: true,
    })

    render(<UserMenu />)

    const trigger = screen.getByRole('button')
    await user.click(trigger)

    await waitFor(() => {
      const logoutButton = screen.getByText(/выход\.\.\./i)
      expect(logoutButton.closest('div')).toHaveAttribute('data-disabled')
    })
  })

  it('displays different role names correctly', () => {
    const roles = [
      { role: UserRole.SYSTEM_ADMIN, display: 'Администратор' },
      { role: UserRole.METHODIST, display: 'Методист' },
      { role: UserRole.ACADEMIC_SECRETARY, display: 'Секретарь' },
      { role: UserRole.TEACHER, display: 'Преподаватель' },
      { role: UserRole.STUDENT, display: 'Студент' },
    ]

    roles.forEach(({ role, display }) => {
      mockUseAuth.mockReturnValue({
        user: { ...mockUser, role },
        isAuthenticated: true,
        isLoading: false,
        error: null,
        login: mockLogin,
        register: mockRegister,
        logout: mockLogoutAuth,
        checkAuth: mockCheckAuth,
        clearError: mockClearError,
      })

      const { unmount } = render(<UserMenu />)
      expect(screen.getByText(display)).toBeInTheDocument()
      unmount()
    })
  })

  it('generates correct initials for single word name', () => {
    mockUseAuth.mockReturnValue({
      user: { ...mockUser, name: 'Admin' },
      isAuthenticated: true,
      isLoading: false,
      error: null,
      login: mockLogin,
      register: mockRegister,
      logout: mockLogoutAuth,
      checkAuth: mockCheckAuth,
      clearError: mockClearError,
    })

    render(<UserMenu />)
    const avatar = screen.getByText('AD') // "Admin" -> "AD"
    expect(avatar).toBeInTheDocument()
  })

  it('generates correct initials for multi-word name', () => {
    mockUseAuth.mockReturnValue({
      user: { ...mockUser, name: 'John Michael Doe' },
      isAuthenticated: true,
      isLoading: false,
      error: null,
      login: mockLogin,
      register: mockRegister,
      logout: mockLogoutAuth,
      checkAuth: mockCheckAuth,
      clearError: mockClearError,
    })

    render(<UserMenu />)
    const avatar = screen.getByText('JM') // "John Michael Doe" -> "JM"
    expect(avatar).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(<UserMenu className="custom-class" />)
    const button = container.querySelector('button')
    expect(button).toHaveClass('custom-class')
  })

  it('has chevron icon in trigger', () => {
    const { container } = render(<UserMenu />)
    const chevron = container.querySelector('svg')
    expect(chevron).toBeInTheDocument()
  })
})
