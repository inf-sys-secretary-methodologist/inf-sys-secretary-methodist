'use client'

import { usePathname } from 'next/navigation'
import { useTranslations } from 'next-intl'
import Link from 'next/link'
import { Palette, Bell, Workflow } from 'lucide-react'
import { cn } from '@/lib/utils'

const settingsNav = [
  { href: '/settings/appearance', labelKey: 'appearance.title' as const, icon: Palette },
  { href: '/settings/notifications', labelKey: 'notifications.title' as const, icon: Bell },
  { href: '/settings/automation', labelKey: 'automation.title' as const, icon: Workflow },
]

export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const t = useTranslations('settings')

  return (
    <div>
      <div className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container mx-auto px-4 sm:px-6">
          <nav className="flex gap-1 overflow-x-auto py-2" aria-label={t('title')}>
            {settingsNav.map(({ href, labelKey, icon: Icon }) => {
              const isActive = pathname === href
              return (
                <Link
                  key={href}
                  href={href}
                  className={cn(
                    'flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition-colors whitespace-nowrap',
                    isActive
                      ? 'bg-primary text-primary-foreground'
                      : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                  )}
                  aria-current={isActive ? 'page' : undefined}
                >
                  <Icon className="h-4 w-4" />
                  {t(labelKey)}
                </Link>
              )
            })}
          </nav>
        </div>
      </div>
      {children}
    </div>
  )
}
