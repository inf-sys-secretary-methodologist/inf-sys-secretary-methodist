'use client'

import { useState, useEffect, useRef } from 'react'
import { X, Save, Loader2, Upload, FileText, Trash2, Tag, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Document } from '@/types/document'
import { documentsApi, UpdateDocumentParams, tagsApi, TagInfo } from '@/lib/api/documents'

interface DocumentEditDialogProps {
  document: Document | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSaved?: () => void
}

export function DocumentEditDialog({
  document,
  open,
  onOpenChange,
  onSaved,
}: DocumentEditDialogProps) {
  const [title, setTitle] = useState('')
  const [subject, setSubject] = useState('')
  const [content, setContent] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isUploading, setIsUploading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [currentFileName, setCurrentFileName] = useState<string | null>(null)
  const [newFile, setNewFile] = useState<File | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Tags state
  const [documentTags, setDocumentTags] = useState<TagInfo[]>([])
  const [allTags, setAllTags] = useState<TagInfo[]>([])
  const [tagSearch, setTagSearch] = useState('')
  const [showTagDropdown, setShowTagDropdown] = useState(false)

  // Reset form when document changes
  useEffect(() => {
    if (document) {
      setTitle(document.name || '')
      setSubject(document.description || '')
      setContent('')
      setError(null)
      setSuccess(null)
      setNewFile(null)
      setDocumentTags([])
      setTagSearch('')
      setShowTagDropdown(false)
    }
  }, [document])

  // Load full document data including content
  useEffect(() => {
    if (document && open) {
      loadDocumentDetails()
    }
  }, [document, open])

  const loadDocumentDetails = async () => {
    if (!document) return
    try {
      const [fullDoc, tags, availableTags] = await Promise.all([
        documentsApi.getById(document.id),
        tagsApi.getDocumentTags(document.id),
        tagsApi.getAll(),
      ])
      setTitle(fullDoc.title || '')
      setSubject(fullDoc.subject || '')
      setContent(fullDoc.content || '')
      setCurrentFileName(fullDoc.file_name || null)
      setDocumentTags(tags)
      setAllTags(availableTags)
    } catch (err) {
      console.error('Failed to load document details:', err)
    }
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      setNewFile(file)
      setError(null)
    }
  }

  const handleFileUpload = async () => {
    if (!document || !newFile) return

    setIsUploading(true)
    setError(null)
    setSuccess(null)

    try {
      await documentsApi.uploadFile(document.id, newFile)
      setSuccess(`Файл "${newFile.name}" успешно загружен`)
      setCurrentFileName(newFile.name)
      setNewFile(null)
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    } catch (err) {
      console.error('Failed to upload file:', err)
      setError('Не удалось загрузить файл')
    } finally {
      setIsUploading(false)
    }
  }

  // Tag management functions
  const handleAddTag = async (tag: TagInfo) => {
    if (!document) return
    if (documentTags.some((t) => t.id === tag.id)) return

    try {
      await tagsApi.addTagToDocument(document.id, tag.id)
      setDocumentTags([...documentTags, tag])
      setTagSearch('')
      setShowTagDropdown(false)
    } catch (err) {
      console.error('Failed to add tag:', err)
      setError('Не удалось добавить тег')
    }
  }

  const handleRemoveTag = async (tagId: number) => {
    if (!document) return

    try {
      await tagsApi.removeTagFromDocument(document.id, tagId)
      setDocumentTags(documentTags.filter((t) => t.id !== tagId))
    } catch (err) {
      console.error('Failed to remove tag:', err)
      setError('Не удалось удалить тег')
    }
  }

  const filteredTags = allTags.filter(
    (tag) =>
      tag.name.toLowerCase().includes(tagSearch.toLowerCase()) &&
      !documentTags.some((t) => t.id === tag.id)
  )

  const handleSave = async () => {
    if (!document) return

    setIsLoading(true)
    setError(null)

    try {
      const params: UpdateDocumentParams = {
        title: title.trim(),
        subject: subject.trim() || undefined,
        content: content.trim() || undefined,
        file_name: currentFileName?.trim() || undefined,
      }

      await documentsApi.update(document.id, params)
      onOpenChange(false)
      onSaved?.()
    } catch (err) {
      console.error('Failed to update document:', err)
      setError('Не удалось сохранить изменения')
    } finally {
      setIsLoading(false)
    }
  }

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} Б`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} КБ`
    return `${(bytes / 1024 / 1024).toFixed(1)} МБ`
  }

  if (!open || !document) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="relative w-full max-w-2xl m-4 bg-white dark:bg-gray-900 rounded-lg shadow-2xl max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700 sticky top-0 bg-white dark:bg-gray-900">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
            Редактирование документа
          </h2>
          <button
            onClick={() => onOpenChange(false)}
            className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
          >
            <X className="h-5 w-5 text-gray-500" />
          </button>
        </div>

        {/* Form */}
        <div className="p-4 space-y-4">
          {error && (
            <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-700 dark:text-red-400 text-sm">
              {error}
            </div>
          )}

          {success && (
            <div className="p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg text-green-700 dark:text-green-400 text-sm">
              {success}
            </div>
          )}

          {/* File Upload Section */}
          <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg bg-gray-50 dark:bg-gray-800/50">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
              Прикреплённый файл
            </label>

            {currentFileName && (
              <div className="flex items-center gap-2 mb-3">
                <FileText className="h-5 w-5 text-blue-500 flex-shrink-0" />
                <input
                  type="text"
                  value={currentFileName}
                  onChange={(e) => setCurrentFileName(e.target.value)}
                  className="flex-1 px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg
                           bg-white dark:bg-gray-800 text-gray-900 dark:text-white
                           focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="Название файла"
                />
              </div>
            )}

            <div className="flex items-center gap-2">
              <input
                ref={fileInputRef}
                type="file"
                onChange={handleFileSelect}
                className="hidden"
                id="file-upload"
              />
              <label
                htmlFor="file-upload"
                className="flex items-center gap-2 px-3 py-2 text-sm border border-gray-300 dark:border-gray-600
                         rounded-lg cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors
                         text-gray-700 dark:text-gray-300"
              >
                <Upload className="h-4 w-4" />
                Выбрать новый файл
              </label>

              {newFile && (
                <>
                  <span className="text-sm text-gray-600 dark:text-gray-400 truncate max-w-[150px]">
                    {newFile.name} ({formatFileSize(newFile.size)})
                  </span>
                  <Button size="sm" onClick={handleFileUpload} disabled={isUploading}>
                    {isUploading ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Загрузить'}
                  </Button>
                  <button
                    onClick={() => {
                      setNewFile(null)
                      if (fileInputRef.current) fileInputRef.current.value = ''
                    }}
                    className="p-1 text-gray-400 hover:text-red-500"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </>
              )}
            </div>

            <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
              При загрузке нового файла создаётся новая версия документа
            </p>
          </div>

          {/* Metadata Fields */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Заголовок *
            </label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                       bg-white dark:bg-gray-800 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Название документа"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Тема / Описание
            </label>
            <input
              type="text"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                       bg-white dark:bg-gray-800 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Краткое описание документа"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Содержание
            </label>
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={6}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                       bg-white dark:bg-gray-800 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
              placeholder="Текстовое содержание документа (опционально)"
            />
          </div>

          {/* Tags Section */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              <Tag className="h-4 w-4 inline-block mr-1" />
              Теги
            </label>

            {/* Current tags */}
            <div className="flex flex-wrap gap-2 mb-3">
              {documentTags.map((tag) => (
                <span
                  key={tag.id}
                  className="inline-flex items-center gap-1 px-3 py-1 bg-blue-100 dark:bg-blue-900/30
                           text-blue-800 dark:text-blue-300 rounded-full text-sm"
                  style={tag.color ? { backgroundColor: `${tag.color}20`, color: tag.color } : {}}
                >
                  {tag.name}
                  <button
                    onClick={() => handleRemoveTag(tag.id)}
                    className="ml-1 hover:bg-blue-200 dark:hover:bg-blue-800 rounded-full p-0.5"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </span>
              ))}
              {documentTags.length === 0 && (
                <span className="text-sm text-gray-500 dark:text-gray-400">Нет тегов</span>
              )}
            </div>

            {/* Add tag input */}
            <div className="relative">
              <div className="flex gap-2">
                <input
                  type="text"
                  value={tagSearch}
                  onChange={(e) => {
                    setTagSearch(e.target.value)
                    setShowTagDropdown(true)
                  }}
                  onFocus={() => setShowTagDropdown(true)}
                  className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                           bg-white dark:bg-gray-800 text-gray-900 dark:text-white
                           focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
                  placeholder="Поиск тегов..."
                />
              </div>

              {/* Dropdown with available tags */}
              {showTagDropdown && filteredTags.length > 0 && (
                <div
                  className="absolute z-10 w-full mt-1 bg-white dark:bg-gray-800 border border-gray-200
                              dark:border-gray-700 rounded-lg shadow-lg max-h-48 overflow-y-auto"
                >
                  {filteredTags.map((tag) => (
                    <button
                      key={tag.id}
                      onClick={() => handleAddTag(tag)}
                      className="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700
                               flex items-center gap-2"
                    >
                      <Plus className="h-4 w-4 text-gray-400" />
                      <span
                        className="px-2 py-0.5 rounded text-sm"
                        style={
                          tag.color ? { backgroundColor: `${tag.color}20`, color: tag.color } : {}
                        }
                      >
                        {tag.name}
                      </span>
                    </button>
                  ))}
                </div>
              )}

              {showTagDropdown && filteredTags.length === 0 && tagSearch && (
                <div
                  className="absolute z-10 w-full mt-1 bg-white dark:bg-gray-800 border border-gray-200
                              dark:border-gray-700 rounded-lg shadow-lg p-3 text-sm text-gray-500"
                >
                  Тег не найден
                </div>
              )}
            </div>

            {/* Click outside to close dropdown */}
            {showTagDropdown && (
              <div className="fixed inset-0 z-0" onClick={() => setShowTagDropdown(false)} />
            )}
          </div>

          <p className="text-xs text-gray-500 dark:text-gray-400">
            При сохранении изменений автоматически создается новая версия документа
          </p>
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-3 p-4 border-t border-gray-200 dark:border-gray-700 sticky bottom-0 bg-white dark:bg-gray-900">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isLoading}>
            Отмена
          </Button>
          <Button onClick={handleSave} disabled={isLoading || !title.trim()}>
            {isLoading ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Сохранение...
              </>
            ) : (
              <>
                <Save className="h-4 w-4 mr-2" />
                Сохранить
              </>
            )}
          </Button>
        </div>
      </div>
    </div>
  )
}
