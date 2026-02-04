'use client'

import { useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import { AlertTriangle, Loader2, ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { StudentRiskCard } from './StudentRiskCard'
import { analyticsApi, StudentRiskInfo, RiskLevel } from '@/lib/api/analytics'

interface AtRiskStudentsListProps {
  riskLevel?: RiskLevel
  pageSize?: number
  onStudentClick?: (student: StudentRiskInfo) => void
}

export function AtRiskStudentsList({
  riskLevel,
  pageSize = 9,
  onStudentClick,
}: AtRiskStudentsListProps) {
  const t = useTranslations('analytics')
  const [students, setStudents] = useState<StudentRiskInfo[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  useEffect(() => {
    const fetchStudents = async () => {
      try {
        setIsLoading(true)
        setError(null)

        let data
        if (riskLevel) {
          data = await analyticsApi.getStudentsByRiskLevel(riskLevel, page, pageSize)
        } else {
          data = await analyticsApi.getAtRiskStudents(page, pageSize)
        }

        setStudents(data.students)
        setTotal(data.total)
      } catch (err) {
        console.error('Failed to fetch at-risk students:', err)
        setError(t('loadError'))
      } finally {
        setIsLoading(false)
      }
    }

    fetchStudents()
  }, [page, pageSize, riskLevel, t])

  const totalPages = Math.ceil(total / pageSize)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-gray-500" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <AlertTriangle className="h-12 w-12 mx-auto text-red-500 mb-4" />
        <p className="text-red-500">{error}</p>
      </div>
    )
  }

  if (students.length === 0) {
    return (
      <div className="text-center py-12">
        <AlertTriangle className="h-12 w-12 mx-auto text-gray-400 mb-4" />
        <p className="text-gray-600 dark:text-gray-400">{t('noAtRiskStudents')}</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Students Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {students.map((student) => (
          <StudentRiskCard key={student.student_id} student={student} onClick={onStudentClick} />
        ))}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            {t('showingStudents', {
              from: (page - 1) * pageSize + 1,
              to: Math.min(page * pageSize, total),
              total,
            })}
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => p - 1)}
              disabled={page === 1}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span className="text-sm text-gray-600 dark:text-gray-400 px-2">
              {page} / {totalPages}
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => p + 1)}
              disabled={page === totalPages}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
