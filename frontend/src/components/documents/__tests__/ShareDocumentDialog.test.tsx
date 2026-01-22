import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ShareDocumentDialog } from '../ShareDocumentDialog'

// Mock ResizeObserver
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}))

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: (namespace: string) => (key: string, _params?: Record<string, unknown>) => {
    const translations: Record<string, Record<string, string>> = {
      'documents.share': {
        title: 'Share Document',
        usersTab: 'Users',
        linksTab: 'Links',
        shareWith: 'Share with',
        selectUser: 'Select user',
        selectRole: 'Select role',
        permission: 'Permission',
        permissionRead: 'View',
        permissionWrite: 'Edit',
        permissionDelete: 'Delete',
        permissionAdmin: 'Admin',
        roleAdmin: 'Admin',
        roleSecretary: 'Secretary',
        roleMethodist: 'Methodist',
        roleTeacher: 'Teacher',
        roleStudent: 'Student',
        share: 'Share',
        cancel: 'Cancel',
        currentPermissions: 'Current permissions',
        noPermissions: 'Not shared with anyone',
        publicLinks: 'Public links',
        createLink: 'Create link',
        linkCopied: 'Link copied',
        expiresAt: 'Expires at',
        maxUses: 'Max uses',
        password: 'Password',
        saving: 'Saving...',
        accessGranted: 'Access granted',
        accessRevoked: 'Access revoked',
        grantError: 'Error granting access',
        revokeError: 'Error revoking access',
        dataLoadError: 'Error loading data',
        linkCreated: 'Link created',
        linkCreateError: 'Error creating link',
        linkCopyError: 'Error copying link',
        linkDeactivated: 'Link deactivated',
        linkDeactivateError: 'Error deactivating link',
        linkDeleted: 'Link deleted',
        linkDeleteError: 'Error deleting link',
        byUser: 'User',
        byRole: 'Role',
      },
      'documents.form': {
        expiresAt: 'Expires at',
      },
      common: {
        cancel: 'Cancel',
        save: 'Save',
        close: 'Close',
        loading: 'Loading...',
      },
    }
    return translations[namespace]?.[key] || key
  },
  useLocale: () => 'en',
}))

// Mock the APIs
jest.mock('@/lib/api/documents', () => ({
  documentsApi: {
    getPermissions: jest.fn().mockResolvedValue([]),
    shareDocument: jest.fn(),
    revokePermission: jest.fn(),
    getPublicLinks: jest.fn().mockResolvedValue([]),
    createPublicLink: jest.fn(),
    deactivatePublicLink: jest.fn(),
    deletePublicLink: jest.fn(),
  },
}))

jest.mock('@/lib/api/users', () => ({
  usersApi: {
    getAll: jest.fn().mockResolvedValue([
      { id: 1, name: 'User 1', email: 'user1@example.com' },
      { id: 2, name: 'User 2', email: 'user2@example.com' },
    ]),
  },
}))

// Mock sonner toast
jest.mock('sonner', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

describe('ShareDocumentDialog', () => {
  const defaultProps = {
    open: true,
    onOpenChange: jest.fn(),
    documentId: 1,
    documentTitle: 'Test Document.pdf',
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders share dialog when open', async () => {
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Share Document')).toBeInTheDocument()
    })
  })

  it('displays document title', async () => {
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Test Document.pdf')).toBeInTheDocument()
    })
  })

  it('renders share tabs', async () => {
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByRole('tab', { name: /users/i })).toBeInTheDocument()
      expect(screen.getByRole('tab', { name: /links/i })).toBeInTheDocument()
    })
  })

  it('does not render when open is false', () => {
    render(<ShareDocumentDialog {...defaultProps} open={false} />)
    expect(screen.queryByText('Share Document')).not.toBeInTheDocument()
  })

  it('calls onOpenChange when dialog is closed', async () => {
    const user = userEvent.setup()
    const onOpenChange = jest.fn()
    render(<ShareDocumentDialog {...defaultProps} onOpenChange={onOpenChange} />)

    await waitFor(() => {
      expect(screen.getByText('Share Document')).toBeInTheDocument()
    })

    // Find and click the close button
    const closeButtons = screen.getAllByRole('button')
    const closeButton = closeButtons.find(
      (btn) =>
        btn.getAttribute('aria-label')?.includes('Close') || btn.querySelector('svg.lucide-x')
    )
    if (closeButton) {
      await user.click(closeButton)
      expect(onOpenChange).toHaveBeenCalledWith(false)
    }
  })

  it('can switch to links tab', async () => {
    const user = userEvent.setup()
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByRole('tab', { name: /links/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('tab', { name: /links/i }))

    await waitFor(() => {
      expect(screen.getByText(/create.*link/i)).toBeInTheDocument()
    })
  })
})
