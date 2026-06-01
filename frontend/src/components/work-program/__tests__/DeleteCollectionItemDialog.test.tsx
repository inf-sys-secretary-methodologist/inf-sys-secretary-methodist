import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { DeleteCollectionItemDialog } from '../DeleteCollectionItemDialog'

const mockToastSuccess = jest.fn()
const mockToastError = jest.fn()
jest.mock('sonner', () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}))

beforeEach(() => {
  jest.clearAllMocks()
})

const wp = { id: 7 } as never

describe('DeleteCollectionItemDialog', () => {
  it('does not render when open=false', () => {
    render(
      <DeleteCollectionItemDialog
        open={false}
        itemLabel="Цель 1"
        onConfirm={jest.fn()}
        onDone={jest.fn()}
        onClose={jest.fn()}
      />
    )
    expect(screen.queryByText('collectionDialog.deleteTitle')).not.toBeInTheDocument()
  })

  it('renders the title + item preview + confirm/cancel when open', () => {
    render(
      <DeleteCollectionItemDialog
        open={true}
        itemLabel="ПК-1 — применять СУБД"
        onConfirm={jest.fn()}
        onDone={jest.fn()}
        onClose={jest.fn()}
      />
    )
    expect(screen.getByText('collectionDialog.deleteTitle')).toBeInTheDocument()
    expect(screen.getByText('ПК-1 — применять СУБД')).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'collectionDialog.deleteConfirm' })
    ).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'collectionDialog.cancel' })).toBeInTheDocument()
  })

  it('calls onConfirm, then onDone + success toast', async () => {
    const onConfirm = jest.fn().mockResolvedValue(wp)
    const onDone = jest.fn()
    const onClose = jest.fn()
    render(
      <DeleteCollectionItemDialog
        open={true}
        itemLabel="Цель 1"
        onConfirm={onConfirm}
        onDone={onDone}
        onClose={onClose}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'collectionDialog.deleteConfirm' }))
    await waitFor(() => expect(onConfirm).toHaveBeenCalled())
    await waitFor(() => expect(onDone).toHaveBeenCalledWith(wp))
    expect(mockToastSuccess).toHaveBeenCalled()
    expect(onClose).toHaveBeenCalled()
  })

  it('keeps the dialog open and shows an error toast when onConfirm rejects', async () => {
    const onConfirm = jest.fn().mockRejectedValue(new Error('boom'))
    const onDone = jest.fn()
    const onClose = jest.fn()
    render(
      <DeleteCollectionItemDialog
        open={true}
        itemLabel="Цель 1"
        onConfirm={onConfirm}
        onDone={onDone}
        onClose={onClose}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'collectionDialog.deleteConfirm' }))
    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(onDone).not.toHaveBeenCalled()
    expect(onClose).not.toHaveBeenCalled()
  })
})
