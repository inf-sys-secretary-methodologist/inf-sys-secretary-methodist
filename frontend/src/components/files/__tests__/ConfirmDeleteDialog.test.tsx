import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { ConfirmDeleteDialog } from '../ConfirmDeleteDialog'

describe('ConfirmDeleteDialog', () => {
  it('renders title and description when open', () => {
    render(
      <ConfirmDeleteDialog
        open={true}
        onConfirm={jest.fn()}
        onCancel={jest.fn()}
      />
    )
    expect(screen.getByText('confirm.deleteTitle')).toBeInTheDocument()
    expect(screen.getByText('confirm.delete')).toBeInTheDocument()
  })

  it('does not render content when closed', () => {
    render(
      <ConfirmDeleteDialog
        open={false}
        onConfirm={jest.fn()}
        onCancel={jest.fn()}
      />
    )
    expect(screen.queryByText('confirm.deleteTitle')).not.toBeInTheDocument()
  })

  it('calls onConfirm when confirm button is clicked', async () => {
    const onConfirm = jest.fn()
    const user = userEvent.setup()

    render(
      <ConfirmDeleteDialog
        open={true}
        onConfirm={onConfirm}
        onCancel={jest.fn()}
      />
    )

    const confirmBtn = screen.getByRole('button', { name: /delete/i })
    await user.click(confirmBtn)
    expect(onConfirm).toHaveBeenCalled()
  })

  it('calls onCancel when cancel button is clicked', async () => {
    const onCancel = jest.fn()
    const user = userEvent.setup()

    render(
      <ConfirmDeleteDialog
        open={true}
        onConfirm={jest.fn()}
        onCancel={onCancel}
      />
    )

    const cancelBtn = screen.getByRole('button', { name: /cancel/i })
    await user.click(cancelBtn)
    expect(onCancel).toHaveBeenCalled()
  })
})
