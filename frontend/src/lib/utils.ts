import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Validates and returns an avatar URL.
 * Returns undefined if the URL is invalid (not a full URL).
 * This prevents 404 errors when avatar paths are stored without presigned URLs.
 */
export function getValidAvatarUrl(avatarUrl: string | null | undefined): string | undefined {
  if (!avatarUrl) return undefined
  // Check if it's a valid full URL (http/https or data URI)
  if (
    avatarUrl.startsWith('http://') ||
    avatarUrl.startsWith('https://') ||
    avatarUrl.startsWith('data:')
  ) {
    return avatarUrl
  }
  // Return undefined for relative paths to show avatar fallback instead of 404
  return undefined
}
