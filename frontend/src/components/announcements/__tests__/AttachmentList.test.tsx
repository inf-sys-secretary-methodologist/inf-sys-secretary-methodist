import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { AttachmentList } from '../AttachmentList'
import type { AnnouncementAttachment } from '@/types/announcements'

const att1: AnnouncementAttachment = {
  id: 1,
  file_name: 'report.pdf',
  file_size: 1024,
  mime_type: 'application/pdf',
  created_at: '2026-04-25T10:00:00Z',
}

const att2: AnnouncementAttachment = {
  id: 2,
  file_name: 'image.png',
  file_size: 2048000,
  mime_type: 'image/png',
  created_at: '2026-04-25T10:00:00Z',
}

describe('AttachmentList', () => {
  it('renders empty state when there are no attachments', () => {
    render(<AttachmentList attachments={[]} />)
    expect(screen.getByText(/noAttachments|no attachments|нет вложений/i)).toBeInTheDocument()
  })

  it('renders attachment names and sizes', () => {
    render(<AttachmentList attachments={[att1, att2]} />)
    expect(screen.getByText('report.pdf')).toBeInTheDocument()
    expect(screen.getByText('image.png')).toBeInTheDocument()
    // Sizes formatted as KB / MB
    expect(screen.getByText(/1\.0\s*KB|1024\s*B/i)).toBeInTheDocument()
    expect(screen.getByText(/2\.0\s*MB/i)).toBeInTheDocument()
  })

  it('shows remove button only when onRemove is provided (editable mode)', async () => {
    const onRemove = jest.fn()
    const user = userEvent.setup()
    const { rerender } = render(<AttachmentList attachments={[att1]} />)
    expect(screen.queryByRole('button', { name: /remove|удалить/i })).not.toBeInTheDocument()

    rerender(<AttachmentList attachments={[att1]} onRemove={onRemove} />)
    const btn = screen.getByRole('button', { name: /remove|удалить/i })
    await user.click(btn)
    expect(onRemove).toHaveBeenCalledWith(1)
  })

  it('renders upload zone only when onUpload is provided', async () => {
    const onUpload = jest.fn().mockResolvedValue(undefined)
    const { rerender } = render(<AttachmentList attachments={[]} />)
    expect(screen.queryByLabelText(/uploadFile|upload|загруз/i)).not.toBeInTheDocument()

    rerender(<AttachmentList attachments={[]} onUpload={onUpload} />)
    expect(screen.getByLabelText(/uploadFile|upload|загруз/i)).toBeInTheDocument()
  })

  it('calls onUpload when a file is selected', async () => {
    const onUpload = jest.fn().mockResolvedValue(undefined)
    const user = userEvent.setup()
    render(<AttachmentList attachments={[]} onUpload={onUpload} />)

    const file = new File(['hi'], 'x.txt', { type: 'text/plain' })
    const input = screen.getByLabelText(/uploadFile|upload|загруз/i) as HTMLInputElement
    await user.upload(input, file)

    expect(onUpload).toHaveBeenCalledWith(file)
  })
})
