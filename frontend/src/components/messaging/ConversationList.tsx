'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import {
  Search,
  Plus,
  Users,
  MessageCircle,
  Loader2,
  UserPlus,
  ArrowLeft,
  Check,
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { cn, getValidAvatarUrl } from '@/lib/utils'
import {
  useConversations,
  useCreateDirectConversation,
  useCreateGroupConversation,
} from '@/hooks/useMessaging'
import { usersApi, type User } from '@/lib/api/users'
import { useAuth } from '@/hooks/useAuth'
import type { Conversation, ConversationType } from '@/types/messaging'

interface ConversationListProps {
  selectedId?: number
  onSelect?: (conversation: Conversation) => void
  className?: string
}

export function ConversationList({ selectedId, onSelect, className }: ConversationListProps) {
  const t = useTranslations('messaging')
  const router = useRouter()
  const [search, setSearch] = useState('')
  const [filterType, setFilterType] = useState<ConversationType | undefined>()

  const { conversations, isLoading, error } = useConversations({
    search: search || undefined,
    type: filterType,
    limit: 50,
  })

  const handleSelect = (conversation: Conversation) => {
    if (onSelect) {
      onSelect(conversation)
    } else {
      router.push(`/messages/${conversation.id}`)
    }
  }

  const formatTime = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const days = Math.floor(diff / (1000 * 60 * 60 * 24))

    if (days === 0) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    } else if (days === 1) {
      return t('time.yesterday')
    } else if (days < 7) {
      return date.toLocaleDateString([], { weekday: 'short' })
    } else {
      return date.toLocaleDateString([], { day: 'numeric', month: 'short' })
    }
  }

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  const getLastMessagePreview = (conversation: Conversation) => {
    if (!conversation.last_message) return t('noMessages')

    const msg = conversation.last_message
    if (msg.is_deleted) return t('messageDeleted')

    if (msg.type === 'image') return '📷 ' + t('image')
    if (msg.type === 'file') return '📎 ' + t('file')
    if (msg.type === 'system') return '⚙️ ' + msg.content

    return msg.content
  }

  return (
    <div className={cn('flex h-full flex-col', className)}>
      {/* Header */}
      <div className="border-b p-4">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">{t('conversations')}</h2>
          <NewConversationDialog />
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder={t('searchConversations')}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>

        {/* Filters */}
        <div className="mt-3 flex gap-2">
          <Button
            variant={filterType === undefined ? 'secondary' : 'ghost'}
            size="sm"
            onClick={() => setFilterType(undefined)}
            className="h-7 text-xs"
          >
            {t('all')}
          </Button>
          <Button
            variant={filterType === 'direct' ? 'secondary' : 'ghost'}
            size="sm"
            onClick={() => setFilterType('direct')}
            className="h-7 text-xs gap-1"
          >
            <MessageCircle className="h-3 w-3" />
            {t('direct')}
          </Button>
          <Button
            variant={filterType === 'group' ? 'secondary' : 'ghost'}
            size="sm"
            onClick={() => setFilterType('group')}
            className="h-7 text-xs gap-1"
          >
            <Users className="h-3 w-3" />
            {t('groups')}
          </Button>
        </div>
      </div>

      {/* Conversation List */}
      <ScrollArea className="flex-1">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
            <p className="text-sm text-destructive">{t('loadError')}</p>
          </div>
        ) : conversations.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
            <div className="rounded-full bg-muted p-3 mb-3">
              <MessageCircle className="h-6 w-6 text-muted-foreground" />
            </div>
            <p className="text-sm text-muted-foreground">{t('noConversations')}</p>
            <p className="mt-1 text-xs text-muted-foreground">{t('startNewConversation')}</p>
          </div>
        ) : (
          <div className="divide-y">
            {conversations.map((conversation) => (
              <button
                key={conversation.id}
                onClick={() => handleSelect(conversation)}
                className={cn(
                  'w-full flex items-start gap-3 p-4 text-left transition-colors hover:bg-accent/50',
                  selectedId === conversation.id && 'bg-accent'
                )}
              >
                {/* Avatar */}
                <div className="relative flex-shrink-0">
                  {conversation.type === 'group' ? (
                    <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                      <Users className="h-5 w-5 text-primary" />
                    </div>
                  ) : (
                    <Avatar className="h-12 w-12">
                      <AvatarImage src={getValidAvatarUrl(conversation.avatar_url)} />
                      <AvatarFallback>{getInitials(conversation.title || 'U')}</AvatarFallback>
                    </Avatar>
                  )}
                </div>

                {/* Content */}
                <div className="min-w-0 flex-1">
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-medium truncate">
                      {conversation.title || t('unknownUser')}
                    </span>
                    {conversation.last_message && (
                      <span className="flex-shrink-0 text-xs text-muted-foreground">
                        {formatTime(conversation.last_message.created_at)}
                      </span>
                    )}
                  </div>
                  <div className="flex items-center justify-between gap-2 mt-0.5">
                    <p className="text-sm text-muted-foreground truncate">
                      {conversation.last_message?.sender_name && (
                        <span className="font-medium">
                          {conversation.last_message.sender_name}:{' '}
                        </span>
                      )}
                      {getLastMessagePreview(conversation)}
                    </p>
                    {conversation.unread_count > 0 && (
                      <Badge variant="default" className="h-5 min-w-5 flex-shrink-0 px-1.5 text-xs">
                        {conversation.unread_count > 99 ? '99+' : conversation.unread_count}
                      </Badge>
                    )}
                  </div>
                </div>
              </button>
            ))}
          </div>
        )}
      </ScrollArea>
    </div>
  )
}

// New Conversation Dialog
interface NewConversationDialogProps {
  onConversationCreated?: (conversation: Conversation) => void
}

function NewConversationDialog({ onConversationCreated }: NewConversationDialogProps) {
  const t = useTranslations('messaging')
  const tCommon = useTranslations('common')
  const router = useRouter()
  const { user: currentUser } = useAuth()
  const [open, setOpen] = useState(false)
  const [step, setStep] = useState<'choose' | 'direct' | 'group'>('choose')
  const [users, setUsers] = useState<User[]>([])
  const [loadingUsers, setLoadingUsers] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedUsers, setSelectedUsers] = useState<User[]>([])
  const [groupName, setGroupName] = useState('')

  const { mutateAsync: createDirect, isPending: isCreatingDirect } = useCreateDirectConversation()
  const { mutateAsync: createGroup, isPending: isCreatingGroup } = useCreateGroupConversation()

  const isCreating = isCreatingDirect || isCreatingGroup

  // Load users when dialog opens
  useEffect(() => {
    if (open && (step === 'direct' || step === 'group')) {
      setLoadingUsers(true)
      usersApi
        .getAll()
        .then((data) => {
          // Filter out current user
          setUsers(data.filter((u) => u.id !== currentUser?.id))
        })
        .catch(console.error)
        .finally(() => setLoadingUsers(false))
    }
  }, [open, step, currentUser?.id])

  // Reset state when dialog closes
  useEffect(() => {
    if (!open) {
      setStep('choose')
      setSearchQuery('')
      setSelectedUsers([])
      setGroupName('')
    }
  }, [open])

  const filteredUsers = users.filter(
    (user) =>
      user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.email.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  const toggleUserSelection = (user: User) => {
    setSelectedUsers((prev) => {
      const isSelected = prev.some((u) => u.id === user.id)
      if (isSelected) {
        return prev.filter((u) => u.id !== user.id)
      }
      return [...prev, user]
    })
  }

  const handleCreateDirect = async (user: User) => {
    try {
      const conversation = await createDirect({ recipient_id: user.id })
      setOpen(false)
      if (conversation) {
        if (onConversationCreated) {
          onConversationCreated(conversation)
        } else {
          router.push(`/messages/${conversation.id}`)
        }
      }
    } catch (error) {
      console.error('Failed to create conversation:', error)
    }
  }

  const handleCreateGroup = async () => {
    if (selectedUsers.length < 1 || !groupName.trim()) return

    try {
      const conversation = await createGroup({
        title: groupName.trim(),
        participant_ids: selectedUsers.map((u) => u.id),
      })
      setOpen(false)
      if (conversation) {
        if (onConversationCreated) {
          onConversationCreated(conversation)
        } else {
          router.push(`/messages/${conversation.id}`)
        }
      }
    } catch (error) {
      console.error('Failed to create group:', error)
    }
  }

  const renderUserList = (onUserClick: (user: User) => void, multiSelect = false) => (
    <>
      <div className="relative mb-4">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder={tCommon('search')}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-9"
        />
      </div>
      <ScrollArea className="h-[300px] -mx-2">
        {loadingUsers ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : filteredUsers.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground text-sm">
            {tCommon('noResults')}
          </div>
        ) : (
          <div className="space-y-1 px-2">
            {filteredUsers.map((user) => {
              const isSelected = multiSelect && selectedUsers.some((u) => u.id === user.id)
              return (
                <button
                  key={user.id}
                  onClick={() => onUserClick(user)}
                  disabled={isCreating}
                  className={cn(
                    'w-full flex items-center gap-3 p-2 rounded-lg text-left transition-colors hover:bg-accent',
                    isSelected && 'bg-accent'
                  )}
                >
                  <Avatar className="h-10 w-10">
                    <AvatarImage src={user.avatar || undefined} />
                    <AvatarFallback>{getInitials(user.name)}</AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">{user.name}</p>
                    <p className="text-sm text-muted-foreground truncate">{user.email}</p>
                  </div>
                  {multiSelect && isSelected && (
                    <div className="flex h-5 w-5 items-center justify-center rounded-full bg-primary">
                      <Check className="h-3 w-3 text-primary-foreground" />
                    </div>
                  )}
                </button>
              )
            })}
          </div>
        )}
      </ScrollArea>
    </>
  )

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="icon" variant="ghost" className="h-8 w-8">
          <Plus className="h-4 w-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {step !== 'choose' && (
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 -ml-1"
                onClick={() => setStep('choose')}
              >
                <ArrowLeft className="h-4 w-4" />
              </Button>
            )}
            {step === 'choose' && t('newConversation')}
            {step === 'direct' && t('directMessage')}
            {step === 'group' && t('createGroup')}
          </DialogTitle>
        </DialogHeader>

        {step === 'choose' && (
          <div className="grid gap-4 py-4">
            <Button
              variant="outline"
              className="justify-start gap-3 h-auto py-4"
              onClick={() => setStep('direct')}
            >
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
                <UserPlus className="h-5 w-5 text-primary" />
              </div>
              <div className="text-left">
                <div className="font-medium">{t('directMessage')}</div>
                <div className="text-sm text-muted-foreground">{t('directMessageDesc')}</div>
              </div>
            </Button>
            <Button
              variant="outline"
              className="justify-start gap-3 h-auto py-4"
              onClick={() => setStep('group')}
            >
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
                <Users className="h-5 w-5 text-primary" />
              </div>
              <div className="text-left">
                <div className="font-medium">{t('createGroup')}</div>
                <div className="text-sm text-muted-foreground">{t('createGroupDesc')}</div>
              </div>
            </Button>
          </div>
        )}

        {step === 'direct' && <div className="py-4">{renderUserList(handleCreateDirect)}</div>}

        {step === 'group' && (
          <div className="py-4 space-y-4">
            <Input
              placeholder={t('groupName')}
              value={groupName}
              onChange={(e) => setGroupName(e.target.value)}
            />
            {selectedUsers.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {selectedUsers.map((user) => (
                  <Badge
                    key={user.id}
                    variant="secondary"
                    className="cursor-pointer"
                    onClick={() => toggleUserSelection(user)}
                  >
                    {user.name} ×
                  </Badge>
                ))}
              </div>
            )}
            {renderUserList(toggleUserSelection, true)}
            <Button
              className="w-full"
              onClick={handleCreateGroup}
              disabled={selectedUsers.length < 1 || !groupName.trim() || isCreating}
            >
              {isCreating ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
              {tCommon('create')}
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
