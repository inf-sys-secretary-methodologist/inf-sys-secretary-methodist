import Link from 'next/link'

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <div className="max-w-md w-full space-y-6 text-center">
        <div className="space-y-2">
          <h1 className="text-9xl font-bold text-gray-900 dark:text-white">404</h1>
          <h2 className="text-2xl font-semibold text-gray-900 dark:text-white">
            Страница не найдена
          </h2>
          <p className="text-gray-600 dark:text-gray-300">
            К сожалению, запрашиваемая страница не существует или была удалена.
          </p>
        </div>

        <Link
          href="/"
          className="inline-block px-6 py-3 rounded-lg font-medium transition-all duration-300 bg-gray-900 dark:bg-white text-white dark:text-gray-900 hover:bg-gray-800 dark:hover:bg-gray-100 hover:scale-105 active:scale-95 shadow-lg hover:shadow-xl"
        >
          Вернуться на главную
        </Link>
      </div>
    </div>
  )
}
