'use client'

import { WifiOff, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'

export default function OfflinePage() {
  const handleRetry = () => {
    window.location.reload()
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-background p-4">
      <div className="text-center space-y-6 max-w-md">
        <div className="flex justify-center">
          <div className="rounded-full bg-muted p-6">
            <WifiOff className="h-12 w-12 text-muted-foreground" />
          </div>
        </div>

        <div className="space-y-2">
          <h1 className="text-2xl font-bold tracking-tight">Нет подключения к интернету</h1>
          <p className="text-muted-foreground">
            Проверьте подключение к интернету и попробуйте снова. Некоторые функции могут быть
            недоступны в офлайн-режиме.
          </p>
        </div>

        <div className="space-y-3">
          <Button onClick={handleRetry} className="w-full">
            <RefreshCw className="mr-2 h-4 w-4" />
            Попробовать снова
          </Button>

          <p className="text-xs text-muted-foreground">
            Вы можете продолжить работу с ранее загруженными данными в кэше.
          </p>
        </div>

        <div className="pt-4 border-t">
          <p className="text-sm text-muted-foreground">Советы для работы офлайн:</p>
          <ul className="mt-2 text-sm text-muted-foreground text-left list-disc list-inside space-y-1">
            <li>Ранее просмотренные страницы доступны из кэша</li>
            <li>Изменения будут синхронизированы при восстановлении связи</li>
            <li>Уведомления придут после подключения</li>
          </ul>
        </div>
      </div>
    </div>
  )
}
