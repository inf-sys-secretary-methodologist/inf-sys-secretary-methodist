'use client'

import { useState, useRef, useCallback, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import {
  Send,
  Paperclip,
  Smile,
  X,
  Image as ImageIcon,
  FileText,
  Sparkles,
  Mic,
  Calendar,
  CheckSquare,
  Loader2,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import type { AttachmentInput } from '@/types/messaging'

interface MessageInputProps {
  onSend: (content: string, attachments?: AttachmentInput[]) => Promise<void>
  onTyping?: () => void
  onStopTyping?: () => void
  disabled?: boolean
  placeholder?: string
  className?: string
  showAiSuggestions?: boolean
}

interface SuggestionChip {
  icon: React.ElementType
  label: string
  color: string
  prompt: string
}

const AI_SUGGESTIONS: SuggestionChip[] = [
  {
    icon: Calendar,
    label: 'scheduleEvent',
    color: 'text-blue-500',
    prompt: 'Давай запланируем встречу...',
  },
  {
    icon: CheckSquare,
    label: 'createTask',
    color: 'text-green-500',
    prompt: 'Создай задачу...',
  },
  {
    icon: FileText,
    label: 'shareDocument',
    color: 'text-amber-500',
    prompt: 'Поделиться документом...',
  },
  {
    icon: Sparkles,
    label: 'askAI',
    color: 'text-purple-500',
    prompt: 'AI помоги мне...',
  },
]

export function MessageInput({
  onSend,
  onTyping,
  onStopTyping,
  disabled = false,
  placeholder,
  className,
  showAiSuggestions = false,
}: MessageInputProps) {
  const t = useTranslations('messaging')
  const [content, setContent] = useState('')
  const [attachments, setAttachments] = useState<AttachmentInput[]>([])
  const [isSending, setIsSending] = useState(false)
  const [showSuggestions, setShowSuggestions] = useState(true)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  const handleTyping = useCallback(() => {
    if (onTyping) {
      onTyping()
    }

    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current)
    }

    typingTimeoutRef.current = setTimeout(() => {
      if (onStopTyping) {
        onStopTyping()
      }
    }, 2000)
  }, [onTyping, onStopTyping])

  useEffect(() => {
    return () => {
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current)
      }
    }
  }, [])

  const handleSend = async () => {
    const trimmedContent = content.trim()
    if (!trimmedContent && attachments.length === 0) return
    if (isSending) return

    setIsSending(true)
    try {
      await onSend(trimmedContent, attachments.length > 0 ? attachments : undefined)
      setContent('')
      setAttachments([])
      setShowSuggestions(true)
      if (onStopTyping) {
        onStopTyping()
      }
    } finally {
      setIsSending(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleContentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value
    setContent(value)
    handleTyping()

    // Hide suggestions when typing
    if (value.trim()) {
      setShowSuggestions(false)
    } else {
      setShowSuggestions(true)
    }
  }

  const handleSuggestionClick = (suggestion: SuggestionChip) => {
    setContent(suggestion.prompt)
    setShowSuggestions(false)
    textareaRef.current?.focus()
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (!files) return

    // TODO: Implement file upload through files API
    // For now, just show a placeholder
    Array.from(files).forEach((file) => {
      const attachment: AttachmentInput = {
        file_id: Date.now(), // Placeholder - should come from file upload
        file_name: file.name,
        file_size: file.size,
        mime_type: file.type,
        url: URL.createObjectURL(file), // Placeholder
      }
      setAttachments((prev) => [...prev, attachment])
    })

    e.target.value = ''
  }

  const removeAttachment = (index: number) => {
    setAttachments((prev) => prev.filter((_, i) => i !== index))
  }

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  const canSend = (content.trim() || attachments.length > 0) && !disabled && !isSending

  return (
    <div className={cn('flex flex-col', className)}>
      {/* AI Suggestions */}
      {showAiSuggestions && showSuggestions && !content && (
        <div className="mb-4 flex flex-wrap justify-center gap-2 px-4">
          {AI_SUGGESTIONS.map((suggestion) => (
            <Badge
              key={suggestion.label}
              variant="secondary"
              className="h-7 min-w-7 cursor-pointer gap-1.5 text-xs rounded-md hover:bg-secondary/80 transition-colors"
              onClick={() => handleSuggestionClick(suggestion)}
            >
              <suggestion.icon className={cn('h-3.5 w-3.5', suggestion.color)} />
              {t(`suggestions.${suggestion.label}`)}
            </Badge>
          ))}
        </div>
      )}

      {/* Attachments Preview */}
      {attachments.length > 0 && (
        <div className="mx-4 mb-2 flex flex-wrap gap-2">
          {attachments.map((attachment, index) => (
            <div
              key={index}
              className="flex items-center gap-2 rounded-md bg-muted px-3 py-2 text-sm"
            >
              {attachment.mime_type.startsWith('image/') ? (
                <ImageIcon className="h-4 w-4 text-muted-foreground" />
              ) : (
                <FileText className="h-4 w-4 text-muted-foreground" />
              )}
              <span className="max-w-32 truncate">{attachment.file_name}</span>
              <span className="text-xs text-muted-foreground">
                ({formatFileSize(attachment.file_size)})
              </span>
              <button
                onClick={() => removeAttachment(index)}
                className="ml-1 rounded-full p-0.5 hover:bg-accent"
              >
                <X className="h-3 w-3" />
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Input Area */}
      <div className="relative rounded-lg border bg-background ring-1 ring-border mx-4 mb-4">
        <div className="relative">
          <Textarea
            ref={textareaRef}
            value={content}
            onChange={handleContentChange}
            onKeyDown={handleKeyDown}
            placeholder={placeholder || t('typeMessage')}
            disabled={disabled}
            className="min-h-[100px] resize-none rounded-b-none border-none py-3 ps-4 pe-12 shadow-none focus-visible:ring-0 bg-transparent"
          />

          {/* Microphone button */}
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  type="button"
                  className="absolute right-3 bottom-3 rounded-full p-2 text-muted-foreground hover:text-foreground hover:bg-accent transition-colors"
                  disabled={disabled}
                >
                  <Mic className="h-4 w-4" />
                </button>
              </TooltipTrigger>
              <TooltipContent>{t('voiceMessage')}</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>

        {/* Bottom Bar */}
        <div className="flex items-center justify-between rounded-b-lg border-t bg-muted/50 px-3 py-2">
          {/* Left Actions */}
          <div className="flex items-center gap-1">
            <input
              ref={fileInputRef}
              type="file"
              multiple
              onChange={handleFileSelect}
              className="hidden"
              accept="image/*,.pdf,.doc,.docx,.xls,.xlsx,.txt"
            />
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="h-7 gap-1.5 text-xs"
              onClick={() => fileInputRef.current?.click()}
              disabled={disabled}
            >
              <Paperclip className="h-3.5 w-3.5 text-muted-foreground" />
              {t('attach')}
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="h-7 gap-1.5 text-xs"
              disabled={disabled}
            >
              <Smile className="h-3.5 w-3.5 text-muted-foreground" />
              {t('emoji')}
            </Button>
          </div>

          {/* Right Actions */}
          <Button
            type="button"
            size="sm"
            className="h-7 gap-1.5 px-3"
            onClick={handleSend}
            disabled={!canSend}
          >
            {isSending ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Send className="h-3.5 w-3.5" />
            )}
            {t('send')}
          </Button>
        </div>
      </div>
    </div>
  )
}
