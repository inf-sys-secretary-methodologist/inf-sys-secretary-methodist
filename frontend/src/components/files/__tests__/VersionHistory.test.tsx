import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { VersionHistory } from '../VersionHistory'
import type { FileVersion } from '@/types/files'

const v1: FileVersion = {
  id: 1,
  version_number: 1,
  size: 1024,
  checksum: 'abc',
  created_by: 1,
  created_at: '2026-04-25T10:00:00Z',
}

const v2: FileVersion = {
  id: 2,
  version_number: 2,
  size: 2048,
  checksum: 'def',
  comment: 'Fixed typo',
  created_by: 1,
  created_at: '2026-04-25T11:00:00Z',
}

describe('VersionHistory', () => {
  it('renders empty state when no versions', () => {
    render(<VersionHistory versions={[]} />)
    expect(screen.getByText('versions.noVersions')).toBeInTheDocument()
  })

  it('renders version numbers', () => {
    render(<VersionHistory versions={[v1, v2]} />)
    expect(screen.getByText(/versions\.version/)).toBeInTheDocument()
  })

  it('renders version comment when present', () => {
    render(<VersionHistory versions={[v1, v2]} />)
    expect(screen.getByText('Fixed typo')).toBeInTheDocument()
  })

  it('calls onDownload when download button is clicked', async () => {
    const onDownload = jest.fn()
    const user = userEvent.setup()

    render(<VersionHistory versions={[v1]} onDownload={onDownload} />)

    const btn = screen.getByRole('button', { name: /versions\.download/i })
    await user.click(btn)
    expect(onDownload).toHaveBeenCalledWith(1)
  })

  it('shows title', () => {
    render(<VersionHistory versions={[v1]} />)
    expect(screen.getByText('versions.title')).toBeInTheDocument()
  })
})
