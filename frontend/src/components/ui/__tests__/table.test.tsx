import { render, screen } from '@testing-library/react'
import {
  Table,
  TableHeader,
  TableBody,
  TableFooter,
  TableHead,
  TableRow,
  TableCell,
  TableCaption,
} from '../table'

describe('Table Components', () => {
  describe('Table', () => {
    it('renders table element', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByRole('table')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(
        <Table className="custom-class">
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByRole('table')).toHaveClass('custom-class')
    })

    it('applies default classes', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByRole('table')).toHaveClass('w-full', 'caption-bottom', 'text-sm')
    })
  })

  describe('TableHeader', () => {
    it('renders thead element', () => {
      render(
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Header</TableHead>
            </TableRow>
          </TableHeader>
        </Table>
      )
      expect(screen.getByRole('rowgroup')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(
        <Table>
          <TableHeader className="custom-class" data-testid="header">
            <TableRow>
              <TableHead>Header</TableHead>
            </TableRow>
          </TableHeader>
        </Table>
      )
      expect(screen.getByTestId('header')).toHaveClass('custom-class')
    })
  })

  describe('TableBody', () => {
    it('renders tbody element', () => {
      render(
        <Table>
          <TableBody data-testid="body">
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByTestId('body')).toBeInTheDocument()
    })
  })

  describe('TableRow', () => {
    it('renders tr element', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByRole('row')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(
        <Table>
          <TableBody>
            <TableRow className="custom-class">
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByRole('row')).toHaveClass('custom-class')
    })
  })

  describe('TableHead', () => {
    it('renders th element', () => {
      render(
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Header</TableHead>
            </TableRow>
          </TableHeader>
        </Table>
      )
      expect(screen.getByRole('columnheader')).toBeInTheDocument()
      expect(screen.getByText('Header')).toBeInTheDocument()
    })
  })

  describe('TableCell', () => {
    it('renders td element', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell content</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByRole('cell')).toBeInTheDocument()
      expect(screen.getByText('Cell content')).toBeInTheDocument()
    })
  })

  describe('TableFooter', () => {
    it('renders tfoot element', () => {
      render(
        <Table>
          <TableFooter data-testid="footer">
            <TableRow>
              <TableCell>Footer</TableCell>
            </TableRow>
          </TableFooter>
        </Table>
      )
      expect(screen.getByTestId('footer')).toBeInTheDocument()
    })
  })

  describe('TableCaption', () => {
    it('renders caption element', () => {
      render(
        <Table>
          <TableCaption>Table caption</TableCaption>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByText('Table caption')).toBeInTheDocument()
    })
  })

  describe('Full Table composition', () => {
    it('renders complete table structure', () => {
      render(
        <Table>
          <TableCaption>Users list</TableCaption>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>John</TableCell>
              <TableCell>john@example.com</TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Jane</TableCell>
              <TableCell>jane@example.com</TableCell>
            </TableRow>
          </TableBody>
          <TableFooter>
            <TableRow>
              <TableCell colSpan={2}>Total: 2 users</TableCell>
            </TableRow>
          </TableFooter>
        </Table>
      )

      expect(screen.getByText('Users list')).toBeInTheDocument()
      expect(screen.getByText('Name')).toBeInTheDocument()
      expect(screen.getByText('Email')).toBeInTheDocument()
      expect(screen.getByText('John')).toBeInTheDocument()
      expect(screen.getByText('john@example.com')).toBeInTheDocument()
      expect(screen.getByText('Total: 2 users')).toBeInTheDocument()
    })
  })
})
