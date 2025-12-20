'use client'

import { useState } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { UserMenu } from '@/components/UserMenu'
import { ThemeSettingsPopover } from '@/components/theme-settings-popover'
import { LanguageSwitcher } from '@/components/language-switcher'
import { NotificationBell } from '@/components/notifications/NotificationBell'
import { MobileNav } from './MobileNav'
import { NavItem } from '@/config/navigation'
import { cn } from '@/lib/utils'

interface AppHeaderProps {
  items: NavItem[]
}

export function AppHeader({ items }: AppHeaderProps) {
  const pathname = usePathname()
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null)
  const t = useTranslations('nav')

  return (
    <header className="sticky top-0 z-50 w-full pt-4 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      {/* Mobile header - below lg */}
      <div className="lg:hidden flex h-14 items-center justify-between px-4 sm:px-6">
        <MobileNav items={items} />
        <div className="flex items-center gap-2">
          <NotificationBell />
          <LanguageSwitcher />
          <ThemeSettingsPopover />
          <UserMenu />
        </div>
      </div>

      {/* Desktop header - lg and above (1024px+) */}
      <div className="hidden lg:flex h-14 items-center justify-between px-6 xl:px-8">
        {/* Left spacer for centering */}
        <div className="w-44" />

        {/* Desktop Navigation - centered */}
        <nav aria-label={t('mainNavigation')}>
          <div
            className="flex items-center gap-1 rounded-full bg-white/80 dark:bg-gray-900/80 backdrop-blur-lg border border-gray-200 dark:border-gray-700 px-3 py-2 shadow-lg"
            role="list"
          >
            {items.map((item, index) => {
              const Icon = item.icon
              const isActive = pathname === item.url
              const isHovered = hoveredIndex === index

              return (
                <Link
                  key={item.url}
                  href={item.url}
                  aria-current={isActive ? 'page' : undefined}
                  onMouseEnter={() => setHoveredIndex(index)}
                  onMouseLeave={() => setHoveredIndex(null)}
                  className="relative focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded-full"
                  role="listitem"
                >
                  <div
                    className={cn(
                      'relative flex items-center gap-1.5 px-3 py-1.5 rounded-full transition-all duration-300',
                      isActive
                        ? 'text-white'
                        : 'text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white'
                    )}
                  >
                    {/* Background glow effect */}
                    {(isActive || isHovered) && (
                      <div
                        className={cn(
                          'absolute inset-0 rounded-full transition-all duration-300',
                          isActive
                            ? 'bg-gradient-to-r from-blue-500 to-purple-600 scale-100'
                            : 'bg-gradient-to-r from-gray-200 to-gray-300 dark:from-gray-700 dark:to-gray-600 scale-95'
                        )}
                        style={{
                          boxShadow: isActive ? '0 0 20px rgba(59, 130, 246, 0.5)' : 'none',
                        }}
                      />
                    )}

                    {/* Content */}
                    <div className="relative z-10 flex items-center gap-1.5">
                      <Icon className="h-4 w-4" />
                      <span className="text-sm font-medium whitespace-nowrap">
                        {t(item.nameKey)}
                      </span>
                    </div>
                  </div>
                </Link>
              )
            })}
          </div>
        </nav>

        {/* Right side actions */}
        <div className="flex items-center justify-end gap-2 w-44">
          <NotificationBell />
          <LanguageSwitcher />
          <ThemeSettingsPopover />
          <UserMenu />
        </div>
      </div>
    </header>
  )
}
