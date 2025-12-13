'use client'

import { useState, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { User, ImagePlus, X, Loader2 } from 'lucide-react'
import Image from 'next/image'
import { ImageCropper } from './ImageCropper'

interface AvatarUploadProps {
  currentAvatar?: string | null
  userName?: string
  onUpload: (file: File) => Promise<void>
  onRemove?: () => Promise<void>
  disabled?: boolean
  className?: string
}

export function AvatarUpload({
  currentAvatar,
  userName,
  onUpload,
  onRemove,
  disabled = false,
  className,
}: AvatarUploadProps) {
  const [isDragging, setIsDragging] = useState(false)
  const [preview, setPreview] = useState<string | null>(currentAvatar || null)
  const [isUploading, setIsUploading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [cropperOpen, setCropperOpen] = useState(false)
  const [imageToCrop, setImageToCrop] = useState<string | null>(null)
  const [originalFileName, setOriginalFileName] = useState<string>('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Update preview when currentAvatar prop changes (e.g., after page refresh)
  useEffect(() => {
    if (currentAvatar) {
      setPreview(currentAvatar)
    }
  }, [currentAvatar])

  const MAX_SIZE = 5 * 1024 * 1024 // 5MB
  const ACCEPTED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']

  const validateFile = (file: File): string | null => {
    if (file.size > MAX_SIZE) {
      return 'Размер файла должен быть менее 5MB'
    }
    if (!ACCEPTED_TYPES.includes(file.type)) {
      return 'Поддерживаются только JPG, PNG, GIF, WebP'
    }
    return null
  }

  const handleFile = async (file: File) => {
    setError(null)

    const validationError = validateFile(file)
    if (validationError) {
      setError(validationError)
      return
    }

    // Store original file name for later use
    setOriginalFileName(file.name)

    // Create data URL and open cropper
    const reader = new FileReader()
    reader.onloadend = () => {
      setImageToCrop(reader.result as string)
      setCropperOpen(true)
    }
    reader.readAsDataURL(file)
  }

  const handleCropComplete = async (croppedBlob: Blob) => {
    setCropperOpen(false)
    setImageToCrop(null)

    // Create preview from cropped blob
    const previewUrl = URL.createObjectURL(croppedBlob)
    setPreview(previewUrl)

    // Convert blob to File for upload
    const croppedFile = new File([croppedBlob], originalFileName || 'avatar.jpg', {
      type: 'image/jpeg',
    })

    // Upload
    setIsUploading(true)
    try {
      await onUpload(croppedFile)
    } catch {
      setError('Ошибка при загрузке файла')
      setPreview(currentAvatar || null)
    } finally {
      setIsUploading(false)
    }
  }

  const handleCropCancel = () => {
    setCropperOpen(false)
    setImageToCrop(null)
    // Reset file input
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
    if (disabled || isUploading) return

    const droppedFile = e.dataTransfer.files[0]
    if (droppedFile) {
      handleFile(droppedFile)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0]
    if (selectedFile) {
      handleFile(selectedFile)
    }
  }

  const handleRemove = async (e: React.MouseEvent) => {
    e.stopPropagation()
    if (!onRemove || isUploading) return

    setIsUploading(true)
    try {
      await onRemove()
      setPreview(null)
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    } catch {
      setError('Ошибка при удалении фото')
    } finally {
      setIsUploading(false)
    }
  }

  const getInitials = (name?: string) => {
    if (!name) return ''
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  return (
    <div className={cn('space-y-3', className)}>
      <div className="flex items-center gap-4">
        {/* Avatar Preview / Placeholder */}
        <div
          className={cn(
            'relative size-20 rounded-full overflow-hidden transition-all duration-300',
            'border-2 border-dashed',
            isDragging
              ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
              : 'border-border bg-muted',
            !disabled && !isUploading && 'cursor-pointer hover:border-blue-400',
            disabled && 'opacity-50 cursor-not-allowed'
          )}
          onClick={() => !disabled && !isUploading && fileInputRef.current?.click()}
          onDragOver={(e) => {
            e.preventDefault()
            if (!disabled && !isUploading) setIsDragging(true)
          }}
          onDragLeave={() => setIsDragging(false)}
          onDrop={handleDrop}
        >
          {isUploading ? (
            <div className="absolute inset-0 flex items-center justify-center bg-background/80">
              <Loader2 className="size-6 animate-spin text-blue-500" />
            </div>
          ) : preview ? (
            <Image
              src={preview}
              alt={userName || 'Аватар'}
              fill
              className="object-cover"
              sizes="80px"
              unoptimized
            />
          ) : userName ? (
            <div className="absolute inset-0 flex items-center justify-center bg-gold-100 dark:bg-gold-900/40 text-gold-700 dark:text-gold-400 text-xl font-semibold">
              {getInitials(userName)}
            </div>
          ) : (
            <div className="absolute inset-0 flex items-center justify-center">
              <User className="size-8 text-muted-foreground" />
            </div>
          )}
        </div>

        {/* Action buttons */}
        <div className="flex flex-col gap-2">
          <button
            type="button"
            onClick={() => fileInputRef.current?.click()}
            disabled={disabled || isUploading}
            className={cn(
              'px-4 py-2 text-sm font-medium rounded-lg transition-colors',
              'bg-blue-600 text-white hover:bg-blue-700',
              'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
              'disabled:opacity-50 disabled:cursor-not-allowed'
            )}
          >
            <span className="flex items-center gap-2">
              <ImagePlus className="size-4" />
              {preview ? 'Изменить фото' : 'Загрузить фото'}
            </span>
          </button>

          {preview && onRemove && (
            <button
              type="button"
              onClick={handleRemove}
              disabled={disabled || isUploading}
              className={cn(
                'px-4 py-2 text-sm font-medium rounded-lg transition-colors',
                'text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20',
                'focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2',
                'disabled:opacity-50 disabled:cursor-not-allowed'
              )}
            >
              <span className="flex items-center gap-2">
                <X className="size-4" />
                Удалить
              </span>
            </button>
          )}
        </div>
      </div>

      {/* Hints */}
      <p className="text-xs text-muted-foreground">JPG, PNG, GIF или WebP. Максимум 5MB.</p>

      {/* Error */}
      {error && <p className="text-xs text-red-500">{error}</p>}

      {/* Hidden input */}
      <input
        ref={fileInputRef}
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp"
        onChange={handleChange}
        disabled={disabled || isUploading}
        className="hidden"
      />

      {/* Image Cropper Dialog */}
      {imageToCrop && (
        <ImageCropper
          image={imageToCrop}
          open={cropperOpen}
          onCropComplete={handleCropComplete}
          onCancel={handleCropCancel}
          aspectRatio={1}
        />
      )}
    </div>
  )
}
