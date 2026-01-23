'use client'

import { useState, useEffect, useCallback, useMemo } from 'react'
import { useTranslations, useLocale } from 'next-intl'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import {
  Trash2,
  Users,
  Link2,
  Copy,
  Check,
  Loader2,
  Calendar,
  Lock,
  Eye,
  Download,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  documentsApi,
  PermissionInfo,
  PublicLinkInfo,
  PermissionLevel,
  UserRole,
} from '@/lib/api/documents'
import { usersApi, User } from '@/lib/api/users'

interface ShareDocumentDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  documentId: number | string
  documentTitle: string
}

export function ShareDocumentDialog({
  open,
  onOpenChange,
  documentId,
  documentTitle,
}: ShareDocumentDialogProps) {
  const t = useTranslations('documents.share')
  const tCommon = useTranslations('common')
  const tForm = useTranslations('documents.form')
  const locale = useLocale()

  const PERMISSION_LABELS = useMemo<Record<PermissionLevel, string>>(
    () => ({
      read: t('permissionRead'),
      write: t('permissionWrite'),
      delete: t('permissionDelete'),
      admin: t('permissionAdmin'),
    }),
    [t]
  )

  const ROLE_LABELS = useMemo<Record<UserRole, string>>(
    () => ({
      admin: t('roleAdmin'),
      secretary: t('roleSecretary'),
      methodist: t('roleMethodist'),
      teacher: t('roleTeacher'),
      student: t('roleStudent'),
    }),
    [t]
  )

  const [activeTab, setActiveTab] = useState('users')
  const [permissions, setPermissions] = useState<PermissionInfo[]>([])
  const [publicLinks, setPublicLinks] = useState<PublicLinkInfo[]>([])
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)

  // Share form state
  const [shareType, setShareType] = useState<'user' | 'role'>('user')
  const [selectedUserId, setSelectedUserId] = useState<string>('')
  const [selectedRole, setSelectedRole] = useState<UserRole | ''>('')
  const [selectedPermission, setSelectedPermission] = useState<PermissionLevel>('read')
  const [expiresAt, setExpiresAt] = useState('')

  // Public link form state
  const [linkPermission, setLinkPermission] = useState<'read' | 'download'>('read')
  const [linkExpiresAt, setLinkExpiresAt] = useState('')
  const [linkMaxUses, setLinkMaxUses] = useState('')
  const [linkPassword, setLinkPassword] = useState('')
  const [copiedLinkId, setCopiedLinkId] = useState<number | null>(null)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const [perms, links, usersList] = await Promise.all([
        documentsApi.getPermissions(documentId),
        documentsApi.getPublicLinks(documentId),
        usersApi.getAll().catch(() => []),
      ])
      setPermissions(perms)
      setPublicLinks(links)
      setUsers(usersList)
    } catch (_error) {
      toast.error(t('dataLoadError'))
    } finally {
      setLoading(false)
    }
  }, [documentId])

  useEffect(() => {
    if (open) {
      loadData()
    }
  }, [open, loadData])

  const handleShare = async () => {
    if (shareType === 'user' && !selectedUserId) {
      toast.error(t('selectUser'))
      return
    }
    if (shareType === 'role' && !selectedRole) {
      toast.error(t('selectRole'))
      return
    }

    setSaving(true)
    try {
      await documentsApi.shareDocument(documentId, {
        user_id: shareType === 'user' ? Number(selectedUserId) : undefined,
        role: shareType === 'role' ? (selectedRole as UserRole) : undefined,
        permission: selectedPermission,
        expires_at: expiresAt || undefined,
      })
      toast.success(t('accessGranted'))
      setSelectedUserId('')
      setSelectedRole('')
      setExpiresAt('')
      await loadData()
    } catch (_error) {
      toast.error(t('grantError'))
    } finally {
      setSaving(false)
    }
  }

  const handleRevokePermission = async (permissionId: number) => {
    try {
      await documentsApi.revokePermission(documentId, permissionId)
      toast.success(t('accessRevoked'))
      setPermissions((prev) => prev.filter((p) => p.id !== permissionId))
    } catch (_error) {
      toast.error(t('revokeError'))
    }
  }

  const handleCreatePublicLink = async () => {
    setSaving(true)
    try {
      const newLink = await documentsApi.createPublicLink(documentId, {
        permission: linkPermission,
        expires_at: linkExpiresAt || undefined,
        max_uses: linkMaxUses ? Number(linkMaxUses) : undefined,
        password: linkPassword || undefined,
      })
      toast.success(t('linkCreated'))
      setPublicLinks((prev) => [newLink, ...prev])
      setLinkExpiresAt('')
      setLinkMaxUses('')
      setLinkPassword('')
    } catch (_error) {
      toast.error(t('linkCreateError'))
    } finally {
      setSaving(false)
    }
  }

  /* c8 ignore start -- Browser API navigator.clipboard */
  const handleCopyLink = async (link: PublicLinkInfo) => {
    try {
      await navigator.clipboard.writeText(link.url)
      setCopiedLinkId(link.id)
      toast.success(t('linkCopied'))
      setTimeout(() => setCopiedLinkId(null), 2000)
    } catch {
      toast.error(t('linkCopyError'))
    }
  }
  /* c8 ignore stop */

  const handleDeactivateLink = async (linkId: number) => {
    try {
      await documentsApi.deactivatePublicLink(documentId, linkId)
      toast.success(t('linkDeactivated'))
      setPublicLinks((prev) => prev.map((l) => (l.id === linkId ? { ...l, is_active: false } : l)))
    } catch (_error) {
      toast.error(t('linkDeactivateError'))
    }
  }

  const handleDeleteLink = async (linkId: number) => {
    try {
      await documentsApi.deletePublicLink(documentId, linkId)
      toast.success(t('linkDeleted'))
      setPublicLinks((prev) => prev.filter((l) => l.id !== linkId))
    } catch (_error) {
      toast.error(t('linkDeleteError'))
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t('title')}</DialogTitle>
          <DialogDescription className="truncate">{documentTitle}</DialogDescription>
        </DialogHeader>

        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="users" className="flex items-center gap-2">
              <Users className="h-4 w-4" />
              {t('usersTab')}
            </TabsTrigger>
            <TabsTrigger value="links" className="flex items-center gap-2">
              <Link2 className="h-4 w-4" />
              {t('linksTab')}
            </TabsTrigger>
          </TabsList>

          {/* Users/Roles Tab */}
          <TabsContent value="users" className="space-y-4">
            {/* Add permission form */}
            <div className="space-y-4 p-4 border rounded-lg bg-muted/50">
              <div className="flex gap-2">
                <Button
                  variant={shareType === 'user' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setShareType('user')}
                >
                  {t('toUser')}
                </Button>
                <Button
                  variant={shareType === 'role' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setShareType('role')}
                >
                  {t('byRole')}
                </Button>
              </div>

              <div className="grid grid-cols-2 gap-4">
                {shareType === 'user' ? (
                  <div className="space-y-2">
                    <Label>{t('user')}</Label>
                    <Select value={selectedUserId} onValueChange={setSelectedUserId}>
                      <SelectTrigger>
                        <SelectValue placeholder={t('selectUser')} />
                      </SelectTrigger>
                      <SelectContent>
                        {users.map((user) => (
                          <SelectItem key={user.id} value={String(user.id)}>
                            {user.name} ({user.email})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                ) : (
                  <div className="space-y-2">
                    <Label>{t('role')}</Label>
                    <Select
                      value={selectedRole}
                      onValueChange={(v) => setSelectedRole(v as UserRole)}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder={t('selectRole')} />
                      </SelectTrigger>
                      <SelectContent>
                        {(Object.keys(ROLE_LABELS) as UserRole[]).map((role) => (
                          <SelectItem key={role} value={role}>
                            {ROLE_LABELS[role]}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                )}

                <div className="space-y-2">
                  <Label>{t('accessLevel')}</Label>
                  <Select
                    value={selectedPermission}
                    onValueChange={(v) => setSelectedPermission(v as PermissionLevel)}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {(Object.keys(PERMISSION_LABELS) as PermissionLevel[]).map((perm) => (
                        <SelectItem key={perm} value={perm}>
                          {PERMISSION_LABELS[perm]}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="space-y-2">
                <Label className="flex items-center gap-2">
                  <Calendar className="h-4 w-4" />
                  {t('expiryOptional')}
                </Label>
                <Input
                  type="datetime-local"
                  value={expiresAt}
                  onChange={(e) => setExpiresAt(e.target.value)}
                />
              </div>

              <Button onClick={handleShare} disabled={saving} className="w-full">
                {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {t('grantAccess')}
              </Button>
            </div>

            {/* Existing permissions */}
            <div className="space-y-2">
              <Label>{t('currentPermissions')}</Label>
              {loading ? (
                <div className="flex justify-center p-4">
                  <Loader2 className="h-6 w-6 animate-spin" />
                </div>
              ) : permissions.length === 0 ? (
                <p className="text-sm text-muted-foreground p-4 text-center">
                  {t('noAccessGranted')}
                </p>
              ) : (
                <div className="space-y-2">
                  {permissions.map((perm) => (
                    <div
                      key={perm.id}
                      className="flex items-center justify-between p-3 border rounded-lg"
                    >
                      <div className="flex flex-col">
                        <span className="font-medium">
                          {perm.user_name || perm.role || t('unknown')}
                        </span>
                        {perm.user_email && (
                          <span className="text-sm text-muted-foreground">{perm.user_email}</span>
                        )}
                        {perm.expires_at && (
                          <span className="text-xs text-muted-foreground">
                            {t('until')}: {new Date(perm.expires_at).toLocaleDateString(locale)}
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge variant="secondary">{PERMISSION_LABELS[perm.permission]}</Badge>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleRevokePermission(perm.id)}
                          aria-label={tForm('revokeAccess')}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </TabsContent>

          {/* Public Links Tab */}
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
                    onValueChange={(v) => setLinkPermission(v as 'read' | 'download')}
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
                    onChange={(e) => setLinkExpiresAt(e.target.value)}
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
                    onChange={(e) => setLinkMaxUses(e.target.value)}
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
                    onChange={(e) => setLinkPassword(e.target.value)}
                  />
                </div>
              </div>

              <Button onClick={handleCreatePublicLink} disabled={saving} className="w-full">
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
                <p className="text-sm text-muted-foreground p-4 text-center">
                  {t('noPublicLinks')}
                </p>
              ) : (
                <div className="space-y-2">
                  {publicLinks.map((link) => (
                    <div
                      key={link.id}
                      className={`p-3 border rounded-lg space-y-2 ${
                        !link.is_active ? 'opacity-50' : ''
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          {link.permission === 'download' ? (
                            <Download className="h-4 w-4" />
                          ) : (
                            <Eye className="h-4 w-4" />
                          )}
                          <span className="text-sm font-mono truncate max-w-[200px]">
                            {link.token}
                          </span>
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
                            onClick={() => handleCopyLink(link)}
                            aria-label={tForm('copyLink')}
                          >
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
                              onClick={() => handleDeactivateLink(link.id)}
                              aria-label={tForm('deactivateLink')}
                            >
                              <Eye className="h-3 w-3" />
                            </Button>
                          )}
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => handleDeleteLink(link.id)}
                            aria-label={tForm('deleteLink')}
                          >
                            <Trash2 className="h-3 w-3 text-destructive" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {tCommon('close')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
