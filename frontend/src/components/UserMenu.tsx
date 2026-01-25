'use client'

import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { LogOut, User as UserIcon, ChevronDown, Bell, Palette } from 'lucide-react'
import { toast } from 'sonner'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { useAuth, useLogout } from '@/hooks/useAuth'
import { cn } from '@/lib/utils'

interface UserMenuProps {
  className?: string
}

export function UserMenu({ className }: UserMenuProps) {
  const { user, isAuthenticated } = useAuth()
  const { logout, isLoading } = useLogout()
  const t = useTranslations('userMenu')
  const tRoles = useTranslations('roles')
  const tAuth = useTranslations('auth')

  if (!isAuthenticated || !user) {
    return null
  }

  /* c8 ignore start - Logout handler and role display helpers */
  const handleLogout = async () => {
    try {
      await logout('/login')
      toast.success(t('logoutSuccess'), {
        description: t('logoutSuccessDesc'),
      })
    } catch (_error) {
      toast.error(t('logoutError'), {
        description: t('logoutErrorDesc'),
      })
    }
  }

  // Get user initials for avatar
  const getInitials = (name?: string): string => {
    if (!name) return 'U'
    const trimmedName = name.trim()
    if (!trimmedName) return 'U'

    const parts = trimmedName.split(' ')
    if (parts.length >= 2) {
      return `${parts[0][0]}${parts[1][0]}`.toUpperCase()
    }
    return trimmedName.slice(0, 2).toUpperCase()
  }

  // Get role display name
  const getRoleDisplayName = (role?: string): string => {
    if (!role) return t('defaultUser')
    try {
      return tRoles(
        role as 'system_admin' | 'methodist' | 'academic_secretary' | 'teacher' | 'student'
      )
    } catch {
      return role
    }
  }
  /* c8 ignore stop */

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        className={cn(
          'flex items-center gap-3 px-3 py-2 rounded-lg',
          'hover:bg-accent transition-colors',
          'focus:outline-none focus:ring-2 focus:ring-ring',
          className
        )}
      >
        <Avatar className="h-9 w-9">
          {(user as unknown as { avatar?: string }).avatar && (
            <AvatarImage
              src={(user as unknown as { avatar?: string }).avatar}
              alt={user?.name || t('avatar')}
            />
          )}
          <AvatarFallback className="bg-primary text-primary-foreground font-medium">
            {getInitials(user?.name)}
          </AvatarFallback>
        </Avatar>

        <div className="hidden xl:flex xl:flex-col xl:items-start text-left">
          <span className="text-sm font-medium text-foreground">
            {user?.name || t('defaultUser')}
          </span>
          <span className="text-xs text-muted-foreground">{getRoleDisplayName(user?.role)}</span>
        </div>

        <ChevronDown className="h-4 w-4 text-muted-foreground hidden xl:block" />
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel>
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-medium">{user?.name || t('defaultUser')}</p>
            <p className="text-xs text-muted-foreground">{user?.email || ''}</p>
            <p className="text-xs text-muted-foreground">{getRoleDisplayName(user?.role)}</p>
          </div>
        </DropdownMenuLabel>

        <DropdownMenuSeparator />

        <DropdownMenuItem asChild>
          <Link href="/profile" className="cursor-pointer">
            <UserIcon className="mr-2 h-4 w-4" />
            <span>{t('profile')}</span>
          </Link>
        </DropdownMenuItem>

        <DropdownMenuItem asChild>
          <Link href="/settings/appearance" className="cursor-pointer">
            <Palette className="mr-2 h-4 w-4" />
            <span>{t('appearance')}</span>
          </Link>
        </DropdownMenuItem>

        <DropdownMenuItem asChild>
          <Link href="/settings/notifications" className="cursor-pointer">
            <Bell className="mr-2 h-4 w-4" />
            <span>{t('notifications')}</span>
          </Link>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem
          onClick={handleLogout}
          disabled={isLoading}
          className="cursor-pointer text-red-600 focus:text-red-600 focus:bg-red-50 dark:focus:bg-red-950/20"
        >
          <LogOut className="mr-2 h-4 w-4" />
          <span>{isLoading ? t('loggingOut') : tAuth('logout')}</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
