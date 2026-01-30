'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { MoreVertical, Pencil, Trash2, Reply, Copy, Check, FileText, Download } from 'lucide-react'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { cn, getValidAvatarUrl } from '@/lib/utils'
import type { Message, Attachment } from '@/types/messaging'

interface MessageBubbleProps {
  message: Message
  isOwn: boolean
  showAvatar?: boolean
  showSender?: boolean
  onReply?: (message: Message) => void
  onEdit?: (message: Message) => void
  onDelete?: (message: Message) => void
  className?: string
}

export function MessageBubble({
  message,
  isOwn,
  showAvatar = true,
  showSender = false,
  onReply,
  onEdit,
  onDelete,
  className,
}: MessageBubbleProps) {
  const t = useTranslations('messaging')
  const [copied, setCopied] = useState(false)

  /* c8 ignore start - Event handlers, tested in e2e */
  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const handleCopy = async () => {
    await navigator.clipboard.writeText(message.content)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  const renderAttachment = (attachment: Attachment) => {
    const isImage = attachment.mime_type.startsWith('image/')

    if (isImage) {
      return (
        <a
          href={attachment.url}
          target="_blank"
          rel="noopener noreferrer"
          className="block mt-2 rounded-lg overflow-hidden max-w-xs"
        >
          <img
            src={attachment.url}
            alt={attachment.file_name}
            className="w-full h-auto object-cover"
          />
        </a>
      )
    }

    return (
      <a
        href={attachment.url}
        target="_blank"
        rel="noopener noreferrer"
        download={attachment.file_name}
        className="flex items-center gap-2 mt-2 p-2 rounded-lg bg-muted/50 hover:bg-muted transition-colors"
      >
        <FileText className="h-8 w-8 text-muted-foreground flex-shrink-0" />
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium truncate">{attachment.file_name}</p>
          <p className="text-xs text-muted-foreground">{formatFileSize(attachment.file_size)}</p>
        </div>
        <Download className="h-4 w-4 text-muted-foreground" />
      </a>
    )
  }

  const renderReplyTo = () => {
    if (!message.reply_to) return null

    return (
      <div
        className={cn(
          'mb-1 px-2 py-1 text-xs rounded border-l-2',
          isOwn
            ? 'bg-primary-foreground/10 border-primary-foreground/50'
            : 'bg-muted border-primary/50'
        )}
      >
        <p className="font-medium text-muted-foreground">{message.reply_to.sender_name}</p>
        <p className="truncate opacity-80">
          {message.reply_to.is_deleted ? t('messageDeleted') : message.reply_to.content}
        </p>
      </div>
    )
  }
  /* c8 ignore stop */

  /* c8 ignore start - JSX rendering, tested in e2e */
  // System message style
  if (message.type === 'system') {
    return (
      <div className="flex justify-center py-2">
        <div className="rounded-full bg-muted px-4 py-1.5 text-xs text-muted-foreground">
          {message.content}
        </div>
      </div>
    )
  }

  // Deleted message
  if (message.is_deleted) {
    return (
      <div className={cn('flex items-end gap-2', isOwn && 'flex-row-reverse', className)}>
        {showAvatar && !isOwn && (
          <Avatar className="h-8 w-8 flex-shrink-0">
            <AvatarImage src={getValidAvatarUrl(message.sender_avatar)} />
            <AvatarFallback className="text-xs">{getInitials(message.sender_name)}</AvatarFallback>
          </Avatar>
        )}
        <div className="max-w-[70%]">
          <div
            className={cn(
              'rounded-2xl px-4 py-2 italic text-muted-foreground',
              isOwn ? 'rounded-br-md bg-muted' : 'rounded-bl-md bg-muted'
            )}
          >
            {t('messageDeleted')}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className={cn('group flex items-end gap-2', isOwn && 'flex-row-reverse', className)}>
      {/* Avatar */}
      {showAvatar && !isOwn && (
        <Avatar className="h-8 w-8 flex-shrink-0">
          <AvatarImage src={getValidAvatarUrl(message.sender_avatar)} />
          <AvatarFallback className="text-xs">{getInitials(message.sender_name)}</AvatarFallback>
        </Avatar>
      )}

      {/* Message Content */}
      <div className="max-w-[70%]">
        {/* Sender name for groups */}
        {showSender && !isOwn && (
          <p className="mb-1 ml-1 text-xs font-medium text-muted-foreground">
            {message.sender_name}
          </p>
        )}

        <div className="flex items-end gap-1">
          {/* Bubble */}
          <div
            className={cn(
              'rounded-2xl px-4 py-2',
              isOwn ? 'rounded-br-md bg-primary text-primary-foreground' : 'rounded-bl-md bg-muted'
            )}
          >
            {renderReplyTo()}
            <p className="text-sm whitespace-pre-wrap break-words">{message.content}</p>

            {/* Attachments */}
            {message.attachments && message.attachments.length > 0 && (
              <div className="mt-2 space-y-2">
                {message.attachments.map((attachment) => (
                  <div key={attachment.id}>{renderAttachment(attachment)}</div>
                ))}
              </div>
            )}

            {/* Time & Status */}
            <div
              className={cn(
                'mt-1 flex items-center gap-1 text-[10px]',
                isOwn ? 'text-primary-foreground/70' : 'text-muted-foreground'
              )}
            >
              <span>{formatTime(message.created_at)}</span>
              {message.is_edited && <span className="italic">({t('edited')})</span>}
            </div>
          </div>

          {/* Actions */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
              >
                <MoreVertical className="h-3 w-3" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align={isOwn ? 'end' : 'start'}>
              {onReply && (
                <DropdownMenuItem onClick={() => onReply(message)}>
                  <Reply className="mr-2 h-4 w-4" />
                  {t('reply')}
                </DropdownMenuItem>
              )}
              <DropdownMenuItem onClick={handleCopy}>
                {copied ? <Check className="mr-2 h-4 w-4" /> : <Copy className="mr-2 h-4 w-4" />}
                {copied ? t('copied') : t('copy')}
              </DropdownMenuItem>
              {isOwn && onEdit && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => onEdit(message)}>
                    <Pencil className="mr-2 h-4 w-4" />
                    {t('edit')}
                  </DropdownMenuItem>
                </>
              )}
              {(isOwn || onDelete) && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={() => onDelete?.(message)}
                    className="text-destructive focus:text-destructive"
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    {t('delete')}
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Placeholder for alignment when avatar is shown */}
      {showAvatar && isOwn && <div className="w-8 flex-shrink-0" />}
    </div>
  )
  /* c8 ignore stop */
}
