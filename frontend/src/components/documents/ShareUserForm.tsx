'use client'

import { memo } from 'react'
import { useTranslations } from 'next-intl'
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
import { Calendar, Loader2 } from 'lucide-react'
import { PermissionLevel, UserRole } from '@/lib/api/documents'
import { User } from '@/lib/api/users'

interface ShareUserFormProps {
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
  users: User[]
  saving: boolean
  onShare: () => void
  permissionLabels: Record<PermissionLevel, string>
  roleLabels: Record<UserRole, string>
}

export const ShareUserForm = memo(function ShareUserForm({
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
  saving,
  onShare,
  permissionLabels,
  roleLabels,
}: ShareUserFormProps) {
  const t = useTranslations('documents.share')

  return (
    <div className="space-y-4 p-4 border rounded-lg bg-muted/50">
      <div className="flex gap-2">
        <Button
          variant={shareType === 'user' ? 'default' : 'outline'}
          size="sm"
          onClick={() => onShareTypeChange('user')}
        >
          {t('toUser')}
        </Button>
        <Button
          variant={shareType === 'role' ? 'default' : 'outline'}
          size="sm"
          onClick={() => onShareTypeChange('role')}
        >
          {t('byRole')}
        </Button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        {shareType === 'user' ? (
          <div className="space-y-2">
            <Label>{t('user')}</Label>
            <Select value={selectedUserId} onValueChange={onSelectedUserIdChange}>
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
            <Select value={selectedRole} onValueChange={(v) => onSelectedRoleChange(v as UserRole)}>
              <SelectTrigger>
                <SelectValue placeholder={t('selectRole')} />
              </SelectTrigger>
              <SelectContent>
                {(Object.keys(roleLabels) as UserRole[]).map((role) => (
                  <SelectItem key={role} value={role}>
                    {roleLabels[role]}
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
            onValueChange={(v) => onSelectedPermissionChange(v as PermissionLevel)}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {(Object.keys(permissionLabels) as PermissionLevel[]).map((perm) => (
                <SelectItem key={perm} value={perm}>
                  {permissionLabels[perm]}
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
          onChange={(e) => onExpiresAtChange(e.target.value)}
        />
      </div>

      <Button onClick={onShare} disabled={saving} className="w-full">
        {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
        {t('grantAccess')}
      </Button>
    </div>
  )
})
