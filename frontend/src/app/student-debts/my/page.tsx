'use client'

import { useTranslations } from 'next-intl'
import { CheckCircle2, Loader2 } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { StudentDebtCard } from '@/components/student-debts/StudentDebtCard'
import { useMyStudentDebts } from '@/hooks/useStudentDebts'
import { useAuthCheck } from '@/hooks/useAuth'

// MyStudentDebtsPage — a student's own academic debts (the /my self-view).
// The backend resolves "own" from the JWT subject, so there is no filter UI:
// the list is the caller's complete debt set. Reached directly from the nav
// (students are redirected here from the registry) and from the dashboard
// widget.
export default function MyStudentDebtsPage() {
  const { isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('studentDebts')

  const enabled = !isLoading && isAuthenticated
  const { items, isLoading: listLoading, error } = useMyStudentDebts(undefined, { enabled })

  if (isLoading || !isAuthenticated) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </AppLayout>
    )
  }

  return (
    <AppLayout>
      <div className="max-w-4xl mx-auto space-y-6">
        <header>
          <h1 className="text-2xl font-bold">{t('my.title')}</h1>
          <p className="text-muted-foreground">{t('my.description')}</p>
        </header>

        {listLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center">
            <p className="font-medium text-destructive">{t('loadFailed')}</p>
          </div>
        ) : items.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <CheckCircle2 className="h-16 w-16 text-emerald-500/30 mb-4" />
            <h3 className="text-lg font-medium">{t('empty.title')}</h3>
            <p className="text-muted-foreground">{t('empty.description')}</p>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {items.map((debt) => (
              <StudentDebtCard key={debt.id} debt={debt} />
            ))}
          </div>
        )}
      </div>
    </AppLayout>
  )
}
