interface Props {
  priority: 'low' | 'medium' | 'high'
}

const config = {
  low:    { label: 'Low',    cls: 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300' },
  medium: { label: 'Medium', cls: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/60 dark:text-yellow-400' },
  high:   { label: 'High',   cls: 'bg-red-100 text-red-600 dark:bg-red-900/60 dark:text-red-400' },
}

export function PriorityBadge({ priority }: Props) {
  const { label, cls } = config[priority] ?? config.medium
  return (
    <span className={`inline-flex items-center rounded px-2 py-0.5 text-xs font-medium ${cls}`}>
      {label}
    </span>
  )
}
