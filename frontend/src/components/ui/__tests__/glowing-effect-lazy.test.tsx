import { GlowingEffect } from '../glowing-effect-lazy'

// Mock next/dynamic
jest.mock('next/dynamic', () => {
  return function mockDynamic(
    _importFn: () => Promise<{ GlowingEffect: React.ComponentType }>,
    _options?: { ssr?: boolean; loading?: () => null }
  ) {
    // Return a mock component
    const MockGlowingEffect = () => null
    MockGlowingEffect.displayName = 'GlowingEffect'
    return MockGlowingEffect
  }
})

describe('GlowingEffect (lazy)', () => {
  it('exports GlowingEffect component', () => {
    expect(GlowingEffect).toBeDefined()
  })

  it('GlowingEffect is a function component', () => {
    expect(typeof GlowingEffect).toBe('function')
  })
})
