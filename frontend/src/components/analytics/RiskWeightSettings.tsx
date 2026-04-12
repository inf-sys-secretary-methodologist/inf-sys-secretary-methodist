'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import { Save } from 'lucide-react'
import { analyticsApi, RiskWeightConfig } from '@/lib/api/analytics'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Slider } from '@/components/ui/slider'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'

export function RiskWeightSettings() {
  const t = useTranslations('analytics')
  const [config, setConfig] = useState<RiskWeightConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const cfg = await analyticsApi.getRiskWeightConfig()
        setConfig(cfg)
      } catch {
        toast.error(t('weights.loadError'))
      } finally {
        setLoading(false)
      }
    }
    fetchConfig()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const handleWeightChange = (key: keyof RiskWeightConfig, value: number) => {
    if (!config) return
    setConfig({ ...config, [key]: value })
  }

  const handleSave = async () => {
    if (!config) return
    setSaving(true)
    try {
      await analyticsApi.updateRiskWeightConfig({
        attendance_weight: config.attendance_weight,
        grade_weight: config.grade_weight,
        submission_weight: config.submission_weight,
        inactivity_weight: config.inactivity_weight,
        high_risk_threshold: config.high_risk_threshold,
        critical_risk_threshold: config.critical_risk_threshold,
      })
      toast.success(t('weights.saveSuccess'))
    } catch {
      toast.error(t('weights.saveError'))
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <Card>
        <CardHeader><Skeleton className="h-6 w-48" /></CardHeader>
        <CardContent className="space-y-4">
          {[1, 2, 3, 4].map(i => <Skeleton key={i} className="h-12 w-full" />)}
        </CardContent>
      </Card>
    )
  }

  if (!config) return null

  const totalWeight = config.attendance_weight + config.grade_weight + config.submission_weight + config.inactivity_weight
  const isValid = Math.abs(totalWeight - 1.0) < 0.02

  const weights = [
    { key: 'attendance_weight' as const, label: t('weights.attendance'), color: 'bg-blue-500' },
    { key: 'grade_weight' as const, label: t('weights.grades'), color: 'bg-green-500' },
    { key: 'submission_weight' as const, label: t('weights.submissions'), color: 'bg-yellow-500' },
    { key: 'inactivity_weight' as const, label: t('weights.inactivity'), color: 'bg-red-500' },
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('weights.title')}</CardTitle>
        <CardDescription>{t('weights.description')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {weights.map(({ key, label, color }) => (
          <div key={key} className="space-y-2">
            <div className="flex items-center justify-between">
              <Label>{label}</Label>
              <span className="text-sm font-mono text-muted-foreground">
                {(config[key] * 100).toFixed(0)}%
              </span>
            </div>
            <div className="flex items-center gap-3">
              <div className={`h-2 w-2 rounded-full ${color}`} />
              <Slider
                value={[config[key] * 100]}
                min={0}
                max={100}
                step={5}
                onValueChange={([v]) => handleWeightChange(key, v / 100)}
                className="flex-1"
              />
            </div>
          </div>
        ))}

        <div className="flex items-center justify-between border-t pt-4">
          <div className="text-sm">
            <span className={isValid ? 'text-green-600' : 'text-destructive font-medium'}>
              {t('weights.total')}: {(totalWeight * 100).toFixed(0)}%
            </span>
            {!isValid && (
              <span className="text-destructive ml-2">({t('weights.mustBe100')})</span>
            )}
          </div>
          <Button onClick={handleSave} disabled={saving || !isValid} size="sm">
            <Save className="mr-2 h-4 w-4" />
            {saving ? t('weights.saving') : t('weights.save')}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
