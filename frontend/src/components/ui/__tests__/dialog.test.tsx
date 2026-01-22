import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
  DialogClose,
} from '../dialog'

describe('Dialog', () => {
  it('renders dialog trigger', () => {
    render(
      <Dialog>
        <DialogTrigger>Open Dialog</DialogTrigger>
        <DialogContent>
          <DialogTitle>Test Title</DialogTitle>
        </DialogContent>
      </Dialog>
    )
    expect(screen.getByText('Open Dialog')).toBeInTheDocument()
  })

  it('opens dialog when trigger is clicked', async () => {
    const user = userEvent.setup()
    render(
      <Dialog>
        <DialogTrigger>Open Dialog</DialogTrigger>
        <DialogContent>
          <DialogTitle>Test Title</DialogTitle>
          <DialogDescription>Test description</DialogDescription>
        </DialogContent>
      </Dialog>
    )

    await user.click(screen.getByText('Open Dialog'))
    expect(screen.getByText('Test Title')).toBeInTheDocument()
    expect(screen.getByText('Test description')).toBeInTheDocument()
  })

  it('closes dialog when close button is clicked', async () => {
    const user = userEvent.setup()
    render(
      <Dialog>
        <DialogTrigger>Open Dialog</DialogTrigger>
        <DialogContent>
          <DialogTitle>Test Title</DialogTitle>
        </DialogContent>
      </Dialog>
    )

    await user.click(screen.getByText('Open Dialog'))
    expect(screen.getByText('Test Title')).toBeInTheDocument()

    const closeButton = screen.getByRole('button', { name: /close/i })
    await user.click(closeButton)

    await waitFor(() => {
      expect(screen.queryByText('Test Title')).not.toBeInTheDocument()
    })
  })

  it('renders DialogHeader with correct styling', async () => {
    const user = userEvent.setup()
    render(
      <Dialog>
        <DialogTrigger>Open Dialog</DialogTrigger>
        <DialogContent>
          <DialogHeader className="test-header">
            <DialogTitle>Test Title</DialogTitle>
          </DialogHeader>
        </DialogContent>
      </Dialog>
    )

    await user.click(screen.getByText('Open Dialog'))
    expect(document.querySelector('.test-header')).toBeInTheDocument()
  })

  it('renders DialogFooter with correct styling', async () => {
    const user = userEvent.setup()
    render(
      <Dialog>
        <DialogTrigger>Open Dialog</DialogTrigger>
        <DialogContent>
          <DialogTitle>Test Title</DialogTitle>
          <DialogFooter className="test-footer">
            <button>Submit</button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    )

    await user.click(screen.getByText('Open Dialog'))
    expect(document.querySelector('.test-footer')).toBeInTheDocument()
  })

  it('renders DialogClose button', async () => {
    const user = userEvent.setup()
    render(
      <Dialog>
        <DialogTrigger>Open Dialog</DialogTrigger>
        <DialogContent>
          <DialogTitle>Test Title</DialogTitle>
          <DialogClose>Cancel</DialogClose>
        </DialogContent>
      </Dialog>
    )

    await user.click(screen.getByText('Open Dialog'))
    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })

  it('closes dialog when DialogClose is clicked', async () => {
    const user = userEvent.setup()
    render(
      <Dialog>
        <DialogTrigger>Open Dialog</DialogTrigger>
        <DialogContent>
          <DialogTitle>Test Title</DialogTitle>
          <DialogClose>Cancel</DialogClose>
        </DialogContent>
      </Dialog>
    )

    await user.click(screen.getByText('Open Dialog'))
    await user.click(screen.getByText('Cancel'))

    await waitFor(() => {
      expect(screen.queryByText('Test Title')).not.toBeInTheDocument()
    })
  })

  it('applies custom className to DialogContent', async () => {
    const user = userEvent.setup()
    render(
      <Dialog>
        <DialogTrigger>Open Dialog</DialogTrigger>
        <DialogContent className="custom-content">
          <DialogTitle>Test Title</DialogTitle>
        </DialogContent>
      </Dialog>
    )

    await user.click(screen.getByText('Open Dialog'))
    expect(document.querySelector('.custom-content')).toBeInTheDocument()
  })

  it('renders with controlled open state', () => {
    render(
      <Dialog open={true}>
        <DialogContent>
          <DialogTitle>Test Title</DialogTitle>
        </DialogContent>
      </Dialog>
    )

    expect(screen.getByText('Test Title')).toBeInTheDocument()
  })

  it('calls onOpenChange when dialog state changes', async () => {
    const user = userEvent.setup()
    const onOpenChange = jest.fn()
    render(
      <Dialog onOpenChange={onOpenChange}>
        <DialogTrigger>Open Dialog</DialogTrigger>
        <DialogContent>
          <DialogTitle>Test Title</DialogTitle>
        </DialogContent>
      </Dialog>
    )

    await user.click(screen.getByText('Open Dialog'))
    expect(onOpenChange).toHaveBeenCalledWith(true)
  })

  it('applies custom className to DialogTitle', async () => {
    const user = userEvent.setup()
    render(
      <Dialog>
        <DialogTrigger>Open Dialog</DialogTrigger>
        <DialogContent>
          <DialogTitle className="custom-title">Test Title</DialogTitle>
        </DialogContent>
      </Dialog>
    )

    await user.click(screen.getByText('Open Dialog'))
    expect(document.querySelector('.custom-title')).toHaveTextContent('Test Title')
  })

  it('applies custom className to DialogDescription', async () => {
    const user = userEvent.setup()
    render(
      <Dialog>
        <DialogTrigger>Open Dialog</DialogTrigger>
        <DialogContent>
          <DialogTitle>Title</DialogTitle>
          <DialogDescription className="custom-desc">Description text</DialogDescription>
        </DialogContent>
      </Dialog>
    )

    await user.click(screen.getByText('Open Dialog'))
    expect(document.querySelector('.custom-desc')).toHaveTextContent('Description text')
  })
})
