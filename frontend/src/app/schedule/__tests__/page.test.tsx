import { render, screen } from '@/test-utils'

jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: jest.fn(),
}))

jest.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({ user: { role: 'academic_secretary' } }),
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
