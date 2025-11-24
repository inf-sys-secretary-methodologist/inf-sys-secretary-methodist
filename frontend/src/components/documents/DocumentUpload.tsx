'use client'

import { useCallback, useState } from 'react'
import { Upload, X, FileText, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DocumentCategory,
  DocumentCategoryLabels,
  ALLOWED_FILE_TYPES,
  ALLOWED_FILE_EXTENSIONS,
  MAX_FILE_SIZE,
  type DocumentUpload,
} from '@/types/document'

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
  const [files, setFiles] = useState<FileWithPreview[]>([])
  const [isDragging, setIsDragging] = useState(false)
  const [category, setCategory] = useState<DocumentCategory>(DocumentCategory.OTHER)
  const [description, setDescription] = useState('')
  const [tags, setTags] = useState('')

  const validateFile = (file: File): string | null => {
    if (!ALLOWED_FILE_TYPES.includes(file.type)) {
      const ext = '.' + file.name.split('.').pop()?.toLowerCase()
      if (!ALLOWED_FILE_EXTENSIONS.includes(ext)) {
        return `Тип файла не поддерживается. Разрешены: ${ALLOWED_FILE_EXTENSIONS.join(', ')}`
      }
    }

    if (file.size > MAX_FILE_SIZE) {
      return `Размер файла превышает максимальный (${MAX_FILE_SIZE / 1024 / 1024}МБ)`
    }

    return null
  }

  const handleFiles = useCallback((newFiles: FileList | File[]) => {
    const fileArray = Array.from(newFiles)
    const validatedFiles: FileWithPreview[] = fileArray.map((file) => {
      const error = validateFile(file)
      return { file, error: error || undefined }
    })

    setFiles((prev) => [...prev, ...validatedFiles])
  }, [])

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

  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index))
  }

  const handleSubmit = async () => {
    const validFiles = files.filter((f) => !f.error)
    if (validFiles.length === 0) return

    const uploads: DocumentUpload[] = validFiles.map(({ file }) => ({
      file,
      category,
      description: description || undefined,
      tags: tags
        ? tags
            .split(',')
            .map((t) => t.trim())
            .filter(Boolean)
        : undefined,
    }))

    await onUpload(uploads)
    setFiles([])
    setDescription('')
    setTags('')
  }

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
          Перетащите файлы сюда
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          или нажмите кнопку ниже для выбора файлов
        </p>
        <label htmlFor="file-upload">
          <Button variant="outline" className="cursor-pointer" disabled={isUploading}>
            Выбрать файлы
          </Button>
        </label>
        <p className="text-xs text-gray-500 dark:text-gray-500 mt-4">
          Поддерживаемые форматы: PDF, DOC, DOCX, XLS, XLSX, TXT, JPG, PNG (макс. 10МБ)
        </p>
      </div>

      {/* File List */}
      {files.length > 0 && (
        <div className="space-y-4">
          <h4 className="font-semibold text-gray-900 dark:text-white">
            Выбранные файлы ({files.length})
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
                      {(fileWithPreview.file.size / 1024).toFixed(2)} КБ
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
              Категория документа
            </label>
            <select
              value={category}
              onChange={(e) => setCategory(e.target.value as DocumentCategory)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                       bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              disabled={isUploading}
            >
              {Object.entries(DocumentCategoryLabels).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Описание (опционально)
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                       bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent
                       resize-none"
              rows={3}
              placeholder="Краткое описание документов..."
              disabled={isUploading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Теги (опционально)
            </label>
            <input
              type="text"
              value={tags}
              onChange={(e) => setTags(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                       bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Введите теги через запятую..."
              disabled={isUploading}
            />
          </div>

          {/* Action Buttons */}
          <div className="flex gap-3 justify-end">
            {onCancel && (
              <Button variant="outline" onClick={onCancel} disabled={isUploading}>
                Отмена
              </Button>
            )}
            <Button
              onClick={handleSubmit}
              disabled={isUploading || validFilesCount === 0 || hasErrors}
            >
              {isUploading ? 'Загрузка...' : `Загрузить (${validFilesCount})`}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
