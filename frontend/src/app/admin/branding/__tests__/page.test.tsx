import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import AdminBrandingPage from '../page'
import type { BrandSettings } from '@/types/branding'

const mockReplace = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace }),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

const mockUseBranding = jest.fn()
const mockUseUpdateBranding = jest.fn()
const mockUpdateBranding = jest.fn()
const mockMutate = jest.fn()

jest.mock('@/hooks/useBranding', () => ({
  useBranding: (opts?: unknown) => mockUseBranding(opts),
  useUpdateBranding: () => mockUseUpdateBranding(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => key,
}))

const seed: BrandSettings = {
  app_name: 'Seeded App',
  tagline: 'seeded tagline',
  logo_url: 'https://example.com/logo.png',
  favicon_url: 'https://example.com/favicon.ico',
  primary_color: '#aabbcc',
  secondary_color: '#001122',
  updated_at: '2026-05-14T12:00:00Z',
}

beforeEach(() => {
  jest.clearAllMocks()
  mockUseAuthCheck.mockReturnValue({
    user: { id: 1, role: 'system_admin' as const },
    isAuthenticated: true,
    isLoading: false,
  })
  mockUseBranding.mockReturnValue({
    config: seed,
    isLoading: false,
    error: undefined,
    mutate: mockMutate,
  })
  mockUseUpdateBranding.mockReturnValue({
    updateBranding: mockUpdateBranding,
    isLoading: false,
    error: null,
    errorCode: null,
  })
})

describe('AdminBrandingPage — role guard', () => {
  it('redirects non-admin to /forbidden', async () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 1, role: 'methodist' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<AdminBrandingPage />)
    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/forbidden'))
  })

  it('renders the page for system_admin', () => {
    render(<AdminBrandingPage />)
    expect(mockReplace).not.toHaveBeenCalled()
    expect(screen.getByTestId('admin-branding-page')).toBeInTheDocument()
    expect(screen.getByTestId('branding-form')).toBeInTheDocument()
  })
})

describe('AdminBrandingPage — form', () => {
  it('populates all 6 fields from the loaded config', () => {
    render(<AdminBrandingPage />)
    expect(screen.getByTestId('branding-input-app-name')).toHaveValue('Seeded App')
    expect(screen.getByTestId('branding-input-tagline')).toHaveValue('seeded tagline')
    expect(screen.getByTestId('branding-input-logo-url')).toHaveValue(
      'https://example.com/logo.png'
    )
    expect(screen.getByTestId('branding-input-favicon-url')).toHaveValue(
      'https://example.com/favicon.ico'
    )
    expect(screen.getByTestId('branding-input-primary-color')).toHaveValue('#aabbcc')
    expect(screen.getByTestId('branding-input-secondary-color')).toHaveValue('#001122')
  })

  it('renders native color pickers next to each hex textbox', () => {
    render(<AdminBrandingPage />)
    expect(screen.getByTestId('branding-picker-primary-color')).toBeInTheDocument()
    expect(screen.getByTestId('branding-picker-secondary-color')).toBeInTheDocument()
  })

  it('submits the form and renders success banner on 200', async () => {
    mockUpdateBranding.mockResolvedValueOnce(seed)
    render(<AdminBrandingPage />)

    fireEvent.change(screen.getByTestId('branding-input-app-name'), {
      target: { value: 'New Name' },
    })
    fireEvent.submit(screen.getByTestId('branding-form'))

    await waitFor(() =>
      expect(mockUpdateBranding).toHaveBeenCalledWith(
        expect.objectContaining({ app_name: 'New Name' })
      )
    )
    await waitFor(() => expect(screen.getByTestId('branding-success')).toBeInTheDocument())
  })

  it('renders the typed error banner on 422 INVALID_COLOR', () => {
    mockUseUpdateBranding.mockReturnValue({
      updateBranding: mockUpdateBranding,
      isLoading: false,
      error: new Error('boom'),
      errorCode: 'INVALID_COLOR',
    })
    render(<AdminBrandingPage />)
    expect(screen.getByTestId('branding-save-error')).toBeInTheDocument()
    expect(screen.getByTestId('branding-save-error')).toHaveTextContent('errors.INVALID_COLOR')
  })

  it('disables submit while saving', () => {
    mockUseUpdateBranding.mockReturnValue({
      updateBranding: mockUpdateBranding,
      isLoading: true,
      error: null,
      errorCode: null,
    })
    render(<AdminBrandingPage />)
    expect(screen.getByTestId('branding-submit')).toBeDisabled()
  })
})

describe('AdminBrandingPage — loading / error', () => {
  it('renders the loading spinner', () => {
    mockUseBranding.mockReturnValue({
      config: undefined,
      isLoading: true,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminBrandingPage />)
    expect(screen.getByTestId('branding-loading')).toBeInTheDocument()
  })

  it('renders the error state on fetch failure', () => {
    mockUseBranding.mockReturnValue({
      config: undefined,
      isLoading: false,
      error: new Error('boom'),
      mutate: jest.fn(),
    })
    render(<AdminBrandingPage />)
    expect(screen.getByTestId('branding-error')).toBeInTheDocument()
  })
})
