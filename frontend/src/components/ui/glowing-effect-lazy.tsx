'use client'

import dynamic from 'next/dynamic'

// Lazy load GlowingEffect to reduce initial bundle (motion ~100KB)
// This wrapper provides the same interface but loads asynchronously
export const GlowingEffect = dynamic(
  () => import('./glowing-effect').then((mod) => mod.GlowingEffect),
  {
    ssr: false,
    loading: () => null, // No visible placeholder needed for visual effects
  }
)
