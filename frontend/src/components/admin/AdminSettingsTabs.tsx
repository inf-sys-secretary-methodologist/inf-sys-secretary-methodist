'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { Workflow, ShieldCheck } from 'lucide-react'
import { cn } from '@/lib/utils'

// System-admin settings tabs.
// Global brand/colors live in /admin/branding (real backend persistence).
// Global notification channels (SMTP/VAPID/Telegram bot) — backlog, not
// in the system settings UI yet. Личные настройки (своя тема, свои
// каналы уведомлений) — /settings/* per roles-and-flows.md (PermissionMatrix).
const tabs = [
  { href: '/admin/settings/automation', labelKey: 'automation.title' as const, icon: Workflow },
  { href: '/admin/settings/security', labelKey: 'security.title' as const, icon: ShieldCheck },
]

export function AdminSettingsTabs() {
  const pathname = usePathname()
  const t = useTranslations('adminSettings')

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
