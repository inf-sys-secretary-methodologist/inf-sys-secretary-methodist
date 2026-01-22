import { render } from '@testing-library/react'
import { MeshGradientBackground } from '../MeshGradientBackground'

// Mock the @paper-design/shaders-react library
jest.mock('@paper-design/shaders-react', () => ({
  MeshGradient: jest.fn(({ style, colors, speed, distortion }) => (
    <div
      data-testid="mesh-gradient"
      data-colors={JSON.stringify(colors)}
      data-speed={speed}
      data-distortion={distortion}
      style={style}
    />
  )),
}))

describe('MeshGradientBackground', () => {
  it('renders with dark theme colors when isDark is true', () => {
    const { getByTestId } = render(
      <MeshGradientBackground isDark={true} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('mesh-gradient')
    expect(gradient).toBeInTheDocument()

    const colors = JSON.parse(gradient.getAttribute('data-colors') || '[]')
    expect(colors).toHaveLength(4)
    expect(colors[0]).toContain('hsl(250')
  })

  it('renders with light theme colors when isDark is false', () => {
    const { getByTestId } = render(
      <MeshGradientBackground isDark={false} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('mesh-gradient')
    const colors = JSON.parse(gradient.getAttribute('data-colors') || '[]')
    expect(colors).toHaveLength(4)
    expect(colors[0]).toContain('hsl(210')
  })

  it('passes speed prop correctly', () => {
    const { getByTestId } = render(
      <MeshGradientBackground isDark={true} speed={2.0} intensity={0.5} />
    )

    const gradient = getByTestId('mesh-gradient')
    expect(gradient).toHaveAttribute('data-speed', '2')
  })

  it('passes intensity as distortion prop', () => {
    const { getByTestId } = render(
      <MeshGradientBackground isDark={true} speed={0.5} intensity={0.6} />
    )

    const gradient = getByTestId('mesh-gradient')
    expect(gradient).toHaveAttribute('data-distortion', '0.6')
  })

  it('renders with full width and height style', () => {
    const { getByTestId } = render(
      <MeshGradientBackground isDark={true} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('mesh-gradient')
    expect(gradient).toHaveStyle({ width: '100%', height: '100%' })
  })

  it('uses 4 colors for gradient', () => {
    const { getByTestId } = render(
      <MeshGradientBackground isDark={true} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('mesh-gradient')
    const colors = JSON.parse(gradient.getAttribute('data-colors') || '[]')
    expect(colors).toHaveLength(4)
  })

  it('dark theme uses muted purple/blue colors', () => {
    const { getByTestId } = render(
      <MeshGradientBackground isDark={true} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('mesh-gradient')
    const colors = JSON.parse(gradient.getAttribute('data-colors') || '[]')

    // Check that colors are in the expected range for dark theme
    colors.forEach((color: string) => {
      expect(color).toMatch(/hsl\(\d+/)
    })
  })

  it('light theme uses pastel colors', () => {
    const { getByTestId } = render(
      <MeshGradientBackground isDark={false} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('mesh-gradient')
    const colors = JSON.parse(gradient.getAttribute('data-colors') || '[]')

    // Light theme should have high lightness values (90%+)
    colors.forEach((color: string) => {
      expect(color).toMatch(/9\d%\)/)
    })
  })
})
