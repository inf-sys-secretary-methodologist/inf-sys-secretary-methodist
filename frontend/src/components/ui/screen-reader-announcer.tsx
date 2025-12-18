'use client'

import { useEffect, useState, useCallback, createContext, useContext, ReactNode } from 'react'

interface AnnouncerContextType {
  announce: (message: string, priority?: 'polite' | 'assertive') => void
}

const AnnouncerContext = createContext<AnnouncerContextType | null>(null)

/**
 * Hook to announce messages to screen readers
 */
export function useAnnouncer() {
  const context = useContext(AnnouncerContext)
  if (!context) {
    throw new Error('useAnnouncer must be used within ScreenReaderAnnouncerProvider')
  }
  return context
}

interface ScreenReaderAnnouncerProviderProps {
  children: ReactNode
}

/**
 * Provider for screen reader announcements.
 * Uses aria-live regions to announce dynamic content changes.
 * WCAG 4.1.3: Status Messages
 */
export function ScreenReaderAnnouncerProvider({ children }: ScreenReaderAnnouncerProviderProps) {
  const [politeMessage, setPoliteMessage] = useState('')
  const [assertiveMessage, setAssertiveMessage] = useState('')

  const announce = useCallback((message: string, priority: 'polite' | 'assertive' = 'polite') => {
    if (priority === 'assertive') {
      setAssertiveMessage('')
      // Small delay to ensure screen reader picks up the change
      setTimeout(() => setAssertiveMessage(message), 100)
    } else {
      setPoliteMessage('')
      setTimeout(() => setPoliteMessage(message), 100)
    }
  }, [])

  // Clear messages after announcement
  useEffect(() => {
    if (politeMessage) {
      const timer = setTimeout(() => setPoliteMessage(''), 1000)
      return () => clearTimeout(timer)
    }
  }, [politeMessage])

  useEffect(() => {
    if (assertiveMessage) {
      const timer = setTimeout(() => setAssertiveMessage(''), 1000)
      return () => clearTimeout(timer)
    }
  }, [assertiveMessage])

  return (
    <AnnouncerContext.Provider value={{ announce }}>
      {children}
      {/* Polite announcements - waits for current speech to finish */}
      <div role="status" aria-live="polite" aria-atomic="true" className="sr-only">
        {politeMessage}
      </div>
      {/* Assertive announcements - interrupts current speech */}
      <div role="alert" aria-live="assertive" aria-atomic="true" className="sr-only">
        {assertiveMessage}
      </div>
    </AnnouncerContext.Provider>
  )
}
