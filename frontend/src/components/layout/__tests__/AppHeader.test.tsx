import { render, screen } from '@testing-library/react'
import { AppHeader } from '../AppHeader'
import type { NavEntry, NavGroup } from '@/config/navigation'
import { FileText, Calendar, MessageCircle, Home, Settings, Users, FolderOpen } from 'lucide-react'

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
      adminGroup: 'Admin',
      users: 'Users',
      files: 'Files',
    }
    return translations[key] || key
  },
}))

// Mock next/navigation
let mockPathname = '/dashboard'
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
  MobileNav: ({ entries }: { entries: NavEntry[] }) => (
    <div data-testid="mobile-nav" data-items={entries.length}>
      Mobile Nav
    </div>
  ),
}))

describe('AppHeader', () => {
  const mockEntries: NavEntry[] = [
    { url: '/dashboard', nameKey: 'dashboard', icon: Home },
    { url: '/documents', nameKey: 'documents', icon: FileText },
    { url: '/calendar', nameKey: 'calendar', icon: Calendar },
    { url: '/messages', nameKey: 'messages', icon: MessageCircle },
    { url: '/settings', nameKey: 'settings', icon: Settings },
  ]

  const mockEntriesWithGroup: NavEntry[] = [
    { url: '/dashboard', nameKey: 'dashboard', icon: Home },
    {
      nameKey: 'adminGroup',
      icon: Settings,
      items: [
        { url: '/admin/users', nameKey: 'users', icon: Users },
        { url: '/admin/files', nameKey: 'files', icon: FolderOpen },
      ],
    } as NavGroup,
  ]

  beforeEach(() => {
    mockPathname = '/dashboard'
  })

  it('renders navigation items', () => {
    render(<AppHeader entries={mockEntries} />)
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Documents')).toBeInTheDocument()
    expect(screen.getByText('Calendar')).toBeInTheDocument()
  })

  it('renders MobileNav component', () => {
    render(<AppHeader entries={mockEntries} />)
    expect(screen.getByTestId('mobile-nav')).toBeInTheDocument()
    expect(screen.getByTestId('mobile-nav')).toHaveAttribute('data-items', '5')
  })

  it('renders UserMenu component', () => {
    render(<AppHeader entries={mockEntries} />)
    expect(screen.getAllByTestId('user-menu').length).toBeGreaterThan(0)
  })

  it('renders ThemeSettingsPopover component', () => {
    render(<AppHeader entries={mockEntries} />)
    expect(screen.getAllByTestId('theme-settings').length).toBeGreaterThan(0)
  })

  it('renders LanguageSwitcher component', () => {
    render(<AppHeader entries={mockEntries} />)
    expect(screen.getAllByTestId('language-switcher').length).toBeGreaterThan(0)
  })

  it('renders NotificationBell component', () => {
    render(<AppHeader entries={mockEntries} />)
    expect(screen.getAllByTestId('notification-bell').length).toBeGreaterThan(0)
  })

  it('renders navigation with correct accessibility attributes', () => {
    render(<AppHeader entries={mockEntries} />)
    const nav = screen.getByRole('navigation', { name: /main navigation/i })
    expect(nav).toBeInTheDocument()
  })

  it('renders navigation items list', () => {
    render(<AppHeader entries={mockEntries} />)
    const list = screen.getByRole('list')
    // The list is rendered inside the navigation
    expect(list).toBeInTheDocument()
  })

  it('renders navigation texts', () => {
    render(<AppHeader entries={mockEntries} />)
    // Text should be visible in the navigation
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Documents')).toBeInTheDocument()
  })

  it('renders navigation element', () => {
    render(<AppHeader entries={mockEntries} />)
    const nav = screen.getByRole('navigation')
    expect(nav).toBeInTheDocument()
  })

  it('renders as sticky header', () => {
    const { container } = render(<AppHeader entries={mockEntries} />)
    const header = container.querySelector('header')
    expect(header).toHaveClass('sticky', 'top-0')
  })

  describe('NavGroup support', () => {
    it('renders NavGroup trigger text', () => {
      render(<AppHeader entries={mockEntriesWithGroup} />)
      // NavGroup should be rendered with its name
      expect(screen.getByText('Admin')).toBeInTheDocument()
    })

    it('highlights group when child item is active', () => {
      mockPathname = '/admin/users'
      render(<AppHeader entries={mockEntriesWithGroup} />)
      // The group trigger should be visible and styled as active
      expect(screen.getByText('Admin')).toBeInTheDocument()
    })

    it('renders group with chevron icon', () => {
      const { container } = render(<AppHeader entries={mockEntriesWithGroup} />)
      // ChevronDown should be rendered inside the group trigger
      const chevrons = container.querySelectorAll('.lucide-chevron-down')
      expect(chevrons.length).toBeGreaterThan(0)
    })

    it('passes entries with group to MobileNav', () => {
      render(<AppHeader entries={mockEntriesWithGroup} />)
      // MobileNav should receive the entries including the group
      expect(screen.getByTestId('mobile-nav')).toHaveAttribute('data-items', '2')
    })
  })
})
