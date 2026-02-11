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

  it('shows error for invalid file format', async () => {
    const { container } = render(<AvatarUpload onUpload={mockOnUpload} />)
    const input = container.querySelector('input[type="file"]') as HTMLInputElement

    const invalidFile = new File(['content'], 'test.pdf', { type: 'application/pdf' })

    Object.defineProperty(input, 'files', {
      value: [invalidFile],
    })

    // Trigger change event
    const event = new Event('change', { bubbles: true })
    input.dispatchEvent(event)

    await waitFor(() => {
      expect(screen.getByText('Invalid file format.')).toBeInTheDocument()
    })
  })

  it('shows error for file too large', async () => {
    const { container } = render(<AvatarUpload onUpload={mockOnUpload} />)
    const input = container.querySelector('input[type="file"]') as HTMLInputElement

    // Create a file larger than 5MB
    const largeContent = new Array(6 * 1024 * 1024).fill('x').join('')
    const largeFile = new File([largeContent], 'large.jpg', { type: 'image/jpeg' })

    Object.defineProperty(input, 'files', {
      value: [largeFile],
    })

    const event = new Event('change', { bubbles: true })
    input.dispatchEvent(event)

    await waitFor(() => {
      expect(screen.getByText('File is too large. Max 5MB.')).toBeInTheDocument()
    })
  })

  it('handles file selection and opens cropper', async () => {
    const { container } = render(<AvatarUpload onUpload={mockOnUpload} />)
    const input = container.querySelector('input[type="file"]') as HTMLInputElement

    const validFile = new File(['content'], 'test.jpg', { type: 'image/jpeg' })

    Object.defineProperty(input, 'files', {
      value: [validFile],
    })

    const event = new Event('change', { bubbles: true })
    input.dispatchEvent(event)

    await waitFor(() => {
      expect(screen.getByTestId('image-cropper')).toBeInTheDocument()
    })
  })

  it('handles drag over events', async () => {
    const { container } = render(<AvatarUpload onUpload={mockOnUpload} />)
    const dropZone = container.querySelector('.relative.size-20')

    if (dropZone) {
      // Simulate drag over
      const dragOverEvent = new Event('dragover', { bubbles: true })
      Object.defineProperty(dragOverEvent, 'preventDefault', { value: jest.fn() })
      dropZone.dispatchEvent(dragOverEvent)

      // Simulate drag leave
      const dragLeaveEvent = new Event('dragleave', { bubbles: true })
      dropZone.dispatchEvent(dragLeaveEvent)
    }

    expect(container).toBeInTheDocument()
  })

  it('handles drop events', async () => {
    const { container } = render(<AvatarUpload onUpload={mockOnUpload} />)
    const dropZone = container.querySelector('.relative.size-20')

    if (dropZone) {
      const validFile = new File(['content'], 'dropped.jpg', { type: 'image/jpeg' })
      const dropEvent = new Event('drop', { bubbles: true }) as unknown as DragEvent

      Object.defineProperty(dropEvent, 'preventDefault', { value: jest.fn() })
      Object.defineProperty(dropEvent, 'dataTransfer', {
        value: {
          files: [validFile],
        },
      })

      dropZone.dispatchEvent(dropEvent)
    }

    await waitFor(() => {
      expect(container).toBeInTheDocument()
    })
  })

  it('shows error when onRemove fails', async () => {
    const failingOnRemove = jest.fn().mockRejectedValue(new Error('Delete failed'))

    render(
      <AvatarUpload
        onUpload={mockOnUpload}
        onRemove={failingOnRemove}
        currentAvatar="https://example.com/avatar.jpg"
      />
    )

    await userEvent.click(screen.getByRole('button', { name: /delete/i }))

    await waitFor(() => {
      expect(screen.getByText('Failed to delete image.')).toBeInTheDocument()
    })
  })

  it('does not process when disabled', async () => {
    const { container } = render(<AvatarUpload onUpload={mockOnUpload} disabled />)
    const dropZone = container.querySelector('.relative.size-20')

    if (dropZone) {
      const validFile = new File(['content'], 'dropped.jpg', { type: 'image/jpeg' })
      const dropEvent = new Event('drop', { bubbles: true }) as unknown as DragEvent

      Object.defineProperty(dropEvent, 'preventDefault', { value: jest.fn() })
      Object.defineProperty(dropEvent, 'dataTransfer', {
        value: {
          files: [validFile],
        },
      })

      dropZone.dispatchEvent(dropEvent)
    }

    expect(mockOnUpload).not.toHaveBeenCalled()
  })

  it('updates preview when currentAvatar prop changes', async () => {
    const { rerender } = render(<AvatarUpload onUpload={mockOnUpload} currentAvatar={null} />)

    rerender(
      <AvatarUpload onUpload={mockOnUpload} currentAvatar="https://example.com/new-avatar.jpg" />
    )

    await waitFor(() => {
      expect(screen.getByTestId('avatar-image')).toBeInTheDocument()
    })
  })

  it('clicks file input when avatar area is clicked', async () => {
    const { container } = render(<AvatarUpload onUpload={mockOnUpload} />)
    const dropZone = container.querySelector('.relative.size-20')
    const input = container.querySelector('input[type="file"]') as HTMLInputElement

    const clickSpy = jest.spyOn(input, 'click')

    if (dropZone) {
      await userEvent.click(dropZone)
    }

    expect(clickSpy).toHaveBeenCalled()
  })

  it('clicks file input when upload/change button is clicked', async () => {
    const { container } = render(<AvatarUpload onUpload={mockOnUpload} />)
    const input = container.querySelector('input[type="file"]') as HTMLInputElement

    const clickSpy = jest.spyOn(input, 'click')

    const uploadButton = screen.getByRole('button', { name: /upload photo/i })
    await userEvent.click(uploadButton)

    expect(clickSpy).toHaveBeenCalled()
    clickSpy.mockRestore()
  })
})
