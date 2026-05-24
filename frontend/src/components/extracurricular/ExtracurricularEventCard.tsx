'use client'

// Stub — replaced in GREEN. Tests assert rendering of title, status
// badge, category badge, audience badge, date, location, capacity,
// and dropdown actions.

import type { ExtracurricularEventSummary } from '@/types/extracurricular'

export interface ExtracurricularEventCardProps {
  event: ExtracurricularEventSummary
  onClick?: () => void
  onEdit?: () => void
  onDelete?: () => void
  onRegister?: () => void
  onUnregister?: () => void
  onPublish?: () => void
  onCancel?: () => void
  onComplete?: () => void
  isRegistered?: boolean
  className?: string
}

export function ExtracurricularEventCard(_props: ExtracurricularEventCardProps) {
  void _props
  return null
}
