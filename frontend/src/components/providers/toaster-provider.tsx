'use client'

import { Toaster } from 'sonner'

export function ToasterProvider() {
  return (
    <Toaster
      position="top-right"
      richColors
      closeButton
      expand={false}
      duration={4000}
      theme="system"
    />
  )
}

// Re-export toast from the same module instance
export { toast } from 'sonner'
