import { Skeleton } from '@/components/ui/skeleton'

export default function AnalyticsLoading() {
  return (
    <div className="min-h-screen bg-background">
      {/* Header skeleton */}
      <header className="sticky top-0 z-50 w-full pt-4 bg-background/95 backdrop-blur">
        <div className="hidden lg:flex h-14 items-center justify-between px-6 xl:px-8">
          <div className="w-44" />
          <div className="flex items-center gap-1 rounded-full bg-white/80 dark:bg-gray-900/80 backdrop-blur-lg border border-gray-200 dark:border-gray-700 px-3 py-2">
            {[...Array(6)].map((_, i) => (
              <Skeleton key={i} className="h-8 w-24 rounded-full" />
            ))}
          </div>
          <div className="flex items-center gap-2 w-44">
            <Skeleton className="h-9 w-9 rounded-full" />
            <Skeleton className="h-9 w-9 rounded-full" />
            <Skeleton className="h-9 w-9 rounded-full" />
            <Skeleton className="h-9 w-9 rounded-full" />
          </div>
        </div>
      </header>

      <main className="px-4 sm:px-6 lg:px-8 py-6 sm:py-8 lg:py-10">
        <div className="max-w-7xl mx-auto space-y-6 sm:space-y-8">
          {/* Title */}
          <div className="text-center space-y-4">
            <Skeleton className="h-10 w-64 mx-auto" />
            <Skeleton className="h-5 w-96 mx-auto" />
          </div>

          {/* Tabs */}
          <div className="flex justify-center">
            <div className="flex gap-2 p-1 rounded-lg bg-muted">
              <Skeleton className="h-9 w-32 rounded-md" />
              <Skeleton className="h-9 w-32 rounded-md" />
              <Skeleton className="h-9 w-32 rounded-md" />
            </div>
          </div>

          {/* KPI Cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {[...Array(4)].map((_, i) => (
              <div
                key={i}
                className="rounded-xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700"
              >
                <div className="flex items-center gap-3 mb-4">
                  <Skeleton className="h-10 w-10 rounded-lg" />
                  <Skeleton className="h-5 w-24" />
                </div>
                <Skeleton className="h-8 w-16 mb-2" />
                <Skeleton className="h-4 w-32" />
              </div>
            ))}
          </div>

          {/* Charts */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="rounded-xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
              <Skeleton className="h-6 w-48 mb-4" />
              <Skeleton className="h-80 w-full rounded-lg" />
            </div>
            <div className="rounded-xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
              <Skeleton className="h-6 w-48 mb-4" />
              <Skeleton className="h-80 w-full rounded-lg" />
            </div>
          </div>

          {/* Table */}
          <div className="rounded-xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
            <Skeleton className="h-6 w-40 mb-4" />
            <div className="space-y-3">
              <Skeleton className="h-10 w-full" />
              {[...Array(5)].map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
