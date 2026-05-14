import { render, screen } from '@testing-library/react'
import { BrandedHeader } from '../BrandedHeader'
import type { BrandSettings } from '@/types/branding'

const mockUseBranding = jest.fn()
jest.mock('@/hooks/useBranding', () => ({
  useBranding: (opts?: unknown) => mockUseBranding(opts),
}))

jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => `t:${key}`,
}))

jest.mock('next/image', () => ({
  __esModule: true,
  default: (props: Record<string, unknown>) => {
    // eslint-disable-next-line jsx-a11y/alt-text, @next/next/no-img-element
    return <img {...(props as { src: string; alt: string })} />
  },
}))

const seed: BrandSettings = {
  app_name: 'Loaded App',
  tagline: 'loaded tagline',
  logo_url: 'https://example.com/logo.png',
  favicon_url: '',
  primary_color: '#ff5733',
  secondary_color: '',
  updated_at: '2026-05-14T12:00:00Z',
}

beforeEach(() => {
  jest.clearAllMocks()
})

describe('BrandedHeader', () => {
  it('reads the public branding config', () => {
    mockUseBranding.mockReturnValue({ config: seed, isLoading: false })
    render(<BrandedHeader titleFallback="authPages.loginWelcome" />)
    expect(mockUseBranding).toHaveBeenCalledWith({ public: true })
  })

  it('renders the configured app name', () => {
    mockUseBranding.mockReturnValue({ config: seed, isLoading: false })
    render(<BrandedHeader titleFallback="authPages.loginWelcome" />)
    expect(screen.getByTestId('branded-title')).toHaveTextContent('Loaded App')
  })

  it('renders the logo when logo_url is set', () => {
    mockUseBranding.mockReturnValue({ config: seed, isLoading: false })
    render(<BrandedHeader titleFallback="authPages.loginWelcome" />)
    expect(screen.getByTestId('branded-logo')).toBeInTheDocument()
  })

  it('renders the tagline when set', () => {
    mockUseBranding.mockReturnValue({ config: seed, isLoading: false })
    render(<BrandedHeader titleFallback="authPages.loginWelcome" />)
    expect(screen.getByTestId('branded-tagline')).toHaveTextContent('loaded tagline')
  })

  it('omits the logo when logo_url is empty', () => {
    mockUseBranding.mockReturnValue({
      config: { ...seed, logo_url: '' },
      isLoading: false,
    })
    render(<BrandedHeader titleFallback="authPages.loginWelcome" />)
    expect(screen.queryByTestId('branded-logo')).not.toBeInTheDocument()
  })

  it('omits the tagline when tagline is empty', () => {
    mockUseBranding.mockReturnValue({
      config: { ...seed, tagline: '' },
      isLoading: false,
    })
    render(<BrandedHeader titleFallback="authPages.loginWelcome" />)
    expect(screen.queryByTestId('branded-tagline')).not.toBeInTheDocument()
  })

  it('falls back to the translated title key during loading', () => {
    mockUseBranding.mockReturnValue({ config: undefined, isLoading: true })
    render(<BrandedHeader titleFallback="authPages.loginWelcome" />)
    expect(screen.getByTestId('branded-title')).toHaveTextContent('t:authPages.loginWelcome')
  })

  it('falls back to the translated title key when fetch fails', () => {
    mockUseBranding.mockReturnValue({ config: undefined, isLoading: false })
    render(<BrandedHeader titleFallback="authPages.loginWelcome" />)
    expect(screen.getByTestId('branded-title')).toHaveTextContent('t:authPages.loginWelcome')
  })
})
