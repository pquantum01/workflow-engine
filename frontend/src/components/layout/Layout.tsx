import { useState, useEffect } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import {
  GitBranch, List, RefreshCw, Menu, X, Sun, Moon, Circle,
} from 'lucide-react'

function useHealthCheck() {
  const [healthy, setHealthy] = useState<boolean | null>(null)
  useEffect(() => {
    const check = () =>
      fetch('/readyz')
        .then((r) => setHealthy(r.ok))
        .catch(() => setHealthy(false))
    check()
    const t = setInterval(check, 30_000)
    return () => clearInterval(t)
  }, [])
  return healthy
}

function useDarkMode() {
  const [dark, setDark] = useState(() => {
    const saved = localStorage.getItem('theme')
    if (saved) return saved === 'dark'
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  })
  useEffect(() => {
    document.documentElement.classList.toggle('dark', dark)
    localStorage.setItem('theme', dark ? 'dark' : 'light')
  }, [dark])
  return [dark, setDark] as const
}

const navItems = [
  { to: '/workflow-defs', label: 'Workflows', icon: GitBranch },
  { to: '/workflow-instances', label: 'Instances', icon: List },
  { to: '/admin/reload', label: 'Admin', icon: RefreshCw },
]

export function Layout() {
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [dark, setDark] = useDarkMode()
  const healthy = useHealthCheck()

  return (
    <div className="flex h-screen bg-zinc-50 dark:bg-zinc-950 text-zinc-900 dark:text-zinc-50">
      {/* Sidebar */}
      <aside
        className={`flex flex-col border-r border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 transition-all duration-200 ${
          sidebarOpen ? 'w-52' : 'w-14'
        }`}
      >
        <div className="flex items-center gap-2 px-3 py-4 border-b border-zinc-200 dark:border-zinc-800">
          <button
            onClick={() => setSidebarOpen((s) => !s)}
            className="p-1 rounded hover:bg-zinc-100 dark:hover:bg-zinc-800 text-zinc-500"
          >
            {sidebarOpen ? <X size={18} /> : <Menu size={18} />}
          </button>
          {sidebarOpen && (
            <span className="text-sm font-semibold tracking-tight text-indigo-600">
              Workflow Engine
            </span>
          )}
        </div>

        <nav className="flex-1 p-2 space-y-1">
          {navItems.map(({ to, label, icon: Icon }) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                `flex items-center gap-2.5 px-2 py-2 rounded-lg text-sm transition-colors ${
                  isActive
                    ? 'bg-indigo-50 dark:bg-indigo-950 text-indigo-600 dark:text-indigo-400 font-medium'
                    : 'text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800'
                }`
              }
            >
              <Icon size={16} className="shrink-0" />
              {sidebarOpen && <span>{label}</span>}
            </NavLink>
          ))}
        </nav>
      </aside>

      {/* Main */}
      <div className="flex flex-col flex-1 overflow-hidden">
        {/* Topbar */}
        <header className="flex items-center justify-between px-6 h-14 border-b border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900">
          <div />
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1.5 text-xs text-zinc-500">
              <Circle
                size={8}
                className={
                  healthy === null
                    ? 'text-zinc-400 fill-zinc-400'
                    : healthy
                    ? 'text-green-500 fill-green-500'
                    : 'text-red-500 fill-red-500'
                }
              />
              {healthy === null ? 'checking' : healthy ? 'healthy' : 'unhealthy'}
            </div>
            <button
              onClick={() => setDark((d) => !d)}
              className="p-1.5 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 text-zinc-500"
            >
              {dark ? <Sun size={16} /> : <Moon size={16} />}
            </button>
          </div>
        </header>

        <main className="flex-1 overflow-auto">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
