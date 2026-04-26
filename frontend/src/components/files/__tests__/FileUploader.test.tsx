import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { FileUploader } from '../FileUploader'

describe('FileUploader', () => {
  it('renders dropzone with title text', () => {
    render(<FileUploader onUpload={jest.fn()} />)
    expect(screen.getByText(/drag.*drop|перетащите/i)).toBeInTheDocument()
  })

  it('calls onUpload when a file is selected via input', async () => {
    const onUpload = jest.fn().mockResolvedValue(undefined)
    const user = userEvent.setup()

    render(<FileUploader onUpload={onUpload} />)

    const file = new File(['content'], 'test.pdf', { type: 'application/pdf' })
    const input = screen.getByTestId('file-upload-input') as HTMLInputElement
    await user.upload(input, file)

    expect(onUpload).toHaveBeenCalledWith(file)
  })

  it('shows uploading state', () => {
    render(<FileUploader onUpload={jest.fn()} uploading />)
    expect(screen.getByText(/uploading|загрузка/i)).toBeInTheDocument()
  })

  it('disables input while uploading', () => {
    render(<FileUploader onUpload={jest.fn()} uploading />)
    const input = screen.getByTestId('file-upload-input') as HTMLInputElement
    expect(input).toBeDisabled()
  })
})
