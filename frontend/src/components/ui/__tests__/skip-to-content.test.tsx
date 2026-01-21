import { render, screen } from '@testing-library/react'
import { SkipToContent } from '../skip-to-content'
import { NextIntlClientProvider } from 'next-intl'

const messages = {
  skipToContent: 'Skip to content',
}

const renderWithIntl = (ui: React.ReactElement) => {
  return render(
    <NextIntlClientProvider locale="en" messages={messages}>
      {ui}
    </NextIntlClientProvider>
  )
}

describe('SkipToContent', () => {
  it('renders skip link', () => {
    renderWithIntl(<SkipToContent />)
    const link = screen.getByRole('link')
    expect(link).toBeInTheDocument()
    // Translation key or translated text should be present
    expect(link.textContent).toBeTruthy()
  })

  it('links to main-content by default', () => {
    renderWithIntl(<SkipToContent />)
    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', '#main-content')
  })

  it('links to custom content ID', () => {
    renderWithIntl(<SkipToContent contentId="custom-content" />)
    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', '#custom-content')
  })

  it('applies custom className', () => {
    renderWithIntl(<SkipToContent className="custom-class" />)
    const link = screen.getByRole('link')
    expect(link).toHaveClass('custom-class')
  })

  it('has sr-only class for screen readers', () => {
    renderWithIntl(<SkipToContent />)
    const link = screen.getByRole('link')
    expect(link).toHaveClass('sr-only')
  })
})
