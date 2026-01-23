import { render, screen, fireEvent } from '@testing-library/react'
import { NavBar, NavItem } from '../tubelight-navbar'
import { Home, Settings, User } from 'lucide-react'

// Mock next/navigation
const mockPathname = jest.fn()
jest.mock('next/navigation', () => ({
  usePathname: () => mockPathname(),
}))

// Mock next/link
jest.mock('next/link', () => {
  const MockLink = ({ children, href, ...props }: { children: React.ReactNode; href: string }) => (
    <a href={href} {...props}>
      {children}
    </a>
  )
  MockLink.displayName = 'MockLink'
  return MockLink
})

describe('NavBar', () => {
  const mockItems: NavItem[] = [
    { name: 'Home', url: '/', icon: Home },
    { name: 'Settings', url: '/settings', icon: Settings },
    { name: 'Profile', url: '/profile', icon: User },
  ]

  beforeEach(() => {
    mockPathname.mockReturnValue('/')
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('renders navigation element', () => {
    render(<NavBar items={mockItems} />)
    expect(screen.getByRole('navigation')).toBeInTheDocument()
  })

  it('renders all nav items', () => {
    render(<NavBar items={mockItems} />)
    expect(screen.getByText('Home')).toBeInTheDocument()
    expect(screen.getByText('Settings')).toBeInTheDocument()
    expect(screen.getByText('Profile')).toBeInTheDocument()
  })

  it('renders links with correct hrefs', () => {
    render(<NavBar items={mockItems} />)
    const links = screen.getAllByRole('link')
    expect(links[0]).toHaveAttribute('href', '/')
    expect(links[1]).toHaveAttribute('href', '/settings')
    expect(links[2]).toHaveAttribute('href', '/profile')
  })

  it('applies custom className', () => {
    render(<NavBar items={mockItems} className="custom-nav" />)
    const nav = screen.getByRole('navigation')
    expect(nav).toHaveClass('custom-nav')
  })

  it('applies default fixed positioning classes', () => {
    render(<NavBar items={mockItems} />)
    const nav = screen.getByRole('navigation')
    expect(nav).toHaveClass('fixed')
    expect(nav).toHaveClass('top-8')
    expect(nav).toHaveClass('left-1/2')
  })

  it('highlights active item based on pathname', () => {
    mockPathname.mockReturnValue('/settings')
    const { container } = render(<NavBar items={mockItems} />)

    // Active item should have white text
    const activeItem = container.querySelector('.text-white')
    expect(activeItem).toBeInTheDocument()
  })

  it('shows glow effect for active item', () => {
    mockPathname.mockReturnValue('/settings')
    const { container } = render(<NavBar items={mockItems} />)

    // Active item should have gradient background
    const gradientElement = container.querySelector('.bg-gradient-to-r.from-blue-500')
    expect(gradientElement).toBeInTheDocument()
  })

  it('shows hover effect on mouse enter', () => {
    const { container } = render(<NavBar items={mockItems} />)
    const links = screen.getAllByRole('link')

    // Hover over non-active item
    fireEvent.mouseEnter(links[1])

    // Should show hover gradient
    const hoverGradient = container.querySelector('.bg-gradient-to-r.from-gray-200')
    expect(hoverGradient).toBeInTheDocument()
  })

  it('removes hover effect on mouse leave', () => {
    const { container } = render(<NavBar items={mockItems} />)
    const links = screen.getAllByRole('link')

    // Hover and then leave
    fireEvent.mouseEnter(links[1])
    fireEvent.mouseLeave(links[1])

    // Hover gradient should be removed for non-active items
    const hoverGradient = container.querySelector('.from-gray-200.scale-95')
    expect(hoverGradient).not.toBeInTheDocument()
  })

  it('renders icons for each item', () => {
    const { container } = render(<NavBar items={mockItems} />)
    const icons = container.querySelectorAll('.h-4.w-4')
    expect(icons).toHaveLength(3)
  })

  it('hides item names on small screens', () => {
    const { container } = render(<NavBar items={mockItems} />)
    const itemNames = container.querySelectorAll('.hidden.lg\\:inline')
    expect(itemNames).toHaveLength(3)
  })

  it('applies backdrop blur to container', () => {
    const { container } = render(<NavBar items={mockItems} />)
    const blurContainer = container.querySelector('.backdrop-blur-lg')
    expect(blurContainer).toBeInTheDocument()
  })

  it('applies shadow to container', () => {
    const { container } = render(<NavBar items={mockItems} />)
    const shadowContainer = container.querySelector('.shadow-lg')
    expect(shadowContainer).toBeInTheDocument()
  })

  it('applies border to container', () => {
    const { container } = render(<NavBar items={mockItems} />)
    const borderContainer = container.querySelector('.border')
    expect(borderContainer).toBeInTheDocument()
  })

  it('applies rounded-full to container', () => {
    const { container } = render(<NavBar items={mockItems} />)
    const roundedContainer = container.querySelector('.rounded-full')
    expect(roundedContainer).toBeInTheDocument()
  })

  it('renders empty when no items provided', () => {
    render(<NavBar items={[]} />)
    const links = screen.queryAllByRole('link')
    expect(links).toHaveLength(0)
  })

  it('handles single item', () => {
    const singleItem: NavItem[] = [{ name: 'Home', url: '/', icon: Home }]
    render(<NavBar items={singleItem} />)
    expect(screen.getByText('Home')).toBeInTheDocument()
    expect(screen.getAllByRole('link')).toHaveLength(1)
  })

  it('active item has box shadow', () => {
    mockPathname.mockReturnValue('/')
    const { container } = render(<NavBar items={mockItems} />)

    // Find element with box shadow style
    const activeGlow = container.querySelector('[style*="box-shadow"]')
    expect(activeGlow).toBeInTheDocument()
  })

  it('non-active items have gray text color', () => {
    mockPathname.mockReturnValue('/')
    const { container } = render(<NavBar items={mockItems} />)

    // Non-active items should have gray text
    const grayTextItems = container.querySelectorAll('.text-gray-700')
    expect(grayTextItems.length).toBeGreaterThan(0)
  })

  it('has z-index for proper layering', () => {
    render(<NavBar items={mockItems} />)
    const nav = screen.getByRole('navigation')
    expect(nav).toHaveClass('z-40')
  })

  it('uses transform to center navbar', () => {
    render(<NavBar items={mockItems} />)
    const nav = screen.getByRole('navigation')
    expect(nav).toHaveClass('-translate-x-1/2')
  })
})
