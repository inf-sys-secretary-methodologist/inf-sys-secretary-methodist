import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { FileFilters, FileFilterValues } from '../FileFilters'

describe('FileFilters', () => {
  const defaultFilters: FileFilterValues = {}

  it('renders search input', () => {
    render(<FileFilters value={defaultFilters} onChange={jest.fn()} />)
    expect(screen.getByPlaceholderText('searchPlaceholder')).toBeInTheDocument()
  })

  it('renders file type select with options', () => {
    render(<FileFilters value={defaultFilters} onChange={jest.fn()} />)
    const select = screen.getByLabelText('filters.type')
    expect(select).toBeInTheDocument()
  })

  it('calls onChange when search text is entered', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()

    render(<FileFilters value={defaultFilters} onChange={onChange} />)

    const input = screen.getByPlaceholderText('searchPlaceholder')
    await user.type(input, 'report')

    expect(onChange).toHaveBeenCalled()
    const lastCall = onChange.mock.calls[onChange.mock.calls.length - 1][0]
    expect(lastCall.search).toContain('report')
  })

  it('calls onChange when file type is selected', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()

    render(<FileFilters value={defaultFilters} onChange={onChange} />)

    const select = screen.getByLabelText('filters.type')
    await user.selectOptions(select, 'image')

    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ fileType: 'image' }))
  })

  it('renders reset button that clears all filters', async () => {
    const onChange = jest.fn()
    const user = userEvent.setup()

    render(
      <FileFilters value={{ search: 'test', fileType: 'image' }} onChange={onChange} />
    )

    const resetBtn = screen.getByRole('button', { name: /reset/i })
    await user.click(resetBtn)
    expect(onChange).toHaveBeenCalledWith({})
  })
})
