'use client'

import { GrainGradient } from '@paper-design/shaders-react'

interface GrainGradientBackgroundProps {
  isDark: boolean
  speed: number
  intensity: number
}

// Color presets for light and dark themes
const darkThemeColors = {
  colorBack: 'hsl(222, 47%, 6%)', // Very dark blue-gray matching dark theme
  colors: ['hsl(250, 70%, 35%)', 'hsl(280, 60%, 30%)', 'hsl(220, 70%, 25%)'], // Muted purples/blues
}

const lightThemeColors = {
  colorBack: 'hsl(0, 0%, 100%)', // Pure white
  colors: ['hsl(210, 40%, 90%)', 'hsl(220, 30%, 85%)', 'hsl(200, 35%, 88%)'], // Very light pastels
}

export function GrainGradientBackground({
  isDark,
  speed,
  intensity,
}: GrainGradientBackgroundProps) {
  const theme = isDark ? darkThemeColors : lightThemeColors

  return (
    <GrainGradient
      style={{ height: '100%', width: '100%' }}
      colorBack={theme.colorBack}
      colors={theme.colors}
      softness={0.8}
      intensity={intensity}
      noise={0.03}
      shape="corners"
      offsetX={0}
      offsetY={0}
      scale={1}
      rotation={0}
      speed={speed}
    />
  )
}
