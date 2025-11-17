'use client'

import { useState } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { LucideIcon } from 'lucide-react'

export interface NavItem {
  name: string
  url: string
  icon: LucideIcon
}

interface NavBarProps {
  items: NavItem[]
  className?: string
}

export function NavBar({ items, className = '' }: NavBarProps) {
  const pathname = usePathname()
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null)

  return (
    <nav className={`fixed top-8 left-1/2 -translate-x-1/2 z-40 ${className}`}>
      <div className="relative flex items-center gap-2 px-4 py-2 rounded-full
                    bg-white/80 dark:bg-gray-900/80 backdrop-blur-lg
                    border border-gray-200 dark:border-gray-700
                    shadow-lg">
        {items.map((item, index) => {
          const Icon = item.icon
          const isActive = pathname === item.url
          const isHovered = hoveredIndex === index

          return (
            <Link
              key={item.url}
              href={item.url}
              onMouseEnter={() => setHoveredIndex(index)}
              onMouseLeave={() => setHoveredIndex(null)}
              className="relative"
            >
              <div
                className={`
                  relative flex items-center gap-2 px-4 py-2 rounded-full
                  transition-all duration-300
                  ${
                    isActive
                      ? 'text-white'
                      : 'text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white'
                  }
                `}
              >
                {/* Background glow effect */}
                {(isActive || isHovered) && (
                  <div
                    className={`
                      absolute inset-0 rounded-full transition-all duration-300
                      ${
                        isActive
                          ? 'bg-gradient-to-r from-blue-500 to-purple-600 scale-100'
                          : 'bg-gradient-to-r from-gray-200 to-gray-300 dark:from-gray-700 dark:to-gray-600 scale-95'
                      }
                    `}
                    style={{
                      boxShadow: isActive
                        ? '0 0 20px rgba(59, 130, 246, 0.5)'
                        : 'none'
                    }}
                  />
                )}

                {/* Content */}
                <div className="relative z-10 flex items-center gap-2">
                  <Icon className="h-4 w-4" />
                  <span className="text-sm font-medium hidden sm:inline">
                    {item.name}
                  </span>
                </div>
              </div>
            </Link>
          )
        })}
      </div>
    </nav>
  )
}
