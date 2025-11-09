import { render, screen, waitFor } from '@/test-utils'
import { ThemeToggleButton } from '../theme-toggle-button'
import userEvent from '@testing-library/user-event'

describe('ThemeToggleButton', () => {
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
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('aria-label', 'Switch to dark theme')
    })
  })

  it('toggles theme when clicked', async () => {
    const user = userEvent.setup()
    render(<ThemeToggleButton />)

    await waitFor(() => {
      expect(screen.getByRole('button')).toBeInTheDocument()
    })

    const button = screen.getByRole('button')

    // Click to switch to dark theme
    await user.click(button)

    await waitFor(() => {
      expect(button).toHaveAttribute('aria-label', 'Switch to light theme')
    })
  })

  it('has proper accessibility attributes', async () => {
    render(<ThemeToggleButton />)

    await waitFor(() => {
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('type', 'button')
      expect(button).toHaveAttribute('aria-label')
    })
  })

  it('applies correct CSS classes', async () => {
    render(<ThemeToggleButton />)

    await waitFor(() => {
      const button = screen.getByRole('button')
      expect(button).toHaveClass('inline-flex')
      expect(button).toHaveClass('rounded-2xl')
    })
  })
})
