import { render, screen, act } from '@testing-library/react'
import { NumberTicker } from '../number-ticker'
import { useMotionValue, useSpring, useInView } from 'motion/react'

// Store change listeners
let changeListener: ((value: number) => void) | null = null
let currentMotionValue = 0

// Mock motion/react
jest.mock('motion/react', () => ({
  useInView: jest.fn(() => true),
  useMotionValue: jest.fn((initial: number) => {
    currentMotionValue = initial
    return {
      get: () => currentMotionValue,
      set: (val: number) => {
        currentMotionValue = val
        // Trigger change listener
        if (changeListener) {
          changeListener(val)
        }
      },
    }
  }),
  useSpring: jest.fn((motionValue) => ({
    on: (event: string, callback: (value: number) => void) => {
      if (event === 'change') {
        changeListener = callback
        // Immediately call with current value
        callback(currentMotionValue)
      }
      // Return unsubscribe function
      return () => {
        changeListener = null
      }
    },
    get: () => motionValue.get(),
  })),
}))

const mockedUseMotionValue = jest.mocked(useMotionValue)
const mockedUseSpring = jest.mocked(useSpring)
const mockedUseInView = jest.mocked(useInView)

describe('NumberTicker', () => {
  beforeEach(() => {
    jest.useFakeTimers()
    changeListener = null
    currentMotionValue = 0
  })

  afterEach(() => {
    jest.useRealTimers()
    jest.clearAllMocks()
  })

  it('renders with initial value', () => {
    render(<NumberTicker value={100} />)
    expect(screen.getByText('0')).toBeInTheDocument()
  })

  it('renders with custom startValue', () => {
    render(<NumberTicker value={100} startValue={50} />)
    expect(screen.getByText('50')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    render(<NumberTicker value={100} className="custom-ticker" />)
    expect(document.querySelector('.custom-ticker')).toBeInTheDocument()
  })

  it('applies default tracking class', () => {
    render(<NumberTicker value={100} />)
    expect(document.querySelector('.tracking-wider')).toBeInTheDocument()
  })

  it('applies tabular-nums class', () => {
    render(<NumberTicker value={100} />)
    expect(document.querySelector('.tabular-nums')).toBeInTheDocument()
  })

  it('passes additional props to span', () => {
    render(<NumberTicker value={100} data-testid="ticker" />)
    expect(screen.getByTestId('ticker')).toBeInTheDocument()
  })

  it('starts animation after delay', () => {
    // Using mockedUseMotionValue from top of file
    const mockSet = jest.fn()
    mockedUseMotionValue.mockReturnValue({
      get: () => 0,
      set: mockSet,
    })

    render(<NumberTicker value={100} delay={1} />)

    // Before delay
    expect(mockSet).not.toHaveBeenCalled()

    // After delay (1 second = 1000ms)
    act(() => {
      jest.advanceTimersByTime(1000)
    })

    expect(mockSet).toHaveBeenCalledWith(100)
  })

  it('sets correct initial value for direction up', () => {
    // Using mockedUseMotionValue from top of file
    mockedUseMotionValue.mockClear()

    render(<NumberTicker value={100} direction="up" startValue={0} />)

    // Should start from startValue (0) when direction is up
    expect(mockedUseMotionValue).toHaveBeenCalledWith(0)
  })

  it('sets correct initial value for direction down', () => {
    // Using mockedUseMotionValue from top of file
    mockedUseMotionValue.mockClear()

    render(<NumberTicker value={100} direction="down" startValue={0} />)

    // Should start from value (100) when direction is down
    expect(mockedUseMotionValue).toHaveBeenCalledWith(100)
  })

  it('animates to value when direction is up', () => {
    // Using mockedUseMotionValue from top of file
    const mockSet = jest.fn()
    mockedUseMotionValue.mockReturnValue({
      get: () => 0,
      set: mockSet,
    })

    render(<NumberTicker value={100} direction="up" />)

    act(() => {
      jest.advanceTimersByTime(0)
    })

    expect(mockSet).toHaveBeenCalledWith(100)
  })

  it('animates to startValue when direction is down', () => {
    // Using mockedUseMotionValue from top of file
    const mockSet = jest.fn()
    mockedUseMotionValue.mockReturnValue({
      get: () => 100,
      set: mockSet,
    })

    render(<NumberTicker value={100} direction="down" startValue={0} />)

    act(() => {
      jest.advanceTimersByTime(0)
    })

    expect(mockSet).toHaveBeenCalledWith(0)
  })

  it('formats number with decimal places', () => {
    // Reset and set up mock to trigger change
    // Using mockedUseSpring from top of file
    mockedUseSpring.mockImplementation(() => ({
      on: (event: string, callback: (value: number) => void) => {
        if (event === 'change') {
          // Call with a value that has decimals
          callback(123.456)
        }
        return () => {}
      },
      get: () => 123.456,
    }))

    render(<NumberTicker value={123.456} decimalPlaces={2} />)

    // Should format with 2 decimal places
    expect(screen.getByText('123.46')).toBeInTheDocument()
  })

  it('formats number without decimal places by default', () => {
    // Using mockedUseSpring from top of file
    mockedUseSpring.mockImplementation(() => ({
      on: (event: string, callback: (value: number) => void) => {
        if (event === 'change') {
          callback(100)
        }
        return () => {}
      },
      get: () => 100,
    }))

    render(<NumberTicker value={100} />)

    expect(screen.getByText('100')).toBeInTheDocument()
  })

  it('does not animate when not in view', () => {
    // Using mockedUseInView, mockedUseMotionValue from top of file
    mockedUseInView.mockReturnValue(false)
    const mockSet = jest.fn()
    mockedUseMotionValue.mockReturnValue({
      get: () => 0,
      set: mockSet,
    })

    render(<NumberTicker value={100} />)

    act(() => {
      jest.advanceTimersByTime(1000)
    })

    // Should not call set because not in view
    expect(mockSet).not.toHaveBeenCalled()
  })

  it('cleans up timer on unmount', () => {
    // Using mockedUseInView from top of file
    mockedUseInView.mockReturnValue(true) // Must be in view to create timer

    const clearTimeoutSpy = jest.spyOn(global, 'clearTimeout')

    const { unmount } = render(<NumberTicker value={100} delay={1} />)

    unmount()

    expect(clearTimeoutSpy).toHaveBeenCalled()

    clearTimeoutSpy.mockRestore()
  })

  it('uses useSpring with correct config', () => {
    // Using mockedUseSpring from top of file
    mockedUseSpring.mockClear()

    render(<NumberTicker value={100} />)

    expect(mockedUseSpring).toHaveBeenCalledWith(expect.anything(), {
      damping: 60,
      stiffness: 100,
    })
  })

  it('uses useInView with correct options', () => {
    // Using mockedUseInView from top of file
    mockedUseInView.mockClear()

    render(<NumberTicker value={100} />)

    expect(mockedUseInView).toHaveBeenCalledWith(expect.anything(), {
      once: true,
      margin: '0px',
    })
  })

  it('formats large numbers with locale', () => {
    // Using mockedUseSpring from top of file
    mockedUseSpring.mockImplementation(() => ({
      on: (event: string, callback: (value: number) => void) => {
        if (event === 'change') {
          callback(1234567)
        }
        return () => {}
      },
      get: () => 1234567,
    }))

    render(<NumberTicker value={1234567} />)

    // en-US locale formats with commas
    expect(screen.getByText('1,234,567')).toBeInTheDocument()
  })

  it('unsubscribes from spring value changes on unmount', () => {
    const unsubscribe = jest.fn()
    // Using mockedUseSpring from top of file
    mockedUseSpring.mockImplementation(() => ({
      on: () => unsubscribe,
      get: () => 0,
    }))

    const { unmount } = render(<NumberTicker value={100} />)

    unmount()

    expect(unsubscribe).toHaveBeenCalled()
  })
})
