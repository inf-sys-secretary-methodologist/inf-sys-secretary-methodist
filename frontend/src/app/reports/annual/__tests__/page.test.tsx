import { render, screen, fireEvent, waitFor } from '@/test-utils'

const mockReplace = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace, push: jest.fn() }),
  useParams: () => ({}),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockDownload = jest.fn()
jest.mock('@/lib/api/annualReport', () => ({
  annualReportApi: {
    download: (year: number) => mockDownload(year),
  },
}))

import AnnualReportPage from '../page'

const methodistAuth = {
  user: { id: 42, role: 'methodist' as const },
  isAuthenticated: true,
  isLoading: false,
}
const adminAuth = {
  user: { id: 1, role: 'system_admin' as const },
  isAuthenticated: true,
  isLoading: false,
}
const teacherAuth = {
  user: { id: 7, role: 'teacher' as const },
  isAuthenticated: true,
  isLoading: false,
}

beforeEach(() => {
  jest.clearAllMocks()
  mockDownload.mockResolvedValue(new Blob(['DOCX'], { type: 'application/octet-stream' }))
})

describe('AnnualReportPage', () => {
  it('renders title, year selector с last 10 years, и download button for methodist', () => {
    mockUseAuthCheck.mockReturnValue(methodistAuth)
    const currentYear = new Date().getFullYear()

    render(<AnnualReportPage />)

    expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument()
    const select = screen.getByRole('combobox') as HTMLSelectElement
    expect(select).toBeInTheDocument()
    expect(select.options).toHaveLength(10)
    expect(select.options[0].value).toBe(String(currentYear))
    expect(select.options[9].value).toBe(String(currentYear - 9))

    expect(
      screen.getByRole('button', { name: /downloadButton|Скачать|Download/i })
    ).toBeInTheDocument()
  })

  it('renders для system_admin as well', () => {
    mockUseAuthCheck.mockReturnValue(adminAuth)

    render(<AnnualReportPage />)
    expect(screen.getByRole('combobox')).toBeInTheDocument()
  })

  it('redirects non-methodist + non-admin role to /forbidden', () => {
    mockUseAuthCheck.mockReturnValue(teacherAuth)

    render(<AnnualReportPage />)

    expect(mockReplace).toHaveBeenCalledWith('/forbidden')
  })

  it('does not redirect while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({ user: null, isAuthenticated: false, isLoading: true })

    render(<AnnualReportPage />)

    expect(mockReplace).not.toHaveBeenCalled()
  })

  it('calls annualReportApi.download(selectedYear) on Download click', async () => {
    mockUseAuthCheck.mockReturnValue(methodistAuth)
    const currentYear = new Date().getFullYear()
    const targetYear = currentYear - 2

    render(<AnnualReportPage />)

    const select = screen.getByRole('combobox') as HTMLSelectElement
    fireEvent.change(select, { target: { value: String(targetYear) } })
    fireEvent.click(screen.getByRole('button', { name: /downloadButton|Скачать|Download/i }))

    await waitFor(() => {
      expect(mockDownload).toHaveBeenCalledWith(targetYear)
    })
  })
})
