import type { LucideIcon } from 'lucide-react'

interface Props {
  icon: LucideIcon
  title: string
  description?: string
  action?: React.ReactNode
}

export function EmptyState({ icon: Icon, title, description, action }: Props) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <Icon size={48} className="text-zinc-300 dark:text-zinc-600 mb-4" />
      <h3 className="text-base font-medium text-zinc-700 dark:text-zinc-300">{title}</h3>
      {description && (
        <p className="text-sm text-zinc-500 mt-1 max-w-sm">{description}</p>
      )}
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}
