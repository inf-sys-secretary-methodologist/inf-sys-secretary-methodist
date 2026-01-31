'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { Bell, Mail, Smartphone, MessageSquare, Clock, RotateCcw, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { AppLayout } from '@/components/layout'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
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
  useNotificationPreferences,
  useToggleChannel,
  useUpdateQuietHours,
  useResetPreferences,
  useTimezones,
} from '@/hooks/useNotifications'
import { TelegramLinkCard } from '@/components/telegram/TelegramLinkCard'
import { PushNotificationSettings } from '@/components/notifications'

const _CHANNEL_KEYS = ['in_app', 'email', 'push', 'slack'] as const

export default function NotificationSettingsPage() {
  const t = useTranslations('settings.notifications')
  const tSettings = useTranslations('settings')
  const tCommon = useTranslations('common')
  const { data: preferences, isLoading } = useNotificationPreferences()
  const { data: timezonesData } = useTimezones()
  const toggleChannel = useToggleChannel()
  const updateQuietHours = useUpdateQuietHours()
  const resetPreferences = useResetPreferences()

  const [quietHoursStart, setQuietHoursStart] = useState('')
  const [quietHoursEnd, setQuietHoursEnd] = useState('')
  const [timezone, setTimezone] = useState('')

  // Initialize form values when preferences load
  const initFormValues = () => {
    if (preferences) {
      setQuietHoursStart(preferences.quiet_hours_start || '22:00')
      setQuietHoursEnd(preferences.quiet_hours_end || '08:00')
      setTimezone(preferences.timezone || 'Europe/Moscow')
    }
  }

  // Initialize on first render with data
  if (preferences && !quietHoursStart && !quietHoursEnd) {
    initFormValues()
  }

  const handleToggleChannel = async (channel: string, enabled: boolean) => {
    try {
      await toggleChannel.mutateAsync({ channel, enabled })
      const channelKey = channel === 'in_app' ? 'inApp' : channel
      const channelName = t(`channels.${channelKey}`)
      const status = enabled ? t('channels.enabled') : t('channels.disabled')
      toast.success(`${channelName} ${status}`)
    } catch {
      toast.error(t('updateError'))
    }
  }

  const handleQuietHoursToggle = async (enabled: boolean) => {
    try {
      await updateQuietHours.mutateAsync({
        enabled,
        start_time: quietHoursStart,
        end_time: quietHoursEnd,
        timezone,
      })
      toast.success(enabled ? t('quietHours.enabled') : t('quietHours.disabled'))
    } catch {
      toast.error(t('updateError'))
    }
  }

  const handleSaveQuietHours = async () => {
    try {
      await updateQuietHours.mutateAsync({
        enabled: preferences?.quiet_hours_enabled ?? false,
        start_time: quietHoursStart,
        end_time: quietHoursEnd,
        timezone,
      })
      toast.success(t('quietHours.saved'))
    } catch {
      toast.error(t('updateError'))
    }
  }

  const handleReset = async () => {
    try {
      await resetPreferences.mutateAsync()
      initFormValues()
      toast.success(t('resetSuccess'))
    } catch {
      toast.error(t('updateError'))
    }
  }

  if (isLoading) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center min-h-[400px]">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </AppLayout>
    )
  }

  return (
    <AppLayout>
      <div className="max-w-2xl mx-auto space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">{t('title')}</h1>
            <p className="text-muted-foreground">{t('subtitle')}</p>
          </div>
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="outline" size="sm">
                <RotateCcw className="h-4 w-4 mr-2" />
                {tSettings('reset')}
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>{tSettings('resetSettings')}</AlertDialogTitle>
                <AlertDialogDescription>{t('resetDescription')}</AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>{tCommon('cancel')}</AlertDialogCancel>
                <AlertDialogAction onClick={handleReset}>{tSettings('reset')}</AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>

        {/* Channels */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bell className="h-5 w-5" />
              {t('channels.title')}
            </CardTitle>
            <CardDescription>{t('channels.subtitle')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Bell className="h-5 w-5 text-muted-foreground" />
                <div>
                  <Label>{t('channels.inApp')}</Label>
                  <p className="text-sm text-muted-foreground">{t('channels.inAppDesc')}</p>
                </div>
              </div>
              <Switch
                checked={preferences?.in_app_enabled ?? true}
                onCheckedChange={(checked) => handleToggleChannel('in_app', checked)}
                disabled={toggleChannel.isPending}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Mail className="h-5 w-5 text-muted-foreground" />
                <div>
                  <Label>{t('channels.email')}</Label>
                  <p className="text-sm text-muted-foreground">{t('channels.emailDesc')}</p>
                </div>
              </div>
              <Switch
                checked={preferences?.email_enabled ?? true}
                onCheckedChange={(checked) => handleToggleChannel('email', checked)}
                disabled={toggleChannel.isPending}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Smartphone className="h-5 w-5 text-muted-foreground" />
                <div>
                  <Label>{t('channels.push')}</Label>
                  <p className="text-sm text-muted-foreground">{t('channels.pushDesc')}</p>
                </div>
              </div>
              <Switch
                checked={preferences?.push_enabled ?? false}
                onCheckedChange={(checked) => handleToggleChannel('push', checked)}
                disabled={toggleChannel.isPending}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <MessageSquare className="h-5 w-5 text-muted-foreground" />
                <div>
                  <Label>{t('channels.slack')}</Label>
                  <p className="text-sm text-muted-foreground">{t('channels.slackDesc')}</p>
                </div>
              </div>
              <Switch
                checked={preferences?.slack_enabled ?? false}
                onCheckedChange={(checked) => handleToggleChannel('slack', checked)}
                disabled={toggleChannel.isPending}
              />
            </div>
          </CardContent>
        </Card>

        {/* Telegram Integration */}
        <TelegramLinkCard />

        {/* Push Notifications */}
        <PushNotificationSettings />

        {/* Quiet Hours */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <Clock className="h-5 w-5" />
                  {t('quietHours.title')}
                </CardTitle>
                <CardDescription>{t('quietHours.subtitle')}</CardDescription>
              </div>
              <Switch
                checked={preferences?.quiet_hours_enabled ?? false}
                onCheckedChange={handleQuietHoursToggle}
                disabled={updateQuietHours.isPending}
              />
            </div>
          </CardHeader>
          {preferences?.quiet_hours_enabled && (
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="quiet-start">{t('quietHours.start')}</Label>
                  <Input
                    id="quiet-start"
                    type="time"
                    value={quietHoursStart}
                    onChange={(e) => setQuietHoursStart(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="quiet-end">{t('quietHours.end')}</Label>
                  <Input
                    id="quiet-end"
                    type="time"
                    value={quietHoursEnd}
                    onChange={(e) => setQuietHoursEnd(e.target.value)}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="timezone">{t('quietHours.timezone')}</Label>
                <Select value={timezone} onValueChange={setTimezone}>
                  <SelectTrigger id="timezone">
                    <SelectValue placeholder={t('quietHours.timezonePlaceholder')} />
                  </SelectTrigger>
                  <SelectContent>
                    {timezonesData?.timezones.map((tz) => (
                      <SelectItem key={tz} value={tz}>
                        {tz}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <Button
                onClick={handleSaveQuietHours}
                disabled={updateQuietHours.isPending}
                className="w-full"
              >
                {updateQuietHours.isPending ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    {t('quietHours.saving')}
                  </>
                ) : (
                  t('quietHours.save')
                )}
              </Button>
            </CardContent>
          )}
        </Card>
      </div>
    </AppLayout>
  )
}
