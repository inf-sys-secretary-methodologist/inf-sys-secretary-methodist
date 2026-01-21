import { render, screen } from '@testing-library/react'
import { MobileNav } from '../MobileNav'
import type { NavItem } from '@/config/navigation'
import { Home, FileText, Calendar } from 'lucide-react'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      openMenu: 'Open menu',
      navigation: 'Navigation',
      mobileNavigation: 'Mobile Navigation',
      close: 'Close',
      dashboard: 'Dashboard',
      documents: 'Documents',
      calendar: 'Calendar',
    }
    return translations[key] || key
  },
}))

// Mock next/navigation
const mockPathname = '/dashboard'
jest.mock('next/navigation', () => ({
  usePathname: () => mockPathname,
}))

// Mock Sheet components
jest.mock('@/components/ui/sheet', () => ({
  Sheet: ({ children, open }: { children: React.ReactNode; open: boolean }) => (
    <div data-testid="sheet" data-open={open}>
      {children}
    </div>
  ),
  SheetTrigger: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="sheet-trigger">{children}</div>
  ),
  SheetContent: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="sheet-content">{children}</div>
  ),
  SheetHeader: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="sheet-header">{children}</div>
  ),
  SheetTitle: ({ children }: { children: React.ReactNode }) => (
    <h2 data-testid="sheet-title">{children}</h2>
  ),
  SheetClose: ({ children }: { children: React.ReactNode }) => (
    <button data-testid="sheet-close">{children}</button>
  ),
}))

describe('MobileNav', () => {
  const mockItems: NavItem[] = [
    { url: '/dashboard', nameKey: 'dashboard', icon: Home },
    { url: '/documents', nameKey: 'documents', icon: FileText },
    { url: '/calendar', nameKey: 'calendar', icon: Calendar },
  ]

  it('renders menu button', () => {
    render(<MobileNav items={mockItems} />)
    expect(screen.getByRole('button', { name: /open menu/i })).toBeInTheDocument()
  })

  it('renders sheet component', () => {
    render(<MobileNav items={mockItems} />)
    expect(screen.getByTestId('sheet')).toBeInTheDocument()
  })

  it('renders sheet content', () => {
    render(<MobileNav items={mockItems} />)
    expect(screen.getByTestId('sheet-content')).toBeInTheDocument()
  })

  it('renders navigation title', () => {
    render(<MobileNav items={mockItems} />)
    expect(screen.getByText('Navigation')).toBeInTheDocument()
  })

  it('renders navigation items', () => {
    render(<MobileNav items={mockItems} />)
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Documents')).toBeInTheDocument()
    expect(screen.getByText('Calendar')).toBeInTheDocument()
  })

  it('renders close button', () => {
    render(<MobileNav items={mockItems} />)
    expect(screen.getByTestId('sheet-close')).toBeInTheDocument()
  })

  it('renders with mobile navigation aria-label', () => {
    render(<MobileNav items={mockItems} />)
    const nav = screen.getByRole('navigation', { name: /mobile navigation/i })
    expect(nav).toBeInTheDocument()
  })

  it('renders navigation links', () => {
    render(<MobileNav items={mockItems} />)
    const links = screen.getAllByRole('link')
    expect(links.length).toBe(3)
  })

  it('marks active link with aria-current', () => {
    render(<MobileNav items={mockItems} />)
    const links = screen.getAllByRole('link')
    const dashboardLink = links.find((link) => link.textContent?.includes('Dashboard'))
    expect(dashboardLink).toHaveAttribute('aria-current', 'page')
  })
})
