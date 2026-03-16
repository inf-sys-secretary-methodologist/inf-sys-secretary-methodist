'use client'

import { useState, useRef, useCallback, useEffect } from 'react'

interface UseSpeechSynthesisOptions {
  lang?: string
  preferredVoiceURI?: string
}

interface UseSpeechSynthesisReturn {
  isSupported: boolean
  isSpeaking: boolean
  voices: SpeechSynthesisVoice[]
  speak: (text: string, voice?: SpeechSynthesisVoice) => void
  pause: () => void
  resume: () => void
  cancel: () => void
}

function stripMarkdown(text: string): string {
  return text
    .replace(/```[\s\S]*?```/g, '')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/#{1,6}\s+/g, '')
    .replace(/\*{3}(.+?)\*{3}/g, '$1')
    .replace(/\*{2}(.+?)\*{2}/g, '$1')
    .replace(/\*(.+?)\*/g, '$1')
    .replace(/_{2}(.+?)_{2}/g, '$1')
    .replace(/_(.+?)_/g, '$1')
    .replace(/~~(.+?)~~/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/^[-*+]\s+/gm, '')
    .replace(/^\d+\.\s+/gm, '')
    .replace(/^>\s+/gm, '')
    .replace(/\n{2,}/g, '. ')
    .trim()
}

export function useSpeechSynthesis({
  lang = 'ru-RU',
  preferredVoiceURI = '',
}: UseSpeechSynthesisOptions = {}): UseSpeechSynthesisReturn {
  const [isSupported, setIsSupported] = useState(false)
  const [isSpeaking, setIsSpeaking] = useState(false)
  const [voices, setVoices] = useState<SpeechSynthesisVoice[]>([])

  const utteranceRef = useRef<SpeechSynthesisUtterance | null>(null)
  const keepAliveRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const clearKeepAlive = useCallback(() => {
    if (keepAliveRef.current) {
      clearInterval(keepAliveRef.current)
      keepAliveRef.current = null
    }
  }, [])

  // Check support on client only to avoid SSR hydration mismatch
  useEffect(() => {
    setIsSupported('speechSynthesis' in window)
  }, [])

  // Load voices
  useEffect(() => {
    if (!isSupported) return

    const loadVoices = () => {
      const allVoices = speechSynthesis.getVoices()
      const langPrefix = lang.split('-')[0]
      const filtered = allVoices.filter(
        (v) => v.lang.startsWith(langPrefix) || v.lang.startsWith(lang)
      )
      setVoices(filtered.length > 0 ? filtered : allVoices)
    }

    loadVoices()
    speechSynthesis.addEventListener('voiceschanged', loadVoices)
    return () => {
      speechSynthesis.removeEventListener('voiceschanged', loadVoices)
    }
  }, [isSupported, lang])

  const speak = useCallback(
    (text: string, voice?: SpeechSynthesisVoice) => {
      if (!isSupported) return

      speechSynthesis.cancel()
      clearKeepAlive()

      const cleanText = stripMarkdown(text)
      if (!cleanText) return

      const utterance = new SpeechSynthesisUtterance(cleanText)
      utterance.lang = lang

      if (voice) {
        utterance.voice = voice
      } else if (preferredVoiceURI) {
        const preferred = voices.find((v) => v.voiceURI === preferredVoiceURI)
        if (preferred) utterance.voice = preferred
      } else if (voices.length > 0) {
        utterance.voice = voices[0]
      }

      utterance.onstart = () => {
        setIsSpeaking(true)
        // Chrome bug: long utterances stop after ~15s without keepalive
        keepAliveRef.current = setInterval(() => {
          speechSynthesis.pause()
          speechSynthesis.resume()
        }, 10000)
      }
      utterance.onend = () => {
        clearKeepAlive()
        setIsSpeaking(false)
      }
      utterance.onerror = () => {
        clearKeepAlive()
        setIsSpeaking(false)
      }
      utterance.onpause = () => setIsSpeaking(false)
      utterance.onresume = () => setIsSpeaking(true)

      utteranceRef.current = utterance
      speechSynthesis.speak(utterance)
    },
    [isSupported, lang, voices, preferredVoiceURI, clearKeepAlive]
  )

  const pause = useCallback(() => {
    if (isSupported) speechSynthesis.pause()
  }, [isSupported])

  const resume = useCallback(() => {
    if (isSupported) speechSynthesis.resume()
  }, [isSupported])

  const cancel = useCallback(() => {
    if (isSupported) {
      clearKeepAlive()
      speechSynthesis.cancel()
      setIsSpeaking(false)
    }
  }, [isSupported, clearKeepAlive])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      clearKeepAlive()
      if (isSupported) {
        speechSynthesis.cancel()
      }
    }
  }, [isSupported, clearKeepAlive])

  return {
    isSupported,
    isSpeaking,
    voices,
    speak,
    pause,
    resume,
    cancel,
  }
}
