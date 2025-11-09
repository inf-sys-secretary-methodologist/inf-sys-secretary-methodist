import { render, screen, waitFor } from '@/test-utils'
import { ThemeToggleButton } from '../theme-toggle-button'
import userEvent from '@testing-library/user-event'

// Mock the useTheme hook
jest.mock('@/hooks/use-theme', () => ({
  useTheme: jest.fn(),
}))

const mockUseTheme = require('@/hooks/use-theme').useTheme

describe('ThemeToggleButton', () => {
  beforeEach(() => {
    // Default mock implementation - light theme
    mockUseTheme.mockReturnValue({
      resolvedTheme: 'light',
      toggleTheme: jest.fn(),
      theme: 'light',
      setTheme: jest.fn(),
      isDark: false,
      isLight: true,
    })
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('renders the theme toggle button', async () => {
    render(<ThemeToggleButton />)

    // Wait for component to mount (handles hydration)
    await waitFor(() => {
      expect(screen.getByRole('button')).toBeInTheDocument()
    })
  })

  it('shows moon icon in light theme by default', async () => {
    render(<ThemeToggleButton />)

    await waitFor(() => {
      const toggle = screen.getByRole('button')
      expect(toggle).toHaveAttribute('aria-label', 'Switch to dark theme')
    })
  })

  it('toggles theme when clicked', async () => {
    const toggleThemeMock = jest.fn()
    mockUseTheme.mockReturnValue({
      resolvedTheme: 'light',
      toggleTheme: toggleThemeMock,
      theme: 'light',
      setTheme: jest.fn(),
      isDark: false,
      isLight: true,
    })

    const user = userEvent.setup()
    render(<ThemeToggleButton />)

    await waitFor(() => {
      expect(screen.getByRole('button')).toBeInTheDocument()
    })

    const toggle = screen.getByRole('button')

    // Click to switch theme
    await user.click(toggle)

    await waitFor(() => {
      expect(toggleThemeMock).toHaveBeenCalledTimes(1)
    })
  })

  it('has proper accessibility attributes', async () => {
    render(<ThemeToggleButton />)

    await waitFor(() => {
      const toggle = screen.getByRole('button')
      expect(toggle).toHaveAttribute('aria-label')
      expect(toggle).toHaveAttribute('tabIndex', '0')
    })
  })

  it('applies correct CSS classes', async () => {
    render(<ThemeToggleButton />)

    await waitFor(() => {
      const toggle = screen.getByRole('button')
      expect(toggle).toHaveClass('flex')
      expect(toggle).toHaveClass('rounded-full')
      expect(toggle).toHaveClass('cursor-pointer')
    })
  })

  it('supports keyboard navigation', async () => {
    const toggleThemeMock = jest.fn()
    mockUseTheme.mockReturnValue({
      resolvedTheme: 'light',
      toggleTheme: toggleThemeMock,
      theme: 'light',
      setTheme: jest.fn(),
      isDark: false,
      isLight: true,
    })

    const user = userEvent.setup()
    render(<ThemeToggleButton />)

    await waitFor(() => {
      expect(screen.getByRole('button')).toBeInTheDocument()
    })

    const toggle = screen.getByRole('button')
    toggle.focus()

    // Press Enter to toggle
    await user.keyboard('{Enter}')

    await waitFor(() => {
      expect(toggleThemeMock).toHaveBeenCalled()
    })

    // Press Space to toggle
    await user.keyboard(' ')

    await waitFor(() => {
      expect(toggleThemeMock).toHaveBeenCalledTimes(2)
    })
  })
})
