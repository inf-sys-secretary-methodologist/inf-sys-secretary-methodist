'use client'

import { useState } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Menu, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
  SheetClose,
} from '@/components/ui/sheet'
import { NavItem } from '@/config/navigation'
import { cn } from '@/lib/utils'

interface MobileNavProps {
  items: NavItem[]
}

export function MobileNav({ items }: MobileNavProps) {
  const [open, setOpen] = useState(false)
  const pathname = usePathname()
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null)

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="icon" className="md:hidden" aria-label="Открыть меню">
          <Menu className="h-5 w-5" />
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="w-72 p-0 [&>button:first-child]:hidden">
        <SheetHeader className="flex flex-row items-center justify-between px-6 py-4">
          <SheetTitle className="text-left">Навигация</SheetTitle>
          <SheetClose className="rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2">
            <X className="h-5 w-5" />
            <span className="sr-only">Закрыть</span>
          </SheetClose>
        </SheetHeader>
        <nav className="flex flex-col gap-2 p-4">
          {items.map((item, index) => {
            const Icon = item.icon
            const isActive = pathname === item.url
            const isHovered = hoveredIndex === index

            return (
              <Link
                key={item.url}
                href={item.url}
                onClick={() => setOpen(false)}
                onMouseEnter={() => setHoveredIndex(index)}
                onMouseLeave={() => setHoveredIndex(null)}
                className="relative"
              >
                <div
                  className={cn(
                    'relative flex items-center gap-3 rounded-lg px-4 py-3 text-sm font-medium transition-all duration-300',
                    isActive
                      ? 'text-white'
                      : 'text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white'
                  )}
                >
                  {/* Background glow effect */}
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

                  {/* Content */}
                  <div className="relative z-10 flex items-center gap-3">
                    <Icon className="h-5 w-5" />
                    <span>{item.name}</span>
                  </div>
                </div>
              </Link>
            )
          })}
        </nav>
      </SheetContent>
    </Sheet>
  )
}
