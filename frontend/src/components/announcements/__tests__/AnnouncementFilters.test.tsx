import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { AnnouncementFilters } from '../AnnouncementFilters'

describe('AnnouncementFilters', () => {
  it('renders search, status, priority, audience controls', () => {
    render(<AnnouncementFilters value={{}} onChange={jest.fn()} />)
    expect(screen.getByPlaceholderText(/searchPlaceholder|search/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/status/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/priority/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/audience/i)).toBeInTheDocument()
  })

  it('calls onChange when search input changes', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()
    render(<AnnouncementFilters value={{}} onChange={onChange} />)
    await user.type(screen.getByPlaceholderText(/searchPlaceholder|search/i), 'x')
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ search: 'x' }))
  })

  it('calls onChange when status is selected', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()
    render(<AnnouncementFilters value={{}} onChange={onChange} />)
    await user.selectOptions(screen.getByLabelText(/status/i), 'published')
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ status: 'published' }))
  })

  it('calls onChange when priority is selected', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()
    render(<AnnouncementFilters value={{}} onChange={onChange} />)
    await user.selectOptions(screen.getByLabelText(/priority/i), 'urgent')
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ priority: 'urgent' }))
  })

  it('calls onChange when target audience is selected', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()
    render(<AnnouncementFilters value={{}} onChange={onChange} />)
    await user.selectOptions(screen.getByLabelText(/audience/i), 'students')
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ target_audience: 'students' }))
  })

  it('toggles pinned-only filter', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()
    render(<AnnouncementFilters value={{}} onChange={onChange} />)
    await user.click(screen.getByLabelText(/pinned|закреп/i))
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ is_pinned: true }))
  })

  it('reflects current filter value', () => {
    render(
      <AnnouncementFilters
        value={{ status: 'archived', priority: 'low', target_audience: 'staff' }}
        onChange={jest.fn()}
      />
    )
    expect((screen.getByLabelText(/status/i) as HTMLSelectElement).value).toBe('archived')
    expect((screen.getByLabelText(/priority/i) as HTMLSelectElement).value).toBe('low')
    expect((screen.getByLabelText(/audience/i) as HTMLSelectElement).value).toBe('staff')
  })

  it('resets all filters when reset button is clicked', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()
    render(
      <AnnouncementFilters
        value={{ status: 'draft', priority: 'high', search: 'x', is_pinned: true }}
        onChange={onChange}
      />
    )
    await user.click(screen.getByRole('button', { name: /reset|сброс/i }))
    expect(onChange).toHaveBeenCalledWith({})
  })
})
