import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MobileNav } from '../MobileNav'
import type { NavEntry, NavGroup } from '@/config/navigation'
import { Home, FileText, Calendar, Settings, Users, FolderOpen } from 'lucide-react'

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
  const mockEntries: NavEntry[] = [
    { url: '/dashboard', nameKey: 'dashboard', icon: Home },
    { url: '/documents', nameKey: 'documents', icon: FileText },
    { url: '/calendar', nameKey: 'calendar', icon: Calendar },
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

  it('renders menu button', () => {
    render(<MobileNav entries={mockEntries} />)
    expect(screen.getByRole('button', { name: /open menu/i })).toBeInTheDocument()
  })

  it('renders sheet component', () => {
    render(<MobileNav entries={mockEntries} />)
    expect(screen.getByTestId('sheet')).toBeInTheDocument()
  })

  it('renders sheet content', () => {
    render(<MobileNav entries={mockEntries} />)
    expect(screen.getByTestId('sheet-content')).toBeInTheDocument()
  })

  it('renders navigation title', () => {
    render(<MobileNav entries={mockEntries} />)
    expect(screen.getByText('Navigation')).toBeInTheDocument()
  })

  it('renders navigation items', () => {
    render(<MobileNav entries={mockEntries} />)
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Documents')).toBeInTheDocument()
    expect(screen.getByText('Calendar')).toBeInTheDocument()
  })

  it('renders close button', () => {
    render(<MobileNav entries={mockEntries} />)
    expect(screen.getByTestId('sheet-close')).toBeInTheDocument()
  })

  it('renders with mobile navigation aria-label', () => {
    render(<MobileNav entries={mockEntries} />)
    const nav = screen.getByRole('navigation', { name: /mobile navigation/i })
    expect(nav).toBeInTheDocument()
  })

  it('renders navigation links', () => {
    render(<MobileNav entries={mockEntries} />)
    const links = screen.getAllByRole('link')
    expect(links.length).toBe(3)
  })

  it('marks active link with aria-current', () => {
    render(<MobileNav entries={mockEntries} />)
    const links = screen.getAllByRole('link')
    const dashboardLink = links.find((link) => link.textContent?.includes('Dashboard'))
    expect(dashboardLink).toHaveAttribute('aria-current', 'page')
  })

  describe('NavGroup support', () => {
    it('renders NavGroup as collapsible', () => {
      render(<MobileNav entries={mockEntriesWithGroup} />)
      expect(screen.getByText('Admin')).toBeInTheDocument()
    })

    it('expands group to show items when clicked', async () => {
      render(<MobileNav entries={mockEntriesWithGroup} />)

      const adminTrigger = screen.getByText('Admin')
      fireEvent.click(adminTrigger)

      await waitFor(() => {
        expect(screen.getByText('Users')).toBeInTheDocument()
        expect(screen.getByText('Files')).toBeInTheDocument()
      })
    })

    it('auto-expands group when child item is active', () => {
      mockPathname = '/admin/users'
      render(<MobileNav entries={mockEntriesWithGroup} />)
      // Group should be expanded because of active child
      expect(screen.getByText('Users')).toBeInTheDocument()
    })

    it('renders chevron icon that rotates when expanded', async () => {
      const { container } = render(<MobileNav entries={mockEntriesWithGroup} />)

      // Initially not expanded (unless auto-expanded due to active)
      const adminTrigger = screen.getByText('Admin')
      fireEvent.click(adminTrigger)

      await waitFor(() => {
        const chevron = container.querySelector('.lucide-chevron-down')
        expect(chevron).toBeInTheDocument()
      })
    })

    it('toggles group expansion on click', async () => {
      render(<MobileNav entries={mockEntriesWithGroup} />)

      const adminTrigger = screen.getByText('Admin')

      // First click - expand
      fireEvent.click(adminTrigger)
      await waitFor(() => {
        expect(screen.getByText('Users')).toBeInTheDocument()
      })

      // Second click - collapse
      fireEvent.click(adminTrigger)
      await waitFor(() => {
        expect(screen.queryByText('Users')).not.toBeInTheDocument()
      })
    })

    it('marks nested link as active', () => {
      mockPathname = '/admin/users'
      render(<MobileNav entries={mockEntriesWithGroup} />)

      const usersLink = screen.getByText('Users').closest('a')
      expect(usersLink).toHaveAttribute('aria-current', 'page')
    })

    it('renders nested items with indentation', () => {
      mockPathname = '/admin/users'
      render(<MobileNav entries={mockEntriesWithGroup} />)

      const usersLink = screen.getByText('Users').closest('a')
      expect(usersLink).toHaveClass('ml-4')
    })
  })
})
