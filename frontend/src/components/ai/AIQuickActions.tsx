'use client'

import { useTranslations } from 'next-intl'
import {
  FileSearch,
  HelpCircle,
  Calendar,
  FileText,
  Users,
  BarChart3,
  BookOpen,
  Sparkles,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { AIQuickAction } from '@/types/ai'

interface AIQuickActionsProps {
  onAction: (prompt: string) => void
  disabled?: boolean
  className?: string
}

export function AIQuickActions({ onAction, disabled, className }: AIQuickActionsProps) {
  const t = useTranslations('ai')

  const quickActions: AIQuickAction[] = [
    {
      id: 'search-documents',
      label: t('quickActions.searchDocuments'),
      prompt: t('quickActions.searchDocumentsPrompt'),
      icon: 'FileSearch',
      category: 'documents',
    },
    {
      id: 'summarize-document',
      label: t('quickActions.summarizeDocument'),
      prompt: t('quickActions.summarizeDocumentPrompt'),
      icon: 'FileText',
      category: 'documents',
    },
    {
      id: 'schedule-help',
      label: t('quickActions.scheduleHelp'),
      prompt: t('quickActions.scheduleHelpPrompt'),
      icon: 'Calendar',
      category: 'schedule',
    },
    {
      id: 'student-analytics',
      label: t('quickActions.studentAnalytics'),
      prompt: t('quickActions.studentAnalyticsPrompt'),
      icon: 'BarChart3',
      category: 'analytics',
    },
    {
      id: 'find-template',
      label: t('quickActions.findTemplate'),
      prompt: t('quickActions.findTemplatePrompt'),
      icon: 'BookOpen',
      category: 'templates',
    },
    {
      id: 'help',
      label: t('quickActions.help'),
      prompt: t('quickActions.helpPrompt'),
      icon: 'HelpCircle',
      category: 'general',
    },
  ]

  const getIcon = (iconName?: string) => {
    switch (iconName) {
      case 'FileSearch':
        return <FileSearch className="h-4 w-4" />
      case 'FileText':
        return <FileText className="h-4 w-4" />
      case 'Calendar':
        return <Calendar className="h-4 w-4" />
      case 'BarChart3':
        return <BarChart3 className="h-4 w-4" />
      case 'BookOpen':
        return <BookOpen className="h-4 w-4" />
      case 'Users':
        return <Users className="h-4 w-4" />
      case 'HelpCircle':
        return <HelpCircle className="h-4 w-4" />
      default:
        return <Sparkles className="h-4 w-4" />
    }
  }

  return (
    <div className={cn('space-y-3', className)}>
      <p className="text-sm text-muted-foreground">{t('quickActionsTitle')}</p>
      <div className="flex flex-wrap gap-2">
        {quickActions.map((action) => (
          <Button
            key={action.id}
            variant="outline"
            size="sm"
            disabled={disabled}
            onClick={() => onAction(action.prompt)}
            className="gap-2 h-auto py-2 px-3"
          >
            {getIcon(action.icon)}
            <span>{action.label}</span>
          </Button>
        ))}
      </div>
    </div>
  )
}

// Compact version for sidebar
interface AIQuickActionChipsProps {
  onAction: (prompt: string) => void
  disabled?: boolean
  className?: string
}

export function AIQuickActionChips({ onAction, disabled, className }: AIQuickActionChipsProps) {
  const t = useTranslations('ai')

  const chips = [
    { label: t('chips.search'), prompt: t('quickActions.searchDocumentsPrompt') },
    { label: t('chips.summarize'), prompt: t('quickActions.summarizeDocumentPrompt') },
    { label: t('chips.schedule'), prompt: t('quickActions.scheduleHelpPrompt') },
    { label: t('chips.analytics'), prompt: t('quickActions.studentAnalyticsPrompt') },
  ]

  return (
    <div className={cn('flex flex-wrap gap-1.5', className)}>
      {chips.map((chip) => (
        <button
          key={chip.label}
          disabled={disabled}
          onClick={() => onAction(chip.prompt)}
          className={cn(
            'px-2.5 py-1 rounded-full text-xs font-medium',
            'bg-muted hover:bg-muted/80 transition-colors',
            'disabled:opacity-50 disabled:cursor-not-allowed'
          )}
        >
          {chip.label}
        </button>
      ))}
    </div>
  )
}
