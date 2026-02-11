'use client'

import { memo } from 'react'
import { useTranslations } from 'next-intl'
import { Check, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface CreateVersionFormProps {
  description: string
  isCreating: boolean
  onDescriptionChange: (description: string) => void
  onCreate: () => void
  onCancel: () => void
}

export const CreateVersionForm = memo(function CreateVersionForm({
  description,
  isCreating,
  onDescriptionChange,
  onCreate,
  onCancel,
}: CreateVersionFormProps) {
  const t = useTranslations('documents.versions')
  const tCommon = useTranslations('common')

  return (
    <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg space-y-3">
      <input
        type="text"
        placeholder={t('descriptionPlaceholder')}
        value={description}
        onChange={(e) => onDescriptionChange(e.target.value)}
        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md
                   bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                   focus:ring-2 focus:ring-blue-500 focus:border-transparent"
      />
      <div className="flex gap-2">
        <Button size="sm" onClick={onCreate} disabled={isCreating}>
          {isCreating ? (
            <Loader2 className="h-4 w-4 animate-spin mr-2" />
          ) : (
            <Check className="h-4 w-4 mr-2" />
          )}
          {t('saveVersion')}
        </Button>
        <Button variant="ghost" size="sm" onClick={onCancel}>
          {tCommon('cancel')}
        </Button>
      </div>
    </div>
  )
})
