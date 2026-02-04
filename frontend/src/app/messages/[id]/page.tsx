'use client'

import { use } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { useAuthCheck } from '@/hooks/useAuth'
import { getAvailableNavEntries } from '@/config/navigation'
import { AppHeader } from '@/components/layout/AppHeader'
import { SkipToContent } from '@/components/ui/skip-to-content'
import { ConversationList } from '@/components/messaging/ConversationList'
import { ConversationView } from '@/components/messaging/ConversationView'
import type { Conversation } from '@/types/messaging'

interface PageProps {
  params: Promise<{ id: string }>
}

export default function ConversationPage({ params }: PageProps) {
  const { id } = use(params)
  const router = useRouter()
  const { user, isLoading } = useAuthCheck()
  const t = useTranslations('common')
  const conversationId = parseInt(id, 10)

  const handleSelectConversation = (conversation: Conversation) => {
    router.push(`/messages/${conversation.id}`)
  }

  const handleBack = () => {
    router.push('/messages')
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

  if (isNaN(conversationId)) {
    router.push('/messages')
    return null
  }

  return (
    <div className="h-screen flex flex-col overflow-hidden">
      <SkipToContent />
      <AppHeader entries={navEntries} />
      <main
        id="main-content"
        tabIndex={-1}
        className="flex-1 flex overflow-hidden focus:outline-none"
      >
        {/* Conversation List - Hidden on mobile */}
        <div className="hidden md:flex md:w-80 lg:w-96 border-r flex-shrink-0 flex-col overflow-hidden">
          <ConversationList selectedId={conversationId} onSelect={handleSelectConversation} />
        </div>

        {/* Conversation View */}
        <div className="flex-1 min-w-0 flex flex-col overflow-hidden">
          <ConversationView conversationId={conversationId} onBack={handleBack} />
        </div>
      </main>
    </div>
  )
}
