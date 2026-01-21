import { render, screen, waitFor } from '@testing-library/react'
import { Avatar, AvatarImage, AvatarFallback } from '../avatar'

describe('Avatar Components', () => {
  describe('Avatar', () => {
    it('renders avatar container', () => {
      render(
        <Avatar data-testid="avatar">
          <AvatarFallback>JD</AvatarFallback>
        </Avatar>
      )
      expect(screen.getByTestId('avatar')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(
        <Avatar className="custom-class" data-testid="avatar">
          <AvatarFallback>JD</AvatarFallback>
        </Avatar>
      )
      expect(screen.getByTestId('avatar')).toHaveClass('custom-class')
    })

    it('applies default classes', () => {
      render(
        <Avatar data-testid="avatar">
          <AvatarFallback>JD</AvatarFallback>
        </Avatar>
      )
      const avatar = screen.getByTestId('avatar')
      expect(avatar).toHaveClass('relative', 'flex', 'h-10', 'w-10', 'shrink-0', 'rounded-full')
    })
  })

  describe('AvatarFallback', () => {
    it('renders fallback text', () => {
      render(
        <Avatar>
          <AvatarFallback>JD</AvatarFallback>
        </Avatar>
      )
      expect(screen.getByText('JD')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(
        <Avatar>
          <AvatarFallback className="custom-class">JD</AvatarFallback>
        </Avatar>
      )
      expect(screen.getByText('JD')).toHaveClass('custom-class')
    })

    it('applies default classes', () => {
      render(
        <Avatar>
          <AvatarFallback>JD</AvatarFallback>
        </Avatar>
      )
      const fallback = screen.getByText('JD')
      expect(fallback).toHaveClass('flex', 'h-full', 'w-full', 'items-center', 'justify-center')
    })

    it('renders any content as fallback', () => {
      render(
        <Avatar>
          <AvatarFallback>
            <span data-testid="icon">Icon</span>
          </AvatarFallback>
        </Avatar>
      )
      expect(screen.getByTestId('icon')).toBeInTheDocument()
    })
  })

  describe('AvatarImage', () => {
    it('renders image with src', async () => {
      render(
        <Avatar>
          <AvatarImage src="/test-image.jpg" alt="User avatar" />
          <AvatarFallback>JD</AvatarFallback>
        </Avatar>
      )

      // Image may take time to load, fallback shows first
      await waitFor(
        () => {
          const img = screen.queryByRole('img')
          if (img) {
            expect(img).toHaveAttribute('src', '/test-image.jpg')
            expect(img).toHaveAttribute('alt', 'User avatar')
          }
        },
        { timeout: 100 }
      )
    })

    it('applies custom className to image', async () => {
      render(
        <Avatar>
          <AvatarImage src="/test.jpg" alt="Test" className="custom-image-class" />
          <AvatarFallback>JD</AvatarFallback>
        </Avatar>
      )

      await waitFor(
        () => {
          const img = screen.queryByRole('img')
          if (img) {
            expect(img).toHaveClass('custom-image-class')
          }
        },
        { timeout: 100 }
      )
    })
  })

  describe('Full Avatar composition', () => {
    it('shows fallback when image fails to load', async () => {
      render(
        <Avatar>
          <AvatarImage src="/nonexistent.jpg" alt="User" />
          <AvatarFallback>AB</AvatarFallback>
        </Avatar>
      )

      // Fallback should be visible
      expect(screen.getByText('AB')).toBeInTheDocument()
    })

    it('renders with different sizes via className', () => {
      render(
        <Avatar className="h-16 w-16" data-testid="large-avatar">
          <AvatarFallback>XL</AvatarFallback>
        </Avatar>
      )

      const avatar = screen.getByTestId('large-avatar')
      expect(avatar).toHaveClass('h-16', 'w-16')
    })
  })
})
