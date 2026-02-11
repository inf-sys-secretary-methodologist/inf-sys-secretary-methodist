import { render, screen } from '@testing-library/react'
import { MessageBubble } from '../MessageBubble'
import type { Message } from '@/types/messaging'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      messageDeleted: 'Message deleted',
      edited: 'edited',
      reply: 'Reply',
      copy: 'Copy',
      copied: 'Copied',
      edit: 'Edit',
      delete: 'Delete',
    }
    return translations[key] || key
  },
}))

// Mock clipboard API
Object.assign(navigator, {
  clipboard: {
    writeText: jest.fn().mockResolvedValue(undefined),
  },
})

describe('MessageBubble', () => {
  const baseMessage: Message = {
    id: 1,
    conversation_id: 1,
    sender_id: 1,
    sender_name: 'John Doe',
    content: 'Hello, world!',
    type: 'text',
    created_at: new Date().toISOString(),
    attachments: [],
    is_edited: false,
    is_deleted: false,
  }

  it('renders message content', () => {
    render(<MessageBubble message={baseMessage} isOwn={false} />)
    expect(screen.getByText('Hello, world!')).toBeInTheDocument()
  })

  it('renders own message with primary background', () => {
    render(<MessageBubble message={baseMessage} isOwn={true} />)
    const bubble = screen.getByText('Hello, world!').closest('.rounded-2xl')
    expect(bubble).toHaveClass('bg-primary')
  })

  it('renders other message with muted background', () => {
    render(<MessageBubble message={baseMessage} isOwn={false} />)
    const bubble = screen.getByText('Hello, world!').closest('.rounded-2xl')
    expect(bubble).toHaveClass('bg-muted')
  })

  it('shows avatar for non-own messages when showAvatar is true', () => {
    render(<MessageBubble message={baseMessage} isOwn={false} showAvatar={true} />)
    // Avatar fallback should show initials
    expect(screen.getByText('JD')).toBeInTheDocument()
  })

  it('hides avatar when showAvatar is false', () => {
    render(<MessageBubble message={baseMessage} isOwn={false} showAvatar={false} />)
    expect(screen.queryByText('JD')).not.toBeInTheDocument()
  })

  it('shows sender name when showSender is true and not own message', () => {
    render(<MessageBubble message={baseMessage} isOwn={false} showSender={true} />)
    expect(screen.getByText('John Doe')).toBeInTheDocument()
  })

  it('hides sender name for own messages even when showSender is true', () => {
    render(<MessageBubble message={baseMessage} isOwn={true} showSender={true} />)
    // Name should not appear separately (only in avatar area)
    const nameElements = screen.queryAllByText('John Doe')
    expect(nameElements.length).toBe(0)
  })

  it('shows edited indicator for edited messages', () => {
    render(<MessageBubble message={{ ...baseMessage, is_edited: true }} isOwn={false} />)
    expect(screen.getByText('(edited)')).toBeInTheDocument()
  })

  it('renders deleted message with placeholder text', () => {
    render(<MessageBubble message={{ ...baseMessage, is_deleted: true }} isOwn={false} />)
    expect(screen.getByText('Message deleted')).toBeInTheDocument()
    expect(screen.queryByText('Hello, world!')).not.toBeInTheDocument()
  })

  it('renders system message centered', () => {
    render(<MessageBubble message={{ ...baseMessage, type: 'system' }} isOwn={false} />)
    const container = screen.getByText('Hello, world!').closest('.flex')
    expect(container).toHaveClass('justify-center')
  })

  it('formats time correctly', () => {
    const date = new Date()
    date.setHours(14, 30, 0)
    render(
      <MessageBubble message={{ ...baseMessage, created_at: date.toISOString() }} isOwn={false} />
    )
    // Time should be displayed (format depends on locale)
    const timeElement = screen.getByText(/\d{1,2}:\d{2}/)
    expect(timeElement).toBeInTheDocument()
  })

  it('renders reply-to reference when present', () => {
    const messageWithReply: Message = {
      ...baseMessage,
      reply_to: {
        id: 0,
        conversation_id: 1,
        sender_id: 2,
        content: 'Original message',
        sender_name: 'Jane Doe',
        type: 'text',
        attachments: [],
        created_at: new Date().toISOString(),
        is_edited: false,
        is_deleted: false,
      },
    }
    render(<MessageBubble message={messageWithReply} isOwn={false} />)
    expect(screen.getByText('Jane Doe')).toBeInTheDocument()
    expect(screen.getByText('Original message')).toBeInTheDocument()
  })

  it('shows deleted text for deleted reply-to', () => {
    const messageWithDeletedReply: Message = {
      ...baseMessage,
      reply_to: {
        id: 0,
        conversation_id: 1,
        sender_id: 2,
        content: 'Original message',
        sender_name: 'Jane Doe',
        type: 'text',
        attachments: [],
        created_at: new Date().toISOString(),
        is_edited: false,
        is_deleted: true,
      },
    }
    render(<MessageBubble message={messageWithDeletedReply} isOwn={false} />)
    expect(screen.getByText('Message deleted')).toBeInTheDocument()
  })

  it('renders image attachments', () => {
    const messageWithImage: Message = {
      ...baseMessage,
      attachments: [
        {
          id: 1,
          file_name: 'image.png',
          mime_type: 'image/png',
          file_size: 1024,
          url: 'https://example.com/image.png',
        },
      ],
    }
    render(<MessageBubble message={messageWithImage} isOwn={false} />)
    const image = screen.getByRole('img')
    // Next.js Image optimizes src, so just check it contains the URL
    expect(image.getAttribute('src')).toContain('example.com')
  })

  it('renders file attachments with download link', () => {
    const messageWithFile: Message = {
      ...baseMessage,
      attachments: [
        {
          id: 1,
          file_name: 'document.pdf',
          mime_type: 'application/pdf',
          file_size: 2048,
          url: 'https://example.com/document.pdf',
        },
      ],
    }
    render(<MessageBubble message={messageWithFile} isOwn={false} />)
    expect(screen.getByText('document.pdf')).toBeInTheDocument()
    expect(screen.getByText('2.0 KB')).toBeInTheDocument()
  })

  it('renders dropdown menu trigger button', () => {
    render(<MessageBubble message={baseMessage} isOwn={false} onReply={jest.fn()} />)
    // The more button should be present (even if not visible until hover)
    const moreButton = screen.getByRole('button')
    expect(moreButton).toBeInTheDocument()
  })

  it('accepts onReply callback', () => {
    const onReply = jest.fn()
    render(<MessageBubble message={baseMessage} isOwn={false} onReply={onReply} />)
    // Just verify component renders with callback
    expect(screen.getByText('Hello, world!')).toBeInTheDocument()
  })

  it('accepts onEdit callback', () => {
    const onEdit = jest.fn()
    render(<MessageBubble message={baseMessage} isOwn={true} onEdit={onEdit} />)
    expect(screen.getByText('Hello, world!')).toBeInTheDocument()
  })

  it('accepts onDelete callback', () => {
    const onDelete = jest.fn()
    render(<MessageBubble message={baseMessage} isOwn={true} onDelete={onDelete} />)
    expect(screen.getByText('Hello, world!')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(
      <MessageBubble message={baseMessage} isOwn={false} className="custom-class" />
    )
    expect(container.firstChild).toHaveClass('custom-class')
  })

  it('formats file sizes correctly', () => {
    const bytesAttachment: Message = {
      ...baseMessage,
      attachments: [
        { id: 1, file_name: 'small.txt', mime_type: 'text/plain', file_size: 500, url: '' },
      ],
    }
    const { rerender } = render(<MessageBubble message={bytesAttachment} isOwn={false} />)
    expect(screen.getByText('500 B')).toBeInTheDocument()

    const kbAttachment: Message = {
      ...baseMessage,
      attachments: [
        { id: 1, file_name: 'medium.txt', mime_type: 'text/plain', file_size: 1024 * 5, url: '' },
      ],
    }
    rerender(<MessageBubble message={kbAttachment} isOwn={false} />)
    expect(screen.getByText('5.0 KB')).toBeInTheDocument()

    const mbAttachment: Message = {
      ...baseMessage,
      attachments: [
        {
          id: 1,
          file_name: 'large.txt',
          mime_type: 'text/plain',
          file_size: 1024 * 1024 * 2.5,
          url: '',
        },
      ],
    }
    rerender(<MessageBubble message={mbAttachment} isOwn={false} />)
    expect(screen.getByText('2.5 MB')).toBeInTheDocument()
  })
})
