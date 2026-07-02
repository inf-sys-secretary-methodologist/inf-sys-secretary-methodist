'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { Copy, RotateCcw, Trash2, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { AppLayout } from '@/components/layout'
import { SettingsTabs } from '@/components/settings/SettingsTabs'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
import {
  useCalendarSubscription,
  createCalendarSubscription,
  rotateCalendarSubscription,
  deleteCalendarSubscription,
} from '@/hooks/useCalendarSubscription'

export default function CalendarSettingsPage() {
  const t = useTranslations('settings.calendar')
  const tCommon = useTranslations('common')
  const { subscription, isLoading, mutate } = useCalendarSubscription()
  const [busy, setBusy] = useState(false)

  const run = async (action: () => Promise<unknown>, successKey: string) => {
    setBusy(true)
    try {
      await action()
      await mutate()
      toast.success(t(successKey))
    } catch {
      toast.error(t('error'))
    } finally {
      setBusy(false)
    }
  }

  const handleCopy = async () => {
    if (!subscription?.url) return
    try {
      await navigator.clipboard.writeText(subscription.url)
      toast.success(t('copied'))
    } catch {
      toast.error(t('error'))
    }
  }

  const isSubscribed = Boolean(subscription?.subscribed && subscription.url)

  return (
    <AppLayout>
      <SettingsTabs />
      <div className="max-w-2xl mx-auto space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>{t('title')}</CardTitle>
            <CardDescription>{t('description')}</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : isSubscribed ? (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>{t('urlLabel')}</Label>
                  <div className="flex gap-2">
                    <Input readOnly value={subscription!.url} className="font-mono text-xs" />
                    <Button variant="outline" onClick={handleCopy}>
                      <Copy className="h-4 w-4 mr-2" />
                      {t('copy')}
                    </Button>
                  </div>
                  <p className="text-sm text-muted-foreground">{t('urlHint')}</p>
                </div>

                <div className="flex flex-wrap gap-2">
                  <AlertDialog>
                    <AlertDialogTrigger asChild>
                      <Button variant="outline" disabled={busy}>
                        <RotateCcw className="h-4 w-4 mr-2" />
                        {t('regenerate')}
                      </Button>
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>{t('regenerateConfirmTitle')}</AlertDialogTitle>
                        <AlertDialogDescription>
                          {t('regenerateConfirmDescription')}
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>{tCommon('cancel')}</AlertDialogCancel>
                        <AlertDialogAction
                          onClick={() => run(rotateCalendarSubscription, 'regenerated')}
                        >
                          {t('regenerate')}
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>

                  <AlertDialog>
                    <AlertDialogTrigger asChild>
                      <Button variant="destructive" disabled={busy}>
                        <Trash2 className="h-4 w-4 mr-2" />
                        {t('delete')}
                      </Button>
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>{t('deleteConfirmTitle')}</AlertDialogTitle>
                        <AlertDialogDescription>
                          {t('deleteConfirmDescription')}
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>{tCommon('cancel')}</AlertDialogCancel>
                        <AlertDialogAction
                          onClick={() => run(deleteCalendarSubscription, 'deleted')}
                        >
                          {t('delete')}
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </div>
              </div>
            ) : (
              <div className="space-y-4">
                <div>
                  <p className="font-medium">{t('notSubscribedTitle')}</p>
                  <p className="text-sm text-muted-foreground">{t('notSubscribedDescription')}</p>
                </div>
                <Button onClick={() => run(createCalendarSubscription, 'created')} disabled={busy}>
                  {busy && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                  {t('create')}
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('howToTitle')}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground whitespace-pre-line">
              {t('howToDescription')}
            </p>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  )
}
