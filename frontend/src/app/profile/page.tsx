'use client'

import { useState, useEffect } from 'react'
import { useAuthStore } from '@/stores/authStore'
import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { AvatarUpload } from '@/components/profile/AvatarUpload'
import { usersApi } from '@/lib/api/users'
import {
  User,
  Mail,
  Calendar,
  Shield,
  Edit2,
  X,
  Check,
  Phone,
  FileText,
  Loader2,
} from 'lucide-react'
import { UserRole } from '@/types/auth'

const roleLabels: Record<UserRole, string> = {
  [UserRole.SYSTEM_ADMIN]: 'Системный администратор',
  [UserRole.METHODIST]: 'Методист',
  [UserRole.ACADEMIC_SECRETARY]: 'Учёный секретарь',
  [UserRole.TEACHER]: 'Преподаватель',
  [UserRole.STUDENT]: 'Студент',
}

function ProfilePage() {
  const { user, checkAuth } = useAuthStore()
  const [isEditing, setIsEditing] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [editedPhone, setEditedPhone] = useState('')
  const [editedBio, setEditedBio] = useState('')
  const [avatarUrl, setAvatarUrl] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (user) {
      setEditedPhone((user as unknown as { phone?: string }).phone || '')
      setEditedBio((user as unknown as { bio?: string }).bio || '')
      // Avatar URL is now returned directly from /api/me as presigned URL
      const avatar = (user as unknown as { avatar?: string }).avatar
      if (avatar) {
        setAvatarUrl(avatar)
      }
    }
  }, [user])

  if (!user) {
    return null
  }

  const handleSave = async () => {
    setIsSaving(true)
    setError(null)
    try {
      await usersApi.updateProfile(user.id, {
        phone: editedPhone,
        bio: editedBio,
      })
      await checkAuth()
      setIsEditing(false)
    } catch (err: unknown) {
      // Extract error message from API response
      const apiError = err as { response?: { data?: { message?: string; error?: string } } }
      const errorMessage =
        apiError?.response?.data?.message ||
        apiError?.response?.data?.error ||
        'Ошибка при сохранении профиля'

      // Translate common validation errors
      let userFriendlyError = errorMessage
      if (errorMessage.includes('e164') || errorMessage.includes('phone')) {
        userFriendlyError = 'Телефон должен быть в международном формате (например, +79991234567)'
      } else if (errorMessage.includes('bio') || errorMessage.includes('max=500')) {
        userFriendlyError = 'Описание не должно превышать 500 символов'
      }

      setError(userFriendlyError)
      console.error('Failed to save profile:', err)
    } finally {
      setIsSaving(false)
    }
  }

  const handleCancel = () => {
    setEditedPhone((user as unknown as { phone?: string }).phone || '')
    setEditedBio((user as unknown as { bio?: string }).bio || '')
    setIsEditing(false)
    setError(null)
  }

  const handleAvatarUpload = async (file: File) => {
    try {
      const response = await usersApi.uploadAvatar(user.id, file)
      if (response.data?.avatar_url) {
        setAvatarUrl(response.data.avatar_url)
      }
      await checkAuth()
    } catch (err) {
      console.error('Failed to upload avatar:', err)
      throw err
    }
  }

  const handleAvatarRemove = async () => {
    try {
      await usersApi.deleteAvatar(user.id)
      setAvatarUrl(null)
      await checkAuth()
    } catch (err) {
      console.error('Failed to delete avatar:', err)
      throw err
    }
  }

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'Не указано'
    const date = new Date(dateString)
    if (isNaN(date.getTime())) return 'Не указано'
    return date.toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  }

  return (
    <AppLayout>
      <div className="max-w-4xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">Профиль пользователя</h1>
            <p className="text-sm sm:text-base text-muted-foreground mt-1">
              Управляйте вашей учётной записью и персональными данными
            </p>
          </div>

          {!isEditing && (
            <Button onClick={() => setIsEditing(true)} className="gap-2 w-full sm:w-auto">
              <Edit2 className="h-4 w-4" />
              Редактировать
            </Button>
          )}
        </div>

        {/* Avatar Card */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg sm:text-xl">Фото профиля</CardTitle>
            <CardDescription>Загрузите фотографию для вашего профиля</CardDescription>
          </CardHeader>
          <CardContent>
            <AvatarUpload
              currentAvatar={avatarUrl}
              userName={user.name}
              onUpload={handleAvatarUpload}
              onRemove={handleAvatarRemove}
              disabled={isSaving}
            />
          </CardContent>
        </Card>

        {/* Profile Information Card */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg sm:text-xl">Персональная информация</CardTitle>
            <CardDescription>Ваши основные данные учётной записи</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 sm:space-y-6">
            {error && (
              <div className="p-3 text-sm text-red-500 bg-red-50 dark:bg-red-900/20 rounded-md">
                {error}
              </div>
            )}

            {/* Name Field */}
            <div className="space-y-2">
              <Label htmlFor="name" className="flex items-center gap-2 text-sm">
                <User className="h-4 w-4 text-muted-foreground" />
                Имя
              </Label>
              <p className="text-sm text-foreground bg-secondary/50 dark:bg-secondary/30 p-3 rounded-md border border-border/50">
                {user.name}
              </p>
            </div>

            {/* Email Field */}
            <div className="space-y-2">
              <Label htmlFor="email" className="flex items-center gap-2 text-sm">
                <Mail className="h-4 w-4 text-muted-foreground" />
                Email
              </Label>
              <p className="text-sm text-foreground bg-secondary/50 dark:bg-secondary/30 p-3 rounded-md border border-border/50 break-all">
                {user.email}
              </p>
            </div>

            {/* Phone Field */}
            <div className="space-y-2">
              <Label htmlFor="phone" className="flex items-center gap-2 text-sm">
                <Phone className="h-4 w-4 text-muted-foreground" />
                Телефон
              </Label>
              {isEditing ? (
                <Input
                  id="phone"
                  type="tel"
                  value={editedPhone}
                  onChange={(e) => setEditedPhone(e.target.value)}
                  placeholder="+7 (999) 123-45-67"
                />
              ) : (
                <p className="text-sm text-foreground bg-secondary/50 dark:bg-secondary/30 p-3 rounded-md border border-border/50">
                  {(user as unknown as { phone?: string }).phone || 'Не указан'}
                </p>
              )}
            </div>

            {/* Bio Field */}
            <div className="space-y-2">
              <Label htmlFor="bio" className="flex items-center gap-2 text-sm">
                <FileText className="h-4 w-4 text-muted-foreground" />О себе
              </Label>
              {isEditing ? (
                <Textarea
                  id="bio"
                  value={editedBio}
                  onChange={(e) => setEditedBio(e.target.value)}
                  placeholder="Расскажите немного о себе..."
                  rows={3}
                />
              ) : (
                <p className="text-sm text-foreground bg-secondary/50 dark:bg-secondary/30 p-3 rounded-md border border-border/50 min-h-[60px]">
                  {(user as unknown as { bio?: string }).bio || 'Не указано'}
                </p>
              )}
            </div>

            {/* Role Field (Read-only) */}
            <div className="space-y-2">
              <Label className="flex items-center gap-2 text-sm">
                <Shield className="h-4 w-4 text-muted-foreground" />
                Роль
              </Label>
              <p className="text-sm text-foreground bg-secondary/50 dark:bg-secondary/30 p-3 rounded-md border border-border/50">
                {roleLabels[user.role]}
              </p>
            </div>

            {/* Edit Actions */}
            {isEditing && (
              <div className="flex flex-col sm:flex-row gap-3 pt-4">
                <Button onClick={handleSave} disabled={isSaving} className="gap-2 w-full sm:w-auto">
                  {isSaving ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Check className="h-4 w-4" />
                  )}
                  Сохранить
                </Button>
                <Button
                  onClick={handleCancel}
                  variant="outline"
                  disabled={isSaving}
                  className="gap-2 w-full sm:w-auto"
                >
                  <X className="h-4 w-4" />
                  Отмена
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Account Information Card */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg sm:text-xl">Информация об учётной записи</CardTitle>
            <CardDescription>Дополнительные данные вашего аккаунта</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Created At */}
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between py-2 border-b gap-1">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Calendar className="h-4 w-4" />
                <span>Дата создания</span>
              </div>
              <p className="text-sm font-medium">{formatDate(user.created_at)}</p>
            </div>

            {/* Updated At */}
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between py-2 border-b gap-1">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Calendar className="h-4 w-4" />
                <span>Последнее обновление</span>
              </div>
              <p className="text-sm font-medium">{formatDate(user.updated_at)}</p>
            </div>

            {/* User ID */}
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between py-2 gap-1">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <User className="h-4 w-4" />
                <span>ID пользователя</span>
              </div>
              <p className="text-xs font-mono bg-secondary/50 dark:bg-secondary/30 px-2 py-1 rounded border border-border/50 break-all">
                {user.id}
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  )
}

export default ProfilePage
