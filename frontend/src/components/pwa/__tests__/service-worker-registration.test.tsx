import { render, act, waitFor } from '@testing-library/react'
import { ServiceWorkerRegistration } from '../service-worker-registration'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      updateAvailable: 'A new version is available. Reload to update?',
    }
    return translations[key] || key
  },
}))

describe('ServiceWorkerRegistration', () => {
  let originalConfirm: typeof window.confirm
  let mockRegistration: {
    update: jest.Mock
    addEventListener: jest.Mock
    installing: ServiceWorker | null
  }
  let mockServiceWorker: {
    register: jest.Mock
    controller: ServiceWorkerContainer['controller']
    addEventListener: jest.Mock
  }

  beforeEach(() => {
    jest.useFakeTimers()

    // Store originals
    originalConfirm = window.confirm

    // Mock confirm
    window.confirm = jest.fn()

    // Create mock registration
    mockRegistration = {
      update: jest.fn(),
      addEventListener: jest.fn(),
      installing: null,
    }

    // Create mock service worker
    mockServiceWorker = {
      register: jest.fn().mockResolvedValue(mockRegistration),
      controller: {} as ServiceWorker,
      addEventListener: jest.fn(),
    }

    // Set up navigator.serviceWorker
    Object.defineProperty(global.navigator, 'serviceWorker', {
      value: mockServiceWorker,
      writable: true,
      configurable: true,
    })
  })

  afterEach(() => {
    jest.useRealTimers()
    jest.clearAllMocks()
    window.confirm = originalConfirm
  })

  it('renders nothing', () => {
    const { container } = render(<ServiceWorkerRegistration />)
    expect(container).toBeEmptyDOMElement()
  })

  it('registers service worker on mount', async () => {
    render(<ServiceWorkerRegistration />)

    await waitFor(() => {
      expect(mockServiceWorker.register).toHaveBeenCalledWith('/sw.js', {
        scope: '/',
        updateViaCache: 'none',
      })
    })
  })

  it('sets up periodic update check every hour', async () => {
    const setIntervalSpy = jest.spyOn(global, 'setInterval')

    render(<ServiceWorkerRegistration />)

    await waitFor(() => {
      expect(setIntervalSpy).toHaveBeenCalledWith(expect.any(Function), 60 * 60 * 1000)
    })

    setIntervalSpy.mockRestore()
  })

  it('calls registration.update periodically', async () => {
    render(<ServiceWorkerRegistration />)

    await waitFor(() => {
      expect(mockRegistration.addEventListener).toHaveBeenCalledWith(
        'updatefound',
        expect.any(Function)
      )
    })

    // Advance time by 1 hour
    act(() => {
      jest.advanceTimersByTime(60 * 60 * 1000)
    })

    expect(mockRegistration.update).toHaveBeenCalled()
  })

  it('adds updatefound event listener to registration', async () => {
    render(<ServiceWorkerRegistration />)

    await waitFor(() => {
      expect(mockRegistration.addEventListener).toHaveBeenCalledWith(
        'updatefound',
        expect.any(Function)
      )
    })
  })

  it('adds controllerchange event listener', async () => {
    render(<ServiceWorkerRegistration />)

    await waitFor(() => {
      expect(mockServiceWorker.addEventListener).toHaveBeenCalledWith(
        'controllerchange',
        expect.any(Function)
      )
    })
  })

  it('does not register if serviceWorker not supported', () => {
    // Remove serviceWorker from navigator
    Object.defineProperty(global.navigator, 'serviceWorker', {
      value: undefined,
      writable: true,
      configurable: true,
    })

    render(<ServiceWorkerRegistration />)

    // Should not throw and should not call register
    expect(mockServiceWorker.register).not.toHaveBeenCalled()
  })

  it('handles registration error gracefully', async () => {
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation()
    mockServiceWorker.register.mockRejectedValueOnce(new Error('Registration failed'))

    render(<ServiceWorkerRegistration />)

    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalledWith(
        'Service Worker registration failed:',
        expect.any(Error)
      )
    })

    consoleSpy.mockRestore()
  })

  describe('update handling', () => {
    let updateFoundHandler: () => void
    let newWorker: { addEventListener: jest.Mock; postMessage: jest.Mock; state: string }

    beforeEach(() => {
      newWorker = {
        addEventListener: jest.fn(),
        postMessage: jest.fn(),
        state: 'installing',
      }

      mockRegistration.installing = newWorker as unknown as ServiceWorker

      mockRegistration.addEventListener.mockImplementation((event: string, handler: () => void) => {
        if (event === 'updatefound') {
          updateFoundHandler = handler
        }
      })
    })

    it('shows confirm dialog when new worker is installed', async () => {
      ;(window.confirm as jest.Mock).mockReturnValue(false)

      render(<ServiceWorkerRegistration />)

      await waitFor(() => {
        expect(mockRegistration.addEventListener).toHaveBeenCalled()
      })

      // Trigger updatefound
      act(() => {
        updateFoundHandler()
      })

      // Get the statechange handler
      const stateChangeHandler = newWorker.addEventListener.mock.calls.find(
        (call) => call[0] === 'statechange'
      )?.[1]

      // Simulate worker installed
      newWorker.state = 'installed'
      act(() => {
        stateChangeHandler?.()
      })

      expect(window.confirm).toHaveBeenCalledWith('A new version is available. Reload to update?')
    })

    it('sends SKIP_WAITING message when user confirms update', async () => {
      ;(window.confirm as jest.Mock).mockReturnValue(true)

      // Mock location.reload
      const originalLocation = window.location
      const reloadMock = jest.fn()
      Object.defineProperty(window, 'location', {
        value: { ...originalLocation, reload: reloadMock },
        writable: true,
      })

      render(<ServiceWorkerRegistration />)

      await waitFor(() => {
        expect(mockRegistration.addEventListener).toHaveBeenCalled()
      })

      // Trigger updatefound
      act(() => {
        updateFoundHandler()
      })

      // Get the statechange handler
      const stateChangeHandler = newWorker.addEventListener.mock.calls.find(
        (call) => call[0] === 'statechange'
      )?.[1]

      // Simulate worker installed
      newWorker.state = 'installed'
      act(() => {
        stateChangeHandler?.()
      })

      expect(newWorker.postMessage).toHaveBeenCalledWith({ type: 'SKIP_WAITING' })
      expect(reloadMock).toHaveBeenCalled()

      // Restore location
      Object.defineProperty(window, 'location', {
        value: originalLocation,
        writable: true,
      })
    })

    it('does nothing when user declines update', async () => {
      ;(window.confirm as jest.Mock).mockReturnValue(false)

      render(<ServiceWorkerRegistration />)

      await waitFor(() => {
        expect(mockRegistration.addEventListener).toHaveBeenCalled()
      })

      // Trigger updatefound
      act(() => {
        updateFoundHandler()
      })

      // Get the statechange handler
      const stateChangeHandler = newWorker.addEventListener.mock.calls.find(
        (call) => call[0] === 'statechange'
      )?.[1]

      // Simulate worker installed
      newWorker.state = 'installed'
      act(() => {
        stateChangeHandler?.()
      })

      expect(newWorker.postMessage).not.toHaveBeenCalled()
    })

    it('does not show prompt if no controller exists', async () => {
      mockServiceWorker.controller = null

      render(<ServiceWorkerRegistration />)

      await waitFor(() => {
        expect(mockRegistration.addEventListener).toHaveBeenCalled()
      })

      // Trigger updatefound
      act(() => {
        updateFoundHandler()
      })

      // Get the statechange handler
      const stateChangeHandler = newWorker.addEventListener.mock.calls.find(
        (call) => call[0] === 'statechange'
      )?.[1]

      // Simulate worker installed
      newWorker.state = 'installed'
      act(() => {
        stateChangeHandler?.()
      })

      // Should not show confirm because no controller (first install)
      expect(window.confirm).not.toHaveBeenCalled()
    })

    it('does not add statechange listener if no installing worker', async () => {
      mockRegistration.installing = null

      render(<ServiceWorkerRegistration />)

      await waitFor(() => {
        expect(mockRegistration.addEventListener).toHaveBeenCalled()
      })

      // Trigger updatefound
      act(() => {
        updateFoundHandler()
      })

      // No statechange listener should be added
      expect(newWorker.addEventListener).not.toHaveBeenCalled()
    })
  })
})
