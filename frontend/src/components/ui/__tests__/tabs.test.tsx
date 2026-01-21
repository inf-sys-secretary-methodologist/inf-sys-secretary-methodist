import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../tabs'

describe('Tabs Components', () => {
  const renderTabs = () => {
    return render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
        <TabsContent value="tab2">Content 2</TabsContent>
      </Tabs>
    )
  }

  describe('Tabs', () => {
    it('renders tabs component', () => {
      renderTabs()
      expect(screen.getByRole('tablist')).toBeInTheDocument()
    })

    it('shows default tab content', () => {
      renderTabs()
      expect(screen.getByText('Content 1')).toBeInTheDocument()
    })
  })

  describe('TabsList', () => {
    it('renders tablist role', () => {
      renderTabs()
      expect(screen.getByRole('tablist')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(
        <Tabs defaultValue="tab1">
          <TabsList className="custom-class">
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">Content</TabsContent>
        </Tabs>
      )
      expect(screen.getByRole('tablist')).toHaveClass('custom-class')
    })

    it('applies default classes', () => {
      renderTabs()
      const tabsList = screen.getByRole('tablist')
      expect(tabsList).toHaveClass('inline-flex', 'h-10', 'items-center', 'rounded-md', 'bg-muted')
    })
  })

  describe('TabsTrigger', () => {
    it('renders tab buttons', () => {
      renderTabs()
      expect(screen.getByRole('tab', { name: 'Tab 1' })).toBeInTheDocument()
      expect(screen.getByRole('tab', { name: 'Tab 2' })).toBeInTheDocument()
    })

    it('switches tabs on click', async () => {
      const user = userEvent.setup()
      renderTabs()

      // Initial state - Tab 1 content visible
      expect(screen.getByText('Content 1')).toBeInTheDocument()

      // Click Tab 2
      await user.click(screen.getByRole('tab', { name: 'Tab 2' }))

      // Tab 2 content should be visible
      expect(screen.getByText('Content 2')).toBeInTheDocument()
    })

    it('applies active state styling', async () => {
      renderTabs()
      const tab1 = screen.getByRole('tab', { name: 'Tab 1' })

      expect(tab1).toHaveAttribute('data-state', 'active')
    })

    it('can be disabled', () => {
      render(
        <Tabs defaultValue="tab1">
          <TabsList>
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
            <TabsTrigger value="tab2" disabled>
              Tab 2
            </TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">Content 1</TabsContent>
          <TabsContent value="tab2">Content 2</TabsContent>
        </Tabs>
      )

      const disabledTab = screen.getByRole('tab', { name: 'Tab 2' })
      expect(disabledTab).toBeDisabled()
    })
  })

  describe('TabsContent', () => {
    it('shows content for active tab', () => {
      renderTabs()
      expect(screen.getByText('Content 1')).toBeInTheDocument()
    })

    it('applies custom className', () => {
      render(
        <Tabs defaultValue="tab1">
          <TabsList>
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1" className="custom-class">
            Content
          </TabsContent>
        </Tabs>
      )
      expect(screen.getByRole('tabpanel')).toHaveClass('custom-class')
    })
  })

  describe('Keyboard navigation', () => {
    it('supports keyboard navigation', async () => {
      const user = userEvent.setup()
      renderTabs()

      const tab1 = screen.getByRole('tab', { name: 'Tab 1' })
      tab1.focus()

      // Press ArrowRight to move to next tab
      await user.keyboard('{ArrowRight}')

      const tab2 = screen.getByRole('tab', { name: 'Tab 2' })
      expect(tab2).toHaveFocus()
    })
  })
})
