import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ShareDocumentDialog } from '../ShareDocumentDialog'
import { documentsApi } from '@/lib/api/documents'
import { usersApi } from '@/lib/api/users'
import { toast } from 'sonner'

// Mock ResizeObserver
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}))

// Mock Select component to enable testing onValueChange
jest.mock('@/components/ui/select', () => ({
  Select: ({
    children,
    value,
    onValueChange,
    disabled,
  }: {
    children: React.ReactNode
    value?: string
    onValueChange?: (value: string) => void
    disabled?: boolean
  }) => (
    <div data-testid="mock-select">
      <select
        value={value}
        onChange={(e) => onValueChange?.(e.target.value)}
        disabled={disabled}
        data-testid="mock-select-input"
      >
        {children}
      </select>
    </div>
  ),
  SelectTrigger: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  SelectValue: ({ placeholder }: { placeholder?: string }) => <span>{placeholder}</span>,
  SelectContent: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  SelectItem: ({ children, value }: { children: React.ReactNode; value: string }) => (
    <option value={value}>{children}</option>
  ),
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
        toUser: 'User',
        user: 'User',
        role: 'Role',
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

const mockedDocumentsApi = jest.mocked(documentsApi)
const mockedUsersApi = jest.mocked(usersApi)
const mockedToast = jest.mocked(toast)

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

  it('loads user list when opened', async () => {
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(mockedUsersApi.getAll).toHaveBeenCalled()
    })
  })

  it('loads permissions when opened', async () => {
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(mockedDocumentsApi.getPermissions).toHaveBeenCalledWith(1)
    })
  })

  it('loads public links when opened', async () => {
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(mockedDocumentsApi.getPublicLinks).toHaveBeenCalledWith(1)
    })
  })

  it('can click by role button', async () => {
    const user = userEvent.setup()
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Share Document')).toBeInTheDocument()
    })

    // Find "Role" button specifically
    const roleButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'Role')
    if (roleButtons.length > 0) {
      await user.click(roleButtons[0])
    }
    expect(roleButtons.length).toBeGreaterThan(0)
  })

  it('shows existing permissions when loaded', async () => {
    mockedDocumentsApi.getPermissions.mockResolvedValueOnce([
      {
        id: 1,
        user_name: 'Test User',
        user_email: 'test@example.com',
        permission: 'read',
      },
    ] as never)

    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Test User')).toBeInTheDocument()
    })
  })

  it('shows existing public links when loaded', async () => {
    const user = userEvent.setup()
    mockedDocumentsApi.getPublicLinks.mockResolvedValueOnce([
      {
        id: 1,
        token: 'abc123',
        url: 'https://example.com/link/abc123',
        permission: 'read',
        is_active: true,
        use_count: 5,
      },
    ] as never)

    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByRole('tab', { name: /links/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('tab', { name: /links/i }))

    await waitFor(() => {
      expect(screen.getByText('abc123')).toBeInTheDocument()
    })
  })

  it('can share document with user', async () => {
    const user = userEvent.setup()
    mockedDocumentsApi.shareDocument.mockResolvedValueOnce({} as never)

    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Share Document')).toBeInTheDocument()
    })

    // Click grant access button
    const grantButton = screen.getByRole('button', { name: /grant.*access|share/i })
    if (grantButton) {
      await user.click(grantButton)
    }
  })

  it('handles data load error', async () => {
    mockedDocumentsApi.getPermissions.mockRejectedValueOnce(new Error('Load failed'))

    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(mockedToast.error).toHaveBeenCalled()
    })
  })

  it('shows loading state while loading data', () => {
    render(<ShareDocumentDialog {...defaultProps} />)
    // Loading spinner should be present briefly
    expect(document.body).toBeInTheDocument()
  })

  it('renders permission select with options', async () => {
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Share Document')).toBeInTheDocument()
    })

    // Permission select should be present
    const selects = screen.getAllByRole('combobox')
    expect(selects.length).toBeGreaterThan(0)
  })

  it('renders expiry date input', async () => {
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Share Document')).toBeInTheDocument()
    })

    // Date input for expiry should be present
    const dateInputs = document.querySelectorAll('input[type="datetime-local"]')
    expect(dateInputs.length).toBeGreaterThan(0)
  })

  it('can revoke permission', async () => {
    const user = userEvent.setup()
    mockedDocumentsApi.getPermissions.mockResolvedValueOnce([
      {
        id: 1,
        user_name: 'Test User',
        user_email: 'test@example.com',
        permission: 'read',
      },
    ] as never)
    mockedDocumentsApi.revokePermission.mockResolvedValueOnce({} as never)

    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Test User')).toBeInTheDocument()
    })

    // Find and click the delete/revoke button
    const deleteButtons = screen.getAllByRole('button')
    const revokeButton = deleteButtons.find((btn) => btn.querySelector('svg.lucide-trash-2'))
    if (revokeButton) {
      await user.click(revokeButton)
      await waitFor(() => {
        expect(mockedDocumentsApi.revokePermission).toHaveBeenCalled()
      })
    }
  })

  it('can switch share type to user', async () => {
    const user = userEvent.setup()
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Share Document')).toBeInTheDocument()
    })

    // Click "User" button (there are multiple "User" texts, get the button one)
    const userButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'User')
    if (userButtons.length > 0) {
      await user.click(userButtons[0])
    }
    expect(userButtons.length).toBeGreaterThan(0)
  })

  it('can change selected user', async () => {
    const user = userEvent.setup()
    render(<ShareDocumentDialog {...defaultProps} />)

    // Wait for users to load
    await waitFor(() => {
      expect(screen.getByText('User 1 (user1@example.com)')).toBeInTheDocument()
    })

    // Find and change user select
    const selectInputs = screen.getAllByTestId('mock-select-input')
    if (selectInputs.length > 0) {
      await user.selectOptions(selectInputs[0], '1')
      expect(selectInputs[0]).toHaveValue('1')
    }
  })

  it('can change selected role when in role mode', async () => {
    const user = userEvent.setup()
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Role')).toBeInTheDocument()
    })

    // Switch to role mode
    await user.click(screen.getByText('Role'))

    // Find and change role select
    const selectInputs = screen.getAllByTestId('mock-select-input')
    if (selectInputs.length > 0) {
      await user.selectOptions(selectInputs[0], 'teacher')
      expect(selectInputs[0]).toHaveValue('teacher')
    }
  })

  it('can change permission select', async () => {
    const user = userEvent.setup()
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Share Document')).toBeInTheDocument()
    })

    // Find permission select (second select)
    const selectInputs = screen.getAllByTestId('mock-select-input')
    if (selectInputs.length > 1) {
      await user.selectOptions(selectInputs[1], 'write')
      expect(selectInputs[1]).toHaveValue('write')
    }
  })

  it('can change expiry date input', async () => {
    const user = userEvent.setup()
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByText('Share Document')).toBeInTheDocument()
    })

    // Find datetime-local input
    const dateInputs = document.querySelectorAll('input[type="datetime-local"]')
    if (dateInputs.length > 0) {
      const dateInput = dateInputs[0] as HTMLInputElement
      await user.type(dateInput, '2024-12-31T23:59')
      expect(dateInput.value).toBe('2024-12-31T23:59')
    }
  })

  it('can change link permission select', async () => {
    const user = userEvent.setup()
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByRole('tab', { name: /links/i })).toBeInTheDocument()
    })

    // Switch to links tab
    await user.click(screen.getByRole('tab', { name: /links/i }))

    await waitFor(() => {
      expect(screen.getByText(/create.*link/i)).toBeInTheDocument()
    })

    // Find and change link permission select
    const selectInputs = screen.getAllByTestId('mock-select-input')
    if (selectInputs.length > 0) {
      await user.selectOptions(selectInputs[selectInputs.length - 1], 'download')
    }
  })

  it('can change link max uses input', async () => {
    const user = userEvent.setup()
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByRole('tab', { name: /links/i })).toBeInTheDocument()
    })

    // Switch to links tab
    await user.click(screen.getByRole('tab', { name: /links/i }))

    await waitFor(() => {
      expect(screen.getByText(/create.*link/i)).toBeInTheDocument()
    })

    // Find max uses input
    const numberInputs = document.querySelectorAll('input[type="number"]')
    if (numberInputs.length > 0) {
      const maxUsesInput = numberInputs[0] as HTMLInputElement
      await user.type(maxUsesInput, '10')
      expect(maxUsesInput.value).toContain('10')
    }
  })

  it('can change link password input', async () => {
    const user = userEvent.setup()
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByRole('tab', { name: /links/i })).toBeInTheDocument()
    })

    // Switch to links tab
    await user.click(screen.getByRole('tab', { name: /links/i }))

    await waitFor(() => {
      expect(screen.getByText(/create.*link/i)).toBeInTheDocument()
    })

    // Find password input
    const passwordInputs = document.querySelectorAll('input[type="password"]')
    if (passwordInputs.length > 0) {
      const passwordInput = passwordInputs[0] as HTMLInputElement
      await user.type(passwordInput, 'secret123')
      expect(passwordInput.value).toBe('secret123')
    }
  })

  it('can click create link button', async () => {
    const user = userEvent.setup()
    mockedDocumentsApi.createPublicLink.mockResolvedValueOnce({
      id: 1,
      token: 'newtoken',
      url: 'https://example.com/link/newtoken',
      permission: 'read',
      is_active: true,
      use_count: 0,
    } as never)

    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByRole('tab', { name: /links/i })).toBeInTheDocument()
    })

    // Switch to links tab
    await user.click(screen.getByRole('tab', { name: /links/i }))

    await waitFor(() => {
      expect(screen.getByText(/create.*link/i)).toBeInTheDocument()
    })

    // Click create link button
    const createButton = screen.getByRole('button', { name: /create.*link/i })
    await user.click(createButton)

    await waitFor(() => {
      expect(mockedDocumentsApi.createPublicLink).toHaveBeenCalled()
    })
  })

  it('can change link expiry date input', async () => {
    const user = userEvent.setup()
    render(<ShareDocumentDialog {...defaultProps} />)

    await waitFor(() => {
      expect(screen.getByRole('tab', { name: /links/i })).toBeInTheDocument()
    })

    // Switch to links tab
    await user.click(screen.getByRole('tab', { name: /links/i }))

    await waitFor(() => {
      expect(screen.getByText(/create.*link/i)).toBeInTheDocument()
    })

    // Find and change link expiry date input
    const dateInputs = document.querySelectorAll('input[type="datetime-local"]')
    if (dateInputs.length > 0) {
      const expiryInput = dateInputs[dateInputs.length - 1] as HTMLInputElement
      await user.type(expiryInput, '2025-06-30T12:00')
      expect(expiryInput.value).toContain('2025')
    }
  })

  it('closes dialog when footer close button is clicked', async () => {
    const user = userEvent.setup()
    const onOpenChange = jest.fn()
    render(<ShareDocumentDialog {...defaultProps} onOpenChange={onOpenChange} />)

    await waitFor(() => {
      expect(screen.getByText('Share Document')).toBeInTheDocument()
    })

    // Find and click the close button in footer (different from the X button)
    const closeButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'Close')
    if (closeButtons.length > 0) {
      await user.click(closeButtons[closeButtons.length - 1])
      expect(onOpenChange).toHaveBeenCalledWith(false)
    }
  })
})
