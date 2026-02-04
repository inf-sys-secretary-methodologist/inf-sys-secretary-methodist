'use client'

import { useState } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { Menu, X, ChevronDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
  SheetClose,
} from '@/components/ui/sheet'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { NavEntry, NavItem, NavGroup, isNavGroup } from '@/config/navigation'
import { cn } from '@/lib/utils'

interface MobileNavProps {
  entries: NavEntry[]
}

export function MobileNav({ entries }: MobileNavProps) {
  const [open, setOpen] = useState(false)
  const [expandedGroups, setExpandedGroups] = useState<string[]>([])
  const pathname = usePathname()
  const [hoveredIndex, setHoveredIndex] = useState<string | null>(null)
  const t = useTranslations('nav')
  const tCommon = useTranslations('common')

  const toggleGroup = (groupKey: string) => {
    setExpandedGroups((prev) =>
      prev.includes(groupKey) ? prev.filter((k) => k !== groupKey) : [...prev, groupKey]
    )
  }

  // Check if any item in the group is active
  const isGroupActive = (group: NavGroup) => {
    return group.items.some((item) => pathname === item.url || pathname.startsWith(item.url + '/'))
  }

  // Render a single nav item
  const renderNavItem = (item: NavItem, indexKey: string, nested = false) => {
    const Icon = item.icon
    const isActive = pathname === item.url || pathname.startsWith(item.url + '/')
    const isHovered = hoveredIndex === indexKey

    return (
      <Link
        key={item.url}
        href={item.url}
        aria-current={isActive ? 'page' : undefined}
        /* c8 ignore next 3 - event handlers for mobile nav, tested in e2e */
        onClick={() => setOpen(false)}
        onMouseEnter={() => setHoveredIndex(indexKey)}
        onMouseLeave={() => setHoveredIndex(null)}
        className={cn(
          'relative focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded-lg',
          nested && 'ml-4'
        )}
      >
        <div
          className={cn(
            'relative flex items-center gap-3 rounded-lg px-4 py-3 text-sm font-medium transition-all duration-300',
            isActive
              ? 'text-white'
              : 'text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white'
          )}
        >
          {/* c8 ignore start - Background glow effect */}
          {(isActive || isHovered) && (
            <div
              className={cn(
                'absolute inset-0 rounded-lg transition-all duration-300',
                isActive
                  ? 'bg-gradient-to-r from-blue-500 to-purple-600 scale-100'
                  : 'bg-gradient-to-r from-gray-200 to-gray-300 dark:from-gray-700 dark:to-gray-600 scale-[0.98]'
              )}
              style={{
                boxShadow: isActive ? '0 0 20px rgba(59, 130, 246, 0.5)' : 'none',
              }}
            />
          )}
          {/* c8 ignore stop */}

          <div className="relative z-10 flex items-center gap-3">
            <Icon className="h-5 w-5" />
            <span>{t(item.nameKey)}</span>
          </div>
        </div>
      </Link>
    )
  }

  // Render a collapsible group
  const renderNavGroup = (group: NavGroup, index: number) => {
    const Icon = group.icon
    const isActive = isGroupActive(group)
    const isExpanded = expandedGroups.includes(group.nameKey) || isActive
    const groupKey = `group-${index}`
    const isHovered = hoveredIndex === groupKey

    return (
      <Collapsible
        key={group.nameKey}
        open={isExpanded}
        onOpenChange={() => toggleGroup(group.nameKey)}
      >
        <CollapsibleTrigger
          /* c8 ignore next 2 - hover handlers for mobile nav, tested in e2e */
          onMouseEnter={() => setHoveredIndex(groupKey)}
          onMouseLeave={() => setHoveredIndex(null)}
          className="w-full relative focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded-lg"
        >
          <div
            className={cn(
              'relative flex items-center justify-between rounded-lg px-4 py-3 text-sm font-medium transition-all duration-300',
              isActive
                ? 'text-white'
                : 'text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white'
            )}
          >
            {/* c8 ignore start - Background glow effect */}
            {(isActive || isHovered) && (
              <div
                className={cn(
                  'absolute inset-0 rounded-lg transition-all duration-300',
                  isActive
                    ? 'bg-gradient-to-r from-blue-500 to-purple-600 scale-100'
                    : 'bg-gradient-to-r from-gray-200 to-gray-300 dark:from-gray-700 dark:to-gray-600 scale-[0.98]'
                )}
                style={{
                  boxShadow: isActive ? '0 0 20px rgba(59, 130, 246, 0.5)' : 'none',
                }}
              />
            )}
            {/* c8 ignore stop */}

            <div className="relative z-10 flex items-center gap-3">
              <Icon className="h-5 w-5" />
              <span>{t(group.nameKey)}</span>
            </div>
            <ChevronDown
              className={cn(
                'relative z-10 h-4 w-4 transition-transform duration-200',
                isExpanded && 'rotate-180'
              )}
            />
          </div>
        </CollapsibleTrigger>
        <CollapsibleContent className="mt-1 space-y-1">
          {group.items.map((item, itemIndex) =>
            renderNavItem(item, `${groupKey}-item-${itemIndex}`, true)
          )}
        </CollapsibleContent>
      </Collapsible>
    )
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="icon" aria-label={t('openMenu')}>
          <Menu className="h-5 w-5" />
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="w-72 p-0 [&>button:first-child]:hidden">
        <SheetHeader className="flex flex-row items-center justify-between px-6 py-4">
          <SheetTitle className="text-left">{t('navigation')}</SheetTitle>
          <SheetClose className="rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2">
            <X className="h-5 w-5" />
            <span className="sr-only">{tCommon('close')}</span>
          </SheetClose>
        </SheetHeader>
        <nav aria-label={t('mobileNavigation')} className="flex flex-col gap-2 p-4">
          {entries.map((entry, index) =>
            isNavGroup(entry) ? renderNavGroup(entry, index) : renderNavItem(entry, `item-${index}`)
          )}
        </nav>
      </SheetContent>
    </Sheet>
  )
}
