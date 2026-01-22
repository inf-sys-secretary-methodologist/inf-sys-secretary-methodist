import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {
  Sheet,
  SheetTrigger,
  SheetContent,
  SheetHeader,
  SheetFooter,
  SheetTitle,
  SheetDescription,
  SheetClose,
} from '../sheet'

describe('Sheet', () => {
  it('renders sheet trigger', () => {
    render(
      <Sheet>
        <SheetTrigger>Open Sheet</SheetTrigger>
        <SheetContent>
          <SheetTitle>Sheet Title</SheetTitle>
        </SheetContent>
      </Sheet>
    )
    expect(screen.getByText('Open Sheet')).toBeInTheDocument()
  })

  it('opens sheet when trigger is clicked', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open Sheet</SheetTrigger>
        <SheetContent>
          <SheetTitle>Sheet Title</SheetTitle>
          <SheetDescription>Sheet description</SheetDescription>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open Sheet'))
    expect(screen.getByText('Sheet Title')).toBeInTheDocument()
    expect(screen.getByText('Sheet description')).toBeInTheDocument()
  })

  it('closes sheet when close button is clicked', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open Sheet</SheetTrigger>
        <SheetContent>
          <SheetTitle>Sheet Title</SheetTitle>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open Sheet'))
    expect(screen.getByText('Sheet Title')).toBeInTheDocument()

    const closeButton = screen.getByRole('button', { name: /close/i })
    await user.click(closeButton)

    await waitFor(() => {
      expect(screen.queryByText('Sheet Title')).not.toBeInTheDocument()
    })
  })

  it('renders with side="right" by default', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent>
          <SheetTitle>Title</SheetTitle>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    const content = screen.getByRole('dialog')
    expect(content).toHaveClass('right-0')
  })

  it('renders with side="left"', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent side="left">
          <SheetTitle>Title</SheetTitle>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    const content = screen.getByRole('dialog')
    expect(content).toHaveClass('left-0')
  })

  it('renders with side="top"', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent side="top">
          <SheetTitle>Title</SheetTitle>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    const content = screen.getByRole('dialog')
    expect(content).toHaveClass('top-0')
  })

  it('renders with side="bottom"', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent side="bottom">
          <SheetTitle>Title</SheetTitle>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    const content = screen.getByRole('dialog')
    expect(content).toHaveClass('bottom-0')
  })

  it('renders SheetHeader with correct styling', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent>
          <SheetHeader className="test-header">
            <SheetTitle>Title</SheetTitle>
          </SheetHeader>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    expect(document.querySelector('.test-header')).toBeInTheDocument()
  })

  it('renders SheetFooter with correct styling', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent>
          <SheetTitle>Title</SheetTitle>
          <SheetFooter className="test-footer">
            <button>Action</button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    expect(document.querySelector('.test-footer')).toBeInTheDocument()
  })

  it('renders SheetClose button', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent>
          <SheetTitle>Title</SheetTitle>
          <SheetClose>Cancel</SheetClose>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })

  it('closes sheet when SheetClose is clicked', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent>
          <SheetTitle>Title</SheetTitle>
          <SheetClose>Cancel</SheetClose>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    await user.click(screen.getByText('Cancel'))

    await waitFor(() => {
      expect(screen.queryByText('Title')).not.toBeInTheDocument()
    })
  })

  it('applies custom className to SheetContent', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent className="custom-content">
          <SheetTitle>Title</SheetTitle>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    expect(document.querySelector('.custom-content')).toBeInTheDocument()
  })

  it('renders with controlled open state', () => {
    render(
      <Sheet open={true}>
        <SheetContent>
          <SheetTitle>Title</SheetTitle>
        </SheetContent>
      </Sheet>
    )

    expect(screen.getByText('Title')).toBeInTheDocument()
  })

  it('calls onOpenChange when sheet state changes', async () => {
    const user = userEvent.setup()
    const onOpenChange = jest.fn()
    render(
      <Sheet onOpenChange={onOpenChange}>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent>
          <SheetTitle>Title</SheetTitle>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    expect(onOpenChange).toHaveBeenCalledWith(true)
  })

  it('applies custom className to SheetTitle', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent>
          <SheetTitle className="custom-title">Custom Title</SheetTitle>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    expect(document.querySelector('.custom-title')).toHaveTextContent('Custom Title')
  })

  it('applies custom className to SheetDescription', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent>
          <SheetTitle>Title</SheetTitle>
          <SheetDescription className="custom-desc">Custom Description</SheetDescription>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    expect(document.querySelector('.custom-desc')).toHaveTextContent('Custom Description')
  })

  it('closes on escape key', async () => {
    const user = userEvent.setup()
    render(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
        <SheetContent>
          <SheetTitle>Title</SheetTitle>
        </SheetContent>
      </Sheet>
    )

    await user.click(screen.getByText('Open'))
    expect(screen.getByText('Title')).toBeInTheDocument()

    await user.keyboard('{Escape}')
    await waitFor(() => {
      expect(screen.queryByText('Title')).not.toBeInTheDocument()
    })
  })
})
