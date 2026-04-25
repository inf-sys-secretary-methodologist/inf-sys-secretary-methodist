'use client'

import * as React from 'react'
import { useState } from 'react'
import { format } from 'date-fns'
import { useTranslations } from 'next-intl'

import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  ANNOUNCEMENT_PRIORITIES,
  TARGET_AUDIENCES,
  type Announcement,
  type AnnouncementPriority,
  type CreateAnnouncementInput,
  type TargetAudience,
} from '@/types/announcements'

interface AnnouncementFormProps {
  announcement?: Announcement
  onSubmit: (input: CreateAnnouncementInput) => Promise<void> | void
  onCancel: () => void
  className?: string
}

const toDateInputValue = (iso?: string) => (iso ? format(new Date(iso), 'yyyy-MM-dd') : '')

export function AnnouncementForm({
  announcement,
  onSubmit,
  onCancel,
  className,
}: AnnouncementFormProps) {
  const t = useTranslations('announcements')

  const [title, setTitle] = useState(announcement?.title ?? '')
  const [content, setContent] = useState(announcement?.content ?? '')
  const [summary, setSummary] = useState(announcement?.summary ?? '')
  const [priority, setPriority] = useState<AnnouncementPriority>(
    announcement?.priority ?? 'normal'
  )
  const [audience, setAudience] = useState<TargetAudience>(
    announcement?.target_audience ?? 'all'
  )
  const [isPinned, setIsPinned] = useState(announcement?.is_pinned ?? false)
  const [tagsInput, setTagsInput] = useState(announcement?.tags?.join(', ') ?? '')
  const [publishAt, setPublishAt] = useState(toDateInputValue(announcement?.publish_at))
  const [expireAt, setExpireAt] = useState(toDateInputValue(announcement?.expire_at))

  const [titleError, setTitleError] = useState<string | null>(null)
  const [contentError, setContentError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    let hasError = false
    if (!title.trim()) {
      setTitleError(t('errors.titleRequired'))
      hasError = true
    }
    if (!content.trim()) {
      setContentError(t('errors.contentRequired'))
      hasError = true
    }
    if (hasError) return

    setTitleError(null)
    setContentError(null)
    setSubmitting(true)
    try {
      const input: CreateAnnouncementInput = {
        title: title.trim(),
        content: content.trim(),
        priority,
        target_audience: audience,
        is_pinned: isPinned,
      }
      if (summary.trim()) input.summary = summary.trim()
      // Parse "YYYY-MM-DD" from <input type="date"> as local midnight,
      // not UTC midnight. new Date("2026-04-30") would produce UTC midnight,
      // shifting the date by the user's UTC offset; appending T00:00:00
      // (no Z) makes the JS Date constructor treat it as local time.
      if (publishAt) input.publish_at = new Date(`${publishAt}T00:00:00`).toISOString()
      if (expireAt) input.expire_at = new Date(`${expireAt}T00:00:00`).toISOString()
      const tagsList = tagsInput
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean)
      if (tagsList.length > 0) input.tags = tagsList
      await onSubmit(input)
    } finally {
      setSubmitting(false)
    }
  }

  const selectClass =
    'flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring'
  const textareaClass =
    'flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring'

  return (
    <form onSubmit={handleSubmit} className={cn('flex flex-col gap-4', className)}>
      <div>
        <label
          htmlFor="ann-form-title"
          className="block text-sm font-medium text-foreground mb-1"
        >
          {t('form.titleLabel')}
        </label>
        <Input
          id="ann-form-title"
          value={title}
          onChange={(e) => {
            setTitle(e.target.value)
            if (titleError) setTitleError(null)
          }}
          placeholder={t('form.titlePlaceholder')}
          aria-invalid={!!titleError}
          aria-describedby={titleError ? 'ann-form-title-error' : undefined}
        />
        {titleError && (
          <p id="ann-form-title-error" className="mt-1 text-sm text-red-600 dark:text-red-400">
            {titleError}
          </p>
        )}
      </div>

      <div>
        <label
          htmlFor="ann-form-summary"
          className="block text-sm font-medium text-foreground mb-1"
        >
          {t('form.summaryLabel')}
        </label>
        <textarea
          id="ann-form-summary"
          value={summary}
          onChange={(e) => setSummary(e.target.value)}
          placeholder={t('form.summaryPlaceholder')}
          rows={2}
          maxLength={1000}
          className={textareaClass}
        />
      </div>

      <div>
        <label
          htmlFor="ann-form-content"
          className="block text-sm font-medium text-foreground mb-1"
        >
          {t('form.contentLabel')}
        </label>
        <textarea
          id="ann-form-content"
          value={content}
          onChange={(e) => {
            setContent(e.target.value)
            if (contentError) setContentError(null)
          }}
          placeholder={t('form.contentPlaceholder')}
          rows={8}
          className={textareaClass}
          aria-invalid={!!contentError}
          aria-describedby={contentError ? 'ann-form-content-error' : undefined}
        />
        {contentError && (
          <p id="ann-form-content-error" className="mt-1 text-sm text-red-600 dark:text-red-400">
            {contentError}
          </p>
        )}
      </div>

      <div className="grid gap-3 sm:grid-cols-2">
        <div>
          <label
            htmlFor="ann-form-priority"
            className="block text-sm font-medium text-foreground mb-1"
          >
            {t('form.priorityLabel')}
          </label>
          <select
            id="ann-form-priority"
            value={priority}
            onChange={(e) => setPriority(e.target.value as AnnouncementPriority)}
            className={selectClass}
          >
            {ANNOUNCEMENT_PRIORITIES.map((p) => (
              <option key={p} value={p}>
                {t(`priority.${p}`)}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label
            htmlFor="ann-form-audience"
            className="block text-sm font-medium text-foreground mb-1"
          >
            {t('form.audienceLabel')}
          </label>
          <select
            id="ann-form-audience"
            value={audience}
            onChange={(e) => setAudience(e.target.value as TargetAudience)}
            className={selectClass}
          >
            {TARGET_AUDIENCES.map((a) => (
              <option key={a} value={a}>
                {t(`audience.${a}`)}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="grid gap-3 sm:grid-cols-2">
        <div>
          <label
            htmlFor="ann-form-publish-at"
            className="block text-sm font-medium text-foreground mb-1"
          >
            {t('form.publishAtLabel')}
          </label>
          <Input
            id="ann-form-publish-at"
            type="date"
            value={publishAt}
            onChange={(e) => setPublishAt(e.target.value)}
          />
        </div>

        <div>
          <label
            htmlFor="ann-form-expire-at"
            className="block text-sm font-medium text-foreground mb-1"
          >
            {t('form.expireAtLabel')}
          </label>
          <Input
            id="ann-form-expire-at"
            type="date"
            value={expireAt}
            onChange={(e) => setExpireAt(e.target.value)}
          />
        </div>
      </div>

      <div>
        <label
          htmlFor="ann-form-tags"
          className="block text-sm font-medium text-foreground mb-1"
        >
          {t('form.tagsLabel')}
        </label>
        <Input
          id="ann-form-tags"
          value={tagsInput}
          onChange={(e) => setTagsInput(e.target.value)}
          placeholder={t('form.tagsPlaceholder')}
        />
      </div>

      <label
        htmlFor="ann-form-pinned"
        className="flex items-center gap-2 text-sm text-foreground cursor-pointer"
      >
        <input
          id="ann-form-pinned"
          type="checkbox"
          checked={isPinned}
          onChange={(e) => setIsPinned(e.target.checked)}
          className="h-4 w-4 rounded border-input"
        />
        <span>{t('form.pinnedLabel')}</span>
      </label>

      <div className="flex gap-2 justify-end">
        <Button type="button" variant="outline" onClick={onCancel} disabled={submitting}>
          {t('form.cancel')}
        </Button>
        <Button type="submit" disabled={submitting}>
          {t('form.save')}
        </Button>
      </div>
    </form>
  )
}
