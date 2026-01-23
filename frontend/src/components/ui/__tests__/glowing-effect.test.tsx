import { render, act } from '@testing-library/react'
import { GlowingEffect } from '../glowing-effect'

// Mock motion/react
jest.mock('motion/react', () => ({
  animate: jest.fn((from, to, options) => {
    // Call onUpdate immediately with final value
    if (options?.onUpdate) {
      options.onUpdate(to)
    }
    return { stop: jest.fn() }
  }),
}))

describe('GlowingEffect', () => {
  let requestAnimationFrameSpy: jest.SpyInstance
  let cancelAnimationFrameSpy: jest.SpyInstance

  beforeEach(() => {
    requestAnimationFrameSpy = jest
      .spyOn(window, 'requestAnimationFrame')
      .mockImplementation((cb) => {
        cb(0)
        return 1
      })
    cancelAnimationFrameSpy = jest.spyOn(window, 'cancelAnimationFrame').mockImplementation()
  })

  afterEach(() => {
    requestAnimationFrameSpy.mockRestore()
    cancelAnimationFrameSpy.mockRestore()
  })

  it('renders without crashing', () => {
    const { container } = render(<GlowingEffect />)
    expect(container).toBeInTheDocument()
  })

  it('renders two div elements', () => {
    const { container } = render(<GlowingEffect />)
    const divs = container.querySelectorAll('div')
    expect(divs.length).toBeGreaterThanOrEqual(2)
  })

  it('applies custom className', () => {
    const { container } = render(<GlowingEffect className="custom-glow" />)
    expect(container.querySelector('.custom-glow')).toBeInTheDocument()
  })

  it('sets blur CSS variable', () => {
    const { container } = render(<GlowingEffect blur={10} />)
    const element = container.querySelector('[style*="--blur"]')
    expect(element).toHaveStyle('--blur: 10px')
  })

  it('sets spread CSS variable', () => {
    const { container } = render(<GlowingEffect spread={30} />)
    const element = container.querySelector('[style*="--spread"]')
    expect(element).toHaveStyle('--spread: 30')
  })

  it('sets border width CSS variable', () => {
    const { container } = render(<GlowingEffect borderWidth={2} />)
    const element = container.querySelector('[style*="--glowingeffect-border-width"]')
    expect(element).toHaveStyle('--glowingeffect-border-width: 2px')
  })

  it('uses default variant gradient', () => {
    const { container } = render(<GlowingEffect />)
    const element = container.querySelector('[style*="--gradient"]')
    expect(element?.getAttribute('style')).toContain('radial-gradient')
  })

  it('uses white variant when specified', () => {
    const { container } = render(<GlowingEffect variant="white" />)
    const element = container.querySelector('[style*="--gradient"]')
    expect(element?.getAttribute('style')).toContain('repeating-conic-gradient')
  })

  it('applies glow class when glow prop is true', () => {
    const { container } = render(<GlowingEffect glow />)
    expect(container.querySelector('.opacity-100')).toBeInTheDocument()
  })

  it('is disabled by default', () => {
    const { container } = render(<GlowingEffect />)
    // When disabled, one div should have !block and another should have !hidden
    expect(container.innerHTML).toContain('hidden')
  })

  it('adds event listeners when not disabled', () => {
    const addEventListenerSpy = jest.spyOn(window, 'addEventListener')
    const bodyAddEventListenerSpy = jest.spyOn(document.body, 'addEventListener')

    render(<GlowingEffect disabled={false} />)

    expect(addEventListenerSpy).toHaveBeenCalledWith('scroll', expect.any(Function), {
      passive: true,
    })
    expect(bodyAddEventListenerSpy).toHaveBeenCalledWith('pointermove', expect.any(Function), {
      passive: true,
    })

    addEventListenerSpy.mockRestore()
    bodyAddEventListenerSpy.mockRestore()
  })

  it('removes event listeners on unmount', () => {
    const removeEventListenerSpy = jest.spyOn(window, 'removeEventListener')
    const bodyRemoveEventListenerSpy = jest.spyOn(document.body, 'removeEventListener')

    const { unmount } = render(<GlowingEffect disabled={false} />)
    unmount()

    expect(removeEventListenerSpy).toHaveBeenCalledWith('scroll', expect.any(Function))
    expect(bodyRemoveEventListenerSpy).toHaveBeenCalledWith('pointermove', expect.any(Function))

    removeEventListenerSpy.mockRestore()
    bodyRemoveEventListenerSpy.mockRestore()
  })

  it('handles pointer move events', () => {
    const { container } = render(<GlowingEffect disabled={false} proximity={100} />)

    // Get the element that has the ref
    const element = container.querySelector('[style*="--active"]') as HTMLElement
    expect(element).toBeInTheDocument()

    // Mock getBoundingClientRect
    jest.spyOn(element, 'getBoundingClientRect').mockReturnValue({
      left: 0,
      top: 0,
      width: 100,
      height: 100,
      right: 100,
      bottom: 100,
      x: 0,
      y: 0,
      toJSON: () => ({}),
    })

    // Simulate pointer move
    act(() => {
      const event = new MouseEvent('pointermove', {
        clientX: 50,
        clientY: 50,
        bubbles: true,
      })
      Object.defineProperty(event, 'x', { value: 50 })
      Object.defineProperty(event, 'y', { value: 50 })
      document.body.dispatchEvent(event)
    })

    // The component should process the event
    expect(requestAnimationFrameSpy).toHaveBeenCalled()
  })

  it('handles scroll events', () => {
    render(<GlowingEffect disabled={false} />)

    act(() => {
      window.dispatchEvent(new Event('scroll'))
    })

    expect(requestAnimationFrameSpy).toHaveBeenCalled()
  })

  it('cancels animation frame on new move', () => {
    const { container } = render(<GlowingEffect disabled={false} proximity={100} />)

    const element = container.querySelector('[style*="--active"]') as HTMLElement
    jest.spyOn(element, 'getBoundingClientRect').mockReturnValue({
      left: 0,
      top: 0,
      width: 100,
      height: 100,
      right: 100,
      bottom: 100,
      x: 0,
      y: 0,
      toJSON: () => ({}),
    })

    // First move
    act(() => {
      const event = new MouseEvent('pointermove', { bubbles: true })
      Object.defineProperty(event, 'x', { value: 50 })
      Object.defineProperty(event, 'y', { value: 50 })
      document.body.dispatchEvent(event)
    })

    // Second move should cancel previous animation frame
    act(() => {
      const event = new MouseEvent('pointermove', { bubbles: true })
      Object.defineProperty(event, 'x', { value: 60 })
      Object.defineProperty(event, 'y', { value: 60 })
      document.body.dispatchEvent(event)
    })

    expect(cancelAnimationFrameSpy).toHaveBeenCalled()
  })

  it('sets active to 0 when mouse is in inactive zone', () => {
    const { container } = render(
      <GlowingEffect disabled={false} inactiveZone={0.9} proximity={1000} />
    )

    const element = container.querySelector('[style*="--active"]') as HTMLElement
    jest.spyOn(element, 'getBoundingClientRect').mockReturnValue({
      left: 0,
      top: 0,
      width: 100,
      height: 100,
      right: 100,
      bottom: 100,
      x: 0,
      y: 0,
      toJSON: () => ({}),
    })

    // Move to center (inside inactive zone)
    act(() => {
      const event = new MouseEvent('pointermove', { bubbles: true })
      Object.defineProperty(event, 'x', { value: 50 })
      Object.defineProperty(event, 'y', { value: 50 })
      document.body.dispatchEvent(event)
    })

    expect(element.style.getPropertyValue('--active')).toBe('0')
  })

  it('sets active to 1 when mouse is near element', () => {
    const { container } = render(
      <GlowingEffect disabled={false} inactiveZone={0} proximity={200} />
    )

    const element = container.querySelector('[style*="--active"]') as HTMLElement
    jest.spyOn(element, 'getBoundingClientRect').mockReturnValue({
      left: 0,
      top: 0,
      width: 100,
      height: 100,
      right: 100,
      bottom: 100,
      x: 0,
      y: 0,
      toJSON: () => ({}),
    })

    // Move near the element
    act(() => {
      const event = new MouseEvent('pointermove', { bubbles: true })
      Object.defineProperty(event, 'x', { value: 50 })
      Object.defineProperty(event, 'y', { value: 50 })
      document.body.dispatchEvent(event)
    })

    expect(element.style.getPropertyValue('--active')).toBe('1')
  })

  it('sets active to 0 when mouse is far from element', () => {
    const { container } = render(<GlowingEffect disabled={false} inactiveZone={0} proximity={10} />)

    const element = container.querySelector('[style*="--active"]') as HTMLElement
    jest.spyOn(element, 'getBoundingClientRect').mockReturnValue({
      left: 0,
      top: 0,
      width: 100,
      height: 100,
      right: 100,
      bottom: 100,
      x: 0,
      y: 0,
      toJSON: () => ({}),
    })

    // Move far from element
    act(() => {
      const event = new MouseEvent('pointermove', { bubbles: true })
      Object.defineProperty(event, 'x', { value: 500 })
      Object.defineProperty(event, 'y', { value: 500 })
      document.body.dispatchEvent(event)
    })

    expect(element.style.getPropertyValue('--active')).toBe('0')
  })

  it('does not add event listeners when disabled', () => {
    const addEventListenerSpy = jest.spyOn(window, 'addEventListener')

    render(<GlowingEffect disabled={true} />)

    // Should not have scroll listener
    expect(addEventListenerSpy).not.toHaveBeenCalledWith(
      'scroll',
      expect.any(Function),
      expect.anything()
    )

    addEventListenerSpy.mockRestore()
  })

  it('applies blur class when blur > 0', () => {
    const { container } = render(<GlowingEffect blur={5} />)
    expect(container.querySelector('.blur-\\[var\\(--blur\\)\\]')).toBeInTheDocument()
  })

  it('has correct displayName', () => {
    expect(GlowingEffect.displayName).toBe('GlowingEffect')
  })
})
