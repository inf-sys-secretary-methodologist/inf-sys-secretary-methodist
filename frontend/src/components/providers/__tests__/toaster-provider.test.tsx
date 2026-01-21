import { render } from '@testing-library/react'
import { ToasterProvider, toast } from '../toaster-provider'

// Mock sonner
jest.mock('sonner', () => ({
  Toaster: (props: Record<string, unknown>) => (
    <div data-testid="toaster" data-props={JSON.stringify(props)} />
  ),
  toast: {
    success: jest.fn(),
    error: jest.fn(),
    info: jest.fn(),
    warning: jest.fn(),
  },
}))

describe('ToasterProvider', () => {
  it('renders Toaster component', () => {
    const { getByTestId } = render(<ToasterProvider />)
    expect(getByTestId('toaster')).toBeInTheDocument()
  })

  it('passes correct props to Toaster', () => {
    const { getByTestId } = render(<ToasterProvider />)
    const toaster = getByTestId('toaster')
    const props = JSON.parse(toaster.getAttribute('data-props') || '{}')

    expect(props.position).toBe('top-right')
    expect(props.richColors).toBe(true)
    expect(props.closeButton).toBe(true)
    expect(props.expand).toBe(false)
    expect(props.duration).toBe(4000)
    expect(props.theme).toBe('system')
  })

  it('re-exports toast from sonner', () => {
    expect(toast).toBeDefined()
    expect(toast.success).toBeDefined()
    expect(toast.error).toBeDefined()
  })
})
