import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DataSourceSelector } from '../DataSourceSelector'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      dataSource: 'Data Source',
      'sources.documents': 'Documents',
      'sources.users': 'Users',
      'sources.events': 'Events',
      'sources.tasks': 'Tasks',
      'sources.students': 'Students',
    }
    return translations[key] || key
  },
}))

// Mock GlowingEffect
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => <div data-testid="glowing-effect" />,
}))

describe('DataSourceSelector', () => {
  const mockOnChange = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders data source label', () => {
    render(<DataSourceSelector selected="documents" onChange={mockOnChange} />)
    expect(screen.getByText('Data Source')).toBeInTheDocument()
  })

  it('renders all data source options', () => {
    render(<DataSourceSelector selected="documents" onChange={mockOnChange} />)

    expect(screen.getByText('Documents')).toBeInTheDocument()
    expect(screen.getByText('Users')).toBeInTheDocument()
    expect(screen.getByText('Events')).toBeInTheDocument()
    expect(screen.getByText('Tasks')).toBeInTheDocument()
    expect(screen.getByText('Students')).toBeInTheDocument()
  })

  it('highlights selected option', () => {
    render(<DataSourceSelector selected="documents" onChange={mockOnChange} />)

    const documentsButton = screen.getByText('Documents').closest('button')
    expect(documentsButton).toHaveClass('bg-gray-900')
  })

  it('calls onChange when option is clicked', async () => {
    render(<DataSourceSelector selected="documents" onChange={mockOnChange} />)

    await userEvent.click(screen.getByText('Users'))

    expect(mockOnChange).toHaveBeenCalledWith('users')
  })

  it('renders GlowingEffect', () => {
    render(<DataSourceSelector selected="documents" onChange={mockOnChange} />)
    expect(screen.getByTestId('glowing-effect')).toBeInTheDocument()
  })

  it('updates when selection changes', () => {
    const { rerender } = render(<DataSourceSelector selected="documents" onChange={mockOnChange} />)

    let documentsButton = screen.getByText('Documents').closest('button')
    expect(documentsButton).toHaveClass('bg-gray-900')

    rerender(<DataSourceSelector selected="users" onChange={mockOnChange} />)

    const usersButton = screen.getByText('Users').closest('button')
    expect(usersButton).toHaveClass('bg-gray-900')

    documentsButton = screen.getByText('Documents').closest('button')
    expect(documentsButton).not.toHaveClass('bg-gray-900')
  })

  it('renders 5 buttons for all data sources', () => {
    render(<DataSourceSelector selected="documents" onChange={mockOnChange} />)
    const buttons = screen.getAllByRole('button')
    expect(buttons).toHaveLength(5)
  })
})
