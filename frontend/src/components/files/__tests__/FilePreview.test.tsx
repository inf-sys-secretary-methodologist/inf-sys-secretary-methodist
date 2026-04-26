import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { FilePreview } from '../FilePreview'

describe('FilePreview', () => {
  it('renders image preview for image mime types', () => {
    render(
      <FilePreview
        fileName="photo.png"
        mimeType="image/png"
        downloadUrl="https://minio/photo.png"
        onClose={jest.fn()}
      />
    )
    const img = screen.getByRole('img')
    expect(img).toHaveAttribute('src', 'https://minio/photo.png')
    expect(img).toHaveAttribute('alt', 'photo.png')
  })

  it('renders iframe for PDF files', () => {
    render(
      <FilePreview
        fileName="doc.pdf"
        mimeType="application/pdf"
        downloadUrl="https://minio/doc.pdf"
        onClose={jest.fn()}
      />
    )
    const iframe = screen.getByTitle('doc.pdf')
    expect(iframe).toHaveAttribute('src', 'https://minio/doc.pdf')
  })

  it('shows no-preview message for unsupported types', () => {
    render(
      <FilePreview
        fileName="data.zip"
        mimeType="application/zip"
        downloadUrl="https://minio/data.zip"
        onClose={jest.fn()}
      />
    )
    expect(screen.getByText('preview.noPreview')).toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', async () => {
    const onClose = jest.fn()
    const user = userEvent.setup()

    render(
      <FilePreview
        fileName="photo.png"
        mimeType="image/png"
        downloadUrl="https://minio/photo.png"
        onClose={onClose}
      />
    )

    const closeBtn = screen.getByRole('button', { name: /preview\.close/i })
    await user.click(closeBtn)
    expect(onClose).toHaveBeenCalled()
  })

  it('displays file name in header', () => {
    render(
      <FilePreview
        fileName="report.pdf"
        mimeType="application/pdf"
        downloadUrl="https://minio/report.pdf"
        onClose={jest.fn()}
      />
    )
    expect(screen.getByText('report.pdf')).toBeInTheDocument()
  })
})
