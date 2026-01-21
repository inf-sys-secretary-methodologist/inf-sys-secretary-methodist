import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AvatarUpload } from '../AvatarUpload'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      avatar: 'Avatar',
      changePhoto: 'Change photo',
      uploadPhoto: 'Upload photo',
      delete: 'Delete',
      formatHint: 'JPG, PNG, GIF or WebP. Max 5MB.',
      sizeError: 'File is too large. Max 5MB.',
      formatError: 'Invalid file format.',
      uploadError: 'Failed to upload image.',
      deleteError: 'Failed to delete image.',
    }
    return translations[key] || key
  },
}))

// Mock next/dynamic
jest.mock('next/dynamic', () => () => {
  const MockImageCropper = () => <div data-testid="image-cropper">Image Cropper</div>
  return MockImageCropper
})

// Mock next/image
jest.mock('next/image', () => ({
  __esModule: true,
  // eslint-disable-next-line @next/next/no-img-element
  default: function MockImage({ alt, src }: { alt: string; src: string }) {
    // eslint-disable-next-line @next/next/no-img-element
    return <img alt={alt} src={src} data-testid="avatar-image" />
  },
}))

describe('AvatarUpload', () => {
  const mockOnUpload = jest.fn().mockResolvedValue(undefined)
  const mockOnRemove = jest.fn().mockResolvedValue(undefined)

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders upload button', () => {
    render(<AvatarUpload onUpload={mockOnUpload} />)
    expect(screen.getByRole('button', { name: /upload photo/i })).toBeInTheDocument()
  })

  it('shows change photo button when avatar exists', () => {
    render(<AvatarUpload onUpload={mockOnUpload} currentAvatar="https://example.com/avatar.jpg" />)
    expect(screen.getByRole('button', { name: /change photo/i })).toBeInTheDocument()
  })

  it('shows delete button when avatar exists and onRemove is provided', () => {
    render(
      <AvatarUpload
        onUpload={mockOnUpload}
        onRemove={mockOnRemove}
        currentAvatar="https://example.com/avatar.jpg"
      />
    )
    expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument()
  })

  it('shows user initials when userName is provided', () => {
    render(<AvatarUpload onUpload={mockOnUpload} userName="John Doe" />)
    expect(screen.getByText('JD')).toBeInTheDocument()
  })

  it('shows format hint', () => {
    render(<AvatarUpload onUpload={mockOnUpload} />)
    expect(screen.getByText(/JPG, PNG, GIF or WebP/)).toBeInTheDocument()
  })

  it('disables buttons when disabled prop is true', () => {
    render(<AvatarUpload onUpload={mockOnUpload} disabled />)
    expect(screen.getByRole('button', { name: /upload photo/i })).toBeDisabled()
  })

  it('applies custom className', () => {
    const { container } = render(<AvatarUpload onUpload={mockOnUpload} className="custom-class" />)
    expect(container.firstChild).toHaveClass('custom-class')
  })

  it('shows avatar image when currentAvatar is provided', () => {
    render(<AvatarUpload onUpload={mockOnUpload} currentAvatar="https://example.com/avatar.jpg" />)
    expect(screen.getByTestId('avatar-image')).toBeInTheDocument()
  })

  it('renders hidden file input', () => {
    const { container } = render(<AvatarUpload onUpload={mockOnUpload} />)
    const input = container.querySelector('input[type="file"]')
    expect(input).toBeInTheDocument()
    expect(input).toHaveClass('hidden')
  })

  it('accepts correct image formats', () => {
    const { container } = render(<AvatarUpload onUpload={mockOnUpload} />)
    const input = container.querySelector('input[type="file"]')
    expect(input).toHaveAttribute('accept', 'image/jpeg,image/png,image/gif,image/webp')
  })

  it('calls onRemove when delete button is clicked', async () => {
    render(
      <AvatarUpload
        onUpload={mockOnUpload}
        onRemove={mockOnRemove}
        currentAvatar="https://example.com/avatar.jpg"
      />
    )

    await userEvent.click(screen.getByRole('button', { name: /delete/i }))

    await waitFor(() => {
      expect(mockOnRemove).toHaveBeenCalled()
    })
  })
})
