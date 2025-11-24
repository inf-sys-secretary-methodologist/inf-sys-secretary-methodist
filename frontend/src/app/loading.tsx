export default function Loading() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="space-y-4 text-center">
        <div className="relative w-16 h-16 mx-auto">
          <div className="absolute top-0 left-0 w-full h-full border-4 border-gray-200 dark:border-gray-700 rounded-full"></div>
          <div className="absolute top-0 left-0 w-full h-full border-4 border-gray-900 dark:border-white border-t-transparent rounded-full animate-spin"></div>
        </div>
        <p className="text-gray-600 dark:text-gray-300 font-medium">03@C7:0...</p>
      </div>
    </div>
  )
}
