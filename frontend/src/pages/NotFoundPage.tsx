import { Link } from 'react-router-dom'

export function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center h-64 text-center">
      <div className="text-4xl font-bold text-zinc-200 dark:text-zinc-800 mb-2">404</div>
      <p className="text-zinc-500 mb-4">Page not found.</p>
      <Link to="/" className="text-indigo-600 hover:text-indigo-700 text-sm">
        Go home
      </Link>
    </div>
  )
}
