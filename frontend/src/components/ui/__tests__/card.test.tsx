import { render, screen } from '@testing-library/react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '../card'

describe('Card Components', () => {
  describe('Card', () => {
    it('renders children', () => {
      render(<Card>Card content</Card>)
      expect(screen.getByText('Card content')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(<Card className="custom-class">Content</Card>)
      expect(screen.getByText('Content')).toHaveClass('custom-class')
    })

    it('applies default classes', () => {
      render(<Card>Content</Card>)
      const card = screen.getByText('Content')
      expect(card).toHaveClass('rounded-xl', 'border', 'bg-card', 'shadow')
    })

    it('forwards ref', () => {
      const ref = { current: null }
      render(<Card ref={ref}>Content</Card>)
      expect(ref.current).toBeInstanceOf(HTMLDivElement)
    })
  })

  describe('CardHeader', () => {
    it('renders children', () => {
      render(<CardHeader>Header content</CardHeader>)
      expect(screen.getByText('Header content')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(<CardHeader className="custom-class">Header</CardHeader>)
      expect(screen.getByText('Header')).toHaveClass('custom-class')
    })

    it('applies default classes', () => {
      render(<CardHeader>Header</CardHeader>)
      expect(screen.getByText('Header')).toHaveClass('flex', 'flex-col', 'p-6')
    })
  })

  describe('CardTitle', () => {
    it('renders children', () => {
      render(<CardTitle>Title text</CardTitle>)
      expect(screen.getByText('Title text')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(<CardTitle className="custom-class">Title</CardTitle>)
      expect(screen.getByText('Title')).toHaveClass('custom-class')
    })

    it('applies default classes', () => {
      render(<CardTitle>Title</CardTitle>)
      expect(screen.getByText('Title')).toHaveClass('font-semibold', 'leading-none')
    })
  })

  describe('CardDescription', () => {
    it('renders children', () => {
      render(<CardDescription>Description text</CardDescription>)
      expect(screen.getByText('Description text')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(<CardDescription className="custom-class">Description</CardDescription>)
      expect(screen.getByText('Description')).toHaveClass('custom-class')
    })

    it('applies default classes', () => {
      render(<CardDescription>Description</CardDescription>)
      expect(screen.getByText('Description')).toHaveClass('text-sm', 'text-muted-foreground')
    })
  })

  describe('CardContent', () => {
    it('renders children', () => {
      render(<CardContent>Content text</CardContent>)
      expect(screen.getByText('Content text')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(<CardContent className="custom-class">Content</CardContent>)
      expect(screen.getByText('Content')).toHaveClass('custom-class')
    })

    it('applies default classes', () => {
      render(<CardContent>Content</CardContent>)
      expect(screen.getByText('Content')).toHaveClass('p-6', 'pt-0')
    })
  })

  describe('CardFooter', () => {
    it('renders children', () => {
      render(<CardFooter>Footer content</CardFooter>)
      expect(screen.getByText('Footer content')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(<CardFooter className="custom-class">Footer</CardFooter>)
      expect(screen.getByText('Footer')).toHaveClass('custom-class')
    })

    it('applies default classes', () => {
      render(<CardFooter>Footer</CardFooter>)
      expect(screen.getByText('Footer')).toHaveClass('flex', 'items-center', 'p-6')
    })
  })

  describe('Full Card composition', () => {
    it('renders complete card structure', () => {
      render(
        <Card>
          <CardHeader>
            <CardTitle>Card Title</CardTitle>
            <CardDescription>Card description text</CardDescription>
          </CardHeader>
          <CardContent>Main content</CardContent>
          <CardFooter>Footer actions</CardFooter>
        </Card>
      )

      expect(screen.getByText('Card Title')).toBeInTheDocument()
      expect(screen.getByText('Card description text')).toBeInTheDocument()
      expect(screen.getByText('Main content')).toBeInTheDocument()
      expect(screen.getByText('Footer actions')).toBeInTheDocument()
    })
  })
})
