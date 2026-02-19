export function TagBadge({ tag }: { tag: string }) {
  return (
    <span className="inline-block px-2.5 py-0.5 rounded-full text-xs font-medium
                     bg-forest-700 text-mint-200 border border-forest-600
                     hover:border-jade-500/50 hover:text-jade-300 transition-colors cursor-default">
      {tag}
    </span>
  )
}
