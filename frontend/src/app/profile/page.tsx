'use client'

import { useState } from 'react'
import { useAuthCheck } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { User, Mail, Calendar, Shield, Edit2, X, Check } from 'lucide-react'
import { UserRole } from '@/types/auth'
import { UserMenu } from '@/components/UserMenu'
import { ThemeToggleButton } from '@/components/theme-toggle-button'
import { NavBar } from '@/components/ui/tubelight-navbar'
import { getAvailableNavItems } from '@/config/navigation'

const roleLabels: Record<UserRole, string> = {
  [UserRole.SYSTEM_ADMIN]: 'Системный администратор',
  [UserRole.METHODIST]: 'Методист',
  [UserRole.ACADEMIC_SECRETARY]: 'Учёный секретарь',
  [UserRole.TEACHER]: 'Преподаватель',
  [UserRole.STUDENT]: 'Студент',
}

function ProfilePage() {
  const { user, isLoading } = useAuthCheck()
  const [isEditing, setIsEditing] = useState(false)
  const [editedName, setEditedName] = useState(user?.name || '')
  const [editedEmail, setEditedEmail] = useState(user?.email || '')

  // Get navigation items filtered by user role
  const navItems = getAvailableNavItems(user?.role)

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="text-muted-foreground">Загрузка...</p>
        </div>
      </div>
    )
  }

  if (!user) {
    return null
  }

  const handleSave = async () => {
    // TODO: Implement API call to update user profile
    console.log('Saving profile:', { name: editedName, email: editedEmail })

    // For now, just close edit mode
    setIsEditing(false)

    // TODO: Update user in authStore after successful API call
    // updateUser({ ...user, name: editedName, email: editedEmail })
  }

  const handleCancel = () => {
    setEditedName(user.name)
    setEditedEmail(user.email)
    setIsEditing(false)
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  }

  return (
    <div className="min-h-screen bg-background p-8">
      {/* Navigation Bar */}
      <NavBar items={navItems} />

      {/* Top Navigation */}
      <div
        className="fixed top-8 right-8 z-50 pointer-events-auto flex items-center gap-3"
        style={{ isolation: 'isolate' }}
      >
        <UserMenu />
        <ThemeToggleButton />
      </div>

      <div className="container mx-auto pt-24 max-w-4xl">
        <div className="space-y-6">
          {/* Header */}
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold tracking-tight">Профиль пользователя</h1>
              <p className="text-muted-foreground mt-1">
                Управляйте вашей учётной записью и персональными данными
              </p>
            </div>

            {!isEditing && (
              <Button onClick={() => setIsEditing(true)} className="gap-2">
                <Edit2 className="h-4 w-4" />
                Редактировать
              </Button>
            )}
          </div>

          {/* Profile Information Card */}
          <Card>
            <CardHeader>
              <CardTitle>Персональная информация</CardTitle>
              <CardDescription>Ваши основные данные учётной записи</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Name Field */}
              <div className="space-y-2">
                <Label htmlFor="name" className="flex items-center gap-2">
                  <User className="h-4 w-4 text-muted-foreground" />
                  Имя
                </Label>
                {isEditing ? (
                  <Input
                    id="name"
                    value={editedName}
                    onChange={(e) => setEditedName(e.target.value)}
                    placeholder="Введите ваше имя"
                  />
                ) : (
                  <p className="text-sm text-foreground bg-muted p-3 rounded-md">{user.name}</p>
                )}
              </div>

              {/* Email Field */}
              <div className="space-y-2">
                <Label htmlFor="email" className="flex items-center gap-2">
                  <Mail className="h-4 w-4 text-muted-foreground" />
                  Email
                </Label>
                {isEditing ? (
                  <Input
                    id="email"
                    type="email"
                    value={editedEmail}
                    onChange={(e) => setEditedEmail(e.target.value)}
                    placeholder="Введите ваш email"
                  />
                ) : (
                  <p className="text-sm text-foreground bg-muted p-3 rounded-md">{user.email}</p>
                )}
              </div>

              {/* Role Field (Read-only) */}
              <div className="space-y-2">
                <Label className="flex items-center gap-2">
                  <Shield className="h-4 w-4 text-muted-foreground" />
                  Роль
                </Label>
                <p className="text-sm text-foreground bg-muted p-3 rounded-md">
                  {roleLabels[user.role]}
                </p>
              </div>

              {/* Edit Actions */}
              {isEditing && (
                <div className="flex gap-3 pt-4">
                  <Button onClick={handleSave} className="gap-2">
                    <Check className="h-4 w-4" />
                    Сохранить
                  </Button>
                  <Button onClick={handleCancel} variant="outline" className="gap-2">
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
              <CardTitle>Информация об учётной записи</CardTitle>
              <CardDescription>Дополнительные данные вашего аккаунта</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Created At */}
              <div className="flex items-center justify-between py-2 border-b">
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Calendar className="h-4 w-4" />
                  <span>Дата создания</span>
                </div>
                <p className="text-sm font-medium">{formatDate(user.createdAt)}</p>
              </div>

              {/* Updated At */}
              <div className="flex items-center justify-between py-2 border-b">
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Calendar className="h-4 w-4" />
                  <span>Последнее обновление</span>
                </div>
                <p className="text-sm font-medium">{formatDate(user.updatedAt)}</p>
              </div>

              {/* User ID */}
              <div className="flex items-center justify-between py-2">
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <User className="h-4 w-4" />
                  <span>ID пользователя</span>
                </div>
                <p className="text-xs font-mono bg-muted px-2 py-1 rounded">{user.id}</p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

export default ProfilePage
