import { render, screen, fireEvent } from '@testing-library/react'
import { Collapsible, CollapsibleTrigger, CollapsibleContent } from '../collapsible'

describe('Collapsible', () => {
  it('renders without crashing', () => {
    render(
      <Collapsible>
        <CollapsibleTrigger>Toggle</CollapsibleTrigger>
        <CollapsibleContent>Content</CollapsibleContent>
      </Collapsible>
    )
    expect(screen.getByText('Toggle')).toBeInTheDocument()
  })

  it('content is hidden by default', () => {
    render(
      <Collapsible>
        <CollapsibleTrigger>Toggle</CollapsibleTrigger>
        <CollapsibleContent>Hidden Content</CollapsibleContent>
      </Collapsible>
    )
    // Radix UI doesn't render CollapsibleContent when closed
    expect(screen.queryByText('Hidden Content')).not.toBeInTheDocument()
  })

  it('shows content when open is true', () => {
    render(
      <Collapsible open>
        <CollapsibleTrigger>Toggle</CollapsibleTrigger>
        <CollapsibleContent>Visible Content</CollapsibleContent>
      </Collapsible>
    )
    expect(screen.getByText('Visible Content')).toBeVisible()
  })

  it('toggles content on trigger click', () => {
    render(
      <Collapsible>
        <CollapsibleTrigger>Toggle</CollapsibleTrigger>
        <CollapsibleContent>Toggle Content</CollapsibleContent>
      </Collapsible>
    )

    const trigger = screen.getByText('Toggle')

    // Initially not in the document
    expect(screen.queryByText('Toggle Content')).not.toBeInTheDocument()

    // Click to open
    fireEvent.click(trigger)
    expect(screen.getByText('Toggle Content')).toBeVisible()

    // Click to close
    fireEvent.click(trigger)
    expect(screen.queryByText('Toggle Content')).not.toBeInTheDocument()
  })

  it('calls onOpenChange when toggled', () => {
    const handleOpenChange = jest.fn()

    render(
      <Collapsible onOpenChange={handleOpenChange}>
        <CollapsibleTrigger>Toggle</CollapsibleTrigger>
        <CollapsibleContent>Content</CollapsibleContent>
      </Collapsible>
    )

    fireEvent.click(screen.getByText('Toggle'))
    expect(handleOpenChange).toHaveBeenCalledWith(true)

    fireEvent.click(screen.getByText('Toggle'))
    expect(handleOpenChange).toHaveBeenCalledWith(false)
  })

  it('can be controlled', () => {
    const { rerender } = render(
      <Collapsible open={false}>
        <CollapsibleTrigger>Toggle</CollapsibleTrigger>
        <CollapsibleContent>Controlled Content</CollapsibleContent>
      </Collapsible>
    )

    expect(screen.queryByText('Controlled Content')).not.toBeInTheDocument()

    rerender(
      <Collapsible open={true}>
        <CollapsibleTrigger>Toggle</CollapsibleTrigger>
        <CollapsibleContent>Controlled Content</CollapsibleContent>
      </Collapsible>
    )

    expect(screen.getByText('Controlled Content')).toBeVisible()
  })

  it('can be disabled', () => {
    render(
      <Collapsible disabled>
        <CollapsibleTrigger>Toggle</CollapsibleTrigger>
        <CollapsibleContent>Content</CollapsibleContent>
      </Collapsible>
    )

    const trigger = screen.getByText('Toggle')
    fireEvent.click(trigger)

    // When disabled, clicking doesn't open the collapsible
    expect(screen.queryByText('Content')).not.toBeInTheDocument()
  })
})
