'use client'

import { useEffect, useState } from 'react'
import { useTheme } from 'next-themes'
import dynamic from 'next/dynamic'
import { useAppearanceStore, useAppearanceHydrated } from '@/stores/appearanceStore'

const GrainGradientBackground = dynamic(
  () => import('./GrainGradientBackground').then((mod) => mod.GrainGradientBackground),
  { ssr: false, loading: () => null }
)

const WarpBackground = dynamic(() => import('./WarpBackground').then((mod) => mod.WarpBackground), {
  ssr: false,
  loading: () => null,
})

const MeshGradientBackground = dynamic(
  () => import('./MeshGradientBackground').then((mod) => mod.MeshGradientBackground),
  { ssr: false, loading: () => null }
)

export function BackgroundProvider() {
  const { resolvedTheme } = useTheme()
  const background = useAppearanceStore((state) => state.background)
  const reducedMotion = useAppearanceStore((state) => state.reducedMotion)
  const hasHydrated = useAppearanceHydrated()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  // Don't render anything until mounted AND hydrated to avoid hydration mismatch
  if (!mounted || !hasHydrated) {
    return null
  }

  // If background is disabled or type is 'none', don't render
  if (!background.enabled || background.type === 'none') {
    return null
  }

  const isDark = resolvedTheme === 'dark'
  const effectiveSpeed = reducedMotion ? 0 : background.speed

  const commonProps = {
    isDark,
    speed: effectiveSpeed,
    intensity: background.intensity,
  }

  return (
    <div className="fixed inset-0 -z-10 pointer-events-none">
      {background.type === 'grain-gradient' && <GrainGradientBackground {...commonProps} />}
      {background.type === 'warp' && <WarpBackground {...commonProps} />}
      {background.type === 'mesh-gradient' && <MeshGradientBackground {...commonProps} />}
    </div>
  )
}
