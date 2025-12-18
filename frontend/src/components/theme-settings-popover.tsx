'use client'

import * as React from 'react'
import { Moon, Sun, Monitor, Palette, Sparkles } from 'lucide-react'
import { useTheme } from 'next-themes'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { cn } from '@/lib/utils'
import { useAppearanceStore, type BackgroundType } from '@/stores/appearanceStore'

const themeOptions = [
  { value: 'light', label: 'Светлая', icon: Sun },
  { value: 'dark', label: 'Тёмная', icon: Moon },
  { value: 'system', label: 'Системная', icon: Monitor },
] as const

const backgroundOptions: { value: BackgroundType; label: string }[] = [
  { value: 'none', label: 'Без фона' },
  { value: 'grain-gradient', label: 'Градиент' },
  { value: 'warp', label: 'Искажение' },
  { value: 'mesh-gradient', label: 'Сетка' },
]

export function ThemeSettingsPopover() {
  const { theme, setTheme } = useTheme()
  const { background, setBackgroundType, setBackgroundEnabled } = useAppearanceStore()
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    return (
      <div
        className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border-2 border-gray-300 bg-white dark:border-gray-700 dark:bg-gray-900 shadow-md opacity-50"
        aria-hidden="true"
      >
        <Palette className="h-5 w-5 text-muted-foreground" />
      </div>
    )
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border-2 border-gray-300 bg-white text-gray-900 transition-all duration-200 hover:bg-gray-100 hover:scale-105 active:scale-95 dark:border-gray-700 dark:bg-gray-900 dark:text-white dark:hover:bg-gray-800 shadow-md hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          aria-label="Настройки темы"
          type="button"
        >
          <Palette className="h-5 w-5" />
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-72 p-4" align="end">
        <div className="space-y-4">
          {/* Theme Selection */}
          <div className="space-y-2">
            <Label className="text-sm font-medium flex items-center gap-2">
              {theme === 'dark' ? (
                <Moon className="h-4 w-4" />
              ) : theme === 'light' ? (
                <Sun className="h-4 w-4" />
              ) : (
                <Monitor className="h-4 w-4" />
              )}
              Тема оформления
            </Label>
            <div className="grid grid-cols-3 gap-2">
              {themeOptions.map((option) => {
                const Icon = option.icon
                const isActive = theme === option.value
                return (
                  <Button
                    key={option.value}
                    variant={isActive ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setTheme(option.value)}
                    className={cn(
                      'flex flex-col items-center gap-1 h-auto py-2',
                      isActive && 'ring-2 ring-primary ring-offset-2'
                    )}
                  >
                    <Icon className="h-4 w-4" />
                    <span className="text-xs">{option.label}</span>
                  </Button>
                )
              })}
            </div>
          </div>

          {/* Background Selection */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="text-sm font-medium flex items-center gap-2">
                <Sparkles className="h-4 w-4" />
                Анимированный фон
              </Label>
              <Switch
                checked={background.enabled}
                onCheckedChange={setBackgroundEnabled}
                aria-label="Включить анимированный фон"
              />
            </div>
            {background.enabled && (
              <div className="grid grid-cols-2 gap-2">
                {backgroundOptions.map((option) => {
                  const isActive = background.type === option.value
                  return (
                    <Button
                      key={option.value}
                      variant={isActive ? 'default' : 'outline'}
                      size="sm"
                      onClick={() => setBackgroundType(option.value)}
                      className={cn('text-xs h-8', isActive && 'ring-2 ring-primary ring-offset-1')}
                    >
                      {option.label}
                    </Button>
                  )
                })}
              </div>
            )}
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
}
