'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { FileText, Code, Eye } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { TemplateInfo } from '@/lib/api/templates'

interface TemplatePreviewDialogProps {
  template: TemplateInfo | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreate: (template: TemplateInfo) => void
}

export function TemplatePreviewDialog({
  template,
  open,
  onOpenChange,
  onCreate,
}: TemplatePreviewDialogProps) {
  const t = useTranslations('templates')
  const [activeTab, setActiveTab] = useState('info')

  if (!template) return null

  const handleCreate = () => {
    onOpenChange(false)
    onCreate(template)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            {template.name}
          </DialogTitle>
          <DialogDescription>{template.description}</DialogDescription>
        </DialogHeader>

        <Tabs value={activeTab} onValueChange={setActiveTab} className="mt-4">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="info" className="flex items-center gap-1">
              <Eye className="h-4 w-4" />
              {t('info')}
            </TabsTrigger>
            <TabsTrigger value="variables" className="flex items-center gap-1">
              <Code className="h-4 w-4" />
              {t('variables')}
            </TabsTrigger>
            <TabsTrigger value="content" className="flex items-center gap-1">
              <FileText className="h-4 w-4" />
              {t('content')}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="info" className="space-y-4 mt-4">
            <div className="grid gap-4">
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600 dark:text-gray-400">{t('code')}:</span>
                <Badge variant="outline">{template.code}</Badge>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600 dark:text-gray-400">
                  {t('variablesCount')}:
                </span>
                <Badge variant="secondary">{template.template_variables?.length || 0}</Badge>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600 dark:text-gray-400">
                  {t('hasTemplate')}:
                </span>
                <Badge variant={template.has_template ? 'default' : 'secondary'}>
                  {template.has_template ? t('yes') : t('no')}
                </Badge>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="variables" className="mt-4">
            {template.template_variables && template.template_variables.length > 0 ? (
              <div className="space-y-3">
                {template.template_variables.map((variable) => (
                  <div
                    key={variable.name}
                    className="rounded-lg border border-gray-200 dark:border-gray-700 p-3"
                  >
                    <div className="flex items-center justify-between mb-1">
                      <span className="font-medium text-gray-900 dark:text-white">
                        {variable.name}
                      </span>
                      <div className="flex gap-2">
                        <Badge variant="outline" className="text-xs">
                          {variable.variable_type}
                        </Badge>
                        {variable.required && (
                          <Badge variant="destructive" className="text-xs">
                            {t('required')}
                          </Badge>
                        )}
                      </div>
                    </div>
                    {variable.description && (
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        {variable.description}
                      </p>
                    )}
                    {variable.default_value && (
                      <p className="text-xs text-gray-500 mt-1">
                        {t('defaultValue')}: {variable.default_value}
                      </p>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-gray-500">{t('noVariables')}</div>
            )}
          </TabsContent>

          <TabsContent value="content" className="mt-4">
            {template.template_content ? (
              <div className="rounded-lg border bg-gray-50 dark:bg-gray-900 p-4 max-h-80 overflow-y-auto">
                <pre className="whitespace-pre-wrap text-sm text-gray-700 dark:text-gray-300 font-mono">
                  {template.template_content}
                </pre>
              </div>
            ) : (
              <div className="text-center py-8 text-gray-500">{t('noContent')}</div>
            )}
          </TabsContent>
        </Tabs>

        <div className="flex justify-end gap-2 mt-6">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('close')}
          </Button>
          <Button onClick={handleCreate} className="bg-blue-600 hover:bg-blue-700 text-white">
            {t('createFromThis')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
