'use client'

import { useState, useEffect } from 'react'
import { Send, Check, Copy, ExternalLink, Loader2, XCircle, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'
import { useTranslations } from 'next-intl'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import {
  useTelegramStatus,
  useGenerateVerificationCode,
  useDisconnectTelegram,
} from '@/hooks/useTelegram'

export function TelegramLinkCard() {
  const t = useTranslations('telegram')
  const { data: status, isLoading: statusLoading, mutate: refreshStatus } = useTelegramStatus()
  const generateCode = useGenerateVerificationCode()
  const disconnectTelegram = useDisconnectTelegram()

  const [copied, setCopied] = useState(false)
  const [expiresIn, setExpiresIn] = useState<number | null>(null)

  // Update expiration countdown
  useEffect(() => {
    if (!generateCode.data?.expires_at) {
      setExpiresIn(null)
      return
    }

    const updateExpiry = () => {
      const expiresAt = new Date(generateCode.data!.expires_at).getTime()
      const now = Date.now()
      const remaining = Math.max(0, Math.floor((expiresAt - now) / 1000))
      setExpiresIn(remaining)

      if (remaining <= 0) {
        generateCode.reset()
      }
    }

    updateExpiry()
    const interval = setInterval(updateExpiry, 1000)
    return () => clearInterval(interval)
  }, [generateCode.data?.expires_at, generateCode])

  const handleGenerateCode = async () => {
    try {
      await generateCode.mutateAsync()
    } catch {
      toast.error(t('generateCodeError'))
    }
  }

  const handleCopyCode = async () => {
    if (!generateCode.data?.code) return

    try {
      await navigator.clipboard.writeText(generateCode.data.code)
      setCopied(true)
      toast.success(t('codeCopied'))
      setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error(t('copyError'))
    }
  }

  const handleOpenBot = () => {
    if (!generateCode.data?.bot_link) return
    window.open(generateCode.data.bot_link, '_blank')
  }

  const handleDisconnect = async () => {
    try {
      await disconnectTelegram.mutateAsync()
      toast.success(t('disconnected'))
      refreshStatus()
    } catch {
      toast.error(t('disconnectError'))
    }
  }

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  if (statusLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Send className="h-5 w-5 flex-shrink-0" />
            Telegram
          </CardTitle>
        </CardHeader>
        <CardContent className="flex justify-center py-8">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    )
  }

  // Connected state
  if (status?.connected) {
    return (
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Send className="h-5 w-5 flex-shrink-0" />
                Telegram
              </CardTitle>
              <CardDescription>{t('accountLinked')}</CardDescription>
            </div>
            <Badge variant="default" className="bg-green-600 hover:bg-green-700">
              <Check className="h-3 w-3 mr-1" />
              {t('connected')}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="bg-muted/50 rounded-lg p-4 space-y-2">
            {status.first_name && (
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">{t('name')}</span>
                <span className="font-medium">{status.first_name}</span>
              </div>
            )}
            {status.username && (
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Username</span>
                <span className="font-medium">@{status.username}</span>
              </div>
            )}
            {status.connected_at && (
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">{t('connectedAt')}</span>
                <span className="font-medium">
                  {new Date(status.connected_at).toLocaleDateString(undefined, {
                    day: 'numeric',
                    month: 'long',
                    year: 'numeric',
                  })}
                </span>
              </div>
            )}
          </div>

          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive" className="w-full">
                <XCircle className="h-4 w-4 mr-2" />
                {t('disconnect')}
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>{t('disconnectTitle')}</AlertDialogTitle>
                <AlertDialogDescription>{t('disconnectDescription')}</AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>{t('cancel')}</AlertDialogCancel>
                <AlertDialogAction
                  onClick={handleDisconnect}
                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                >
                  {t('confirmDisconnect')}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </CardContent>
      </Card>
    )
  }

  // Not connected - show verification flow
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Send className="h-5 w-5 flex-shrink-0" />
          Telegram
        </CardTitle>
        <CardDescription>{t('linkDescription')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {!generateCode.data ? (
          // Step 1: Generate code button
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">{t('linkInstructions')}</p>
            <Button
              onClick={handleGenerateCode}
              disabled={generateCode.isPending}
              className="w-full"
            >
              {generateCode.isPending ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  {t('generating')}
                </>
              ) : (
                <>
                  <Send className="h-4 w-4 mr-2" />
                  {t('getCode')}
                </>
              )}
            </Button>
          </div>
        ) : (
          // Step 2: Show code and instructions
          <div className="space-y-4">
            <div className="bg-muted rounded-lg p-4 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">{t('yourCode')}</span>
                {expiresIn !== null && expiresIn > 0 && (
                  <Badge variant="outline" className="text-xs">
                    {t('expiresIn', { time: formatTime(expiresIn) })}
                  </Badge>
                )}
              </div>
              <div className="flex items-center gap-2">
                <code className="flex-1 text-center text-2xl font-mono font-bold tracking-wider bg-background rounded px-4 py-2">
                  {generateCode.data.code}
                </code>
                <Button
                  variant="outline"
                  size="icon"
                  onClick={handleCopyCode}
                  aria-label={t('copyCode')}
                >
                  {copied ? (
                    <Check className="h-4 w-4 text-green-600" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>

            <div className="space-y-2 text-sm">
              <p className="font-medium">{t('howToLink')}</p>
              <ol className="list-decimal list-inside space-y-1 text-muted-foreground">
                <li>{t('step1')}</li>
                <li>{t('step2')}</li>
                <li>{t('step3')}</li>
              </ol>
            </div>

            <div className="flex gap-2">
              <Button onClick={handleOpenBot} className="flex-1">
                <ExternalLink className="h-4 w-4 mr-2" />
                {t('openBot')}
              </Button>
              <Button
                variant="outline"
                size="icon"
                onClick={handleGenerateCode}
                disabled={generateCode.isPending}
                aria-label={t('getNewCode')}
              >
                <RefreshCw className={`h-4 w-4 ${generateCode.isPending ? 'animate-spin' : ''}`} />
              </Button>
            </div>

            <Button
              variant="ghost"
              onClick={() => {
                generateCode.reset()
                refreshStatus()
              }}
              className="w-full text-muted-foreground"
            >
              {t('cancel')}
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
