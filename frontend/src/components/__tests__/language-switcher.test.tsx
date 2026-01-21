import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { LanguageSwitcher } from '../language-switcher'

// Mock next-intl
jest.mock('next-intl', () => ({
  useLocale: jest.fn(() => 'ru'),
}))

// Mock next/navigation
const mockRefresh = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    refresh: mockRefresh,
  }),
}))

describe('LanguageSwitcher', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // Mock document.cookie
    Object.defineProperty(document, 'cookie', {
      writable: true,
      value: '',
    })
  })

  it('renders language switcher button', () => {
    render(<LanguageSwitcher />)
    expect(screen.getByRole('button', { name: /change language/i })).toBeInTheDocument()
  })

  it('opens dropdown menu on click', async () => {
    const user = userEvent.setup()
    render(<LanguageSwitcher />)

    const button = screen.getByRole('button', { name: /change language/i })
    await user.click(button)

    // Should show language options
    expect(screen.getByText('Русский')).toBeInTheDocument()
    expect(screen.getByText('English')).toBeInTheDocument()
    expect(screen.getByText('Français')).toBeInTheDocument()
  })

  it('shows flag emojis for languages', async () => {
    const user = userEvent.setup()
    render(<LanguageSwitcher />)

    const button = screen.getByRole('button', { name: /change language/i })
    await user.click(button)

    // Should show flag emojis
    expect(screen.getByText('🇷🇺')).toBeInTheDocument()
    expect(screen.getByText('🇬🇧')).toBeInTheDocument()
    expect(screen.getByText('🇫🇷')).toBeInTheDocument()
  })

  it('calls router.refresh when language is changed', async () => {
    const user = userEvent.setup()
    render(<LanguageSwitcher />)

    const button = screen.getByRole('button', { name: /change language/i })
    await user.click(button)

    const englishOption = screen.getByText('English')
    await user.click(englishOption)

    expect(mockRefresh).toHaveBeenCalled()
  })

  it('sets locale cookie when language is changed', async () => {
    const user = userEvent.setup()
    render(<LanguageSwitcher />)

    const button = screen.getByRole('button', { name: /change language/i })
    await user.click(button)

    const englishOption = screen.getByText('English')
    await user.click(englishOption)

    expect(document.cookie).toContain('NEXT_LOCALE=en')
  })

  it('highlights current locale', async () => {
    const user = userEvent.setup()
    render(<LanguageSwitcher />)

    const button = screen.getByRole('button', { name: /change language/i })
    await user.click(button)

    // Russian should have accent class (current locale)
    const russianItem = screen.getByText('Русский').closest('[role="menuitem"]')
    expect(russianItem).toHaveClass('bg-accent')
  })
})
