interface Props {
  className?: string
}

export function Skeleton({ className = '' }: Props) {
  return (
    <div className={`animate-pulse bg-zinc-200 dark:bg-zinc-800 rounded ${className}`} />
  )
}
