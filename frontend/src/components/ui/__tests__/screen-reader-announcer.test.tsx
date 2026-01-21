import { render, screen, act } from '@testing-library/react'
import { ScreenReaderAnnouncerProvider, useAnnouncer } from '../screen-reader-announcer'

// Test component that uses the useAnnouncer hook
function TestComponent() {
  const { announce } = useAnnouncer()
  return (
    <div>
      <button onClick={() => announce('Polite message')}>Announce Polite</button>
      <button onClick={() => announce('Assertive message', 'assertive')}>Announce Assertive</button>
    </div>
  )
}

describe('ScreenReaderAnnouncerProvider', () => {
  beforeEach(() => {
    jest.useFakeTimers()
  })

  afterEach(() => {
    jest.useRealTimers()
  })

  it('renders children', () => {
    render(
      <ScreenReaderAnnouncerProvider>
        <div data-testid="child">Child content</div>
      </ScreenReaderAnnouncerProvider>
    )
    expect(screen.getByTestId('child')).toBeInTheDocument()
  })

  it('provides aria-live regions', () => {
    render(
      <ScreenReaderAnnouncerProvider>
        <div>Content</div>
      </ScreenReaderAnnouncerProvider>
    )

    expect(screen.getByRole('status')).toBeInTheDocument()
    expect(screen.getByRole('alert')).toBeInTheDocument()
  })

  it('announces polite messages', async () => {
    render(
      <ScreenReaderAnnouncerProvider>
        <TestComponent />
      </ScreenReaderAnnouncerProvider>
    )

    const button = screen.getByRole('button', { name: /announce polite/i })

    act(() => {
      button.click()
    })

    act(() => {
      jest.advanceTimersByTime(150)
    })

    expect(screen.getByRole('status')).toHaveTextContent('Polite message')
  })

  it('announces assertive messages', () => {
    render(
      <ScreenReaderAnnouncerProvider>
        <TestComponent />
      </ScreenReaderAnnouncerProvider>
    )

    const button = screen.getByRole('button', { name: /announce assertive/i })

    act(() => {
      button.click()
    })

    act(() => {
      jest.advanceTimersByTime(150)
    })

    expect(screen.getByRole('alert')).toHaveTextContent('Assertive message')
  })

  it('clears messages after timeout', () => {
    render(
      <ScreenReaderAnnouncerProvider>
        <TestComponent />
      </ScreenReaderAnnouncerProvider>
    )

    const button = screen.getByRole('button', { name: /announce polite/i })

    act(() => {
      button.click()
    })

    act(() => {
      jest.advanceTimersByTime(150)
    })

    expect(screen.getByRole('status')).toHaveTextContent('Polite message')

    act(() => {
      jest.advanceTimersByTime(1000)
    })

    expect(screen.getByRole('status')).toHaveTextContent('')
  })
})

describe('useAnnouncer', () => {
  it('throws error when used outside provider', () => {
    const consoleError = jest.spyOn(console, 'error').mockImplementation(() => {})

    function ComponentWithoutProvider() {
      useAnnouncer()
      return null
    }

    expect(() => render(<ComponentWithoutProvider />)).toThrow(
      'useAnnouncer must be used within ScreenReaderAnnouncerProvider'
    )

    consoleError.mockRestore()
  })
})
