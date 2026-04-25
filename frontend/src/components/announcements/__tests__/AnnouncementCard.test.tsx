import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { AnnouncementCard } from '../AnnouncementCard'
import {
  ANNOUNCEMENT_STATUSES,
  ANNOUNCEMENT_PRIORITIES,
  type Announcement,
  type AnnouncementStatus,
  type AnnouncementPriority,
} from '@/types/announcements'

const baseAnnouncement: Announcement = {
  id: 1,
  title: 'Важное объявление',
  content: 'Текст объявления для всех студентов',
  summary: 'Краткое описание',
  author_id: 1,
  author: { id: 1, name: 'Иван Петров', email: 'ivan@example.com' },
  status: 'published',
  priority: 'high',
  target_audience: 'students',
  is_pinned: false,
  view_count: 42,
  tags: ['важное', 'студенты'],
  created_at: '2026-04-25T10:00:00Z',
  updated_at: '2026-04-25T10:00:00Z',
}

describe('AnnouncementCard', () => {
  it('renders title, summary, and content', () => {
    render(<AnnouncementCard announcement={baseAnnouncement} />)
    expect(screen.getByText('Важное объявление')).toBeInTheDocument()
    expect(screen.getByText('Краткое описание')).toBeInTheDocument()
  })

  it('shows priority and status badges', () => {
    render(<AnnouncementCard announcement={baseAnnouncement} />)
    expect(screen.getByText('priority.high')).toBeInTheDocument()
    expect(screen.getByText('status.published')).toBeInTheDocument()
  })

  it('shows target audience badge', () => {
    render(<AnnouncementCard announcement={baseAnnouncement} />)
    expect(screen.getByText('audience.students')).toBeInTheDocument()
  })

  it('renders author name', () => {
    render(<AnnouncementCard announcement={baseAnnouncement} />)
    expect(screen.getByText('Иван Петров')).toBeInTheDocument()
  })

  it('renders view count', () => {
    render(<AnnouncementCard announcement={baseAnnouncement} />)
    expect(screen.getByText(/42/)).toBeInTheDocument()
  })

  it('renders tags', () => {
    render(<AnnouncementCard announcement={baseAnnouncement} />)
    expect(screen.getByText(/важное/i)).toBeInTheDocument()
    expect(screen.getByText(/студенты/i)).toBeInTheDocument()
  })

  it('shows pinned indicator when is_pinned=true', () => {
    render(<AnnouncementCard announcement={{ ...baseAnnouncement, is_pinned: true }} />)
    expect(screen.getByTestId('announcement-pinned-indicator')).toBeInTheDocument()
  })

  it('does not show pinned indicator when is_pinned=false', () => {
    render(<AnnouncementCard announcement={baseAnnouncement} />)
    expect(screen.queryByTestId('announcement-pinned-indicator')).not.toBeInTheDocument()
  })

  it('shows attachment count when attachments are present', () => {
    render(
      <AnnouncementCard
        announcement={{
          ...baseAnnouncement,
          attachments: [
            { id: 1, file_name: 'doc.pdf', file_size: 1024, mime_type: 'application/pdf', created_at: '' },
            { id: 2, file_name: 'pic.png', file_size: 2048, mime_type: 'image/png', created_at: '' },
          ],
        }}
      />
    )
    expect(screen.getByTestId('announcement-attachment-count')).toHaveTextContent('2')
  })

  it('calls onClick when title is clicked', async () => {
    const onClick = jest.fn()
    const user = userEvent.setup()
    render(<AnnouncementCard announcement={baseAnnouncement} onClick={onClick} />)
    await user.click(screen.getByText('Важное объявление'))
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  // Table-driven for status × 3
  it.each(ANNOUNCEMENT_STATUSES)('renders status badge for status=%s', (status: AnnouncementStatus) => {
    render(<AnnouncementCard announcement={{ ...baseAnnouncement, status }} />)
    expect(screen.getByText(`status.${status}`)).toBeInTheDocument()
  })

  // Table-driven for priority × 4
  it.each(ANNOUNCEMENT_PRIORITIES)(
    'renders priority badge for priority=%s',
    (priority: AnnouncementPriority) => {
      render(<AnnouncementCard announcement={{ ...baseAnnouncement, priority }} />)
      expect(screen.getByText(`priority.${priority}`)).toBeInTheDocument()
    }
  )
})
