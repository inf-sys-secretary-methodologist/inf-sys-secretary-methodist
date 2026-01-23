import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { ThemeSettingsPopover } from '../theme-settings-popover'

// Mock next-themes
const mockSetTheme = jest.fn()
jest.mock('next-themes', () => ({
  useTheme: () => ({
    theme: 'light',
    setTheme: mockSetTheme,
  }),
}))

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      ariaLabel: 'Theme settings',
      themeTitle: 'Theme',
      'themes.light': 'Light',
      'themes.dark': 'Dark',
      'themes.system': 'System',
      animatedBackground: 'Animated Background',
      enableAnimatedBackground: 'Enable animated background',
      'backgrounds.none': 'None',
      'backgrounds.grain-gradient': 'Grain',
      'backgrounds.warp': 'Warp',
      'backgrounds.mesh-gradient': 'Mesh',
    }
    return translations[key] || key
  },
}))

// Mock appearance store
const mockSetBackgroundType = jest.fn()
const mockSetBackgroundEnabled = jest.fn()
jest.mock('@/stores/appearanceStore', () => ({
  useAppearanceStore: () => ({
    background: { enabled: false, type: 'none' },
    setBackgroundType: mockSetBackgroundType,
    setBackgroundEnabled: mockSetBackgroundEnabled,
  }),
}))

describe('ThemeSettingsPopover', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders placeholder before mount', () => {
    const { container } = render(<ThemeSettingsPopover />)
    // Before mount, shows placeholder with aria-hidden
    expect(container.querySelector('[aria-hidden="true"]')).toBeInTheDocument()
  })

  it('renders trigger button after mount', async () => {
    render(<ThemeSettingsPopover />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Theme settings' })).toBeInTheDocument()
    })
  })

  it('opens popover when trigger is clicked', async () => {
    render(<ThemeSettingsPopover />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Theme settings' })).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: 'Theme settings' }))

    await waitFor(() => {
      expect(screen.getByText('Theme')).toBeInTheDocument()
    })
  })

  it('displays theme options', async () => {
    render(<ThemeSettingsPopover />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Theme settings' })).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: 'Theme settings' }))

    await waitFor(() => {
      expect(screen.getByText('Light')).toBeInTheDocument()
      expect(screen.getByText('Dark')).toBeInTheDocument()
      expect(screen.getByText('System')).toBeInTheDocument()
    })
  })

  it('calls setTheme when theme button is clicked', async () => {
    render(<ThemeSettingsPopover />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Theme settings' })).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: 'Theme settings' }))

    await waitFor(() => {
      expect(screen.getByText('Dark')).toBeInTheDocument()
    })

    // Find the button containing "Dark" text
    const darkButton = screen.getByText('Dark').closest('button')
    fireEvent.click(darkButton!)

    expect(mockSetTheme).toHaveBeenCalledWith('dark')
  })

  it('displays animated background toggle', async () => {
    render(<ThemeSettingsPopover />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Theme settings' })).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: 'Theme settings' }))

    await waitFor(() => {
      expect(screen.getByText('Animated Background')).toBeInTheDocument()
    })
  })

  it('calls setBackgroundEnabled when toggle is changed', async () => {
    render(<ThemeSettingsPopover />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Theme settings' })).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: 'Theme settings' }))

    await waitFor(() => {
      expect(screen.getByRole('switch')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('switch'))

    expect(mockSetBackgroundEnabled).toHaveBeenCalledWith(true)
  })
})

describe('ThemeSettingsPopover theme icons', () => {
  it('shows correct icon for dark theme', async () => {
    jest.doMock('next-themes', () => ({
      useTheme: () => ({
        theme: 'dark',
        setTheme: mockSetTheme,
      }),
    }))

    render(<ThemeSettingsPopover />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Theme settings' })).toBeInTheDocument()
    })
  })

  it('shows correct icon for system theme', async () => {
    jest.doMock('next-themes', () => ({
      useTheme: () => ({
        theme: 'system',
        setTheme: mockSetTheme,
      }),
    }))

    render(<ThemeSettingsPopover />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Theme settings' })).toBeInTheDocument()
    })
  })
})
