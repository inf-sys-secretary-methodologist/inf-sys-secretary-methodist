'use client'

import { useState, useEffect, useCallback } from 'react'

const STORAGE_KEY = 'favorite_templates'
const RECENT_KEY = 'recent_templates'
const MAX_RECENT = 5

interface UseFavoriteTemplatesReturn {
  favorites: number[]
  recentlyUsed: number[]
  isFavorite: (id: number) => boolean
  toggleFavorite: (id: number) => void
  addToRecent: (id: number) => void
  clearRecent: () => void
}

export function useFavoriteTemplates(): UseFavoriteTemplatesReturn {
  const [favorites, setFavorites] = useState<number[]>([])
  const [recentlyUsed, setRecentlyUsed] = useState<number[]>([])

  // Load from localStorage on mount
  useEffect(() => {
    try {
      const savedFavorites = localStorage.getItem(STORAGE_KEY)
      if (savedFavorites) {
        setFavorites(JSON.parse(savedFavorites))
      }

      const savedRecent = localStorage.getItem(RECENT_KEY)
      if (savedRecent) {
        setRecentlyUsed(JSON.parse(savedRecent))
      }
    } catch (error) {
      console.error('Failed to load favorites from localStorage:', error)
    }
  }, [])

  // Save favorites to localStorage
  const saveFavorites = useCallback((newFavorites: number[]) => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newFavorites))
    } catch (error) {
      console.error('Failed to save favorites to localStorage:', error)
    }
  }, [])

  // Save recent to localStorage
  const saveRecent = useCallback((newRecent: number[]) => {
    try {
      localStorage.setItem(RECENT_KEY, JSON.stringify(newRecent))
    } catch (error) {
      console.error('Failed to save recent to localStorage:', error)
    }
  }, [])

  const isFavorite = useCallback(
    (id: number) => {
      return favorites.includes(id)
    },
    [favorites]
  )

  const toggleFavorite = useCallback(
    (id: number) => {
      setFavorites((prev) => {
        const newFavorites = prev.includes(id) ? prev.filter((fId) => fId !== id) : [...prev, id]
        saveFavorites(newFavorites)
        return newFavorites
      })
    },
    [saveFavorites]
  )

  const addToRecent = useCallback(
    (id: number) => {
      setRecentlyUsed((prev) => {
        // Remove if already exists, add to front
        const filtered = prev.filter((rId) => rId !== id)
        const newRecent = [id, ...filtered].slice(0, MAX_RECENT)
        saveRecent(newRecent)
        return newRecent
      })
    },
    [saveRecent]
  )

  const clearRecent = useCallback(() => {
    setRecentlyUsed([])
    try {
      localStorage.removeItem(RECENT_KEY)
    } catch (error) {
      console.error('Failed to clear recent from localStorage:', error)
    }
  }, [])

  return {
    favorites,
    recentlyUsed,
    isFavorite,
    toggleFavorite,
    addToRecent,
    clearRecent,
  }
}
