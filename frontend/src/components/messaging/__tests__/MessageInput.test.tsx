import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MessageInput } from '../MessageInput'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      typeMessage: 'Type a message...',
      voiceMessage: 'Voice message',
      attach: 'Attach',
      emoji: 'Emoji',
      send: 'Send',
      'suggestions.scheduleEvent': 'Schedule event',
      'suggestions.createTask': 'Create task',
      'suggestions.shareDocument': 'Share document',
      'suggestions.askAI': 'Ask AI',
    }
    return translations[key] || key
  },
}))

describe('MessageInput', () => {
  const mockOnSend = jest.fn().mockResolvedValue(undefined)

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders textarea with placeholder', () => {
    render(<MessageInput onSend={mockOnSend} />)
    expect(screen.getByPlaceholderText('Type a message...')).toBeInTheDocument()
  })

  it('renders with custom placeholder', () => {
    render(<MessageInput onSend={mockOnSend} placeholder="Custom placeholder" />)
    expect(screen.getByPlaceholderText('Custom placeholder')).toBeInTheDocument()
  })

  it('renders send button', () => {
    render(<MessageInput onSend={mockOnSend} />)
    expect(screen.getByRole('button', { name: /send/i })).toBeInTheDocument()
  })

  it('renders attach button', () => {
    render(<MessageInput onSend={mockOnSend} />)
    expect(screen.getByRole('button', { name: /attach/i })).toBeInTheDocument()
  })

  it('renders emoji button', () => {
    render(<MessageInput onSend={mockOnSend} />)
    expect(screen.getByRole('button', { name: /emoji/i })).toBeInTheDocument()
  })

  it('send button is disabled when textarea is empty', () => {
    render(<MessageInput onSend={mockOnSend} />)
    expect(screen.getByRole('button', { name: /send/i })).toBeDisabled()
  })

  it('send button is enabled when textarea has content', async () => {
    render(<MessageInput onSend={mockOnSend} />)
    const textarea = screen.getByPlaceholderText('Type a message...')

    await userEvent.type(textarea, 'Hello')

    expect(screen.getByRole('button', { name: /send/i })).not.toBeDisabled()
  })

  it('calls onSend when send button is clicked', async () => {
    render(<MessageInput onSend={mockOnSend} />)
    const textarea = screen.getByPlaceholderText('Type a message...')

    await userEvent.type(textarea, 'Hello World')
    await userEvent.click(screen.getByRole('button', { name: /send/i }))

    await waitFor(() => {
      expect(mockOnSend).toHaveBeenCalledWith('Hello World', undefined)
    })
  })

  it('calls onSend when Enter is pressed', async () => {
    render(<MessageInput onSend={mockOnSend} />)
    const textarea = screen.getByPlaceholderText('Type a message...')

    await userEvent.type(textarea, 'Test message{Enter}')

    await waitFor(() => {
      expect(mockOnSend).toHaveBeenCalledWith('Test message', undefined)
    })
  })

  it('does not send on Shift+Enter', async () => {
    render(<MessageInput onSend={mockOnSend} />)
    const textarea = screen.getByPlaceholderText('Type a message...')

    fireEvent.change(textarea, { target: { value: 'Test' } })
    fireEvent.keyDown(textarea, { key: 'Enter', shiftKey: true })

    expect(mockOnSend).not.toHaveBeenCalled()
  })

  it('clears textarea after sending', async () => {
    render(<MessageInput onSend={mockOnSend} />)
    const textarea = screen.getByPlaceholderText('Type a message...') as HTMLTextAreaElement

    await userEvent.type(textarea, 'Message to clear')
    await userEvent.click(screen.getByRole('button', { name: /send/i }))

    await waitFor(() => {
      expect(textarea.value).toBe('')
    })
  })

  it('disables input when disabled prop is true', () => {
    render(<MessageInput onSend={mockOnSend} disabled />)
    expect(screen.getByPlaceholderText('Type a message...')).toBeDisabled()
  })

  it('calls onTyping when typing', async () => {
    const onTyping = jest.fn()
    render(<MessageInput onSend={mockOnSend} onTyping={onTyping} />)
    const textarea = screen.getByPlaceholderText('Type a message...')

    await userEvent.type(textarea, 'H')

    expect(onTyping).toHaveBeenCalled()
  })

  it('shows AI suggestions when enabled and input is empty', () => {
    render(<MessageInput onSend={mockOnSend} showAiSuggestions />)
    expect(screen.getByText('Schedule event')).toBeInTheDocument()
    expect(screen.getByText('Create task')).toBeInTheDocument()
    expect(screen.getByText('Share document')).toBeInTheDocument()
    expect(screen.getByText('Ask AI')).toBeInTheDocument()
  })

  it('hides AI suggestions when input has content', async () => {
    render(<MessageInput onSend={mockOnSend} showAiSuggestions />)
    const textarea = screen.getByPlaceholderText('Type a message...')

    await userEvent.type(textarea, 'Some content')

    expect(screen.queryByText('Schedule event')).not.toBeInTheDocument()
  })

  it('fills input when AI suggestion is clicked', async () => {
    render(<MessageInput onSend={mockOnSend} showAiSuggestions />)

    await userEvent.click(screen.getByText('Schedule event'))

    const textarea = screen.getByPlaceholderText('Type a message...') as HTMLTextAreaElement
    expect(textarea.value).toBe('Давай запланируем встречу...')
  })

  it('applies custom className', () => {
    const { container } = render(
      <MessageInput onSend={mockOnSend} className="custom-input-class" />
    )
    expect(container.firstChild).toHaveClass('custom-input-class')
  })

  it('calls onStopTyping after sending', async () => {
    const onStopTyping = jest.fn()
    render(<MessageInput onSend={mockOnSend} onStopTyping={onStopTyping} />)
    const textarea = screen.getByPlaceholderText('Type a message...')

    await userEvent.type(textarea, 'Test message')
    await userEvent.click(screen.getByRole('button', { name: /send/i }))

    await waitFor(() => {
      expect(onStopTyping).toHaveBeenCalled()
    })
  })

  it('shows suggestions when input is cleared', async () => {
    render(<MessageInput onSend={mockOnSend} showAiSuggestions />)
    const textarea = screen.getByPlaceholderText('Type a message...')

    // Type something - suggestions should hide
    await userEvent.type(textarea, 'test')
    expect(screen.queryByText('Schedule event')).not.toBeInTheDocument()

    // Clear input - suggestions should reappear
    await userEvent.clear(textarea)
    expect(screen.getByText('Schedule event')).toBeInTheDocument()
  })

  it('does not send empty message', async () => {
    render(<MessageInput onSend={mockOnSend} />)
    const textarea = screen.getByPlaceholderText('Type a message...')

    // Type whitespace only
    fireEvent.change(textarea, { target: { value: '   ' } })

    // Send button should still be disabled
    expect(screen.getByRole('button', { name: /send/i })).toBeDisabled()
  })

  it('calls onStopTyping after typing stops (timeout)', async () => {
    jest.useFakeTimers()
    const onStopTyping = jest.fn()
    const onTyping = jest.fn()
    render(<MessageInput onSend={mockOnSend} onTyping={onTyping} onStopTyping={onStopTyping} />)
    const textarea = screen.getByPlaceholderText('Type a message...')

    fireEvent.change(textarea, { target: { value: 'H' } })

    // Advance timers by 2 seconds
    jest.advanceTimersByTime(2000)

    expect(onStopTyping).toHaveBeenCalled()
    jest.useRealTimers()
  })

  it('does not call onSend when already sending', async () => {
    const slowOnSend = jest
      .fn()
      .mockImplementation(() => new Promise((resolve) => setTimeout(resolve, 1000)))
    render(<MessageInput onSend={slowOnSend} />)
    const textarea = screen.getByPlaceholderText('Type a message...')

    await userEvent.type(textarea, 'Test')

    // Click send twice quickly
    const sendButton = screen.getByRole('button', { name: /send/i })
    await userEvent.click(sendButton)

    // Second click should be ignored
    await waitFor(() => {
      expect(slowOnSend).toHaveBeenCalledTimes(1)
    })
  })
})
