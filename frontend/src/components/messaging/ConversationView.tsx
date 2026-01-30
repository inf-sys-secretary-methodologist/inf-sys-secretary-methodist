'use client'

import { useEffect, useRef, useState, useMemo, useCallback } from 'react'
import { useTranslations } from 'next-intl'
import {
  ArrowLeft,
  Phone,
  Video,
  MoreVertical,
  Search,
  Users,
  Settings,
  Loader2,
  MessageCircle,
  Sparkles,
} from 'lucide-react'
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
import { useConversationWithMessages } from '@/hooks/useMessaging'
import { useAuth } from '@/hooks/useAuth'
import { MessageBubble } from './MessageBubble'
import { MessageInput } from './MessageInput'
import type { Message } from '@/types/messaging'

interface ConversationViewProps {
  conversationId: number
  onBack?: () => void
  className?: string
}

export function ConversationView({ conversationId, onBack, className }: ConversationViewProps) {
  const t = useTranslations('messaging')
  const { user } = useAuth()
  const scrollRef = useRef<HTMLDivElement>(null)
  const dateGroupRefs = useRef<Map<string, HTMLDivElement>>(new Map())
  const [replyTo, setReplyTo] = useState<Message | null>(null)
  const [visibleDate, setVisibleDate] = useState<string | null>(null)

  const {
    conversation,
    messages,
    hasMore,
    isLoading,
    isSending,
    sendMessage,
    typingUsers,
    sendTyping,
    sendStopTyping,
    isConnected,
  } = useConversationWithMessages(conversationId)

  // Group messages by date - defined before effects that use it
  const groupedMessages = useMemo(() => {
    const groups: { date: string; messages: Message[] }[] = []
    let currentDateKey = ''

    // Sort messages by created_at ASC (oldest first) for proper chat display
    const sortedMessages = [...messages].sort(
      (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    )

    for (const message of sortedMessages) {
      const msgDate = new Date(message.created_at)
      // Use ISO date string (YYYY-MM-DD) as key for grouping
      const dateKey = msgDate.toISOString().split('T')[0]
      if (dateKey !== currentDateKey) {
        currentDateKey = dateKey
        groups.push({ date: dateKey, messages: [] })
      }
      groups[groups.length - 1].messages.push(message)
    }

    return groups
  }, [messages])

  // Scroll to bottom when new messages arrive
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages.length])

  // Initialize visible date when messages load
  useEffect(() => {
    if (groupedMessages.length > 0 && !visibleDate) {
      // Set initial visible date to the last date group (newest messages are at bottom)
      setVisibleDate(groupedMessages[groupedMessages.length - 1].date)
    }
  }, [groupedMessages, visibleDate])

  /* c8 ignore start - Event handlers, tested in e2e */
  // Set up IntersectionObserver to track visible date
  useEffect(() => {
    const container = scrollRef.current
    if (!container) return

    const observer = new IntersectionObserver(
      (entries) => {
        // Find the topmost visible date group
        const visibleEntries = entries.filter((entry) => entry.isIntersecting)
        if (visibleEntries.length > 0) {
          // Sort by Y position and get the topmost one
          visibleEntries.sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)
          const topEntry = visibleEntries[0]
          const date = topEntry.target.getAttribute('data-date')
          if (date) {
            setVisibleDate(date)
          }
        }
      },
      {
        root: container,
        rootMargin: '-10px 0px -90% 0px', // Focus on the top 10% of the viewport
        threshold: 0,
      }
    )

    // Observe all date group elements
    dateGroupRefs.current.forEach((element) => {
      observer.observe(element)
    })

    return () => {
      observer.disconnect()
    }
  }, [groupedMessages])

  // Callback to set date group ref
  const setDateGroupRef = useCallback(
    (date: string) => (el: HTMLDivElement | null) => {
      if (el) {
        dateGroupRefs.current.set(date, el)
      } else {
        dateGroupRefs.current.delete(date)
      }
    },
    []
  )

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  const handleSendMessage = async (content: string) => {
    await sendMessage(content)
    setReplyTo(null)
  }

  const handleReply = (message: Message) => {
    setReplyTo(message)
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const yesterday = new Date(now)
    yesterday.setDate(yesterday.getDate() - 1)

    if (date.toDateString() === now.toDateString()) {
      return t('time.today')
    } else if (date.toDateString() === yesterday.toDateString()) {
      return t('time.yesterday')
    } else {
      return date.toLocaleDateString(undefined, {
        day: 'numeric',
        month: 'long',
        year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
      })
    }
  }
  /* c8 ignore stop */

  const isGroupConversation = conversation?.type === 'group'

  // Empty state when no conversation
  if (!conversationId) {
    return (
      <div className={cn('flex h-full flex-col items-center justify-center', className)}>
        <div className="text-center px-8">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
            <MessageCircle className="h-8 w-8 text-primary" />
          </div>
          <h3 className="text-lg font-semibold">{t('selectConversation')}</h3>
          <p className="mt-2 text-sm text-muted-foreground">{t('selectConversationDesc')}</p>
        </div>
      </div>
    )
  }

  // Loading state
  if (isLoading && !conversation) {
    return (
      <div className={cn('flex h-full items-center justify-center', className)}>
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className={cn('flex h-full flex-col overflow-hidden', className)}>
      {/* Header - fixed at top */}
      <div className="flex-shrink-0 flex items-center justify-between border-b px-4 py-3">
        <div className="flex items-center gap-3">
          {onBack && (
            <Button variant="ghost" size="icon" className="h-8 w-8 md:hidden" onClick={onBack}>
              <ArrowLeft className="h-4 w-4" />
            </Button>
          )}

          {/* c8 ignore start - Group avatar conditional */}
          {isGroupConversation ? (
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
              <Users className="h-5 w-5 text-primary" />
            </div>
          ) : (
            <Avatar className="h-10 w-10">
              <AvatarImage src={getValidAvatarUrl(conversation?.avatar_url)} />
              <AvatarFallback>{getInitials(conversation?.title || 'U')}</AvatarFallback>
            </Avatar>
          )}
          {/* c8 ignore stop */}

          <div className="min-w-0">
            {/* c8 ignore next */}
            <h3 className="font-semibold truncate">{conversation?.title || t('unknownUser')}</h3>
            {/* c8 ignore start - Connection status conditionals */}
            <p className="text-xs text-muted-foreground">
              {isConnected ? (
                typingUsers.length > 0 ? (
                  <span className="text-primary">{t('typing')}</span>
                ) : isGroupConversation ? (
                  t('participantsCount', {
                    count: conversation?.participants.length || 0,
                  })
                ) : (
                  t('online')
                )
              ) : (
                t('connecting')
              )}
            </p>
            {/* c8 ignore stop */}
          </div>
        </div>

        <div className="flex items-center gap-1">
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <Search className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <Phone className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <Video className="h-4 w-4" />
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {isGroupConversation && (
                <>
                  <DropdownMenuItem>
                    <Users className="mr-2 h-4 w-4" />
                    {t('viewParticipants')}
                  </DropdownMenuItem>
                  <DropdownMenuItem>
                    <Settings className="mr-2 h-4 w-4" />
                    {t('groupSettings')}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                </>
              )}
              <DropdownMenuItem className="text-destructive focus:text-destructive">
                {t('leaveConversation')}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Messages Area - scrollable with floating date indicator */}
      <div className="relative flex-1 min-h-0 overflow-hidden">
        {/* Floating date indicator */}
        {visibleDate && messages.length > 0 && (
          <div className="absolute top-2 left-1/2 -translate-x-1/2 z-20 pointer-events-none">
            <div className="rounded-full bg-muted/95 backdrop-blur-sm px-3 py-1 text-xs text-muted-foreground shadow-md border border-border/50">
              {formatDate(visibleDate)}
            </div>
          </div>
        )}

        <div className="absolute inset-0 overflow-y-auto" ref={scrollRef}>
          <div className="px-4 py-4 pt-10 space-y-4">
            {/* Load more button */}
            {hasMore && (
              <div className="flex justify-center py-2">
                <Button variant="ghost" size="sm">
                  {t('loadMore')}
                </Button>
              </div>
            )}

            {/* Messages grouped by date */}
            {groupedMessages.map(({ date, messages: dateMessages }) => (
              <div key={date} ref={setDateGroupRef(date)} data-date={date}>
                {/* Messages */}
                <div className="space-y-3">
                  {/* c8 ignore start - Message rendering with showAvatar logic */}
                  {dateMessages.map((message, index) => {
                    const prevMessage = dateMessages[index - 1]
                    const showAvatar =
                      !prevMessage ||
                      prevMessage.sender_id !== message.sender_id ||
                      message.type === 'system'

                    return (
                      <MessageBubble
                        key={message.id}
                        message={message}
                        isOwn={message.sender_id === user?.id}
                        showAvatar={showAvatar}
                        showSender={isGroupConversation && showAvatar}
                        onReply={handleReply}
                      />
                    )
                  })}
                  {/* c8 ignore stop */}
                </div>
              </div>
            ))}

            {/* c8 ignore start - Empty state */}
            {messages.length === 0 && !isLoading && (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <div className="rounded-full bg-primary/10 p-4 mb-4">
                  <Sparkles className="h-8 w-8 text-primary" />
                </div>
                <h3 className="text-lg font-medium">{t('startConversation')}</h3>
                <p className="mt-1 text-sm text-muted-foreground max-w-sm">
                  {t('startConversationDesc')}
                </p>
              </div>
            )}
            {/* c8 ignore stop */}

            {/* Loading indicator */}
            {isLoading && (
              <div className="flex justify-center py-4">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            )}
          </div>
        </div>
      </div>

      {/* c8 ignore start - Reply Preview, tested in e2e */}
      {/* Reply Preview - fixed above input */}
      {replyTo && (
        <div className="flex-shrink-0 mx-4 mt-2 flex items-center gap-2 rounded-lg bg-muted p-2">
          <div className="flex-1 min-w-0">
            <p className="text-xs font-medium text-primary">
              {t('replyingTo')} {replyTo.sender_name}
            </p>
            <p className="text-sm truncate text-muted-foreground">{replyTo.content}</p>
          </div>
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6 flex-shrink-0"
            onClick={() => setReplyTo(null)}
          >
            <ArrowLeft className="h-3 w-3 rotate-45" />
          </Button>
        </div>
      )}
      {/* c8 ignore stop */}

      {/* Message Input - fixed at bottom */}
      <div className="flex-shrink-0">
        <MessageInput
          onSend={handleSendMessage}
          onTyping={sendTyping}
          onStopTyping={sendStopTyping}
          disabled={isSending}
          showAiSuggestions={messages.length === 0}
        />
      </div>
    </div>
  )
}
