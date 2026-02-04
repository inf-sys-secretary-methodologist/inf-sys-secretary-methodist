'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { MessageCircle, Sparkles, Users, Send } from 'lucide-react'
import { useAuthCheck } from '@/hooks/useAuth'
import { getAvailableNavEntries } from '@/config/navigation'
import { AppHeader } from '@/components/layout/AppHeader'
import { SkipToContent } from '@/components/ui/skip-to-content'
import { ConversationList } from '@/components/messaging/ConversationList'
import { ConversationView } from '@/components/messaging/ConversationView'
import { cn } from '@/lib/utils'
import type { Conversation } from '@/types/messaging'

export default function MessagesPage() {
  const { user, isLoading } = useAuthCheck()
  const [selectedConversation, setSelectedConversation] = useState<Conversation | null>(null)
  const t = useTranslations('common')

  const handleSelectConversation = (conversation: Conversation) => {
    setSelectedConversation(conversation)
  }

  const handleBack = () => {
    setSelectedConversation(null)
  }

  // Get navigation entries filtered by user role
  const navEntries = getAvailableNavEntries(user?.role)

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="text-muted-foreground">{t('loading')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex flex-col">
      <SkipToContent />
      <AppHeader entries={navEntries} />
      <main id="main-content" tabIndex={-1} className="flex-1 flex focus:outline-none">
        {/* Conversation List - Hidden on mobile when conversation is selected */}
        <div
          className={cn(
            'w-full md:w-80 lg:w-96 border-r flex-shrink-0',
            selectedConversation ? 'hidden md:flex md:flex-col' : 'flex flex-col'
          )}
        >
          <ConversationList
            selectedId={selectedConversation?.id}
            onSelect={handleSelectConversation}
          />
        </div>

        {/* Conversation View - Hidden on mobile when no conversation selected */}
        <div
          className={cn(
            'flex-1 min-w-0',
            !selectedConversation ? 'hidden md:flex md:flex-col' : 'flex flex-col'
          )}
        >
          {selectedConversation ? (
            <ConversationView conversationId={selectedConversation.id} onBack={handleBack} />
          ) : (
            <EmptyState />
          )}
        </div>
      </main>
    </div>
  )
}

function EmptyState() {
  const t = useTranslations('messaging')

  return (
    <div className="flex h-full flex-col items-center justify-center px-8">
      <div className="relative mb-8">
        {/* Main icon */}
        <div className="flex h-20 w-20 items-center justify-center rounded-2xl bg-gradient-to-br from-primary/20 to-primary/5 shadow-lg">
          <MessageCircle className="h-10 w-10 text-primary" />
        </div>
        {/* Decorative elements */}
        <div className="absolute -top-2 -right-2 flex h-8 w-8 items-center justify-center rounded-full bg-blue-500/10">
          <Sparkles className="h-4 w-4 text-blue-500" />
        </div>
        <div className="absolute -bottom-1 -left-1 flex h-6 w-6 items-center justify-center rounded-full bg-green-500/10">
          <Users className="h-3 w-3 text-green-500" />
        </div>
        <div className="absolute top-1/2 -right-4 flex h-5 w-5 items-center justify-center rounded-full bg-amber-500/10">
          <Send className="h-2.5 w-2.5 text-amber-500" />
        </div>
      </div>

      <div className="text-center max-w-md">
        <h2 className="text-2xl font-bold tracking-tight">{t('welcome')}</h2>
        <p className="mt-3 text-muted-foreground leading-relaxed">{t('welcomeDesc')}</p>
      </div>

      <div className="mt-8 flex flex-col items-center gap-4">
        <div className="flex items-center gap-3 text-sm text-muted-foreground">
          <div className="flex items-center gap-1.5">
            <div className="h-2 w-2 rounded-full bg-green-500" />
            {t('realTimeSync')}
          </div>
          <span className="text-muted-foreground/50">•</span>
          <div className="flex items-center gap-1.5">
            <div className="h-2 w-2 rounded-full bg-blue-500" />
            {t('fileSharing')}
          </div>
          <span className="text-muted-foreground/50">•</span>
          <div className="flex items-center gap-1.5">
            <div className="h-2 w-2 rounded-full bg-purple-500" />
            {t('groupChats')}
          </div>
        </div>
      </div>
    </div>
  )
}
