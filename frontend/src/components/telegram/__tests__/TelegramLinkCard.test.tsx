import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TelegramLinkCard } from '../TelegramLinkCard'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: { time?: string }) => {
    const translations: Record<string, string> = {
      linkDescription: 'Connect your Telegram account to receive notifications',
      linkInstructions: 'Click the button below to get a verification code',
      getCode: 'Get Code',
      generating: 'Generating...',
      yourCode: 'Your verification code',
      expiresIn: `Expires in ${params?.time}`,
      copyCode: 'Copy code',
      codeCopied: 'Code copied!',
      howToLink: 'How to link',
      step1: 'Open the bot in Telegram',
      step2: 'Send the verification code',
      step3: 'Done! You will receive notifications',
      openBot: 'Open Bot',
      getNewCode: 'Get new code',
      cancel: 'Cancel',
      accountLinked: 'Your Telegram account is linked',
      connected: 'Connected',
      name: 'Name',
      connectedAt: 'Connected at',
      disconnect: 'Disconnect',
      disconnectTitle: 'Disconnect Telegram?',
      disconnectDescription: 'You will stop receiving notifications',
      confirmDisconnect: 'Disconnect',
      disconnected: 'Telegram disconnected',
      generateCodeError: 'Failed to generate code',
      copyError: 'Failed to copy',
      disconnectError: 'Failed to disconnect',
    }
    return translations[key] || key
  },
}))

// Mock sonner
jest.mock('sonner', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

// Mock telegram hooks
const mockStatus = {
  connected: false,
  first_name: null as string | null,
  username: null as string | null,
  connected_at: null as string | null,
}

const mockGenerateCode = {
  data: null as null | { code: string; bot_link: string; expires_at: string },
  isPending: false,
  mutateAsync: jest.fn(),
  reset: jest.fn(),
}

const mockDisconnectTelegram = {
  mutateAsync: jest.fn(),
}

jest.mock('@/hooks/useTelegram', () => ({
  useTelegramStatus: () => ({
    data: mockStatus,
    isLoading: false,
    mutate: jest.fn(),
  }),
  useGenerateVerificationCode: () => mockGenerateCode,
  useDisconnectTelegram: () => mockDisconnectTelegram,
}))

describe('TelegramLinkCard', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockStatus.connected = false
    mockStatus.first_name = null
    mockStatus.username = null
    mockStatus.connected_at = null
    mockGenerateCode.data = null
    mockGenerateCode.isPending = false
  })

  it('renders telegram card with title', () => {
    render(<TelegramLinkCard />)
    expect(screen.getByText('Telegram')).toBeInTheDocument()
  })

  it('shows link description when not connected', () => {
    render(<TelegramLinkCard />)
    expect(
      screen.getByText('Connect your Telegram account to receive notifications')
    ).toBeInTheDocument()
  })

  it('shows get code button when not connected', () => {
    render(<TelegramLinkCard />)
    expect(screen.getByRole('button', { name: /get code/i })).toBeInTheDocument()
  })

  it('calls generate code when button is clicked', async () => {
    mockGenerateCode.mutateAsync.mockResolvedValue({
      code: 'ABC123',
      bot_link: 'https://t.me/testbot',
      expires_at: new Date(Date.now() + 300000).toISOString(),
    })

    render(<TelegramLinkCard />)

    await userEvent.click(screen.getByRole('button', { name: /get code/i }))

    expect(mockGenerateCode.mutateAsync).toHaveBeenCalled()
  })

  it('shows verification code when generated', () => {
    mockGenerateCode.data = {
      code: 'ABC123',
      bot_link: 'https://t.me/testbot',
      expires_at: new Date(Date.now() + 300000).toISOString(),
    }

    render(<TelegramLinkCard />)

    expect(screen.getByText('ABC123')).toBeInTheDocument()
    expect(screen.getByText('Your verification code')).toBeInTheDocument()
  })

  it('shows instructions when code is generated', () => {
    mockGenerateCode.data = {
      code: 'ABC123',
      bot_link: 'https://t.me/testbot',
      expires_at: new Date(Date.now() + 300000).toISOString(),
    }

    render(<TelegramLinkCard />)

    expect(screen.getByText('How to link')).toBeInTheDocument()
    expect(screen.getByText('Open the bot in Telegram')).toBeInTheDocument()
    expect(screen.getByText('Send the verification code')).toBeInTheDocument()
  })

  it('shows open bot button when code is generated', () => {
    mockGenerateCode.data = {
      code: 'ABC123',
      bot_link: 'https://t.me/testbot',
      expires_at: new Date(Date.now() + 300000).toISOString(),
    }

    render(<TelegramLinkCard />)

    expect(screen.getByRole('button', { name: /open bot/i })).toBeInTheDocument()
  })

  it('shows connected status when linked', () => {
    mockStatus.connected = true
    mockStatus.first_name = 'John'
    mockStatus.username = 'johndoe'
    mockStatus.connected_at = new Date().toISOString()

    render(<TelegramLinkCard />)

    expect(screen.getByText('Connected')).toBeInTheDocument()
    expect(screen.getByText('John')).toBeInTheDocument()
    expect(screen.getByText('@johndoe')).toBeInTheDocument()
  })

  it('shows disconnect button when connected', () => {
    mockStatus.connected = true
    mockStatus.first_name = 'John'

    render(<TelegramLinkCard />)

    expect(screen.getByRole('button', { name: /disconnect/i })).toBeInTheDocument()
  })

  it('shows cancel button when code is displayed', () => {
    mockGenerateCode.data = {
      code: 'ABC123',
      bot_link: 'https://t.me/testbot',
      expires_at: new Date(Date.now() + 300000).toISOString(),
    }

    render(<TelegramLinkCard />)

    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
  })

  it('resets code when cancel is clicked', async () => {
    mockGenerateCode.data = {
      code: 'ABC123',
      bot_link: 'https://t.me/testbot',
      expires_at: new Date(Date.now() + 300000).toISOString(),
    }

    render(<TelegramLinkCard />)

    await userEvent.click(screen.getByRole('button', { name: /cancel/i }))

    expect(mockGenerateCode.reset).toHaveBeenCalled()
  })
})
