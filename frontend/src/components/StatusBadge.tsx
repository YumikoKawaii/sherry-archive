import type { MangaStatus } from '../types/manga'

const config: Record<MangaStatus, { label: string; classes: string }> = {
  ongoing:   { label: 'Ongoing',   classes: 'bg-jade-500/15 text-jade-400 border-jade-500/30' },
  completed: { label: 'Completed', classes: 'bg-mint-100/10 text-mint-200 border-mint-200/20' },
  hiatus:    { label: 'Hiatus',    classes: 'bg-amber-500/15 text-amber-400 border-amber-500/30' },
}

export function StatusBadge({ status }: { status: MangaStatus }) {
  const { label, classes } = config[status] ?? config.ongoing
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${classes}`}>
      {label}
    </span>
  )
}
