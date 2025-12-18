'use client'

import { MeshGradient } from '@paper-design/shaders-react'

interface MeshGradientBackgroundProps {
  isDark: boolean
  speed: number
  intensity: number
}

// Color presets for light and dark themes
const darkThemeColors = [
  'hsl(250, 50%, 20%)',
  'hsl(280, 45%, 18%)',
  'hsl(220, 55%, 15%)',
  'hsl(260, 50%, 22%)',
]

const lightThemeColors = [
  'hsl(210, 25%, 95%)',
  'hsl(220, 20%, 92%)',
  'hsl(200, 25%, 94%)',
  'hsl(230, 15%, 96%)',
]

export function MeshGradientBackground({ isDark, speed, intensity }: MeshGradientBackgroundProps) {
  const colors = isDark ? darkThemeColors : lightThemeColors

  return (
    <MeshGradient
      style={{ width: '100%', height: '100%' }}
      colors={colors}
      speed={speed}
      distortion={intensity}
    />
  )
}
