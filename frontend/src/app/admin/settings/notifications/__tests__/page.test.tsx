import { render, screen } from '@/test-utils'

jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: jest.fn(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

import AdminNotificationsPage from '../page'

describe('AdminNotificationsPage', () => {
  it('renders page title', () => {
    render(<AdminNotificationsPage />)
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('notifications.title')
  })

  it('renders SMTP section', () => {
    render(<AdminNotificationsPage />)
    expect(screen.getByText('notifications.smtpTitle')).toBeInTheDocument()
  })

  it('renders push section', () => {
    render(<AdminNotificationsPage />)
    expect(screen.getByText('notifications.pushTitle')).toBeInTheDocument()
  })

  it('renders telegram section', () => {
    render(<AdminNotificationsPage />)
    expect(screen.getByText('notifications.telegramTitle')).toBeInTheDocument()
  })
})
