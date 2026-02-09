'use client'

import { useTranslations } from 'next-intl'
import { cn } from '@/lib/utils'
import {
  FileText,
  ScrollText,
  Mail,
  ClipboardList,
  FileSignature,
  Briefcase,
  LayoutGrid,
  Star,
  Clock,
} from 'lucide-react'

export type TemplateCategory =
  | 'all'
  | 'favorites'
  | 'recent'
  | 'orders'
  | 'memos'
  | 'letters'
  | 'protocols'
  | 'contracts'
  | 'other'

interface CategoryConfig {
  key: TemplateCategory
  icon: React.ComponentType<{ className?: string }>
  // Codes that belong to this category
  codes?: string[]
}

export const CATEGORY_CONFIG: CategoryConfig[] = [
  { key: 'all', icon: LayoutGrid },
  { key: 'favorites', icon: Star },
  { key: 'recent', icon: Clock },
  { key: 'orders', icon: ScrollText, codes: ['order_main', 'order_hr', 'order_admin'] },
  { key: 'memos', icon: FileText, codes: ['memo', 'directive'] },
  { key: 'letters', icon: Mail, codes: ['business_letter'] },
  { key: 'protocols', icon: ClipboardList, codes: ['protocol'] },
  { key: 'contracts', icon: FileSignature, codes: ['contract'] },
  { key: 'other', icon: Briefcase, codes: ['job_instruction'] },
]

interface TemplateCategoryTabsProps {
  activeCategory: TemplateCategory
  onCategoryChange: (category: TemplateCategory) => void
  counts?: Record<TemplateCategory, number>
  className?: string
}

export function TemplateCategoryTabs({
  activeCategory,
  onCategoryChange,
  counts,
  className,
}: TemplateCategoryTabsProps) {
  const t = useTranslations('templates.categories')

  return (
    <div className={cn('flex flex-wrap gap-2', className)}>
      {CATEGORY_CONFIG.map(({ key, icon: Icon }) => {
        const isActive = activeCategory === key
        const count = counts?.[key]
        const showCount = count !== undefined && count > 0

        return (
          <button
            key={key}
            onClick={() => onCategoryChange(key)}
            className={cn(
              'inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200',
              'border focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
              isActive
                ? 'bg-primary text-primary-foreground border-primary shadow-sm'
                : 'bg-card text-muted-foreground border-border hover:bg-accent hover:text-accent-foreground'
            )}
          >
            <Icon className="h-4 w-4" />
            <span>{t(key)}</span>
            {showCount && (
              <span
                className={cn(
                  'inline-flex items-center justify-center min-w-[20px] h-5 px-1.5 rounded-full text-xs font-medium',
                  isActive
                    ? 'bg-primary-foreground/20 text-primary-foreground'
                    : 'bg-muted text-muted-foreground'
                )}
              >
                {count}
              </span>
            )}
          </button>
        )
      })}
    </div>
  )
}

// Helper function to get category for a template code
export function getCategoryForCode(code: string): TemplateCategory {
  for (const config of CATEGORY_CONFIG) {
    if (config.codes?.includes(code)) {
      return config.key
    }
  }
  return 'other'
}

// Helper function to filter templates by category
export function filterTemplatesByCategory<T extends { code: string; id: number }>(
  templates: T[],
  category: TemplateCategory,
  favorites: number[],
  recent: number[]
): T[] {
  switch (category) {
    case 'all':
      return templates
    case 'favorites':
      return templates.filter((t) => favorites.includes(t.id))
    case 'recent':
      // Return in order of recent usage
      return recent
        .map((id) => templates.find((t) => t.id === id))
        .filter((t): t is T => t !== undefined)
    default:
      const config = CATEGORY_CONFIG.find((c) => c.key === category)
      if (config?.codes) {
        return templates.filter((t) => config.codes!.includes(t.code))
      }
      return templates
  }
}

// Helper to count templates per category
export function countTemplatesByCategory<T extends { code: string; id: number }>(
  templates: T[],
  favorites: number[],
  recent: number[]
): Record<TemplateCategory, number> {
  const counts: Record<TemplateCategory, number> = {
    all: templates.length,
    favorites: templates.filter((t) => favorites.includes(t.id)).length,
    recent: recent.filter((id) => templates.some((t) => t.id === id)).length,
    orders: 0,
    memos: 0,
    letters: 0,
    protocols: 0,
    contracts: 0,
    other: 0,
  }

  for (const template of templates) {
    const category = getCategoryForCode(template.code)
    if (category !== 'all' && category !== 'favorites' && category !== 'recent') {
      counts[category]++
    }
  }

  return counts
}
