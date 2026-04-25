import { render, screen, waitFor } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { AnnouncementForm } from '../AnnouncementForm'
import type { Announcement } from '@/types/announcements'

const baseAnnouncement: Announcement = {
  id: 1,
  title: 'Old title',
  content: 'Old content',
  summary: 'Old summary',
  author_id: 1,
  status: 'draft',
  priority: 'normal',
  target_audience: 'all',
  is_pinned: false,
  view_count: 0,
  tags: ['existing'],
  created_at: '2026-04-25T10:00:00Z',
  updated_at: '2026-04-25T10:00:00Z',
}

describe('AnnouncementForm', () => {
  it('renders required inputs', () => {
    render(<AnnouncementForm onSubmit={jest.fn()} onCancel={jest.fn()} />)
    expect(screen.getByLabelText(/titleLabel|title/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/contentLabel|content/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/summaryLabel|summary/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/priorityLabel|priority/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/audienceLabel|audience/i)).toBeInTheDocument()
  })

  it('pre-fills fields when editing existing announcement', () => {
    render(<AnnouncementForm announcement={baseAnnouncement} onSubmit={jest.fn()} onCancel={jest.fn()} />)
    expect((screen.getByLabelText(/titleLabel|title/i) as HTMLInputElement).value).toBe('Old title')
    expect((screen.getByLabelText(/contentLabel|content/i) as HTMLTextAreaElement).value).toBe(
      'Old content'
    )
    expect((screen.getByLabelText(/summaryLabel|summary/i) as HTMLTextAreaElement).value).toBe(
      'Old summary'
    )
  })

  it('calls onSubmit with form data when submitted', async () => {
    const onSubmit = jest.fn().mockResolvedValue(undefined)
    const user = userEvent.setup()
    render(<AnnouncementForm onSubmit={onSubmit} onCancel={jest.fn()} />)

    await user.type(screen.getByLabelText(/titleLabel|title/i), 'New title')
    await user.type(screen.getByLabelText(/contentLabel|content/i), 'Body')
    await user.selectOptions(screen.getByLabelText(/priorityLabel|priority/i), 'urgent')
    await user.selectOptions(screen.getByLabelText(/audienceLabel|audience/i), 'students')
    await user.click(screen.getByRole('button', { name: /save|сохран/i }))

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalled()
    })
    const submitted = onSubmit.mock.calls[0][0]
    expect(submitted.title).toBe('New title')
    expect(submitted.content).toBe('Body')
    expect(submitted.priority).toBe('urgent')
    expect(submitted.target_audience).toBe('students')
  })

  it('blocks submit when title is empty', async () => {
    const onSubmit = jest.fn()
    const user = userEvent.setup()
    render(<AnnouncementForm onSubmit={onSubmit} onCancel={jest.fn()} />)

    await user.type(screen.getByLabelText(/contentLabel|content/i), 'Body')
    await user.click(screen.getByRole('button', { name: /save|сохран/i }))

    expect(onSubmit).not.toHaveBeenCalled()
    expect(screen.getByText(/titleRequired|обязательно|required/i)).toBeInTheDocument()
  })

  it('blocks submit when content is empty', async () => {
    const onSubmit = jest.fn()
    const user = userEvent.setup()
    render(<AnnouncementForm onSubmit={onSubmit} onCancel={jest.fn()} />)

    await user.type(screen.getByLabelText(/titleLabel|title/i), 'Title')
    await user.click(screen.getByRole('button', { name: /save|сохран/i }))

    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('toggles is_pinned', async () => {
    const onSubmit = jest.fn().mockResolvedValue(undefined)
    const user = userEvent.setup()
    render(<AnnouncementForm onSubmit={onSubmit} onCancel={jest.fn()} />)

    await user.type(screen.getByLabelText(/titleLabel|title/i), 'T')
    await user.type(screen.getByLabelText(/contentLabel|content/i), 'C')
    await user.click(screen.getByLabelText(/pinned|закреп/i))
    await user.click(screen.getByRole('button', { name: /save|сохран/i }))

    await waitFor(() => expect(onSubmit).toHaveBeenCalled())
    expect(onSubmit.mock.calls[0][0].is_pinned).toBe(true)
  })

  it('calls onCancel when cancel button is clicked', async () => {
    const onCancel = jest.fn()
    const user = userEvent.setup()
    render(<AnnouncementForm onSubmit={jest.fn()} onCancel={onCancel} />)
    await user.click(screen.getByRole('button', { name: /cancel|отмен/i }))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })
})
