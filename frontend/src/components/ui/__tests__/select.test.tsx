import { render, screen } from '@testing-library/react'
import {
  Select,
  SelectTrigger,
  SelectContent,
  SelectItem,
  SelectValue,
  SelectGroup,
  SelectLabel,
  SelectSeparator,
  SelectScrollUpButton,
  SelectScrollDownButton,
} from '../select'

// Mock ResizeObserver
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}))

describe('Select', () => {
  it('renders select trigger', () => {
    render(
      <Select>
        <SelectTrigger>
          <SelectValue placeholder="Select an option" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="option1">Option 1</SelectItem>
        </SelectContent>
      </Select>
    )
    expect(screen.getByRole('combobox')).toBeInTheDocument()
    expect(screen.getByText('Select an option')).toBeInTheDocument()
  })

  it('displays selected value with defaultValue', () => {
    render(
      <Select defaultValue="option2">
        <SelectTrigger>
          <SelectValue placeholder="Select an option" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="option1">Option 1</SelectItem>
          <SelectItem value="option2">Option 2</SelectItem>
        </SelectContent>
      </Select>
    )

    expect(screen.getByRole('combobox')).toHaveTextContent('Option 2')
  })

  it('applies custom className to SelectTrigger', () => {
    render(
      <Select>
        <SelectTrigger className="custom-trigger">
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="a">A</SelectItem>
        </SelectContent>
      </Select>
    )

    expect(screen.getByRole('combobox')).toHaveClass('custom-trigger')
  })

  it('disables select when disabled prop is true', () => {
    render(
      <Select disabled>
        <SelectTrigger>
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="a">A</SelectItem>
        </SelectContent>
      </Select>
    )

    expect(screen.getByRole('combobox')).toBeDisabled()
  })

  it('works with controlled value', () => {
    const { rerender } = render(
      <Select value="option1">
        <SelectTrigger>
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="option1">Option 1</SelectItem>
          <SelectItem value="option2">Option 2</SelectItem>
        </SelectContent>
      </Select>
    )

    expect(screen.getByRole('combobox')).toHaveTextContent('Option 1')

    rerender(
      <Select value="option2">
        <SelectTrigger>
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="option1">Option 1</SelectItem>
          <SelectItem value="option2">Option 2</SelectItem>
        </SelectContent>
      </Select>
    )

    expect(screen.getByRole('combobox')).toHaveTextContent('Option 2')
  })

  it('has correct displayName for all components', () => {
    expect(SelectTrigger.displayName).toBe('SelectTrigger')
    expect(SelectContent.displayName).toBe('SelectContent')
    expect(SelectItem.displayName).toBe('SelectItem')
    expect(SelectLabel.displayName).toBe('SelectLabel')
    expect(SelectSeparator.displayName).toBe('SelectSeparator')
    expect(SelectScrollUpButton.displayName).toBe('SelectScrollUpButton')
    expect(SelectScrollDownButton.displayName).toBe('SelectScrollDownButton')
  })

  it('renders with chevron icon', () => {
    const { container } = render(
      <Select>
        <SelectTrigger>
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="a">A</SelectItem>
        </SelectContent>
      </Select>
    )

    // Chevron icon should be rendered
    expect(container.querySelector('svg')).toBeInTheDocument()
  })

  it('sets aria-expanded to false when closed', () => {
    render(
      <Select>
        <SelectTrigger>
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="a">A</SelectItem>
        </SelectContent>
      </Select>
    )

    expect(screen.getByRole('combobox')).toHaveAttribute('aria-expanded', 'false')
  })

  it('renders trigger with type button', () => {
    render(
      <Select>
        <SelectTrigger>
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="a">A</SelectItem>
        </SelectContent>
      </Select>
    )

    expect(screen.getByRole('combobox')).toHaveAttribute('type', 'button')
  })

  it('renders with data-state attribute', () => {
    render(
      <Select>
        <SelectTrigger>
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="a">A</SelectItem>
        </SelectContent>
      </Select>
    )

    expect(screen.getByRole('combobox')).toHaveAttribute('data-state', 'closed')
  })

  it('renders trigger with role combobox', () => {
    render(
      <Select>
        <SelectTrigger>
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="a">A</SelectItem>
        </SelectContent>
      </Select>
    )

    expect(screen.getByRole('combobox')).toBeInTheDocument()
  })

  it('shows placeholder text when no value is selected', () => {
    render(
      <Select>
        <SelectTrigger>
          <SelectValue placeholder="Choose an item" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="a">A</SelectItem>
        </SelectContent>
      </Select>
    )

    expect(screen.getByText('Choose an item')).toBeInTheDocument()
  })

  it('renders with SelectGroup', () => {
    render(
      <Select defaultValue="apple">
        <SelectTrigger>
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectLabel>Fruits</SelectLabel>
            <SelectItem value="apple">Apple</SelectItem>
            <SelectItem value="banana">Banana</SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
    )

    // With defaultValue, the combobox should show the selected item
    expect(screen.getByRole('combobox')).toHaveTextContent('Apple')
  })
})
