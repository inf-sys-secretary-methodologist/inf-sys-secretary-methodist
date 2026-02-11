'use client'

import { useState, useEffect, useCallback, useMemo } from 'react'
import { useTranslations } from 'next-intl'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Users, Link2 } from 'lucide-react'
import { toast } from 'sonner'
import {
  documentsApi,
  PermissionInfo,
  PublicLinkInfo,
  PermissionLevel,
  UserRole,
} from '@/lib/api/documents'
import { usersApi, User } from '@/lib/api/users'
import { PermissionsTab } from './PermissionsTab'
import { PublicLinksTab } from './PublicLinksTab'

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

  /* c8 ignore start - Data loading and API handlers, tested in e2e */
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
  }, [documentId, t])

  useEffect(() => {
    if (open) {
      loadData()
    }
  }, [open, loadData])

  const handleShare = useCallback(async () => {
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
  }, [
    shareType,
    selectedUserId,
    selectedRole,
    selectedPermission,
    expiresAt,
    documentId,
    t,
    loadData,
  ])

  const handleRevokePermission = useCallback(
    async (permissionId: number) => {
      try {
        await documentsApi.revokePermission(documentId, permissionId)
        toast.success(t('accessRevoked'))
        setPermissions((prev) => prev.filter((p) => p.id !== permissionId))
      } catch (_error) {
        toast.error(t('revokeError'))
      }
    },
    [documentId, t]
  )

  const handleCreatePublicLink = useCallback(async () => {
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
  }, [documentId, linkPermission, linkExpiresAt, linkMaxUses, linkPassword, t])

  const handleCopyLink = useCallback(
    async (link: PublicLinkInfo) => {
      try {
        await navigator.clipboard.writeText(link.url)
        setCopiedLinkId(link.id)
        toast.success(t('linkCopied'))
        setTimeout(() => setCopiedLinkId(null), 2000)
      } catch {
        toast.error(t('linkCopyError'))
      }
    },
    [t]
  )

  const handleDeactivateLink = useCallback(
    async (linkId: number) => {
      try {
        await documentsApi.deactivatePublicLink(documentId, linkId)
        toast.success(t('linkDeactivated'))
        setPublicLinks((prev) =>
          prev.map((l) => (l.id === linkId ? { ...l, is_active: false } : l))
        )
      } catch (_error) {
        toast.error(t('linkDeactivateError'))
      }
    },
    [documentId, t]
  )

  const handleDeleteLink = useCallback(
    async (linkId: number) => {
      try {
        await documentsApi.deletePublicLink(documentId, linkId)
        toast.success(t('linkDeleted'))
        setPublicLinks((prev) => prev.filter((l) => l.id !== linkId))
      } catch (_error) {
        toast.error(t('linkDeleteError'))
      }
    },
    [documentId, t]
  )
  /* c8 ignore stop */

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

          <PermissionsTab
            shareType={shareType}
            onShareTypeChange={setShareType}
            selectedUserId={selectedUserId}
            onSelectedUserIdChange={setSelectedUserId}
            selectedRole={selectedRole}
            onSelectedRoleChange={setSelectedRole}
            selectedPermission={selectedPermission}
            onSelectedPermissionChange={setSelectedPermission}
            expiresAt={expiresAt}
            onExpiresAtChange={setExpiresAt}
            users={users}
            permissions={permissions}
            loading={loading}
            saving={saving}
            permissionLabels={PERMISSION_LABELS}
            roleLabels={ROLE_LABELS}
            onShare={handleShare}
            onRevokePermission={handleRevokePermission}
          />

          <PublicLinksTab
            linkPermission={linkPermission}
            onLinkPermissionChange={setLinkPermission}
            linkExpiresAt={linkExpiresAt}
            onLinkExpiresAtChange={setLinkExpiresAt}
            linkMaxUses={linkMaxUses}
            onLinkMaxUsesChange={setLinkMaxUses}
            linkPassword={linkPassword}
            onLinkPasswordChange={setLinkPassword}
            publicLinks={publicLinks}
            loading={loading}
            saving={saving}
            copiedLinkId={copiedLinkId}
            onCreatePublicLink={handleCreatePublicLink}
            onCopyLink={handleCopyLink}
            onDeactivateLink={handleDeactivateLink}
            onDeleteLink={handleDeleteLink}
          />
        </Tabs>

        <DialogFooter>
          {/* c8 ignore next - Close dialog button */}
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {tCommon('close')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
