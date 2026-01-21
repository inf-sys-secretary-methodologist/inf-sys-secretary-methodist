import { render, screen } from '@testing-library/react'
import { TextDiff, TextDiffSideBySide, TextDiffInline } from '../TextDiff'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      before: 'Before',
      after: 'After',
      removed: 'Removed',
      added: 'Added',
      noChanges: 'No changes detected',
      versionA: 'Version A',
      versionB: 'Version B',
    }
    return translations[key] || key
  },
}))

describe('TextDiff', () => {
  it('shows no changes message when texts are identical', () => {
    render(<TextDiff oldText="Hello World" newText="Hello World" />)
    expect(screen.getByText('No changes detected')).toBeInTheDocument()
  })

  it('displays diff when texts are different', () => {
    render(<TextDiff oldText="Hello" newText="Hello World" />)
    expect(screen.queryByText('No changes detected')).not.toBeInTheDocument()
  })

  it('shows legend with labels', () => {
    render(<TextDiff oldText="Line 1" newText="Line 2" />)
    expect(screen.getByText(/Before.*Removed/)).toBeInTheDocument()
    expect(screen.getByText(/After.*Added/)).toBeInTheDocument()
  })

  it('uses custom labels when provided', () => {
    render(
      <TextDiff
        oldText="Old content"
        newText="New content"
        oldLabel="Original"
        newLabel="Modified"
      />
    )
    expect(screen.getByText(/Original.*Removed/)).toBeInTheDocument()
    expect(screen.getByText(/Modified.*Added/)).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(<TextDiff oldText="A" newText="B" className="custom-diff-class" />)
    expect(container.firstChild).toHaveClass('custom-diff-class')
  })

  it('handles empty strings', () => {
    render(<TextDiff oldText="" newText="New content" />)
    expect(screen.queryByText('No changes detected')).not.toBeInTheDocument()
  })

  it('shows removed lines with minus indicator', () => {
    render(<TextDiff oldText="Line to remove" newText="" />)
    expect(screen.getByText('-')).toBeInTheDocument()
  })

  it('shows added lines with plus indicator', () => {
    render(<TextDiff oldText="" newText="Line to add" />)
    expect(screen.getByText('+')).toBeInTheDocument()
  })
})

describe('TextDiffSideBySide', () => {
  it('shows no changes message when texts are identical', () => {
    render(<TextDiffSideBySide oldText="Same text" newText="Same text" />)
    expect(screen.getByText('No changes detected')).toBeInTheDocument()
  })

  it('displays side-by-side diff when texts differ', () => {
    render(<TextDiffSideBySide oldText="Old version" newText="New version" />)
    expect(screen.getByText('Version A')).toBeInTheDocument()
    expect(screen.getByText('Version B')).toBeInTheDocument()
  })

  it('uses custom labels', () => {
    render(
      <TextDiffSideBySide oldText="A" newText="B" oldLabel="Left Side" newLabel="Right Side" />
    )
    expect(screen.getByText('Left Side')).toBeInTheDocument()
    expect(screen.getByText('Right Side')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(
      <TextDiffSideBySide oldText="X" newText="Y" className="side-by-side-class" />
    )
    expect(container.firstChild).toHaveClass('side-by-side-class')
  })

  it('renders grid layout with two columns', () => {
    const { container } = render(<TextDiffSideBySide oldText="Left" newText="Right" />)
    const grid = container.querySelector('.grid-cols-2')
    expect(grid).toBeInTheDocument()
  })
})

describe('TextDiffInline', () => {
  it('shows no changes message when texts are identical', () => {
    render(<TextDiffInline oldText="Same text" newText="Same text" />)
    expect(screen.getByText('No changes detected')).toBeInTheDocument()
  })

  it('displays inline word-level diff', () => {
    render(<TextDiffInline oldText="Hello world" newText="Hello there" />)
    expect(screen.queryByText('No changes detected')).not.toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(
      <TextDiffInline oldText="A" newText="B" className="inline-diff-class" />
    )
    expect(container.firstChild).toHaveClass('inline-diff-class')
  })

  it('shows removed text with strikethrough styling', () => {
    const { container } = render(<TextDiffInline oldText="remove this" newText="" />)
    const removed = container.querySelector('.line-through')
    expect(removed).toBeInTheDocument()
  })

  it('shows added text with highlight', () => {
    const { container } = render(<TextDiffInline oldText="" newText="add this" />)
    const added = container.querySelector('.bg-green-200')
    expect(added).toBeInTheDocument()
  })
})
