import { Skeleton } from '@/components/ui/skeleton'

export default function DashboardLoading() {
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
          {/* Welcome section */}
          <div className="space-y-2">
            <Skeleton className="h-9 w-80" />
            <Skeleton className="h-5 w-64" />
          </div>

          {/* Stats grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {[...Array(4)].map((_, i) => (
              <div
                key={i}
                className="rounded-xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700"
              >
                <div className="flex items-center justify-between">
                  <Skeleton className="h-10 w-10 rounded-lg" />
                  <Skeleton className="h-6 w-16" />
                </div>
                <div className="mt-4 space-y-2">
                  <Skeleton className="h-8 w-20" />
                  <Skeleton className="h-4 w-32" />
                </div>
              </div>
            ))}
          </div>

          {/* Charts grid */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="rounded-xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
              <Skeleton className="h-6 w-40 mb-4" />
              <Skeleton className="h-64 w-full rounded-lg" />
            </div>
            <div className="rounded-xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
              <Skeleton className="h-6 w-40 mb-4" />
              <Skeleton className="h-64 w-full rounded-lg" />
            </div>
          </div>

          {/* Recent activity */}
          <div className="rounded-xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
            <Skeleton className="h-6 w-48 mb-4" />
            <div className="space-y-3">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="flex items-center gap-4">
                  <Skeleton className="h-10 w-10 rounded-full" />
                  <div className="flex-1 space-y-1">
                    <Skeleton className="h-4 w-3/4" />
                    <Skeleton className="h-3 w-1/2" />
                  </div>
                  <Skeleton className="h-4 w-16" />
                </div>
              ))}
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
