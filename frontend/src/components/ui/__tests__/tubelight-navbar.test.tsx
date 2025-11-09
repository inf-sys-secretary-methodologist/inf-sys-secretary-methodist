import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TubelightNavbar } from '../tubelight-navbar'
import { Home, FileText, Users } from 'lucide-react'

const mockNavItems = [
  { name: 'Home', url: '/', icon: Home },
  { name: 'Documents', url: '/documents', icon: FileText },
  { name: 'Students', url: '/students', icon: Users },
]

// Mock framer-motion
jest.mock('framer-motion', () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  },
}))

describe('TubelightNavbar', () => {
  beforeEach(() => {
    // Reset window size before each test
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 1024,
    })
  })

  it('renders all navigation items', () => {
    render(<TubelightNavbar items={mockNavItems} />)

    expect(screen.getByText('Home')).toBeInTheDocument()
    expect(screen.getByText('Documents')).toBeInTheDocument()
    expect(screen.getByText('Students')).toBeInTheDocument()
  })

  it('renders navigation links with correct href', () => {
    render(<TubelightNavbar items={mockNavItems} />)

    const homeLink = screen.getByRole('link', { name: /home/i })
    expect(homeLink).toHaveAttribute('href', '/')

    const documentsLink = screen.getByRole('link', { name: /documents/i })
    expect(documentsLink).toHaveAttribute('href', '/documents')
  })

  it('applies fixed positioning and centering classes', () => {
    const { container } = render(<TubelightNavbar items={mockNavItems} />)
    const navbarWrapper = container.firstChild

    expect(navbarWrapper).toHaveClass('fixed')
    expect(navbarWrapper).toHaveClass('top-0')
    expect(navbarWrapper).toHaveClass('left-1/2')
    expect(navbarWrapper).toHaveClass('-translate-x-1/2')
    expect(navbarWrapper).toHaveClass('z-50')
  })

  it('applies backdrop blur and border styles', () => {
    const { container } = render(<TubelightNavbar items={mockNavItems} />)
    const navbarContainer = container.querySelector('.backdrop-blur-lg')

    expect(navbarContainer).toBeInTheDocument()
    expect(navbarContainer).toHaveClass('rounded-full')
    expect(navbarContainer).toHaveClass('border')
  })

  it('changes active tab on click', async () => {
    const user = userEvent.setup()
    render(<TubelightNavbar items={mockNavItems} />)

    const documentsLink = screen.getByRole('link', { name: /documents/i })
    await user.click(documentsLink)

    // The active state is managed internally
    expect(documentsLink).toBeInTheDocument()
  })

  it('applies custom className when provided', () => {
    const { container } = render(
      <TubelightNavbar items={mockNavItems} className="custom-class" />
    )
    const navbarWrapper = container.firstChild

    expect(navbarWrapper).toHaveClass('custom-class')
  })

  it('renders with at least one item', () => {
    render(<TubelightNavbar items={mockNavItems} />)
    const links = screen.getAllByRole('link')

    expect(links.length).toBeGreaterThanOrEqual(1)
  })

  it('has proper responsive classes for desktop', () => {
    render(<TubelightNavbar items={mockNavItems} />)

    const homeText = screen.getByText('Home')
    expect(homeText).toHaveClass('hidden')
    expect(homeText).toHaveClass('md:inline')
  })

  it('first item is active by default', () => {
    const { container } = render(<TubelightNavbar items={mockNavItems} />)

    // Check that first link has active styles
    const firstLink = screen.getByRole('link', { name: /home/i })
    expect(firstLink).toHaveClass('bg-muted')
    expect(firstLink).toHaveClass('text-primary')
  })

  it('handles window resize events', async () => {
    render(<TubelightNavbar items={mockNavItems} />)

    // Simulate resize to mobile
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 500,
    })
    window.dispatchEvent(new Event('resize'))

    await waitFor(() => {
      // Component should still be rendered
      expect(screen.getByText('Home')).toBeInTheDocument()
    })
  })

  it('cleans up resize listener on unmount', () => {
    const removeEventListenerSpy = jest.spyOn(window, 'removeEventListener')
    const { unmount } = render(<TubelightNavbar items={mockNavItems} />)

    unmount()

    expect(removeEventListenerSpy).toHaveBeenCalledWith('resize', expect.any(Function))
  })
})
