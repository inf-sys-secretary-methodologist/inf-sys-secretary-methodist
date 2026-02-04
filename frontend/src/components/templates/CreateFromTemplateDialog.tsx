'use client'

import { useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import { Loader2, FileText, CheckCircle } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { templatesApi, TemplateInfo, TemplateVariable } from '@/lib/api/templates'

interface CreateFromTemplateDialogProps {
  template: TemplateInfo | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateFromTemplateDialog({
  template,
  open,
  onOpenChange,
}: CreateFromTemplateDialogProps) {
  const t = useTranslations('templates')
  const tCommon = useTranslations('common')
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [isPreviewLoading, setIsPreviewLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [documentTitle, setDocumentTitle] = useState('')
  const [variables, setVariables] = useState<Record<string, string>>({})
  const [previewContent, setPreviewContent] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  // Initialize variables from template
  useEffect(() => {
    if (template?.template_variables) {
      const initialVars: Record<string, string> = {}
      template.template_variables.forEach((v) => {
        initialVars[v.name] = v.default_value || ''
      })
      setVariables(initialVars)
    }
    setDocumentTitle('')
    setPreviewContent(null)
    setError(null)
    setSuccess(false)
  }, [template])

  const handleVariableChange = (name: string, value: string) => {
    setVariables((prev) => ({ ...prev, [name]: value }))
    setPreviewContent(null) // Clear preview when variables change
  }

  const handlePreview = async () => {
    if (!template) return

    try {
      setIsPreviewLoading(true)
      setError(null)
      const content = await templatesApi.preview(template.id, variables)
      setPreviewContent(content)
    } catch (err) {
      console.error('Preview failed:', err)
      setError(t('previewError'))
    } finally {
      setIsPreviewLoading(false)
    }
  }

  const handleCreate = async () => {
    if (!template || !documentTitle.trim()) return

    try {
      setIsLoading(true)
      setError(null)
      await templatesApi.createDocument(template.id, {
        title: documentTitle.trim(),
        variables,
      })
      setSuccess(true)
      setTimeout(() => {
        onOpenChange(false)
        router.push(`/documents`)
      }, 1500)
    } catch (err) {
      console.error('Create failed:', err)
      setError(t('createError'))
    } finally {
      setIsLoading(false)
    }
  }

  const renderVariableInput = (variable: TemplateVariable) => {
    const value = variables[variable.name] || ''

    if (variable.variable_type === 'select' && variable.options) {
      return (
        <select
          id={variable.name}
          value={value}
          onChange={(e) => handleVariableChange(variable.name, e.target.value)}
          className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <option value="">{t('selectOption')}</option>
          {variable.options.map((opt) => (
            <option key={opt} value={opt}>
              {opt}
            </option>
          ))}
        </select>
      )
    }

    if (variable.variable_type === 'date') {
      return (
        <Input
          id={variable.name}
          type="date"
          value={value}
          onChange={(e) => handleVariableChange(variable.name, e.target.value)}
        />
      )
    }

    if (variable.variable_type === 'number') {
      return (
        <Input
          id={variable.name}
          type="number"
          value={value}
          onChange={(e) => handleVariableChange(variable.name, e.target.value)}
        />
      )
    }

    return (
      <Input
        id={variable.name}
        type="text"
        value={value}
        onChange={(e) => handleVariableChange(variable.name, e.target.value)}
        placeholder={variable.description || ''}
      />
    )
  }

  if (success) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-lg">
          <div className="flex flex-col items-center justify-center py-8">
            <CheckCircle className="h-16 w-16 text-green-500 mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
              {t('createSuccess')}
            </h2>
            <p className="text-gray-600 dark:text-gray-400 mt-2">{t('redirecting')}</p>
          </div>
        </DialogContent>
      </Dialog>
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            {t('createFromTemplate')}
          </DialogTitle>
          <DialogDescription>
            {template?.name} - {template?.description}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Document Title */}
          <div className="space-y-2">
            <Label htmlFor="title">{t('documentTitle')}</Label>
            <Input
              id="title"
              value={documentTitle}
              onChange={(e) => setDocumentTitle(e.target.value)}
              placeholder={t('enterTitle')}
            />
          </div>

          {/* Template Variables */}
          {template?.template_variables && template.template_variables.length > 0 && (
            <div className="space-y-4">
              <h3 className="text-sm font-medium text-gray-900 dark:text-white">
                {t('fillVariables')}
              </h3>
              {template.template_variables.map((variable) => (
                <div key={variable.name} className="space-y-2">
                  <Label htmlFor={variable.name} className="flex items-center gap-1">
                    {variable.name}
                    {variable.required && <span className="text-red-500">*</span>}
                  </Label>
                  {variable.description && (
                    <p className="text-xs text-gray-500">{variable.description}</p>
                  )}
                  {renderVariableInput(variable)}
                </div>
              ))}
            </div>
          )}

          {/* Preview */}
          {previewContent && (
            <div className="space-y-2">
              <h3 className="text-sm font-medium text-gray-900 dark:text-white">
                {t('previewTitle')}
              </h3>
              <div className="rounded-lg border bg-gray-50 dark:bg-gray-900 p-4 max-h-60 overflow-y-auto">
                <pre className="whitespace-pre-wrap text-sm text-gray-700 dark:text-gray-300">
                  {previewContent}
                </pre>
              </div>
            </div>
          )}

          {/* Error */}
          {error && (
            <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 p-3">
              <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
            </div>
          )}
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {tCommon('cancel')}
          </Button>
          <Button variant="outline" onClick={handlePreview} disabled={isPreviewLoading}>
            {isPreviewLoading && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
            {t('preview')}
          </Button>
          <Button
            onClick={handleCreate}
            disabled={isLoading || !documentTitle.trim()}
            className="bg-blue-600 hover:bg-blue-700 text-white"
          >
            {isLoading && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
            {t('createDocument')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
