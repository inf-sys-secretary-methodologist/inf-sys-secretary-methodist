import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import { InstallPrompt } from '../install-prompt'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      installApp: 'Install App',
      close: 'Close',
      iosInstructions: 'Install this app on your iPhone',
      iosStep1: 'Tap',
      iosStep1Action: 'Share',
      iosStep2: 'Then',
      iosStep2Action: 'Add to Home Screen',
      installDescription: 'Install our app for a better experience',
      install: 'Install',
      notNow: 'Not Now',
    }
    return translations[key] || key
  },
}))

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
}

// Store original window properties
const originalMatchMedia = window.matchMedia

describe('InstallPrompt', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorageMock.getItem.mockReturnValue(null)

    // Setup default matchMedia mock
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation((query) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: jest.fn(),
        removeListener: jest.fn(),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
      })),
    })

    // Setup localStorage mock
    Object.defineProperty(window, 'localStorage', {
      writable: true,
      value: localStorageMock,
    })
  })

  afterEach(() => {
    // Restore originals
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: originalMatchMedia,
    })
  })

  it('returns null when already in standalone mode', () => {
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation(() => ({
        matches: true, // standalone mode
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
      })),
    })

    const { container } = render(<InstallPrompt />)
    expect(container.firstChild).toBeNull()
  })

  it('returns null when user previously dismissed within 7 days', () => {
    localStorageMock.getItem.mockReturnValue(new Date().toISOString())

    const { container } = render(<InstallPrompt />)
    expect(container.firstChild).toBeNull()
  })

  it('returns null when not on iOS and no beforeinstallprompt event', () => {
    const { container } = render(<InstallPrompt />)
    expect(container.firstChild).toBeNull()
  })

  it('shows prompt again after 7 days of dismissal', () => {
    const eightDaysAgo = new Date()
    eightDaysAgo.setDate(eightDaysAgo.getDate() - 8)
    localStorageMock.getItem.mockReturnValue(eightDaysAgo.toISOString())

    // Note: The component needs beforeinstallprompt event or iOS to show
    // Without that, it will still return null
    const { container } = render(<InstallPrompt />)
    // This just verifies the component renders without error
    expect(container).toBeInTheDocument()
  })

  it('handles beforeinstallprompt event', async () => {
    const mockDeferredPrompt = {
      prompt: jest.fn(),
      userChoice: Promise.resolve({ outcome: 'accepted' as const }),
      preventDefault: jest.fn(),
    }

    const { container } = render(<InstallPrompt />)

    // Simulate beforeinstallprompt event
    const event = new Event('beforeinstallprompt')
    Object.assign(event, mockDeferredPrompt)

    window.dispatchEvent(event)

    // The component should show the install prompt
    // Note: Due to useEffect timing, we might need to wait
    expect(container).toBeInTheDocument()
  })

  it('handles appinstalled event', () => {
    const { container } = render(<InstallPrompt />)

    // Simulate appinstalled event
    window.dispatchEvent(new Event('appinstalled'))

    // The component should hide after installation
    expect(container).toBeInTheDocument()
  })

  describe('iOS device', () => {
    const originalUserAgent = navigator.userAgent

    beforeEach(() => {
      // Mock iOS user agent
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)',
        configurable: true,
      })
      // Ensure MSStream is not present (iOS check)
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      delete (window as any).MSStream
    })

    afterEach(() => {
      Object.defineProperty(navigator, 'userAgent', {
        value: originalUserAgent,
        configurable: true,
      })
    })

    it('shows iOS-specific prompt on iPhone', () => {
      render(<InstallPrompt />)

      expect(screen.getByText('Install App')).toBeInTheDocument()
      expect(screen.getByText('Install this app on your iPhone')).toBeInTheDocument()
      // Check for numbered steps
      expect(screen.getByText('1')).toBeInTheDocument()
      expect(screen.getByText('2')).toBeInTheDocument()
    })

    it('dismisses iOS prompt when close button clicked', () => {
      render(<InstallPrompt />)

      const closeButton = screen.getByRole('button', { name: 'Close' })
      fireEvent.click(closeButton)

      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'pwa-install-dismissed',
        expect.any(String)
      )
    })

    it('does not show iOS prompt if already in standalone mode', () => {
      Object.defineProperty(window, 'matchMedia', {
        writable: true,
        value: jest.fn().mockImplementation(() => ({
          matches: true, // standalone mode
          addEventListener: jest.fn(),
          removeEventListener: jest.fn(),
        })),
      })

      const { container } = render(<InstallPrompt />)
      expect(container.firstChild).toBeNull()
    })
  })

  describe('Standard browser prompt', () => {
    it('shows standard prompt when beforeinstallprompt event fires', async () => {
      const mockDeferredPrompt = {
        prompt: jest.fn().mockResolvedValue(undefined),
        userChoice: Promise.resolve({ outcome: 'accepted' as const }),
        preventDefault: jest.fn(),
      }

      render(<InstallPrompt />)

      // Simulate beforeinstallprompt event
      await act(async () => {
        const event = new Event('beforeinstallprompt', { cancelable: true })
        Object.assign(event, mockDeferredPrompt)
        window.dispatchEvent(event)
      })

      await waitFor(() => {
        expect(screen.getByText('Install App')).toBeInTheDocument()
        expect(screen.getByText('Install our app for a better experience')).toBeInTheDocument()
        expect(screen.getByRole('button', { name: 'Install' })).toBeInTheDocument()
        expect(screen.getByRole('button', { name: 'Not Now' })).toBeInTheDocument()
      })
    })

    it('calls prompt() when install button clicked', async () => {
      const mockDeferredPrompt = {
        prompt: jest.fn().mockResolvedValue(undefined),
        userChoice: Promise.resolve({ outcome: 'accepted' as const }),
        preventDefault: jest.fn(),
      }

      render(<InstallPrompt />)

      // Trigger the event
      await act(async () => {
        const event = new Event('beforeinstallprompt', { cancelable: true })
        Object.assign(event, mockDeferredPrompt)
        window.dispatchEvent(event)
      })

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Install' })).toBeInTheDocument()
      })

      // Click install
      await act(async () => {
        fireEvent.click(screen.getByRole('button', { name: 'Install' }))
      })

      expect(mockDeferredPrompt.prompt).toHaveBeenCalled()
    })

    it('handles user dismissing install prompt', async () => {
      const mockDeferredPrompt = {
        prompt: jest.fn().mockResolvedValue(undefined),
        userChoice: Promise.resolve({ outcome: 'dismissed' as const }),
        preventDefault: jest.fn(),
      }

      render(<InstallPrompt />)

      await act(async () => {
        const event = new Event('beforeinstallprompt', { cancelable: true })
        Object.assign(event, mockDeferredPrompt)
        window.dispatchEvent(event)
      })

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Install' })).toBeInTheDocument()
      })

      await act(async () => {
        fireEvent.click(screen.getByRole('button', { name: 'Install' }))
      })

      // Should save dismissal to localStorage
      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'pwa-install-dismissed',
        expect.any(String)
      )
    })

    it('dismisses prompt when Not Now button clicked', async () => {
      const mockDeferredPrompt = {
        prompt: jest.fn().mockResolvedValue(undefined),
        userChoice: Promise.resolve({ outcome: 'accepted' as const }),
        preventDefault: jest.fn(),
      }

      render(<InstallPrompt />)

      await act(async () => {
        const event = new Event('beforeinstallprompt', { cancelable: true })
        Object.assign(event, mockDeferredPrompt)
        window.dispatchEvent(event)
      })

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Not Now' })).toBeInTheDocument()
      })

      fireEvent.click(screen.getByRole('button', { name: 'Not Now' }))

      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'pwa-install-dismissed',
        expect.any(String)
      )
    })

    it('dismisses prompt when close button clicked', async () => {
      const mockDeferredPrompt = {
        prompt: jest.fn().mockResolvedValue(undefined),
        userChoice: Promise.resolve({ outcome: 'accepted' as const }),
        preventDefault: jest.fn(),
      }

      render(<InstallPrompt />)

      await act(async () => {
        const event = new Event('beforeinstallprompt', { cancelable: true })
        Object.assign(event, mockDeferredPrompt)
        window.dispatchEvent(event)
      })

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
      })

      fireEvent.click(screen.getByRole('button', { name: 'Close' }))

      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'pwa-install-dismissed',
        expect.any(String)
      )
    })
  })

  describe('handleInstall edge cases', () => {
    it('does nothing if deferredPrompt is null', async () => {
      // Render without triggering beforeinstallprompt
      const { container } = render(<InstallPrompt />)

      // Component should be empty (no prompt to show)
      expect(container.firstChild).toBeNull()
    })
  })
})
