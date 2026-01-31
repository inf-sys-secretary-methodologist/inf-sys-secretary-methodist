'use client'

import { useState } from 'react'
import {
  Bell,
  BellOff,
  Check,
  Loader2,
  Smartphone,
  Trash2,
  TestTube,
  AlertTriangle,
} from 'lucide-react'
import { toast } from 'sonner'
import { useTranslations } from 'next-intl'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
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
import { usePushNotifications } from '@/hooks/usePushNotifications'
import { cn } from '@/lib/utils'

export function PushNotificationSettings() {
  const t = useTranslations('push')
  const [isSendingTest, setIsSendingTest] = useState(false)

  const {
    isSupported,
    permission,
    isEnabled,
    isLocallySubscribed,
    subscriptions,
    totalDevices,
    isLoading,
    isSubscribing,
    isUnsubscribing,
    error,
    subscribe,
    unsubscribe,
    removeSubscription,
    testNotification,
  } = usePushNotifications()

  /* c8 ignore start - Action handlers with browser APIs */
  const handleEnablePush = async () => {
    try {
      const result = await subscribe()
      if (result) {
        toast.success(t('enabledSuccess'))
      }
    } catch (err) {
      if (err instanceof Error && err.message.includes('denied')) {
        toast.error(t('permissionDenied'))
      } else {
        toast.error(t('enableError'))
      }
    }
  }

  const handleDisablePush = async () => {
    try {
      await unsubscribe()
      toast.success(t('disabledSuccess'))
    } catch {
      toast.error(t('disableError'))
    }
  }

  const handleRemoveDevice = async (subscriptionId: number) => {
    try {
      await removeSubscription(subscriptionId)
      toast.success(t('deviceRemoved'))
    } catch {
      toast.error(t('removeError'))
    }
  }

  const handleTestNotification = async () => {
    setIsSendingTest(true)
    try {
      await testNotification()
      toast.success(t('testSent'))
    } catch {
      toast.error(t('testError'))
    } finally {
      setIsSendingTest(false)
    }
  }
  /* c8 ignore stop */

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString(undefined, {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
    })
  }

  // Not supported
  if (!isSupported) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <BellOff className="h-5 w-5 flex-shrink-0" />
            {t('title')}
          </CardTitle>
          <CardDescription>{t('notSupported')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2 text-muted-foreground text-sm">
            <AlertTriangle className="h-4 w-4" />
            <span>{t('notSupportedDescription')}</span>
          </div>
        </CardContent>
      </Card>
    )
  }

  // Permission denied
  if (permission === 'denied') {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <BellOff className="h-5 w-5 flex-shrink-0" />
            {t('title')}
          </CardTitle>
          <CardDescription>{t('permissionBlocked')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <AlertTriangle className="h-5 w-5 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5" />
              <div className="space-y-1">
                <p className="text-sm font-medium text-amber-800 dark:text-amber-200">
                  {t('permissionBlockedTitle')}
                </p>
                <p className="text-sm text-amber-700 dark:text-amber-300">
                  {t('permissionBlockedDescription')}
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  /* c8 ignore start - Loading state */
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Bell className="h-5 w-5 flex-shrink-0" />
            {t('title')}
          </CardTitle>
        </CardHeader>
        <CardContent className="flex justify-center py-8">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    )
  }
  /* c8 ignore stop */

  // Enabled state - show devices and management options
  if (isLocallySubscribed || isEnabled) {
    return (
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Bell className="h-5 w-5 flex-shrink-0" />
                {t('title')}
              </CardTitle>
              <CardDescription>{t('enabledDescription')}</CardDescription>
            </div>
            <Badge variant="default" className="bg-green-600 hover:bg-green-700">
              <Check className="h-3 w-3 mr-1" />
              {t('enabled')}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Toggle switch */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Smartphone className="h-5 w-5 text-muted-foreground" />
              <div>
                <Label>{t('browserNotifications')}</Label>
                <p className="text-sm text-muted-foreground">{t('browserNotificationsDesc')}</p>
              </div>
            </div>
            <Switch
              checked={true}
              onCheckedChange={(checked) => {
                if (!checked) {
                  handleDisablePush()
                }
              }}
              disabled={isUnsubscribing}
            />
          </div>

          {/* Subscribed devices */}
          {subscriptions.length > 0 && (
            <div className="space-y-3">
              <Label className="text-sm font-medium">{t('devices', { count: totalDevices })}</Label>
              <div className="space-y-2">
                {subscriptions.map((sub) => (
                  <div
                    key={sub.id}
                    className={cn(
                      'flex items-center justify-between rounded-lg border p-3',
                      sub.is_active ? 'bg-background' : 'bg-muted/50 opacity-60'
                    )}
                  >
                    <div className="flex items-center gap-3">
                      <Smartphone className="h-4 w-4 text-muted-foreground" />
                      <div>
                        <p className="text-sm font-medium">
                          {sub.device_name || t('unknownDevice')}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {t('addedOn', { date: formatDate(sub.created_at) })}
                        </p>
                      </div>
                    </div>
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8">
                          <Trash2 className="h-4 w-4 text-muted-foreground hover:text-destructive" />
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>{t('removeDeviceTitle')}</AlertDialogTitle>
                          <AlertDialogDescription>
                            {t('removeDeviceDescription')}
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>{t('cancel')}</AlertDialogCancel>
                          <AlertDialogAction
                            onClick={() => handleRemoveDevice(sub.id)}
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                          >
                            {t('remove')}
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Test notification button */}
          <Button
            variant="outline"
            onClick={handleTestNotification}
            disabled={isSendingTest}
            className="w-full"
          >
            {isSendingTest ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                {t('sendingTest')}
              </>
            ) : (
              <>
                <TestTube className="h-4 w-4 mr-2" />
                {t('sendTest')}
              </>
            )}
          </Button>

          {/* Disable all button */}
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive" className="w-full">
                <BellOff className="h-4 w-4 mr-2" />
                {t('disableAll')}
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>{t('disableTitle')}</AlertDialogTitle>
                <AlertDialogDescription>{t('disableDescription')}</AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>{t('cancel')}</AlertDialogCancel>
                <AlertDialogAction
                  onClick={handleDisablePush}
                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                >
                  {t('confirmDisable')}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>

          {/* Error display */}
          {error && (
            <div className="text-sm text-destructive flex items-center gap-2">
              <AlertTriangle className="h-4 w-4" />
              {error.message}
            </div>
          )}
        </CardContent>
      </Card>
    )
  }

  // Not enabled - show enable button
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Bell className="h-5 w-5 flex-shrink-0" />
          {t('title')}
        </CardTitle>
        <CardDescription>{t('description')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm text-muted-foreground">{t('enableInstructions')}</p>
        <Button onClick={handleEnablePush} disabled={isSubscribing} className="w-full">
          {isSubscribing ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              {t('enabling')}
            </>
          ) : (
            <>
              <Bell className="h-4 w-4 mr-2" />
              {t('enable')}
            </>
          )}
        </Button>
      </CardContent>
    </Card>
  )
}
