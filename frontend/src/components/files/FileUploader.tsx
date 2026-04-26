'use client'

import { useCallback, useRef, useState } from 'react'
import { useTranslations } from 'next-intl'
import { Upload } from 'lucide-react'
import { cn } from '@/lib/utils'

interface FileUploaderProps {
  onUpload: (file: File) => Promise<void> | void
  uploading?: boolean
  className?: string
}

export function FileUploader({ onUpload, uploading = false, className }: FileUploaderProps) {
  const t = useTranslations('files')
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragOver, setDragOver] = useState(false)

  const handleFile = useCallback(
    (file: File) => {
      if (uploading) return
      onUpload(file)
    },
    [onUpload, uploading]
  )

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      setDragOver(false)
      const file = e.dataTransfer.files[0]
      if (file) handleFile(file)
    },
    [handleFile]
  )

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(true)
  }, [])

  const handleDragLeave = useCallback(() => {
    setDragOver(false)
  }, [])

  const handleInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0]
      if (file) handleFile(file)
      if (inputRef.current) inputRef.current.value = ''
    },
    [handleFile]
  )

  return (
    <label
      htmlFor="file-uploader-input"
      onDrop={handleDrop}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      className={cn(
        'flex flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed p-8 text-center transition-colors cursor-pointer',
        dragOver
          ? 'border-primary bg-primary/5'
          : 'border-muted-foreground/25 hover:border-primary/50',
        uploading && 'pointer-events-none opacity-60',
        className
      )}
    >
      <Upload className="h-8 w-8 text-muted-foreground" />
      {uploading ? (
        <p className="text-sm text-muted-foreground">{t('dropzone.uploading')}</p>
      ) : (
        <>
          <p className="text-sm font-medium">{t('dropzone.title')}</p>
          <p className="text-xs text-muted-foreground">{t('dropzone.subtitle')}</p>
        </>
      )}
      <input
        ref={inputRef}
        id="file-uploader-input"
        type="file"
        data-testid="file-upload-input"
        onChange={handleInputChange}
        disabled={uploading}
        className="sr-only"
      />
    </label>
  )
}
