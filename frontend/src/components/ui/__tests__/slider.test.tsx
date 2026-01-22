import { render, screen, fireEvent } from '@testing-library/react'
import { Slider } from '../slider'

// Mock ResizeObserver
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}))

describe('Slider', () => {
  it('renders slider component', () => {
    render(<Slider defaultValue={[50]} />)
    expect(screen.getByRole('slider')).toBeInTheDocument()
  })

  it('displays default value', () => {
    render(<Slider defaultValue={[50]} max={100} />)
    const slider = screen.getByRole('slider')
    expect(slider).toHaveAttribute('aria-valuenow', '50')
  })

  it('uses min and max values', () => {
    render(<Slider defaultValue={[25]} min={0} max={100} />)
    const slider = screen.getByRole('slider')
    expect(slider).toHaveAttribute('aria-valuemin', '0')
    expect(slider).toHaveAttribute('aria-valuemax', '100')
  })

  it('applies custom className', () => {
    const { container } = render(<Slider defaultValue={[50]} className="custom-slider" />)
    expect(container.querySelector('.custom-slider')).toBeInTheDocument()
  })

  it('is disabled when disabled prop is true', () => {
    render(<Slider defaultValue={[50]} disabled />)
    const slider = screen.getByRole('slider')
    expect(slider).toHaveAttribute('data-disabled')
  })

  it('calls onValueChange when value changes', async () => {
    const onValueChange = jest.fn()
    render(<Slider defaultValue={[50]} max={100} onValueChange={onValueChange} />)

    const slider = screen.getByRole('slider')
    fireEvent.keyDown(slider, { key: 'ArrowRight' })

    expect(onValueChange).toHaveBeenCalled()
  })

  it('increments value with arrow right key', () => {
    render(<Slider defaultValue={[50]} max={100} step={1} />)

    const slider = screen.getByRole('slider')
    fireEvent.keyDown(slider, { key: 'ArrowRight' })

    expect(slider).toHaveAttribute('aria-valuenow', '51')
  })

  it('decrements value with arrow left key', () => {
    render(<Slider defaultValue={[50]} max={100} step={1} />)

    const slider = screen.getByRole('slider')
    fireEvent.keyDown(slider, { key: 'ArrowLeft' })

    expect(slider).toHaveAttribute('aria-valuenow', '49')
  })

  it('respects step value', () => {
    render(<Slider defaultValue={[50]} max={100} step={10} />)

    const slider = screen.getByRole('slider')
    fireEvent.keyDown(slider, { key: 'ArrowRight' })

    expect(slider).toHaveAttribute('aria-valuenow', '60')
  })

  it('does not exceed max value', () => {
    render(<Slider defaultValue={[99]} max={100} step={5} />)

    const slider = screen.getByRole('slider')
    fireEvent.keyDown(slider, { key: 'ArrowRight' })

    expect(slider).toHaveAttribute('aria-valuenow', '100')
  })

  it('does not go below min value', () => {
    render(<Slider defaultValue={[1]} min={0} max={100} step={5} />)

    const slider = screen.getByRole('slider')
    fireEvent.keyDown(slider, { key: 'ArrowLeft' })

    expect(slider).toHaveAttribute('aria-valuenow', '0')
  })

  it('renders with controlled value', () => {
    const { rerender } = render(<Slider value={[30]} max={100} />)
    expect(screen.getByRole('slider')).toHaveAttribute('aria-valuenow', '30')

    rerender(<Slider value={[70]} max={100} />)
    expect(screen.getByRole('slider')).toHaveAttribute('aria-valuenow', '70')
  })

  it('has correct display name', () => {
    expect(Slider.displayName).toBe('Slider')
  })

  it('supports orientation prop', () => {
    render(<Slider defaultValue={[50]} orientation="vertical" />)
    const slider = screen.getByRole('slider')
    expect(slider.closest('[data-orientation="vertical"]')).toBeInTheDocument()
  })

  it('handles Home key to go to minimum', () => {
    render(<Slider defaultValue={[50]} min={0} max={100} />)

    const slider = screen.getByRole('slider')
    fireEvent.keyDown(slider, { key: 'Home' })

    expect(slider).toHaveAttribute('aria-valuenow', '0')
  })

  it('handles End key to go to maximum', () => {
    render(<Slider defaultValue={[50]} min={0} max={100} />)

    const slider = screen.getByRole('slider')
    fireEvent.keyDown(slider, { key: 'End' })

    expect(slider).toHaveAttribute('aria-valuenow', '100')
  })
})
