'use client'

import { useCallback, useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import { Upload, X, FileText, AlertCircle, Tag, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DocumentCategory,
  ALLOWED_FILE_TYPES,
  ALLOWED_FILE_EXTENSIONS,
  MAX_FILE_SIZE,
  type DocumentUpload,
} from '@/types/document'
import { tagsApi, TagInfo } from '@/lib/api/documents'

interface DocumentUploadProps {
  onUpload: (uploads: DocumentUpload[]) => Promise<void>
  onCancel?: () => void
  isUploading?: boolean
  className?: string
}

interface FileWithPreview {
  file: File
  preview?: string
  error?: string
}

export function DocumentUploadComponent({
  onUpload,
  onCancel,
  isUploading = false,
  className = '',
}: DocumentUploadProps) {
  const t = useTranslations('documents.uploadForm')
  const tDocs = useTranslations('documents')
  const tCommon = useTranslations('common')
  const [files, setFiles] = useState<FileWithPreview[]>([])
  const [isDragging, setIsDragging] = useState(false)
  const [category, setCategory] = useState<DocumentCategory>(DocumentCategory.EDUCATIONAL)
  const [description, setDescription] = useState('')
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([])
  const [availableTags, setAvailableTags] = useState<TagInfo[]>([])
  const [isLoadingTags, setIsLoadingTags] = useState(false)

  /* c8 ignore start - Tags loading effect, tested in e2e */
  // Load available tags on mount
  useEffect(() => {
    const loadTags = async () => {
      setIsLoadingTags(true)
      try {
        const tags = await tagsApi.getAll()
        setAvailableTags(tags)
      } catch (err) {
        console.error('Failed to load tags:', err)
      } finally {
        setIsLoadingTags(false)
      }
    }
    loadTags()
  }, [])
  /* c8 ignore stop */

  const validateFile = useCallback(
    (file: File): string | null => {
      if (!ALLOWED_FILE_TYPES.includes(file.type)) {
        const ext = '.' + file.name.split('.').pop()?.toLowerCase()
        if (!ALLOWED_FILE_EXTENSIONS.includes(ext)) {
          return t('typeNotSupported', { extensions: ALLOWED_FILE_EXTENSIONS.join(', ') })
        }
      }

      if (file.size > MAX_FILE_SIZE) {
        return t('sizeExceeded', { size: String(MAX_FILE_SIZE / 1024 / 1024) })
      }

      return null
    },
    [t]
  )

  const handleFiles = useCallback(
    (newFiles: FileList | File[]) => {
      const fileArray = Array.from(newFiles)
      const validatedFiles: FileWithPreview[] = fileArray.map((file) => {
        const error = validateFile(file)
        return { file, error: error || undefined }
      })

      setFiles((prev) => [...prev, ...validatedFiles])
    },
    [validateFile]
  )

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      e.stopPropagation()
      setIsDragging(false)

      if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
        handleFiles(e.dataTransfer.files)
      }
    },
    [handleFiles]
  )

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)
  }, [])

  const handleFileInput = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (e.target.files && e.target.files.length > 0) {
        handleFiles(e.target.files)
      }
    },
    [handleFiles]
  )

  /* c8 ignore start - File handlers, tested in e2e */
  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index))
  }

  const handleSubmit = async () => {
    const validFiles = files.filter((f) => !f.error)
    if (validFiles.length === 0) return

    // Get tag names from selected IDs
    const selectedTagNames = selectedTagIds
      .map((id) => availableTags.find((t) => t.id === id)?.name)
      .filter(Boolean) as string[]

    const uploads: DocumentUpload[] = validFiles.map(({ file }) => ({
      file,
      category,
      description: description || undefined,
      tags: selectedTagNames.length > 0 ? selectedTagNames : undefined,
    }))

    await onUpload(uploads)
    setFiles([])
    setDescription('')
    setSelectedTagIds([])
  }

  const toggleTag = (tagId: number) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId]
    )
  }
  /* c8 ignore stop */

  const validFilesCount = files.filter((f) => !f.error).length
  const hasErrors = files.some((f) => f.error)

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Drag & Drop Zone */}
      <div
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        className={`
          relative border-2 border-dashed rounded-lg p-12 text-center transition-all
          ${
            isDragging
              ? 'border-blue-500 bg-blue-50 dark:bg-blue-950/20'
              : 'border-gray-300 dark:border-gray-700 hover:border-gray-400 dark:hover:border-gray-600'
          }
          ${isUploading ? 'opacity-50 pointer-events-none' : ''}
        `}
      >
        <input
          type="file"
          multiple
          onChange={handleFileInput}
          className="hidden"
          id="file-upload"
          accept={ALLOWED_FILE_TYPES.join(',')}
          disabled={isUploading}
        />

        <Upload className="mx-auto h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
          {t('dragAndDrop')}
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">{t('orClickButton')}</p>
        <Button
          variant="outline"
          className="cursor-pointer"
          disabled={isUploading}
          onClick={() => document.getElementById('file-upload')?.click()}
        >
          {t('selectFiles')}
        </Button>
        <p className="text-xs text-gray-500 dark:text-gray-500 mt-4">{t('supportedFormats')}</p>
      </div>

      {/* File List */}
      {files.length > 0 && (
        <div className="space-y-4">
          <h4 className="font-semibold text-gray-900 dark:text-white">
            {t('selectedFiles')} ({files.length})
          </h4>
          <div className="space-y-2">
            {files.map((fileWithPreview, index) => (
              <div
                key={index}
                className={`
                  flex items-center justify-between p-3 rounded-lg border
                  ${
                    fileWithPreview.error
                      ? 'border-red-300 bg-red-50 dark:bg-red-950/20 dark:border-red-800'
                      : 'border-gray-200 bg-white dark:bg-gray-900 dark:border-gray-700'
                  }
                `}
              >
                <div className="flex items-center gap-3 flex-1 min-w-0">
                  {fileWithPreview.error ? (
                    <AlertCircle className="h-5 w-5 text-red-500 flex-shrink-0" />
                  ) : (
                    <FileText className="h-5 w-5 text-gray-400 flex-shrink-0" />
                  )}
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                      {fileWithPreview.file.name}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      {tCommon('fileSize.kb', {
                        size: (fileWithPreview.file.size / 1024).toFixed(2),
                      })}
                    </p>
                    {fileWithPreview.error && (
                      <p className="text-xs text-red-600 dark:text-red-400 mt-1">
                        {fileWithPreview.error}
                      </p>
                    )}
                  </div>
                </div>
                <button
                  onClick={() => removeFile(index)}
                  className="p-1 hover:bg-gray-100 dark:hover:bg-gray-800 rounded"
                  disabled={isUploading}
                >
                  <X className="h-4 w-4 text-gray-500" />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Upload Options */}
      {files.length > 0 && (
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {t('documentCategory')}
            </label>
            <select
              value={category}
              onChange={(e) => setCategory(e.target.value as DocumentCategory)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                       bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              disabled={isUploading}
            >
              {Object.values(DocumentCategory).map((value) => (
                <option key={value} value={value}>
                  {tDocs(`categories.${value}`)}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {t('descriptionOptional')}
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                       bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent
                       resize-none"
              rows={3}
              placeholder={t('descriptionPlaceholder')}
              disabled={isUploading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              <Tag className="h-4 w-4 inline-block mr-1" />
              {t('tagsOptional')}
            </label>
            {isLoadingTags ? (
              <p className="text-sm text-gray-500">{t('loadingTags')}</p>
            ) : availableTags.length === 0 ? (
              <p className="text-sm text-gray-500">{t('noTagsAvailable')}</p>
            ) : (
              <div className="flex flex-wrap gap-2">
                {availableTags.map((tag) => {
                  const isSelected = selectedTagIds.includes(tag.id)
                  return (
                    <button
                      key={tag.id}
                      type="button"
                      onClick={() => toggleTag(tag.id)}
                      disabled={isUploading}
                      className={`
                        inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium
                        transition-all duration-150 border
                        ${
                          /* c8 ignore next 3 - Tag selection styling */
                          isSelected
                            ? 'bg-blue-100 dark:bg-blue-900/40 text-blue-800 dark:text-blue-300 border-blue-300 dark:border-blue-700'
                            : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 border-gray-300 dark:border-gray-600 hover:bg-gray-200 dark:hover:bg-gray-700'
                        }
                        ${/* c8 ignore next */ isUploading ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
                      `}
                      style={
                        tag.color && isSelected
                          ? {
                              backgroundColor: `${tag.color}20`,
                              color: tag.color,
                              borderColor: tag.color,
                            }
                          : {}
                      }
                    >
                      {isSelected && <Check className="h-3.5 w-3.5" />}
                      {tag.name}
                    </button>
                  )
                })}
              </div>
            )}
            {selectedTagIds.length > 0 && (
              <p className="text-xs text-gray-500 mt-2">
                {t('selectedTags')}: {selectedTagIds.length}
              </p>
            )}
          </div>

          {/* Action Buttons */}
          <div className="flex gap-3 justify-end">
            {onCancel && (
              <Button variant="outline" onClick={onCancel} disabled={isUploading}>
                {tCommon('cancel')}
              </Button>
            )}
            <Button
              onClick={handleSubmit}
              disabled={isUploading || validFilesCount === 0 || hasErrors}
            >
              {/* c8 ignore next */}
              {isUploading ? t('uploading') : t('uploadCount', { count: validFilesCount })}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
