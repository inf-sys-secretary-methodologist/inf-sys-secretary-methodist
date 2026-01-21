import { render, screen, fireEvent } from '@testing-library/react'
import { ErrorBoundary, ErrorBoundaryWrapper, TranslatedErrorBoundary } from '../ErrorBoundary'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      title: 'An Error Occurred',
      defaultMessage: 'This component encountered an error.',
      details: 'Error details',
      showStack: 'Show stack',
      retry: 'Try Again',
    }
    return translations[key] || key
  },
}))

// Component that throws an error
function ThrowError({ shouldThrow }: { shouldThrow: boolean }) {
  if (shouldThrow) {
    throw new Error('Test error message')
  }
  return <div>No error</div>
}

describe('ErrorBoundary', () => {
  // Suppress console.error for expected errors
  const originalError = console.error
  beforeAll(() => {
    console.error = jest.fn()
  })
  afterAll(() => {
    console.error = originalError
  })

  it('renders children when no error', () => {
    render(
      <ErrorBoundary>
        <div data-testid="child">Child content</div>
      </ErrorBoundary>
    )
    expect(screen.getByTestId('child')).toBeInTheDocument()
    expect(screen.getByText('Child content')).toBeInTheDocument()
  })

  it('renders error UI when child throws', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(screen.getByText('An Error Occurred')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /try again/i })).toBeInTheDocument()
  })

  it('renders custom fallback when provided', () => {
    render(
      <ErrorBoundary fallback={<div>Custom error UI</div>}>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(screen.getByText('Custom error UI')).toBeInTheDocument()
  })

  it('renders custom error message when provided', () => {
    render(
      <ErrorBoundary errorMessage="Something went wrong">
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
  })

  it('calls onError callback when error occurs', () => {
    const onError = jest.fn()
    render(
      <ErrorBoundary onError={onError}>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(onError).toHaveBeenCalled()
    expect(onError.mock.calls[0][0]).toBeInstanceOf(Error)
    expect(onError.mock.calls[0][0].message).toBe('Test error message')
  })

  it('renders retry button that can be clicked', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(screen.getByText('An Error Occurred')).toBeInTheDocument()

    // Click retry button - this resets the hasError state, then re-catches the error
    const retryButton = screen.getByRole('button', { name: /try again/i })
    expect(retryButton).toBeInTheDocument()

    // Verify button click doesn't throw
    expect(() => fireEvent.click(retryButton)).not.toThrow()
  })

  it('shows error details in development mode', () => {
    const originalEnv = process.env.NODE_ENV
    Object.defineProperty(process.env, 'NODE_ENV', { value: 'development', writable: true })

    render(
      <ErrorBoundary showDetails={true}>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(screen.getByText('Test error message')).toBeInTheDocument()

    Object.defineProperty(process.env, 'NODE_ENV', { value: originalEnv, writable: true })
  })

  it('uses custom translations when provided', () => {
    render(
      <ErrorBoundary
        translations={{
          title: 'Custom Title',
          defaultMessage: 'Custom message',
          retry: 'Retry Now',
        }}
      >
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(screen.getByText('Custom Title')).toBeInTheDocument()
    expect(screen.getByText('Custom message')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /retry now/i })).toBeInTheDocument()
  })
})

describe('ErrorBoundaryWrapper', () => {
  const originalError = console.error
  beforeAll(() => {
    console.error = jest.fn()
  })
  afterAll(() => {
    console.error = originalError
  })

  it('renders children', () => {
    render(
      <ErrorBoundaryWrapper>
        <div data-testid="child">Content</div>
      </ErrorBoundaryWrapper>
    )
    expect(screen.getByTestId('child')).toBeInTheDocument()
  })

  it('catches errors like ErrorBoundary', () => {
    render(
      <ErrorBoundaryWrapper>
        <ThrowError shouldThrow={true} />
      </ErrorBoundaryWrapper>
    )
    expect(screen.getByText('An Error Occurred')).toBeInTheDocument()
  })
})

describe('TranslatedErrorBoundary', () => {
  const originalError = console.error
  beforeAll(() => {
    console.error = jest.fn()
  })
  afterAll(() => {
    console.error = originalError
  })

  it('renders children', () => {
    render(
      <TranslatedErrorBoundary>
        <div data-testid="child">Content</div>
      </TranslatedErrorBoundary>
    )
    expect(screen.getByTestId('child')).toBeInTheDocument()
  })

  it('uses translated strings for error UI', () => {
    render(
      <TranslatedErrorBoundary>
        <ThrowError shouldThrow={true} />
      </TranslatedErrorBoundary>
    )
    // Should use translated title from useTranslations mock
    expect(screen.getByText('An Error Occurred')).toBeInTheDocument()
  })
})
