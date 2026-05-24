import { render, screen, fireEvent } from '@testing-library/react'
import { ExtracurricularEventFilters } from '../ExtracurricularEventFilters'
import type { ExtracurricularEventFilterParams } from '@/types/extracurricular'

// next-intl auto-mocked — t returns key verbatim. Parity test in
// hooks/__tests__/extracurricular.i18n.test.ts loads real JSON.

describe('ExtracurricularEventFilters', () => {
  it('renders status select containing all canonical statuses', () => {
    const onChange = jest.fn()
    render(<ExtracurricularEventFilters value={{}} onChange={onChange} />)
    const select = screen.getByLabelText('status.label') as HTMLSelectElement
    const optionValues = Array.from(select.options).map((o) => o.value)
    expect(optionValues).toEqual(
      expect.arrayContaining(['', 'draft', 'published', 'canceled', 'completed'])
    )
  })

  it('renders category select containing all canonical categories', () => {
    const onChange = jest.fn()
    render(<ExtracurricularEventFilters value={{}} onChange={onChange} />)
    const select = screen.getByLabelText('category.label') as HTMLSelectElement
    const optionValues = Array.from(select.options).map((o) => o.value)
    expect(optionValues).toEqual(
      expect.arrayContaining(['', 'academic', 'cultural', 'sports', 'volunteer', 'professional'])
    )
  })

  it('calls onChange with status patch', () => {
    const onChange = jest.fn()
    render(<ExtracurricularEventFilters value={{}} onChange={onChange} />)
    fireEvent.change(screen.getByLabelText('status.label'), { target: { value: 'published' } })
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ status: 'published' }))
  })

  it('calls onChange with category patch', () => {
    const onChange = jest.fn()
    render(<ExtracurricularEventFilters value={{}} onChange={onChange} />)
    fireEvent.change(screen.getByLabelText('category.label'), { target: { value: 'sports' } })
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ category: 'sports' }))
  })

  it('drops the filter when "all" (empty value) is selected', () => {
    const onChange = jest.fn()
    const value: ExtracurricularEventFilterParams = { status: 'draft', category: 'cultural' }
    render(<ExtracurricularEventFilters value={value} onChange={onChange} />)
    fireEvent.change(screen.getByLabelText('status.label'), { target: { value: '' } })
    expect(onChange).toHaveBeenCalledWith(
      expect.not.objectContaining({ status: expect.any(String) })
    )
  })

  it('reflects current value in select state', () => {
    const onChange = jest.fn()
    const value: ExtracurricularEventFilterParams = { status: 'canceled', category: 'volunteer' }
    render(<ExtracurricularEventFilters value={value} onChange={onChange} />)
    const status = screen.getByLabelText('status.label') as HTMLSelectElement
    const category = screen.getByLabelText('category.label') as HTMLSelectElement
    expect(status.value).toBe('canceled')
    expect(category.value).toBe('volunteer')
  })
})
