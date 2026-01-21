import { render, screen } from '@testing-library/react'
import { AppHeader } from '../AppHeader'
import type { NavItem } from '@/config/navigation'
import { FileText, Calendar, MessageCircle, Home, Settings } from 'lucide-react'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      dashboard: 'Dashboard',
      documents: 'Documents',
      calendar: 'Calendar',
      messages: 'Messages',
      settings: 'Settings',
      mainNavigation: 'Main Navigation',
    }
    return translations[key] || key
  },
}))

// Mock next/navigation
const mockPathname = '/dashboard'
jest.mock('next/navigation', () => ({
  usePathname: () => mockPathname,
}))

// Mock child components
jest.mock('@/components/UserMenu', () => ({
  UserMenu: () => <div data-testid="user-menu">User Menu</div>,
}))

jest.mock('@/components/theme-settings-popover', () => ({
  ThemeSettingsPopover: () => <div data-testid="theme-settings">Theme Settings</div>,
}))

jest.mock('@/components/language-switcher', () => ({
  LanguageSwitcher: () => <div data-testid="language-switcher">Language Switcher</div>,
}))

jest.mock('@/components/notifications/NotificationBell', () => ({
  NotificationBell: () => <div data-testid="notification-bell">Notification Bell</div>,
}))

jest.mock('../MobileNav', () => ({
  MobileNav: ({ items }: { items: NavItem[] }) => (
    <div data-testid="mobile-nav" data-items={items.length}>
      Mobile Nav
    </div>
  ),
}))

describe('AppHeader', () => {
  const mockItems: NavItem[] = [
    { url: '/dashboard', nameKey: 'dashboard', icon: Home },
    { url: '/documents', nameKey: 'documents', icon: FileText },
    { url: '/calendar', nameKey: 'calendar', icon: Calendar },
    { url: '/messages', nameKey: 'messages', icon: MessageCircle },
    { url: '/settings', nameKey: 'settings', icon: Settings },
  ]

  it('renders navigation items', () => {
    render(<AppHeader items={mockItems} />)
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Documents')).toBeInTheDocument()
    expect(screen.getByText('Calendar')).toBeInTheDocument()
  })

  it('renders MobileNav component', () => {
    render(<AppHeader items={mockItems} />)
    expect(screen.getByTestId('mobile-nav')).toBeInTheDocument()
    expect(screen.getByTestId('mobile-nav')).toHaveAttribute('data-items', '5')
  })

  it('renders UserMenu component', () => {
    render(<AppHeader items={mockItems} />)
    expect(screen.getAllByTestId('user-menu').length).toBeGreaterThan(0)
  })

  it('renders ThemeSettingsPopover component', () => {
    render(<AppHeader items={mockItems} />)
    expect(screen.getAllByTestId('theme-settings').length).toBeGreaterThan(0)
  })

  it('renders LanguageSwitcher component', () => {
    render(<AppHeader items={mockItems} />)
    expect(screen.getAllByTestId('language-switcher').length).toBeGreaterThan(0)
  })

  it('renders NotificationBell component', () => {
    render(<AppHeader items={mockItems} />)
    expect(screen.getAllByTestId('notification-bell').length).toBeGreaterThan(0)
  })

  it('renders navigation with correct accessibility attributes', () => {
    render(<AppHeader items={mockItems} />)
    const nav = screen.getByRole('navigation', { name: /main navigation/i })
    expect(nav).toBeInTheDocument()
  })

  it('renders navigation items list', () => {
    render(<AppHeader items={mockItems} />)
    const list = screen.getByRole('list')
    // The list is rendered inside the navigation
    expect(list).toBeInTheDocument()
  })

  it('renders navigation texts', () => {
    render(<AppHeader items={mockItems} />)
    // Text should be visible in the navigation
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Documents')).toBeInTheDocument()
  })

  it('renders navigation element', () => {
    render(<AppHeader items={mockItems} />)
    const nav = screen.getByRole('navigation')
    expect(nav).toBeInTheDocument()
  })

  it('renders as sticky header', () => {
    const { container } = render(<AppHeader items={mockItems} />)
    const header = container.querySelector('header')
    expect(header).toHaveClass('sticky', 'top-0')
  })
})
