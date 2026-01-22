import { render } from '@testing-library/react'
import { GrainGradientBackground } from '../GrainGradientBackground'

// Mock the @paper-design/shaders-react library
jest.mock('@paper-design/shaders-react', () => ({
  GrainGradient: jest.fn(
    ({ style, colorBack, colors, softness, intensity, noise, shape, speed }) => (
      <div
        data-testid="grain-gradient"
        data-color-back={colorBack}
        data-colors={JSON.stringify(colors)}
        data-softness={softness}
        data-intensity={intensity}
        data-noise={noise}
        data-shape={shape}
        data-speed={speed}
        style={style}
      />
    )
  ),
}))

describe('GrainGradientBackground', () => {
  it('renders with dark theme colors when isDark is true', () => {
    const { getByTestId } = render(
      <GrainGradientBackground isDark={true} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('grain-gradient')
    expect(gradient).toBeInTheDocument()
    expect(gradient).toHaveAttribute('data-color-back', 'hsl(222, 47%, 6%)')
  })

  it('renders with light theme colors when isDark is false', () => {
    const { getByTestId } = render(
      <GrainGradientBackground isDark={false} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('grain-gradient')
    expect(gradient).toBeInTheDocument()
    expect(gradient).toHaveAttribute('data-color-back', 'hsl(0, 0%, 100%)')
  })

  it('passes speed prop correctly', () => {
    const { getByTestId } = render(
      <GrainGradientBackground isDark={true} speed={1.5} intensity={0.8} />
    )

    const gradient = getByTestId('grain-gradient')
    expect(gradient).toHaveAttribute('data-speed', '1.5')
  })

  it('passes intensity prop correctly', () => {
    const { getByTestId } = render(
      <GrainGradientBackground isDark={true} speed={0.5} intensity={0.3} />
    )

    const gradient = getByTestId('grain-gradient')
    expect(gradient).toHaveAttribute('data-intensity', '0.3')
  })

  it('renders with full width and height style', () => {
    const { getByTestId } = render(
      <GrainGradientBackground isDark={true} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('grain-gradient')
    expect(gradient).toHaveStyle({ width: '100%', height: '100%' })
  })

  it('uses corners shape', () => {
    const { getByTestId } = render(
      <GrainGradientBackground isDark={true} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('grain-gradient')
    expect(gradient).toHaveAttribute('data-shape', 'corners')
  })

  it('has fixed softness value of 0.8', () => {
    const { getByTestId } = render(
      <GrainGradientBackground isDark={true} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('grain-gradient')
    expect(gradient).toHaveAttribute('data-softness', '0.8')
  })

  it('has fixed noise value of 0.03', () => {
    const { getByTestId } = render(
      <GrainGradientBackground isDark={true} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('grain-gradient')
    expect(gradient).toHaveAttribute('data-noise', '0.03')
  })

  it('uses different colors array for dark theme', () => {
    const { getByTestId } = render(
      <GrainGradientBackground isDark={true} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('grain-gradient')
    const colors = JSON.parse(gradient.getAttribute('data-colors') || '[]')
    expect(colors).toHaveLength(3)
    expect(colors[0]).toContain('hsl(250')
  })

  it('uses different colors array for light theme', () => {
    const { getByTestId } = render(
      <GrainGradientBackground isDark={false} speed={0.5} intensity={0.8} />
    )

    const gradient = getByTestId('grain-gradient')
    const colors = JSON.parse(gradient.getAttribute('data-colors') || '[]')
    expect(colors).toHaveLength(3)
    expect(colors[0]).toContain('hsl(210')
  })
})
