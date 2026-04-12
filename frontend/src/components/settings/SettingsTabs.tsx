'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { Palette, Bell, Workflow } from 'lucide-react'
import { cn } from '@/lib/utils'

const tabs = [
  { href: '/settings/appearance', labelKey: 'appearance.title' as const, icon: Palette },
  { href: '/settings/notifications', labelKey: 'notifications.title' as const, icon: Bell },
  { href: '/settings/automation', labelKey: 'automation.title' as const, icon: Workflow },
]

export function SettingsTabs() {
  const pathname = usePathname()
  const t = useTranslations('settings')

  return (
    <div className="mb-6">
      <nav className="flex gap-1 overflow-x-auto">
        {tabs.map(({ href, labelKey, icon: Icon }) => {
          const isActive = pathname === href
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                'flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-all whitespace-nowrap border',
                isActive
                  ? 'border-primary bg-primary text-primary-foreground shadow-sm'
                  : 'border-transparent text-muted-foreground hover:border-border hover:text-foreground'
              )}
            >
              <Icon className="h-4 w-4" />
              {t(labelKey)}
            </Link>
          )
        })}
      </nav>
    </div>
  )
}
