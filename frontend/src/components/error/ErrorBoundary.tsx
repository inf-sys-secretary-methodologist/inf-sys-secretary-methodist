'use client'

import React, { Component, ErrorInfo, ReactNode } from 'react'
import { AlertCircle, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface Props {
  children: ReactNode
  /**
   * Fallback component to render when an error occurs
   */
  fallback?: ReactNode
  /**
   * Callback function called when an error is caught
   */
  onError?: (error: Error, errorInfo: ErrorInfo) => void
  /**
   * Custom error message to display
   */
  errorMessage?: string
  /**
   * Whether to show error details in development mode
   */
  showDetails?: boolean
}

interface State {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
}

/**
 * Error Boundary component for catching React errors in component tree
 *
 * @example
 * ```tsx
 * <ErrorBoundary>
 *   <MyComponent />
 * </ErrorBoundary>
 * ```
 *
 * @example With custom fallback
 * ```tsx
 * <ErrorBoundary fallback={<CustomErrorUI />}>
 *   <MyComponent />
 * </ErrorBoundary>
 * ```
 *
 * @example With error handler
 * ```tsx
 * <ErrorBoundary
 *   onError={(error, errorInfo) => {
 *     logErrorToService(error, errorInfo)
 *   }}
 * >
 *   <MyComponent />
 * </ErrorBoundary>
 * ```
 */
export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    }
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    // Update state so the next render will show the fallback UI
    return {
      hasError: true,
      error,
    }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Log error information
    console.error('ErrorBoundary caught an error:', {
      error,
      errorInfo,
      componentStack: errorInfo.componentStack,
      timestamp: new Date().toISOString(),
    })

    // Update state with error info
    this.setState({
      errorInfo,
    })

    // Call custom error handler if provided
    this.props.onError?.(error, errorInfo)

    // TODO: Send to error tracking service (e.g., Sentry)
    // logErrorToService(error, errorInfo)
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    })
  }

  render() {
    const { hasError, error, errorInfo } = this.state
    const { children, fallback, errorMessage, showDetails = true } = this.props

    if (hasError) {
      // Render custom fallback if provided
      if (fallback) {
        return fallback
      }

      // Default error UI
      return (
        <div className="p-6 rounded-lg border border-destructive/20 bg-destructive/5">
          <div className="flex items-start gap-4">
            {/* Error Icon */}
            <div className="flex-shrink-0">
              <div className="rounded-full bg-destructive/10 p-3">
                <AlertCircle className="h-6 w-6 text-destructive" />
              </div>
            </div>

            <div className="flex-1 space-y-4">
              {/* Error Title */}
              <div>
                <h3 className="text-lg font-semibold text-destructive">
                  Произошла ошибка
                </h3>
                <p className="text-sm text-muted-foreground mt-1">
                  {errorMessage ||
                    'Этот компонент столкнулся с ошибкой и не может быть отображён.'}
                </p>
              </div>

              {/* Error Details (Development only or if showDetails is true) */}
              {showDetails &&
                process.env.NODE_ENV === 'development' &&
                error && (
                  <div className="space-y-2">
                    <div className="p-3 rounded bg-destructive/10 border border-destructive/20">
                      <p className="text-xs font-semibold text-destructive mb-1">
                        Сообщение об ошибке:
                      </p>
                      <p className="text-xs text-destructive/80 font-mono break-words">
                        {error.message}
                      </p>
                    </div>

                    {errorInfo?.componentStack && (
                      <details className="text-xs">
                        <summary className="cursor-pointer text-muted-foreground hover:text-foreground">
                          Показать стек компонентов
                        </summary>
                        <pre className="mt-2 p-3 rounded bg-muted overflow-x-auto text-xs">
                          {errorInfo.componentStack}
                        </pre>
                      </details>
                    )}
                  </div>
                )}

              {/* Reset Button */}
              <div>
                <Button
                  onClick={this.handleReset}
                  variant="outline"
                  size="sm"
                  className="gap-2"
                >
                  <RefreshCw className="h-4 w-4" />
                  Попробовать снова
                </Button>
              </div>
            </div>
          </div>
        </div>
      )
    }

    return children
  }
}

/**
 * Hook-based wrapper for ErrorBoundary (for convenience)
 *
 * @example
 * ```tsx
 * function MyComponent() {
 *   return (
 *     <ErrorBoundaryWrapper>
 *       <ChildComponent />
 *     </ErrorBoundaryWrapper>
 *   )
 * }
 * ```
 */
export function ErrorBoundaryWrapper({
  children,
  ...props
}: Omit<Props, 'children'> & { children: ReactNode }) {
  return <ErrorBoundary {...props}>{children}</ErrorBoundary>
}
