'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import {
  Copy,
  Check,
  User,
  Bot,
  FileText,
  ExternalLink,
  ChevronDown,
  ChevronUp,
  Loader2,
} from 'lucide-react'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { AIMessage, DocumentSource } from '@/types/ai'

interface AIMessageBubbleProps {
  message: AIMessage
  className?: string
}

export function AIMessageBubble({ message, className }: AIMessageBubbleProps) {
  const t = useTranslations('ai')
  const [copied, setCopied] = useState(false)
  const [sourcesExpanded, setSourcesExpanded] = useState(false)

  const isUser = message.role === 'user'
  const isStreaming = message.status === 'streaming'
  const isPending = message.status === 'pending'
  const isError = message.status === 'error'

  const handleCopy = async () => {
    await navigator.clipboard.writeText(message.content)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <div className={cn('group flex gap-3', isUser && 'flex-row-reverse', className)}>
      {/* Avatar */}
      <Avatar className={cn('h-8 w-8 flex-shrink-0', isUser ? 'bg-primary' : 'bg-muted')}>
        <AvatarFallback className={cn(isUser ? 'bg-primary text-primary-foreground' : 'bg-muted')}>
          {isUser ? <User className="h-4 w-4" /> : <Bot className="h-4 w-4" />}
        </AvatarFallback>
      </Avatar>

      {/* Message Content */}
      <div className={cn('flex-1 space-y-2', isUser && 'flex flex-col items-end')}>
        {/* Bubble */}
        <div
          className={cn(
            'max-w-[85%] rounded-2xl px-4 py-2.5',
            isUser ? 'rounded-br-md bg-primary text-primary-foreground' : 'rounded-bl-md bg-muted',
            isError && 'bg-destructive/10 border border-destructive/20'
          )}
        >
          {isPending ? (
            <div className="flex items-center gap-2 text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span className="text-sm">{t('thinking')}</span>
            </div>
          ) : isError ? (
            <p className="text-sm text-destructive">{message.error_message || t('error')}</p>
          ) : (
            <>
              <div className="text-sm whitespace-pre-wrap break-words prose prose-sm dark:prose-invert max-w-none">
                {message.content}
                {isStreaming && (
                  <span className="inline-block w-2 h-4 ml-1 bg-current animate-pulse" />
                )}
              </div>

              {/* Time & Actions */}
              <div
                className={cn(
                  'mt-2 flex items-center gap-2 text-[10px]',
                  isUser ? 'text-primary-foreground/70' : 'text-muted-foreground'
                )}
              >
                <span>{formatTime(message.created_at)}</span>
                {message.model && !isUser && <span className="opacity-60">{message.model}</span>}
                {!isStreaming && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-5 w-5 opacity-0 group-hover:opacity-100 transition-opacity"
                    onClick={handleCopy}
                  >
                    {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                  </Button>
                )}
              </div>
            </>
          )}
        </div>

        {/* Sources */}
        {!isUser && message.sources && message.sources.length > 0 && (
          <SourcesList
            sources={message.sources}
            expanded={sourcesExpanded}
            onToggle={() => setSourcesExpanded(!sourcesExpanded)}
          />
        )}
      </div>
    </div>
  )
}

interface SourcesListProps {
  sources: DocumentSource[]
  expanded: boolean
  onToggle: () => void
}

function SourcesList({ sources, expanded, onToggle }: SourcesListProps) {
  const t = useTranslations('ai')
  const displayedSources = expanded ? sources : sources.slice(0, 2)
  const hasMore = sources.length > 2

  return (
    <div className="max-w-[85%] space-y-2">
      <button
        onClick={onToggle}
        className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors"
      >
        <FileText className="h-3.5 w-3.5" />
        <span>
          {t('sources')} ({sources.length})
        </span>
        {hasMore &&
          (expanded ? (
            <ChevronUp className="h-3.5 w-3.5" />
          ) : (
            <ChevronDown className="h-3.5 w-3.5" />
          ))}
      </button>

      <div className="space-y-1.5">
        {displayedSources.map((source, index) => (
          <SourceCard key={`${source.document_id}-${index}`} source={source} />
        ))}
      </div>
    </div>
  )
}

interface SourceCardProps {
  source: DocumentSource
}

function SourceCard({ source }: SourceCardProps) {
  return (
    <a
      href={`/documents/${source.document_id}`}
      target="_blank"
      rel="noopener noreferrer"
      className="block p-2 rounded-lg bg-muted/50 hover:bg-muted transition-colors border border-border/50"
    >
      <div className="flex items-start gap-2">
        <FileText className="h-4 w-4 text-muted-foreground flex-shrink-0 mt-0.5" />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5">
            <p className="text-xs font-medium truncate">{source.document_title}</p>
            <ExternalLink className="h-3 w-3 text-muted-foreground flex-shrink-0" />
          </div>
          <p className="text-xs text-muted-foreground line-clamp-2 mt-0.5">{source.chunk_text}</p>
          <div className="flex items-center gap-2 mt-1 text-[10px] text-muted-foreground/70">
            <span>{Math.round(source.similarity_score * 100)}% match</span>
            {source.page_number && <span>p. {source.page_number}</span>}
          </div>
        </div>
      </div>
    </a>
  )
}
