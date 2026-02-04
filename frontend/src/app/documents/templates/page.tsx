'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { useAuthCheck } from '@/hooks/useAuth'
import { AppLayout } from '@/components/layout'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { Button } from '@/components/ui/button'
import { FileText, ArrowLeft } from 'lucide-react'
import Link from 'next/link'
import {
  TemplateList,
  CreateFromTemplateDialog,
  TemplatePreviewDialog,
} from '@/components/templates'
import { TemplateInfo } from '@/lib/api/templates'
import { canEdit } from '@/lib/auth/permissions'

export default function TemplatesPage() {
  const { user } = useAuthCheck()
  const t = useTranslations('templates')
  const userCanEdit = canEdit(user?.role)

  const [previewTemplate, setPreviewTemplate] = useState<TemplateInfo | null>(null)
  const [createTemplate, setCreateTemplate] = useState<TemplateInfo | null>(null)

  const handlePreview = (template: TemplateInfo) => {
    setPreviewTemplate(template)
  }

  const handleCreate = (template: TemplateInfo) => {
    setCreateTemplate(template)
  }

  const handleCreateFromPreview = (template: TemplateInfo) => {
    setPreviewTemplate(null)
    setCreateTemplate(template)
  }

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto space-y-6 sm:space-y-8">
        {/* Page Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-2">
            <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-gray-900 dark:text-white">
              {t('title')}
            </h1>
            <p className="text-base sm:text-lg text-gray-600 dark:text-gray-300">{t('subtitle')}</p>
          </div>
          <Link href="/documents">
            <Button variant="outline" className="flex items-center gap-2">
              <ArrowLeft className="h-4 w-4" />
              {t('backToDocuments')}
            </Button>
          </Link>
        </div>

        {/* Templates List */}
        <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 lg:p-8 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
          <GlowingEffect
            spread={40}
            glow={true}
            disabled={false}
            proximity={64}
            inactiveZone={0.01}
            borderWidth={3}
          />
          <div className="relative z-10">
            <div className="flex items-center gap-3 mb-6">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400">
                <FileText className="h-5 w-5" />
              </div>
              <div>
                <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                  {t('availableTemplates')}
                </h2>
                <p className="text-sm text-gray-600 dark:text-gray-400">{t('selectToCreate')}</p>
              </div>
            </div>

            <TemplateList onPreview={handlePreview} onCreate={handleCreate} canEdit={userCanEdit} />
          </div>
        </div>
      </div>

      {/* Preview Dialog */}
      <TemplatePreviewDialog
        template={previewTemplate}
        open={previewTemplate !== null}
        onOpenChange={(open) => !open && setPreviewTemplate(null)}
        onCreate={handleCreateFromPreview}
      />

      {/* Create Dialog */}
      <CreateFromTemplateDialog
        template={createTemplate}
        open={createTemplate !== null}
        onOpenChange={(open) => !open && setCreateTemplate(null)}
      />
    </AppLayout>
  )
}
