'use client'

import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { LogOut, Settings, User as UserIcon, ChevronDown } from 'lucide-react'
import { toast } from 'sonner'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { useAuth, useLogout } from '@/hooks/useAuth'
import { cn } from '@/lib/utils'

interface UserMenuProps {
  className?: string
}

export function UserMenu({ className }: UserMenuProps) {
  const router = useRouter()
  const { user, isAuthenticated } = useAuth()
  const { logout, isLoading } = useLogout()

  if (!isAuthenticated || !user) {
    return null
  }

  const handleLogout = async () => {
    try {
      await logout('/login')
      toast.success('Выход выполнен успешно', {
        description: 'До скорой встречи!',
      })
    } catch (error) {
      toast.error('Ошибка выхода', {
        description: 'Попробуйте еще раз',
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
    if (!role) return 'Пользователь'
    const roleMap: Record<string, string> = {
      system_admin: 'Администратор',
      methodist: 'Методист',
      academic_secretary: 'Секретарь',
      teacher: 'Преподаватель',
      student: 'Студент',
    }
    return roleMap[role] || role
  }

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
          <AvatarFallback className="bg-primary text-primary-foreground font-medium">
            {getInitials(user?.name)}
          </AvatarFallback>
        </Avatar>

        <div className="hidden md:flex md:flex-col md:items-start text-left">
          <span className="text-sm font-medium text-foreground">{user?.name || 'Пользователь'}</span>
          <span className="text-xs text-muted-foreground">
            {getRoleDisplayName(user?.role)}
          </span>
        </div>

        <ChevronDown className="h-4 w-4 text-muted-foreground hidden md:block" />
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel>
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-medium">{user?.name || 'Пользователь'}</p>
            <p className="text-xs text-muted-foreground">{user?.email || ''}</p>
            <p className="text-xs text-muted-foreground">
              {getRoleDisplayName(user?.role)}
            </p>
          </div>
        </DropdownMenuLabel>

        <DropdownMenuSeparator />

        <DropdownMenuItem asChild>
          <Link href="/profile" className="cursor-pointer">
            <UserIcon className="mr-2 h-4 w-4" />
            <span>Профиль</span>
          </Link>
        </DropdownMenuItem>

        <DropdownMenuItem asChild>
          <Link href="/settings" className="cursor-pointer">
            <Settings className="mr-2 h-4 w-4" />
            <span>Настройки</span>
          </Link>
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem
          onClick={handleLogout}
          disabled={isLoading}
          className="cursor-pointer text-red-600 focus:text-red-600 focus:bg-red-50 dark:focus:bg-red-950/20"
        >
          <LogOut className="mr-2 h-4 w-4" />
          <span>{isLoading ? 'Выход...' : 'Выйти'}</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
