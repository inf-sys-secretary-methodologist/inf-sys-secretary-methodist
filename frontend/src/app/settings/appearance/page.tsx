'use client'

import { useEffect, useState } from 'react'
import { useTheme } from 'next-themes'
import { Palette, Sun, Moon, Monitor, Sparkles, RotateCcw, Eye, Zap } from 'lucide-react'
import { toast } from 'sonner'
import { AppLayout } from '@/components/layout'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Slider } from '@/components/ui/slider'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { useAppearanceStore, type BackgroundType } from '@/stores/appearanceStore'
import { cn } from '@/lib/utils'

const backgroundTypes: { value: BackgroundType; label: string; description: string }[] = [
  { value: 'none', label: 'Без фона', description: 'Стандартный однотонный фон' },
  {
    value: 'grain-gradient',
    label: 'Градиент с зернистостью',
    description: 'Мягкий градиент с эффектом зернистости',
  },
  { value: 'warp', label: 'Искажение', description: 'Динамичный волнистый эффект' },
  {
    value: 'mesh-gradient',
    label: 'Сетчатый градиент',
    description: 'Плавный многоцветный градиент',
  },
]

const themeOptions = [
  { value: 'light', label: 'Светлая', icon: Sun },
  { value: 'dark', label: 'Тёмная', icon: Moon },
  { value: 'system', label: 'Системная', icon: Monitor },
]

export default function AppearanceSettingsPage() {
  const { theme, setTheme } = useTheme()
  const {
    background,
    reducedMotion,
    setBackgroundType,
    setBackgroundEnabled,
    setBackgroundSpeed,
    setBackgroundIntensity,
    setReducedMotion,
    resetToDefaults,
  } = useAppearanceStore()

  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  const handleReset = () => {
    resetToDefaults()
    setTheme('system')
    toast.success('Настройки внешнего вида сброшены')
  }

  if (!mounted) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
        </div>
      </AppLayout>
    )
  }

  return (
    <AppLayout>
      <div className="max-w-2xl mx-auto space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Внешний вид</h1>
            <p className="text-muted-foreground">Настройте тему и анимированный фон</p>
          </div>
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="outline" size="sm">
                <RotateCcw className="h-4 w-4 mr-2" />
                Сбросить
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Сбросить настройки?</AlertDialogTitle>
                <AlertDialogDescription>
                  Все настройки внешнего вида будут сброшены к значениям по умолчанию.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Отмена</AlertDialogCancel>
                <AlertDialogAction onClick={handleReset}>Сбросить</AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>

        {/* Theme Selection */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Palette className="h-5 w-5" />
              Тема оформления
            </CardTitle>
            <CardDescription>Выберите цветовую тему интерфейса</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 gap-3">
              {themeOptions.map((option) => {
                const Icon = option.icon
                const isSelected = theme === option.value
                return (
                  <button
                    key={option.value}
                    onClick={() => setTheme(option.value)}
                    className={cn(
                      'flex flex-col items-center gap-2 p-4 rounded-lg border-2 transition-all',
                      isSelected
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50 hover:bg-muted/50'
                    )}
                  >
                    <Icon
                      className={cn(
                        'h-6 w-6',
                        isSelected ? 'text-primary' : 'text-muted-foreground'
                      )}
                    />
                    <span
                      className={cn(
                        'text-sm font-medium',
                        isSelected ? 'text-primary' : 'text-muted-foreground'
                      )}
                    >
                      {option.label}
                    </span>
                  </button>
                )
              })}
            </div>
          </CardContent>
        </Card>

        {/* Background Settings */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <Sparkles className="h-5 w-5" />
                  Анимированный фон
                </CardTitle>
                <CardDescription>Настройте динамический фон приложения</CardDescription>
              </div>
              <Switch checked={background.enabled} onCheckedChange={setBackgroundEnabled} />
            </div>
          </CardHeader>
          {background.enabled && (
            <CardContent className="space-y-6">
              {/* Background Type */}
              <div className="space-y-2">
                <Label className="flex items-center gap-2">
                  <Eye className="h-4 w-4" />
                  Тип фона
                </Label>
                <Select
                  value={background.type}
                  onValueChange={(value) => setBackgroundType(value as BackgroundType)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Выберите тип фона" />
                  </SelectTrigger>
                  <SelectContent>
                    {backgroundTypes.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        <div className="flex flex-col">
                          <span>{type.label}</span>
                          <span className="text-xs text-muted-foreground">{type.description}</span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {background.type !== 'none' && (
                <>
                  {/* Speed Control */}
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <Label className="flex items-center gap-2">
                        <Zap className="h-4 w-4" />
                        Скорость анимации
                      </Label>
                      <span className="text-sm text-muted-foreground">
                        {background.speed.toFixed(1)}x
                      </span>
                    </div>
                    <Slider
                      value={[background.speed]}
                      onValueChange={([value]) => setBackgroundSpeed(value)}
                      min={0.1}
                      max={2}
                      step={0.1}
                      disabled={reducedMotion}
                    />
                    <p className="text-xs text-muted-foreground">
                      {reducedMotion
                        ? 'Отключено из-за режима уменьшения движения'
                        : 'Скорость воспроизведения анимации'}
                    </p>
                  </div>

                  {/* Intensity Control */}
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <Label>Интенсивность</Label>
                      <span className="text-sm text-muted-foreground">
                        {Math.round(background.intensity * 100)}%
                      </span>
                    </div>
                    <Slider
                      value={[background.intensity]}
                      onValueChange={([value]) => setBackgroundIntensity(value)}
                      min={0.1}
                      max={1}
                      step={0.05}
                    />
                    <p className="text-xs text-muted-foreground">Яркость и насыщенность эффекта</p>
                  </div>
                </>
              )}
            </CardContent>
          )}
        </Card>

        {/* Accessibility */}
        <Card>
          <CardHeader>
            <CardTitle>Доступность</CardTitle>
            <CardDescription>Настройки для улучшения удобства использования</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Уменьшить движение</Label>
                <p className="text-sm text-muted-foreground">
                  Отключает анимации для людей с вестибулярными нарушениями
                </p>
              </div>
              <Switch checked={reducedMotion} onCheckedChange={setReducedMotion} />
            </div>
          </CardContent>
        </Card>

        {/* Preview Info */}
        <Card className="border-dashed">
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground text-center">
              Изменения применяются автоматически. Посмотрите на фон страницы, чтобы увидеть
              результат.
            </p>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  )
}
