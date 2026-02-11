'use client'

import { memo } from 'react'
import { useTranslations, useLocale } from 'next-intl'
import { TabsContent } from '@/components/ui/tabs'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Trash2, Loader2 } from 'lucide-react'
import { PermissionInfo, PermissionLevel, UserRole } from '@/lib/api/documents'
import { User } from '@/lib/api/users'
import { ShareUserForm } from './ShareUserForm'

interface PermissionsTabProps {
  // Form state
  shareType: 'user' | 'role'
  onShareTypeChange: (type: 'user' | 'role') => void
  selectedUserId: string
  onSelectedUserIdChange: (id: string) => void
  selectedRole: UserRole | ''
  onSelectedRoleChange: (role: UserRole | '') => void
  selectedPermission: PermissionLevel
  onSelectedPermissionChange: (permission: PermissionLevel) => void
  expiresAt: string
  onExpiresAtChange: (expiresAt: string) => void

  // Data
  users: User[]
  permissions: PermissionInfo[]
  loading: boolean
  saving: boolean

  // Labels
  permissionLabels: Record<PermissionLevel, string>
  roleLabels: Record<UserRole, string>

  // Actions
  onShare: () => void
  onRevokePermission: (permissionId: number) => void
}

export const PermissionsTab = memo(function PermissionsTab({
  shareType,
  onShareTypeChange,
  selectedUserId,
  onSelectedUserIdChange,
  selectedRole,
  onSelectedRoleChange,
  selectedPermission,
  onSelectedPermissionChange,
  expiresAt,
  onExpiresAtChange,
  users,
  permissions,
  loading,
  saving,
  permissionLabels,
  roleLabels,
  onShare,
  onRevokePermission,
}: PermissionsTabProps) {
  const t = useTranslations('documents.share')
  const tForm = useTranslations('documents.form')
  const locale = useLocale()

  return (
    <TabsContent value="users" className="space-y-4">
      {/* Add permission form */}
      <ShareUserForm
        shareType={shareType}
        onShareTypeChange={onShareTypeChange}
        selectedUserId={selectedUserId}
        onSelectedUserIdChange={onSelectedUserIdChange}
        selectedRole={selectedRole}
        onSelectedRoleChange={onSelectedRoleChange}
        selectedPermission={selectedPermission}
        onSelectedPermissionChange={onSelectedPermissionChange}
        expiresAt={expiresAt}
        onExpiresAtChange={onExpiresAtChange}
        users={users}
        saving={saving}
        onShare={onShare}
        permissionLabels={permissionLabels}
        roleLabels={roleLabels}
      />

      {/* Existing permissions */}
      <div className="space-y-2">
        <Label>{t('currentPermissions')}</Label>
        {loading ? (
          <div className="flex justify-center p-4">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : permissions.length === 0 ? (
          <p className="text-sm text-muted-foreground p-4 text-center">{t('noAccessGranted')}</p>
        ) : (
          <div className="space-y-2">
            {permissions.map((perm) => (
              <div
                key={perm.id}
                className="flex items-center justify-between p-3 border rounded-lg"
              >
                <div className="flex flex-col">
                  <span className="font-medium">
                    {/* c8 ignore next */}
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
                  <Badge variant="secondary">{permissionLabels[perm.permission]}</Badge>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => onRevokePermission(perm.id)}
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
  )
})
