import { render, screen } from '@testing-library/react'
import { ThemeProvider } from '../theme-provider'

// Mock next-themes
jest.mock('next-themes', () => ({
  ThemeProvider: ({ children, ...props }: { children: React.ReactNode }) => (
    <div data-testid="theme-provider" data-props={JSON.stringify(props)}>
      {children}
    </div>
  ),
}))

describe('ThemeProvider', () => {
  it('renders children', () => {
    render(
      <ThemeProvider>
        <div data-testid="child">Content</div>
      </ThemeProvider>
    )
    expect(screen.getByTestId('child')).toBeInTheDocument()
    expect(screen.getByText('Content')).toBeInTheDocument()
  })

  it('passes props to NextThemesProvider', () => {
    render(
      <ThemeProvider attribute="class" defaultTheme="dark" enableSystem>
        <div>Content</div>
      </ThemeProvider>
    )
    const provider = screen.getByTestId('theme-provider')
    const props = JSON.parse(provider.getAttribute('data-props') || '{}')
    expect(props).toEqual({
      attribute: 'class',
      defaultTheme: 'dark',
      enableSystem: true,
    })
  })

  it('wraps content with NextThemesProvider', () => {
    render(
      <ThemeProvider>
        <button>Click me</button>
      </ThemeProvider>
    )
    expect(screen.getByTestId('theme-provider')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Click me' })).toBeInTheDocument()
  })
})
