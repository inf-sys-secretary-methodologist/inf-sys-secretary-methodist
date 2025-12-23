'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { Button } from '@/components/ui/button'
import {
  ArrowLeft,
  Save,
  Download,
  FileSpreadsheet,
  FileText,
  FileDown,
  ChevronDown,
} from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

interface ReportBuilderHeaderProps {
  reportName: string
  onNameChange: (name: string) => void
  onSave: () => void
  onExport: (format: 'pdf' | 'xlsx' | 'csv') => void
  onBack: () => void
}

export function ReportBuilderHeader({
  reportName,
  onNameChange,
  onSave,
  onExport,
  onBack,
}: ReportBuilderHeaderProps) {
  const t = useTranslations('reports.builder')
  const [isEditing, setIsEditing] = useState(false)

  return (
    <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          onClick={onBack}
          className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
        >
          <ArrowLeft className="h-5 w-5" />
        </Button>

        {isEditing ? (
          <input
            type="text"
            value={reportName}
            onChange={(e) => onNameChange(e.target.value)}
            onBlur={() => setIsEditing(false)}
            onKeyDown={(e) => e.key === 'Enter' && setIsEditing(false)}
            autoFocus
            className="text-xl sm:text-2xl font-bold bg-transparent border-b-2 border-gray-300 dark:border-gray-600 focus:border-gray-900 dark:focus:border-white outline-none text-gray-900 dark:text-white"
          />
        ) : (
          <h1
            onClick={() => setIsEditing(true)}
            className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white cursor-pointer hover:opacity-80 transition-opacity"
          >
            {reportName}
          </h1>
        )}
      </div>

      <div className="flex items-center gap-2 sm:gap-3">
        <Button
          onClick={onSave}
          variant="outline"
          className="flex items-center gap-2 border-gray-300 dark:border-gray-600"
        >
          <Save className="h-4 w-4" />
          <span className="hidden sm:inline">{t('save')}</span>
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button className="flex items-center gap-2 bg-gray-900 dark:bg-white text-white dark:text-gray-900 hover:bg-gray-800 dark:hover:bg-gray-100">
              <Download className="h-4 w-4" />
              <span className="hidden sm:inline">{t('export')}</span>
              <ChevronDown className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <DropdownMenuItem onClick={() => onExport('pdf')} className="cursor-pointer">
              <FileText className="h-4 w-4 mr-2" />
              {t('exportPdf')}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onExport('xlsx')} className="cursor-pointer">
              <FileSpreadsheet className="h-4 w-4 mr-2" />
              {t('exportExcel')}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onExport('csv')} className="cursor-pointer">
              <FileDown className="h-4 w-4 mr-2" />
              {t('exportCsv')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  )
}
