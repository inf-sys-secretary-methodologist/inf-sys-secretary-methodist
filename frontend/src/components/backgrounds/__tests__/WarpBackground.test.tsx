import { render } from '@testing-library/react'
import { WarpBackground } from '../WarpBackground'

// Mock the @paper-design/shaders-react library
jest.mock('@paper-design/shaders-react', () => ({
  Warp: jest.fn(
    ({
      style,
      proportion,
      softness,
      distortion,
      swirl,
      swirlIterations,
      shape,
      shapeScale,
      scale,
      rotation,
      speed,
      colors,
    }) => (
      <div
        data-testid="warp-gradient"
        data-proportion={proportion}
        data-softness={softness}
        data-distortion={distortion}
        data-swirl={swirl}
        data-swirl-iterations={swirlIterations}
        data-shape={shape}
        data-shape-scale={shapeScale}
        data-scale={scale}
        data-rotation={rotation}
        data-speed={speed}
        data-colors={JSON.stringify(colors)}
        style={style}
      />
    )
  ),
}))

describe('WarpBackground', () => {
  it('renders with dark theme colors when isDark is true', () => {
    const { getByTestId } = render(<WarpBackground isDark={true} speed={0.5} intensity={0.8} />)

    const warp = getByTestId('warp-gradient')
    expect(warp).toBeInTheDocument()

    const colors = JSON.parse(warp.getAttribute('data-colors') || '[]')
    expect(colors).toHaveLength(4)
    expect(colors[0]).toContain('hsl(250')
  })

  it('renders with light theme colors when isDark is false', () => {
    const { getByTestId } = render(<WarpBackground isDark={false} speed={0.5} intensity={0.8} />)

    const warp = getByTestId('warp-gradient')
    const colors = JSON.parse(warp.getAttribute('data-colors') || '[]')
    expect(colors).toHaveLength(4)
    expect(colors[0]).toContain('hsl(210')
  })

  it('passes speed prop correctly', () => {
    const { getByTestId } = render(<WarpBackground isDark={true} speed={2.0} intensity={0.5} />)

    const warp = getByTestId('warp-gradient')
    expect(warp).toHaveAttribute('data-speed', '2')
  })

  it('scales intensity by 0.5 for distortion', () => {
    const { getByTestId } = render(<WarpBackground isDark={true} speed={0.5} intensity={0.8} />)

    const warp = getByTestId('warp-gradient')
    // intensity * 0.5 = 0.8 * 0.5 = 0.4
    expect(warp).toHaveAttribute('data-distortion', '0.4')
  })

  it('renders with full width and height style', () => {
    const { getByTestId } = render(<WarpBackground isDark={true} speed={0.5} intensity={0.8} />)

    const warp = getByTestId('warp-gradient')
    expect(warp).toHaveStyle({ width: '100%', height: '100%' })
  })

  it('uses checks shape', () => {
    const { getByTestId } = render(<WarpBackground isDark={true} speed={0.5} intensity={0.8} />)

    const warp = getByTestId('warp-gradient')
    expect(warp).toHaveAttribute('data-shape', 'checks')
  })

  it('has fixed proportion value of 0.45', () => {
    const { getByTestId } = render(<WarpBackground isDark={true} speed={0.5} intensity={0.8} />)

    const warp = getByTestId('warp-gradient')
    expect(warp).toHaveAttribute('data-proportion', '0.45')
  })

  it('has fixed softness value of 1', () => {
    const { getByTestId } = render(<WarpBackground isDark={true} speed={0.5} intensity={0.8} />)

    const warp = getByTestId('warp-gradient')
    expect(warp).toHaveAttribute('data-softness', '1')
  })

  it('has fixed swirl value of 0.6', () => {
    const { getByTestId } = render(<WarpBackground isDark={true} speed={0.5} intensity={0.8} />)

    const warp = getByTestId('warp-gradient')
    expect(warp).toHaveAttribute('data-swirl', '0.6')
  })

  it('has fixed swirl iterations of 8', () => {
    const { getByTestId } = render(<WarpBackground isDark={true} speed={0.5} intensity={0.8} />)

    const warp = getByTestId('warp-gradient')
    expect(warp).toHaveAttribute('data-swirl-iterations', '8')
  })

  it('has fixed shape scale of 0.15', () => {
    const { getByTestId } = render(<WarpBackground isDark={true} speed={0.5} intensity={0.8} />)

    const warp = getByTestId('warp-gradient')
    expect(warp).toHaveAttribute('data-shape-scale', '0.15')
  })

  it('has scale of 1 and rotation of 0', () => {
    const { getByTestId } = render(<WarpBackground isDark={true} speed={0.5} intensity={0.8} />)

    const warp = getByTestId('warp-gradient')
    expect(warp).toHaveAttribute('data-scale', '1')
    expect(warp).toHaveAttribute('data-rotation', '0')
  })
})
