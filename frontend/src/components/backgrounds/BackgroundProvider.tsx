'use client'

import { useEffect, useState } from 'react'
import { useTheme } from 'next-themes'
import { useAppearanceStore } from '@/stores/appearanceStore'
import { GrainGradientBackground } from './GrainGradientBackground'
import { WarpBackground } from './WarpBackground'
import { MeshGradientBackground } from './MeshGradientBackground'

export function BackgroundProvider() {
  const { resolvedTheme } = useTheme()
  const background = useAppearanceStore((state) => state.background)
  const reducedMotion = useAppearanceStore((state) => state.reducedMotion)
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    // Small delay to ensure zustand has hydrated from localStorage
    const timer = setTimeout(() => setMounted(true), 50)
    return () => clearTimeout(timer)
  }, [])

  // Don't render anything until mounted to avoid hydration mismatch
  if (!mounted) {
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
