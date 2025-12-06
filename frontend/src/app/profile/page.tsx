'use client'

import { useState } from 'react'
import { useAuthCheck } from '@/hooks/useAuth'
import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { User, Mail, Calendar, Shield, Edit2, X, Check } from 'lucide-react'
import { UserRole } from '@/types/auth'

const roleLabels: Record<UserRole, string> = {
  [UserRole.SYSTEM_ADMIN]: 'Системный администратор',
  [UserRole.METHODIST]: 'Методист',
  [UserRole.ACADEMIC_SECRETARY]: 'Учёный секретарь',
  [UserRole.TEACHER]: 'Преподаватель',
  [UserRole.STUDENT]: 'Студент',
}

function ProfilePage() {
  const { user } = useAuthCheck()
  const [isEditing, setIsEditing] = useState(false)
  const [editedName, setEditedName] = useState(user?.name || '')
  const [editedEmail, setEditedEmail] = useState(user?.email || '')

  if (!user) {
    return null
  }

  const handleSave = async () => {
    // TODO: Implement API call to update user profile
    console.log('Saving profile:', { name: editedName, email: editedEmail })
    setIsEditing(false)
  }

  const handleCancel = () => {
    setEditedName(user.name)
    setEditedEmail(user.email)
    setIsEditing(false)
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

        {/* Profile Information Card */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg sm:text-xl">Персональная информация</CardTitle>
            <CardDescription>Ваши основные данные учётной записи</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 sm:space-y-6">
            {/* Name Field */}
            <div className="space-y-2">
              <Label htmlFor="name" className="flex items-center gap-2 text-sm">
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
              <Label htmlFor="email" className="flex items-center gap-2 text-sm">
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
                <p className="text-sm text-foreground bg-muted p-3 rounded-md break-all">
                  {user.email}
                </p>
              )}
            </div>

            {/* Role Field (Read-only) */}
            <div className="space-y-2">
              <Label className="flex items-center gap-2 text-sm">
                <Shield className="h-4 w-4 text-muted-foreground" />
                Роль
              </Label>
              <p className="text-sm text-foreground bg-muted p-3 rounded-md">
                {roleLabels[user.role]}
              </p>
            </div>

            {/* Edit Actions */}
            {isEditing && (
              <div className="flex flex-col sm:flex-row gap-3 pt-4">
                <Button onClick={handleSave} className="gap-2 w-full sm:w-auto">
                  <Check className="h-4 w-4" />
                  Сохранить
                </Button>
                <Button onClick={handleCancel} variant="outline" className="gap-2 w-full sm:w-auto">
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
              <p className="text-xs font-mono bg-muted px-2 py-1 rounded break-all">{user.id}</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  )
}

export default ProfilePage
