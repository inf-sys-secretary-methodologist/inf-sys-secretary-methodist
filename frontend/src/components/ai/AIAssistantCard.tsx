'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import { useTranslations } from 'next-intl'
import {
  Send,
  Square,
  Trash2,
  Plus,
  MessageSquare,
  Settings,
  Sparkles,
  Bot,
  Loader2,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { cn } from '@/lib/utils'
import { useAIChat, useAIConversations, useDeleteAIConversation } from '@/hooks/useAIChat'
import { AIMessageBubble } from './AIMessageBubble'
import { AIQuickActions, AIQuickActionChips } from './AIQuickActions'
import type { AIConversation } from '@/types/ai'

interface AIAssistantCardProps {
  className?: string
  showHistory?: boolean
  defaultConversationId?: number | null
}

export function AIAssistantCard({
  className,
  showHistory = true,
  defaultConversationId = null,
}: AIAssistantCardProps) {
  const t = useTranslations('ai')
  const [conversationId, setConversationId] = useState<number | null>(defaultConversationId)
  const [inputValue, setInputValue] = useState('')
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [conversationToDelete, setConversationToDelete] = useState<number | null>(null)

  const viewportRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const prevMessageCountRef = useRef(0)
  const isUserScrolledUpRef = useRef(false)
  const scrollRafRef = useRef<number | null>(null)

  const { conversation, messages, isLoading, isStreaming, isPending, sendMessage, stopGeneration } =
    useAIChat(conversationId)

  const { conversations, isLoading: isLoadingConversations } = useAIConversations({ limit: 50 })
  const { mutateAsync: deleteConversation, isPending: isDeleting } = useDeleteAIConversation()

  // Check if viewport is scrolled near the bottom
  const isNearBottom = useCallback(() => {
    const viewport = viewportRef.current
    if (!viewport) return true
    const threshold = 100
    return viewport.scrollHeight - viewport.scrollTop - viewport.clientHeight < threshold
  }, [])

  // Track user scroll position
  useEffect(() => {
    const viewport = viewportRef.current
    if (!viewport) return
    const handleScroll = () => {
      isUserScrolledUpRef.current = !isNearBottom()
    }
    viewport.addEventListener('scroll', handleScroll, { passive: true })
    return () => viewport.removeEventListener('scroll', handleScroll)
  }, [isNearBottom])

  // Scroll to bottom: instant on conversation load, RAF-deduplicated during streaming
  useEffect(() => {
    const viewport = viewportRef.current
    if (!viewport) return

    const messageCount = messages.length
    const prevCount = prevMessageCountRef.current
    prevMessageCountRef.current = messageCount

    // Conversation switched or initial load — instant scroll
    if (prevCount === 0 && messageCount > 0) {
      viewport.scrollTop = viewport.scrollHeight
      isUserScrolledUpRef.current = false
      return
    }

    // Don't auto-scroll if user scrolled up
    if (isUserScrolledUpRef.current) return

    // Deduplicate scroll calls via RAF to avoid stacking smooth animations
    if (scrollRafRef.current !== null) {
      cancelAnimationFrame(scrollRafRef.current)
    }
    scrollRafRef.current = requestAnimationFrame(() => {
      scrollRafRef.current = null
      viewport.scrollTop = viewport.scrollHeight
    })
  }, [messages])

  // Cleanup scroll RAF on unmount
  useEffect(() => {
    return () => {
      if (scrollRafRef.current !== null) {
        cancelAnimationFrame(scrollRafRef.current)
      }
    }
  }, [])

  // Reset scroll tracking when conversation changes
  useEffect(() => {
    prevMessageCountRef.current = 0
    isUserScrolledUpRef.current = false
  }, [conversationId])

  // Focus input on mount
  useEffect(() => {
    inputRef.current?.focus()
  }, [conversationId])

  const handleSend = useCallback(async () => {
    if (!inputValue.trim() || isStreaming || isPending) return

    const content = inputValue.trim()
    setInputValue('')

    const newConversationId = await sendMessage(content)
    if (!conversationId && newConversationId) {
      setConversationId(newConversationId)
    }
  }, [inputValue, isStreaming, isPending, sendMessage, conversationId])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleQuickAction = (prompt: string) => {
    setInputValue(prompt)
    inputRef.current?.focus()
  }

  const handleNewConversation = () => {
    setConversationId(null)
    setInputValue('')
  }

  const handleSelectConversation = (conv: AIConversation) => {
    setConversationId(conv.id)
  }

  const handleDeleteConversation = async () => {
    if (!conversationToDelete) return

    await deleteConversation(conversationToDelete)
    if (conversationId === conversationToDelete) {
      setConversationId(null)
    }
    setDeleteDialogOpen(false)
    setConversationToDelete(null)
  }

  const confirmDelete = (id: number) => {
    setConversationToDelete(id)
    setDeleteDialogOpen(true)
  }

  const hasMessages = messages.length > 0

  return (
    <div className={cn('flex h-full overflow-hidden', className)}>
      {/* Sidebar - Conversation History */}
      {showHistory && (
        <div className="hidden md:flex md:w-72 lg:w-80 flex-col border-r">
          <ConversationSidebar
            conversations={conversations}
            currentId={conversationId}
            isLoading={isLoadingConversations}
            onSelect={handleSelectConversation}
            onNew={handleNewConversation}
            onDelete={confirmDelete}
          />
        </div>
      )}

      {/* Main Chat Area */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Header */}
        <header className="flex items-center justify-between px-4 py-3 border-b">
          <div className="flex items-center gap-3">
            {/* Mobile History Button */}
            {showHistory && (
              <Sheet>
                <SheetTrigger asChild>
                  <Button variant="ghost" size="icon" className="md:hidden">
                    <MessageSquare className="h-5 w-5" />
                  </Button>
                </SheetTrigger>
                <SheetContent side="left" className="w-72 p-0">
                  <SheetHeader className="p-4 border-b">
                    <SheetTitle>{t('history')}</SheetTitle>
                  </SheetHeader>
                  <ConversationSidebar
                    conversations={conversations}
                    currentId={conversationId}
                    isLoading={isLoadingConversations}
                    onSelect={(conv) => {
                      handleSelectConversation(conv)
                    }}
                    onNew={handleNewConversation}
                    onDelete={confirmDelete}
                    compact
                  />
                </SheetContent>
              </Sheet>
            )}

            <div className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-primary/20 to-primary/5">
                <Sparkles className="h-4 w-4 text-primary" />
              </div>
              <div>
                <h1 className="text-sm font-semibold">{t('title')}</h1>
                {conversation && (
                  <p className="text-xs text-muted-foreground truncate max-w-[200px]">
                    {conversation.title}
                  </p>
                )}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon" onClick={handleNewConversation}>
              <Plus className="h-5 w-5" />
            </Button>
            <Button variant="ghost" size="icon">
              <Settings className="h-5 w-5" />
            </Button>
          </div>
        </header>

        {/* Messages Area */}
        <ScrollArea className="flex-1" viewportRef={viewportRef}>
          <div className="p-4 space-y-4">
            {isLoading && !hasMessages ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : !hasMessages ? (
              <EmptyState onQuickAction={handleQuickAction} />
            ) : (
              messages.map((message, index) => (
                <AIMessageBubble
                  key={message.id !== -1 ? message.id : `streaming-${index}`}
                  message={message}
                />
              ))
            )}
          </div>
        </ScrollArea>

        {/* Quick Action Chips */}
        {hasMessages && (
          <div className="px-4 py-2 border-t">
            <AIQuickActionChips onAction={handleQuickAction} disabled={isStreaming || isPending} />
          </div>
        )}

        {/* Input Area */}
        <div className="p-4 border-t">
          <div className="flex items-center gap-2">
            <Input
              ref={inputRef}
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={t('inputPlaceholder')}
              disabled={isStreaming}
              className="flex-1"
            />
            {isStreaming ? (
              <Button variant="destructive" size="icon" onClick={stopGeneration}>
                <Square className="h-4 w-4" />
              </Button>
            ) : (
              <Button size="icon" onClick={handleSend} disabled={!inputValue.trim() || isPending}>
                <Send className="h-4 w-4" />
              </Button>
            )}
          </div>
          <p className="text-xs text-muted-foreground mt-2 text-center">{t('disclaimer')}</p>
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('deleteConversation')}</AlertDialogTitle>
            <AlertDialogDescription>{t('deleteConversationConfirm')}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('cancel')}</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteConversation}
              disabled={isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? <Loader2 className="h-4 w-4 animate-spin" /> : t('delete')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

interface ConversationSidebarProps {
  conversations: AIConversation[]
  currentId: number | null
  isLoading: boolean
  onSelect: (conversation: AIConversation) => void
  onNew: () => void
  onDelete: (id: number) => void
  compact?: boolean
}

function ConversationSidebar({
  conversations,
  currentId,
  isLoading,
  onSelect,
  onNew,
  onDelete,
  compact = false,
}: ConversationSidebarProps) {
  const t = useTranslations('ai')

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24))

    if (diffDays === 0) return t('today')
    if (diffDays === 1) return t('yesterday')
    if (diffDays < 7) return t('daysAgo', { days: diffDays })
    return date.toLocaleDateString()
  }

  return (
    <div className="flex flex-col h-full">
      {!compact && (
        <div className="p-4">
          <Button onClick={onNew} className="w-full gap-2">
            <Plus className="h-4 w-4" />
            {t('newConversation')}
          </Button>
        </div>
      )}

      <ScrollArea className="flex-1">
        <div className={cn('space-y-1', compact ? 'p-2' : 'px-2 pb-4')}>
          {isLoading ? (
            <div className="flex items-center justify-center py-4">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : conversations.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-4">{t('noConversations')}</p>
          ) : (
            conversations.map((conv) => (
              <div
                key={conv.id}
                className={cn(
                  'group relative flex items-center gap-2 rounded-lg px-3 py-2 cursor-pointer transition-colors',
                  currentId === conv.id ? 'bg-primary/10 text-primary' : 'hover:bg-muted'
                )}
                onClick={() => onSelect(conv)}
              >
                <MessageSquare className="h-4 w-4 flex-shrink-0" />
                <div className="flex-1 min-w-0 pr-6">
                  <p className="text-sm font-medium truncate">{conv.title}</p>
                  <p className="text-xs text-muted-foreground truncate">
                    {formatDate(conv.updated_at)}
                  </p>
                </div>
                <button
                  className="absolute right-2 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 transition-opacity p-1 rounded-md hover:bg-destructive/10 text-muted-foreground hover:text-destructive"
                  onClick={(e) => {
                    e.stopPropagation()
                    onDelete(conv.id)
                  }}
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              </div>
            ))
          )}
        </div>
      </ScrollArea>
    </div>
  )
}

interface EmptyStateProps {
  onQuickAction: (prompt: string) => void
}

function EmptyState({ onQuickAction }: EmptyStateProps) {
  const t = useTranslations('ai')

  return (
    <div className="flex flex-col items-center justify-center py-12 px-4">
      <div className="relative mb-6">
        <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-primary/20 to-primary/5 shadow-lg">
          <Bot className="h-8 w-8 text-primary" />
        </div>
        <div className="absolute -top-1 -right-1 flex h-6 w-6 items-center justify-center rounded-full bg-blue-500/10">
          <Sparkles className="h-3 w-3 text-blue-500" />
        </div>
      </div>

      <h2 className="text-xl font-semibold mb-2">{t('welcomeTitle')}</h2>
      <p className="text-muted-foreground text-center max-w-md mb-8">{t('welcomeDescription')}</p>

      <AIQuickActions onAction={onQuickAction} />
    </div>
  )
}
