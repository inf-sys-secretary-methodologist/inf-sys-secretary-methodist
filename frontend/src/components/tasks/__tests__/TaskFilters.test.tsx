import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { TaskFilters } from '../TaskFilters'

describe('TaskFilters', () => {
  it('renders search, status and priority controls', () => {
    render(<TaskFilters value={{}} onChange={jest.fn()} />)
    expect(screen.getByPlaceholderText(/searchPlaceholder|search/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/status/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/priority/i)).toBeInTheDocument()
  })

  it('calls onChange when search input changes', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()
    render(<TaskFilters value={{}} onChange={onChange} />)

    const search = screen.getByPlaceholderText(/searchPlaceholder|search/i)
    // Single keystroke is enough — the controlled component forwards each
    // change to the parent. Cumulative string would require a stateful wrapper.
    await user.type(search, 'r')

    expect(onChange).toHaveBeenCalled()
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ search: 'r' }))
  })

  it('calls onChange when status is selected', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()
    render(<TaskFilters value={{}} onChange={onChange} />)

    const status = screen.getByLabelText(/status/i)
    await user.selectOptions(status, 'in_progress')

    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ status: 'in_progress' }))
  })

  it('calls onChange when priority is selected', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()
    render(<TaskFilters value={{}} onChange={onChange} />)

    const priority = screen.getByLabelText(/priority/i)
    await user.selectOptions(priority, 'urgent')

    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ priority: 'urgent' }))
  })

  it('reflects current filter value', () => {
    render(<TaskFilters value={{ status: 'completed', priority: 'low' }} onChange={jest.fn()} />)
    const status = screen.getByLabelText(/status/i) as HTMLSelectElement
    const priority = screen.getByLabelText(/priority/i) as HTMLSelectElement
    expect(status.value).toBe('completed')
    expect(priority.value).toBe('low')
  })

  it('resets filters when reset button is clicked', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()
    render(
      <TaskFilters value={{ status: 'completed', priority: 'high', search: 'x' }} onChange={onChange} />
    )

    await user.click(screen.getByRole('button', { name: /reset|сброс/i }))
    expect(onChange).toHaveBeenCalledWith({})
  })
})
