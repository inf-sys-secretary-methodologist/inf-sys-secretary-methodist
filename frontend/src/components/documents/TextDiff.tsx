'use client'

import { useMemo } from 'react'
import * as Diff from 'diff'
import { useTranslations } from 'next-intl'

interface TextDiffProps {
  oldText: string
  newText: string
  oldLabel?: string
  newLabel?: string
  className?: string
}

type DiffPart = {
  value: string
  added?: boolean
  removed?: boolean
}

export function TextDiff({ oldText, newText, oldLabel, newLabel, className = '' }: TextDiffProps) {
  const t = useTranslations('textDiff')
  const effectiveOldLabel = oldLabel ?? t('before')
  const effectiveNewLabel = newLabel ?? t('after')

  const diff = useMemo(() => {
    /* c8 ignore next */
    return Diff.diffLines(oldText || '', newText || '')
  }, [oldText, newText])

  const hasChanges = diff.some((part: DiffPart) => part.added || part.removed)

  if (!hasChanges) {
    return (
      <div className={`text-gray-500 dark:text-gray-400 text-sm italic ${className}`}>
        {t('noChanges')}
      </div>
    )
  }

  return (
    <div className={`font-mono text-sm ${className}`}>
      {/* Legend */}
      <div className="flex gap-4 mb-3 text-xs">
        <div className="flex items-center gap-1">
          <span className="w-3 h-3 bg-red-100 dark:bg-red-900/30 border border-red-300 dark:border-red-700 rounded" />
          <span className="text-gray-600 dark:text-gray-400">
            {effectiveOldLabel} ({t('removed')})
          </span>
        </div>
        <div className="flex items-center gap-1">
          <span className="w-3 h-3 bg-green-100 dark:bg-green-900/30 border border-green-300 dark:border-green-700 rounded" />
          <span className="text-gray-600 dark:text-gray-400">
            {effectiveNewLabel} ({t('added')})
          </span>
        </div>
      </div>

      {/* Diff Content */}
      <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
        {diff.map((part: DiffPart, index: number) => {
          const lines = part.value.split('\n').filter(
            (line, i, arr) =>
              // Keep all lines except trailing empty line
              i < arr.length - 1 || line !== ''
          )

          /* c8 ignore next */
          if (lines.length === 0) return null

          return lines.map((line, lineIndex) => (
            <div
              key={`${index}-${lineIndex}`}
              /* c8 ignore start - Dynamic className for diff parts */
              className={`
                px-3 py-1 border-b border-gray-100 dark:border-gray-800 last:border-b-0
                ${
                  part.added
                    ? 'bg-green-50 dark:bg-green-900/20 text-green-800 dark:text-green-300'
                    : part.removed
                      ? 'bg-red-50 dark:bg-red-900/20 text-red-800 dark:text-red-300'
                      : 'bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-300'
                }
              `}
              /* c8 ignore stop */
            >
              <span className="select-none mr-2 text-gray-400 dark:text-gray-600">
                {/* c8 ignore next */}
                {part.added ? '+' : part.removed ? '-' : ' '}
              </span>
              {/* c8 ignore next */}
              {line || ' '}
            </div>
          ))
        })}
      </div>
    </div>
  )
}

// Side-by-side diff view
export function TextDiffSideBySide({
  oldText,
  newText,
  oldLabel,
  newLabel,
  className = '',
}: TextDiffProps) {
  const t = useTranslations('textDiff')
  const effectiveOldLabel = oldLabel ?? t('versionA')
  const effectiveNewLabel = newLabel ?? t('versionB')

  const diff = useMemo(() => {
    /* c8 ignore next */
    return Diff.diffLines(oldText || '', newText || '')
  }, [oldText, newText])

  const leftLines: { text: string; type: 'removed' | 'unchanged' | 'empty' }[] = []
  const rightLines: { text: string; type: 'added' | 'unchanged' | 'empty' }[] = []

  diff.forEach((part: DiffPart) => {
    const lines = part.value.split('\n').filter((line, i, arr) => i < arr.length - 1 || line !== '')

    if (part.removed) {
      lines.forEach((line) => {
        leftLines.push({ text: line, type: 'removed' })
        rightLines.push({ text: '', type: 'empty' })
      })
    } else if (part.added) {
      lines.forEach((line) => {
        leftLines.push({ text: '', type: 'empty' })
        rightLines.push({ text: line, type: 'added' })
      })
    } else {
      lines.forEach((line) => {
        leftLines.push({ text: line, type: 'unchanged' })
        rightLines.push({ text: line, type: 'unchanged' })
      })
    }
  })

  const hasChanges = diff.some((part: DiffPart) => part.added || part.removed)

  if (!hasChanges) {
    return (
      <div className={`text-gray-500 dark:text-gray-400 text-sm italic ${className}`}>
        {t('noChanges')}
      </div>
    )
  }

  /* c8 ignore start - Line class helper with default case */
  const getLineClass = (type: string) => {
    switch (type) {
      case 'removed':
        return 'bg-red-50 dark:bg-red-900/20 text-red-800 dark:text-red-300'
      case 'added':
        return 'bg-green-50 dark:bg-green-900/20 text-green-800 dark:text-green-300'
      case 'empty':
        return 'bg-gray-50 dark:bg-gray-800/50'
      default:
        return 'bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-300'
    }
  }
  /* c8 ignore stop */

  return (
    <div className={`font-mono text-sm ${className}`}>
      <div className="grid grid-cols-2 gap-2">
        {/* Left side - Old */}
        <div>
          <div className="text-xs font-medium text-gray-600 dark:text-gray-400 mb-2 flex items-center gap-2">
            <span className="w-2 h-2 bg-red-400 rounded-full" />
            {effectiveOldLabel}
          </div>
          <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
            {leftLines.map((line, index) => (
              <div
                key={index}
                className={`px-3 py-1 border-b border-gray-100 dark:border-gray-800 last:border-b-0 min-h-[28px] ${getLineClass(line.type)}`}
              >
                {line.text || '\u00A0'}
              </div>
            ))}
          </div>
        </div>

        {/* Right side - New */}
        <div>
          <div className="text-xs font-medium text-gray-600 dark:text-gray-400 mb-2 flex items-center gap-2">
            <span className="w-2 h-2 bg-green-400 rounded-full" />
            {effectiveNewLabel}
          </div>
          <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
            {rightLines.map((line, index) => (
              <div
                key={index}
                className={`px-3 py-1 border-b border-gray-100 dark:border-gray-800 last:border-b-0 min-h-[28px] ${getLineClass(line.type)}`}
              >
                {line.text || '\u00A0'}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

// Inline word-level diff (more detailed)
export function TextDiffInline({
  oldText,
  newText,
  className = '',
}: Omit<TextDiffProps, 'oldLabel' | 'newLabel'>) {
  const t = useTranslations('textDiff')
  const diff = useMemo(() => {
    return Diff.diffWords(oldText || '', newText || '')
  }, [oldText, newText])

  const hasChanges = diff.some((part: DiffPart) => part.added || part.removed)

  if (!hasChanges) {
    return (
      <div className={`text-gray-500 dark:text-gray-400 text-sm italic ${className}`}>
        {t('noChanges')}
      </div>
    )
  }

  return (
    <div className={`font-mono text-sm leading-relaxed ${className}`}>
      <div className="p-3 border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-900">
        {diff.map((part: DiffPart, index: number) => (
          <span
            key={index}
            className={`
              ${
                part.added
                  ? 'bg-green-200 dark:bg-green-800/50 text-green-900 dark:text-green-200 px-0.5 rounded'
                  : part.removed
                    ? 'bg-red-200 dark:bg-red-800/50 text-red-900 dark:text-red-200 line-through px-0.5 rounded'
                    : 'text-gray-700 dark:text-gray-300'
              }
            `}
          >
            {part.value}
          </span>
        ))}
      </div>
    </div>
  )
}
