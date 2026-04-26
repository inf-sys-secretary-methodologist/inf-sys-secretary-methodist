import { render, screen } from '@/test-utils'

jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: jest.fn(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

import SchedulePage from '../page'

describe('SchedulePage', () => {
  it('renders schedule title', () => {
    render(<SchedulePage />)
    expect(screen.getByText('title')).toBeInTheDocument()
  })

  it('shows coming soon message', () => {
    render(<SchedulePage />)
    expect(screen.getByText('comingSoon')).toBeInTheDocument()
  })
})
