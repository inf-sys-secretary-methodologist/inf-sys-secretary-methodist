import { render, screen } from '@testing-library/react'
import { BackgroundProvider } from '../BackgroundProvider'
import type { BackgroundType } from '@/stores/appearanceStore'

// Mock next-themes
jest.mock('next-themes', () => ({
  useTheme: () => ({ resolvedTheme: 'light' }),
}))

// Mock appearance store
const mockAppearanceStore = {
  background: {
    enabled: true,
    type: 'grain-gradient' as BackgroundType,
    speed: 1,
    intensity: 0.5,
  },
  reducedMotion: false,
}

jest.mock('@/stores/appearanceStore', () => ({
  useAppearanceStore: (selector: (state: typeof mockAppearanceStore) => unknown) =>
    selector(mockAppearanceStore),
  useAppearanceHydrated: () => true,
}))

// Mock the background components
jest.mock('../GrainGradientBackground', () => ({
  GrainGradientBackground: ({
    isDark,
    speed,
    intensity,
  }: {
    isDark: boolean
    speed: number
    intensity: number
  }) => (
    <div
      data-testid="grain-gradient"
      data-isdark={isDark}
      data-speed={speed}
      data-intensity={intensity}
    />
  ),
}))

jest.mock('../WarpBackground', () => ({
  WarpBackground: ({
    isDark,
    speed,
    intensity,
  }: {
    isDark: boolean
    speed: number
    intensity: number
  }) => (
    <div data-testid="warp" data-isdark={isDark} data-speed={speed} data-intensity={intensity} />
  ),
}))

jest.mock('../MeshGradientBackground', () => ({
  MeshGradientBackground: ({
    isDark,
    speed,
    intensity,
  }: {
    isDark: boolean
    speed: number
    intensity: number
  }) => (
    <div
      data-testid="mesh-gradient"
      data-isdark={isDark}
      data-speed={speed}
      data-intensity={intensity}
    />
  ),
}))

describe('BackgroundProvider', () => {
  beforeEach(() => {
    // Reset the mock state
    mockAppearanceStore.background = {
      enabled: true,
      type: 'grain-gradient' as BackgroundType,
      speed: 1,
      intensity: 0.5,
    }
    mockAppearanceStore.reducedMotion = false
  })

  it('renders grain-gradient background by default', () => {
    render(<BackgroundProvider />)
    expect(screen.getByTestId('grain-gradient')).toBeInTheDocument()
  })

  it('passes correct props to background component', () => {
    render(<BackgroundProvider />)
    const bg = screen.getByTestId('grain-gradient')
    expect(bg).toHaveAttribute('data-isdark', 'false')
    expect(bg).toHaveAttribute('data-speed', '1')
    expect(bg).toHaveAttribute('data-intensity', '0.5')
  })

  it('renders warp background when type is warp', () => {
    mockAppearanceStore.background.type = 'warp' as BackgroundType
    render(<BackgroundProvider />)
    expect(screen.getByTestId('warp')).toBeInTheDocument()
  })

  it('renders mesh-gradient background when type is mesh-gradient', () => {
    mockAppearanceStore.background.type = 'mesh-gradient' as BackgroundType
    render(<BackgroundProvider />)
    expect(screen.getByTestId('mesh-gradient')).toBeInTheDocument()
  })

  it('returns null when background is disabled', () => {
    mockAppearanceStore.background.enabled = false
    const { container } = render(<BackgroundProvider />)
    expect(container.firstChild).toBeNull()
  })

  it('returns null when background type is none', () => {
    mockAppearanceStore.background.type = 'none' as BackgroundType
    const { container } = render(<BackgroundProvider />)
    expect(container.firstChild).toBeNull()
  })

  it('sets speed to 0 when reducedMotion is enabled', () => {
    mockAppearanceStore.reducedMotion = true
    render(<BackgroundProvider />)
    const bg = screen.getByTestId('grain-gradient')
    expect(bg).toHaveAttribute('data-speed', '0')
  })

  it('renders with fixed positioning', () => {
    const { container } = render(<BackgroundProvider />)
    expect(container.firstChild).toHaveClass('fixed', 'inset-0')
  })
})
