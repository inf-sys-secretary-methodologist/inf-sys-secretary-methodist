import { render, screen } from '@/test-utils'

jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: jest.fn(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

import AdminAutomationPage from '../page'

describe('AdminAutomationPage', () => {
  it('renders page title', () => {
    render(<AdminAutomationPage />)
    expect(screen.getByText('automation.title')).toBeInTheDocument()
  })

  it('renders workflows section', () => {
    render(<AdminAutomationPage />)
    expect(screen.getByText('automation.workflowsTitle')).toBeInTheDocument()
  })
})
