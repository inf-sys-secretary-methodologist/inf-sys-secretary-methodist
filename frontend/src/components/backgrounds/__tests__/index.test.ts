// Mock the shader library before imports
jest.mock('@paper-design/shaders-react', () => ({
  GrainGradient: jest.fn(() => null),
  MeshGradient: jest.fn(() => null),
  Warp: jest.fn(() => null),
}))

// Test that all exports are available from the index
import {
  BackgroundProvider,
  GrainGradientBackground,
  WarpBackground,
  MeshGradientBackground,
} from '../index'

describe('backgrounds/index exports', () => {
  it('exports BackgroundProvider', () => {
    expect(BackgroundProvider).toBeDefined()
  })

  it('exports GrainGradientBackground', () => {
    expect(GrainGradientBackground).toBeDefined()
  })

  it('exports WarpBackground', () => {
    expect(WarpBackground).toBeDefined()
  })

  it('exports MeshGradientBackground', () => {
    expect(MeshGradientBackground).toBeDefined()
  })
})
