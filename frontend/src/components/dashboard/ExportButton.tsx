'use client'

import { useState } from 'react'
import { Download, FileSpreadsheet, FileText, Loader2 } from 'lucide-react'
import { exportDashboard } from '@/hooks/useDashboard'
import type { ExportInput } from '@/types/dashboard'

interface ExportButtonProps {
  className?: string
}

export function ExportButton({ className }: ExportButtonProps) {
  const [isExporting, setIsExporting] = useState(false)
  const [showMenu, setShowMenu] = useState(false)

  const handleExport = async (format: 'pdf' | 'xlsx') => {
    setIsExporting(true)
    setShowMenu(false)

    try {
      const input: ExportInput = {
        format,
        sections: ['stats', 'trends', 'activity'],
      }

      const result = await exportDashboard(input)

      // Download the file
      const link = document.createElement('a')
      link.href = result.file_url
      link.download = result.file_name
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    } catch (error) {
      console.error('Export failed:', error)
    } finally {
      setIsExporting(false)
    }
  }

  return (
    <div className={`relative ${className}`}>
      <button
        onClick={() => setShowMenu(!showMenu)}
        disabled={isExporting}
        className="flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-105 active:scale-95 hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {isExporting ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : (
          <Download className="h-4 w-4" />
        )}
        Экспорт
      </button>

      {showMenu && !isExporting && (
        <div className="absolute top-full right-0 mt-2 w-48 rounded-lg bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 shadow-lg z-50">
          <button
            onClick={() => handleExport('pdf')}
            className="flex items-center gap-3 w-full px-4 py-3 text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors rounded-t-lg"
          >
            <FileText className="h-4 w-4 text-red-500" />
            <span className="text-gray-900 dark:text-white">Экспорт в PDF</span>
          </button>
          <button
            onClick={() => handleExport('xlsx')}
            className="flex items-center gap-3 w-full px-4 py-3 text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors rounded-b-lg"
          >
            <FileSpreadsheet className="h-4 w-4 text-green-500" />
            <span className="text-gray-900 dark:text-white">Экспорт в Excel</span>
          </button>
        </div>
      )}
    </div>
  )
}
