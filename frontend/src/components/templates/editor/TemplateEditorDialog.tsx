'use client'

import { useState, useEffect, useMemo, useCallback } from 'react'
import { useTranslations } from 'next-intl'
import {
  Loader2,
  Save,
  Eye,
  EyeOff,
  Code,
  FileText,
  AlertTriangle,
  CheckCircle,
} from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { templatesApi, TemplateInfo, TemplateVariable } from '@/lib/api/templates'
import { VariablesList } from './VariablesList'
import { AddVariableDialog } from './AddVariableDialog'
import { cn } from '@/lib/utils'

interface TemplateEditorDialogProps {
  template: TemplateInfo | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave?: () => void
}

// Extract variable names from template content
function extractVariables(content: string): string[] {
  const regex = /\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}/g
  const matches = new Set<string>()
  let match
  while ((match = regex.exec(content)) !== null) {
    matches.add(match[1])
  }
  return Array.from(matches)
}

// Highlight variables in content
function highlightContent(content: string, declaredVariables: string[]): React.ReactNode {
  const parts: React.ReactNode[] = []
  const regex = /(\{\{[a-zA-Z_][a-zA-Z0-9_]*\}\})/g
  let lastIndex = 0
  let match

  while ((match = regex.exec(content)) !== null) {
    // Add text before match
    if (match.index > lastIndex) {
      parts.push(content.slice(lastIndex, match.index))
    }

    // Add highlighted variable
    const varName = match[1].slice(2, -2)
    const isDeclared = declaredVariables.includes(varName)
    parts.push(
      <span
        key={match.index}
        className={cn(
          'px-1 py-0.5 rounded font-medium',
          isDeclared
            ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400'
            : 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400'
        )}
      >
        {match[1]}
      </span>
    )

    lastIndex = match.index + match[0].length
  }

  // Add remaining text
  if (lastIndex < content.length) {
    parts.push(content.slice(lastIndex))
  }

  return parts
}

export function TemplateEditorDialog({
  template,
  open,
  onOpenChange,
  onSave,
}: TemplateEditorDialogProps) {
  const t = useTranslations('templates.editor')
  const tCommon = useTranslations('common')
  const [content, setContent] = useState('')
  const [variables, setVariables] = useState<TemplateVariable[]>([])
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)
  const [showPreview, setShowPreview] = useState(true)
  const [activeTab, setActiveTab] = useState<'content' | 'variables'>('content')
  const [addVariableOpen, setAddVariableOpen] = useState(false)
  const [editingVariable, setEditingVariable] = useState<{
    variable: TemplateVariable
    index: number
  } | null>(null)

  // Load template data
  useEffect(() => {
    if (open && template) {
      setContent(template.template_content || '')
      setVariables(template.template_variables || [])
      setError(null)
      setSuccess(false)
    }
  }, [open, template])

  // Extract and analyze variables
  const contentVariables = useMemo(() => extractVariables(content), [content])
  const declaredNames = useMemo(() => variables.map((v) => v.name), [variables])

  const undeclaredVariables = useMemo(
    () => contentVariables.filter((v) => !declaredNames.includes(v)),
    [contentVariables, declaredNames]
  )

  const unusedVariables = useMemo(
    () => declaredNames.filter((v) => !contentVariables.includes(v)),
    [contentVariables, declaredNames]
  )

  const hasChanges = useMemo(() => {
    if (!template) return false
    const originalContent = template.template_content || ''
    const originalVariables = template.template_variables || []
    return (
      content !== originalContent || JSON.stringify(variables) !== JSON.stringify(originalVariables)
    )
  }, [template, content, variables])

  const handleSave = async () => {
    if (!template) return

    try {
      setIsSaving(true)
      setError(null)
      await templatesApi.update(template.id, {
        template_content: content,
        template_variables: variables,
      })
      setSuccess(true)
      setTimeout(() => {
        onSave?.()
        onOpenChange(false)
      }, 1500)
    } catch (err) {
      console.error('Failed to save template:', err)
      setError(t('saveError'))
    } finally {
      setIsSaving(false)
    }
  }

  const handleAddVariable = useCallback((variable: TemplateVariable) => {
    setVariables((prev) => [...prev, variable])
  }, [])

  const handleEditVariable = useCallback((variable: TemplateVariable, index: number) => {
    setEditingVariable({ variable, index })
    setAddVariableOpen(true)
  }, [])

  const handleSaveVariable = useCallback(
    (variable: TemplateVariable) => {
      if (editingVariable) {
        setVariables((prev) => {
          const newVars = [...prev]
          newVars[editingVariable.index] = variable
          return newVars
        })
        setEditingVariable(null)
      } else {
        handleAddVariable(variable)
      }
    },
    [editingVariable, handleAddVariable]
  )

  const insertVariable = (varName: string) => {
    const textarea = document.getElementById('template-content') as HTMLTextAreaElement
    if (textarea) {
      const start = textarea.selectionStart
      const end = textarea.selectionEnd
      const text = content
      const before = text.substring(0, start)
      const after = text.substring(end)
      const newContent = `${before}{{${varName}}}${after}`
      setContent(newContent)
      // Set cursor after inserted variable
      setTimeout(() => {
        textarea.focus()
        const newPos = start + varName.length + 4
        textarea.setSelectionRange(newPos, newPos)
      }, 0)
    }
  }

  if (!template) return null

  if (success) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-lg">
          <div className="flex flex-col items-center justify-center py-8">
            <CheckCircle className="h-16 w-16 text-green-500 mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
              {t('saveSuccess')}
            </h2>
          </div>
        </DialogContent>
      </Dialog>
    )
  }

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-5xl max-h-[90vh] flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Code className="h-5 w-5" />
              {t('editTemplate')}: {template.name}
            </DialogTitle>
            <DialogDescription>
              {t('editDescription')} ({template.code})
            </DialogDescription>
          </DialogHeader>

          {/* Validation warnings */}
          {(undeclaredVariables.length > 0 || unusedVariables.length > 0) && (
            <div className="flex items-center gap-2 p-2 rounded-lg bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800">
              <AlertTriangle className="h-4 w-4 text-amber-600 shrink-0" />
              <span className="text-sm text-amber-800 dark:text-amber-200">
                {undeclaredVariables.length > 0 && (
                  <span>
                    {t('undeclaredCount', { count: undeclaredVariables.length })}
                    {unusedVariables.length > 0 && ' • '}
                  </span>
                )}
                {unusedVariables.length > 0 && (
                  <span>{t('unusedCount', { count: unusedVariables.length })}</span>
                )}
              </span>
            </div>
          )}

          {/* Main content area */}
          <div className="flex-1 overflow-hidden">
            <Tabs
              value={activeTab}
              onValueChange={(v) => setActiveTab(v as 'content' | 'variables')}
              className="h-full flex flex-col"
            >
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="content" className="gap-2">
                  <FileText className="h-4 w-4" />
                  {t('content')}
                </TabsTrigger>
                <TabsTrigger value="variables" className="gap-2">
                  <Code className="h-4 w-4" />
                  {t('variables')}
                  {variables.length > 0 && (
                    <Badge variant="secondary" className="ml-1">
                      {variables.length}
                    </Badge>
                  )}
                </TabsTrigger>
              </TabsList>

              <TabsContent value="content" className="flex-1 mt-4 overflow-hidden">
                <div className={cn('h-full', showPreview ? 'grid grid-cols-2 gap-4' : '')}>
                  {/* Editor */}
                  <div className="flex flex-col h-full">
                    <div className="flex items-center justify-between mb-2">
                      <Label htmlFor="template-content">{t('templateContent')}</Label>
                      <div className="flex items-center gap-2">
                        {/* Quick insert buttons */}
                        {variables.length > 0 && (
                          <div className="flex items-center gap-1">
                            <span className="text-xs text-gray-500 mr-1">{t('insert')}:</span>
                            {variables.slice(0, 3).map((v) => (
                              <Button
                                key={v.name}
                                size="sm"
                                variant="outline"
                                className="h-6 px-2 text-xs"
                                onClick={() => insertVariable(v.name)}
                              >
                                {v.name}
                              </Button>
                            ))}
                            {variables.length > 3 && (
                              <span className="text-xs text-gray-500">+{variables.length - 3}</span>
                            )}
                          </div>
                        )}
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => setShowPreview(!showPreview)}
                          className="h-8"
                        >
                          {showPreview ? (
                            <EyeOff className="h-4 w-4" />
                          ) : (
                            <Eye className="h-4 w-4" />
                          )}
                        </Button>
                      </div>
                    </div>
                    <textarea
                      id="template-content"
                      value={content}
                      onChange={(e) => setContent(e.target.value)}
                      className="flex-1 min-h-[300px] w-full rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-4 text-sm font-mono resize-none focus:outline-none focus:ring-2 focus:ring-primary"
                      placeholder={t('contentPlaceholder')}
                    />
                  </div>

                  {/* Preview */}
                  {showPreview && (
                    <div className="flex flex-col h-full">
                      <Label className="mb-2">{t('preview')}</Label>
                      <div className="flex-1 min-h-[300px] rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/50 p-4 overflow-auto">
                        <pre className="text-sm font-mono whitespace-pre-wrap text-gray-700 dark:text-gray-300">
                          {highlightContent(content, declaredNames)}
                        </pre>
                      </div>
                    </div>
                  )}
                </div>
              </TabsContent>

              <TabsContent value="variables" className="flex-1 mt-4 overflow-auto">
                <VariablesList
                  variables={variables}
                  onChange={setVariables}
                  onAdd={() => {
                    setEditingVariable(null)
                    setAddVariableOpen(true)
                  }}
                  onEdit={handleEditVariable}
                  undeclaredVariables={undeclaredVariables}
                  unusedVariables={unusedVariables}
                />
              </TabsContent>
            </Tabs>
          </div>

          {/* Error */}
          {error && (
            <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 p-3">
              <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
            </div>
          )}

          {/* Footer */}
          <div className="flex items-center justify-between pt-4 border-t">
            <div className="text-sm text-gray-500">
              {hasChanges && <span className="text-amber-600">{t('unsavedChanges')}</span>}
            </div>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                {tCommon('cancel')}
              </Button>
              <Button onClick={handleSave} disabled={isSaving || !hasChanges} className="gap-2">
                {isSaving ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Save className="h-4 w-4" />
                )}
                {tCommon('save')}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Add/Edit Variable Dialog */}
      <AddVariableDialog
        open={addVariableOpen}
        onOpenChange={(open) => {
          setAddVariableOpen(open)
          if (!open) setEditingVariable(null)
        }}
        onSave={handleSaveVariable}
        editVariable={editingVariable?.variable}
        existingNames={declaredNames}
      />
    </>
  )
}
