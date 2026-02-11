'use client'

import { useState, useCallback, Suspense } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { useAuthCheck } from '@/hooks/useAuth'
import { AppLayout } from '@/components/layout'
import { ReportBuilderHeader } from '@/components/reports/ReportBuilderHeader'
import { DataSourceSelector } from '@/components/reports/DataSourceSelector'
import { FieldSelector } from '@/components/reports/FieldSelector'
import { FilterBuilder } from '@/components/reports/FilterBuilder'
import { ReportPreview } from '@/components/reports/ReportPreview'
import type { DataSourceType, SelectedField, ReportFilter, ReportField } from '@/types/reports'

function ReportBuilderContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { user: _user } = useAuthCheck()
  const t = useTranslations('reports.builder')

  // Get initial values from URL params
  const initialSource = (searchParams.get('source') as DataSourceType) || 'documents'
  // TODO: Planned feature - Use templateId for loading template data from backend
  const _templateId = searchParams.get('template')

  // State
  const [reportName, setReportName] = useState(t('untitledReport'))
  const [dataSource, setDataSource] = useState<DataSourceType>(initialSource)
  const [selectedFields, setSelectedFields] = useState<SelectedField[]>([])
  const [filters, setFilters] = useState<ReportFilter[]>([])
  const [activeTab, setActiveTab] = useState<'fields' | 'filters' | 'preview'>('fields')

  // Handlers
  const handleDataSourceChange = useCallback((source: DataSourceType) => {
    setDataSource(source)
    setSelectedFields([])
    setFilters([])
  }, [])

  const handleAddField = useCallback((field: ReportField) => {
    setSelectedFields((prev) => {
      if (prev.some((f) => f.field.id === field.id)) {
        return prev
      }
      return [...prev, { field, order: prev.length }]
    })
  }, [])

  const handleRemoveField = useCallback((fieldId: string) => {
    setSelectedFields((prev) => prev.filter((f) => f.field.id !== fieldId))
  }, [])

  const handleReorderFields = useCallback((newOrder: SelectedField[]) => {
    setSelectedFields(newOrder.map((f, i) => ({ ...f, order: i })))
  }, [])

  const handleAddFilter = useCallback((filter: ReportFilter) => {
    setFilters((prev) => [...prev, filter])
  }, [])

  const handleRemoveFilter = useCallback((filterId: string) => {
    setFilters((prev) => prev.filter((f) => f.id !== filterId))
  }, [])

  const handleUpdateFilter = useCallback((filterId: string, updates: Partial<ReportFilter>) => {
    setFilters((prev) => prev.map((f) => (f.id === filterId ? { ...f, ...updates } : f)))
  }, [])

  const handleSave = useCallback(() => {
    // TODO: Planned feature - Implement save to backend API
    // This will persist the report configuration to the database
    console.warn('Save not implemented yet:', {
      name: reportName,
      dataSource,
      fields: selectedFields,
      filters,
    })
  }, [reportName, dataSource, selectedFields, filters])

  const handleExport = useCallback((format: 'pdf' | 'xlsx' | 'csv') => {
    // TODO: Planned feature - Implement export functionality
    // This will generate report in the specified format (pdf/xlsx/csv)
    console.warn('Export not implemented yet:', format)
  }, [])

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <ReportBuilderHeader
          reportName={reportName}
          onNameChange={setReportName}
          onSave={handleSave}
          onExport={handleExport}
          onBack={() => router.push('/reports')}
        />

        {/* Data Source Selector */}
        <DataSourceSelector selected={dataSource} onChange={handleDataSourceChange} />

        {/* Tab Navigation */}
        <div className="flex border-b border-gray-200 dark:border-gray-700">
          <button
            onClick={() => setActiveTab('fields')}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === 'fields'
                ? 'border-gray-900 dark:border-white text-gray-900 dark:text-white'
                : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
            }`}
          >
            {t('tabs.fields')}
            {selectedFields.length > 0 && (
              <span className="ml-2 px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded-full text-xs">
                {selectedFields.length}
              </span>
            )}
          </button>
          <button
            onClick={() => setActiveTab('filters')}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === 'filters'
                ? 'border-gray-900 dark:border-white text-gray-900 dark:text-white'
                : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
            }`}
          >
            {t('tabs.filters')}
            {filters.length > 0 && (
              <span className="ml-2 px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded-full text-xs">
                {filters.length}
              </span>
            )}
          </button>
          <button
            onClick={() => setActiveTab('preview')}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === 'preview'
                ? 'border-gray-900 dark:border-white text-gray-900 dark:text-white'
                : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
            }`}
          >
            {t('tabs.preview')}
          </button>
        </div>

        {/* Tab Content */}
        <div className="min-h-[400px]">
          {activeTab === 'fields' && (
            <FieldSelector
              dataSource={dataSource}
              selectedFields={selectedFields}
              onAddField={handleAddField}
              onRemoveField={handleRemoveField}
              onReorderFields={handleReorderFields}
            />
          )}

          {activeTab === 'filters' && (
            <FilterBuilder
              dataSource={dataSource}
              filters={filters}
              onAddFilter={handleAddFilter}
              onRemoveFilter={handleRemoveFilter}
              onUpdateFilter={handleUpdateFilter}
            />
          )}

          {activeTab === 'preview' && (
            <ReportPreview
              dataSource={dataSource}
              selectedFields={selectedFields}
              filters={filters}
            />
          )}
        </div>
      </div>
    </AppLayout>
  )
}

export default function ReportBuilderPage() {
  return (
    <Suspense
      fallback={<div className="flex items-center justify-center min-h-screen">Loading...</div>}
    >
      <ReportBuilderContent />
    </Suspense>
  )
}
