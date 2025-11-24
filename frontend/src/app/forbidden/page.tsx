'use client'

import { useRouter } from 'next/navigation'
import { ShieldX, ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'

export default function ForbiddenPage() {
  const router = useRouter()

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <div className="text-center space-y-6 max-w-md">
        {/* Icon */}
        <div className="flex justify-center">
          <div className="rounded-full bg-destructive/10 p-6">
            <ShieldX className="h-16 w-16 text-destructive" />
          </div>
        </div>

        {/* Title */}
        <div className="space-y-2">
          <h1 className="text-4xl font-bold tracking-tight">403</h1>
          <h2 className="text-2xl font-semibold">Доступ запрещен</h2>
        </div>

        {/* Description */}
        <p className="text-muted-foreground">
          У вас нет прав доступа к этой странице. Пожалуйста, свяжитесь с администратором, если вы
          считаете, что это ошибка.
        </p>

        {/* Actions */}
        <div className="flex flex-col sm:flex-row gap-3 justify-center">
          <Button onClick={() => router.back()} variant="outline" className="gap-2">
            <ArrowLeft className="h-4 w-4" />
            Назад
          </Button>
          <Button onClick={() => router.push('/dashboard')}>На главную</Button>
        </div>
      </div>
    </div>
  )
}
