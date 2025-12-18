'use client'

import { Warp } from '@paper-design/shaders-react'

interface WarpBackgroundProps {
  isDark: boolean
  speed: number
  intensity: number
}

// Color presets for light and dark themes
const darkThemeColors = [
  'hsl(250, 50%, 25%)',
  'hsl(280, 45%, 20%)',
  'hsl(220, 55%, 22%)',
  'hsl(260, 50%, 28%)',
]

const lightThemeColors = [
  'hsl(210, 30%, 92%)',
  'hsl(220, 25%, 88%)',
  'hsl(200, 30%, 90%)',
  'hsl(230, 20%, 94%)',
]

export function WarpBackground({ isDark, speed, intensity }: WarpBackgroundProps) {
  const colors = isDark ? darkThemeColors : lightThemeColors

  return (
    <Warp
      style={{ width: '100%', height: '100%' }}
      proportion={0.45}
      softness={1}
      distortion={intensity * 0.5}
      swirl={0.6}
      swirlIterations={8}
      shape="checks"
      shapeScale={0.15}
      scale={1}
      rotation={0}
      speed={speed}
      colors={colors}
    />
  )
}
