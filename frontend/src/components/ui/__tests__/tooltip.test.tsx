import { render, screen } from '@testing-library/react'
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from '../tooltip'

describe('Tooltip Components', () => {
  describe('TooltipProvider', () => {
    it('renders children', () => {
      render(
        <TooltipProvider>
          <div data-testid="child">Content</div>
        </TooltipProvider>
      )
      expect(screen.getByTestId('child')).toBeInTheDocument()
    })
  })

  describe('Tooltip', () => {
    it('renders children', () => {
      render(
        <Tooltip>
          <div data-testid="child">Content</div>
        </Tooltip>
      )
      expect(screen.getByTestId('child')).toBeInTheDocument()
    })
  })

  describe('TooltipTrigger', () => {
    it('renders children in span', () => {
      render(
        <TooltipTrigger>
          <button>Trigger</button>
        </TooltipTrigger>
      )
      expect(screen.getByRole('button', { name: 'Trigger' })).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(
        <TooltipTrigger className="custom-class" data-testid="trigger">
          Trigger
        </TooltipTrigger>
      )
      expect(screen.getByTestId('trigger')).toHaveClass('custom-class')
    })

    it('uses asChild to clone element', () => {
      render(
        <TooltipTrigger asChild>
          <button data-testid="button">Button</button>
        </TooltipTrigger>
      )
      expect(screen.getByTestId('button')).toBeInTheDocument()
    })

    it('applies group class for hover behavior', () => {
      render(<TooltipTrigger data-testid="trigger">Trigger</TooltipTrigger>)
      expect(screen.getByTestId('trigger')).toHaveClass('group/tooltip')
    })
  })

  describe('TooltipContent', () => {
    it('renders with tooltip role', () => {
      render(<TooltipContent>Tooltip text</TooltipContent>)
      expect(screen.getByRole('tooltip')).toBeInTheDocument()
      expect(screen.getByText('Tooltip text')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(<TooltipContent className="custom-class">Content</TooltipContent>)
      expect(screen.getByRole('tooltip')).toHaveClass('custom-class')
    })

    it('positions on top by default', () => {
      render(<TooltipContent>Content</TooltipContent>)
      expect(screen.getByRole('tooltip')).toHaveClass('bottom-full')
    })

    it('positions on bottom', () => {
      render(<TooltipContent side="bottom">Content</TooltipContent>)
      expect(screen.getByRole('tooltip')).toHaveClass('top-full')
    })

    it('positions on left', () => {
      render(<TooltipContent side="left">Content</TooltipContent>)
      expect(screen.getByRole('tooltip')).toHaveClass('right-full')
    })

    it('positions on right', () => {
      render(<TooltipContent side="right">Content</TooltipContent>)
      expect(screen.getByRole('tooltip')).toHaveClass('left-full')
    })
  })

  describe('Full Tooltip composition', () => {
    it('renders complete tooltip structure', () => {
      render(
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger data-testid="trigger">Hover me</TooltipTrigger>
            <TooltipContent>Tooltip content</TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )

      expect(screen.getByTestId('trigger')).toBeInTheDocument()
      expect(screen.getByText('Hover me')).toBeInTheDocument()
      expect(screen.getByRole('tooltip')).toHaveTextContent('Tooltip content')
    })
  })
})
