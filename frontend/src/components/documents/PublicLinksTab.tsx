'use client'

import { memo } from 'react'
import { useTranslations, useLocale } from 'next-intl'
import { TabsContent } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Trash2, Copy, Check, Loader2, Calendar, Lock, Eye, Download } from 'lucide-react'
import { PublicLinkInfo } from '@/lib/api/documents'

interface PublicLinksTabProps {
  // Form state
  linkPermission: 'read' | 'download'
  onLinkPermissionChange: (permission: 'read' | 'download') => void
  linkExpiresAt: string
  onLinkExpiresAtChange: (expiresAt: string) => void
  linkMaxUses: string
  onLinkMaxUsesChange: (maxUses: string) => void
  linkPassword: string
  onLinkPasswordChange: (password: string) => void

  // Data
  publicLinks: PublicLinkInfo[]
  loading: boolean
  saving: boolean
  copiedLinkId: number | null

  // Actions
  onCreatePublicLink: () => void
  onCopyLink: (link: PublicLinkInfo) => void
  onDeactivateLink: (linkId: number) => void
  onDeleteLink: (linkId: number) => void
}

export const PublicLinksTab = memo(function PublicLinksTab({
  linkPermission,
  onLinkPermissionChange,
  linkExpiresAt,
  onLinkExpiresAtChange,
  linkMaxUses,
  onLinkMaxUsesChange,
  linkPassword,
  onLinkPasswordChange,
  publicLinks,
  loading,
  saving,
  copiedLinkId,
  onCreatePublicLink,
  onCopyLink,
  onDeactivateLink,
  onDeleteLink,
}: PublicLinksTabProps) {
  const t = useTranslations('documents.share')
  const tForm = useTranslations('documents.form')
  const locale = useLocale()

  return (
    <TabsContent value="links" className="space-y-4">
      {/* Create link form */}
      <div className="space-y-4 p-4 border rounded-lg bg-muted/50">
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label className="flex items-center gap-2">
              {linkPermission === 'read' ? (
                <Eye className="h-4 w-4" />
              ) : (
                <Download className="h-4 w-4" />
              )}
              {t('accessType')}
            </Label>
            <Select
              value={linkPermission}
              onValueChange={(v) => onLinkPermissionChange(v as 'read' | 'download')}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="read">{t('viewOnly')}</SelectItem>
                <SelectItem value="download">{t('viewAndDownload')}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label className="flex items-center gap-2">
              <Calendar className="h-4 w-4" />
              {t('expiry')}
            </Label>
            <Input
              type="datetime-local"
              value={linkExpiresAt}
              onChange={(e) => onLinkExpiresAtChange(e.target.value)}
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label>{t('maxUses')}</Label>
            <Input
              type="number"
              min="1"
              placeholder={t('noLimitPlaceholder')}
              value={linkMaxUses}
              onChange={(e) => onLinkMaxUsesChange(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label className="flex items-center gap-2">
              <Lock className="h-4 w-4" />
              {t('passwordOptional')}
            </Label>
            <Input
              type="password"
              placeholder={t('noPasswordPlaceholder')}
              value={linkPassword}
              onChange={(e) => onLinkPasswordChange(e.target.value)}
            />
          </div>
        </div>

        <Button onClick={onCreatePublicLink} disabled={saving} className="w-full">
          {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {t('createPublicLink')}
        </Button>
      </div>

      {/* Existing links */}
      <div className="space-y-2">
        <Label>{t('existingLinks')}</Label>
        {loading ? (
          <div className="flex justify-center p-4">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : publicLinks.length === 0 ? (
          <p className="text-sm text-muted-foreground p-4 text-center">{t('noPublicLinks')}</p>
        ) : (
          <div className="space-y-2">
            {publicLinks.map((link) => (
              <div
                key={link.id}
                /* c8 ignore next 3 - Inactive link styling */
                className={`p-3 border rounded-lg space-y-2 ${!link.is_active ? 'opacity-50' : ''}`}
              >
                {/* c8 ignore start - Link display with permission icons */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    {link.permission === 'download' ? (
                      <Download className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                    <span className="text-sm font-mono truncate max-w-[200px]">{link.token}</span>
                    {link.has_password && <Lock className="h-3 w-3" />}
                  </div>
                  <div className="flex items-center gap-1">
                    {!link.is_active && (
                      <Badge variant="outline" className="text-xs">
                        {t('inactive')}
                      </Badge>
                    )}
                    <Badge variant="secondary" className="text-xs">
                      {link.use_count} {t('uses')}
                    </Badge>
                  </div>
                </div>

                <div className="flex items-center justify-between text-xs text-muted-foreground">
                  <span>
                    {link.expires_at
                      ? `${t('until')} ${new Date(link.expires_at).toLocaleDateString(locale)}`
                      : t('unlimited')}
                    {link.max_uses && ` • ${t('maxUsesLabel', { count: link.max_uses })}`}
                  </span>
                  <div className="flex gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7"
                      onClick={() => onCopyLink(link)}
                      aria-label={tForm('copyLink')}
                    >
                      {/* c8 ignore next 3 - Copy feedback icon */}
                      {copiedLinkId === link.id ? (
                        <Check className="h-3 w-3 text-green-500" />
                      ) : (
                        <Copy className="h-3 w-3" />
                      )}
                    </Button>
                    {link.is_active && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() => onDeactivateLink(link.id)}
                        aria-label={tForm('deactivateLink')}
                      >
                        <Eye className="h-3 w-3" />
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7"
                      onClick={() => onDeleteLink(link.id)}
                      aria-label={tForm('deleteLink')}
                    >
                      <Trash2 className="h-3 w-3 text-destructive" />
                    </Button>
                  </div>
                </div>
                {/* c8 ignore stop */}
              </div>
            ))}
          </div>
        )}
      </div>
    </TabsContent>
  )
})
