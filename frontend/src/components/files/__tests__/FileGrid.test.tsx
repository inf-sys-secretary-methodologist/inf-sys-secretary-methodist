import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { FileGrid } from '../FileGrid'
import type { FileItem } from '@/types/files'

const file1: FileItem = {
  id: 1,
  original_name: 'report.pdf',
  size: 1048576,
  mime_type: 'application/pdf',
  checksum: 'abc',
  uploaded_by: 1,
  is_temporary: false,
  created_at: '2026-04-25T10:00:00Z',
  updated_at: '2026-04-25T10:00:00Z',
}

const file2: FileItem = {
  id: 2,
  original_name: 'photo.png',
  size: 2048,
  mime_type: 'image/png',
  checksum: 'def',
  uploaded_by: 2,
  is_temporary: false,
  created_at: '2026-04-25T11:00:00Z',
  updated_at: '2026-04-25T11:00:00Z',
}

describe('FileGrid', () => {
  it('renders empty state when no files', () => {
    render(<FileGrid files={[]} />)
    expect(screen.getByText('noFiles')).toBeInTheDocument()
  })

  it('renders file names', () => {
    render(<FileGrid files={[file1, file2]} />)
    expect(screen.getByText('report.pdf')).toBeInTheDocument()
    expect(screen.getByText('photo.png')).toBeInTheDocument()
  })

  it('calls onDownload when download button is clicked', async () => {
    const onDownload = jest.fn()
    const user = userEvent.setup()

    render(<FileGrid files={[file1]} onDownload={onDownload} />)

    const downloadBtn = screen.getByRole('button', { name: /download/i })
    await user.click(downloadBtn)
    expect(onDownload).toHaveBeenCalledWith(1)
  })

  it('calls onDelete when delete button is clicked', async () => {
    const onDelete = jest.fn()
    const user = userEvent.setup()

    render(<FileGrid files={[file1]} onDelete={onDelete} />)

    const deleteBtn = screen.getByRole('button', { name: /delete/i })
    await user.click(deleteBtn)
    expect(onDelete).toHaveBeenCalledWith(1)
  })

  it('calls onPreview when file row is clicked', async () => {
    const onPreview = jest.fn()
    const user = userEvent.setup()

    render(<FileGrid files={[file1]} onPreview={onPreview} />)

    await user.click(screen.getByText('report.pdf'))
    expect(onPreview).toHaveBeenCalledWith(file1)
  })

  it('calls onVersions when versions button is clicked', async () => {
    const onVersions = jest.fn()
    const user = userEvent.setup()

    render(<FileGrid files={[file1]} onVersions={onVersions} />)

    const versionsBtn = screen.getByRole('button', { name: /versions\.title/i })
    await user.click(versionsBtn)
    expect(onVersions).toHaveBeenCalledWith(1)
  })
})
