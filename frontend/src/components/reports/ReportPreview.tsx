'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { RefreshCw, AlertCircle, FileText } from 'lucide-react'
import type {
  DataSourceType,
  SelectedField,
  ReportFilter,
  ReportPreviewData,
} from '@/types/reports'

interface ReportPreviewProps {
  dataSource: DataSourceType
  selectedFields: SelectedField[]
  filters: ReportFilter[]
}

// Mock data generator for preview
function generateMockData(
  dataSource: DataSourceType,
  fields: SelectedField[],
  _filters: ReportFilter[]
): ReportPreviewData {
  if (fields.length === 0) {
    return { columns: [], rows: [], totalCount: 0 }
  }

  const columns = fields.map((f) => ({
    key: f.field.id,
    label: f.field.label,
  }))

  // Generate mock rows based on data source
  const mockRowsCount = Math.floor(Math.random() * 10) + 5
  const rows: Record<string, unknown>[] = []

  for (let i = 0; i < mockRowsCount; i++) {
    const row: Record<string, unknown> = {}
    fields.forEach((f) => {
      switch (f.field.type) {
        case 'string':
          row[f.field.id] = `${f.field.label} ${i + 1}`
          break
        case 'number':
          row[f.field.id] = Math.floor(Math.random() * 1000)
          break
        case 'date':
          const date = new Date()
          date.setDate(date.getDate() - Math.floor(Math.random() * 30))
          row[f.field.id] = date.toISOString().split('T')[0]
          break
        case 'boolean':
          row[f.field.id] = Math.random() > 0.5
          break
        case 'enum':
          if (f.field.enumValues && f.field.enumValues.length > 0) {
            row[f.field.id] =
              f.field.enumValues[Math.floor(Math.random() * f.field.enumValues.length)]
          }
          break
        default:
          row[f.field.id] = `Value ${i + 1}`
      }
    })
    rows.push(row)
  }

  return {
    columns,
    rows,
    totalCount: rows.length,
  }
}

export function ReportPreview({ dataSource, selectedFields, filters }: ReportPreviewProps) {
  const t = useTranslations('reports.builder')
  const [isLoading, setIsLoading] = useState(false)
  const [previewData, setPreviewData] = useState<ReportPreviewData | null>(null)

  const handleRefresh = () => {
    if (selectedFields.length === 0) return

    setIsLoading(true)
    // Simulate API call
    setTimeout(() => {
      const data = generateMockData(dataSource, selectedFields, filters)
      setPreviewData(data)
      setIsLoading(false)
    }, 500)
  }

  const formatCellValue = (value: unknown, type: string): string => {
    if (value === null || value === undefined) return '-'
    if (typeof value === 'boolean') return value ? '✓' : '✗'
    if (type === 'date' && typeof value === 'string') {
      return new Date(value).toLocaleDateString()
    }
    if (typeof value === 'number') {
      return value.toLocaleString()
    }
    return String(value)
  }

  return (
    <div className="relative overflow-hidden rounded-xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
      <GlowingEffect
        spread={40}
        glow={true}
        disabled={false}
        proximity={64}
        inactiveZone={0.01}
        borderWidth={2}
      />
      <div className="relative z-10">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{t('preview')}</h3>
          <Button
            onClick={handleRefresh}
            variant="outline"
            size="sm"
            disabled={selectedFields.length === 0 || isLoading}
            className="flex items-center gap-2"
          >
            <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            {t('refreshPreview')}
          </Button>
        </div>

        {selectedFields.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <div className="w-16 h-16 rounded-full bg-yellow-100 dark:bg-yellow-900/20 flex items-center justify-center mb-4">
              <AlertCircle className="h-8 w-8 text-yellow-500" />
            </div>
            <p className="text-gray-600 dark:text-gray-300 mb-2">{t('selectFieldsFirst')}</p>
            <p className="text-sm text-gray-400 dark:text-gray-500">{t('selectFieldsHint')}</p>
          </div>
        ) : !previewData ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <div className="w-16 h-16 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-4">
              <FileText className="h-8 w-8 text-gray-400" />
            </div>
            <p className="text-gray-600 dark:text-gray-300 mb-4">{t('clickToPreview')}</p>
            <Button onClick={handleRefresh} disabled={isLoading}>
              {isLoading ? t('loading') : t('generatePreview')}
            </Button>
          </div>
        ) : (
          <div className="space-y-4">
            {/* Results Count */}
            <div className="flex items-center justify-between text-sm text-gray-500 dark:text-gray-400">
              <span>{t('showingResults', { count: previewData.rows.length })}</span>
              {filters.length > 0 && <span>{t('filtersApplied', { count: filters.length })}</span>}
            </div>

            {/* Data Table */}
            <div className="border rounded-lg overflow-hidden">
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-gray-50 dark:bg-gray-800">
                      {previewData.columns.map((col) => (
                        <TableHead
                          key={col.key}
                          className="font-semibold text-gray-900 dark:text-white whitespace-nowrap"
                        >
                          {t(
                            `fields.${dataSource}.${selectedFields.find((f) => f.field.id === col.key)?.field.name}`,
                            { defaultValue: col.label }
                          )}
                        </TableHead>
                      ))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {previewData.rows.map((row, rowIndex) => (
                      <TableRow
                        key={rowIndex}
                        className="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
                      >
                        {previewData.columns.map((col) => {
                          const field = selectedFields.find((f) => f.field.id === col.key)
                          return (
                            <TableCell
                              key={col.key}
                              className="text-gray-700 dark:text-gray-300 whitespace-nowrap"
                            >
                              {formatCellValue(row[col.key], field?.field.type || 'string')}
                            </TableCell>
                          )
                        })}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>

            {/* Preview Note */}
            <p className="text-xs text-gray-400 dark:text-gray-500 text-center">
              {t('previewNote')}
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
